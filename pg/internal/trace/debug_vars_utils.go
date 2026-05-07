package trace

import (
	"encoding/json"
	"expvar"
	"sync"

	"github.com/Deimvis/go-ext/go1.25/xcheck/xmust"
)

// TODO: more efficient impl:
// impl similar to semaphore internals:
//   - String() sets flag that forbids any operations,
//     then sleeps until current workers finish
//   - Adding new acquire requires write lock to root map,
//
// then updating single acquire doesn't require sync
// since all operations to single acquire are considered to be single threaded
type objVar struct {
	m  map[string]any
	mu sync.RWMutex
}

func (ov *objVar) Write(fn func(map[string]any)) {
	ov.mu.Lock()
	defer ov.mu.Unlock()
	fn(ov.m)
}

func (ov *objVar) String() string {
	ov.mu.RLock()
	defer ov.mu.RUnlock()
	return string(xmust.Do(json.Marshal(ov.m)))
}

type arrVar[T expvar.Var] []T

func (av *arrVar[T]) String() string {
	return string(xmust.Do(json.Marshal(av)))
}
