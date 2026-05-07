package types

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Pool interface {
	// TODO: wrap (impl translate package and migrate pgfx to use these interfaces)
	// Acquire(ctx context.Context) (Conn, error)
	Acquire(ctx context.Context) (*pgx.Conn, error)
}

// TODO: consider renaming to PoolProvider
type PoolManager interface {
	// TODO: wrap (impl translate package and migrate pgfx to use these interfaces)
	GetPool(ConnMode) *pgxpool.Pool
}
