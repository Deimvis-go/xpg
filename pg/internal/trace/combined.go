package trace

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TODO: maybe rename to ChainedTracer (pros: more specific type of "combining", cons: "chained" sounds like it's a single tracer which is chained, not a chain of tracers)

func NewCombinedTracer(tracers ...Tracer) *CombinedTracer {
	return &CombinedTracer{tracers: tracers}
}

type CombinedTracer struct {
	tracers []Tracer
}

func (t *CombinedTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	for _, t := range t.tracers {
		ctx = t.TraceQueryStart(ctx, conn, data)
	}
	return ctx
}

func (t *CombinedTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	for _, t := range t.tracers {
		t.TraceQueryEnd(ctx, conn, data)
	}
}

func (t *CombinedTracer) TraceAcquireStart(ctx context.Context, pool *pgxpool.Pool, data pgxpool.TraceAcquireStartData) context.Context {
	for _, t := range t.tracers {
		ctx = t.TraceAcquireStart(ctx, pool, data)
	}
	return ctx
}

func (t *CombinedTracer) TraceAcquireEnd(ctx context.Context, pool *pgxpool.Pool, data pgxpool.TraceAcquireEndData) {
	for _, t := range t.tracers {
		t.TraceAcquireEnd(ctx, pool, data)
	}
}

func (t *CombinedTracer) TraceRelease(pool *pgxpool.Pool, data pgxpool.TraceReleaseData) {
	for _, t := range t.tracers {
		t.TraceRelease(pool, data)
	}
}

func (t *CombinedTracer) TraceConnectStart(ctx context.Context, data pgx.TraceConnectStartData) context.Context {
	for _, t := range t.tracers {
		ctx = t.TraceConnectStart(ctx, data)
	}
	return ctx
}

func (t *CombinedTracer) TraceConnectEnd(ctx context.Context, data pgx.TraceConnectEndData) {
	for _, t := range t.tracers {
		t.TraceConnectEnd(ctx, data)
	}
}
