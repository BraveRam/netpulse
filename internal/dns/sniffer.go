package dns

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/BraveRam/netpulse/internal/model"
)

type transactionKey struct {
	id                     uint16
	clientIP, serverIP     string
	clientPort, serverPort int
}

// Sniffer captures live DNS queries and answers.
type Sniffer struct {
	mu           sync.RWMutex
	ringBuffer   *RingBuffer
	ctx          context.Context
	cancel       context.CancelFunc
	privileged   bool
	statusMsg    string
	pendingTimes map[transactionKey]time.Time
	sockFD       int
}

// NewSniffer creates a Sniffer.
func NewSniffer(bufferCapacity int) *Sniffer {
	ctx, cancel := context.WithCancel(context.Background())
	return &Sniffer{
		ringBuffer:   NewRingBuffer(bufferCapacity),
		ctx:          ctx,
		cancel:       cancel,
		sockFD:       -1,
		pendingTimes: make(map[transactionKey]time.Time),
	}
}

// Start initiates the packet sniffing loop in a background goroutine.
func (s *Sniffer) Start() {
	go s.captureLoop()
}

// captureLoop opens raw socket or falls back gracefully.
func (s *Sniffer) captureLoop() {
	// htons(syscall.ETH_P_IP) = 0x0800 -> 0x0008 on little endian
	ethPIP := int(htons(syscall.ETH_P_IP))

	fd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_DGRAM, ethPIP)
	if err != nil {
		s.mu.Lock()
		s.privileged = false
		if os.IsPermission(err) || err == syscall.EPERM {
			s.statusMsg = "Unprivileged (run with sudo or CAP_NET_RAW for live DNS packet capture)"
		} else {
			s.statusMsg = fmt.Sprintf("Raw capture unavailable: %v", err)
		}
		s.mu.Unlock()
		return
	}

	s.mu.Lock()
	s.sockFD = fd
	s.privileged = true
	s.statusMsg = "Active (capturing port 53 UDP frames)"
	s.mu.Unlock()

	defer func() {
		_ = syscall.Close(fd)
	}()

	buf := make([]byte, 2048)

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
		}

		// Read packet
		n, _, err := syscall.Recvfrom(fd, buf, 0)
		if err != nil {
			select {
			case <-s.ctx.Done():
				return
			default:
				time.Sleep(10 * time.Millisecond)
				continue
			}
		}

		if n < 28 { // IP header (20) + UDP header (8) minimum
			continue
		}

		s.processPacket(buf[:n])
	}
}

func (s *Sniffer) processPacket(data []byte) {
	if len(data) < 20 {
		return
	}
	// IP Header minimum length is 20 bytes
	versionIHL := data[0]
	if (versionIHL >> 4) != 4 {
		return // IPv4 only for SOCK_DGRAM ETH_P_IP
	}

	ihl := int(versionIHL&0x0F) * 4
	if ihl < 20 || len(data) < ihl+8 {
		return
	}

	protocol := data[9]
	if protocol != 17 { // 17 = UDP
		return
	}

	srcIP := net.IP(data[12:16]).String()
	dstIP := net.IP(data[16:20]).String()

	udpHeader := data[ihl : ihl+8]
	srcPort := binary.BigEndian.Uint16(udpHeader[0:2])
	dstPort := binary.BigEndian.Uint16(udpHeader[2:4])

	if srcPort != 53 && dstPort != 53 {
		return
	}

	dnsPayload := data[ihl+8:]
	if len(dnsPayload) < 12 {
		return
	}

	query, err := ParseDNSPacket(dnsPayload)
	if err != nil || query == nil || query.Domain == "" {
		return
	}

	query.Timestamp = time.Now()
	query.ClientIP = srcIP
	query.ClientPort = int(srcPort)
	query.ServerIP = dstIP
	query.ServerPort = int(dstPort)

	if query.IsResponse {
		query.ClientIP, query.ServerIP = query.ServerIP, query.ClientIP
		query.ClientPort, query.ServerPort = query.ServerPort, query.ClientPort
	}
	key := transactionKey{query.ID, query.ClientIP, query.ServerIP, query.ClientPort, query.ServerPort}

	s.mu.Lock()
	if !query.IsResponse {
		s.pendingTimes[key] = query.Timestamp
	} else {
		if startTime, ok := s.pendingTimes[key]; ok {
			query.Latency = query.Timestamp.Sub(startTime)
			delete(s.pendingTimes, key)
		}
	}
	s.mu.Unlock()

	s.ringBuffer.Add(query)
}

// AddManualRecord manually inserts a DNS query event (e.g. for testing).
func (s *Sniffer) AddManualRecord(q *model.DNSQuery) {
	s.ringBuffer.Add(q)
}

// GetQueries returns recent DNS queries.
func (s *Sniffer) GetQueries() []*model.DNSQuery {
	return s.ringBuffer.GetAll()
}

// Status returns operational status and message.
func (s *Sniffer) Status() (bool, string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.privileged, s.statusMsg
}

// Stop terminates the packet capture loop.
func (s *Sniffer) Stop() {
	s.cancel()
	s.mu.Lock()
	if s.sockFD >= 0 {
		_ = syscall.Close(s.sockFD)
		s.sockFD = -1
	}
	s.mu.Unlock()
}

func htons(i uint16) uint16 {
	return (i<<8)&0xff00 | (i>>8)&0x00ff
}
