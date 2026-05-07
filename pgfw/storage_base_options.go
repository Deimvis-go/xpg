package pgfw

import "github.com/Deimvis-go/xpg/pg/pgconn/pgconnprovider"

type StorageBaseOption func(*storageBaseCfg)

func StorageBaseWithConnProviderPromMetrics(m pgconnprovider.FallbackedPrometheusMetrics) StorageBaseOption {
	return func(c *storageBaseCfg) {
		c.fbConnProviderPromMetrics = append(c.fbConnProviderPromMetrics, m)
	}
}

type storageBaseCfg struct {
	fbConnProviderPromMetrics []pgconnprovider.FallbackedPrometheusMetrics
}
