package collector

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ProcessMeta holds metadata about a process that owns a socket.
type ProcessMeta struct {
	PID         int
	ProcessName string
	Cmdline     string
	User        string
	UID         string
}

// InodeMapper maintains a cache of socket inode to ProcessMeta mappings.
type InodeMapper struct {
	mu         sync.RWMutex
	cache      map[uint64]*ProcessMeta
	lastScan   time.Time
	cacheTTL   time.Duration
	userLookup map[string]string // UID -> Username cache
}

// NewInodeMapper creates a new InodeMapper with specified cache TTL.
func NewInodeMapper(ttl time.Duration) *InodeMapper {
	if ttl <= 0 {
		ttl = 1500 * time.Millisecond
	}
	return &InodeMapper{
		cache:      make(map[uint64]*ProcessMeta),
		cacheTTL:   ttl,
		userLookup: make(map[string]string),
	}
}

// Refresh scans /proc to rebuild the socket inode to PID mapping.
func (m *InodeMapper) Refresh() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// If cache is fresh, skip scanning
	if time.Since(m.lastScan) < m.cacheTTL && len(m.cache) > 0 {
		return nil
	}

	entries, err := os.ReadDir("/proc")
	if err != nil {
		return err
	}

	newCache := make(map[uint64]*ProcessMeta, len(m.cache))

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Only check directories that are numeric (PIDs)
		pid, err := strconv.Atoi(name)
		if err != nil || pid <= 0 {
			continue
		}

		// Read /proc/[pid]/fd
		fdPath := filepath.Join("/proc", name, "fd")
		fds, err := os.ReadDir(fdPath)
		if err != nil {
			// Permission denied or process exited, ignore
			continue
		}

		var meta *ProcessMeta

		for _, fd := range fds {
			linkTarget, err := os.Readlink(filepath.Join(fdPath, fd.Name()))
			if err != nil {
				continue
			}

			// Looking for socket:[12345]
			if strings.HasPrefix(linkTarget, "socket:[") && strings.HasSuffix(linkTarget, "]") {
				inodeStr := linkTarget[8 : len(linkTarget)-1]
				inode, err := strconv.ParseUint(inodeStr, 10, 64)
				if err != nil {
					continue
				}

				if meta == nil {
					meta = m.resolveProcessMeta(pid, name)
				}
				newCache[inode] = meta
			}
		}
	}

	m.cache = newCache
	m.lastScan = time.Now()
	return nil
}

// Lookup finds the process metadata associated with a socket inode.
func (m *InodeMapper) Lookup(inode uint64) *ProcessMeta {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cache[inode]
}

// resolveProcessMeta reads process name, command line, and user for a PID.
func (m *InodeMapper) resolveProcessMeta(pid int, pidStr string) *ProcessMeta {
	procDir := filepath.Join("/proc", pidStr)

	// Read process name from /proc/[pid]/comm
	var procName string
	commData, err := os.ReadFile(filepath.Join(procDir, "comm"))
	if err == nil {
		procName = strings.TrimSpace(string(commData))
	}

	// Read full cmdline from /proc/[pid]/cmdline
	var cmdline string
	cmdData, err := os.ReadFile(filepath.Join(procDir, "cmdline"))
	if err == nil {
		// Null bytes separate arguments
		parts := bytes.Split(cmdData, []byte{0})
		var cleanParts []string
		for _, p := range parts {
			if len(p) > 0 {
				cleanParts = append(cleanParts, string(p))
			}
		}
		cmdline = strings.Join(cleanParts, " ")
		if procName == "" && len(cleanParts) > 0 {
			procName = filepath.Base(cleanParts[0])
		}
	}

	if procName == "" {
		procName = fmt.Sprintf("pid-%d", pid)
	}

	// Read UID from /proc/[pid]/status
	var uidStr, userName string
	statusFile, err := os.Open(filepath.Join(procDir, "status"))
	if err == nil {
		scanner := bufio.NewScanner(statusFile)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "Uid:") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					uidStr = fields[1]
				}
				break
			}
		}
		statusFile.Close()
	}

	if uidStr != "" {
		if cachedUser, ok := m.userLookup[uidStr]; ok {
			userName = cachedUser
		} else {
			if u, err := user.LookupId(uidStr); err == nil {
				userName = u.Username
			} else {
				userName = uidStr
			}
			m.userLookup[uidStr] = userName
		}
	}

	return &ProcessMeta{
		PID:         pid,
		ProcessName: procName,
		Cmdline:     cmdline,
		User:        userName,
		UID:         uidStr,
	}
}
