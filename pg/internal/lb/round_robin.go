package lb

import (
	"math"
	"math/rand/v2"
	"slices"
	"sync/atomic"
	"time"

	"github.com/Deimvis/go-ext/go1.25/xcheck/xmust"
	"github.com/Deimvis/go-ext/go1.25/xoptional"
	"github.com/Deimvis/go-ext/go1.25/xslices"
)

func NewRRHostsLB(hosts []string, rand xoptional.T[Rand]) *roundRobinHostsLoadBalancer {
	xmust.Lt(uint64(len(hosts)), math.MaxUint64/2, "too many hosts")
	hosts = slices.Clone(hosts)
	if rand.HasValue() {
		rand.Value().Shuffle(len(hosts), xslices.NewSwapFn(hosts))
	}
	return &roundRobinHostsLoadBalancer{hosts: hosts}
}

type roundRobinHostsLoadBalancer struct {
	hosts []string

	i atomic.Uint64
}

var _ HostsLoadBalancer = (*roundRobinHostsLoadBalancer)(nil)
var _ DeterministicHostsLoadBalancer = (*roundRobinHostsLoadBalancer)(nil)
var _ ForwardPrecomputedHostsLoadBalancer = (*roundRobinHostsLoadBalancer)(nil)

func (rr *roundRobinHostsLoadBalancer) Next() string {
	nextInd := rr.advanceForward_fast(1)
	var ind uint64
	if nextInd == 0 {
		ind = uint64(len(rr.hosts)) - 1
	} else {
		ind = nextInd - 1
	}
	return rr.hosts[ind]
}

func (rr *roundRobinHostsLoadBalancer) WithSeed(seed uint64) DeterministicHostsLoadBalancer {
	hosts := slices.Clone(rr.hosts)
	rand := rand.New(rand.NewPCG(seed, uint64(time.Now().UnixNano())))
	rand.Shuffle(len(hosts), xslices.NewSwapFn(hosts))
	return &roundRobinHostsLoadBalancer{hosts: hosts}
}

func (rr *roundRobinHostsLoadBalancer) AdvanceForward(delta uint64) {
	rr.advanceForward_fast(delta)
}

// _fast expects that number of calls never overflows max uint64 (~1.8x10^19)
// 5-year capacity: 114 billions of calls per second.
func (rr *roundRobinHostsLoadBalancer) advanceForward_fast(delta uint64) uint64 {
	return rr.i.Add(delta) % uint64(len(rr.hosts))
}

// _safe is always correct, but slower, especially in highly concurrent env.
func (rr *roundRobinHostsLoadBalancer) advanceForward_safe(delta uint64) uint64 {
	n := uint64(len(rr.hosts))
	if n == 1 {
		return 0 // :)
	}
	delta = delta % n
	calcTarget := func(cur uint64) uint64 {
		// safe calc avoiding cur+delta > math.MaxUint64
		cur = cur % n
		remain := n - cur
		if delta < remain {
			return cur + delta
		} else {
			return delta - remain
		}
	}
	cur := rr.i.Load()
	target := calcTarget(cur)
	for !rr.i.CompareAndSwap(cur, target) {
		cur = rr.i.Load()
		target = calcTarget(cur)
	}
	return target
}
