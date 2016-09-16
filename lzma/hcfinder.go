package lzma

import (
	"errors"
	"fmt"

	"github.com/ulikunitz/xz/internal/buz"
)

// Base represents the offset base for the uint32 truncated offset
// values stored in the hash tables. Note that the truncated value 0
// is interpreted a no offset instead of offset zero.
type base struct{ o int64 }

// off returns the full offset for the truncated offset.
func (b base) off(u uint32) int64 { return b.o + int64(u) }

// short converts an offset into a truncacted value.
func (b base) short(off int64) uint32 {
	n := off - b.o
	if !(0 < n && n < 1<<32) {
		panic("offset out of range")
	}
	return uint32(n)
}

// htable provides a simple hash table. Old values for the same slot
// will be forgotten.
type htable struct {
	dict  *dict
	table []uint32
	mask  uint32
	base  base
}

// newHTable creates new htable instance.
func newHTable(size int, d *dict) *htable {
	if size > d.capacity {
		size = d.capacity
	}
	size = int(clp2u32(uint32(size)))
	return &htable{
		dict: d,
		// hash table of truncated offset values; zero entries
		// mean no offset stored
		table: make([]uint32, size),
		mask:  uint32(size - 1),
		// base -1 makes truncated offset 1 offset 0
		base: base{-1},
	}
}

// rebase shifts the base. All non-zero truncated offsets in the hash
// table will be recomputed. For extremely large dictionary capacities
// the shift will be smaller than the dictionary capacity to limit the
// number of rebases required.
func (h *htable) rebase() {
	// We need to limit the maximum dictionary length to limit the
	// number of rebase steps.
	const max = 3 << 30
	c := int64(h.dict.dictLen())
	if c > max {
		c = max
	}
	b := h.dict.head - c - 1
	delta64 := b - h.base.o
	switch {
	case delta64 < 0:
		panic("decreasing base unexpected")
	case delta64 == 0:
		return
	case delta64 >= 1<<32:
		panic("delta too large")
	}
	delta := uint32(delta64)
	h.base.o = b
	for i, u := range h.table {
		if u == 0 {
			continue
		}
		h.table[i] = dozu32(u, delta)
	}
}

// put makes an entry in the hash table. Old entries will be
// overwritten. If necessary the truncated offsets will be rebased.
func (h *htable) put(key uint32, off int64) {
	// check for rebase
	if h.base.o+(1<<32) <= off {
		h.rebase()
	}
	h.table[key&h.mask] = h.base.short(off)
}

// get returns the offset for the given key.
func (h *htable) get(key uint32) (off int64, ok bool) {
	u := h.table[key&h.mask]
	if u == 0 {
		return 0, false
	}
	return h.base.off(u), true
}

// hchain provides a hash table with chained hashing. The table might be
// resized during usage.
type hchain struct {
	htable
	chain []uint32
}

// newHChain creates a new instance for a hash chain.
func newHChain(size int, d *dict) *hchain {
	return &hchain{
		htable: *newHTable(size, d),
		chain:  make([]uint32, len(d.buf.data)),
	}
}

func (h *hchain) put(key uint32, off int64) {
	// TODO: are we putting here the head all the time
	i, ok := h.dict.indexOff(off)
	if !ok {
		panic("offset not inside dictionary")
	}
	if h.base.o+(1<<32) <= off {
		h.rebase()
	}
	p := &h.table[key&h.mask]
	u := *p
	*p = h.base.short(off)
	p = &h.chain[i]
	if u == 0 {
		*p = 0
		return
	}
	delta := off - h.base.off(u)
	if delta < 0 {
		panic("negative delta")
	}
	if delta >= int64(h.dict.capacity) {
		*p = 0
		return
	}
	*p = uint32(delta)
}

// get returns the distances found for the key. It is not checked of
// distances actually matcht the key.
func (h *hchain) get(key uint32, distances []int) int {
	u := h.table[key&h.mask]
	if u == 0 {
		return 0
	}
	off := h.base.off(u)
	i, ok := h.dict.indexOff(off)
	if !ok {
		return 0
	}
	dictLen := int64(h.dict.dictLen())
	dist := h.dict.head - off
	if !(0 <= dist && dist <= dictLen) {
		return 0
	}
	dataLen := len(h.dict.buf.data)
	n := 0
	for {
		distances[n] = int(dist)
		n++
		if n >= len(distances) {
			return n
		}
		delta := h.chain[i]
		if delta == 0 {
			return n
		}
		dist += int64(delta)
		if dist > dictLen {
			h.chain[i] = 0
			return n
		}
		i -= int(delta)
		if i < 0 {
			i += dataLen
		}
	}
}

// hcFinder provides find matcher that is based on hash tables using
// simple chaining.
type hcFinder struct {
	// dictionary providing Write, Pos, ByteAt and CopyN
	dict *dict
	// slice with hash values
	hash buz.Hash
	// number of bytes that the finder is in front of dict.pos();
	// supports hashing in FindMatches
	ahead int
	// hash tables for non-maximum hash sizes; starting with word
	// length 2
	ht []*htable
	// table of hash entries
	hc *hchain
	// word length for the hash chain
	n int
	// preallocated array; capacity is depth
	distances []int
	// preallocated data having the capacity of niceLen
	data []byte
}

