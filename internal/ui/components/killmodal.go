package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/plxor/netpulse/internal/model"
	"github.com/plxor/netpulse/internal/ui/theme"
)

// RenderKillModal renders the process termination confirmation modal.
func RenderKillModal(styles *theme.Styles, conn *model.Connection, width, height int) string {
	if conn == nil {
		return ""
	}

	p := styles.Palette

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(p.Danger).
		Render("⚠️  TERMINATE BACKGROUND PROCESS / SOCKET")

	warnText := lipgloss.NewStyle().
		Foreground(p.Warning).
		Render("You are about to send a termination signal to this process.")

	var sb strings.Builder
	sb.WriteString(header + "\n" + warnText + "\n\n")

	labelStyle := lipgloss.NewStyle().Foreground(p.TextMuted).Width(16)
	valStyle := lipgloss.NewStyle().Foreground(p.Text).Bold(true)

	sb.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Target Process:"), valStyle.Render(conn.ProcessName)))
	sb.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Process PID:"), valStyle.Render(fmt.Sprintf("%d", conn.PID))))
	sb.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Owner User:"), valStyle.Render(conn.User)))
	sb.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Active Socket:"), valStyle.Render(conn.RemoteEndpoint())))
	sb.WriteString("\n")

	promptStyle := lipgloss.NewStyle().
		Background(p.Highlight).
		Foreground(p.Text).
		Padding(0, 1)

	optSigterm := lipgloss.NewStyle().Foreground(p.Warning).Bold(true).Render("[y] Terminate (SIGTERM)")
	optSigkill := lipgloss.NewStyle().Foreground(p.Danger).Bold(true).Render("[Y] Force Kill (SIGKILL)")
	optCancel := lipgloss.NewStyle().Foreground(p.TextMuted).Render("[n / Esc] Cancel")

	actions := fmt.Sprintf("%s   %s   %s", optSigterm, optSigkill, optCancel)
	sb.WriteString(promptStyle.Render(actions))

	modalWidth := 68
	if modalWidth > width-4 {
		modalWidth = width - 4
	}

	modalBox := lipgloss.NewStyle().
		BorderStyle(lipgloss.DoubleBorder()).
		BorderForeground(p.Danger).
		Background(p.Surface).
		Padding(1, 2)

	content := modalBox.Width(modalWidth).Render(sb.String())
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}
