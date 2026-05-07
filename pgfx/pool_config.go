package pgfx

import (
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/Deimvis-go/xpg/pg"
)

type PgxpoolConfigROParams struct {
	fx.In

	ConnConfig   *pg.ConnConfig
	PoolConfig   *pg.PoolConfig    `name:"ro"`
	AfterConnect pg.AfterConnectFn `optional:"true"`

	Logger *zap.SugaredLogger

	QueryStartCounter            *prometheus.CounterVec   `name:"pg_query_start" optional:"true"`
	QueryFinishCounter           *prometheus.CounterVec   `name:"pg_query_finish" optional:"true"`
	QueryDurationHistogram       *prometheus.HistogramVec `name:"pg_query_duration" optional:"true"`
	QueryTimeoutHistogram        *prometheus.HistogramVec `name:"pg_query_timeout" optional:"true"`
	ConnAcquireStartCounter      *prometheus.CounterVec   `name:"pg_conn_acquire_start" optional:"true"`
	ConnAcquireFinishCounter     *prometheus.CounterVec   `name:"pg_conn_acquire_finish" optional:"true"`
	ConnAcquireDurationHistogram *prometheus.HistogramVec `name:"pg_conn_acquire_duration" optional:"true"`
	ConnAcquireTimeoutHistogram  *prometheus.HistogramVec `name:"pg_conn_acquire_timeout" optional:"true"`
	ConnectStartCounter          *prometheus.CounterVec   `name:"pg_connect_start" optional:"true"`
	ConnectFinishCounter         *prometheus.CounterVec   `name:"pg_connect_finish" optional:"true"`
	ConnectDurationHistogram     *prometheus.HistogramVec `name:"pg_connect_duration" optional:"true"`
	ConnectTimeoutHistogram      *prometheus.HistogramVec `name:"pg_connect_timeout" optional:"true"`
	ReleaseCounter               *prometheus.CounterVec   `name:"pg_conn_release" optional:"true"`
}

type PgxpoolConfigROResult struct {
	fx.Out

	PgxpoolConfig *pgxpool.Config `name:"ro"`
}

func NewPgxpoolConfigRO(p PgxpoolConfigROParams) PgxpoolConfigROResult {
	config := pg.NewPgxpoolConfig(pg.PgxpoolConfigParams{
		ConnConfig:   p.ConnConfig,
		PoolConfig:   p.PoolConfig,
		AfterConnect: p.AfterConnect,

		Logger: p.Logger,

		QueryStartCounter:            p.QueryStartCounter,
		QueryFinishCounter:           p.QueryFinishCounter,
		QueryDurationHistogram:       p.QueryDurationHistogram,
		QueryTimeoutHistogram:        p.QueryTimeoutHistogram,
		ConnAcquireStartCounter:      p.ConnAcquireStartCounter,
		ConnAcquireFinishCounter:     p.ConnAcquireFinishCounter,
		ConnAcquireDurationHistogram: p.ConnAcquireDurationHistogram,
		ConnAcquireTimeoutHistogram:  p.ConnAcquireTimeoutHistogram,
		ConnectStartCounter:          p.ConnectStartCounter,
		ConnectFinishCounter:         p.ConnectFinishCounter,
		ConnectDurationHistogram:     p.ConnectDurationHistogram,
		ConnectTimeoutHistogram:      p.ConnectTimeoutHistogram,
		ReleaseCounter:               p.ReleaseCounter,
	})
	config.ConnConfig.ValidateConnect = pgconn.ValidateConnectTargetSessionAttrsPreferStandby // prefer standy means use read-only if any available and any host otherwise
	return PgxpoolConfigROResult{PgxpoolConfig: config}
}

type PgxpoolConfigRWParams struct {
	fx.In

	ConnConfig   *pg.ConnConfig
	PoolConfig   *pg.PoolConfig    `name:"rw"`
	AfterConnect pg.AfterConnectFn `optional:"true"`

	Logger *zap.SugaredLogger

	QueryStartCounter            *prometheus.CounterVec   `name:"pg_query_start" optional:"true"`
	QueryFinishCounter           *prometheus.CounterVec   `name:"pg_query_finish" optional:"true"`
	QueryDurationHistogram       *prometheus.HistogramVec `name:"pg_query_duration" optional:"true"`
	QueryTimeoutHistogram        *prometheus.HistogramVec `name:"pg_query_timeout" optional:"true"`
	ConnAcquireStartCounter      *prometheus.CounterVec   `name:"pg_conn_acquire_start" optional:"true"`
	ConnAcquireFinishCounter     *prometheus.CounterVec   `name:"pg_conn_acquire_finish" optional:"true"`
	ConnAcquireDurationHistogram *prometheus.HistogramVec `name:"pg_conn_acquire_duration" optional:"true"`
	ConnAcquireTimeoutHistogram  *prometheus.HistogramVec `name:"pg_conn_acquire_timeout" optional:"true"`
	ConnectStartCounter          *prometheus.CounterVec   `name:"pg_connect_start" optional:"true"`
	ConnectFinishCounter         *prometheus.CounterVec   `name:"pg_connect_finish" optional:"true"`
	ConnectDurationHistogram     *prometheus.HistogramVec `name:"pg_connect_duration" optional:"true"`
	ConnectTimeoutHistogram      *prometheus.HistogramVec `name:"pg_connect_timeout" optional:"true"`
	ReleaseCounter               *prometheus.CounterVec   `name:"pg_conn_release" optional:"true"`
}

type PgxpoolConfigRWResult struct {
	fx.Out

	PgxpoolConfig *pgxpool.Config `name:"rw"`
}

func NewPgxpoolConfigRW(p PgxpoolConfigRWParams) PgxpoolConfigRWResult {
	// disable load balancing, since there is only single master
	cf := p.ConnConfig
	cf.HostsLoadBalancing = nil
	cf.ServersLoadBalancing = nil

	config := pg.NewPgxpoolConfig(pg.PgxpoolConfigParams{
		ConnConfig:   p.ConnConfig,
		PoolConfig:   p.PoolConfig,
		AfterConnect: p.AfterConnect,

		Logger: p.Logger,

		QueryStartCounter:            p.QueryStartCounter,
		QueryFinishCounter:           p.QueryFinishCounter,
		QueryDurationHistogram:       p.QueryDurationHistogram,
		QueryTimeoutHistogram:        p.QueryTimeoutHistogram,
		ConnAcquireStartCounter:      p.ConnAcquireStartCounter,
		ConnAcquireFinishCounter:     p.ConnAcquireFinishCounter,
		ConnAcquireDurationHistogram: p.ConnAcquireDurationHistogram,
		ConnAcquireTimeoutHistogram:  p.ConnAcquireTimeoutHistogram,
		ConnectStartCounter:          p.ConnectStartCounter,
		ConnectFinishCounter:         p.ConnectFinishCounter,
		ConnectDurationHistogram:     p.ConnectDurationHistogram,
		ConnectTimeoutHistogram:      p.ConnectTimeoutHistogram,
		ReleaseCounter:               p.ReleaseCounter,
	})
	config.ConnConfig.ValidateConnect = pgconn.ValidateConnectTargetSessionAttrsReadWrite
	return PgxpoolConfigRWResult{PgxpoolConfig: config}
}
