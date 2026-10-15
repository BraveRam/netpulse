package components

import (
	"net"
	"testing"

	"github.com/plxor/netpulse/internal/model"
)

func TestRenderSparkline(t *testing.T) {
	// Empty slice
	if res := RenderSparkline(nil, 10); res != "" {
		t.Errorf("Expected empty sparkline for nil data, got %q", res)
	}

	// Uniform values
	dataUniform := []float64{100, 100, 100}
	resUniform := RenderSparkline(dataUniform, 10)
	if len([]rune(resUniform)) != 3 {
		t.Errorf("Expected 3 runes, got %d", len([]rune(resUniform)))
	}

	// Dynamic values
	dataDynamic := []float64{10, 20, 50, 100, 200}
	resDynamic := RenderSparkline(dataDynamic, 5)
	runes := []rune(resDynamic)
	if len(runes) != 5 {
		t.Errorf("Expected 5 runes, got %d", len(runes))
	}
	// First rune should be smaller than last rune
	if runes[0] >= runes[len(runes)-1] {
		t.Errorf("Expected ascending sparkline runes, got %s", resDynamic)
	}
}

func TestFilterConnections(t *testing.T) {
	conns := []*model.Connection{
		{
			PID:         101,
			ProcessName: "curl",
			RemoteIPStr: "1.1.1.1",
			RemotePort:  443,
			State:       "ESTABLISHED",
			ASOrg:       "Cloudflare, Inc.",
			CountryCode: "US",
		},
		{
			PID:         202,
			ProcessName: "spotify",
			RemoteIPStr: "35.186.224.25",
			RemotePort:  443,
			State:       "ESTABLISHED",
			ASOrg:       "Google LLC",
			CountryCode: "US",
		},
		{
			PID:         303,
			ProcessName: "sshd",
			RemoteIPStr: "0.0.0.0",
			RemotePort:  22,
			State:       "LISTEN",
			CountryCode: "LOCAL",
		},
	}

	// Filter by process name
	filtered := FilterConnections(conns, "curl")
	if len(filtered) != 1 || filtered[0].ProcessName != "curl" {
		t.Fatalf("Expected 1 curl connection, got %d", len(filtered))
	}

	// Filter by state
	filteredListen := FilterConnections(conns, "LISTEN")
	if len(filteredListen) != 1 || filteredListen[0].ProcessName != "sshd" {
		t.Fatalf("Expected 1 sshd listen connection, got %d", len(filteredListen))
	}

	// Filter by org
	filteredOrg := FilterConnections(conns, "Cloudflare")
	if len(filteredOrg) != 1 || filteredOrg[0].PID != 101 {
		t.Fatalf("Expected Cloudflare connection, got %d", len(filteredOrg))
	}

	// Filter with empty string (returns all)
	if all := FilterConnections(conns, ""); len(all) != 3 {
		t.Fatalf("Expected all 3 connections with empty query, got %d", len(all))
	}
}

func TestTableViewNavigation(t *testing.T) {
	tv := NewTableView()
	if tv.Cursor != 0 {
		t.Errorf("Expected initial cursor 0, got %d", tv.Cursor)
	}

	tv.MoveDown(5)
	if tv.Cursor != 1 {
		t.Errorf("Expected cursor 1 after MoveDown, got %d", tv.Cursor)
	}

	tv.MoveUp()
	if tv.Cursor != 0 {
		t.Errorf("Expected cursor 0 after MoveUp, got %d", tv.Cursor)
	}

	tv.MoveUp() // should clamp at 0
	if tv.Cursor != 0 {
		t.Errorf("Expected cursor to remain 0, got %d", tv.Cursor)
	}
}

func TestModelFormatters(t *testing.T) {
	c := &model.Connection{
		PID:         4102,
		ProcessName: "curl",
		RemoteIP:    net.ParseIP("1.1.1.1"),
		RemoteIPStr: "1.1.1.1",
		RemotePort:  443,
		State:       "ESTABLISHED",
	}

	if c.ProcessDisplay() != "curl [4102]" {
		t.Errorf("Unexpected process display: %s", c.ProcessDisplay())
	}
	if c.RemoteEndpoint() != "1.1.1.1:443" {
		t.Errorf("Unexpected remote endpoint: %s", c.RemoteEndpoint())
	}

	c.Domain = "one.one.one.one"
	if c.RemoteEndpoint() != "one.one.one.one:443" {
		t.Errorf("Unexpected remote endpoint with domain: %s", c.RemoteEndpoint())
	}

	if speed := model.FormatSpeed(1500000); speed != "1.43 MB/s" {
		t.Errorf("FormatSpeed unexpected: %s", speed)
	}
	if bytes := model.FormatBytes(1073741824); bytes != "1.00 GB" {
		t.Errorf("FormatBytes unexpected: %s", bytes)
	}
}
