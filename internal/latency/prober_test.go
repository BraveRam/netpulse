package latency

import (
	"net"
	"testing"
	"time"

	"github.com/plxor/netpulse/internal/model"
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
