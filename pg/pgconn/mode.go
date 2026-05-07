package pgconn

import (
	"fmt"

	"github.com/jackc/pgx/v5"
	pgxconn "github.com/jackc/pgx/v5/pgconn"
	"github.com/Deimvis/go-ext/go1.25/xptr"
	"github.com/Deimvis-go/xpg/pg/internal/types"
)

type Mode = types.ConnMode

const (
	RO Mode = types.ConnMode_RO
	RW Mode = types.ConnMode_RW
)

// TODO: measure perf
func TryExtractConnMode(connConfig *pgx.ConnConfig) *Mode {
	vc := connConfig.ValidateConnect
	if vc == nil {
		return nil
	}
	match := func(fn1, fn2 pgxconn.ValidateConnectFunc) bool {
		return fmt.Sprintf("%v", fn1) == fmt.Sprintf("%v", fn2)
	}
	if match(vc, pgxconn.ValidateConnectTargetSessionAttrsReadWrite) ||
		match(vc, pgxconn.ValidateConnectTargetSessionAttrsPrimary) {
		return xptr.T(RW)
	}
	if match(vc, pgxconn.ValidateConnectTargetSessionAttrsReadOnly) ||
		match(vc, pgxconn.ValidateConnectTargetSessionAttrsStandby) {
		return xptr.T(RO)
	}
	return nil
}
