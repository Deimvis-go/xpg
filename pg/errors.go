package pg

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	// IMPORTANT: use pgconn from pgx/v5, otherwise cast to pgconn.PgError will fail
)

const (
	UniqueViolationCode = "23505"
)

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	ok := errors.As(err, &pgErr)
	if !ok {
		return false
	}
	return pgErr.Code == UniqueViolationCode
}

func IsNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
