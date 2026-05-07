package lb

// TODO: move to xalgo/xsched (or smth like this) as SchedulingAlgorithm

type Algo string

var (
	LBA_RoundRobin Algo = "round_robin"
	LBA_Random     Algo = "random"
)
