package conns

import (
	"time"

	"github.com/Deimvis-go/xpg/pg/internal/types"
)

// TODO: AcquireWithDynTimeout(func(ctx, mode) xoptional.T[time.Duration])
// TODO: AcquireWithDynAttemptTimeout(func(ctx, mode) xoptional.T[time.Duration])

func AcquireWithMeta(m types.ConnAcquireTracingMeta) types.ConnAcquireOption {
	return func(cfg *types.ConnAcquireConfig) {
		cfg.Meta.SetValue(m)
	}
}

// AcquireWithRWFallback tells Acquire operation
// to attempt acquire conn with given mode first,
// and if failed then attempt to acquire conn
// with RW mode.
// It is useful because sometimes acquire
// includes multiple stages,
// where each stage independently attempts
// to acquire connection, and option
// allows to make RW mode attempt on each stage
// right after first attempt failed.
func AcquireWithRWFallback() types.ConnAcquireOption {
	return func(c *types.ConnAcquireConfig) {
		c.RWFallback = true
	}
}

// AcquireWithTimeout allows to set timeout
// on full Acquire call.
// It works similar to setting timeout in
// context passed to Acquire call,
// but in case actual Acquire call is delayed
// (when lazy approach is used),
// it won't work as expected and context's
// timeout will include all the time before
// actual Acquire call happens.
// AcquireWithTimeout sets timeout when
// actual Acquire call happens.
func AcquireWithTimeout(t time.Duration) types.ConnAcquireOption {
	return func(c *types.ConnAcquireConfig) {
		c.Timeout.SetValue(t)
	}
}

// AcquireWithAttemptTimeout allows to set timeout
// on each acquire attempt within Acquire call.
// Acquire call may include multiple attempts
// (e.g. when using AcquireWithRWFallback).
func AcquireWithAttemptTimeout(t time.Duration) types.ConnAcquireOption {
	return func(c *types.ConnAcquireConfig) {
		c.AttemptTimeout.SetValue(t)
	}
}

// TODO: implement AcquireWithRetries

// TODO: implement <some_name_here>
// option will ask to acquire non one-time conn,
// which is useful when user is going to make many requests
// and wants to make sure the same connect will be used
// (mostly for performance sake)
// P.S. it's mostly because connect implementation
// may be one-time - it means it is acquired
// only when connect is used (e.g. pool manager as conn provider)
// TODO: maybe generalize and allow to ask to avoid acquiring lazy conns
// and acquire only conns that "already stored locally and does not require
// any network requests".
// func AcquirePersistent() ConnAcquireOption {
// }

func NewAcquireConfig(opts ...types.ConnAcquireOption) types.ConnAcquireConfig {
	cfg := types.ConnAcquireConfig{
		RWFallback: false,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}
