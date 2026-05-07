package pgconnprovider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Deimvis/go-ext/go1.25/ext"
	"github.com/Deimvis/go-ext/go1.25/xoptional"
	"github.com/Deimvis-go/xpg/pg/internal/conns"
	"github.com/Deimvis-go/xpg/pg/internal/types"
	"github.com/Deimvis-go/xpg/pg/pgtrace"
)

func NewFallbacked(provs ...types.ConnProvider) fallbacked {
	return fallbacked{provs: provs, stats: &fallbackedStats{}}
}

type fallbacked struct {
	provs []types.ConnProvider
	hooks FallbackedHooks
	stats *fallbackedStats
}

var _ types.ConnProvider = fallbacked{}
var _ types.ConnProviderMeta = fallbacked{}

func (f fallbacked) Acquire(ctx context.Context, mode types.ConnMode, opts ...types.ConnAcquireOption) (types.Conn, xoptional.T[types.ConnOwnership], error) {
	cfg := conns.NewAcquireConfig(opts...)
	if cfg.Meta.HasValue() {
		ctx = pgtrace.CtxWithConnAcquireMeta(ctx, cfg.Meta.Value())
	}
	if cfg.Timeout.HasValue() {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, cfg.Timeout.Value())
		defer cancel()
	}
	// TODO: maybe modify config so underlying conn providers
	// do not apply options we already applied (e.g. total timeout).
	// Note: actually if new deadline is further than current,
	// context.WithTimeout/WithDeadline does nothinges.

	var errs []error
	for i, prov := range f.provs {
		attempt := FallbackedAttemptState{
			Index:    i,
			Provider: prov,
		}
		var conn types.Conn
		var own xoptional.T[types.ConnOwnership]
		err := f.withAttemptHooks(ctx, attempt, func() error {
			var err error
			f.stats.recordAttempt(attempt, func() {
				if provInt, ok := prov.(types.ConnProviderInternals); ok {
					conn, own, err = provInt.AcquireWithConfig(ctx, mode, cfg)
				} else {
					conn, own, err = prov.Acquire(ctx, mode, opts...)
				}
			})
			return err
		})

		if err == nil {
			return conn, own, err
		}
		if errors.Is(err, context.DeadlineExceeded) {
			break
		}
		errs = append(errs, err)
	}
	return nil, xoptional.New[types.ConnOwnership](), errors.Join(errs...)
}

func (f fallbacked) AcquireManaged(ctx context.Context, mode types.ConnMode, opts ...types.ConnAcquireOption) (types.Conn, error) {
	cfg := conns.NewAcquireConfig(opts...)
	if cfg.Timeout.HasValue() {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, cfg.Timeout.Value())
		defer cancel()
	}

	var errs []error
	for i, prov := range f.provs {
		attempt := FallbackedAttemptState{
			Index:    i,
			Provider: prov,
		}
		var conn types.Conn
		err := f.withAttemptHooks(ctx, attempt, func() error {
			var err error
			f.stats.recordAttempt(attempt, func() {
				if provInt, ok := prov.(types.ConnProviderInternals); ok {
					conn, err = provInt.AcquireManagedWithConfig(ctx, mode, cfg)
				} else {
					conn, err = prov.AcquireManaged(ctx, mode, opts...)
				}
			})
			return err
		})

		if err == nil {
			return conn, err
		}
		if errors.Is(err, context.DeadlineExceeded) {
			break
		}
		errs = append(errs, err)
	}
	return nil, errors.Join(errs...)
}

// WithHooks returns a shallow copy of
// fallbacked conn provider with
// given hooks set.
func (f fallbacked) WithHooks(hooks FallbackedHooks) fallbacked {
	fNew := f.ShallowClone()
	fNew.hooks = hooks
	return fNew
}

// ShallowClone returns a shallow copy of
// fallbacked conn provider
// (underlying conn providers are not cloned,
// stats are cloned with current state and all exports).
func (f fallbacked) ShallowClone() fallbacked {
	return fallbacked{provs: f.provs, stats: f.stats.clone()}
}

func (f fallbacked) Type() string {
	typeParams := ext.Map(f.provs, func(p types.ConnProvider) string {
		if pm, ok := p.(types.ConnProviderMeta); ok {
			return pm.Type()
		}
		return fmt.Sprintf("<%T>", p)
	})
	return fmt.Sprintf("%s[%s]", fallbackedTypeBase, strings.Join(typeParams, ","))
}

func (f fallbacked) GenericType() string {
	return fmt.Sprintf("%s[...ConnProvider]", fallbackedTypeBase)
}

func (f fallbacked) AcquireType() xoptional.T[string] {
	typeParams := ext.Map(f.provs, func(p types.ConnProvider) string {
		if pm, ok := p.(types.ConnProviderMeta); ok {
			if aType := pm.AcquireType(); aType.HasValue() {
				return aType.Value()
			}
		}
		return fmt.Sprintf("<%T>", p)
	})
	v := fmt.Sprintf("%s[%s]", fallbackedAcquireTypeBase, strings.Join(typeParams, ","))
	return xoptional.New(v)
}

func (f fallbacked) withAttemptHooks(ctx context.Context, attempt FallbackedAttemptState, payloadFn func() error) error {
	ectx := NewEventContext(ctx)
	return newWithStartOkFailHooks(
		ext.IfElse(f.hooks.OnAttemptStart != nil,
			func() error {
				return f.hooks.OnAttemptStart(ectx, attempt)
			},
			nil,
		),
		ext.IfElse(f.hooks.OnAttemptFinishOk != nil,
			func() error {
				return f.hooks.OnAttemptFinishOk(ectx, attempt)
			},
			nil,
		),
		ext.IfElse(f.hooks.OnAttemptFinishFail != nil,
			func(e error) error {
				return f.hooks.OnAttemptFinishFail(ectx, attempt, FallbackedFail{err: e})
			},
			nil,
		),
	)(payloadFn)
}

const (
	fallbackedTypeBase        = "fallbacked"
	fallbackedAcquireTypeBase = "fallbacked"
)
