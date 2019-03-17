// bloom.go - Bloom filter.
// Written in 2015 by Yawning Angel; 2019 by Will Scott
//
// To the extent possible under law, the author(s) have dedicated all copyright
// and related and neighboring rights to this software to the public domain
// worldwide. This software is distributed without any warranty.
//
// You should have received a copy of the CC0 Public Domain Dedication along
// with this software. If not, see <http://creativecommons.org/publicdomain/zero/1.0/>.

// Package bloom implements a Bloom Filter.
package bloom

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math/bits"
	"strconv"

	"github.com/dchest/siphash"
)

// Filter is a delta-compressable bloom filter.
// following the logic from http://www.eecs.harvard.edu/~michaelm/NEWWORK/postscripts/cbf2.pdf
type Filter struct {
	b        [][]byte

	k1, k2 uint64

	mask uint64
	nrEntriesMax int
	nrEntries    []int
}

// New constructs a new Filter with a filter set size of 2^mLn2
// which allows an entry factor up to load before dropping layers
// at new deltas.
func New(rand io.Reader, mLn2 int, load float64) (*Filter, error) {
	const maxMln2 = strconv.IntSize - 1

	var key [16]byte
	if _, err := io.ReadFull(rand, key[:]); err != nil {
		return nil, err
	}

	if load <= 0.0 || load > 1.0 {
		return nil, fmt.Errorf("invalid load rate: %v", load)
	}

	if mLn2 > maxMln2 {
		return nil, fmt.Errorf("requested filter too large: %d", mLn2)
	}

	m := uint64(1 << uint64(mLn2))
	n := float64(m) * load

	if uint64(n) > (1 << uint(maxMln2)) {
		return nil, fmt.Errorf("requested filter too large (nrEntriesMax overflow): %d", mLn2)
	}

	f := new(Filter)
	f.k1 = binary.BigEndian.Uint64(key[0:8])
	f.k2 = binary.BigEndian.Uint64(key[8:16])
	f.mask = m - 1
	f.nrEntriesMax = int(n)
	f.b = [][]byte{make([]byte, m/8)}
	f.nrEntries = make([]int, 1)
	return f, nil
}

// MaxEntries returns the maximum capacity of the Filter.
func (f *Filter) MaxEntries() int {
	return f.nrEntriesMax
}

// Entries returns the number of entries that have been inserted into the
// Filter.
func (f *Filter) Entries() int {
	entries := 0
	for i := 0; i < len(f.nrEntries); i++ {
		entries += f.nrEntries[i]
	}
	return entries
}

// TestAndSet tests the Filter for a given value's membership, adds the value
// to the filter, and returns true iff it was present at the time of the call.
func (f *Filter) TestAndSet(b []byte) bool {
	h := f.hash(b)
	// Just return true iff the entry is present.
	if f.test(h) {
		return true
	}

	// Add and return false.
	f.add(h)
	f.nrEntries[0]++
	return false
}

func (f *Filter) Import(layer []byte) error {
	if len(layer) != len(f.b[0]) {
		return errors.New("Invalid layer size")
	}
	f.b = append([][]byte{layer}, f.b...)
	c := f.count(layer)
	f.nrEntries = append([]int{c}, f.nrEntries...)
	f.checkExpiry()
	return nil
}

func (f *Filter) Delta() []byte {
	newLayer := make([]byte, len(f.b[0]))
	f.b = append([][]byte{newLayer}, f.b...)
	f.nrEntries = append([]int{0}, f.nrEntries...)
	f.checkExpiry()
	return f.b[1]
}

// Test tests the Filter for a given value's membership and returns true iff
// it is present (or a false positive).
func (f *Filter) Test(b []byte) bool {
	return f.test(f.hash(b))
}


func (f *Filter) count(b []byte) int {
	var cnt int
	for i := 0; i < len(b); i++ {
		cnt += bits.OnesCount8(b[i])
	}
	return cnt
}

func (f *Filter) checkExpiry() {
	ecnt := len(f.nrEntries)
	ecntM := float64(ecnt + 1) / float64(ecnt)
	if float64(f.Entries()) * ecntM >= float64(f.MaxEntries()) {
		f.nrEntries = f.nrEntries[:ecnt-1]
		f.b = f.b[:ecnt-1]
	}
}

func (f *Filter) hash(b []byte) uint64 {
	h, _ := siphash.Hash128(f.k1, f.k2, b)
	h &= f.mask
	return h
}

func (f *Filter) test(hash uint64) bool {
	for i := 0; i < len(f.b); i++ {
		if 0 != f.b[i][hash/8]&(1<<(hash&7)) {
			return true
		}
	}
	return false
}

func (f *Filter) add(hash uint64) {
	f.b[0][hash/8] |= (1 << (hash & 7))
}
