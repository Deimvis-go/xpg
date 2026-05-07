package types

import (
	"context"
	"time"

	"github.com/Deimvis/go-ext/go1.25/xoptional"
)

type ConnAcquireOption func(*ConnAcquireConfig)
type ConnAcquireFn func(context.Context, ConnMode, ...ConnAcquireOption) (Conn, xoptional.T[ConnOwnership], error)
type ConnAcquireManagedFn func(context.Context, ConnMode, ...ConnAcquireOption) (Conn, error)
type SilentConnAcquireFn func(context.Context, ConnMode, ...ConnAcquireOption) Conn

type ConnAcquireConfig struct {
	RWFallback     bool
	Timeout        xoptional.T[time.Duration]
	AttemptTimeout xoptional.T[time.Duration]
	Meta           xoptional.T[ConnAcquireTracingMeta]
}

type ConnOwnership interface {
	Take() (OwnedConn, error)
	MustTake() OwnedConn
	// TODO: add Steal() that will allow attempting stealing connection ownership
	// need to add semantics on "first Take call" side whether to allow stealing or not.
	// (like TakeWeakly())
	// TODO: add method to take weakly and allow someone else to take it
	// UPD: no, it's bad, conn is single-threaded
	// (TakeShared())
}

type ConnAcquireTracingMeta struct {
	ConnMode xoptional.T[ConnMode]
}
