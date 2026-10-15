package theme

import (
	"github.com/charmbracelet/lipgloss"
)

// Styles provides dynamic lipgloss styles derived from the active Palette.
type Styles struct {
	Palette Palette

	// Main Layout
	AppContainer lipgloss.Style
	HeaderBox    lipgloss.Style
	FooterBox    lipgloss.Style

	// Typography & Accents
	Title        lipgloss.Style
	Subtitle     lipgloss.Style
	Badge        lipgloss.Style
	MutedText    lipgloss.Style
	HighlightRow lipgloss.Style
	NormalRow    lipgloss.Style

	// Status Badges
	StateEstablished lipgloss.Style
	StateListen      lipgloss.Style
	StateTimeWait    lipgloss.Style
	StateCloseWait   lipgloss.Style
	StateOther       lipgloss.Style

	// Latency Indicators
	LatencyFast lipgloss.Style
	LatencyMid  lipgloss.Style
	LatencySlow lipgloss.Style
	LatencyNone lipgloss.Style

	// Table Headers & Borders
	TableHeader lipgloss.Style
	TableCell   lipgloss.Style
	TableBorder lipgloss.Style

	// Pane Boxes
	PaneBox       lipgloss.Style
	PaneActiveBox lipgloss.Style
	PaneTitle     lipgloss.Style

	// Dialog & Modals
	ModalBox    lipgloss.Style
	ModalHeader lipgloss.Style
	ModalBody   lipgloss.Style

	// Search Bar
	SearchBar lipgloss.Style
	SearchPrompt lipgloss.Style
}

// NewStyles generates a set of styles configured for the given Palette.
func NewStyles(p Palette) *Styles {
	return &Styles{
		Palette: p,

		AppContainer: lipgloss.NewStyle().
			Background(p.Background).
			Foreground(p.Text),

		HeaderBox: lipgloss.NewStyle().
			Background(p.Surface).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(p.Border).
			BorderBottom(true).
			Padding(0, 1),

		FooterBox: lipgloss.NewStyle().
			Background(p.Surface).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(p.Border).
			BorderTop(true).
			Padding(0, 1),

		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(p.Primary),

		Subtitle: lipgloss.NewStyle().
			Foreground(p.TextMuted),

		Badge: lipgloss.NewStyle().
			Padding(0, 1).
			Bold(true),

		MutedText: lipgloss.NewStyle().
			Foreground(p.TextMuted),

		HighlightRow: lipgloss.NewStyle().
			Background(p.Highlight).
			Foreground(p.Primary).
			Bold(true),

		NormalRow: lipgloss.NewStyle().
			Foreground(p.Text),

		// TCP States
		StateEstablished: lipgloss.NewStyle().
			Foreground(p.Success).
			Bold(true),

		StateListen: lipgloss.NewStyle().
			Foreground(p.Primary).
			Bold(true),

		StateTimeWait: lipgloss.NewStyle().
			Foreground(p.Warning),

		StateCloseWait: lipgloss.NewStyle().
			Foreground(p.Danger),

		StateOther: lipgloss.NewStyle().
			Foreground(p.TextMuted),

		// Latency
		LatencyFast: lipgloss.NewStyle().
			Foreground(p.Success),

		LatencyMid: lipgloss.NewStyle().
			Foreground(p.Warning),

		LatencySlow: lipgloss.NewStyle().
			Foreground(p.Danger),

		LatencyNone: lipgloss.NewStyle().
			Foreground(p.TextMuted),

		// Table
		TableHeader: lipgloss.NewStyle().
			Bold(true).
			Foreground(p.Secondary).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(p.Border),

		TableCell: lipgloss.NewStyle().
			Padding(0, 1),

		TableBorder: lipgloss.NewStyle().
			Foreground(p.Border),

		// Panes
		PaneBox: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(p.Border).
			Background(p.Surface).
			Padding(0, 1),

		PaneActiveBox: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(p.Primary).
			Background(p.Surface).
			Padding(0, 1),

		PaneTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(p.Primary),

		// Modal
		ModalBox: lipgloss.NewStyle().
			BorderStyle(lipgloss.DoubleBorder()).
			BorderForeground(p.Secondary).
			Background(p.Surface).
			Padding(1, 2),

		ModalHeader: lipgloss.NewStyle().
			Bold(true).
			Foreground(p.Secondary),

		ModalBody: lipgloss.NewStyle().
			Foreground(p.Text),

		// Search
		SearchBar: lipgloss.NewStyle().
			Background(p.Surface).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(p.Primary).
			Padding(0, 1),

		SearchPrompt: lipgloss.NewStyle().
			Bold(true).
			Foreground(p.Primary),
	}
}
