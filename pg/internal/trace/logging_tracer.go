package trace

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Deimvis/go-ext/go1.25/ext"
	"github.com/Deimvis/go-ext/go1.25/xfallback/xfb"
	"github.com/Deimvis/go-ext/go1.25/xoptional"
	"github.com/Deimvis/go-ext/go1.25/xptr"
	"github.com/Deimvis-go/logs/logs"
	"github.com/Deimvis-go/xpg/pg/pgconn"
	"github.com/Deimvis-go/xprometheus/xprometheus"
)

const MAX_TRACE_QUERY_LENGTH = 4096
const MAX_TRACE_ARG_LENGTH = 1024

func NewLoggingTracer(lg logs.KVCtxLogger) *LoggingTracer {
	return &LoggingTracer{lg: lg}
}

type LoggingTracer struct {
	lg logs.KVCtxLogger
}

func (t *LoggingTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	query := truncateQuery(data.SQL)
	args := ext.Map(data.Args, truncateArg)
	queryId := uuid.New().String()
	meta := ctxValueOr(ctx, queryMetaCtxKey{}, QueryMeta{})
	t.lg.Debug(ctx, "PG Query - start", "query_id", queryId, "query", query, "args", args,
		"query_name", xoptional.ValueOr(meta.QueryName, xprometheus.LabelUnknown),
		"conn_mode", xoptional.ValueCastOr(meta.ConnMode, pgconn.Mode.String, xprometheus.LabelUnknown))
	ctx = context.WithValue(ctx, queryIdCtxKey{}, queryId)
	ctx = context.WithValue(ctx, queryStartCtxKey{}, time.Now())
	return ctx
}

func (t *LoggingTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	queryId := ctx.Value(queryIdCtxKey{}).(string)
	start := ctx.Value(queryStartCtxKey{}).(time.Time)
	elapsed := time.Since(start).Seconds()
	meta := ctxValueOr(ctx, queryMetaCtxKey{}, QueryMeta{})
	t.lg.Debug(ctx, "PG Query - finish", "query_id", queryId, "elapsed", elapsed, "tag", data.CommandTag, "err", data.Err,
		"query_name", xoptional.ValueOr(meta.QueryName, xprometheus.LabelUnknown),
		"conn_mode", xoptional.ValueCastOr(meta.ConnMode, pgconn.Mode.String, xprometheus.LabelUnknown))
}

func (t *LoggingTracer) TraceAcquireStart(ctx context.Context, p *pgxpool.Pool, _ pgxpool.TraceAcquireStartData) context.Context {
	ctx, acquireId := ctxValueOrSetLazy(ctx, acquireIdCtxKey{}, genAcquireId)
	st := p.Stat()
	meta := ctxValueOr(ctx, connAcquireMetaCtxKey{}, ConnAcquireMeta{})
	t.lg.Debug(ctx, "PG Conn Acquire - start", "acquire_id", acquireId,
		"total_conns", st.TotalConns(), "acquired_conns", st.AcquiredConns(),
		"idle_conns", st.IdleConns(), "constructing_conns", st.ConstructingConns(),
		"conn_mode", xoptional.ValueCastOr(meta.ConnMode, pgconn.Mode.String, xprometheus.LabelUnknown))
	ctx = context.WithValue(ctx, acquireIdCtxKey{}, acquireId)
	ctx = context.WithValue(ctx, acquireStartCtxKey{}, time.Now())
	return ctx
}

func (t *LoggingTracer) TraceAcquireEnd(ctx context.Context, _ *pgxpool.Pool, data pgxpool.TraceAcquireEndData) {
	acquireId := ctx.Value(acquireIdCtxKey{}).(string)
	start := ctx.Value(acquireStartCtxKey{}).(time.Time)
	elapsed := time.Since(start).Seconds()
	meta := ctxValueOr(ctx, connAcquireMetaCtxKey{}, ConnAcquireMeta{})
	t.lg.Debug(ctx, "PG Conn Acquire - finish", "acquire_id", acquireId, "elapsed", elapsed, "err", data.Err,
		"conn_mode", xoptional.ValueCastOr(meta.ConnMode, pgconn.Mode.String, xprometheus.LabelUnknown))
}

func (t *LoggingTracer) TraceRelease(pool *pgxpool.Pool, _ pgxpool.TraceReleaseData) {
	t.lg.KV().Debug("PG Conn Release - start")
}

func (t *LoggingTracer) TraceConnectStart(ctx context.Context, data pgx.TraceConnectStartData) context.Context {
	connectId := uuid.New().String()
	t.lg.Debug(ctx, "PG Connect - start", "connect_id", connectId, "timeout", data.ConnConfig.ConnectTimeout)
	ctx = context.WithValue(ctx, connectIdCtxKey{}, connectId)
	ctx = context.WithValue(ctx, connectStartCtxKey{}, time.Now())
	return ctx
}

func (t *LoggingTracer) TraceConnectEnd(ctx context.Context, data pgx.TraceConnectEndData) {
	connectId := ctx.Value(connectIdCtxKey{}).(string)
	start := ctx.Value(connectStartCtxKey{}).(time.Time)
	elapsed := time.Since(start).Seconds()
	t.lg.Debug(ctx, "PG Connect - finish", "connect_id", connectId, "elapsed", elapsed, "err", data.Err)
}

func truncateQuery(query string) string {
	if len(query) > MAX_TRACE_QUERY_LENGTH {
		return fmt.Sprintf("%s...<long query of length ~%d>", query[:MAX_TRACE_QUERY_LENGTH-100], len(query))
	}
	return query
}

func truncateArg(arg any) any {
	length := calcApproxLength(reflect.ValueOf(arg))
	if length != nil && *length > MAX_TRACE_ARG_LENGTH {
		return fmt.Sprintf("<long value of length ~%d>", *length)
	}
	return arg
}

func calcApproxLength(v reflect.Value) *int {
	var length *int = nil
	switch v.Kind() {
	case reflect.Struct:
		l := 0
		for i := 0; i < v.NumField(); i++ {
			l += xfb.OnNilv(calcApproxLength(v.Field(i)), 0)
		}
		length = &l
	case reflect.Slice:
		l := 0
		for i := 0; i < v.Len(); i++ {
			l += xfb.OnNilv(calcApproxLength(v.Index(i)), 0)
		}
		length = &l
	case reflect.String:
		length = xptr.T(v.Len())
	default:
		length = xptr.T(int(v.Type().Size()))
	}
	return length
}
