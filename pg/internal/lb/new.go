package lb

import (
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/Deimvis/go-ext/go1.25/xcheck/xmust"
	"github.com/Deimvis/go-ext/go1.25/xoptional"
)

// TODO: write generic load balancer algorithm with item type parameter and use it there

type HostsLoadBalancer interface {
	Next() string
}

type DeterministicHostsLoadBalancer interface {
	HostsLoadBalancer
	// WithSeed creates a clone with given seed
	// being set.
	WithSeed(seed uint64) DeterministicHostsLoadBalancer
}

type ForwardPrecomputedHostsLoadBalancer interface {
	HostsLoadBalancer
	// AdvanceForward makes load balancer
	// to skip N results, as it was called
	// Next() N times.
	AdvanceForward(uint64)
}

func NewHostsLoadBalancer(hosts []string, algo Algo, opts ...HostsLoadBalancerOption) HostsLoadBalancer {
	cfg := hostsLoadBalancerCfg{}
	for _, opt := range opts {
		opt(&cfg)
	}

	xmust.True(len(hosts) > 0)
	rand_ := xoptional.New[Rand]()
	if cfg.seed.HasValue() {
		src := rand.NewPCG(cfg.seed.Value(), uint64(time.Now().UnixNano()))
		rand_.SetValue(rand.New(src))
	}

	switch algo {
	case LBA_RoundRobin:
		return NewRRHostsLB(hosts, rand_)
	case LBA_Random:
		return NewRandomHostsLB(hosts, xoptional.ValueOr(rand_, globalRand))
	}
	panic(fmt.Errorf("got unexpected hosts load balancing algo: %s", algo))
}

type HostsLoadBalancerOption func(c *hostsLoadBalancerCfg)

// WithSeed makes hosts load balancer output determinist
func WithSeed(seed uint64) HostsLoadBalancerOption {
	return func(c *hostsLoadBalancerCfg) {
		c.seed = xoptional.New(seed)
	}
}

type hostsLoadBalancerCfg struct {
	seed xoptional.T[uint64]
}
