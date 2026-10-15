package theme

import (
	"github.com/charmbracelet/lipgloss"
)

// Palette defines the colors for a TUI theme.
type Palette struct {
	Name       string
	Primary    lipgloss.Color
	Secondary  lipgloss.Color
	Accent     lipgloss.Color
	Success    lipgloss.Color
	Warning    lipgloss.Color
	Danger     lipgloss.Color
	Info       lipgloss.Color
	Background lipgloss.Color
	Surface    lipgloss.Color
	Border     lipgloss.Color
	Text       lipgloss.Color
	TextMuted  lipgloss.Color
	Highlight  lipgloss.Color
}

// Available Themes
var (
	Cyberpunk = Palette{
		Name:       "Cyberpunk",
		Primary:    lipgloss.Color("#00f3ff"), // Electric Cyan
		Secondary:  lipgloss.Color("#ff007f"), // Neon Magenta
		Accent:     lipgloss.Color("#9d00ff"), // Electric Violet
		Success:    lipgloss.Color("#00ff66"), // Neon Green
		Warning:    lipgloss.Color("#ffe600"), // Cyber Yellow
		Danger:     lipgloss.Color("#ff3366"), // Hot Pink/Red
		Info:       lipgloss.Color("#00c8ff"), // Bright Blue
		Background: lipgloss.Color("#090a10"), // Obsidian Black
		Surface:    lipgloss.Color("#131722"), // Dark Surface
		Border:     lipgloss.Color("#2a334d"), // Border Blue
		Text:       lipgloss.Color("#f0f4fc"), // Crisp White
		TextMuted:  lipgloss.Color("#657394"), // Slate Muted
		Highlight:  lipgloss.Color("#1e293b"), // Row Selection
	}

	Catppuccin = Palette{
		Name:       "Catppuccin Mocha",
		Primary:    lipgloss.Color("#89b4fa"), // Blue
		Secondary:  lipgloss.Color("#cba6f7"), // Mauve
		Accent:     lipgloss.Color("#f5c2e7"), // Pink
		Success:    lipgloss.Color("#a6e3a1"), // Green
		Warning:    lipgloss.Color("#fab387"), // Peach
		Danger:     lipgloss.Color("#f38ba8"), // Red
		Info:       lipgloss.Color("#74c7ec"), // Sapphire
		Background: lipgloss.Color("#1e1e2e"), // Base
		Surface:    lipgloss.Color("#181825"), // Mantle
		Border:     lipgloss.Color("#45475a"), // Surface1
		Text:       lipgloss.Color("#cdd6f4"), // Text
		TextMuted:  lipgloss.Color("#6c7086"), // Overlay0
		Highlight:  lipgloss.Color("#313244"), // Surface0
	}

	TokyoNight = Palette{
		Name:       "Tokyo Night",
		Primary:    lipgloss.Color("#7dcfff"), // Cyan
		Secondary:  lipgloss.Color("#bb9af7"), // Purple
		Accent:     lipgloss.Color("#ff9e64"), // Orange
		Success:    lipgloss.Color("#9ece6a"), // Green
		Warning:    lipgloss.Color("#e0af68"), // Yellow
		Danger:     lipgloss.Color("#f7768e"), // Red
		Info:       lipgloss.Color("#7aa2f7"), // Blue
		Background: lipgloss.Color("#1a1b26"), // Storm Dark
		Surface:    lipgloss.Color("#24283b"), // Surface
		Border:     lipgloss.Color("#414868"), // Border
		Text:       lipgloss.Color("#c0caf5"), // Text
		TextMuted:  lipgloss.Color("#565f89"), // Comment
		Highlight:  lipgloss.Color("#292e42"), // Selection
	}

	Nord = Palette{
		Name:       "Nord",
		Primary:    lipgloss.Color("#88c0d0"), // Frost Cyan
		Secondary:  lipgloss.Color("#81a1c1"), // Frost Blue
		Accent:     lipgloss.Color("#b48ead"), // Aurora Purple
		Success:    lipgloss.Color("#a3be8c"), // Aurora Green
		Warning:    lipgloss.Color("#ebcb8b"), // Aurora Yellow
		Danger:     lipgloss.Color("#bf616a"), // Aurora Red
		Info:       lipgloss.Color("#5e81ac"), // Frost Deep
		Background: lipgloss.Color("#2e3440"), // Polar Night Dark
		Surface:    lipgloss.Color("#3b4252"), // Polar Night
		Border:     lipgloss.Color("#4c566a"), // Polar Night Light
		Text:       lipgloss.Color("#eceff4"), // Snow Storm
		TextMuted:  lipgloss.Color("#d8dee9"), // Snow Storm Muted
		Highlight:  lipgloss.Color("#434c5e"), // Selection
	}
)

var AllThemes = []Palette{Cyberpunk, Catppuccin, TokyoNight, Nord}
