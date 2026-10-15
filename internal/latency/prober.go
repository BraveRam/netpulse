package latency

import (
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/BraveRam/netpulse/internal/model"
)

// RTTRecord stores measured latency and expiry.
type RTTRecord struct {
	LatencyMs  int
	MeasuredAt time.Time
}

// LatencyProber measures active TCP handshake latency to remote endpoints with TTL caching.
type LatencyProber struct {
	mu       sync.RWMutex
	cache    map[string]*RTTRecord
	inFlight map[string]bool
	cacheTTL time.Duration
	sem      chan struct{} // Concurrency limiter
}

// NewLatencyProber creates a LatencyProber.
func NewLatencyProber(cacheTTL time.Duration, maxConcurrency int) *LatencyProber {
	if cacheTTL <= 0 {
		cacheTTL = 20 * time.Second
	}
	if maxConcurrency <= 0 {
		maxConcurrency = 6
	}

	return &LatencyProber{
		cache:    make(map[string]*RTTRecord),
		inFlight: make(map[string]bool),
		cacheTTL: cacheTTL,
		sem:      make(chan struct{}, maxConcurrency),
	}
}

// Probe returns the cached latency in ms, or -1 if not yet measured or unavailable.
// If not cached or expired, it triggers an async probe.
func (p *LatencyProber) Probe(conn *model.Connection) int {
	if conn == nil || (conn.Protocol != "tcp" && conn.Protocol != "tcp6") || conn.State == "LISTEN" || conn.RemotePort == 0 || conn.RemoteIP == nil {
		return -1
	}

	// Skip loopback or local addresses for probing
	if conn.RemoteIP.IsLoopback() || conn.RemoteIP.IsUnspecified() {
		return 0
	}

	target := net.JoinHostPort(conn.RemoteIPStr, strconv.Itoa(conn.RemotePort))

	p.mu.RLock()
	record, found := p.cache[target]
	p.mu.RUnlock()

	now := time.Now()
	if found && now.Sub(record.MeasuredAt) < p.cacheTTL {
		return record.LatencyMs
	}

	p.triggerAsyncProbe(target, conn.Protocol)

	if found {
		return record.LatencyMs
	}
	return -1
}

func (p *LatencyProber) triggerAsyncProbe(target, proto string) {
	p.mu.Lock()
	if p.inFlight[target] {
		p.mu.Unlock()
		return
	}
	p.inFlight[target] = true
	p.mu.Unlock()

	go func() {
		defer func() {
			p.mu.Lock()
			delete(p.inFlight, target)
			p.mu.Unlock()
		}()

		// Acquire semaphore slot
		select {
		case p.sem <- struct{}{}:
			defer func() { <-p.sem }()
		default:
			// Busy, skip this probe cycle
			return
		}

		network := "tcp"
		if proto == "tcp6" {
			network = "tcp6"
		}

		start := time.Now()
		c, err := net.DialTimeout(network, target, 400*time.Millisecond)
		rtt := time.Since(start)

		latencyMs := -1
		if err == nil {
			_ = c.Close()
			latencyMs = int(rtt.Milliseconds())
			if latencyMs == 0 {
				latencyMs = 1
			}
		}

		p.mu.Lock()
		p.cache[target] = &RTTRecord{
			LatencyMs:  latencyMs,
			MeasuredAt: time.Now(),
		}
		p.mu.Unlock()
	}()
}
