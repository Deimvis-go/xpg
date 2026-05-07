package lb

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoundRobinHostsLoadBalancer(t *testing.T) {
	hosts := []string{"a", "b", "c"}
	hlb := NewHostsLoadBalancer(hosts, LBA_RoundRobin)
	for i := range 10 {
		h := hlb.Next()
		require.Equal(t, hosts[i%len(hosts)], h)
	}
}

func TestRoundRobinHostsLoadBalancer_Concurrency(t *testing.T) {
	hosts := []string{"a", "b", "c"}
	hlb := NewHostsLoadBalancer(hosts, LBA_RoundRobin)

	threads := 100
	callsPerThread := 1000

	wg := sync.WaitGroup{}
	for range threads {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range callsPerThread {
				h := hlb.Next()
				require.Contains(t, hosts, h)
			}
		}()
	}
	wg.Wait()
}
