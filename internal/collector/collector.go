package collector

import (
	"time"

	"github.com/BraveRam/netpulse/internal/model"
)

// Collector is the high-level interface for acquiring live socket and bandwidth metrics.
type Collector interface {
	GetConnections() ([]*model.Connection, error)
	GetBandwidth() (*model.BandwidthStats, error)
}

// SystemCollector is the default collector on Linux systems.
type SystemCollector struct {
	parser    *LinuxSocketParser
	bandwidth *BandwidthMonitor
}

// NewSystemCollector creates a configured SystemCollector.
func NewSystemCollector(inodeCacheTTL time.Duration) *SystemCollector {
	mapper := NewInodeMapper(inodeCacheTTL)
	parser := NewLinuxSocketParser(mapper)
	bandwidth := NewBandwidthMonitor()

	return &SystemCollector{
		parser:    parser,
		bandwidth: bandwidth,
	}
}

// GetConnections retrieves all active TCP and UDP sockets mapped to their processes.
func (c *SystemCollector) GetConnections() ([]*model.Connection, error) {
	return c.parser.ParseAll()
}

// GetBandwidth returns current network throughput rates and sparkline history.
func (c *SystemCollector) GetBandwidth() (*model.BandwidthStats, error) {
	return c.bandwidth.Update()
}
