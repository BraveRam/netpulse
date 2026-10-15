package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/plxor/netpulse/internal/model"
	"github.com/plxor/netpulse/internal/ui/theme"
)

// RenderDetailsModal displays an in-depth telemetry inspection modal for a selected connection.
func RenderDetailsModal(styles *theme.Styles, conn *model.Connection, width, height int) string {
	if conn == nil {
		return ""
	}

	p := styles.Palette

	title := styles.ModalHeader.Render(fmt.Sprintf("🔍 CONNECTION INSPECTOR: %s", conn.RemoteEndpoint()))
	sub := styles.MutedText.Render("Detailed socket telemetry and process attribution")

	sectionStyle := lipgloss.NewStyle().Bold(true).Foreground(p.Primary)
	labelStyle := lipgloss.NewStyle().Foreground(p.TextMuted).Width(16)
	valStyle := lipgloss.NewStyle().Foreground(p.Text)

	var sb strings.Builder
	sb.WriteString(title + "\n" + sub + "\n\n")

	// 1. Process Section
	sb.WriteString(sectionStyle.Render("📦 PROCESS ATTRIBUTION") + "\n")
	sb.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Process Name:"), valStyle.Render(conn.ProcessName)))
	sb.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Process PID:"), valStyle.Render(fmt.Sprintf("%d", conn.PID))))
	sb.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("User:"), valStyle.Render(conn.User)))
	if conn.Cmdline != "" {
		cmdDisplay := conn.Cmdline
		if len(cmdDisplay) > 70 {
			cmdDisplay = cmdDisplay[:67] + "…"
		}
		sb.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Command Line:"), valStyle.Render(cmdDisplay)))
	}
	sb.WriteString("\n")

	// 2. Socket Details
	sb.WriteString(sectionStyle.Render("🔌 SOCKET METRICS") + "\n")
	sb.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Protocol:"), valStyle.Render(fmt.Sprintf("%s (%s)", conn.Protocol, conn.Family))))
	sb.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Local Address:"), valStyle.Render(conn.LocalEndpoint())))
	sb.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Remote Address:"), valStyle.Render(conn.RemoteEndpoint())))
	sb.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("TCP State:"), renderStateBadge(styles, conn.State)))
	sb.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Socket Inode:"), valStyle.Render(fmt.Sprintf("%d", conn.Inode))))
	sb.WriteString("\n")

	// 3. Geolocation & Network ASN
	sb.WriteString(sectionStyle.Render("🌍 GEOLOCATION & NETWORK ENRICHMENT") + "\n")
	countryDisplay := fmt.Sprintf("%s %s (%s)", conn.FlagEmoji, conn.CountryName, conn.CountryCode)
	sb.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Country:"), valStyle.Render(countryDisplay)))
	if conn.City != "" {
		sb.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("City:"), valStyle.Render(conn.City)))
	}
	sb.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Autonomous Sys:"), valStyle.Render(conn.ASN)))
	sb.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Organization:"), valStyle.Render(conn.ASOrg)))
	if conn.Domain != "" {
		sb.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Reverse DNS:"), valStyle.Render(conn.Domain)))
	}
	latencyDisplay := renderLatencyBadge(styles, conn.LatencyMs)
	sb.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Active RTT:"), latencyDisplay))
	sb.WriteString("\n")

	helpHint := styles.MutedText.Render("[Esc / Enter / q] Close Inspector")
	sb.WriteString(helpHint)

	modalWidth := 74
	if modalWidth > width-4 {
		modalWidth = width - 4
	}

	content := styles.ModalBox.Width(modalWidth).Render(sb.String())
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}
