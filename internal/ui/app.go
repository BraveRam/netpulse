package ui

import (
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/plxor/netpulse/internal/collector"
	"github.com/plxor/netpulse/internal/dns"
	"github.com/plxor/netpulse/internal/enricher"
	"github.com/plxor/netpulse/internal/killer"
	"github.com/plxor/netpulse/internal/latency"
	"github.com/plxor/netpulse/internal/model"
	"github.com/plxor/netpulse/internal/ui/components"
	"github.com/plxor/netpulse/internal/ui/theme"
)

type ViewMode int

const (
	ViewNormal ViewMode = iota
	ViewFilter
	ViewDetails
	ViewKill
)

type SortMode int

const (
	SortByState SortMode = iota
	SortByProcess
	SortByRemote
	SortByLatency
	SortByCountry
)

var sortModeNames = []string{
	"State",
	"Process",
	"Remote",
	"Latency",
	"Country",
}

// TickMsg signals the UI update loop to collect fresh network metrics.
type TickMsg time.Time

// AppModel is the primary Bubble Tea model managing netpulse TUI state.
type AppModel struct {
	mode          ViewMode
	collector     collector.Collector
	enricher      *enricher.Enricher
	prober        *latency.LatencyProber
	sniffer       *dns.Sniffer
	killer        *killer.Killer
	themeIndex    int
	styles        *theme.Styles
	keys          KeyMap
	table         *components.TableView
	search        *components.SearchBar
	sortMode      SortMode
	dnsPaneOpen   bool
	paused        bool
	connections   []*model.Connection
	filteredConns []*model.Connection
	stats         *model.BandwidthStats
	width         int
	height        int
	tickInterval  time.Duration
	statusMsg     string
	statusExpiry  time.Time
}

// Config configures the AppModel instance.
type Config struct {
	TickInterval time.Duration
	GeoIPPath    string
	EnableDNS    bool
	ThemeIndex   int
}

// NewAppModel creates and initializes the netpulse AppModel.
func NewAppModel(cfg Config) *AppModel {
	if cfg.TickInterval <= 0 {
		cfg.TickInterval = 1 * time.Second
	}
	if cfg.ThemeIndex < 0 || cfg.ThemeIndex >= len(theme.AllThemes) {
		cfg.ThemeIndex = 0
	}

	activeTheme := theme.AllThemes[cfg.ThemeIndex]
	styles := theme.NewStyles(activeTheme)

	sniff := dns.NewSniffer(250)
	if cfg.EnableDNS {
		sniff.Start()
	}

	m := &AppModel{
		mode:         ViewNormal,
		collector:    collector.NewSystemCollector(1500 * time.Millisecond),
		enricher:     enricher.NewEnricher(cfg.GeoIPPath),
		prober:       latency.NewLatencyProber(20*time.Second, 6),
		sniffer:      sniff,
		killer:       killer.NewKiller(),
		themeIndex:   cfg.ThemeIndex,
		styles:       styles,
		keys:         DefaultKeyMap(),
		table:        components.NewTableView(),
		search:       components.NewSearchBar(styles),
		sortMode:     SortByState,
		dnsPaneOpen:  false,
		paused:       false,
		tickInterval: cfg.TickInterval,
	}

	return m
}

func (m *AppModel) Init() tea.Cmd {
	return tea.Batch(
		tea.Tick(m.tickInterval, func(t time.Time) tea.Msg {
			return TickMsg(t)
		}),
	)
}

