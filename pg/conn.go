package pg

import (
	"github.com/Deimvis-go/xpg/pg/internal/types"
	"github.com/Deimvis-go/xpg/pg/pgconn"
)

type Conn = types.Conn
type ConnMode = types.ConnMode
type ConnOwnership = types.ConnOwnership

var (
	CM_RO ConnMode = types.ConnMode_RO
	CM_RW ConnMode = types.ConnMode_RW
)

var TryExtractConnMode = pgconn.TryExtractConnMode
