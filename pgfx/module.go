package pgfx

import (
	"go.uber.org/fx"

	"github.com/Deimvis-go/xpg/pg"
)

// FX Module.
//
// Requires:
// *pg.ConnConfig
// *pg.PoolConfig[`name:"ro"`]
// *pg.PoolConfig[`name:"rw"`]
// pg.AfterConnectFn (optional)
// *zap.SugaredLogger
//
// Provides:
// pg.PoolManager
var Module = fx.Module("pg",
	fx.Provide(
		fx.Private,
		NewPgxpoolConfigRO,
		fx.Annotate(
			pg.NewPostgresConnectionPool,
			fx.ParamTags(`name:"ro"`, ``),
			fx.ResultTags(`name:"ro"`),
		),
	),
	fx.Provide(
		fx.Private,
		NewPgxpoolConfigRW,
		fx.Annotate(
			pg.NewPostgresConnectionPool,
			fx.ParamTags(`name:"rw"`, ``),
			fx.ResultTags(`name:"rw"`),
		),
	),
	fx.Provide(
		pg.NewPoolManager,
	),
)

// TODO: add NewModule function, allowing to use multiple
// pgfx.Module within single options.
// It should allow passing prefix (aka namespace),
// so that module will use arguments provided by
// "namespace::"+name and return
// values also prefixed with namespace.
//
// func NewModule(prefix string) fx.Option {
// 	newPgxpoolConfigRO := NewPgxpoolConfigRO
// 	if prefix != "" {
//
// 	}
// 	return fx.Module(prefix+"pg",
// 		fx.Provide(
// 			fx.Private,
// 			NewPgxpoolConfigRO,
// 			fx.Annotate(
// 				pg.NewPostgresConnectionPool,
// 				fx.ParamTags(`name:"ro"`, ``),
// 				fx.ResultTags(`name:"ro"`),
// 			),
// 		),
// 		fx.Provide(
// 			fx.Private,
// 			NewPgxpoolConfigRW,
// 			fx.Annotate(
// 				pg.NewPostgresConnectionPool,
// 				fx.ParamTags(`name:"rw"`, ``),
// 				fx.ResultTags(`name:"rw"`),
// 			),
// 		),
// 		fx.Provide(
// 			pg.NewPoolManager,
// 		),
// 	)
// }
