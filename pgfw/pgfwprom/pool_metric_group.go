package pgfwprom

import (
	"errors"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/Deimvis/go-ext/go1.25/xoptional"
	"github.com/Deimvis-go/xprometheus/xprometheus"
	"github.com/Deimvis-go/xprometheus/xprometheus/xprommetric"
)

func NewConnPoolMetrics(pms ConnPoolMetricsStruct) ConnPoolMetrics {
	return ConnPoolMetrics{
		StructMetricGroup: xprometheus.NewStructMetricGroup(pms),
		mls:               xoptional.New[ConnPoolMetaLabels](),
		mlsPrecompiled:    false,

		metricsStateMu:            &sync.Mutex{},
		connAcquireCount:          0,
		connAcquireCancelledCount: 0,
	}
}

type ConnPoolMetrics struct {
	*xprometheus.StructMetricGroup[ConnPoolMetricsStruct]
	mls            xoptional.T[ConnPoolMetaLabels]
	mlsPrecompiled bool

	// metrics state
	metricsStateMu                           *sync.Mutex
	connAcquireCount                         int64
	connAcquireDurationTotal                 time.Duration
	connAcquireCancelledCount                int64
	connAcquireConnAvailabilityWaitTimeTotal time.Duration
}

// ConnPoolMetricsStruct metrics:
// - must have variable labels: []
// - may have meta labels: ["pool_name", "conn_mode"]
// Meta labels are constant labels
// describing pool's meta.
// One may set meta labels using
// prometheus const labels.
//
// TODO: tags for autoinit
// `prom:"metric('min_conns_count', ls=['pool_name','conn_mode'])"`
// then call InitUninitialized(WithMetricNameResolver(tag or fallback infer from field name))
type ConnPoolMetricsStruct struct {
	ConfigMinConnsCount *prometheus.GaugeVec
	ConfigMaxConnsCount *prometheus.GaugeVec

	CurrentConstructingConnsCount *prometheus.GaugeVec
	CurrentAcquiredConnsCount     *prometheus.GaugeVec
	CurrentIdleConnsCount         *prometheus.GaugeVec
	// total = constructing + acquired + idle
	CurrentTotalConnsCount *prometheus.GaugeVec

	// NOTE: as of now (2025-11-20),
	// in pgx stats ConnAcquireDurationTotal == ConnAcquireConnAvailabilityWaitTimeTotal
	// (they are the same)
	ConnAcquireCount         *prometheus.CounterVec
	ConnAcquireDurationTotal xprommetric.HavingUnit[
		*prometheus.CounterVec,
		time.Duration,
		float64,
	]
	ConnAcquireCancelledCount                *prometheus.CounterVec
	ConnAcquireConnAvailabilityWaitTimeTotal xprommetric.HavingUnit[
		*prometheus.CounterVec,
		time.Duration,
		float64,
	]
}

func (pm *ConnPoolMetrics) Bind(p pool) MetricRecorder {
	return boundConnPoolMetrics{pm: pm, p: p}
}

// Clone creates a copy,
// which operates over the same metrics,
// but may store different state,
// which is useful for recording different
// pools and exporting to the same set of metrics.
func (pm ConnPoolMetrics) Clone() ConnPoolMetrics {
	pmCopy := ConnPoolMetrics{
		StructMetricGroup: pm.StructMetricGroup.Clone(),
		mls:               pm.mls,
		mlsPrecompiled:    pm.mlsPrecompiled,
		metricsStateMu:    &sync.Mutex{},
	}
	return pmCopy
}

type boundConnPoolMetrics struct {
	pm *ConnPoolMetrics
	p  pool
}

