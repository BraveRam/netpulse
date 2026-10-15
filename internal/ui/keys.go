package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/plxor/netpulse/internal/ui/theme"
)

// KeyMap defines the keybindings for netpulse.
type KeyMap struct {
	Up          key.Binding
	Down        key.Binding
	PageUp      key.Binding
	PageDown    key.Binding
	Home        key.Binding
	End         key.Binding
	Inspect     key.Binding
	Filter      key.Binding
	ToggleDNS   key.Binding
	Sort        key.Binding
	Theme       key.Binding
	Pause       key.Binding
	Kill        key.Binding
	Cancel      key.Binding
	Quit        key.Binding
}

// DefaultKeyMap returns the standard key bindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup"),
			key.WithHelp("pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown"),
			key.WithHelp("pgdn", "page down"),
		),
		Home: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("g/home", "top"),
		),
		End: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("G/end", "bottom"),
		),
		Inspect: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "inspect"),
		),
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		ToggleDNS: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "dns sniffer"),
		),
		Sort: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "sort mode"),
		),
		Theme: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "theme"),
		),
		Pause: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "pause"),
		),
		Kill: key.NewBinding(
			key.WithKeys("k"),
			key.WithHelp("k", "kill proc"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back/cancel"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

// RenderFooter builds the bottom navigation and status line.
func RenderFooter(styles *theme.Styles, statusMsg string, sortName string, width int) string {
	keyStyle := styles.Title
	descStyle := styles.MutedText
	sep := styles.MutedText.Render(" • ")

	items := []string{
		fmt.Sprintf("%s %s", keyStyle.Render("[↑/↓]"), descStyle.Render("Nav")),
		fmt.Sprintf("%s %s", keyStyle.Render("[Enter]"), descStyle.Render("Inspect")),
		fmt.Sprintf("%s %s", keyStyle.Render("[/]"), descStyle.Render("Filter")),
		fmt.Sprintf("%s %s", keyStyle.Render("[d]"), descStyle.Render("DNS")),
		fmt.Sprintf("%s %s", keyStyle.Render("[s]"), descStyle.Render(fmt.Sprintf("Sort (%s)", sortName))),
		fmt.Sprintf("%s %s", keyStyle.Render("[k]"), descStyle.Render("Kill")),
		fmt.Sprintf("%s %s", keyStyle.Render("[Space]"), descStyle.Render("Pause")),
		fmt.Sprintf("%s %s", keyStyle.Render("[t]"), descStyle.Render("Theme")),
		fmt.Sprintf("%s %s", keyStyle.Render("[q]"), descStyle.Render("Quit")),
	}

	left := strings.Join(items, sep)

	right := ""
	if statusMsg != "" {
		right = styles.StateTimeWait.Render(statusMsg)
	}

	gap := width - lipgloss.Width(left) - lipgloss.Width(right) - 4
	if gap < 1 {
		gap = 1
	}
	spacer := strings.Repeat(" ", gap)

	content := fmt.Sprintf("%s%s%s", left, spacer, right)
	return styles.FooterBox.Width(width).Render(content)
}