// newHCFinder creates a new instanse of a hash chain finder. n is the
// word length by the largest hash table.
func newHCFinder(n int, dictCap int, bufSize int, niceLen int, depth int,
) (hc *hcFinder, err error) {
	if n < 2 {
		return nil, errors.New(
			"lzma: word length n must be greater or equal 2")
	}
	if !(2 <= niceLen && niceLen <= maxMatchLen) {
		return nil, fmt.Errorf(
			"lzma: niceLen must be in [%d,%d]", 2, maxMatchLen)
	}
	if depth < 1 {
		return nil, errors.New("lzma: depth must be larger or equal 1")
	}
	dict, err := newDict(dictCap, bufSize)
	if err != nil {
		return nil, err
	}
	f := &hcFinder{
		n:         n,
		dict:      dict,
		hash:      buz.MakeHash(n),
		ht:        make([]*htable, n-2),
		hc:        newHChain(1<<27, dict),
		distances: make([]int, depth),
		data:      make([]byte, niceLen),
	}
	for i := range f.ht {
		var size int
		switch i + 2 {
		case 2:
			size = 1 << 16
		case 3:
			size = 1 << 24
		default:
			size = 1 << 26
		}
		f.ht[i] = newHTable(size, dict)
	}
	return f, nil
}

// Dict returns the dictionary.
func (f *hcFinder) Dict() *dict {
	return f.dict
}

// Depth returns the depth applied by the hash finder.
func (f *hcFinder) Depth() int {
	return cap(f.distances)
}

// mustDiscard forwards the head. All issue will be noted as panics. For
// purists: These panics are not supposed to happen and are bugs.
func (f *hcFinder) mustDiscard(n int) {
	discarded, err := f.dict.Discard(n)
	if err != nil {
		panic(err)
	}
	if discarded != n {
		panic(fmt.Errorf("method Discard returned %d; want %d",
			discarded, n))
	}
}

// enters the current set of hash values into the hash tables.
func (f *hcFinder) enterHashes() {
	for i, h := range f.hash {
		off := f.dict.head
		if i < f.n-2 {
			f.ht[i].put(h, off)
		} else {
			f.hc.put(h, off)
		}
	}
}

// Skip skips over n byte of input. Bytes are added to the dictionary
// and hashes are still computed.
func (f *hcFinder) Skip(n int) {
	if n < 0 {
		panic("lzma: can't skip negative number of bytes")
	}
	if n > f.dict.Buffered() {
		panic(errors.New("lzma: skip request exceeds buffered size"))
	}
	if n <= f.ahead {
		f.mustDiscard(n)
		f.ahead -= n
		return
	}
	n -= f.ahead
	f.mustDiscard(f.ahead)
	f.ahead = 0
	for ; n > 0; n-- {
		k, _ := f.dict.Peek(f.data[:f.n])
		f.hash.Compute(f.data[:k])
		f.enterHashes()
		f.mustDiscard(1)
	}
}

// betterMatch checks whether there is a match that is greater than
// maxLen on the given distance.
func (f *hcFinder) betterMatch(dist int, maxLen int) (m operation, ok bool) {
	if len(f.data) < maxLen+1 {
		panic("called with len(f.data) < maxLen+1")
	}
	buf := &f.dict.buf
	if dist < 0 {
		dist += len(buf.data)
	}
	if !(1 <= dist && dist <= f.dict.dictLen()) {
		return m, false
	}
	if f.dict.ByteAt(dist-maxLen) != f.data[maxLen] {
		return m, false
	}
	n := f.dict.matchLen(dist, f.data)
	if !(2 <= n && n <= maxMatchLen) {
		return m, false
	}
	return match(uint32(dist), n), true
}

// onlyFindMatches searches only for matches and doesn't extend the
// dictionary.
func (f *hcFinder) onlyfindMatches(m []operation) int {
	if len(m) == 0 {
		return 0
	}
	if len(f.data) == 0 {
		panic("no data")
	}
	var n, maxLen int
	hi := len(f.hash) - 1
	if hi+2 == f.n {
		// search hash chain
		k := f.hc.get(f.hash[hi], f.distances)
		hi--
		for _, dist := range f.distances[:k] {
			bm, ok := f.betterMatch(dist, maxLen)
			if !ok {
				continue
			}
			maxLen = int(bm.len)
			m[n] = bm
			n++
			if n == len(m) || len(f.data) == maxLen {
				return n
			}
		}
	}
	for ; hi >= 0; hi-- {
		// search the hash tables for the smaller hash lengths
		ht := f.ht[hi]
		off, ok := ht.get(f.hash[hi])
		if !ok {
			continue
		}
		bm, ok := f.betterMatch(int(f.dict.head-off), maxLen)
		if !ok {
			continue
		}
		maxLen = int(bm.len)
		m[n] = bm
		n++
		if n == len(m) || len(f.data) == maxLen {
			return n
		}
	}
	return n
}

// FindMatches finds for the current position and returns them in the
// given array. The number of matches entered into m are returned. The
// matches in the slice will have increasing lengths. The dictionary
// will be moved by one byte and hashes for the head position will be
// entered into the dictionary.
func (f *hcFinder) FindMatches(m []operation) int {
	if f.dict.Buffered() == 0 {
		panic(errors.New("lzma: no data buffered"))
	}
	k, _ := f.dict.Peek(f.data[:cap(f.data)])
	f.data = f.data[:k]
	if f.n < k {
		k = f.n
	}
	f.hash.Compute(f.data[:k])
	n := f.onlyfindMatches(m)
	if f.ahead == 0 {
		f.enterHashes()
		f.ahead++
	}
	return n
}
