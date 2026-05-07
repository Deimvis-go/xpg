package pgconnprovider

import (
	"context"

	"github.com/Deimvis/go-ext/go1.25/xoptional"
	"github.com/Deimvis-go/xpg/pg/internal/types"
)

func Silence(p types.ConnProvider, onErr func(error)) types.SilentConnProvider {
	return silencedProvider{p: p, onErr: onErr}
}

type silencedProvider struct {
	p     types.ConnProvider
	onErr func(error)
}

func (sp silencedProvider) Acquire(ctx context.Context, m types.ConnMode, opts ...types.ConnAcquireOption) (types.Conn, xoptional.T[types.ConnOwnership]) {
	conn, free, err := sp.p.Acquire(ctx, m, opts...)
	if err != nil {
		sp.onErr(err)
	}
	return conn, free
}

func (sp silencedProvider) AcquireManaged(ctx context.Context, m types.ConnMode, opts ...types.ConnAcquireOption) types.Conn {
	conn, err := sp.p.AcquireManaged(ctx, m, opts...)
	if err != nil {
		sp.onErr(err)
	}
	return conn
}
