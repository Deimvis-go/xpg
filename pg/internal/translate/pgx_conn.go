package translate

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Deimvis/go-ext/go1.25/xoptional"
	"github.com/Deimvis-go/xpg/pg/internal/types"
)

func FromPgxConnToStandaloneConn(pgxconn *pgx.Conn, mode types.ConnMode) types.StandaloneConn {
	var sconn types.StandaloneConn = pgxconn
	conn := types.NewConnReflectInplace(sconn, types.NewConnMetaInplace(mode, xoptional.New(false)))
	return conn.(types.StandaloneConn)
}

func FromPgxConnToConn(pgxconn *pgx.Conn, mode types.ConnMode) (types.Conn, types.ConnOwnership) {
	var sconn types.StandaloneConn = pgxconn
	freeFn := sconn.Close
	meta := types.NewConnMetaInplace(mode, xoptional.New(false))
	meta.OwnershipTaken_.SetValue(false)
	meta.OwnedConn_ = types.NewOwnedConnInplace(freeFn)
	conn := types.NewConnReflectInplace(sconn, meta)
	return conn, types.NewConnOwnershipInplace(meta)
}

func FromPgxpoolConnToPoolConn(pgxconn *pgxpool.Conn, mode types.ConnMode) types.PoolConn {
	var sconn types.PoolConn = pgxconn
	conn := types.NewConnReflectInplace(sconn, types.NewConnMetaInplace(mode, xoptional.New(false)))
	return conn.(types.PoolConn)
}

func FromPgxpoolConnToConn(pgxconn *pgxpool.Conn, mode types.ConnMode) (types.Conn, types.ConnOwnership) {
	var sconn types.PoolConn = pgxconn
	freeFn := func(context.Context) error {
		sconn.Release()
		return nil
	}
	meta := types.NewConnMetaInplace(mode, xoptional.New(false))
	meta.OwnershipTaken_.SetValue(false)
	meta.OwnedConn_ = types.NewOwnedConnInplace(freeFn)
	conn := types.NewConnReflectInplace(sconn, types.NewConnMetaInplace(mode, xoptional.New(false)))
	return conn, types.NewConnOwnershipInplace(meta)
}

func FromPgxPoolToConn(pgxpool *pgxpool.Pool, mode types.ConnMode) types.Conn {
	return types.NewConnReflectInplace(pgxpool, types.NewConnMetaInplace(mode, xoptional.New(true)))
}
