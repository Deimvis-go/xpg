package pg

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
)

type PoolManagerParams struct {
	fx.In

	PoolRO *pgxpool.Pool `name:"ro"`
	PoolRW *pgxpool.Pool `name:"rw"`
}

func NewPoolManager(p PoolManagerParams) PoolManager {
	pm := &poolManager{
		poolRO: p.PoolRO,
		poolRW: p.PoolRW,
	}
	return pm
}

type poolManager struct {
	poolRO *pgxpool.Pool
	poolRW *pgxpool.Pool
}

func (pm *poolManager) GetPool(mode ConnMode) *pgxpool.Pool {
	switch mode {
	case CM_RO:
		return pm.poolRO
	case CM_RW:
		return pm.poolRW
	}
	panic(fmt.Errorf("unsupported conn mode: %s", mode))
}
