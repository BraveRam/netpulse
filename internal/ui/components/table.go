package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/plxor/netpulse/internal/model"
	"github.com/plxor/netpulse/internal/ui/theme"
)

// TableView renders the primary Connections Radar table with scrolling and highlighting.
type TableView struct {
	Cursor int
	Offset int
}

// NewTableView creates a TableView.
func NewTableView() *TableView {
	return &TableView{
		Cursor: 0,
		Offset: 0,
	}
}

// MoveUp moves cursor up by one row.
func (t *TableView) MoveUp() {
	if t.Cursor > 0 {
		t.Cursor--
	}
}

// MoveDown moves cursor down by one row within max limit.
func (t *TableView) MoveDown(max int) {
	if max <= 0 {
		t.Cursor = 0
		return
	}
	if t.Cursor < max-1 {
		t.Cursor++
	}
}

// PageUp scrolls up by one page height.
func (t *TableView) PageUp(pageSize int) {
	t.Cursor -= pageSize
	if t.Cursor < 0 {
		t.Cursor = 0
	}
}

// PageDown scrolls down by one page height.
func (t *TableView) PageDown(pageSize, max int) {
	t.Cursor += pageSize
	if t.Cursor >= max {
		t.Cursor = max - 1
	}
	if t.Cursor < 0 {
		t.Cursor = 0
	}
}

// Render draws the table rows and header.
func (t *TableView) Render(styles *theme.Styles, conns []*model.Connection, width, height int) string {
	if len(conns) == 0 {
		emptyMsg := styles.MutedText.Render("No network connections matching filter.")
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, emptyMsg)
	}

	// Clamp cursor
	if t.Cursor >= len(conns) {
		t.Cursor = len(conns) - 1
	}
	if t.Cursor < 0 {
		t.Cursor = 0
	}

	// Calculate visible rows
	headerHeight := 2 // header row + border
	visibleRows := height - headerHeight
	if visibleRows < 1 {
		visibleRows = 1
	}

	// Adjust offset for scrolling
	if t.Cursor < t.Offset {
		t.Offset = t.Cursor
	} else if t.Cursor >= t.Offset+visibleRows {
		t.Offset = t.Cursor - visibleRows + 1
	}

	// Calculate column widths dynamically based on terminal width
	// Base widths: Flag: 9, Process: 18, Remote: 26, Org: 22, State: 13, Latency: 10, Traffic: 12
	colFlag := 9
	colState := 13
	colLatency := 10
	colTraffic := 12

	remaining := width - (colFlag + colState + colLatency + colTraffic + 6) // margins
	if remaining < 40 {
		remaining = 40
	}
	colProcess := remaining * 25 / 100
	if colProcess < 14 {
		colProcess = 14
	}
	colRemote := remaining * 45 / 100
	if colRemote < 20 {
		colRemote = 20
	}
	colOrg := remaining - colProcess - colRemote
	if colOrg < 14 {
		colOrg = 14
	}

	var sb strings.Builder

	// Header
	headerStr := fmt.Sprintf(
		"  %-*s %-*s %-*s %-*s %-*s %-*s %-*s",
		colFlag, "FLAG",
		colProcess, "PROCESS",
		colRemote, "REMOTE ENDPOINT",
		colOrg, "ORG / ASN",
		colState, "STATE",
		colLatency, "LATENCY",
		colTraffic, "TRAFFIC",
	)
	sb.WriteString(styles.TableHeader.Width(width).Render(headerStr))
	sb.WriteString("\n")

	// Render visible rows
	end := t.Offset + visibleRows
	if end > len(conns) {
		end = len(conns)
	}

	for i := t.Offset; i < end; i++ {
		c := conns[i]
		isSelected := (i == t.Cursor)

		// Flag & Country
		flagStr := fmt.Sprintf("%s %-2s", c.FlagEmoji, c.CountryCode)
		if len(flagStr) > colFlag {
			flagStr = flagStr[:colFlag]
		}

		// Process
		procStr := c.ProcessDisplay()
		if len(procStr) > colProcess {
			procStr = procStr[:colProcess-1] + "…"
		}

		// Remote Endpoint
		remoteStr := c.RemoteEndpoint()
		if len(remoteStr) > colRemote {
			remoteStr = remoteStr[:colRemote-1] + "…"
		}

		// Org / ASN
		orgStr := c.ASOrg
		if orgStr == "" {
			orgStr = c.ASN
		}
		if orgStr == "" {
			orgStr = "-"
		}
		if len(orgStr) > colOrg {
			orgStr = orgStr[:colOrg-1] + "…"
		}

		// State
		stateStr := renderStateBadge(styles, c.State)

		// Latency
		latencyStr := renderLatencyBadge(styles, c.LatencyMs)

		// Traffic Rate
		trafficStr := "-"
		if c.RateRx > 0 || c.RateTx > 0 {
			trafficStr = model.FormatSpeed(c.RateRx + c.RateTx)
		}

		cursorPrefix := "  "
		if isSelected {
			cursorPrefix = "❯ "
		}

		rowContent := fmt.Sprintf(
			"%s%-*s %-*s %-*s %-*s %-*s %-*s %-*s",
			cursorPrefix,
			colFlag, flagStr,
			colProcess, procStr,
			colRemote, remoteStr,
			colOrg, orgStr,
			colState, stateStr,
			colLatency, latencyStr,
			colTraffic, trafficStr,
		)

		if isSelected {
			sb.WriteString(styles.HighlightRow.Width(width).Render(rowContent))
		} else {
			sb.WriteString(styles.NormalRow.Width(width).Render(rowContent))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func renderStateBadge(styles *theme.Styles, state string) string {
	switch state {
	case "ESTABLISHED":
		return styles.StateEstablished.Render("ESTABLISHED")
	case "LISTEN":
		return styles.StateListen.Render("LISTEN")
	case "TIME_WAIT":
		return styles.StateTimeWait.Render("TIME_WAIT")
	case "CLOSE_WAIT":
		return styles.StateCloseWait.Render("CLOSE_WAIT")
	default:
		return styles.StateOther.Render(state)
	}
}

func renderLatencyBadge(styles *theme.Styles, ms int) string {
	if ms < 0 {
		return styles.LatencyNone.Render("● -")
	}
	if ms == 0 {
		return styles.LatencyFast.Render("● <1ms")
	}
	str := fmt.Sprintf("● %dms", ms)
	if ms < 60 {
		return styles.LatencyFast.Render(str)
	}
	if ms < 180 {
		return styles.LatencyMid.Render(str)
	}
	return styles.LatencySlow.Render(str)
}
