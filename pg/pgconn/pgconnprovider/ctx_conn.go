package pgconnprovider

import (
	"context"
	"errors"

	"github.com/Deimvis/go-ext/go1.25/xoptional"
	"github.com/Deimvis-go/xpg/pg/internal/conns"
	"github.com/Deimvis-go/xpg/pg/internal/types"
	"github.com/Deimvis-go/xpg/pg/pgtrace"
)

// NewCtxConn constructs conn provider
// which provides connection from context (if any).
func NewCtxConn() types.ConnProvider {
	return ctxConn{}
}

type ctxConn struct{}

// interface guards
var _ types.ConnProvider = ctxConn{}
var _ types.ConnProviderMeta = ctxConn{}
var _ types.ConnProviderInternals = ctxConn{}

func (cc ctxConn) Acquire(ctx context.Context, mode types.ConnMode, opts ...types.ConnAcquireOption) (types.Conn, xoptional.T[types.ConnOwnership], error) {
	return cc.AcquireWithConfig(ctx, mode, conns.NewAcquireConfig(opts...))
}

func (cc ctxConn) AcquireManaged(ctx context.Context, mode types.ConnMode, opts ...types.ConnAcquireOption) (types.Conn, error) {
	return cc.AcquireManagedWithConfig(ctx, mode, conns.NewAcquireConfig(opts...))
}

func (cc ctxConn) AcquireWithConfig(ctx context.Context, mode types.ConnMode, cfg types.ConnAcquireConfig) (types.Conn, xoptional.T[types.ConnOwnership], error) {
	conn, err := cc.AcquireManagedWithConfig(ctx, mode, cfg)
	return conn, xoptional.New[types.ConnOwnership](), err
}

func (cc ctxConn) AcquireManagedWithConfig(ctx context.Context, mode types.ConnMode, cfg types.ConnAcquireConfig) (types.Conn, error) {
	if cfg.Meta.HasValue() {
		ctx = pgtrace.CtxWithConnAcquireMeta(ctx, cfg.Meta.Value())
	}
	// assuming it works fast, so ignoring timeouts from config
	conn, err := cc.acquire(ctx, mode)
	if err != nil && cfg.RWFallback && mode != types.ConnMode_RW && !errors.Is(err, context.DeadlineExceeded) {
		conn, err = cc.acquire(ctx, types.ConnMode_RW)
	}
	return conn, err
}

func (cc ctxConn) acquire(ctx context.Context, mode types.ConnMode) (types.Conn, error) {
	ctxKey := conns.CtxKey(mode)
	conn := ctx.Value(ctxKey)
	if conn == nil {
		return nil, conns.ErrNoCtxConn
	}
	return conn.(types.Conn), nil
}

func (cc ctxConn) Type() string {
	return ctxConnType
}

func (cc ctxConn) GenericType() string {
	return ctxConnType
}

func (cc ctxConn) AcquireType() xoptional.T[string] {
	return xoptional.New(ctxConnAcquireType)
}

const (
	ctxConnType        = "ctx_conn"
	ctxConnAcquireType = "context"
)
