package components

import (
	"fmt"
	"strings"

	"github.com/BraveRam/netpulse/internal/model"
	"github.com/BraveRam/netpulse/internal/ui/theme"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
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

	// Reserve display cells for fixed columns, the cursor, and six separators.
	// Divide the rest without imposing minimums that would wrap terminal rows.
	colFlag, colState, colLatency, colTraffic := 8, 11, 9, 10
	remaining := width - (colFlag + colState + colLatency + colTraffic + 8)
	if remaining < 0 {
		remaining = 0
	}
	colProcess := remaining * 25 / 100
	colRemote := remaining * 45 / 100
	colOrg := remaining - colProcess - colRemote
	columns := []int{colFlag, colProcess, colRemote, colOrg, colState, colLatency, colTraffic}
	formatRow := func(prefix string, cells ...string) string {
		for i := range cells {
			cells[i] = fitTableCell(cells[i], columns[i])
		}
		return ansi.Truncate(prefix+strings.Join(cells, " "), width, "")
	}

	var sb strings.Builder

	// Header
	headerStr := formatRow("  ", "FLAG", "PROCESS", "REMOTE ENDPOINT", "ORG / ASN", "STATE", "LATENCY", "TRAFFIC")
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

		procStr := c.ProcessDisplay()
		remoteStr := c.RemoteEndpoint()

		// Org / ASN
		orgStr := c.ASOrg
		if orgStr == "" {
			orgStr = c.ASN
		}
		if orgStr == "" {
			orgStr = "-"
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

		rowContent := formatRow(cursorPrefix, flagStr, procStr, remoteStr, orgStr, stateStr, latencyStr, trafficStr)

		if isSelected {
			sb.WriteString(styles.HighlightRow.Width(width).Render(rowContent))
		} else {
			sb.WriteString(styles.NormalRow.Width(width).Render(rowContent))
		}
		sb.WriteString("\n")
	}

	return strings.TrimSuffix(sb.String(), "\n")
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

// fitTableCell preserves Unicode graphemes and ANSI styles while aligning columns.
func fitTableCell(value string, width int) string {
	if width <= 0 {
		return ""
	}
	value = strings.NewReplacer("\n", " ", "\r", " ", "\t", " ").Replace(value)
	value = ansi.Truncate(value, width, "…")
	return value + strings.Repeat(" ", width-ansi.StringWidth(value))
}
