package pgtrace

import (
	"github.com/Deimvis-go/xpg/pg/internal/trace"
)

type ConnAcquireMeta = trace.ConnAcquireMeta

var CtxWithConnAcquireMeta = trace.CtxWithConnAcquireMeta
var CtxConnAcquireMeta = trace.CtxConnAcquireMeta

type QueryMeta = trace.QueryMeta

var CtxWithQueryMeta = trace.CtxWithQueryMeta
var CtxQueryMeta = trace.CtxQueryMeta
