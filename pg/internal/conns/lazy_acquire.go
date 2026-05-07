package conns

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/Deimvis/go-ext/go1.25/xoptional"
	"github.com/Deimvis/go-ext/go1.25/xslices"
	"github.com/Deimvis-go/xpg/pg/internal/types"
)

// LazifyAcquire returns makes "acquiring"
// a lazy operation, delayed until connection
// is used for the first time.
//
// Note that context passed to acquire call
// will be used in delayed acquire
// and its deadline will be applied
// on delayed operation.
// In case timeout should be set on acquire call
// then AcquireWithTimeout option should be used.
//
// If acquire fails during first use of connection
// then error is returned and subsequent connection
// uses will return the same error caused by acquire operation.
// In case retries should be used on acquire call
// then it should be configured for acquire operation.
func LazifyAcquire(fn types.ConnAcquireManagedFn) types.ConnAcquireManagedFn {
	return func(ctx context.Context, mode types.ConnMode, opts ...types.ConnAcquireOption) (types.Conn, error) {
		return &lazyConn{aFn: fn, aCtx: ctx, aMode: mode, aOpts: opts}, nil
	}
}

// lazyConn represents lazy-initialized conn.
// lazyConn uses context supplied on acquiring,
// rather than context given on first connection use.
// Implementation note:
// conn is assumed to be single-threaded,
// so lazy initialization does not need
// any synchronization
// (conn is assumed to be associated with a single session,
// which is associated with at most one transaction at a time).
type lazyConn struct {
	aFn   types.ConnAcquireManagedFn
	aCtx  context.Context
	aMode types.ConnMode
	aOpts []types.ConnAcquireOption
	aErr  xoptional.T[types.LazyConnAcquireError]

	conn xoptional.T[types.Conn]
}

// interface guards
var _ types.Conn = (*lazyConn)(nil)
var _ types.StandaloneConn = (*lazyConn)(nil)
var _ types.PoolConn = (*lazyConn)(nil)
var _ types.LazyConn = (*lazyConn)(nil)
var _ types.ConnMeta = (*lazyConn)(nil)
var _ types.ConnReflect = (*lazyConn)(nil)

func (lc *lazyConn) Exec(ctx context.Context, sql string, args ...any) (commandTag pgconn.CommandTag, err error) {
	conn, err := lc.connOrAcquire()
	if err != nil {
		return pgconn.CommandTag{}, err
	}
	return conn.Exec(ctx, sql, args...)
}

func (lc *lazyConn) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	conn, err := lc.connOrAcquire()
	if err != nil {
		return nil, err
	}
	return conn.Query(ctx, sql, args...)
}

func (lc *lazyConn) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	conn, err := lc.connOrAcquire()
	if err != nil {
		return errRow{err: err}
	}
	return conn.QueryRow(ctx, sql, args...)
}

func (lc *lazyConn) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	conn, err := lc.connOrAcquire()
	if err != nil {
		return 0, err
	}
	return conn.CopyFrom(ctx, tableName, columnNames, rowSrc)
}

func (lc *lazyConn) Begin(ctx context.Context) (pgx.Tx, error) {
	conn, err := lc.connOrAcquire()
	if err != nil {
		return nil, err
	}
	return conn.Begin(ctx)
}

func (lc *lazyConn) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	conn, err := lc.connOrAcquire()
	if err != nil {
		return errBatchResults{err}
	}
	return conn.SendBatch(ctx, b)
}

func (lc *lazyConn) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	conn, err := lc.connOrAcquire()
	if err != nil {
		return nil, err
	}
	return conn.BeginTx(ctx, txOptions)
}

func (lc *lazyConn) Close(ctx context.Context) error {
	if !lc.Acquired() {
		lc.aErr.SetValue(lc.newLazyAcquireError_emulated(errors.New("conn closed")))
		return nil
	}
	conn, err := lc.connOrAcquire()
	if err != nil {
		return err
	}
	if sc, ok := conn.(types.StandaloneConn); ok {
		return sc.Close(ctx)
	}
	return nil
}

