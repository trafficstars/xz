package lzma

import (
	"errors"
	"fmt"

	"github.com/ulikunitz/xz/internal/buz"
)

// Maximum value. The maximum table len results in a 512 MiB table field
// for type htable and a 256 MiB table for type hchain.
const (
	maxUint32   = 1<<32 - 1
	maxTableLen = 1 << 26
)

// load provides the limit for the load of a hash table. Is the load
// higher the table size will be doubled.
const load float64 = 0.75

// ptr is a reference to the actual dictionary. Since we need a zero
// value for an uninitialized pointer, we increase index values by 1.
type ptr uint32

// pointer converts an index into a pointer. The pointer returned is
// never null.
func pointer(idx int) ptr {
	if !(0 <= idx && int64(idx) < maxUint32) {
		panic("index out of range")
	}
	return ptr(idx + 1)
}

// index returns the index value of the pointer. The function panics if
// called for a zero value pointer.
func (p ptr) index() int {
	if p == 0 {
		panic("no index for null pointer")
	}
	return int(p - 1)
}

// htEntry describes an entry into the hash table. The key is stored to
// support resizing.
type htEntry struct {
	key uint32
	p   ptr
}

// htable provides a simple hash table. Old values for the same slot
// will be forgotten.
type htable struct {
	table   []htEntry
	entries int
	max     int
	mask    uint32
}

// resize changes the size of the table. While it is possible to
// decrease the size of an hash table, expect the hash table to lose
// entries and a resize during the next put call.
func (h *htable) resize(n int) {
	if n == 0 {
		*h = htable{}
		return
	}
	if !(1 <= n && n <= maxTableLen) {
		panic("n out of range")
	}
	n = int(clp2u32(uint32(n)))
	t := h.table
	h.table = make([]htEntry, n)
	h.mask = uint32(n - 1)
	h.max = int(load * float64(n))
	h.entries = 0
	for _, e := range t {
		if e.p != 0 {
			pe := &h.table[e.key&h.mask]
			if pe.p != 0 {
				panic("unexpected double entry")
			}
			h.entries++
			*pe = e
		}
	}
}

// put makes an entry in the hash table. The table might be resized by the
// function. The old entry in the slot might be overwritten.
func (h *htable) put(key uint32, p ptr) {
	if p == 0 {
		panic("null pointer not supported")
	}
	if h.entries >= h.max && len(h.table) <= maxTableLen>>1 {
		var n int
		if len(h.table) == 0 {
			n = 256
		} else {
			n = 2 * len(h.table)
		}
		h.resize(n)
	}
	e := &h.table[key&h.mask]
	if e.p == 0 {
		h.entries++
	}
	*e = htEntry{key: key, p: p}
}

// get returns the pointer for key if available.
func (h *htable) get(key uint32) (p ptr, ok bool) {
	if len(h.table) == 0 {
		return 0, false
	}
	p = h.table[key&h.mask].p
	return p, p != 0
}

// hcEntry describes an entry in the chain field of the hchain type.
type hcEntry struct {
	key  uint32
	prev ptr
	next ptr
}

// hchain provides a hash table with chained hashing. The table might be
// resized during usage.
type hchain struct {
	table   []ptr
	chain   []hcEntry
	entries int
	max     int
	mask    uint32
}

// newHChain creates a new instance for a hash chain. The dictSize
// describes the complete size, dictCap + bufSize,  of the dictionary to
// be able to address all bytes of it.
func newHChain(dictSize int, tableLen int) *hchain {
	if tableLen < 4 {
		tableLen = 256
	}
	if !(1 <= tableLen && tableLen <= maxTableLen) {
		panic("tableLen out of range")
	}
	n := int(clp2u32(uint32(tableLen)))
	return &hchain{
		chain: make([]hcEntry, dictSize),
		table: make([]ptr, n),
		mask:  uint32(n - 1),
		max:   int(load * float64(n)),
	}
}

// resize changes the size of the hash chain.
func (h *hchain) resize(n int) {
	if n < 4 {
		n = 256
	}
	if !(1 <= n && n <= maxTableLen) {
		panic("n out of range")
	}
	n = int(clp2u32(uint32(n)))
	t := h.table
	h.table = make([]ptr, n)
	h.mask = uint32(n - 1)
	h.max = int(load * float64(n))
	h.entries = 0
	for _, head := range t {
		if head == 0 {
			continue
		}
		e := h.chain[head.index()]
		p := e.prev
		tail := p
		for {
			e = h.chain[p.index()]
			prev := e.prev
			h._put(e.key, p)
			p = prev
			if p == tail {
				break
			}
		}
	}
}

// removes pointer from hash chain
func (h *hchain) remove(p ptr) {
	if p == 0 {
		return
	}
	e := &h.chain[p.index()]
	if e.next == 0 {
		return
	}
	// handle head entry
	head := &h.table[e.key&h.mask]
	if p == *head {
		if p != e.next {
			*head = e.next
		} else {
			h.entries--
			*head = 0
		}
	}
	if p != e.next {
		eprev := &h.chain[e.prev.index()]
		enext := &h.chain[e.next.index()]
		eprev.next = e.next
		enext.prev = e.prev
	}
	*e = hcEntry{}
}

