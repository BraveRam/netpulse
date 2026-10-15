package components

import (
	"fmt"
	"github.com/BraveRam/netpulse/internal/ui/theme"
	"github.com/charmbracelet/x/ansi"
	"net"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/BraveRam/netpulse/internal/model"
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

func TestTableLabelsAndColumnAlignment(t *testing.T) {
	styles := theme.NewStyles(theme.AllThemes[0])
	conns := []*model.Connection{
		{PID: 1234, ProcessName: "éditeur日本語", FlagEmoji: "🇺🇸", CountryCode: "US", RemoteIPStr: "1.1.1.1", RemotePort: 443, State: "ESTABLISHED", ASOrg: "Société réseau 日本語", LatencyMs: 42},
		{PID: 1235, ProcessName: "local", FlagEmoji: "📍", CountryCode: "LOCAL", State: "LISTEN", LatencyMs: -1},
	}
	for _, width := range []int{80, 120, 190} {
		t.Run(fmt.Sprint(width), func(t *testing.T) {
			view := NewTableView().Render(styles, conns, width, 6)
			if !utf8.ValidString(view) {
				t.Fatal("invalid UTF-8 in table")
			}
			plain := ansi.Strip(view)
			if !strings.Contains(plain, "🇺🇸 US") || !strings.Contains(plain, "📍 LOCAL") {
				t.Fatalf("country labels truncated:\n%s", plain)
			}
			lines := strings.Split(plain, "\n")
			if len(lines) > 6 {
				t.Fatalf("table exceeded height: %d", len(lines))
			}
			for _, line := range lines {
				if got := ansi.StringWidth(line); got != width {
					t.Fatalf("line width %d, want %d: %q", got, width, line)
				}
			}
			stateCol := ansi.StringWidth(lines[0][:strings.Index(lines[0], "STATE")])
			for i, state := range []string{"ESTABLISHED", "LISTEN"} {
				line := lines[i+2]
				if got := ansi.StringWidth(line[:strings.Index(line, state)]); got != stateCol {
					t.Fatalf("state column %d, want %d", got, stateCol)
				}
			}
		})
	}
}

func TestTableCellANSIAndUnicodeWidth(t *testing.T) {
	value := "\x1b[32m🇪🇹 日本語 société\x1b[0m"
	got := fitTableCell(value, 9)
	if !utf8.ValidString(got) || ansi.StringWidth(got) != 9 {
		t.Fatalf("invalid cell: %q", got)
	}
	if !strings.HasSuffix(strings.TrimSpace(ansi.Strip(got)), "…") {
		t.Fatalf("missing truncation marker: %q", got)
	}
}

func TestTableScrollingFitsViewport(t *testing.T) {
	conns := make([]*model.Connection, 30)
	for i := range conns {
		conns[i] = &model.Connection{PID: i + 1, ProcessName: "process", State: "ESTABLISHED", LatencyMs: -1}
	}
	table := NewTableView()
	table.Cursor = len(conns) - 1
	view := table.Render(theme.NewStyles(theme.AllThemes[0]), conns, 80, 12)
	lines := strings.Split(ansi.Strip(view), "\n")
	if len(lines) != 12 {
		t.Fatalf("table height %d, want 12", len(lines))
	}
	if !strings.Contains(lines[len(lines)-1], "❯") {
		t.Fatal("selected row is outside viewport")
	}
}

func TestHeaderSummaryIncludesOther(t *testing.T) {
	stats := &model.BandwidthStats{TotalConnections: 7, ActiveEstablished: 1, ActiveListen: 1, ActiveUDP: 3}
	view := ansi.Strip(RenderHeader(theme.NewStyles(theme.AllThemes[0]), stats, 190, false, "Cyberpunk"))
	for _, label := range []string{"TOTAL 7", "ESTAB 1", "LISTEN 1", "UDP 3", "OTHER 2"} {
		if !strings.Contains(view, label) {
			t.Fatalf("missing %q in header", label)
		}
	}
}
