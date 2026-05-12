package pgfw

import (
	"context"

	"github.com/Deimvis-go/xpg/pg"
	"github.com/Deimvis/models/utility/go/dmutil"
	"github.com/jackc/pgx/v5"
)

type SearchResultRow interface {
	GetTotalCount() int
}

type MakeSqlQueryFn func(onlySize bool, customPagination *dmutil.OffsetLimitPagination) string

func Search[SelectT SearchResultRow](db pg.PG, ctx context.Context, query *string, offset int64, makeSqlQuery MakeSqlQueryFn) (parsedRows []SelectT, totalCount int, err error) {
	var args []interface{}
	if query != nil {
		args = append(args, *query)
	}
	rows, err := db.Query(ctx, makeSqlQuery(false /*only_count*/, nil), args...)
	if err != nil {
		return nil, 0, err
	}
	parsedRows, err = pgx.CollectRows(rows, pgx.RowToStructByName[SelectT])
	if err != nil {
		return nil, 0, err
	}

	if len(parsedRows) > 0 {
		totalCount = parsedRows[0].GetTotalCount()
	} else {
		if offset == 0 {
			totalCount = 0
		} else {
			customPagination := dmutil.OffsetLimitPagination{
				Offset: 0,
				Limit:  1,
			}
			err = db.QueryRow(ctx, makeSqlQuery(true /* only_count */, &customPagination), args...).Scan(&totalCount)
			if err != nil {
				if pg.IsNoRows(err) {
					totalCount = 0
				} else {
					return nil, 0, err
				}
			}
		}
	}

	return parsedRows, totalCount, nil
}
