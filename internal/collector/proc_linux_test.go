package collector

import (
	"net"
	"os"
	"path/filepath"
	"testing"
)

func TestParseHexAddressIPv4(t *testing.T) {
	tests := []struct {
		hexStr   string
		expected net.IP
		port     int
	}{
		{"0100007F:A7AB", net.IPv4(127, 0, 0, 1), 42923},
		{"00000000:0035", net.IPv4(0, 0, 0, 0), 53},
		{"08080808:0050", net.IPv4(8, 8, 8, 8), 80},
		{"0101A8C0:01BB", net.IPv4(192, 168, 1, 1), 443}, // C0 = 192, A8 = 168, 01 = 1, 01 = 1
	}

	for _, tt := range tests {
		ip, port, err := parseHexAddress(tt.hexStr, "IPv4")
		if err != nil {
			t.Fatalf("Failed to parse %s: %v", tt.hexStr, err)
		}
		if !ip.Equal(tt.expected) {
			t.Errorf("For %s, expected IP %s, got %s", tt.hexStr, tt.expected, ip)
		}
		if port != tt.port {
			t.Errorf("For %s, expected port %d, got %d", tt.hexStr, tt.port, port)
		}
	}
}

func TestBandwidthMonitor(t *testing.T) {
	mon := NewBandwidthMonitor()
	stats, err := mon.Update()
	if err != nil {
		t.Fatalf("BandwidthMonitor.Update() failed: %v", err)
	}
	if stats == nil {
		t.Fatal("Expected non-nil BandwidthStats")
	}
}

func TestUDPStates(t *testing.T) {
	for _, tt := range []struct{ proto, remote, state, want string }{
		{"udp", "00000000:0000", "07", "LISTEN"},
		{"udp", "08080808:0035", "01", "ESTABLISHED"},
		{"udp6", "00000000000000000000000000000000:0000", "07", "LISTEN"},
	} {
		t.Run(tt.proto+tt.want, func(t *testing.T) {
			local, family := "00000000:0035", "IPv4"
			if tt.proto == "udp6" {
				local, family = "00000000000000000000000000000000:0035", "IPv6"
			}
			path := filepath.Join(t.TempDir(), "sockets")
			data := "header\n0: " + local + " " + tt.remote + " " + tt.state + " 0:0 0:0 0 1000 0 123\n"
			if err := os.WriteFile(path, []byte(data), 0600); err != nil {
				t.Fatal(err)
			}
			conns, err := NewLinuxSocketParser(nil).parseFile(path, tt.proto, family)
			if err != nil {
				t.Fatal(err)
			}
			if len(conns) != 1 || conns[0].State != tt.want {
				t.Fatalf("expected %s, got %+v", tt.want, conns)
			}
		})
	}
}
