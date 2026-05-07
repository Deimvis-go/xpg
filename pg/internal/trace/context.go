package trace

import (
	"context"

	"github.com/google/uuid"
)

func ctxValueOr[T any](ctx context.Context, key any, fallback T) T {
	v := ctx.Value(key)
	if v != nil {
		return v.(T)
	}
	return fallback
}

func ctxValueOrSet[T any](ctx context.Context, key any, fb T) (context.Context, T) {
	v := ctx.Value(key)
	if v != nil {
		return ctx, v.(T)
	}
	ctx = context.WithValue(ctx, key, fb)
	return ctx, fb
}

func ctxValueOrSetLazy[T any](ctx context.Context, key any, lazyFb func() T) (context.Context, T) {
	v := ctx.Value(key)
	if v != nil {
		return ctx, v.(T)
	}
	fb := lazyFb()
	ctx = context.WithValue(ctx, key, fb)
	return ctx, fb
}

func ctxWithValueNX(ctx context.Context, key any, value any) context.Context {
	if ctx.Value(key) != nil {
		return ctx
	}
	return context.WithValue(ctx, key, value)
}

func ctxWithValueNXLazy(ctx context.Context, key any, lazyValue func() any) context.Context {
	if ctx.Value(key) != nil {
		return ctx
	}
	return context.WithValue(ctx, key, lazyValue())
}

func genQueryId() string {
	return uuid.New().String()
}

func genAcquireId() string {
	return uuid.New().String()
}

func genConnectId() string {
	return uuid.New().String()
}

// TODO: consider creating internal meta for operation,
// it will reduce number of keys in context

type queryIdCtxKey struct{}
type queryStartCtxKey struct{}
type acquireIdCtxKey struct{}
type acquireStartCtxKey struct{}
type connectIdCtxKey struct{}
type connectStartCtxKey struct{}
