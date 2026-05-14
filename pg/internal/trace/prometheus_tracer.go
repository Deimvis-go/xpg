package trace

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/Deimvis/go-ext/go1.25/xoptional"
	"github.com/Deimvis-go/xpg/pg/pgconn"
	"github.com/Deimvis-go/xprometheus/prom"
	"go.uber.org/fx"
)

type PrometheusTracerParams struct {
	fx.In
}

func NewPrometheusTracer(
	queryStartCounter *prometheus.CounterVec,
	queryFinishCounter *prometheus.CounterVec,
	queryDurationHistogram *prometheus.HistogramVec,
	queryTimeoutHistogram *prometheus.HistogramVec,
	connAcquireStartCounter *prometheus.CounterVec,
	connAcquireFinishCounter *prometheus.CounterVec,
	connAcquireDurationHistogram *prometheus.HistogramVec,
	connAcquireTimeoutHistogram *prometheus.HistogramVec,
	connectStartCounter *prometheus.CounterVec,
	connectFinishCounter *prometheus.CounterVec,
	connectDurationHistogram *prometheus.HistogramVec,
	connectTimeoutHistogram *prometheus.HistogramVec,
	connReleaseCounter *prometheus.CounterVec,
) *PrometheusTracer {
	return &PrometheusTracer{
		queryStartCounter:            queryStartCounter,
		queryFinishCounter:           queryFinishCounter,
		queryDurationHistogram:       queryDurationHistogram,
		queryTimeoutHistogram:        queryTimeoutHistogram,
		connAcquireStartCounter:      connAcquireStartCounter,
		connAcquireFinishCounter:     connAcquireFinishCounter,
		connAcquireDurationHistogram: connAcquireDurationHistogram,
		connAcquireTimeoutHistogram:  connAcquireTimeoutHistogram,
		connectStartCounter:          connectStartCounter,
		connectFinishCounter:         connectFinishCounter,
		connectDurationHistogram:     connectDurationHistogram,
		connectTimeoutHistogram:      connectTimeoutHistogram,
		connReleaseCounter:           connReleaseCounter,
	}
}

type PrometheusTracer struct {
	// variable labels: ["query_name", "conn_mode"]
	queryStartCounter     *prometheus.CounterVec
	queryTimeoutHistogram *prometheus.HistogramVec
	// variable labels: ["query_name", "conn_mode", "error"]
	queryFinishCounter     *prometheus.CounterVec
	queryDurationHistogram *prometheus.HistogramVec

	// variable labels: ["conn_mode"]
	connAcquireStartCounter     *prometheus.CounterVec
	connAcquireTimeoutHistogram *prometheus.HistogramVec
	// variable labels: ["conn_mode", "error"]
	connAcquireFinishCounter     *prometheus.CounterVec
	connAcquireDurationHistogram *prometheus.HistogramVec

	// variable labels: ["conn_mode"]
	connectStartCounter     *prometheus.CounterVec
	connectTimeoutHistogram *prometheus.HistogramVec
	// variable labels: ["conn_mode", "error"]
	connectFinishCounter     *prometheus.CounterVec
	connectDurationHistogram *prometheus.HistogramVec

	// variable labels: ["conn_mode"]
	connReleaseCounter *prometheus.CounterVec
}

func (t *PrometheusTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	ctx, _ = ctxValueOrSetLazy(ctx, queryStartCtxKey{}, time.Now)
	meta := ctxValueOr(ctx, queryMetaCtxKey{}, QueryMeta{})
	labels := prometheus.Labels{
		"query_name": xoptional.ValueOr(meta.QueryName, prom.LabelUnknown),
		"conn_mode":  xoptional.ValueCastOr(meta.ConnMode, pgconn.Mode.String, prom.LabelUnknown),
	}
	if t.queryStartCounter != nil {
		t.queryStartCounter.With(labels).Inc()
	}
	if dedl, ok := ctx.Deadline(); t.queryTimeoutHistogram != nil && ok {
		t.queryTimeoutHistogram.With(labels).Observe(float64(time.Until(dedl)))
	}
	return ctx
}

func (t *PrometheusTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	start := ctx.Value(queryStartCtxKey{}).(time.Time)
	elapsed := time.Since(start).Seconds()

	meta := ctxValueOr(ctx, queryMetaCtxKey{}, QueryMeta{})
	errStr := ""
	if data.Err != nil {
		if errors.Is(data.Err, context.DeadlineExceeded) {
			errStr = kContextDeadlineExceededLabel
		} else {
			errStr = fmt.Sprintf("<%T>", data.Err)
		}
	}
	labels := prometheus.Labels{
		"query_name": xoptional.ValueOr(meta.QueryName, prom.LabelUnknown),
		"conn_mode":  xoptional.ValueCastOr(meta.ConnMode, pgconn.Mode.String, prom.LabelUnknown),
		"error":      errStr,
	}
	if t.queryFinishCounter != nil {
		t.queryFinishCounter.With(labels).Inc()
	}
	if t.queryDurationHistogram != nil {
		t.queryDurationHistogram.With(labels).Observe(elapsed)
	}
}

