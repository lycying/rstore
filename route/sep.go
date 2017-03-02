package route

import (
	"hash/fnv"
	"math/rand"
	"strconv"
	"strings"
)

type SepRule struct {
	Rule

	Slots   []string
	OptSlot int
}

func NewSepRule(key string, sep string, optSlot int, units []EndPointer) *SepRule {
	r := &SepRule{}
	r.Units = units
	r.N = len(r.Units)

	r.Slots = strings.Split(key, sep)
	r.OptSlot = optSlot
	return r
}

func (r *SepRule) RoundRobin() (EndPointer, error) {
	n := rand.Intn(r.N)
	return r.Units[n], nil
}

func (r *SepRule) Hash() (EndPointer, error) {
	slot := r.Slots[r.OptSlot]

	h := fnv.New32a()
	h.Write([]byte(slot))
	hs := int(h.Sum32())

	return r.Units[hs%r.N], nil
}

func (r *SepRule) LastNNumberHash(n int) (EndPointer, error) {
	slot := r.Slots[r.OptSlot]
	lastN := slot[len(slot)-n:]
	hs, err := strconv.Atoi(lastN)
	if err != nil {
		return nil, err
	}
	return r.Units[hs%r.N], nil
}

func (r *SepRule) FirstNNumberHash(n int) (EndPointer, error) {
	slot := r.Slots[r.OptSlot]
	lastN := slot[len(slot)-n:]
	hs, err := strconv.Atoi(lastN)
	if err != nil {
		return nil, err
	}
	return r.Units[hs%r.N], nil
}
