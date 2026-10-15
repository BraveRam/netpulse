package model

import "fmt"

// BandwidthStats holds network interface counters and throughput rates.
type BandwidthStats struct {
	TotalRxBytes uint64
	TotalTxBytes uint64
	RateRx       float64 // Bytes per second
	RateTx       float64 // Bytes per second

	// History buffer for sparkline generation
	RxHistory []float64
	TxHistory []float64

	// Connection summary counts
	ActiveTCP         int
	ActiveUDP         int
	ActiveListen      int
	ActiveEstablished int
	TotalConnections  int
}

// FormatSpeed formats bytes per second into human-readable rate, e.g. "1.4 MB/s".
func FormatSpeed(bytesPerSec float64) string {
	const (
		kb = 1024.0
		mb = 1024.0 * kb
		gb = 1024.0 * mb
	)

	switch {
	case bytesPerSec >= gb:
		return fmt.Sprintf("%.2f GB/s", bytesPerSec/gb)
	case bytesPerSec >= mb:
		return fmt.Sprintf("%.2f MB/s", bytesPerSec/mb)
	case bytesPerSec >= kb:
		return fmt.Sprintf("%.1f KB/s", bytesPerSec/kb)
	default:
		return fmt.Sprintf("%.0f B/s", bytesPerSec)
	}
}

// FormatBytes formats total byte counts into human-readable size, e.g. "14.2 GB".
func FormatBytes(b uint64) string {
	const (
		kb = 1024
		mb = 1024 * kb
		gb = 1024 * mb
		tb = 1024 * gb
	)

	switch {
	case b >= tb:
		return fmt.Sprintf("%.2f TB", float64(b)/float64(tb))
	case b >= gb:
		return fmt.Sprintf("%.2f GB", float64(b)/float64(gb))
	case b >= mb:
		return fmt.Sprintf("%.1f MB", float64(b)/float64(mb))
	case b >= kb:
		return fmt.Sprintf("%.1f KB", float64(b)/float64(kb))
	default:
		return fmt.Sprintf("%d B", b)
	}
}
