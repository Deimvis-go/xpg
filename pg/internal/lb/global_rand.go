package lb

import "math/rand/v2"

type Rand interface {
	IntN(n int) int
	Shuffle(n int, swap func(i, j int))
}

// NOTE: use this hacky approach sicne math/rand/v2.globalRand is unexported

type globalRandProxy struct{}

var _ Rand = (*globalRandProxy)(nil)

func (r *globalRandProxy) IntN(n int) int {
	return rand.IntN(n)
}

func (r *globalRandProxy) Shuffle(n int, swap func(i, j int)) {
	rand.Shuffle(n, swap)
}

var (
	globalRand Rand = &globalRandProxy{}
)
