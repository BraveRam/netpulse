package dns

import (
	"sync"

	"github.com/plxor/netpulse/internal/model"
)

// RingBuffer stores a fixed-capacity history of DNS queries.
type RingBuffer struct {
	mu       sync.RWMutex
	capacity int
	entries  []*model.DNSQuery
}

// NewRingBuffer creates a RingBuffer with specified capacity.
func NewRingBuffer(capacity int) *RingBuffer {
	if capacity <= 0 {
		capacity = 200
	}
	return &RingBuffer{
		capacity: capacity,
		entries:  make([]*model.DNSQuery, 0, capacity),
	}
}

// Add appends a query, evicting the oldest if capacity is reached.
func (r *RingBuffer) Add(q *model.DNSQuery) {
	if q == nil {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.entries) >= r.capacity {
		// Drop oldest
		r.entries = r.entries[1:]
	}
	r.entries = append(r.entries, q)
}

// GetAll returns a snapshot of all stored queries (newest first).
func (r *RingBuffer) GetAll() []*model.DNSQuery {
	r.mu.RLock()
	defer r.mu.RUnlock()

	n := len(r.entries)
	res := make([]*model.DNSQuery, n)
	for i := 0; i < n; i++ {
		res[i] = r.entries[n-1-i]
	}
	return res
}

// Count returns number of items in the buffer.
func (r *RingBuffer) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.entries)
}
