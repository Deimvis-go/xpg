package pgfx

import (
	"context"

	"github.com/jackc/pgx/v5"
	"go.uber.org/fx"

	"github.com/Deimvis-go/xpg/pg"
)

func NewPostgresConnection(lc fx.Lifecycle) *pgx.Conn {
	con := pg.NewPostgresConnection()
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return con.Close(ctx)
		},
	})
	return con
}
