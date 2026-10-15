# ⚡ netpulse

<div align="center">

```
  _  _ ___ _____ ___ _   _ _    ___ ___ 
 | \| | __|_   _| _ \ | | | |  / __| __|
 | .` | _|  | | |  _/ |_| | |__\__ \ _| 
 |_|\_|___| |_| |_|  \___/|____|___/___| 
```

**Real-Time Terminal Radar for Network Activity, Sockets, and DNS Telemetry**

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Built with Charm](https://img.shields.io/badge/Built%20with-Charm-ff69b4.svg)](https://charm.sh)

</div>

---

## 🎯 Overview

**`netpulse`** is a blazing-fast, visually captivating Terminal User Interface (TUI) built in Go with the [Charmbracelet](https://charm.sh) suite (`bubbletea`, `lipgloss`, `bubbles`).

It operates like a real-time terminal radar for your machine's network stack:
- Maps out every outgoing TCP/UDP connection, listening socket, and process attribution.
- Enriches network flows with **country flags**, **autonomous system (ASN) organization names**, and **reverse DNS hostnames**.
- Actively probes **round-trip latency (RTT)** with color-coded meters.
- Displays live **sparkline bandwidth graphs** (` ▂▃▅▇`) for incoming and outgoing throughput.
- Features a **Live DNS Query Sniffer** tracking live domain lookups made on the host.
- Includes an **instant kill switch** (`k`) to terminate rogue processes or background connections on sight.

---

## 📸 Radar Interface Preview

```
⚡ NETPULSE @my-host                            [LIVE] [Cyberpunk]
TOTAL 34  ESTAB 12  LISTEN 8  UDP 14            ▼ 1.42 MB/s  ▂▃▅▇   ▲ 340 KB/s   ▂▄▆
──────────────────────────────────────────────────────────────────────────────────────────
  FLAG      PROCESS            REMOTE ENDPOINT             ORG / ASN          STATE        LATENCY    TRAFFIC
  🇺🇸 US     curl [4102]        api.github.com:443          GitHub, Inc.       ESTABLISHED  ● 18ms     1.4 MB/s
  🇩🇪 DE     spotify [2134]     audio-fa.spotify.com:443    Fastly, Inc.       ESTABLISHED  ● 42ms     320 KB/s
  🇺🇸 US     chrome [9812]      1.1.1.1:443                 Cloudflare, Inc.   ESTABLISHED  ● 14ms     -
  📍 LOCAL  sshd [802]         *:*                         Local System       LISTEN       ● -        -
  📍 LOCAL  dnsmasq [612]      *:*                         Local System       LISTEN       ● -        -
  🇯🇵 JP     ssh [14022]        jump.tokyo.net:22           Hetzner Online     ESTABLISHED  ● 182ms    48 KB/s
