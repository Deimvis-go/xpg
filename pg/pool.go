package pg

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func NewPostgresConnectionPool(pgxpoolCfg *pgxpool.Config, logger *zap.SugaredLogger) *pgxpool.Pool {
	logger.Infow("Will create pgx pool", "config", fmt.Sprintf("%+v", pgxpoolCfg))
	pool, err := pgxpool.NewWithConfig(context.Background(), pgxpoolCfg)
	if err != nil {
		panic(fmt.Errorf("failed to connect to PostgreSQL database: %w", err))
	}
	return pool
}
