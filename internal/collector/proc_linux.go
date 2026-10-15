package collector

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/BraveRam/netpulse/internal/model"
)

// LinuxSocketParser parses /proc/net/tcp, /proc/net/tcp6, /proc/net/udp, /proc/net/udp6.
type LinuxSocketParser struct {
	inodeMapper *InodeMapper
}

// NewLinuxSocketParser creates a LinuxSocketParser.
func NewLinuxSocketParser(mapper *InodeMapper) *LinuxSocketParser {
	return &LinuxSocketParser{
		inodeMapper: mapper,
	}
}

// ParseAll reads all TCP and UDP sockets from /proc/net/.
func (p *LinuxSocketParser) ParseAll() ([]*model.Connection, error) {
	if p.inodeMapper != nil {
		_ = p.inodeMapper.Refresh()
	}

	var allConns []*model.Connection

	// Parse TCP IPv4
	tcp4, err := p.parseFile("/proc/net/tcp", "tcp", "IPv4")
	if err == nil {
		allConns = append(allConns, tcp4...)
	}

	// Parse TCP IPv6
	tcp6, err := p.parseFile("/proc/net/tcp6", "tcp6", "IPv6")
	if err == nil {
		allConns = append(allConns, tcp6...)
	}

	// Parse UDP IPv4
	udp4, err := p.parseFile("/proc/net/udp", "udp", "IPv4")
	if err == nil {
		allConns = append(allConns, udp4...)
	}

	// Parse UDP IPv6
	udp6, err := p.parseFile("/proc/net/udp6", "udp6", "IPv6")
	if err == nil {
		allConns = append(allConns, udp6...)
	}

	return allConns, nil
}

// parseFile parses a single /proc/net/{tcp,tcp6,udp,udp6} file.
func (p *LinuxSocketParser) parseFile(filePath, proto, family string) ([]*model.Connection, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var connections []*model.Connection
	scanner := bufio.NewScanner(file)

	// Skip header line
	if scanner.Scan() {
		// header: sl local_address rem_address st ...
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 10 {
			continue
		}

		// fields[1] = local_address (ip:port)
		// fields[2] = rem_address (ip:port)
		// fields[3] = st (TCP state in hex)
		// fields[9] = inode
		localIP, localPort, err := parseHexAddress(fields[1], family)
		if err != nil {
			continue
		}

		remoteIP, remotePort, err := parseHexAddress(fields[2], family)
		if err != nil {
			continue
		}

		stateCode, err := strconv.ParseUint(fields[3], 16, 8)
		if err != nil {
			continue
		}

		inode, _ := strconv.ParseUint(fields[9], 10, 64)

		var stateStr string
		if strings.HasPrefix(proto, "udp") {
			if remoteIP.IsUnspecified() || remotePort == 0 {
				stateStr = "LISTEN"
			} else {
				stateStr = "ESTABLISHED"
			}
		} else {
			stateStr = model.TCPStates[uint8(stateCode)]
			if stateStr == "" {
				stateStr = fmt.Sprintf("STATE_%02X", stateCode)
			}
		}

		conn := &model.Connection{
			ID:          fmt.Sprintf("%s:%s:%d-%s:%d", proto, localIP.String(), localPort, remoteIP.String(), remotePort),
			Protocol:    proto,
			Family:      family,
			LocalIP:     localIP,
			LocalIPStr:  localIP.String(),
			LocalPort:   localPort,
			RemoteIP:    remoteIP,
			RemoteIPStr: remoteIP.String(),
			RemotePort:  remotePort,
			State:       stateStr,
			Inode:       inode,
			LatencyMs:   -1,
		}

		// Enrich process information from InodeMapper
		if p.inodeMapper != nil && inode > 0 {
			if meta := p.inodeMapper.Lookup(inode); meta != nil {
				conn.PID = meta.PID
				conn.ProcessName = meta.ProcessName
				conn.Cmdline = meta.Cmdline
				conn.User = meta.User
			}
		}

		connections = append(connections, conn)
	}

	return connections, scanner.Err()
}

// parseHexAddress parses an "IP:Port" hex pair from /proc/net.
func parseHexAddress(hexAddr, family string) (net.IP, int, error) {
	parts := strings.Split(hexAddr, ":")
	if len(parts) != 2 {
		return nil, 0, fmt.Errorf("invalid address format: %s", hexAddr)
	}

	port64, err := strconv.ParseUint(parts[1], 16, 16)
	if err != nil {
		return nil, 0, err
	}
	port := int(port64)

	ipHex := parts[0]
	if family == "IPv4" {
		if len(ipHex) != 8 {
			return nil, 0, fmt.Errorf("invalid IPv4 hex length: %d", len(ipHex))
		}
		val, err := strconv.ParseUint(ipHex, 16, 32)
		if err != nil {
			return nil, 0, err
		}
		ip := net.IPv4(
			byte(val),
			byte(val>>8),
			byte(val>>16),
			byte(val>>24),
		)
		return ip, port, nil
	}

	// IPv6 hex has 32 hex digits (four 32-bit words)
	if len(ipHex) != 32 {
		return nil, 0, fmt.Errorf("invalid IPv6 hex length: %d", len(ipHex))
	}

	rawBytes, err := hex.DecodeString(ipHex)
	if err != nil {
		return nil, 0, err
	}

	// Reorder 4-byte words from little-endian to big-endian
	var ip16 [16]byte
	for i := 0; i < 4; i++ {
		offset := i * 4
		ip16[offset+0] = rawBytes[offset+3]
		ip16[offset+1] = rawBytes[offset+2]
		ip16[offset+2] = rawBytes[offset+1]
		ip16[offset+3] = rawBytes[offset+0]
	}

	ip := net.IP(ip16[:])
	// If it's an IPv4-mapped IPv6 address (::ffff:x.x.x.x), convert to IPv4
	if ip4 := ip.To4(); ip4 != nil {
		return ip4, port, nil
	}

	return ip, port, nil
}
