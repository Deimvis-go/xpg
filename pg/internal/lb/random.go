package lb

import (
	"math/rand/v2"
	"time"
)

func NewRandomHostsLB(hosts []string, rand Rand) *randomHostsLoadBalancer {
	return &randomHostsLoadBalancer{hosts: hosts, rand: rand}
}

type randomHostsLoadBalancer struct {
	hosts []string
	rand  Rand
}

var _ HostsLoadBalancer = (*randomHostsLoadBalancer)(nil)
var _ DeterministicHostsLoadBalancer = (*randomHostsLoadBalancer)(nil)
var _ ForwardPrecomputedHostsLoadBalancer = (*randomHostsLoadBalancer)(nil)

func (rnd *randomHostsLoadBalancer) Next() string {
	ind := rnd.rand.IntN(len(rnd.hosts))
	return rnd.hosts[ind]
}

func (rnd *randomHostsLoadBalancer) WithSeed(seed uint64) DeterministicHostsLoadBalancer {
	return &randomHostsLoadBalancer{
		hosts: rnd.hosts,
		rand:  rand.New(rand.NewPCG(seed, uint64(time.Now().UnixNano()))),
	}
}

func (rnd *randomHostsLoadBalancer) AdvanceForward(delta uint64) {}