──────────────────────────────────────────────────────────────────────────────────────────
📡 LIVE DNS QUERY SNIFFER  ● Active
TIME     CLIENT          TYPE   QUERY DOMAIN                     STATUS     RTT      ANSWERS
10:45:02 127.0.0.1:53210 A      api.github.com                   NOERROR    12ms     140.82.112.5
10:45:03 127.0.0.1:49214 AAAA   cloudflare.com                   NOERROR    18ms     2606:4700::6810:85e5
──────────────────────────────────────────────────────────────────────────────────────────
[↑/↓] Nav • [Enter] Inspect • [/] Filter • [d] DNS • [s] Sort • [k] Kill • [Space] Pause • [t] Theme • [q] Quit
```

---

## 🚀 Key Features

### 1. Live Connections Radar
- **High-Speed Linux `/proc/net` Engine**: Zero-allocation hex parser extracting TCP/UDP sockets, states, and file descriptor inodes in milliseconds.
- **Process Attribution**: Automatically matches socket inodes to running PIDs, process names, command-line arguments, and process owner usernames via `/proc/[pid]/fd`.
- **Country Flags & GeoIP**: Displays country flag emojis (`🇺🇸 US`, `🇩🇪 DE`, `🇯🇵 JP`, `📍 Local`) for every connection. Detects RFC1918, loopback, and carrier NAT addresses automatically.
- **Autonomous System (ASN) & Organization Attribution**: Resolves remote endpoints to their parent autonomous system (`Google LLC`, `Cloudflare`, `Amazon AWS`, `Hetzner`, `Apple`) using built-in CIDR maps and asynchronous DNS TXT lookups (`origin.asn.cymru.com`).
- **Live Latency / RTT Meter**: Probes TCP handshake round-trip time in milliseconds:
  - `● <60ms` Neon Green
  - `● 60-180ms` Amber Yellow
  - `● >180ms` Bright Red
- **Sparkline Bandwidth Meters**: Dynamic unicode sparklines (` ▂▃▄▅▆▇█`) visualizing incoming and outgoing interface byte rates over time.

### 2. Live DNS Query Sniffer (`d`)
- Dedicated toggleable lower pane capturing live outbound DNS queries and answers on UDP port 53.
- Shows timestamp, client port, record type (A, AAAA, CNAME, MX, TXT, HTTPS), status code (NOERROR, NXDOMAIN), round-trip query latency, and resolved answer IP addresses.

### 3. Instant Kill Switch (`k`)
- Terminate rogue background sockets or suspicious processes with safety guards.
- Supports graceful termination (`SIGTERM` via `y`) or immediate force kill (`SIGKILL` via `Y`).
- Safety protections prevent terminating `PID 1` (init/systemd), kernel threads, or `netpulse` itself.

### 4. Interactive Telemetry Inspector (`Enter`)
- Pressing `Enter` on any highlighted connection displays a deep-dive inspection card with:
  - Full executable command line and arguments
  - Socket inodes and file descriptor details
  - Geolocation (Country, City, Flag)
  - ASN and Organization
  - Reverse DNS hostname aliases
  - Active RTT latency

### 5. Instant Filtering & Sorting (`/` & `s`)
- **Live Search (`/`)**: Instant filter across Process Name, PID, IP, Domain, Port, Country, ASN, or State.
- **Sorting Modes (`s`)**: Cycle sort ordering by State, Process Name, Remote Host, Latency RTT, or Country Code.

### 6. Cyberpunk & Modern Themes (`t`)
- Press `t` to cycle dynamically through themes:
  - ⚡ **Cyberpunk**: Electric Cyan, Neon Magenta, Neon Green, Obsidian
  - 🌸 **Catppuccin Mocha**: Mauve, Sapphire, Peach, Surface0
  - 🌃 **Tokyo Night**: Deep Indigo, Electric Purple, Cyan
  - ❄️ **Nord**: Arctic Blue, Polar Night, Aurora Green

---

## 📦 Installation & Build

### Prerequisites
- **Go 1.22+**
- Linux (x86_64 or arm64)

### Build from Source
```bash
git clone https://github.com/plxor/netpulse.git
cd netpulse

make build
```

The compiled binary will be placed at `bin/netpulse`.

### Privileged vs. Normal Execution
`netpulse` runs seamlessly as an unprivileged user, displaying connections and processes owned by your user account.

To inspect all system processes and capture raw DNS frames:
```bash
sudo ./bin/netpulse
```
Or grant Linux capabilities to the binary:
```bash
sudo setcap cap_net_raw,cap_net_admin+ep ./bin/netpulse
./bin/netpulse
```

---

## ⌨️ Keybindings

| Key | Action |
|:---|:---|
| `↑` / `k` | Move cursor up |
| `↓` / `j` | Move cursor down |
| `PgUp` / `PgDn` | Scroll page up / down |
| `g` / `G` | Jump to top / bottom |
| `Enter` | Open detailed Connection Inspector modal |
| `/` | Open instant live filter / search bar |
| `d` | Toggle Live DNS Query Sniffer pane |
| `s` | Cycle sort mode (State, Process, Remote, Latency, Country) |
| `k` | Open Process Termination / Kill modal |
| `Space` | Pause / Resume real-time telemetry updates |
| `t` | Cycle color theme (Cyberpunk, Catppuccin, Tokyo Night, Nord) |
| `Esc` | Clear search / close open modal |
| `q` / `Ctrl+C` | Exit `netpulse` |

---

## ⚙️ CLI Flags

```text
Usage: netpulse [flags]

Flags:
  -dns
    	Enable real-time DNS packet sniffer (default true)
  -geoip string
    	Path to optional MaxMind GeoLite2-City.mmdb
  -interval duration
    	Telemetry refresh interval (e.g. 500ms, 1s, 2s) (default 1s)
  -theme string
    	Initial UI theme (cyberpunk, catppuccin, tokyonight, nord) (default "cyberpunk")
  -version
    	Print version and exit
```

---

## 🧪 Testing

Run the full test suite with race detector:
```bash
make test
```

Generate synthetic demo traffic to test connections, RTT meters, and DNS sniffer:
```bash
./scripts/demo_traffic.sh
```

---

## 📄 License
MIT License. Crafted with ❤️ for terminal power users.
