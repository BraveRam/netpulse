package components

import (
	"fmt"
	"strings"

	"github.com/BraveRam/netpulse/internal/model"
	"github.com/BraveRam/netpulse/internal/ui/theme"
	"github.com/charmbracelet/bubbles/textinput"
)

// SearchBar wraps textinput.Model for live filtering.
type SearchBar struct {
	Input textinput.Model
}

// NewSearchBar creates an initialized SearchBar.
func NewSearchBar(styles *theme.Styles) *SearchBar {
	ti := textinput.New()
	ti.Placeholder = "Filter by process, PID, IP, domain, port, org, state..."
	ti.Prompt = "🔍 / "
	ti.PromptStyle = styles.SearchPrompt
	ti.CharLimit = 64
	ti.Width = 50

	return &SearchBar{
		Input: ti,
	}
}

// FilterConnections filters a slice of connections according to the query.
func FilterConnections(conns []*model.Connection, query string) []*model.Connection {
	trimmed := strings.TrimSpace(strings.ToLower(query))
	if trimmed == "" {
		return conns
	}

	var matched []*model.Connection
	for _, c := range conns {
		if connectionMatches(c, trimmed) {
			matched = append(matched, c)
		}
	}
	return matched
}

func connectionMatches(c *model.Connection, q string) bool {
	if strings.Contains(strings.ToLower(c.ProcessName), q) {
		return true
	}
	if strings.Contains(fmt.Sprintf("%d", c.PID), q) {
		return true
	}
	if strings.Contains(strings.ToLower(c.RemoteIPStr), q) {
		return true
	}
	if strings.Contains(strings.ToLower(c.Domain), q) {
		return true
	}
	if strings.Contains(fmt.Sprintf("%d", c.RemotePort), q) {
		return true
	}
	if strings.Contains(strings.ToLower(c.CountryCode), q) {
		return true
	}
	if strings.Contains(strings.ToLower(c.CountryName), q) {
		return true
	}
	if strings.Contains(strings.ToLower(c.ASN), q) {
		return true
	}
	if strings.Contains(strings.ToLower(c.ASOrg), q) {
		return true
	}
	if strings.Contains(strings.ToLower(c.State), q) {
		return true
	}
	if strings.Contains(strings.ToLower(c.Protocol), q) {
		return true
	}
	return false
}
