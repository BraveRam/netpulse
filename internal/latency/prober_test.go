package latency

import (
	"net"
	"testing"
	"time"

	"github.com/BraveRam/netpulse/internal/model"
)

func TestLatencyProberLocalListener(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	defer ln.Close()

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	tcpAddr := ln.Addr().(*net.TCPAddr)
	prober := NewLatencyProber(10*time.Second, 2)

	conn := &model.Connection{
		RemoteIP:    tcpAddr.IP,
		RemoteIPStr: tcpAddr.IP.String(),
		RemotePort:  tcpAddr.Port,
		Protocol:    "tcp",
		State:       "ESTABLISHED",
	}

	// Loopback address returns 0ms immediately
	rtt := prober.Probe(conn)
	if rtt != 0 {
		t.Errorf("Expected 0ms for loopback, got %d", rtt)
	}
}

func TestIPv6ProbeUsesValidEndpoint(t *testing.T) {
	p := NewLatencyProber(time.Minute, 1)
	target := "[2001:db8::1]:443"
	// A cached successful measurement must be reused without dialing.
	p.cache[target] = &RTTRecord{LatencyMs: 42, MeasuredAt: time.Now()}
	c := &model.Connection{RemoteIP: net.ParseIP("2001:db8::1"), RemoteIPStr: "2001:db8::1", RemotePort: 443, Protocol: "tcp6", State: "ESTABLISHED"}
	if got := p.Probe(c); got != 42 {
		t.Fatalf("expected cached IPv6 RTT 42, got %d", got)
	}
}

func TestUDPNotProbedAsTCP(t *testing.T) {
	p := NewLatencyProber(time.Minute, 1)
	c := &model.Connection{RemoteIP: net.ParseIP("127.0.0.1"), RemoteIPStr: "127.0.0.1", RemotePort: 53, Protocol: "udp", State: "ESTABLISHED"}
	if got := p.Probe(c); got != -1 {
		t.Fatalf("expected no TCP latency for UDP, got %d", got)
	}
}
