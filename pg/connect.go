package pg

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

func NewPostgresConnection() *pgx.Conn {
	con, err := pgx.Connect(context.Background(), os.Getenv("POSTGRES_URL"))
	if err != nil {
		panic(fmt.Errorf("failed to connect to PostgreSQL database: %w", err))
	}
	return con
}
