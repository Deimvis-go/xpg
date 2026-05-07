package pgfx

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/Deimvis-go/xpg/pg"
)

func NewPostgresConnectionPool(lc fx.Lifecycle, pgxpoolCfg *pgxpool.Config, logger *zap.SugaredLogger) *pgxpool.Pool {
	pool := pg.NewPostgresConnectionPool(pgxpoolCfg, logger)
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			pool.Close()
			return nil
		},
	})
	return pool
}
