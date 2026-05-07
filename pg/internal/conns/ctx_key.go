package conns

import (
	"errors"
	"fmt"

	"github.com/Deimvis-go/xpg/pg/internal/types"
)

func CtxKey(mode types.ConnMode) string {
	return fmt.Sprintf("pg_conn_%s_ctx_key__github.com/Deimvis-go/xpg/pg", mode)
}

var ErrNoCtxConn = errors.New("no conn in context")
