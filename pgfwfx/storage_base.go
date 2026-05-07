package pgfwfx

import (
	"github.com/Deimvis-go/logs/logs"
	"github.com/Deimvis-go/xpg/pg"
	"github.com/Deimvis-go/xpg/pg/pgconn/pgconnprovider"
	"github.com/Deimvis-go/xpg/pgfw"
	"go.uber.org/fx"
)

type StorageBaseParams struct {
	fx.In

	Pm pg.PoolManager
	Lg logs.KVCtxLogger

	CPPromMetrics pgconnprovider.FallbackedPrometheusMetrics `name:"pg_storage_base" optional:"true"`
}

func NewStorageBase(p StorageBaseParams) *pgfw.StorageBase {
	opts := []pgfw.StorageBaseOption{}
	if p.CPPromMetrics != nil {
		opts = append(opts, pgfw.StorageBaseWithConnProviderPromMetrics(p.CPPromMetrics))
	}
	return pgfw.NewStorageBase(p.Pm, p.Lg, opts...)
}
