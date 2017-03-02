package route

import ()

type RulePath interface {
	RoundRobin() (EndPointer, error)
	Hash() (EndPointer, error)
	FirstNNumberHash(n int) (EndPointer, error)
	LastNNumberHash(n int) (EndPointer, error)
}

type Rule struct {
	RulePath
	Units []EndPointer
	N     int
}
