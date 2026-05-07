package pgfw

import (
	"github.com/Deimvis/go-ext/go1.25/xcheck/xmust"
	"github.com/Deimvis-go/xpg/pg"
)

type forEachPGSettings struct {
	parallel bool
	untilFn  func(any, error) bool // finished when untilFn returns true
}

type ForEachPGOption func(*forEachPGSettings)

// WithSequential - aka not parallel
func WithSequential() ForEachPGOption {
	return func(opts *forEachPGSettings) {
		opts.parallel = false
	}
}

func WithUntilFound() ForEachPGOption {
	return func(opts *forEachPGSettings) {
		opts.untilFn = func(_ any, err error) bool {
			return err == nil || !pg.IsNoRows(err)
		}
	}
}

func ForEachPG[T any](fn func(pg.PG) (T, error), pgs []pg.PG, opts ...ForEachPGOption) (T, error) {
	settings := forEachPGSettings{
		parallel: false,
		untilFn:  nil,
	}
	for _, opt := range opts {
		opt(&settings)
	}

	xmust.True(!settings.parallel, "parallel is not supported yet")
	var v T
	var err error
	for _, pg := range pgs {
		v, err = fn(pg)
		if settings.untilFn != nil && settings.untilFn(v, err) {
			break
		}
	}
	return v, err
}

func ForEachPGLazy[T any](fn func(pg.PG) (T, error), lazyPgs []func() (pg.PG, error), opts ...ForEachPGOption) (T, error) {
	settings := forEachPGSettings{
		parallel: false,
		untilFn:  nil,
	}
	for _, opt := range opts {
		opt(&settings)
	}

	xmust.True(!settings.parallel, "parallel is not supported yet")
	var v T
	var err error
	for _, lpg := range lazyPgs {
		var pg pg.PG
		pg, err = lpg()
		if err != nil {
			break
		}
		v, err = fn(pg)
		if settings.untilFn != nil && settings.untilFn(v, err) {
			break
		}
	}
	return v, err
}
