package lzma

import (
	"github.com/ulikunitz/xz/internal/buz"
)

const (
	maxUint32   = 1<<32 - 1
	maxTableLen = 1 << 26
)

const load float64 = 0.75

type ptr uint32

// pointer converts an index into a pointer. The pointer returned is
// never null.
func pointer(idx int) ptr {
	if !(0 <= idx && int64(idx) < maxUint32) {
		panic("index out of range")
	}
	return ptr(idx + 1)
}

func (p ptr) index() int {
	if p == 0 {
		panic("no index for null pointer")
	}
	return int(p - 1)
}

type htEntry struct {
	key uint32
	p   ptr
}

type htable struct {
	table   []htEntry
	entries int
	max     int
	mask    uint32
}

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

func (h *htable) get(key uint32) (p ptr, ok bool) {
	p = h.table[key&h.mask].p
	return p, p != 0
}

type hcEntry struct {
	key  uint32
	next ptr
}

type hchain struct {
	table   []ptr
	chain   []hcEntry
	entries int
	max     int
	mask    uint32
}

func newHChain(dictCap int, tableLen int) *hchain {
	if tableLen < 4 {
		tableLen = 256
	}
	if !(1 <= tableLen && tableLen <= maxTableLen) {
		panic("tableLen out of range")
	}
	n := int(clp2u32(uint32(tableLen)))
	return &hchain{
		chain: make([]hcEntry, dictCap),
		table: make([]ptr, n),
		mask:  uint32(n - 1),
		max:   int(load * float64(n)),
	}
}

func (h *hchain) resize(n int) {
	if n < 4 {
		n = 256
	}
	if !(1 <= n && n <= maxTableLen) {
		panic("n out of range")
	}
	n = int(clp2u32(uint32(n)))
	h.mask = uint32(n - 1)
	h.max = int(load * float64(n))
	h.entries = 0
	t := h.table
	h.table = make([]ptr, n)
	for _, p := range t {
		for p != 0 {
			e := &h.chain[p.index()]
			pptr := &h.table[e.key&h.mask]
			next := *pptr
			if next == 0 {
				h.entries++
			}
			*pptr = p
			p = e.next
			e.next = next
		}
	}
}

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
	e := &h.chain[p.index()]
	// remove old table entry
	pptr := &h.table[e.key&h.mask]
	if *pptr == p {
		if e.next == 0 {
			h.entries--
		}
		*pptr = e.next
	}
	// add p to table
	pptr = &h.table[key&h.mask]
	next := *pptr
	if next == 0 {
		h.entries++
	}
	*pptr = p
	*e = hcEntry{key, next}
}

func (h *hchain) get(key uint32, ptrs []ptr) int {
	n := 0
	p := h.table[key&h.mask]
	for {
		if p == 0 {
			return n
		}
		ptrs[n] = p
		n++
		if n >= len(ptrs) {
			return n
		}
		p = h.chain[p.index()].next
	}
}

type hcFinder struct {
	// word length for the hasd chain
	n int
	// dictionary providing Write, Pos, ByteAt and CopyN
	dict *dict
	// slice with hash values
	hash buz.Hash
	// number of bytes that the hasher is in front of pos();
	// supports hashing in FindMatches
	ahead int
	// hash tables for non-maximum hash sizes; starting with word
	// length 2
	ht []htable
	// table of hash entries
	hc *hchain
}

func newHCFinder(dictCap int, bufSize int, niceLen int, depth int,
) (hc *hcFinder, err error) {
	panic("TODO")
}
