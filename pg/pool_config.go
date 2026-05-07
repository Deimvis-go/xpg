package pg

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/Deimvis/go-ext/go1.25/xoptional"
	"github.com/Deimvis/models/utility/golang/dmutil"
)

type AfterConnectFn func(ctx context.Context, conn *pgx.Conn) error

type PoolConfig struct {
	MinConns               *int32             `yaml:"min_conns"`
	MaxConns               *int32             `yaml:"max_conns"`
	MaxConnLifetimeS       xoptional.T[int64] `yaml:"max_conn_lifetime_s"`
	MaxConnLifetimeJitterS *int               `yaml:"max_conn_lifetime_jitter_s"`
	ConnectTimeoutMs       xoptional.T[int64] `yaml:"connect_timeout_ms"`
	// TODO: make Tracers required
	Tracers *PoolTracersConfig `yaml:"tracers"`

	// TODO: add fields, wrap original Acquire to use these ?
	// AcquireTimeoutMs xoptional.T[int64] `yaml:"acquire_timeout_ms"`
	// AcquireAttemptTimeoutMs xoptional.T[int64] `yaml:"acquire_timeout_ms"`
}

type PoolTracersConfig struct {
	Logging    *LoggingTracerConfig    `yaml:"logging"`
	Prometheus *PrometheusTracerConfig `yaml:"prometheus"`
	DebugVars  *DebugVarsTracerConfig  `yaml:"debug_vars"`
}

type LoggingTracerConfig struct {
	dmutil.Option `yaml:",inline"`
}

type PrometheusTracerConfig struct {
	dmutil.Option `yaml:",inline"`
}

type DebugVarsTracerConfig struct {
	dmutil.Option      `yaml:",inline"`
	VarsRootNameFormat string `yaml:"vars_root_name_format"`
}

func (dvt *DebugVarsTracerConfig) ValidateSelf() error {
	if !strings.Contains(dvt.VarsRootNameFormat, "%s") {
		return errors.New("vars root name format must contain '%s' for pool name")
	}
	return nil
}
