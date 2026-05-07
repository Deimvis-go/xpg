package pgpool

import (
	"context"
	"errors"
	"time"

	"github.com/Deimvis/go-ext/go1.25/xcontext"
	"github.com/Deimvis/go-ext/go1.25/xoptional"
	"github.com/Deimvis-go/xpg/pg/internal/conns"
	"github.com/Deimvis-go/xpg/pg/internal/translate"
	"github.com/Deimvis-go/xpg/pg/internal/types"
	"github.com/Deimvis-go/xpg/pg/pgtrace"
)

// AsOneTimeConnProvider returns ConnProvider from PoolManager,
// which provides a one-time connect on each command -
// it means that on each command new connect is acquired from pool,
// and after command if finished, it is
// automatically released back to the pool.
// TODO: support AcquireNonOneTime option to hint
// acquire connection with ownership
// (this provider would return error).
func AsOneTimeConnProvider(pm types.PoolManager) types.ConnProvider {
	return asOneTimeConnProvider{pm: pm}
}

// AsPersistentConnProvider returns ConnProvider from Poolmanager,
// which provides a persistent (non one-time) connect.
func AsPersistentConnProvider(pm types.PoolManager) types.ConnProvider {
	return asPersistentConnProvider{pm: pm}
}

type asOneTimeConnProvider struct {
	pm types.PoolManager
}

// interface guards
var _ types.ConnProvider = asOneTimeConnProvider{}
var _ types.ConnProviderMeta = asOneTimeConnProvider{}
var _ types.ConnProviderInternals = asOneTimeConnProvider{}

func (acp asOneTimeConnProvider) Acquire(ctx context.Context, mode types.ConnMode, opts ...types.ConnAcquireOption) (types.Conn, xoptional.T[types.ConnOwnership], error) {
	return acp.AcquireWithConfig(ctx, mode, conns.NewAcquireConfig(opts...))
}

func (acp asOneTimeConnProvider) AcquireManaged(ctx context.Context, mode types.ConnMode, opts ...types.ConnAcquireOption) (types.Conn, error) {
	return acp.AcquireManagedWithConfig(ctx, mode, conns.NewAcquireConfig(opts...))
}

func (acp asOneTimeConnProvider) AcquireWithConfig(ctx context.Context, mode types.ConnMode, cfg types.ConnAcquireConfig) (types.Conn, xoptional.T[types.ConnOwnership], error) {
	conn, err := acp.AcquireManagedWithConfig(ctx, mode, cfg)
	return conn, xoptional.New[types.ConnOwnership](), err
}

func (acp asOneTimeConnProvider) AcquireManagedWithConfig(ctx context.Context, mode types.ConnMode, cfg types.ConnAcquireConfig) (types.Conn, error) {
	// assuming it works fast, so ignoring timeouts from config
	if cfg.Meta.HasValue() {
		ctx = pgtrace.CtxWithConnAcquireMeta(ctx, cfg.Meta.Value())
	}
	conn, err := acp.acquire(ctx, mode)
	if err != nil && cfg.RWFallback && mode != types.ConnMode_RW && !errors.Is(err, context.DeadlineExceeded) {
		conn, err = acp.acquire(ctx, types.ConnMode_RW)
	}
	return conn, err
}

func (acp asOneTimeConnProvider) acquire(_ context.Context, mode types.ConnMode) (types.Conn, error) {
	pool := acp.pm.GetPool(mode)
	return translate.FromPgxPoolToConn(pool, mode), nil
}

func (acp asOneTimeConnProvider) Type() string {
	return asOnetimeType
}

func (acp asOneTimeConnProvider) GenericType() string {
	return asOnetimeType
}

func (acp asOneTimeConnProvider) AcquireType() xoptional.T[string] {
	return xoptional.New(asOnetimeAcquireType)
}

type asPersistentConnProvider struct {
	pm types.PoolManager
}

// interface guards
var _ types.ConnProvider = asPersistentConnProvider{}
var _ types.ConnProviderMeta = asPersistentConnProvider{}
var _ types.ConnProviderInternals = asPersistentConnProvider{}

func (acp asPersistentConnProvider) Acquire(ctx context.Context, mode types.ConnMode, opts ...types.ConnAcquireOption) (types.Conn, xoptional.T[types.ConnOwnership], error) {
	return acp.AcquireWithConfig(ctx, mode, conns.NewAcquireConfig(opts...))
}

func (acp asPersistentConnProvider) AcquireManaged(ctx context.Context, mode types.ConnMode, opts ...types.ConnAcquireOption) (types.Conn, error) {
	return acp.AcquireManagedWithConfig(ctx, mode, conns.NewAcquireConfig(opts...))
}

func (acp asPersistentConnProvider) AcquireWithConfig(ctx context.Context, mode types.ConnMode, cfg types.ConnAcquireConfig) (types.Conn, xoptional.T[types.ConnOwnership], error) {
	if cfg.Timeout.HasValue() {
		cancel := xcontext.WithTimeoutIn(&ctx, cfg.Timeout.Value())
		defer cancel()
	}
	if cfg.Meta.HasValue() {
		ctx = pgtrace.CtxWithConnAcquireMeta(ctx, cfg.Meta.Value())
	}

	conn, own, err := acp.acquire(ctx, mode, cfg.AttemptTimeout)
	if err != nil && cfg.RWFallback && mode != types.ConnMode_RW && !errors.Is(err, context.DeadlineExceeded) {
		conn, own, err = acp.acquire(ctx, types.ConnMode_RW, cfg.AttemptTimeout)
	}
	return conn, xoptional.New(own), err
}

func (acp asPersistentConnProvider) AcquireManagedWithConfig(ctx context.Context, mode types.ConnMode, cfg types.ConnAcquireConfig) (types.Conn, error) {
	return nil, errors.New("no managed conns")
}

func (acp asPersistentConnProvider) acquire(
	ctx context.Context,
	mode types.ConnMode,
	timeout xoptional.T[time.Duration],
) (types.Conn, types.ConnOwnership, error) {
	if timeout.HasValue() {
		cancel := xcontext.WithTimeoutIn(&ctx, timeout.Value())
		defer cancel()
	}
	pool := acp.pm.GetPool(mode)
	pgxconn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, nil, err
	}
	conn, own := translate.FromPgxpoolConnToConn(pgxconn, mode)
	return conn, own, nil
}

func (acp asPersistentConnProvider) Type() string {
	return asPersistentType
}

func (acp asPersistentConnProvider) GenericType() string {
	return asPersistentType
}

func (acp asPersistentConnProvider) AcquireType() xoptional.T[string] {
	return xoptional.New(asPersistentAcquireType)
}

const (
	asOnetimeType        = "onetime_from_pool"
	asOnetimeAcquireType = "onetime_from_pool"

	asPersistentType        = "persistent_from_pool"
	asPersistentAcquireType = "persistent_from_pool"
)
