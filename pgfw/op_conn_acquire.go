package pgfw

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Deimvis/go-ext/go1.25/xoptional"
	"github.com/Deimvis-go/xpg/pg"
	"github.com/Deimvis-go/xpg/pg/pgtrace"
)

// pgx based interfaces
type conn = *pgxpool.Conn // TODO: hide pgx.Conn under proper interface, but allow to cast to *pgx.Conn
// TODO: consider adding functionality to wrap pgxpool.Pool into this local pool interface, so connection mode
// is propagated automatically into trace
type pool interface {
	Acquire(ctx context.Context) (conn, error)
}

type AcquireConnOption func(*acquireConnCfg)

func AcquireConn(p pool, ctx context.Context, opts ...AcquireConnOption) (pg.Conn, error) {
	cfg := acquireConnCfg{}
	for _, opt := range opts {
		opt(&cfg)
	}

	if cfg.meta.HasValue() {
		ctx = pgtrace.CtxWithConnAcquireMeta(ctx, cfg.meta.Value())
	}
	return p.Acquire(ctx)
}

func WithMeta(m pgtrace.ConnAcquireMeta) AcquireConnOption {
	return func(c *acquireConnCfg) {
		c.meta.SetValue(m)
	}
}

type acquireConnCfg struct {
	meta xoptional.T[pgtrace.ConnAcquireMeta]
}
