package pgconnprovider

import (
	"fmt"

	"github.com/Deimvis/go-ext/go1.25/ext"
	"github.com/Deimvis-go/xpg/pg/internal/types"
)

type FallbackedHooks struct {
	OnAttemptStart      func(EventContext, FallbackedAttemptState) error
	OnAttemptFinishOk   func(EventContext, FallbackedAttemptState) error
	OnAttemptFinishFail func(EventContext, FallbackedAttemptState, FallbackedFail) error
}

type FallbackedAttemptState struct {
	Index    int
	Provider types.ConnProvider
}

func (fas FallbackedAttemptState) AcquireTypeOr(fb string) string {
	if m, ok := fas.Provider.(types.ConnProviderMeta); ok {
		if acquireType := m.AcquireType(); acquireType.HasValue() {
			return acquireType.Value()
		}
	}
	return fb
}

type FallbackedFail struct {
	// Implementation note:
	// struct type is used in order to prevent
	// users from passing it
	// to fail hook return value.
	err error
}

func (ff FallbackedFail) Error() error {
	return ff.err
}

func newWithStartOkFailHooks(
	start func() error,
	ok func() error,
	fail func(error) error,
) func(payloadFn func() error) error {
	return func(payloadFn func() error) error {
		failCb := func(err error) {
			if fail != nil {
				fail(err)
			}
		}
		if fail != nil {
			defer ext.OnPanic(func(r any) {
				if err, ok := r.(error); ok {
					failCb(err)
				} else {
					err := fmt.Errorf("panicked on %v", r)
					failCb(err)
				}
				panic(r)
			})
		}

		if start != nil {
			err := start()
			if err != nil {
				failCb(err)
				return err
			}
		}

		err := payloadFn()
		if err != nil {
			failCb(err)
			return err
		}

		if ok != nil {
			err := ok()
			if err != nil {
				failCb(err)
				return err
			}
		}
		return nil
	}
}
