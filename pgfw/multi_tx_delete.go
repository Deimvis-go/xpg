package pgfw

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/Deimvis-go/xpg/pg"
)

// NOTE: very experimental.
// MultiTxDelete performs row deletion
// in multiple sequential transactions.
func MultiTxDelete(conn pg.Conn, ctx context.Context, p MultiTxDeleteParams) error {
	// TODO: escape to avoid sql injection
	// TODO: support query args to canonize queries
	query := fmt.Sprintf(multiTxDeleteQuery,
		p.IdColumn, p.Table, p.DeleteCondition, p.MaxRowsPerTx,
		p.Table, p.IdColumn, p.IdColumn,
	)
	for {
		cmdTag, err := conn.Exec(ctx, query)
		if err != nil {
			return err
		}
		if p.QueryFinishHook != nil {
			p.QueryFinishHook(ctx, cmdTag)
		}
		if uint64(cmdTag.RowsAffected()) < p.MaxRowsPerTx {
			break
		}
		time.Sleep(p.TxInterval)
	}
	return nil
}

type MultiTxDeleteParams struct {
	Table           string
	IdColumn        string
	DeleteCondition string
	MaxRowsPerTx    uint64
	TxInterval      time.Duration
	QueryFinishHook func(context.Context, pgconn.CommandTag)
}

const (
	multiTxDeleteQuery = `
WITH to_delete AS (
    SELECT "%s"
    FROM "%s"
    WHERE %s
    LIMIT %d
)
DELETE FROM "%s" as t1
USING to_delete as t2
WHERE (t1."%s" = t2."%s")
`
)
