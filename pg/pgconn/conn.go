package pgconn

import "github.com/Deimvis-go/xpg/pg/internal/types"

type Conn = types.Conn
type StandaloneConn = types.StandaloneConn
type PoolConn = types.PoolConn
type ConnReflect = types.ConnReflect
type ConnMeta = types.ConnMeta
type OwnedConn = types.OwnedConn

func RevealReflect(c Conn) (ConnReflect, bool) {
	if cr, ok := c.(ConnReflect); ok {
		return cr, true
	}
	return nil, false
}
