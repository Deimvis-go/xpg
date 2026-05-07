package types

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/Deimvis/go-ext/go1.25/xoptional"
)

type ConnMode string

const (
	ConnMode_RO ConnMode = "ro"
	ConnMode_RW ConnMode = "rw"
)

func (cm ConnMode) String() string {
	return string(cm)
}

type Conn interface {
	Exec(ctx context.Context, sql string, args ...any) (commandTag pgconn.CommandTag, err error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
	Begin(ctx context.Context) (pgx.Tx, error)
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

type StandaloneConn interface {
	Conn
	// Close closes connection.
	// Safe to call on already closed connection -
	// nothing will happen.
	Close(ctx context.Context) error
}

type PoolConn interface {
	Conn
	// Release releases connection back to the pool.
	// Safe to call on already released connection -
	// nothing will happen.
	Release()
}

type ConnReflect interface {
	Conn
	MutableConnMeta
	// TODO: add methods to mutate some conn state (MutableConn interface)
	// TODO: add method TakeOwnership() error
}

// LazyConn is a conn with lazy acquisition.
// Any error returned by LazyConn
// may be LazyConnAcquireError, which means
// it was caused by acquire operation.
type LazyConn interface {
	Conn
	// Acquire initializes lazy conn.
	// Safe to call on already acquired connection -
	// nothing will happen.
	Acquire() LazyConnAcquireError
	// Acquired returns whether acquire operation
	// was called.
	// If acquire operation failed,
	// Acquired still returns true.
	// If connection became invalid
	// (e.g. because it was closed or released),
	// Acquired still returns true.
	Acquired() bool
	AcquireArgs() (context.Context, ConnMode, []ConnAcquireOption)
	// TODO: add method to get internal conn
}

// ConnMeta describes connection.
//
// Note that IsOneTime and IsLazy return values
// are independent and one value does not
// imply value for other one.
// If one needs to check whether connection
// is currently "acquired",
// he should check that both
// IsOneTime and IsLazy are false
type ConnMeta interface {
	Mode() xoptional.T[ConnMode]
	// OwnershipTaken returns whether connect
	// ownership is already taken and
	// it does not need to be "freed".
	// In other words, if OwnershipTaken returns true,
	// it means one already taken connection
	// ownership and is obligated to "free"
	// the connection.
	OwnershipTaken() xoptional.T[bool]
	// IsOneTime returns whether connect
	// is one-time, which means that
	// connect is actually
	// "acquired" when command starts
	// and "freed" when command finishes,
	// so each command actually uses different connections.
	// If information is unkonwn returned value has no value.
	IsOneTime() xoptional.T[bool]
	// IsLazy returns whether connect
	// is lazy, which means that
	// connect is actually
	// "acquired" when it is first time used.
	// When lazy conn "acquires"
	// an actual conn, it ends being lazy
	// and IsLazy would return false.
	IsLazy() xoptional.T[bool]
	// TODO: smth like Host()
}

type MutableConnMeta interface {
	ConnMeta
	TakeOwnership() (OwnedConn, error)
}

type OwnedConn interface {
	FreeConn(context.Context) error
}

type ConnFreeFn func(context.Context) error

// LazyConnAcquireError is error caused by
// delayed acquire operation.
type LazyConnAcquireError interface {
	error
	AcquireArgs() (context.Context, ConnMode, []ConnAcquireOption)
	// TODO: not sure to expose this
	// // emulated means it was originated by lazy conn
	// // rather than by acquire function call itself
	// IsEmulated() bool
}
