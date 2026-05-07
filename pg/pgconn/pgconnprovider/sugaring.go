package pgconnprovider

import (
	"context"
	"fmt"

	"github.com/Deimvis/go-ext/go1.25/xoptional"
	"github.com/Deimvis-go/xpg/pg/internal/conns"
	"github.com/Deimvis-go/xpg/pg/internal/types"
)

func Sugar(p types.ConnProvider) types.SugaredConnProvider {
	return sugared{p: p}
}

func SugarSilent(p types.SilentConnProvider) types.SugaredSilentConnProvider {
	return sugaredSilent{p: p}
}

// TODO: unsure that it's good to fallback to wrapping
// sugared provider in another struct
// in order to implement desugared one,
// at least I would like to hint somehow that
// if implementation of SugaredConnProvider is not known
// then it will be wrapped
// in order to implement ConnProvider.
//
// func Desugar(sp types.SugaredConnProvider) types.ConnProvider {
// 	if impl, ok := sp.(sugared); ok {
// 		return impl.p
// 	}
// 	return desugared{sp: sp}
// }

type sugared struct {
	p types.ConnProvider
}

var _ types.SugaredConnProvider = sugared{}
var _ types.ConnProviderMeta = sugared{}

func (s sugared) R(ctx context.Context) (types.Conn, xoptional.T[types.ConnOwnership], error) {
	return s.p.Acquire(ctx, types.ConnMode_RO, conns.AcquireWithRWFallback())
}

func (s sugared) RO(ctx context.Context) (types.Conn, xoptional.T[types.ConnOwnership], error) {
	return s.p.Acquire(ctx, types.ConnMode_RO)
}

func (s sugared) RW(ctx context.Context) (types.Conn, xoptional.T[types.ConnOwnership], error) {
	return s.p.Acquire(ctx, types.ConnMode_RW)
}

func (s sugared) ManagedR(ctx context.Context) (types.Conn, error) {
	return s.p.AcquireManaged(ctx, types.ConnMode_RO, conns.AcquireWithRWFallback())
}

func (s sugared) ManagedRO(ctx context.Context) (types.Conn, error) {
	return s.p.AcquireManaged(ctx, types.ConnMode_RO)
}

func (s sugared) ManagedRW(ctx context.Context) (types.Conn, error) {
	return s.p.AcquireManaged(ctx, types.ConnMode_RW)
}

func (s sugared) Type() string {
	var paramType string
	if pm, ok := s.p.(types.ConnProviderMeta); ok {
		paramType = pm.Type()
	} else {
		paramType = fmt.Sprintf("<%T>", s.p)
	}
	return fmt.Sprintf("%s[%s]", sugaredTypeBase, paramType)
}

func (s sugared) GenericType() string {
	return fmt.Sprintf("%s[ConnProvider]", sugaredTypeBase)
}

func (s sugared) AcquireType() xoptional.T[string] {
	if pm, ok := s.p.(types.ConnProviderMeta); ok {
		return pm.AcquireType()
	}
	return xoptional.New[string]()
}

type sugaredSilent struct {
	p types.SilentConnProvider
}

func (s sugaredSilent) R(ctx context.Context) (types.Conn, xoptional.T[types.ConnOwnership]) {
	return s.p.Acquire(ctx, types.ConnMode_RO, conns.AcquireWithRWFallback())
}

func (s sugaredSilent) RO(ctx context.Context) (types.Conn, xoptional.T[types.ConnOwnership]) {
	return s.p.Acquire(ctx, types.ConnMode_RO)
}

func (s sugaredSilent) RW(ctx context.Context) (types.Conn, xoptional.T[types.ConnOwnership]) {
	return s.p.Acquire(ctx, types.ConnMode_RW)
}

func (s sugaredSilent) ManagedR(ctx context.Context) types.Conn {
	return s.p.AcquireManaged(ctx, types.ConnMode_RO, conns.AcquireWithRWFallback())
}

func (s sugaredSilent) ManagedRO(ctx context.Context) types.Conn {
	return s.p.AcquireManaged(ctx, types.ConnMode_RO)
}

func (s sugaredSilent) ManagedRW(ctx context.Context) types.Conn {
	return s.p.AcquireManaged(ctx, types.ConnMode_RW)
}

type desugared struct {
	sp types.SugaredConnProvider
}

// TODO: impl
// var _ types.ConnProvider = desugared{}
var _ types.ConnProviderMeta = desugared{}

func (d desugared) Acquire(ctx context.Context, mode types.ConnMode) (types.Conn, xoptional.T[types.ConnOwnership], error) {
	switch mode {
	case types.ConnMode_RO:
		return d.sp.RO(ctx)
	case types.ConnMode_RW:
		return d.sp.RW(ctx)
	}
	panic(fmt.Errorf("unexpected mode %s", mode.String()))
}

func (d desugared) AcquireManaged(ctx context.Context, mode types.ConnMode) (types.Conn, error) {
	switch mode {
	case types.ConnMode_RO:
		return d.sp.ManagedRO(ctx)
	case types.ConnMode_RW:
		return d.sp.ManagedRW(ctx)
	}
	panic(fmt.Errorf("unexpected mode %s", mode.String()))
}

func (s desugared) Type() string {
	var paramType string
	if pm, ok := s.sp.(types.ConnProviderMeta); ok {
		paramType = pm.Type()
	} else {
		paramType = fmt.Sprintf("<%T>", s.sp)
	}
	return fmt.Sprintf("%s[%s]", desugaredTypeBase, paramType)
}

func (s desugared) GenericType() string {
	return fmt.Sprintf("%s[ConnProvider]", desugaredTypeBase)
}

func (s desugared) AcquireType() xoptional.T[string] {
	if pm, ok := s.sp.(types.ConnProviderMeta); ok {
		return pm.AcquireType()
	}
	return xoptional.New[string]()
}

const (
	sugaredTypeBase   = "sugared"
	desugaredTypeBase = "desugared"
)