func (t *PrometheusTracer) TraceAcquireStart(ctx context.Context, p *pgxpool.Pool, _ pgxpool.TraceAcquireStartData) context.Context {
	ctx, _ = ctxValueOrSetLazy(ctx, acquireStartCtxKey{}, time.Now)
	meta := ctxValueOr(ctx, connAcquireMetaCtxKey{}, ConnAcquireMeta{})
	labels := prometheus.Labels{
		"conn_mode": xoptional.ValueCastOr(meta.ConnMode, pgconn.Mode.String, prom.LabelUnknown),
	}
	if t.connAcquireStartCounter != nil {
		t.connAcquireStartCounter.With(labels).Inc()
	}
	if dedl, ok := ctx.Deadline(); t.connAcquireTimeoutHistogram != nil && ok {
		t.connAcquireTimeoutHistogram.With(labels).Observe(float64(time.Until(dedl)))
	}
	return ctx
}

func (t *PrometheusTracer) TraceAcquireEnd(ctx context.Context, p *pgxpool.Pool, data pgxpool.TraceAcquireEndData) {
	start := ctx.Value(acquireStartCtxKey{}).(time.Time)
	elapsed := time.Since(start).Seconds()

	meta := ctxValueOr(ctx, connAcquireMetaCtxKey{}, ConnAcquireMeta{})
	errStr := ""
	if data.Err != nil {
		if errors.Is(data.Err, context.DeadlineExceeded) {
			errStr = kContextDeadlineExceededLabel
		} else {
			errStr = fmt.Sprintf("<%T>", data.Err)
		}
	}
	labels := prometheus.Labels{
		"conn_mode": xoptional.ValueCastOr(meta.ConnMode, pgconn.Mode.String, prom.LabelUnknown),
		"error":     errStr,
	}
	if t.connAcquireFinishCounter != nil {
		t.connAcquireFinishCounter.With(labels).Inc()
	}
	if t.connAcquireDurationHistogram != nil {
		t.connAcquireDurationHistogram.With(labels).Observe(elapsed)
	}
}

func (t *PrometheusTracer) TraceRelease(pool *pgxpool.Pool, d pgxpool.TraceReleaseData) {
	if t.connReleaseCounter != nil {
		// TODO: find proper approach for conn mode determination
		// funcEq := func(fn1, fn2 any) bool {
		// 	return reflect.ValueOf(fn1).Pointer() == reflect.ValueOf(fn2).Pointer()
		// }
		// if funcEq(d.Conn.Config().ValidateConnect, pgconn.ValidateConnectTargetSessionAttrsPreferStandby) ||
		// 	funcEq(d.Conn.Config().ValidateConnect, pgconn.ValidateConnectTargetSessionAttrsStandby) ||
		// 	funcEq(d.Conn.Config().ValidateConnect, pgconn.ValidateConnectTargetSessionAttrsReadOnly) {
		// 	connMode := "ro"
		// } else {
		// 	connMode := "rw"
		// }
		labels := prometheus.Labels{
			"conn_mode": prom.LabelUnknown,
		}
		t.connReleaseCounter.With(labels).Inc()
	}
}

func (t *PrometheusTracer) TraceConnectStart(ctx context.Context, _ pgx.TraceConnectStartData) context.Context {
	ctx, _ = ctxValueOrSetLazy(ctx, connectStartCtxKey{}, time.Now)
	// if connect is a part of acquire, we may use acquire meta
	ameta := ctxValueOr(ctx, connectMetaCtxKey{}, ConnAcquireMeta{})
	// TODO: support connect-specific meta
	// meta := ctxValueOr(ctx, connectMetaCtxKey{}, ConnectMeta{})
	labels := prometheus.Labels{
		"conn_mode": xoptional.ValueCastOr(ameta.ConnMode, pgconn.Mode.String, prom.LabelUnknown),
	}
	if t.connectStartCounter != nil {
		t.connectStartCounter.With(labels).Inc()
	}
	if dedl, ok := ctx.Deadline(); t.connectTimeoutHistogram != nil && ok {
		t.connectTimeoutHistogram.With(labels).Observe(float64(time.Until(dedl)))
	}
	return ctx
}

func (t *PrometheusTracer) TraceConnectEnd(ctx context.Context, data pgx.TraceConnectEndData) {
	start := ctx.Value(connectStartCtxKey{}).(time.Time)
	elapsed := time.Since(start).Seconds()
	// if connect is a part of acquire, we may use acquire meta
	ameta := ctxValueOr(ctx, connectMetaCtxKey{}, ConnAcquireMeta{})
	// TODO: support connect-specific meta
	// meta := ctxValueOr(ctx, connectMetaCtxKey{}, ConnectMeta{})
	errStr := ""
	if data.Err != nil {
		if errors.Is(data.Err, context.DeadlineExceeded) {
			errStr = kContextDeadlineExceededLabel
		} else {
			errStr = fmt.Sprintf("<%T>", data.Err)
		}
	}
	labels := prometheus.Labels{
		"conn_mode": xoptional.ValueCastOr(ameta.ConnMode, pgconn.Mode.String, prom.LabelUnknown),
		"error":     errStr,
	}
	if t.connectFinishCounter != nil {
		t.connectFinishCounter.With(labels).Inc()
	}
	if t.connectDurationHistogram != nil {
		t.connectDurationHistogram.With(labels).Observe(elapsed)
	}
}

const (
	kContextDeadlineExceededLabel = "timeout"
)
