package conns

import (
	"context"
	reflect "reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/Deimvis/go-ext/go1.25/ext"
	"github.com/Deimvis/go-ext/go1.25/xcheck/xmust"
	"github.com/Deimvis/go-ext/go1.25/xslices"
	"github.com/Deimvis-go/xpg/pg/internal/types"
)

//go:generate mockgen -source=../types/conn.go -destination=lazy_acquire_mocks_test.go -package=conns

func TestLazifyAcquire(t *testing.T) {
	acquired := false
	t.Run("as-Conn", func(t *testing.T) {
		t.Run("proxies", func(t *testing.T) {
			newConnMock := func(ctrl *gomock.Controller) types.Conn {
				return NewMockConn(ctrl)
			}
			testProxiesTo(t, newConnMock, newLazyConn[types.Conn](&acquired),
				ProxiesWithExpectations(expectAcquired[types.Conn](&acquired)))
		})
	})
	t.Run("as-StandaloneConn", func(t *testing.T) {
		t.Run("proxies", func(t *testing.T) {
			newConnMock := func(ctrl *gomock.Controller) types.StandaloneConn {
				return NewMockStandaloneConn(ctrl)
			}
			testProxiesTo(t, newConnMock, newLazyConn[types.StandaloneConn](&acquired),
				ProxiesWithExpectations(expectAcquired[types.StandaloneConn](&acquired)),
				ProxiesMethods[types.StandaloneConn](func(m reflect.Method) bool {
					return m.Name != "Close"
				}))
		})
		t.Run("proxies-after-acquire", func(t *testing.T) {
			newConnMock := func(ctrl *gomock.Controller) types.StandaloneConn {
				return NewMockStandaloneConn(ctrl)
			}
			testProxiesTo(t, newConnMock, newLazyConnAcquired[types.StandaloneConn](&acquired),
				ProxiesWithExpectations(expectAcquired[types.StandaloneConn](&acquired)),
				ProxiesMethods[types.StandaloneConn](func(m reflect.Method) bool {
					return m.Name == "Close"
				}))
		})
		t.Run("Close/emulated-error", func(t *testing.T) {
			aFn := func(context.Context, types.ConnMode, ...types.ConnAcquireOption) (types.Conn, error) {
				return nil, nil
			}
			conn, err := LazifyAcquire(aFn)(context.Background(), types.ConnMode_RO)
			require.NoError(t, err)
			sconn := conn.(types.StandaloneConn)
			err = sconn.Close(context.Background())
			require.NoError(t, err)
			_, err = sconn.Exec(context.Background(), `SELECT 1`)
			require.Error(t, err)
			require.Implements(t, (*types.LazyConnAcquireError)(nil), err)
			require.Equal(t, "conn closed", err.Error())
			lae := err.(lazyAcquireError)
			require.True(t, lae.isEmulated)
		})
	})
	t.Run("as-PoolConn", func(t *testing.T) {
		t.Run("proxies", func(t *testing.T) {
			newConnMock := func(ctrl *gomock.Controller) types.PoolConn {
				return NewMockPoolConn(ctrl)
			}
			testProxiesTo(t, newConnMock, newLazyConn[types.PoolConn](&acquired),
				ProxiesWithExpectations(expectAcquired[types.PoolConn](&acquired)),
				ProxiesMethods[types.PoolConn](func(m reflect.Method) bool {
					return m.Name != "Release"
				}))
		})
		t.Run("proxies-after-acquire", func(t *testing.T) {
			newConnMock := func(ctrl *gomock.Controller) types.PoolConn {
				return NewMockPoolConn(ctrl)
			}
			testProxiesTo(t, newConnMock, newLazyConnAcquired[types.PoolConn](&acquired),
				ProxiesWithExpectations(expectAcquired[types.PoolConn](&acquired)),
				ProxiesMethods[types.PoolConn](func(m reflect.Method) bool {
					return m.Name == "Release"
				}))
		})
		t.Run("Release/emulated-error", func(t *testing.T) {
			aFn := func(context.Context, types.ConnMode, ...types.ConnAcquireOption) (types.Conn, error) {
				return nil, nil
			}
			conn, err := LazifyAcquire(aFn)(context.Background(), types.ConnMode_RO)
			require.NoError(t, err)
			pconn := conn.(types.PoolConn)
			pconn.Release()
			_, err = pconn.Exec(context.Background(), `SELECT 1`)
			require.Error(t, err)
			require.Implements(t, (*types.LazyConnAcquireError)(nil), err)
			require.Equal(t, "conn released", err.Error())
			lae := err.(lazyAcquireError)
			require.True(t, lae.isEmulated)
		})
	})
	t.Run("as-LazyConn", func(t *testing.T) {
		t.Run("methods-work", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mock := NewMockLazyConn(ctrl)

			ctx := context.Background()
			mode := types.ConnMode_RO
			opts := []types.ConnAcquireOption{}
			aFn := func(actCtx context.Context, actMode types.ConnMode, actOpts ...types.ConnAcquireOption) (types.Conn, error) {
				require.Equal(t, ctx, actCtx)
				require.Equal(t, mode, actMode)
				require.Equal(t, opts, actOpts)
				return mock, nil
			}
			lazyConn, err := LazifyAcquire(aFn)(ctx, mode, opts...)
			require.NoError(t, err)
			lc := lazyConn.(types.LazyConn)

			require.False(t, lc.Acquired())
			require.NoError(t, lc.Acquire())
			require.True(t, lc.Acquired())
			actCtx, actMode, actOpts := lc.AcquireArgs()
			require.Equal(t, ctx, actCtx)
			require.Equal(t, mode, actMode)
			require.Equal(t, opts, actOpts)
		})
		t.Run("Acquire/twice-ok", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mock := NewMockLazyConn(ctrl)
			lc := newLazyConn[types.LazyConn](&acquired)(mock)
			require.False(t, lc.Acquired())
			require.NoError(t, lc.Acquire())
			require.True(t, lc.Acquired())
			require.NoError(t, lc.Acquire())
			require.True(t, lc.Acquired())
		})
		t.Run("Acquire/error-propagated", func(t *testing.T) {
			aCtx := context.Background()
			aMode := types.ConnMode_RO
			aOpts := []types.ConnAcquireOption{}
			aErr := customErr{msg: "acquire failed"}
			aFn := func(context.Context, types.ConnMode, ...types.ConnAcquireOption) (types.Conn, error) {
				return nil, aErr
			}
			lazyConn, err := LazifyAcquire(aFn)(aCtx, aMode, aOpts...)
			require.NoError(t, err)
			lc := lazyConn.(types.LazyConn)

			actErr := lc.Acquire()
			require.ErrorIs(t, actErr, aErr)
			actCtx, actMode, actOpts := actErr.AcquireArgs()
			require.Equal(t, aCtx, actCtx)
			require.Equal(t, aMode, actMode)
			require.Equal(t, aOpts, actOpts)
		})
	})
	t.Run("as-ConnReflect", func(t *testing.T) {
		t.Run("proxies", func(t *testing.T) {
			newConnMock := func(ctrl *gomock.Controller) types.ConnReflect {
				return NewMockConnReflect(ctrl)
			}
			testProxiesTo(t, newConnMock, newLazyConn[types.ConnReflect](&acquired),
				ProxiesWithExpectations(expectAcquired[types.ConnReflect](&acquired)),
				ProxiesMethods[types.ConnReflect](func(m reflect.Method) bool {
					return !xslices.Has([]string{"IsLazy", "OwnershipTaken", "TakeOwnership"}, m.Name)
				}))
		})
		t.Run("IsLazy-works-and-doesnt-causes-acquire", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mock := NewMockConnReflect(ctrl)
			cr := newLazyConn[types.ConnReflect](&acquired)(mock)
			require.True(t, cr.IsLazy().HasValue())
			require.True(t, cr.IsLazy().Value())
			require.NoError(t, any(cr).(types.LazyConn).Acquire())
			require.True(t, cr.IsLazy().HasValue())
			require.False(t, cr.IsLazy().Value())
		})
	})
	t.Run("acquire-error-propagated", func(t *testing.T) {
		t.Run("Exec", func(t *testing.T) {
			aCtx := context.Background()
			aMode := types.ConnMode_RO
			aOpts := []types.ConnAcquireOption{}
			aErr := customErr{msg: "acquire failed"}
			aFn := func(context.Context, types.ConnMode, ...types.ConnAcquireOption) (types.Conn, error) {
				return nil, aErr
			}
			lazyConn, err := LazifyAcquire(aFn)(aCtx, aMode, aOpts...)
			require.NoError(t, err)

			_, actErr := lazyConn.Exec(context.Background(), `SELECT 1`)
			require.Implements(t, (*types.LazyConnAcquireError)(nil), actErr)
			require.ErrorIs(t, actErr, aErr)
			actCtx, actMode, actOpts := actErr.(types.LazyConnAcquireError).AcquireArgs()
			require.Equal(t, aCtx, actCtx)
			require.Equal(t, aMode, actMode)
			require.Equal(t, aOpts, actOpts)
		})
	})
}