func (h *hchain) _put(key uint32, p ptr) {
	if p == 0 {
		panic(errors.New("_put: null pointer argument"))
	}
	head := &h.table[key&h.mask]
	e := &h.chain[p.index()]
	e.key = key
	if *head == 0 {
		e.prev, e.next = p, p
		*head = p
	} else {
		e.next = *head
		enext := &h.chain[head.index()]
		*head = p
		e.prev = enext.prev
		eprev := &h.chain[e.prev.index()]
		enext.prev = p
		eprev.next = p
	}
}

// put puts aan entry into the hash chain. The old chain will still
// reference it, but the get function will take care of it. It assumes
// that all entries after p in the chain are not relevant anymore.
func (h *hchain) put(key uint32, p ptr) {
	if p == 0 {
		panic("null pointer not supported")
	}
	if h.entries >= h.max && len(h.table) <= maxTableLen>>1 {
		// resize table
		var n int
		if len(h.table) == 0 {
			n = 256
		} else {
			n = 2 * len(h.table)
		}
		h.resize(n)
	}
	h.remove(p)
	h._put(key, p)
}

// get returns the the pointers found for the key.
func (h *hchain) get(key uint32, ptrs []ptr) int {
	head := h.table[key&h.mask]
	if head == 0 {
		return 0
	}
	n := 0
	p := head
	for {
		e := h.chain[p.index()]
		if e.key == key {
			ptrs[n] = p
			n++
			if n >= len(ptrs) {
				return n
			}
		}
		p = h.chain[p.index()].next
		if p == head {
			return n
		}
	}
}

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
	ht []htable
	// table of hash entries
	hc *hchain
	// word length for the hasd chain
	n int
	// nice length
	niceLen int
	// segment for which a match needs to be found
	seg segment
	// preallocated array; capacity is depth
	ptrs []ptr
	// preallocated word to hash
	word []byte
}

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
		n:       n,
		niceLen: niceLen,
		dict:    dict,
		hash:    buz.MakeHash(n),
		ht:      make([]htable, n-2),
		hc:      newHChain(dict.buf.Cap()+1, 1<<18),
		ptrs:    make([]ptr, depth),
		word:    make([]byte, n),
	}
	for i := range f.ht {
		f.ht[i].resize(1 << (10 + 2*uint(i)))
	}
	return f, nil
}

func (f *hcFinder) Dict() *dict {
	return f.dict
}

func (f *hcFinder) Depth() int {
	return cap(f.ptrs)
}

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

func (f *hcFinder) enterHashes() {
	for i, h := range f.hash {
		p := pointer(f.dict.buf.rear)
		if i < f.n-2 {
			f.ht[i].put(h, p)
		} else {
			f.hc.put(h, p)
		}
	}
}

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
		k, _ := f.dict.Peek(f.word)
		f.hash.Compute(f.word[:k])
		f.enterHashes()
		f.mustDiscard(1)
	}
}

func (f *hcFinder) betterMatch(p ptr, maxLen int) (m operation, ok bool) {
	if f.seg.Len() < maxLen+1 {
		panic("called with seg.Len() < maxLen+1")
	}
	buf := &f.dict.buf
	dist := buf.rear - p.index()
	if dist < 0 {
		dist += len(buf.data)
	}
	if !(1 <= dist && dist <= f.dict.dictLen()) {
		return m, false
	}
	if f.dict.ByteAt(dist-maxLen) != f.dict.ByteAt(-maxLen) {
		return m, false
	}
	pseg, err := f.dict.segment(dist, f.seg.Len())
	if err != nil {
		panic(err)
	}
	n := prefixLenSegments(pseg, f.seg)
	if !(2 <= n && n <= maxMatchLen) {
		return m, false
	}
	return match(uint32(dist), n), true
}

func (f *hcFinder) onlyfindMatches(m []operation) int {
	if len(m) == 0 {
		return 0
	}
	if f.seg.Len() == 0 {
		panic("no data")
	}
	var n, maxLen int
	hi := len(f.hash) - 1
	if hi+2 == f.n {
		// search hash chain
		k := f.hc.get(f.hash[hi], f.ptrs)
		hi--
		for _, p := range f.ptrs[:k] {
			bm, ok := f.betterMatch(p, maxLen)
			if !ok {
				continue
			}
			maxLen = int(bm.len)
			m[n] = bm
			n++
			if n == len(m) || f.seg.Len() == maxLen {
				return n
			}
		}
	}
	for ; hi >= 0; hi-- {
		// search the hash tables for the smaller hash lengths
		ht := f.ht[hi]
		p, ok := ht.get(f.hash[hi])
		if !ok {
			continue
		}
		bm, ok := f.betterMatch(p, maxLen)
		if !ok {
			continue
		}
		maxLen = int(bm.len)
		m[n] = bm
		n++
		if n == len(m) || f.seg.Len() == maxLen {
			return n
		}
	}
	return n
}

func (f *hcFinder) FindMatches(m []operation) int {
	if f.dict.Buffered() == 0 {
		panic(errors.New("lzma: no data buffered"))
	}
	var err error
	f.seg, err = f.dict.segment(0, f.niceLen)
	if err != nil {
		panic(err)
	}
	k, _ := f.seg.Peek(f.word)
	f.hash.Compute(f.word[:k])
	n := f.onlyfindMatches(m)
	if f.ahead == 0 {
		f.enterHashes()
		f.ahead++
	}
	return n
}
