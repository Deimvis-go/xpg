package pgconnprovider

import (
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/Deimvis-go/xprometheus/xprometheus"
)

func (f *fallbacked) Stats() FallbackedStats {
	return f.stats
}

type FallbackedStats interface {
	// Snapshot returns a snapshot of stored
	// statistics.
	// Statistic values are not synchronized.
	Snapshot() FallbackedStatsSnapshot
	// RegisterPrometheusExport registers
	// prometheus metrics where
	// statistics will be
	// automatically exported.
	RegisterPrometheusExport(FallbackedPrometheusMetrics) error
}

type FallbackedStatsSnapshot struct {
	AcquireAttemptCountTotal uint64
}

func NewFallbackedPrometheusMetrics(s FallbackedPrometheusMetricsStruct) FallbackedPrometheusMetrics {
	return xprometheus.NewStructMetricGroup(s)
}

type FallbackedPrometheusMetrics = *xprometheus.StructMetricGroup[FallbackedPrometheusMetricsStruct]

type FallbackedPrometheusMetricsStruct struct {
	// variable labels: ["ind", "acquire_type"]
	AcquireAttempt *xprometheus.IntervalMetricGroup
}

type fallbackedStats struct {
	acquireAttemptCount atomic.Uint64

	promExportsMu sync.RWMutex
	promExports   []FallbackedPrometheusMetrics
}

var _ FallbackedStats = (*fallbackedStats)(nil)

func (fs *fallbackedStats) Snapshot() FallbackedStatsSnapshot {
	return FallbackedStatsSnapshot{
		AcquireAttemptCountTotal: fs.acquireAttemptCount.Load(),
	}
}

func (fs *fallbackedStats) RegisterPrometheusExport(m FallbackedPrometheusMetrics) error {
	fs.promExportsMu.Lock()
	defer fs.promExportsMu.Unlock()
	fs.promExports = append(fs.promExports, m)
	return nil
}

func (fs *fallbackedStats) recordAttempt(attempt FallbackedAttemptState, fn func()) {
	fs.acquireAttemptCount.Add(1)

	var promExports []FallbackedPrometheusMetrics
	func() {
		fs.promExportsMu.RLock()
		defer fs.promExportsMu.RUnlock()
		promExports = fs.promExports
	}()

	ls := make(prometheus.Labels)
	if attempt.Index < 100 {
		ls["ind"] = strconv.FormatInt(int64(attempt.Index), 10)
		ls["acquire_type"] = attempt.AcquireTypeOr(xprometheus.LabelUnknown)
	} else {
		// avoid labels high cardinality
		ls["ind"] = xprometheus.LabelHighCardinality
		ls["acquire_type"] = xprometheus.LabelHighCardinality
	}
	for _, e := range promExports {
		if e.Struct().AcquireAttempt != nil {
			e.Struct().AcquireAttempt.StartC().With(ls).Inc()
		}
	}
	startTime := time.Now()
	fn()
	elapsed := time.Since(startTime)
	for _, e := range promExports {
		if e.Struct().AcquireAttempt != nil {
			e.Struct().AcquireAttempt.FinishC().With(ls).Inc()
			dur := e.Struct().AcquireAttempt.ScaleDuration(elapsed)
			e.Struct().AcquireAttempt.DurationH().Metric().With(ls).Observe(dur)
		}
	}
}

func (fs *fallbackedStats) clone() *fallbackedStats {
	fs.promExportsMu.Lock()
	defer fs.promExportsMu.Unlock()
	stats := &fallbackedStats{
		promExports: fs.promExports,
	}
	stats.acquireAttemptCount.Add(fs.acquireAttemptCount.Load())
	return stats
}