func (bpm boundConnPoolMetrics) Record() error {
	var errs []error

	config := bpm.p.Config() // config snapshot
	stats := bpm.p.Stat()    // stats snapshot
	var connAcquireInc uint64 = 0
	var connAcquireDurationInc time.Duration = 0
	var connAcquireCancelledInc uint64 = 0
	var connAcquireConnAvailabilityWaitTimeInc time.Duration = 0
	func() {
		bpm.pm.metricsStateMu.Lock()
		defer bpm.pm.metricsStateMu.Unlock()
		if stats.AcquireCount() > bpm.pm.connAcquireCount {
			connAcquireInc = uint64(stats.AcquireCount() - bpm.pm.connAcquireCount)
			bpm.pm.connAcquireCount = stats.AcquireCount()
		}
		if stats.AcquireDuration() > bpm.pm.connAcquireDurationTotal {
			connAcquireDurationInc = stats.AcquireDuration() - bpm.pm.connAcquireDurationTotal
			bpm.pm.connAcquireDurationTotal = stats.AcquireDuration()
		}
		if stats.CanceledAcquireCount() > bpm.pm.connAcquireCancelledCount {
			connAcquireCancelledInc = uint64(stats.CanceledAcquireCount() - bpm.pm.connAcquireCancelledCount)
			bpm.pm.connAcquireCancelledCount = stats.CanceledAcquireCount()
		}
		if stats.EmptyAcquireWaitTime() > bpm.pm.connAcquireConnAvailabilityWaitTimeTotal {
			connAcquireConnAvailabilityWaitTimeInc = stats.EmptyAcquireWaitTime() - bpm.pm.connAcquireConnAvailabilityWaitTimeTotal
			bpm.pm.connAcquireConnAvailabilityWaitTimeTotal = stats.EmptyAcquireWaitTime()
		}
	}()

	s := bpm.pm.Struct()
	ls := prometheus.Labels{}
	if bpm.pm.mls.HasValue() && !bpm.pm.mlsPrecompiled {
		for name, value := range bpm.pm.mls.Value().ToPrometheusLabels() {
			ls[name] = value
		}
	}

	gaugeSet := func(g *prometheus.GaugeVec, v float64) {
		if g != nil {
			g.With(ls).Set(v)
		}
	}
	counterSet := func(c *prometheus.CounterVec, inc float64) {
		if c != nil {
			c.With(ls).Add(inc)
		}
	}

	gaugeSet(s.ConfigMinConnsCount, float64(config.MinConns))
	gaugeSet(s.ConfigMaxConnsCount, float64(config.MaxConns))

	gaugeSet(s.CurrentConstructingConnsCount, float64(stats.ConstructingConns()))
	gaugeSet(s.CurrentAcquiredConnsCount, float64(stats.AcquiredConns()))
	gaugeSet(s.CurrentIdleConnsCount, float64(stats.IdleConns()))
	gaugeSet(s.CurrentTotalConnsCount, float64(stats.TotalConns()))

	counterSet(s.ConnAcquireCount, float64(connAcquireInc))
	{
		m := s.ConnAcquireDurationTotal
		if m != nil {
			v, err := m.Scale(connAcquireDurationInc)
			if err != nil {
				errs = append(errs, err)
			} else {
				m.Metric().With(ls).Add(v)
			}
		}
	}
	counterSet(s.ConnAcquireCancelledCount, float64(connAcquireCancelledInc))
	{
		m := s.ConnAcquireConnAvailabilityWaitTimeTotal
		if m != nil {
			v, err := m.Scale(connAcquireConnAvailabilityWaitTimeInc)
			if err != nil {
				errs = append(errs, err)
			} else {
				m.Metric().With(ls).Add(v)
			}
		}
	}
	return errors.Join(errs...)
}

func curryWithIn[T interface {
	CurryWith(prometheus.Labels) (T, error)
}](
	v *T,
	ls prometheus.Labels,
) error {
	var err error
	*v, err = (*v).CurryWith(ls)
	return err
}

type pool interface {
	Config() *pgxpool.Config
	Stat() *pgxpool.Stat
}

type MetricRecorder interface {
	Record() error
}
