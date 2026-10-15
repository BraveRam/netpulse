package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/BraveRam/netpulse/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

const version = "1.0.0"

const banner = "  _  _ ___ _____ ___ _   _ _    ___ ___ \n" +
	" | \\| | __|_   _| _ \\ | | | |  / __| __|\n" +
	" | .` | _|  | | |  _/ |_| | |__\\__ \\ _| \n" +
	" |_|\\_|___| |_| |_|  \\___/|____|___/___|\n" +
	" Real-Time Network Telemetry & Socket Radar"

func main() {
	var (
		intervalFlag = flag.Duration("interval", 1*time.Second, "Telemetry refresh interval (e.g. 500ms, 1s, 2s)")
		geoipFlag    = flag.String("geoip", "", "Path to optional MaxMind GeoLite2-City.mmdb")
		themeFlag    = flag.String("theme", "cyberpunk", "Initial UI theme (cyberpunk, catppuccin, tokyonight, nord)")
		dnsFlag      = flag.Bool("dns", true, "Enable real-time DNS packet sniffer")
		versionFlag  = flag.Bool("version", false, "Print version and exit")
	)

	flag.Usage = func() {
		fmt.Printf("%s\nVersion: %s\n\nUsage: netpulse [flags]\n\nFlags:\n", banner, version)
		flag.PrintDefaults()
		fmt.Println("\nKeybindings:")
		fmt.Println("  ↑ / ↓ / j  Navigate connection rows")
		fmt.Println("  Enter          Inspect highlighted connection telemetry")
		fmt.Println("  /              Filter by process, PID, IP, domain, port, org, state")
		fmt.Println("  d              Toggle Live DNS query sniffer pane")
		fmt.Println("  s              Cycle sort column (State, Process, Remote, Latency, Country)")
		fmt.Println("  k              Open instant kill dialog for selected process")
		fmt.Println("  Space          Pause / Resume real-time telemetry")
		fmt.Println("  t              Cycle color theme (Cyberpunk, Catppuccin, Tokyo Night, Nord)")
		fmt.Println("  q / Ctrl+C     Quit netpulse")
	}

	flag.Parse()

	if *versionFlag {
		fmt.Printf("netpulse v%s\n", version)
		os.Exit(0)
	}

	themeIndex := 0
	switch strings.ToLower(*themeFlag) {
	case "catppuccin", "cat":
		themeIndex = 1
	case "tokyonight", "tokyo":
		themeIndex = 2
	case "nord":
		themeIndex = 3
	default:
		themeIndex = 0 // Cyberpunk
	}

	cfg := ui.Config{
		TickInterval: *intervalFlag,
		GeoIPPath:    *geoipFlag,
		EnableDNS:    *dnsFlag,
		ThemeIndex:   themeIndex,
	}

	model := ui.NewAppModel(cfg)

	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running netpulse: %v\n", err)
		os.Exit(1)
	}
}
