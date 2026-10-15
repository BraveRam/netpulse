package enricher

import (
	"context"
	"net"
	"strings"
	"sync"
	"time"
)

// RDNSResolver performs concurrent, rate-limited reverse DNS lookups with caching.
type RDNSResolver struct {
	mu         sync.RWMutex
	cache      map[string]string // IP -> Hostname
	inFlight   map[string]bool
	jobs       chan string
	ctx        context.Context
	cancel     context.CancelFunc
	maxWorkers int
}

// NewRDNSResolver creates an RDNSResolver with a background worker pool.
func NewRDNSResolver(workers int) *RDNSResolver {
	if workers <= 0 {
		workers = 4
	}

	ctx, cancel := context.WithCancel(context.Background())
	r := &RDNSResolver{
		cache:      make(map[string]string),
		inFlight:   make(map[string]bool),
		jobs:       make(chan string, 1000),
		ctx:        ctx,
		cancel:     cancel,
		maxWorkers: workers,
	}

	// Start worker pool
	for i := 0; i < workers; i++ {
		go r.worker()
	}

	return r
}

// Lookup checks the cache for an IP's reverse DNS hostname.
// If not found, it enqueues a background lookup and returns "".
func (r *RDNSResolver) Lookup(ipStr string) string {
	if ipStr == "" || ipStr == "0.0.0.0" || ipStr == "::" {
		return ""
	}

	r.mu.RLock()
	hostname, found := r.cache[ipStr]
	r.mu.RUnlock()

	if found {
		return hostname
	}

	r.enqueue(ipStr)
	return ""
}

// enqueue adds an IP to the background lookup channel if not already in flight.
func (r *RDNSResolver) enqueue(ipStr string) {
	r.mu.Lock()
	if r.inFlight[ipStr] {
		r.mu.Unlock()
		return
	}
	r.inFlight[ipStr] = true
	r.mu.Unlock()

	select {
	case r.jobs <- ipStr:
	default:
		// Queue full, drop job
		r.mu.Lock()
		delete(r.inFlight, ipStr)
		r.mu.Unlock()
	}
}

// worker processes background reverse DNS queries.
func (r *RDNSResolver) worker() {
	for {
		select {
		case <-r.ctx.Done():
			return
		case ipStr, ok := <-r.jobs:
			if !ok {
				return
			}
			r.resolveOne(ipStr)
		}
	}
}

func (r *RDNSResolver) resolveOne(ipStr string) {
	defer func() {
		r.mu.Lock()
		delete(r.inFlight, ipStr)
		r.mu.Unlock()
	}()

	// Query with short timeout
	resolver := &net.Resolver{}
	ctx, cancel := context.WithTimeout(r.ctx, 1500*time.Millisecond)
	names, err := resolver.LookupAddr(ctx, ipStr)
	cancel()

	var hostname string
	if err == nil && len(names) > 0 {
		hostname = strings.TrimSuffix(names[0], ".")
	}

	r.mu.Lock()
	// Cache even empty string to avoid repeatedly re-querying dead IPs
	if len(r.cache) > 10000 {
		// Prune cache
		for k := range r.cache {
			delete(r.cache, k)
			if len(r.cache) <= 5000 {
				break
			}
		}
	}
	r.cache[ipStr] = hostname
	r.mu.Unlock()
}

// Close stops background workers.
func (r *RDNSResolver) Close() {
	r.cancel()
}
