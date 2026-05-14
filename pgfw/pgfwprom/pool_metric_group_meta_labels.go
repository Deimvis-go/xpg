package pgfwprom

import (
	"errors"
	"time"

	"github.com/Deimvis/go-ext/go1.25/ext"
	"github.com/Deimvis-go/xprometheus/prom/prommetric"
)

// SetMetaLebls allows to pre-set constant meta labels.
// These meta labels should be registered in metric
// as variable labels.
// If meta labels are precompiled, it will return error.
func (pm *ConnPoolMetrics) SetMetaLabels(mls ConnPoolMetaLabels) error {
	if pm.mlsPrecompiled {
		return errors.New("meta labels are precompiled")
	}
	pm.mls.SetValue(mls)
	return nil
}

// MetaLabels returns meta labels curried
// into metrics.
// Note that if meta labels are attached
// using prometheus constant labels,
// then they won't be observed by ConnPoolMetrics,
// and won't be returned here.
func (pm ConnPoolMetrics) MetaLabels() (ConnPoolMetaLabels, bool) {
	if pm.mls.HasValue() {
		return pm.mls.Value(), true
	}
	return ConnPoolMetaLabels{}, false
}

// It must be called at most ones, because
// it uses labels currying.
// After successful currying existing fields
// are replaced with their curried copies,
// so struct metrics that would be post-initialized
// won't have meta labels, unless one
// curry them manually.
func (pm *ConnPoolMetrics) PrecompileMetaLabels() error {
	if !pm.mls.HasValue() {
		return errors.New("meta labels not set")
	}
	if pm.mlsPrecompiled {
		return errors.New("meta labels already precompiled")
	}
	pm.mlsPrecompiled = true
	ls := pm.mls.Value().ToPrometheusLabels()
	s := *pm.Struct()
	err := ext.UntilFirstErr(
		func() error {
			if s.ConfigMinConnsCount == nil {
				return nil
			}
			return curryWithIn(&s.ConfigMinConnsCount, ls)
		},
		func() error {
			if s.ConfigMaxConnsCount == nil {
				return nil
			}
			return curryWithIn(&s.ConfigMaxConnsCount, ls)
		},

		func() error {
			if s.CurrentConstructingConnsCount == nil {
				return nil
			}
			return curryWithIn(&s.CurrentConstructingConnsCount, ls)
		},
		func() error {
			if s.CurrentAcquiredConnsCount == nil {
				return nil
			}
			return curryWithIn(&s.CurrentAcquiredConnsCount, ls)
		},
		func() error {
			if s.CurrentIdleConnsCount == nil {
				return nil
			}
			return curryWithIn(&s.CurrentIdleConnsCount, ls)
		},
		func() error {
			if s.CurrentTotalConnsCount == nil {
				return nil
			}
			return curryWithIn(&s.CurrentTotalConnsCount, ls)
		},

		func() error {
			if s.ConnAcquireCount == nil {
				return nil
			}
			return curryWithIn(&s.ConnAcquireCount, ls)
		},
		func() error {
			if s.ConnAcquireDurationTotal == nil {
				return nil
			}
			new, err := s.ConnAcquireDurationTotal.Metric().CurryWith(ls)
			if err != nil {
				return err
			}
			s.ConnAcquireDurationTotal = prommetric.NewHavingUnit(
				new,
				prommetric.UnitScaler[time.Duration, float64](s.ConnAcquireDurationTotal),
			)
			return nil
		},
		func() error {
			if s.ConnAcquireCancelledCount == nil {
				return nil
			}
			return curryWithIn(&s.ConnAcquireCancelledCount, ls)
		},
		func() error {
			if s.ConnAcquireConnAvailabilityWaitTimeTotal == nil {
				return nil
			}
			new, err := s.ConnAcquireConnAvailabilityWaitTimeTotal.Metric().CurryWith(ls)
			if err != nil {
				return err
			}
			s.ConnAcquireConnAvailabilityWaitTimeTotal = prommetric.NewHavingUnit(
				new,
				prommetric.UnitScaler[time.Duration, float64](s.ConnAcquireConnAvailabilityWaitTimeTotal),
			)
			return nil
		},
	)
	if err != nil {
		return err
	}
	*pm.Struct() = s
	return nil
}

// WithMetaLabels is a shortcut for .Clone().SetMetaLabels(mls)
func (pm ConnPoolMetrics) WithMetaLabels(mls ConnPoolMetaLabels) ConnPoolMetrics {
	pmCopy := pm.Clone()
	pmCopy.SetMetaLabels(mls)
	return pmCopy
}

// WithMetaLabelsPrecompiled is a shortcut for .Clone().SetMetaLabels(mls).PrecompileMetaLabels()
func (pm ConnPoolMetrics) WithPrecompiledMetaLabels(mls ConnPoolMetaLabels) (*ConnPoolMetrics, error) {
	pmCopy := pm.Clone()
	pmCopy.SetMetaLabels(mls)
	err := pmCopy.PrecompileMetaLabels()
	if err != nil {
		return nil, err
	}
	return &pmCopy, nil
}
