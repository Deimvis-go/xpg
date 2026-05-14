package trace

import (
	"context"
	"encoding/json"
	"expvar"
	"runtime"
	"sync"
	"time"
	"unsafe"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Deimvis/go-ext/go1.25/xoptional"
	"github.com/Deimvis/go-ext/go1.25/xptr"
	"github.com/Deimvis-go/xpg/pg/pgconn"
	"github.com/Deimvis-go/xprometheus/prom"
)

func NewDebugVarsTracer(varsRootName string) *DebugVarsTracer {
	varsRoot := expvar.NewMap(varsRootName)
	acquired := &objVar{m: make(map[string]any)}
	varsRoot.Set("acquired_conns", acquired)
	tracer := &DebugVarsTracer{
		varsRoot: varsRoot,
		acquired: acquired,
	}
	tracer.acquiresPool.New = func() any {
		return &acquireDebugInfo{
			startStackBuf: make([]byte, stackBufInitSize),
		}
	}
	return tracer
}

type DebugVarsTracer struct {
	varsRoot          *expvar.Map
	acquired          *objVar
	connPtr2acquireId sync.Map // unsafe.Pointer->string

	acquiresPool sync.Pool
}

func (t *DebugVarsTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	return ctx
}

func (t *DebugVarsTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
}

func (t *DebugVarsTracer) TraceAcquireStart(ctx context.Context, p *pgxpool.Pool, _ pgxpool.TraceAcquireStartData) context.Context {
	ctx, acquireId := ctxValueOrSetLazy(ctx, acquireIdCtxKey{}, genAcquireId)
	ctx, startTime := ctxValueOrSetLazy(ctx, acquireStartCtxKey{}, time.Now)
	meta := ctxValueOr(ctx, connAcquireMetaCtxKey{}, ConnAcquireMeta{})
	debugInfo := t.acquiresPool.Get().(*acquireDebugInfo)
	debugInfo.reset()
	debugInfo.acquireId = acquireId
	debugInfo.SetStartStack()
	debugInfo.startTs = startTime.Unix()
	debugInfo.connMode = xoptional.ValueCastOr(meta.ConnMode, pgconn.Mode.String, prom.LabelUnknown)
	t.acquired.Write(func(m map[string]any) {
		m[acquireId] = debugInfo
	})
	return ctx
}

func (t *DebugVarsTracer) TraceAcquireEnd(ctx context.Context, _ *pgxpool.Pool, data pgxpool.TraceAcquireEndData) {
	acquireId := ctx.Value(acquireIdCtxKey{}).(string)
	if data.Err != nil {
		t.acquired.Write(func(m map[string]any) {
			debugInfo := m[acquireId].(*acquireDebugInfo)
			delete(m, acquireId)
			debugInfo.reset()
			t.acquiresPool.Put(debugInfo)
		})
		return
	}
	t.acquired.Write(func(m map[string]any) {
		m[acquireId].(*acquireDebugInfo).finishTs = xptr.T(time.Now().Unix())
	})
	t.connPtr2acquireId.Store(unsafe.Pointer(data.Conn), acquireId)
	// TODO: add ttl to record, so we avoid leaking memory to this record
	// even if some concurrent release happens
}

func (t *DebugVarsTracer) TraceRelease(pool *pgxpool.Pool, data pgxpool.TraceReleaseData) {
	acquireId_, ok := t.connPtr2acquireId.Load(unsafe.Pointer(data.Conn))
	if !ok {
		return
	}
	defer t.connPtr2acquireId.Delete(unsafe.Pointer(data.Conn))
	acquireId := acquireId_.(string)
	t.acquired.Write(func(m map[string]any) {
		debugInfo := m[acquireId].(*acquireDebugInfo)
		delete(m, acquireId)
		debugInfo.reset()
		t.acquiresPool.Put(debugInfo)
	})
}

func (t *DebugVarsTracer) TraceConnectStart(ctx context.Context, data pgx.TraceConnectStartData) context.Context {
	return ctx
}

func (t *DebugVarsTracer) TraceConnectEnd(ctx context.Context, data pgx.TraceConnectEndData) {
}

type acquireDebugInfo struct {
	acquireId      string
	connMode       string
	startTs        int64
	finishTs       *int64
	startStackBuf  []byte
	startStackSize int
}

func (acq *acquireDebugInfo) SetStartStack() {
	for {
		n := runtime.Stack(acq.startStackBuf, false)
		if n < len(acq.startStackBuf) {
			acq.startStackSize = n
			return
		}
		acq.startStackBuf = make([]byte, 2*len(acq.startStackBuf))
	}
}

func (acq *acquireDebugInfo) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		AcquireId       string `json:"acquire_id"`
		ConnMode        string `json:"conn_mode"`
		StartTs         int64  `json:"start_ts"`
		FinishTs        *int64 `json:"finish_ts"`
		StartStackTrace string `json:"start_stack_trace"`
	}{
		AcquireId:       acq.acquireId,
		ConnMode:        acq.connMode,
		StartTs:         acq.startTs,
		FinishTs:        acq.finishTs,
		StartStackTrace: string(acq.startStackBuf[:acq.startStackSize]),
	})
}

func (acq *acquireDebugInfo) reset() {
	acq.acquireId = ""
	acq.connMode = ""
	acq.startTs = 0
	acq.finishTs = nil
}

const stackBufInitSize = 1024
