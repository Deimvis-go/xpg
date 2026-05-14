package pgfw

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/Deimvis/go-ext/go1.25/xoptional"
	"github.com/Deimvis-go/logs/logs"
	"github.com/Deimvis-go/xpg/pg"
	"github.com/Deimvis-go/xpg/pg/pgconn"
	"github.com/Deimvis-go/xpg/pg/pgconn/pgconnprovider"
	"github.com/Deimvis-go/xpg/pg/pgpool"
	"github.com/Deimvis-go/xprometheus/prom"
)

func NewStorageBase(pm pg.PoolManager, lg logs.KVCtxLogger, opts ...StorageBaseOption) *StorageBase {
	cfg := storageBaseCfg{}
	for _, opt := range opts {
		opt(&cfg)
	}

	connProvider := pgconnprovider.NewFallbacked(
		pgconnprovider.NewCtxConn(),
		pgpool.AsOneTimeConnProvider(pm),
	).WithHooks(pgconnprovider.FallbackedHooks{
		OnAttemptStart: func(
			ectx pgconnprovider.EventContext,
			attempt pgconnprovider.FallbackedAttemptState,
		) error {
			if attempt.Index > 0 {
				lg.Info(ectx.Context(), pgConnAcquireFallbackLogMsg,
					"ind", attempt.Index,
					"acquire_type", attempt.AcquireTypeOr(prom.LabelUnknown))
			}
			return nil
		},
	})
	for _, promMetrics := range cfg.fbConnProviderPromMetrics {
		connProvider.Stats().RegisterPrometheusExport(promMetrics)
	}
	return &StorageBase{
		conn: pgconnprovider.SugarSilent(pgconnprovider.Silence(
			connProvider,
			func(err error) { panic(err) },
		)),
		pm: pm,
	}
}

type StorageBase struct {
	conn pgconn.SugaredSilentProvider
	pm   pg.PoolManager
}

// TODO: add shortcut so that one can:
//   conn = StorageBase.ManagedConn()
//   conn.RO()

func (b *StorageBase) ROConn(ctx context.Context) pg.Conn {
	return b.conn.ManagedRO(ctx)
}

func (b *StorageBase) RWConn(ctx context.Context) pg.Conn {
	return b.conn.ManagedRW(ctx)
}

// RConn attempt to acquire RO conn first and RW second.
func (b *StorageBase) RConn(ctx context.Context) pg.Conn {
	return b.conn.ManagedR(ctx)
}

func (b *StorageBase) OwneableROConn(ctx context.Context) (pg.Conn, xoptional.T[pg.ConnOwnership]) {
	return b.conn.RO(ctx)
}

func (b *StorageBase) OwneableRWConn(ctx context.Context) (pg.Conn, xoptional.T[pg.ConnOwnership]) {
	return b.conn.RW(ctx)
}

// RConn attempt to acquire RO conn first and RW second.
func (b *StorageBase) OwneableRConn(ctx context.Context) (pg.Conn, xoptional.T[pg.ConnOwnership]) {
	return b.conn.R(ctx)
}

// TODO: add Lazy* for conns
// TODO: add NonOneTime* for conns (or Persistent*)

func (b *StorageBase) BeginTx(ctx context.Context) (pgx.Tx, error) {
	// TODO: figure out how to get ownable connection and
	// put its ownership inside Tx ownership,
	// so connection is "freed" along with transaction
	return b.conn.ManagedRW(ctx).Begin(ctx)
}

func (b *StorageBase) BeginCustomTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error) {
	return b.conn.ManagedRW(ctx).BeginTx(ctx, opts)
}

const (
	pgConnAcquireFallbackLogMsg = "PG Conn Acquire fallbacks to next conn provider"
)
