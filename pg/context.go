package pg

import (
	"context"
	"crypto/rand"
	"encoding/binary"

	"github.com/Deimvis/go-ext/go1.25/xcheck/xmust"
	"github.com/Deimvis-go/xpg/pg/internal/conns"
	"github.com/Deimvis-go/xpg/pg/internal/types"
)

// DEPRECATED. TODO: replace with CtxConnOrAcquire(ctx, mode, connProvider)
// CtxConnOrPool returns connection from context, if any, and returns pool otherwise.
// In any case no connection management is required -
// connection from context is managed by one who placed it here;
// pool automatically releases connection after each query.
func CtxConnOrPool(ctx context.Context, mode ConnMode, pm PoolManager) ExtPG {
	// TODO: implement pg conn ownership transferring and reuse connection from main request
	var pg ExtPG
	ctxConn := CtxConn(ctx, mode)
	if ctxConn != nil {
		pg = ctxConn
	} else {
		pg = pm.GetPool(mode)
	}
	return pg
}

// CtxConn returns connection from context, if any, nil otherwise.
func CtxConn(ctx context.Context, mode ConnMode) types.Conn {
	ctxKey := conns.CtxKey(mode)
	conn := ctx.Value(ctxKey)
	if conn == nil {
		return nil
	}
	return conn.(types.Conn)
}

// TODO: return ConnOwnership on acquire and specify
// that it acquires persistent conn (not one-time),
// or better use ConnProvider

// CtxAcquireConn acquires connection and puts it into new context which is returned.
func CtxAcquireConn(ctx context.Context, mode ConnMode, pm PoolManager) context.Context {
	ctxKey := conns.CtxKey(mode)
	conn := xmust.Do(pm.GetPool(CM_RW).Acquire(ctx))
	return context.WithValue(ctx, ctxKey, conn)
}

// CtxReleaseConn releases connection from context.
func CtxReleaseConn(ctx *context.Context, mode ConnMode) {
	ctxKey := conns.CtxKey(mode)
	conn := (*ctx).Value(ctxKey).(types.PoolConn)
	conn.Release()
	*ctx = context.WithValue(*ctx, ctxKey, nil)
}

var CtxConnKey = conns.CtxKey

// ConnectSeedCtxKey returns context key
// for storing connect operation seed that is used
// to seed PostgreSQL servers load balancer.
// Since this seed is scoped to single connect
// operation, it makes load balancer to start
// working independently of concurrent connect
// operations, which provides useful guarantees,
// such that round robin load balancer won't return
// same host for this connect operation
// if retry happens (which is a normal case,
// since connect may be looking for certain
// server - leader/replica).
func ConnectSeedCtxKey() any {
	return connectSeedCtxKey{}
}

func CtxWithConnectSeed(ctx context.Context, seed uint64) context.Context {
	return context.WithValue(ctx, connectSeedCtxKey{}, seed)
}

func CtxConnectSeed(ctx context.Context) (uint64, bool) {
	seed := ctx.Value(connectSeedCtxKey{})
	if seed != nil {
		return seed.(uint64), true
	}
	return 0, false
}

func GenConnectSeed() uint64 {
	var seed uint64
	binary.Read(rand.Reader, binary.LittleEndian, &seed)
	return seed
}

type connectSeedCtxKey struct{}