func (lc *lazyConn) Release() {
	if !lc.Acquired() {
		lc.aErr.SetValue(lc.newLazyAcquireError_emulated(errors.New("conn released")))
		return
	}
	conn, err := lc.connOrAcquire()
	if err != nil {
		return
	}
	if pc, ok := conn.(types.PoolConn); ok {
		pc.Release()
		return
	}
}

func (lc *lazyConn) Mode() xoptional.T[types.ConnMode] {
	conn, err := lc.connOrAcquire()
	if err != nil {
		return xoptional.New[types.ConnMode]()
	}
	if cm, ok := conn.(types.ConnMeta); ok {
		return cm.Mode()
	}
	return xoptional.New[types.ConnMode]()
}

func (lc *lazyConn) OwnershipTaken() xoptional.T[bool] {
	return xoptional.New(true)
}

func (lc *lazyConn) IsOneTime() xoptional.T[bool] {
	conn, err := lc.connOrAcquire()
	if err != nil {
		return xoptional.New[bool]()
	}
	if cm, ok := conn.(types.ConnMeta); ok {
		return cm.IsOneTime()
	}
	return xoptional.New[bool]()
}

func (lc *lazyConn) IsLazy() xoptional.T[bool] {
	return xoptional.New(!lc.conn.HasValue())
}

func (lc *lazyConn) TakeOwnership() (types.OwnedConn, error) {
	return nil, errors.New("ownership is already taken")
}

func (lc *lazyConn) Acquire() types.LazyConnAcquireError {
	_, err := lc.connOrAcquire()
	return err
}

func (lc *lazyConn) Acquired() bool {
	return lc.conn.HasValue()
}

func (lc *lazyConn) AcquireArgs() (context.Context, types.ConnMode, []types.ConnAcquireOption) {
	return lc.aCtx, lc.aMode, xslices.Copy(lc.aOpts)
}

func (lc *lazyConn) connOrAcquire() (types.Conn, types.LazyConnAcquireError) {
	if lc.aErr.HasValue() {
		err := lc.aErr.Value()
		if err != nil {
			return nil, err
		}
		return lc.conn.Value(), nil
	}
	conn, err := lc.aFn(lc.aCtx, lc.aMode, lc.aOpts...)
	var laErr types.LazyConnAcquireError = nil
	if err != nil {
		laErr = lc.newLazyAcquireError_natural(err)
	}
	lc.conn.SetValue(conn)
	lc.aErr.SetValue(laErr)
	return conn, laErr
}

func (lc *lazyConn) newLazyAcquireError_natural(err error) types.LazyConnAcquireError {
	return lazyAcquireError{error: err, isEmulated: false, aCtx: lc.aCtx, aMode: lc.aMode, aOpts: lc.aOpts}
}

func (lc *lazyConn) newLazyAcquireError_emulated(err error) types.LazyConnAcquireError {
	return lazyAcquireError{error: err, isEmulated: true, aCtx: lc.aCtx, aMode: lc.aMode, aOpts: lc.aOpts}
}

type lazyAcquireError struct {
	error
	isEmulated bool
	aCtx       context.Context
	aMode      types.ConnMode
	aOpts      []types.ConnAcquireOption
}

// interface guards
var _ types.LazyConnAcquireError = lazyAcquireError{}

func (lae lazyAcquireError) AcquireArgs() (context.Context, types.ConnMode, []types.ConnAcquireOption) {
	return lae.aCtx, lae.aMode, lae.aOpts
}

func (lae lazyAcquireError) IsEmulated() bool {
	return lae.isEmulated
}

func (lae lazyAcquireError) Unwrap() error {
	return lae.error
}

type errRow struct {
	err error
}

// interface guards
var _ pgx.Row = errRow{}

func (er errRow) Scan(dest ...any) error {
	return er.err
}

type errBatchResults struct {
	err error
}

// interface guards
var _ pgx.BatchResults = errBatchResults{}

func (ebr errBatchResults) Exec() (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, ebr.err
}

func (ebr errBatchResults) Query() (pgx.Rows, error) {
	return nil, ebr.err
}

func (ebr errBatchResults) QueryRow() pgx.Row {
	return errRow{err: ebr.err}
}

func (ebr errBatchResults) Close() error {
	return ebr.err
}
