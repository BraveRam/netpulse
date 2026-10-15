<div align="center">

# netpulse

**See what your Linux machine is talking to.**

A terminal dashboard for live sockets, process ownership, DNS activity, and network throughput.

[Getting started](#getting-started) · [Controls](#controls) · [Configuration](#configuration) · [Development](#development)

</div>

---

Keep network activity in view without switching between process lists, socket tools, and packet captures. Netpulse reads Linux `/proc` data, enriches remote endpoints in the background, and brings the results into one searchable terminal interface.

## At a glance

| View | What it tells you |
| --- | --- |
| Connections | TCP and UDP sockets, local and remote endpoints, socket state, and owning process |
| Endpoint context | Reverse DNS names, ASN organizations, and country/city data from an optional MaxMind database |
| Network activity | Aggregate interface throughput with receive/transmit sparklines |
| DNS pane | Captured UDP DNS queries and responses, answer records, and request timing |
| Inspector | Process command line, user, socket inode, endpoints, and enrichment details |
| Process controls | Confirmed graceful termination or force kill of the selected process |

TCP handshake probes provide a periodically cached latency estimate for established TCP connections. This measures a new connection to the remote endpoint, rather than the RTT of the existing socket.

## Getting started

**Requirements:** Linux, Go 1.25 or newer, and a terminal. A wide terminal works best for the full connection table.

### Install with Go

```bash
go install github.com/BraveRam/netpulse/cmd/netpulse@latest
netpulse
```

If `netpulse` is not found, add Go's binary directory to your `PATH`:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

Add that line to your shell profile to keep it across terminal sessions. If you set a custom `GOBIN`, use that directory instead.

Start as your normal user to inspect available sockets and process metadata. Access to other users' process details depends on `/proc` permissions.

For live DNS capture and broader process visibility:

```bash
sudo "$(go env GOPATH)/bin/netpulse"
```

To enable packet capture without running the interface as root:

```bash
sudo setcap cap_net_raw+ep "$(go env GOPATH)/bin/netpulse"
netpulse
```

Use your `GOBIN` path in these commands if configured. Reinstalling replaces the binary, so capabilities may need to be granted again. This capability enables capture; it does not grant permission to terminate another user's processes.

### Build from source

```bash
git clone https://github.com/BraveRam/netpulse.git
cd netpulse
make build
./bin/netpulse
```

For a source build, use `sudo ./bin/netpulse` for privileged execution, or apply `setcap` to `./bin/netpulse`.

## Controls

| Key | Action |
| --- | --- |
| `↑` / `↓` / `j` | Move through connections |
| `PgUp` / `PgDn` | Move ten rows |
| `g` / `G` | Jump to the first / last connection |
| `Enter` | Open the inspector; press again to close |
| `/` | Filter connections; `Enter` keeps the filter |
| `Esc` | Leave the current view; in the normal view, clear the filter |
| `d` | Toggle the DNS pane |
| `s` | Cycle sorting by state, process, remote endpoint, latency, or country |
| `k` | Open the process termination dialog |
| `y` / `Y` | Confirm `SIGTERM` / `SIGKILL` in the termination dialog |
| `n` / `Esc` | Cancel process termination |
| `Space` | Pause / resume connection and bandwidth refreshes |
| `t` | Cycle Cyberpunk, Catppuccin Mocha, Tokyo Night, and Nord themes |
| `q` / `Ctrl+C` | Quit from the normal view |

Filters match process names, PIDs, IP addresses, domains, ports, organizations, ASNs, countries, and states. The termination dialog keeps the process selected when it was opened, even if the connection list refreshes.

## Configuration

```bash
# Refresh twice per second with the Nord theme
netpulse -interval 500ms -theme nord

# Run without DNS packet capture
netpulse -dns=false

# Use a locally obtained MaxMind City or Country database
netpulse -geoip /path/to/GeoLite2-City.mmdb
```

| Flag | Default | Purpose |
| --- | --- | --- |
| `-interval` | `1s` | Connection and bandwidth refresh interval |
| `-theme` | `cyberpunk` | `cyberpunk`, `catppuccin`, `tokyonight`, or `nord` |
| `-dns` | `true` | Start the DNS packet sniffer |
| `-geoip` | Auto-discover | Path to a MaxMind City or Country `.mmdb` file |
| `-version` | — | Print the version and exit |

Without an explicit path, Netpulse checks `/var/lib/GeoIP` and `/usr/share/GeoIP` for `GeoLite2-City.mmdb` and `GeoLite2-Country.mmdb`. Private and local addresses are identified directly. Missing databases or unmatched public addresses fall back to unknown geography; ASN metadata may still provide a country code.

## How collection works

- **Sockets and processes:** parses `/proc/net/{tcp,tcp6,udp,udp6}` and maps socket inodes to `/proc/<pid>/fd` entries. Process ownership is cached briefly.
- **Enrichment:** combines built-in ASN ranges, background IPv4 ASN DNS queries, cached reverse DNS, and an optional local GeoIP database. External lookups create DNS traffic of their own.
- **Throughput:** reads interface byte counters from `/proc/net/dev`. Rates are aggregate interface measurements; per-connection byte rates are not currently collected.
- **DNS capture:** uses a Linux packet socket for IPv4 UDP traffic on port 53. IPv6 DNS transport, TCP DNS, DNS over HTTPS, and DNS over TLS are not decoded.

Pausing freezes connection and bandwidth refreshes. Background DNS capture continues. Process termination refuses PID 1 and Netpulse itself, and remains subject to normal Linux permissions.

## Development

```bash
make build          # Build bin/netpulse
make test           # Run all tests with the race detector
go vet ./...        # Run static checks
```

Regression tests cover UDP socket states, IPv6 probe addressing, DNS endpoint direction and transaction matching, MaxMind database lookups, and the process termination shortcut and target selection.

For a small stream of local and outbound demo traffic:

```bash
./scripts/demo_traffic.sh
```

The implementation uses [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Bubbles](https://github.com/charmbracelet/bubbles), [Lip Gloss](https://github.com/charmbracelet/lipgloss), and [MaxMind DB Reader](https://github.com/oschwald/maxminddb-golang).
