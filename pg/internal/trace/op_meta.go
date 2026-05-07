package trace

import (
	"context"

	"github.com/Deimvis/go-ext/go1.25/xoptional"
	"github.com/Deimvis-go/xpg/pg/internal/types"
)

type ConnAcquireMeta = types.ConnAcquireTracingMeta

func CtxWithConnAcquireMeta(ctx context.Context, v ConnAcquireMeta) context.Context {
	return context.WithValue(ctx, connAcquireMetaCtxKey{}, v)
}

func CtxConnAcquireMeta(ctx context.Context) (ConnAcquireMeta, bool) {
	v := ctx.Value(connAcquireMetaCtxKey{})
	if v == nil {
		return ConnAcquireMeta{}, false
	}
	return v.(ConnAcquireMeta), true
}

type QueryMeta struct {
	QueryName xoptional.T[string]
	ConnMode  xoptional.T[types.ConnMode]
}

func CtxWithQueryMeta(ctx context.Context, v QueryMeta) context.Context {
	return context.WithValue(ctx, queryMetaCtxKey{}, v)
}

func CtxQueryMeta(ctx context.Context) (QueryMeta, bool) {
	v := ctx.Value(queryMetaCtxKey{})
	if v == nil {
		return QueryMeta{}, false
	}
	return v.(QueryMeta), true
}

type (
	connAcquireMetaCtxKey struct{}
	queryMetaCtxKey       struct{}
	connectMetaCtxKey     struct{}
)
