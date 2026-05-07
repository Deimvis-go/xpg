package types

import (
	"context"

	"github.com/Deimvis/go-ext/go1.25/xoptional"
)

type ConnProvider interface {
	// Acquire returns conn and possibly its ownership
	// which should be taken.
	Acquire(context.Context, ConnMode, ...ConnAcquireOption) (Conn, xoptional.T[ConnOwnership], error)
	// AcquireManaged is a version of Acquire that
	// guarantees to return connection which ownership
	// is already taken, and so connection
	// should not be managed.
	AcquireManaged(context.Context, ConnMode, ...ConnAcquireOption) (Conn, error)
}

type SilentConnProvider interface {
	Acquire(context.Context, ConnMode, ...ConnAcquireOption) (Conn, xoptional.T[ConnOwnership])
	AcquireManaged(context.Context, ConnMode, ...ConnAcquireOption) Conn
}

type SugaredConnProvider interface {
	// R attempt to acquire RO conn first and RW second.
	R(context.Context) (Conn, xoptional.T[ConnOwnership], error)
	RO(context.Context) (Conn, xoptional.T[ConnOwnership], error)
	RW(context.Context) (Conn, xoptional.T[ConnOwnership], error)

	ManagedR(context.Context) (Conn, error)
	ManagedRO(context.Context) (Conn, error)
	ManagedRW(context.Context) (Conn, error)

	// TODO: add Lazy* for conns
}

type SugaredSilentConnProvider interface {
	// R attempt to acquire RO conn first and RW second.
	R(context.Context) (Conn, xoptional.T[ConnOwnership])
	RO(context.Context) (Conn, xoptional.T[ConnOwnership])
	RW(context.Context) (Conn, xoptional.T[ConnOwnership])

	ManagedR(context.Context) Conn
	ManagedRO(context.Context) Conn
	ManagedRW(context.Context) Conn

	// TODO: add Lazy* for conns
}

type ConnProviderMeta interface {
	// Type returns complete type name of current
	// conn provider instance.
	Type() string
	// GenericType returns generic type name of
	// conn provider.
	// Must be independent of runtime state (stateless).
	GenericType() string
	// AcquireType returns acquire approach name
	// of this conn provider.
	// Must be independent of runtime state (stateless).
	AcquireType() xoptional.T[string]
}

// ConnProviderInternals is used as optimization
// for case when many options are passed.
// ConnProvider implementing ConnProviderInternals
// allows passing preprocessed options
// in a form of config.
// This is useful optimization for aggregating providers,
// like Fallbacked conn provider.
type ConnProviderInternals interface {
	AcquireWithConfig(context.Context, ConnMode, ConnAcquireConfig) (Conn, xoptional.T[ConnOwnership], error)
	AcquireManagedWithConfig(context.Context, ConnMode, ConnAcquireConfig) (Conn, error)
}
