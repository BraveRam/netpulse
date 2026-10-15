package components

import (
	"fmt"
	"os"

	"github.com/BraveRam/netpulse/internal/model"
	"github.com/BraveRam/netpulse/internal/ui/theme"
	"github.com/charmbracelet/lipgloss"
)

// RenderHeader builds the top stats banner for netpulse.
func RenderHeader(styles *theme.Styles, stats *model.BandwidthStats, width int, paused bool, themeName string) string {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "localhost"
	}

	p := styles.Palette

	// Brand / Title
	pulseIcon := styles.Title.Render("⚡ NETPULSE")
	hostBadge := styles.MutedText.Render(fmt.Sprintf("@%s", hostname))
	titleGroup := fmt.Sprintf("%s %s", pulseIcon, hostBadge)

	// Sockets summary badges
	var total, estab, listen, wait, udp int
	if stats != nil {
		total = stats.TotalConnections
		estab = stats.ActiveEstablished
		listen = stats.ActiveListen
		udp = stats.ActiveUDP
		wait = stats.TotalConnections - (estab + listen + udp)
		if wait < 0 {
			wait = 0
		}
	}

	badgeStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Bold(true)

	totalBadge := badgeStyle.Background(p.Surface).Foreground(p.Text).Render(fmt.Sprintf("TOTAL %d", total))
	estabBadge := badgeStyle.Background(p.Surface).Foreground(p.Success).Render(fmt.Sprintf("ESTAB %d", estab))
	listenBadge := badgeStyle.Background(p.Surface).Foreground(p.Primary).Render(fmt.Sprintf("LISTEN %d", listen))
	udpBadge := badgeStyle.Background(p.Surface).Foreground(p.Secondary).Render(fmt.Sprintf("UDP %d", udp))
	statsGroup := fmt.Sprintf("%s %s %s %s", totalBadge, estabBadge, listenBadge, udpBadge)

	// Bandwidth sparklines
	var rxRateStr, txRateStr, rxSpark, txSpark string
	if stats != nil {
		rxRateStr = model.FormatSpeed(stats.RateRx)
		txRateStr = model.FormatSpeed(stats.RateTx)
		rxSpark = RenderSparkline(stats.RxHistory, 10)
		txSpark = RenderSparkline(stats.TxHistory, 10)
	} else {
		rxRateStr = "0 B/s"
		txRateStr = "0 B/s"
	}

	rxBadge := lipgloss.NewStyle().Foreground(p.Primary).Render(fmt.Sprintf("▼ %s %s", rxRateStr, rxSpark))
	txBadge := lipgloss.NewStyle().Foreground(p.Secondary).Render(fmt.Sprintf("▲ %s %s", txRateStr, txSpark))
	bandwidthGroup := fmt.Sprintf("%s   %s", rxBadge, txBadge)

	// Mode indicator
	modeBadge := ""
	if paused {
		modeBadge = badgeStyle.Background(p.Danger).Foreground(p.Background).Render("PAUSED")
	} else {
		modeBadge = badgeStyle.Background(p.Highlight).Foreground(p.Success).Render("LIVE")
	}

	themeBadge := styles.MutedText.Render(fmt.Sprintf("[%s]", themeName))
	statusGroup := fmt.Sprintf("%s %s", modeBadge, themeBadge)

	// Layout line 1: Title & Host on left, Mode & Theme on right
	line1Left := titleGroup
	line1Right := statusGroup
	gap1 := width - lipgloss.Width(line1Left) - lipgloss.Width(line1Right) - 2
	if gap1 < 1 {
		gap1 = 1
	}
	spacer1 := lipgloss.NewStyle().Width(gap1).Render("")
	row1 := fmt.Sprintf("%s%s%s", line1Left, spacer1, line1Right)

	// Layout line 2: Sockets on left, Bandwidth on right
	line2Left := statsGroup
	line2Right := bandwidthGroup
	gap2 := width - lipgloss.Width(line2Left) - lipgloss.Width(line2Right) - 2
	if gap2 < 1 {
		gap2 = 1
	}
	spacer2 := lipgloss.NewStyle().Width(gap2).Render("")
	row2 := fmt.Sprintf("%s%s%s", line2Left, spacer2, line2Right)

	content := fmt.Sprintf("%s\n%s", row1, row2)
	return styles.HeaderBox.Width(width).Render(content)
}
