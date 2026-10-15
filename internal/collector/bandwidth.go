package collector

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/plxor/netpulse/internal/model"
)

const maxHistoryPoints = 25

// BandwidthMonitor tracks global network interface byte rates and history.
type BandwidthMonitor struct {
	mu           sync.Mutex
	lastRxBytes  uint64
	lastTxBytes  uint64
	lastTime     time.Time
	rxHistory    []float64
	txHistory    []float64
	currentRxRate float64
	currentTxRate float64
	totalRxBytes uint64
	totalTxBytes uint64
	initialized  bool
}

// NewBandwidthMonitor creates a BandwidthMonitor.
func NewBandwidthMonitor() *BandwidthMonitor {
	return &BandwidthMonitor{
		rxHistory: make([]float64, 0, maxHistoryPoints),
		txHistory: make([]float64, 0, maxHistoryPoints),
	}
}

// Update reads /proc/net/dev and calculates the current throughput rates.
func (b *BandwidthMonitor) Update() (*model.BandwidthStats, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	rx, tx, err := readProcNetDev()
	if err != nil {
		return &model.BandwidthStats{}, err
	}

	now := time.Now()
	b.totalRxBytes = rx
	b.totalTxBytes = tx

	if !b.initialized {
		b.lastRxBytes = rx
		b.lastTxBytes = tx
		b.lastTime = now
		b.initialized = true
		return &model.BandwidthStats{
			TotalRxBytes: rx,
			TotalTxBytes: tx,
			RxHistory:    b.rxHistory,
			TxHistory:    b.txHistory,
		}, nil
	}

	deltaSec := now.Sub(b.lastTime).Seconds()
	if deltaSec > 0.05 {
		if rx >= b.lastRxBytes {
			b.currentRxRate = float64(rx-b.lastRxBytes) / deltaSec
		} else {
			b.currentRxRate = 0
		}

		if tx >= b.lastTxBytes {
			b.currentTxRate = float64(tx-b.lastTxBytes) / deltaSec
		} else {
			b.currentTxRate = 0
		}

		b.rxHistory = append(b.rxHistory, b.currentRxRate)
		if len(b.rxHistory) > maxHistoryPoints {
			b.rxHistory = b.rxHistory[1:]
		}

		b.txHistory = append(b.txHistory, b.currentTxRate)
		if len(b.txHistory) > maxHistoryPoints {
			b.txHistory = b.txHistory[1:]
		}

		b.lastRxBytes = rx
		b.lastTxBytes = tx
		b.lastTime = now
	}

	// Copy history slices for caller safety
	rxCopy := make([]float64, len(b.rxHistory))
	copy(rxCopy, b.rxHistory)
	txCopy := make([]float64, len(b.txHistory))
	copy(txCopy, b.txHistory)

	return &model.BandwidthStats{
		TotalRxBytes:  b.totalRxBytes,
		TotalTxBytes:  b.totalTxBytes,
		RateRx:        b.currentRxRate,
		RateTx:        b.currentTxRate,
		RxHistory:     rxCopy,
		TxHistory:     txCopy,
	}, nil
}

// readProcNetDev parses /proc/net/dev to sum total Rx and Tx bytes across non-loopback interfaces.
func readProcNetDev() (uint64, uint64, error) {
	file, err := os.Open("/proc/net/dev")
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	var totalRx, totalTx uint64
	var fallbackRx, fallbackTx uint64
	hasPhysical := false

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		colonIdx := strings.Index(line, ":")
		if colonIdx == -1 {
			continue
		}

		ifaceName := strings.TrimSpace(line[:colonIdx])
		dataPart := strings.TrimSpace(line[colonIdx+1:])
		fields := strings.Fields(dataPart)
		if len(fields) < 9 {
			continue
		}

		rxBytes, errRx := strconv.ParseUint(fields[0], 10, 64)
		txBytes, errTx := strconv.ParseUint(fields[8], 10, 64)
		if errRx != nil || errTx != nil {
			continue
		}

		if ifaceName == "lo" {
			fallbackRx += rxBytes
			fallbackTx += txBytes
			continue
		}

		hasPhysical = true
		totalRx += rxBytes
		totalTx += txBytes
	}

	// If only loopback is present (e.g. in test sandbox), use loopback
	if !hasPhysical {
		return fallbackRx, fallbackTx, nil
	}

	return totalRx, totalTx, nil
}
