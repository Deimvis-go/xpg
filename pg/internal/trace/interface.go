package trace

import (
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Tracer interface {
	pgx.QueryTracer
	pgx.ConnectTracer
	pgxpool.AcquireTracer
	pgxpool.ReleaseTracer
}
