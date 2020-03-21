// Package cuckoofilter implements cuckoo filters from the paper
// "Cuckoo Filter: Practically Better Than Bloom" by Fan et al.
// https://www.cs.cmu.edu/~dga/papers/cuckoo-conext2014.pdf
package cuckoofilter

import (
	"encoding/binary"
	"errors"
	"math/rand"

	"github.com/zhenjl/cityhash"
)

// 4 entries per bucket is suggested by the paper in section 5.1,
// "Optimal bucket size"
const entriesPerBucket = 4

// With 4 entries per bucket, we can expect up to 95% load factor
const loadFactor = 0.95

// Length of fingerprints in bits
const fpBits = 16

// Arbitrarily chosen value
const maxDisplacements = 500

// ErrTooFull is returned when a filter is too full and needs to be
// resized.
var ErrTooFull = errors.New("cuckoo filter too full")

// Filter is an implementation of a cuckoo filter.
type Filter struct {
	nBuckets uint32
	table    [][entriesPerBucket]uint16
}

func nearestPowerOfTwo(val uint32) uint32 {
	for i := uint32(0); i < 32; i++ {
		if pow := uint32(1) << i; pow >= val {
			return pow
		}
	}

	panic("will never happen")
}

// New returns a new cuckoo filter sized for the maximum number of
// keys passed in as maxKeys.
func New(maxKeys uint32) *Filter {
	nBuckets := nearestPowerOfTwo(maxKeys / entriesPerBucket)

	// If load factor is above the max value, we'll likely hit the
	// max number of fingerprint displacements.  In that case,
	// expand the number of buckets.
	if float64(maxKeys)/float64(nBuckets)/entriesPerBucket > loadFactor {
		nBuckets <<= 1
	}

	f := &Filter{
		nBuckets: nBuckets,
		table:    make([][entriesPerBucket]uint16, nBuckets),
	}

	return f
}

func hash(data []byte) uint64 {
	return cityhash.CityHash64(data, uint32(len(data)))
}

func (f *Filter) bucketIndex(hv uint32) uint32 {
	return hv % f.nBuckets
}

func (f *Filter) fingerprint(hv uint32) uint16 {
	fp := uint16(hv & ((1 << fpBits) - 1))

	// gross
	if fp == 0 {
		fp = 1
	}

	return fp
}

func (f *Filter) alternateIndex(idx uint32, fp uint16) uint32 {
	d := make([]byte, 2)
	binary.LittleEndian.PutUint16(d, fp)
	hv := hash(d)
	return f.bucketIndex(idx ^ uint32(hv))
}

func (f *Filter) matchPosition(idx uint32, fp uint16) int {
	for i := 0; i < entriesPerBucket; i++ {
		if f.table[idx][i] == fp {
			return i
		}
	}

	return -1
}

func (f *Filter) emptyPosition(idx uint32) int {
	return f.matchPosition(idx, 0)
}

// Add adds an element to the cuckoo filter.  If the filter is too
// heavily loaded, ErrTooFull may be returned, which signifies that
// the filter must be rebuilt with an increased maxKeys parameter.
func (f *Filter) Add(d []byte) error {
	h := hash(d)

	fp := f.fingerprint(uint32(h))
	i1 := f.bucketIndex(uint32(h >> 32))
	i2 := f.alternateIndex(i1, fp)

	if i := f.emptyPosition(i1); i != -1 {
		f.table[i1][i] = fp
		return nil
	}

	if i := f.emptyPosition(i2); i != -1 {
		f.table[i2][i] = fp
		return nil
	}

	// Choose which index to use randomly
	idx := [2]uint32{i1, i2}[rand.Intn(2)]

	for i := 0; i < maxDisplacements; i++ {
		j := uint32(rand.Intn(entriesPerBucket))

		fp, f.table[idx][j] = f.table[idx][j], fp
		idx = f.alternateIndex(idx, fp)

		if ni := f.emptyPosition(idx); ni != -1 {
			f.table[idx][ni] = fp
			return nil
		}
	}

	return ErrTooFull
}

// Contains returns whether an element may be present in the set.
// Cuckoo filters are probablistic data structures which can return
// false positives.  False negatives are not possible.
func (f *Filter) Contains(d []byte) bool {
	h := hash(d)

	fp := f.fingerprint(uint32(h))
	i1 := f.bucketIndex(uint32(h >> 32))
	i2 := f.alternateIndex(i1, fp)

	return f.matchPosition(i1, fp) != -1 || f.matchPosition(i2, fp) != -1
}

// Delete deletes an element from the set.  To delete an item safely,
// it must have been previously inserted.  Deleting a non-inserted
// item might unintentionally remove a real, different item.
func (f *Filter) Delete(d []byte) bool {
	h := hash(d)

	fp := f.fingerprint(uint32(h))
	i1 := f.bucketIndex(uint32(h >> 32))
	i2 := f.alternateIndex(i1, fp)

	if i := f.matchPosition(i1, fp); i != -1 {
		f.table[i1][i] = 0
		return true
	}

	if i := f.matchPosition(i2, fp); i != -1 {
		f.table[i2][i] = 0
		return true
	}

	return false
}
