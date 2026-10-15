package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/plxor/netpulse/internal/model"
	"github.com/plxor/netpulse/internal/ui/theme"
)

// RenderDNSPane renders the Live DNS Query Sniffer bottom pane.
func RenderDNSPane(styles *theme.Styles, queries []*model.DNSQuery, privileged bool, statusMsg string, width, height int) string {
	p := styles.Palette

	title := styles.PaneTitle.Render("📡 LIVE DNS QUERY SNIFFER")
	statusBadge := ""
	if privileged {
		statusBadge = lipgloss.NewStyle().Foreground(p.Success).Render("● Active")
	} else {
		statusBadge = lipgloss.NewStyle().Foreground(p.Warning).Render("⚠ Limited")
	}

	headerLine := fmt.Sprintf("%s  %s", title, statusBadge)
	if statusMsg != "" && !privileged {
		headerLine += fmt.Sprintf("  %s", styles.MutedText.Render(fmt.Sprintf("(%s)", statusMsg)))
	}

	var sb strings.Builder
	sb.WriteString(headerLine)
	sb.WriteString("\n")

	// Table header for DNS
	// TIME (8), CLIENT (15), TYPE (5), DOMAIN (30), STATUS (8), LATENCY (8), ANSWERS (remaining)
	dnsHeader := fmt.Sprintf(
		"%-8s %-15s %-6s %-32s %-10s %-8s %s",
		"TIME", "CLIENT", "TYPE", "QUERY DOMAIN", "STATUS", "RTT", "ANSWERS",
	)
	sb.WriteString(styles.TableHeader.Render(dnsHeader))
	sb.WriteString("\n")

	rowsAvailable := height - 4
	if rowsAvailable < 1 {
		rowsAvailable = 1
	}

	if len(queries) == 0 {
		emptyMsg := styles.MutedText.Render("  No DNS queries captured yet. Listening on port 53...")
		sb.WriteString(emptyMsg)
		sb.WriteString("\n")
	} else {
		limit := len(queries)
		if limit > rowsAvailable {
			limit = rowsAvailable
		}

		for i := 0; i < limit; i++ {
			q := queries[i]
			timeStr := q.Timestamp.Format("15:04:05")
			clientStr := fmt.Sprintf("%s:%d", q.ClientIP, q.ClientPort)
			if len(clientStr) > 15 {
				clientStr = clientStr[:14] + "…"
			}

			domainStr := q.Domain
			if len(domainStr) > 32 {
				domainStr = domainStr[:31] + "…"
			}

			statusStr := q.RCode
			if statusStr == "NOERROR" {
				statusStr = styles.LatencyFast.Render("NOERROR")
			} else if statusStr != "" {
				statusStr = styles.LatencySlow.Render(statusStr)
			} else {
				statusStr = "-"
			}

			latencyStr := "-"
			if q.Latency > 0 {
				latencyStr = fmt.Sprintf("%dms", q.Latency.Milliseconds())
			}

			answersStr := q.AnswerSummary()

			row := fmt.Sprintf(
				"%-8s %-15s %-6s %-32s %-10s %-8s %s",
				timeStr, clientStr, q.Type, domainStr, statusStr, latencyStr, answersStr,
			)
			sb.WriteString(styles.NormalRow.Render(row))
			sb.WriteString("\n")
		}
	}

	return styles.PaneBox.Width(width).Height(height).Render(sb.String())
}
