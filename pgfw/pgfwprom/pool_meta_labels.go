package pgfwprom

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/Deimvis/go-ext/go1.25/xoptional"
)

// TODO: impl when pool is generalized to interface
// func NewPoolMetaLabels(p pg.ConnPool) PoolMetaLabels

type ConnPoolMetaLabels struct {
	PoolName xoptional.T[string]
	ConnMode xoptional.T[string]
}

func (mpl ConnPoolMetaLabels) ToPrometheusLabels() prometheus.Labels {
	fields := map[string]xoptional.T[string]{
		"pool_name": mpl.PoolName,
		"conn_mode": mpl.ConnMode,
	}
	ls := make(prometheus.Labels, len(fields))
	for name, value := range fields {
		if value.HasValue() {
			ls[name] = value.Value()
		}
	}
	return ls
}