func newLazyConn[ConnT types.Conn](acquired *bool) func(ConnT) ConnT {
	xmust.NotNilPtr(acquired)
	return func(conn ConnT) ConnT {
		aFn := func(ctx context.Context, mode types.ConnMode, opts ...types.ConnAcquireOption) (types.Conn, error) {
			*acquired = true
			return conn, nil
		}
		lazyConn, err := LazifyAcquire(aFn)(context.Background(), types.ConnMode_RO) // any args
		xmust.NoErr(err)
		return lazyConn.(ConnT)
	}
}

func newLazyConnAcquired[ConnT types.Conn](acquired *bool) func(ConnT) ConnT {
	new := newLazyConn[ConnT](acquired)
	return func(conn ConnT) ConnT {
		lazyConn := new(conn)
		xmust.NoErr(any(lazyConn).(types.LazyConn).Acquire())
		return lazyConn
	}
}

func expectAcquired[InterfaceT any](acquired *bool) func(ctrl *gomock.Controller, mock InterfaceT, method reflect.Method) func(t *testing.T) {
	xmust.NotNilPtr(acquired)
	return func(ctrl *gomock.Controller, mock InterfaceT, method reflect.Method) func(t *testing.T) {
		*acquired = false
		return func(t *testing.T) {
			require.True(t, *acquired)
		}
	}
}