func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case TickMsg:
		if !m.paused {
			m.refreshData()
		}
		// Schedule next tick
		cmds = append(cmds, tea.Tick(m.tickInterval, func(t time.Time) tea.Msg {
			return TickMsg(t)
		}))

	case tea.KeyMsg:
		// Clear expired status messages
		if !m.statusExpiry.IsZero() && time.Now().After(m.statusExpiry) {
			m.statusMsg = ""
		}

		switch m.mode {
		case ViewFilter:
			switch {
			case key.Matches(msg, m.keys.Cancel):
				m.mode = ViewNormal
				m.search.Input.Blur()
			case key.Matches(msg, m.keys.Inspect): // Enter
				m.mode = ViewNormal
				m.search.Input.Blur()
			default:
				m.search.Input, cmd = m.search.Input.Update(msg)
				cmds = append(cmds, cmd)
				m.applyFilterAndSort()
			}

		case ViewDetails:
			switch {
			case key.Matches(msg, m.keys.Cancel), key.Matches(msg, m.keys.Inspect), key.Matches(msg, m.keys.Quit):
				m.mode = ViewNormal
			}

		case ViewKill:
			selected := m.selectedConnection()
			switch msg.String() {
			case "y":
				if selected != nil {
					res := m.killer.KillProcess(selected.PID, selected.ProcessName, false)
					m.showStatus(res.Message, 3*time.Second)
				}
				m.mode = ViewNormal
			case "Y":
				if selected != nil {
					res := m.killer.KillProcess(selected.PID, selected.ProcessName, true)
					m.showStatus(res.Message, 3*time.Second)
				}
				m.mode = ViewNormal
			case "n", "esc":
				m.mode = ViewNormal
			}

		case ViewNormal:
			switch {
			case key.Matches(msg, m.keys.Quit):
				m.enricher.Close()
				m.sniffer.Stop()
				return m, tea.Quit

			case key.Matches(msg, m.keys.Up):
				m.table.MoveUp()

			case key.Matches(msg, m.keys.Down):
				m.table.MoveDown(len(m.filteredConns))

			case key.Matches(msg, m.keys.PageUp):
				m.table.PageUp(10)

			case key.Matches(msg, m.keys.PageDown):
				m.table.PageDown(10, len(m.filteredConns))

			case key.Matches(msg, m.keys.Home):
				m.table.Cursor = 0

			case key.Matches(msg, m.keys.End):
				if len(m.filteredConns) > 0 {
					m.table.Cursor = len(m.filteredConns) - 1
				}

			case key.Matches(msg, m.keys.Inspect):
				if m.selectedConnection() != nil {
					m.mode = ViewDetails
				}

			case key.Matches(msg, m.keys.Filter):
				m.mode = ViewFilter
				m.search.Input.Focus()
				return m, textinput.Blink

			case key.Matches(msg, m.keys.ToggleDNS):
				m.dnsPaneOpen = !m.dnsPaneOpen

			case key.Matches(msg, m.keys.Sort):
				m.sortMode = (m.sortMode + 1) % SortMode(len(sortModeNames))
				m.applyFilterAndSort()
				m.showStatus("Sort: "+sortModeNames[m.sortMode], 1500*time.Millisecond)

			case key.Matches(msg, m.keys.Theme):
				m.themeIndex = (m.themeIndex + 1) % len(theme.AllThemes)
				m.styles = theme.NewStyles(theme.AllThemes[m.themeIndex])
				m.search = components.NewSearchBar(m.styles)
				m.showStatus("Theme: "+theme.AllThemes[m.themeIndex].Name, 1500*time.Millisecond)

			case key.Matches(msg, m.keys.Pause):
				m.paused = !m.paused
				if m.paused {
					m.showStatus("Telemetry Paused", 2*time.Second)
				} else {
					m.showStatus("Telemetry Resumed", 1500*time.Millisecond)
				}

			case key.Matches(msg, m.keys.Kill):
				if m.selectedConnection() != nil {
					m.mode = ViewKill
				}

			case key.Matches(msg, m.keys.Cancel):
				if m.search.Input.Value() != "" {
					m.search.Input.SetValue("")
					m.applyFilterAndSort()
				}
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *AppModel) refreshData() {
	conns, err := m.collector.GetConnections()
	if err == nil {
		m.connections = conns
		// Enrich connections with GeoIP, flags, ASN, rDNS
		m.enricher.Enrich(m.connections)

		// Probe active TCP connections for RTT
		for _, c := range m.connections {
			if c.State == "ESTABLISHED" {
				c.LatencyMs = m.prober.Probe(c)
			}
		}
	}

	bwStats, err := m.collector.GetBandwidth()
	if err == nil && bwStats != nil {
		// Update connection counts
		var estab, listen, udp, tcp int
		for _, c := range m.connections {
			if strings.HasPrefix(c.Protocol, "tcp") {
				tcp++
			} else if strings.HasPrefix(c.Protocol, "udp") {
				udp++
			}
			if c.State == "ESTABLISHED" {
				estab++
			} else if c.State == "LISTEN" {
				listen++
			}
		}
		bwStats.TotalConnections = len(m.connections)
		bwStats.ActiveEstablished = estab
		bwStats.ActiveListen = listen
		bwStats.ActiveTCP = tcp
		bwStats.ActiveUDP = udp
		m.stats = bwStats
	}

	m.applyFilterAndSort()
}

func (m *AppModel) applyFilterAndSort() {
	query := m.search.Input.Value()
	filtered := components.FilterConnections(m.connections, query)

	// Sort
	switch m.sortMode {
	case SortByState:
		sort.SliceStable(filtered, func(i, j int) bool {
			score := func(s string) int {
				switch s {
				case "ESTABLISHED":
					return 0
				case "SYN_SENT", "SYN_RECV":
					return 1
				case "LISTEN":
					return 2
				case "TIME_WAIT", "CLOSE_WAIT":
					return 3
				default:
					return 4
				}
			}
			return score(filtered[i].State) < score(filtered[j].State)
		})
	case SortByProcess:
		sort.SliceStable(filtered, func(i, j int) bool {
			return strings.ToLower(filtered[i].ProcessName) < strings.ToLower(filtered[j].ProcessName)
		})
	case SortByRemote:
		sort.SliceStable(filtered, func(i, j int) bool {
			return strings.ToLower(filtered[i].RemoteEndpoint()) < strings.ToLower(filtered[j].RemoteEndpoint())
		})
	case SortByLatency:
		sort.SliceStable(filtered, func(i, j int) bool {
			li := filtered[i].LatencyMs
			lj := filtered[j].LatencyMs
			if li < 0 {
				li = 99999
			}
			if lj < 0 {
				lj = 99999
			}
			return li < lj
		})
	case SortByCountry:
		sort.SliceStable(filtered, func(i, j int) bool {
			return filtered[i].CountryCode < filtered[j].CountryCode
		})
	}

	m.filteredConns = filtered

	// Bounds check cursor
	if m.table.Cursor >= len(m.filteredConns) {
		m.table.Cursor = len(m.filteredConns) - 1
	}
	if m.table.Cursor < 0 {
		m.table.Cursor = 0
	}
}

func (m *AppModel) selectedConnection() *model.Connection {
	if len(m.filteredConns) == 0 || m.table.Cursor < 0 || m.table.Cursor >= len(m.filteredConns) {
		return nil
	}
	return m.filteredConns[m.table.Cursor]
}

func (m *AppModel) showStatus(msg string, duration time.Duration) {
	m.statusMsg = msg
	m.statusExpiry = time.Now().Add(duration)
}

func (m *AppModel) View() string {
	if m.width <= 0 || m.height <= 0 {
		return "Initializing netpulse..."
	}

	w := m.width
	h := m.height

	headerHeight := 4
	footerHeight := 2
	bodyHeight := h - headerHeight - footerHeight
	if bodyHeight < 5 {
		bodyHeight = 5
	}

	headerView := components.RenderHeader(m.styles, m.stats, w, m.paused, theme.AllThemes[m.themeIndex].Name)
	footerView := RenderFooter(m.styles, m.statusMsg, sortModeNames[m.sortMode], w)

	var bodyView string

	switch m.mode {
	case ViewDetails:
		bodyView = components.RenderDetailsModal(m.styles, m.selectedConnection(), w, bodyHeight)
	case ViewKill:
		bodyView = components.RenderKillModal(m.styles, m.selectedConnection(), w, bodyHeight)
	default:
		// Normal or Filter mode
		if m.dnsPaneOpen {
			tableH := bodyHeight * 60 / 100
			dnsH := bodyHeight - tableH
			if tableH < 4 {
				tableH = 4
			}
			if dnsH < 4 {
				dnsH = 4
			}

			tableView := m.table.Render(m.styles, m.filteredConns, w, tableH)
			priv, status := m.sniffer.Status()
			dnsView := components.RenderDNSPane(m.styles, m.sniffer.GetQueries(), priv, status, w, dnsH)
			bodyView = lipgloss.JoinVertical(lipgloss.Left, tableView, dnsView)
		} else {
			searchBarView := ""
			tableHeight := bodyHeight
			if m.mode == ViewFilter || m.search.Input.Value() != "" {
				searchBarView = m.styles.SearchBar.Width(w).Render(m.search.Input.View()) + "\n"
				tableHeight -= 2
			}
			tableView := m.table.Render(m.styles, m.filteredConns, w, tableHeight)
			if searchBarView != "" {
				bodyView = searchBarView + tableView
			} else {
				bodyView = tableView
			}
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, headerView, bodyView, footerView)
}
