package pgfw

import (
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/Deimvis-go/xpg/pg"
)

// NOTE: maybe rewrite TryGet and TryGetPtr as xpgfb.OnNoRows(v, err, nil, err) and xpgfb.OnNoRowsToPtr(v, err, nil, err)

func TryGet[T any](v T, err error) (T, error) {
	if err != nil {
		var anyv T
		if pg.IsNoRows(err) {
			return anyv, nil
		}
		return anyv, err
	}
	return v, nil
}

func TryGetPtr[T any](v T, err error) (*T, error) {
	if err != nil {
		if pg.IsNoRows(err) {
			return nil, nil
		}
		return nil, err
	}
	return &v, nil
}

func Has[T any](v T, err error) (bool, error) {
	var anyValue bool
	if err != nil {
		if pg.IsNoRows(err) {
			return false, nil
		}
		return anyValue, err
	}
	return true, nil
}

func Add(cmdTag pgconn.CommandTag, err error) (bool, error) {
	if err != nil {
		if pg.IsUniqueViolation(err) {
			return false, nil
		}
		return false, err
	}
	return cmdTag.RowsAffected() != 0, nil
}

func AddBatch(rowsAdded int64, err error) (bool, error) {
	if err != nil {
		if pg.IsUniqueViolation(err) {
			return false, nil
		}
		return false, err
	}
	return rowsAdded != 0, nil
}