type customErr struct {
	msg string
}

func (c customErr) Error() string {
	return c.msg
}

// TODO: migrate code below to xtest

type proxiesToOption[InterfacetT any] func(*proxiesToCfg[InterfacetT])

func testProxiesTo[InterfaceT any](
	t *testing.T,
	newGomock func(*gomock.Controller) InterfaceT,
	newProxy func(InterfaceT) InterfaceT,
	opts ...proxiesToOption[InterfaceT],
) {
	t.Helper()

	cfg := proxiesToCfg[InterfaceT]{
		addExpectations: nil,
		methodFilterFn:  func(reflect.Method) bool { return true },
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	intT := reflect.TypeOf((*InterfaceT)(nil)).Elem()
	xmust.Eq(intT.Kind(), reflect.Interface, "bug: failed to obtain reflect.Type of interface type parameter")

	for i := 0; i < intT.NumMethod(); i++ {
		method := intT.Method(i)
		if cfg.methodFilterFn != nil && !cfg.methodFilterFn(method) {
			continue
		}
		args := make([]reflect.Value, method.Type.NumIn())
		for j := 0; j < method.Type.NumIn(); j++ {
			args[j] = reflect.Zero(method.Type.In(j))
		}

		t.Run(method.Name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mock := newGomock(ctrl)
			ctrl.RecordCallWithMethodType(mock, method.Name, method.Type, ext.Map(args, reflect.Value.Interface)...).Times(1)
			if cfg.addExpectations != nil {
				cb := cfg.addExpectations(ctrl, mock, method)
				defer cb(t)
			}

			proxy := newProxy(mock)
			proxyMethod := reflect.ValueOf(proxy).MethodByName(method.Name)
			require.Truef(t, proxyMethod.IsValid(), "no method '%s' in proxy", method.Name)
			_ = proxyMethod.Call(args)
		})
	}
}

func ProxiesWithExpectations[InterfaceT any](fn func(ctrl *gomock.Controller, mock InterfaceT, method reflect.Method) func(t *testing.T)) proxiesToOption[InterfaceT] {
	return func(cfg *proxiesToCfg[InterfaceT]) {
		cfg.addExpectations = fn
	}
}

func ProxiesMethods[InterfaceT any](filterFn func(reflect.Method) bool) proxiesToOption[InterfaceT] {
	return func(cfg *proxiesToCfg[InterfaceT]) {
		cfg.methodFilterFn = filterFn
	}
}

type proxiesToCfg[InterfaceT any] struct {
	addExpectations func(ctrl *gomock.Controller, mock InterfaceT, method reflect.Method) func(t *testing.T)
	methodFilterFn  func(method reflect.Method) bool
}
