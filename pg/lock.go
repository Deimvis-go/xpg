package pg

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func AdvXactLock(ctx context.Context, tx pgx.Tx, key string) error {
	_, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock( hashtext($1) );", key)
	return err
}

func AdvXactUnlock(ctx context.Context, tx pgx.Tx, key string) error {
	_, err := tx.Exec(ctx, "SELECT pg_advisory_xact_unlock( hashtext($1) )", key)
	return err
}

func TryAdvXactLock(ctx context.Context, tx pgx.Tx, key string) (acquired bool, err error) {
	row := tx.QueryRow(ctx, "SELECT pg_try_advisory_xact_lock( hashtext($1) )", key)
	err = row.Scan(&acquired)
	return acquired, err
}
