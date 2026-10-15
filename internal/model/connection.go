package model

import (
	"fmt"
	"net"
)

// TCPStates maps Linux TCP state codes to human-readable strings.
var TCPStates = map[uint8]string{
	0x01: "ESTABLISHED",
	0x02: "SYN_SENT",
	0x03: "SYN_RECV",
	0x04: "FIN_WAIT1",
	0x05: "FIN_WAIT2",
	0x06: "TIME_WAIT",
	0x07: "CLOSE",
	0x08: "CLOSE_WAIT",
	0x09: "LAST_ACK",
	0x0A: "LISTEN",
	0x0B: "CLOSING",
	0x0C: "NEW_SYN_RECV",
}

// Connection represents a live network socket / connection.
type Connection struct {
	ID          string // proto:localIP:port-remoteIP:port
	Protocol    string // "tcp", "tcp6", "udp", "udp6"
	Family      string // "IPv4", "IPv6"
	LocalIP     net.IP
	LocalIPStr  string
	LocalPort   int
	RemoteIP    net.IP
	RemoteIPStr string
	RemotePort  int
	State       string // e.g. "ESTABLISHED", "LISTEN"

	// Process Information
	PID         int
	ProcessName string
	Cmdline     string
	User        string
	Inode       uint64

	// GeoIP & Metadata Enrichment
	CountryCode string // "US", "DE", "JP", "LOCAL"
	CountryName string // "United States", "Local Network"
	FlagEmoji   string // "🇺🇸", "📍"
	City        string
	Domain      string // Reverse DNS resolved host
	ASN         string // e.g. "AS15169"
	ASOrg       string // e.g. "Google LLC"

	// Latency / RTT Probe
	LatencyMs int // RTT in milliseconds (-1 if not measured/applicable)

	// Traffic Metrics
	TxBytes uint64
	RxBytes uint64
	RateTx  float64 // Bytes per second
	RateRx  float64 // Bytes per second
}

// EndpointString returns a formatted "host:port" or "ip:port" representation.
func (c *Connection) RemoteEndpoint() string {
	if c.State == "LISTEN" {
		return "*:*"
	}
	host := c.Domain
	if host == "" {
		host = c.RemoteIPStr
	}
	if host == "" || host == "0.0.0.0" || host == "::" {
		return "*:*"
	}
	return fmt.Sprintf("%s:%d", host, c.RemotePort)
}

// LocalEndpoint returns formatted local address.
func (c *Connection) LocalEndpoint() string {
	return fmt.Sprintf("%s:%d", c.LocalIPStr, c.LocalPort)
}

// ProcessDisplay returns a formatted process name with PID, e.g. "curl [4102]".
func (c *Connection) ProcessDisplay() string {
	if c.PID <= 0 {
		return "- [?]"
	}
	name := c.ProcessName
	if name == "" {
		name = "unknown"
	}
	return fmt.Sprintf("%s [%d]", name, c.PID)
}
