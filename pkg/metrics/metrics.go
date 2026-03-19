package metrics

import (
	"sync"

	"github.com/containrrr/watchtower/pkg/types"
)

var (
	metrics     *Metrics
	metricsOnce sync.Once
)

// Metric is the data points of a single scan
type Metric struct {
	Scanned int
	Updated int
	Failed  int
}

// Metrics is the handler processing all individual scan metrics
// Simplified version without Prometheus integration
type Metrics struct {
	channel chan *Metric

	// Simple counters for tracking
	scanned int
	updated int
	failed  int
	total   int
	skipped int
	mu      sync.RWMutex
}

// NewMetric returns a Metric with the counts taken from the appropriate types.Report fields
func NewMetric(report types.Report) *Metric {
	return &Metric{
		Scanned: len(report.Scanned()),
		// Note: This is for backwards compatibility. ideally, stale containers should be counted separately
		Updated: len(report.Updated()) + len(report.Stale()),
		Failed:  len(report.Failed()),
	}
}

// QueueIsEmpty checks whether any messages are enqueued in the channel
func (metrics *Metrics) QueueIsEmpty() bool {
	return len(metrics.channel) == 0
}

// Register registers metrics for an executed scan
func (metrics *Metrics) Register(metric *Metric) {
	metrics.channel <- metric
}

// Default creates a new metrics handler if none exists, otherwise returns the existing one
func Default() *Metrics {
	metricsOnce.Do(func() {
		metrics = &Metrics{
			// Increased buffer size to handle high-throughput scenarios
			// This prevents blocking when many containers are updated simultaneously
			channel: make(chan *Metric, 100),
		}

		go metrics.HandleUpdate(metrics.channel)
	})

	return metrics
}

// RegisterScan fetches a metric handler and enqueues a metric
func RegisterScan(metric *Metric) {
	metrics := Default()
	metrics.Register(metric)
}

// HandleUpdate dequeue the metric channel and processes it
func (metrics *Metrics) HandleUpdate(channel <-chan *Metric) {
	for change := range channel {
		if change == nil {
			// Update was skipped and rescheduled
			metrics.mu.Lock()
			metrics.total++
			metrics.skipped++
			metrics.scanned = 0
			metrics.updated = 0
			metrics.failed = 0
			metrics.mu.Unlock()
			continue
		}
		// Update metrics with the new values
		metrics.mu.Lock()
		metrics.total++
		metrics.scanned = change.Scanned
		metrics.updated = change.Updated
		metrics.failed = change.Failed
		metrics.mu.Unlock()
	}
}

// GetScanned returns the number of containers scanned in the last scan
func (metrics *Metrics) GetScanned() int {
	metrics.mu.RLock()
	defer metrics.mu.RUnlock()
	return metrics.scanned
}

// GetUpdated returns the number of containers updated in the last scan
func (metrics *Metrics) GetUpdated() int {
	metrics.mu.RLock()
	defer metrics.mu.RUnlock()
	return metrics.updated
}

// GetFailed returns the number of containers where update failed in the last scan
func (metrics *Metrics) GetFailed() int {
	metrics.mu.RLock()
	defer metrics.mu.RUnlock()
	return metrics.failed
}

// GetTotal returns the total number of scans since watchtower started
func (metrics *Metrics) GetTotal() int {
	metrics.mu.RLock()
	defer metrics.mu.RUnlock()
	return metrics.total
}

// GetSkipped returns the total number of skipped scans since watchtower started
func (metrics *Metrics) GetSkipped() int {
	metrics.mu.RLock()
	defer metrics.mu.RUnlock()
	return metrics.skipped
}