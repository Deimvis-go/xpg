package pgconn

import (
	"github.com/Deimvis-go/xpg/pg/internal/conns"
	"github.com/Deimvis-go/xpg/pg/internal/types"
)

type AcquireFn = types.ConnAcquireFn
type AcquireManagedFn = types.ConnAcquireManagedFn
type AcquireOption = types.ConnAcquireOption

var (
	AcquireWithMeta           = conns.AcquireWithMeta
	AcquireWithRWFallback     = conns.AcquireWithRWFallback
	AcquireWithTimeout        = conns.AcquireWithTimeout
	AcquireWithAttemptTimeout = conns.AcquireWithAttemptTimeout
)

var LazifyAcquire = conns.LazifyAcquire
