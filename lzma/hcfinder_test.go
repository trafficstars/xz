package lzma

import (
	"testing"
)

func TestHTable(t *testing.T) {
	dict, err := newDict(4096, 256)
	if err != nil {
		t.Fatalf("newDict error %s", err)
	}
	ht := newHTable(2<<16, dict)
	dict.Write([]byte("ballack"))
	dict.Discard(2)
	ht.put(0, 10)
	ht.put(12, 11)
	check := func(key uint32, off int64) {
		o, ok := ht.get(key)
		if !ok {
			t.Fatalf("get(%d) returns no value; want %d", key, off)
		}
		if o != off {
			t.Fatalf("get(%d) returns %d; want %d", key, o, off)
		}
	}
	check(0, 10)
	check(12, 11)
	ht.rebase()
	t.Log("rebased")
	check(0, 10)
	check(12, 11)
}

/*
func TestHChain(t *testing.T) {
	dict, err := newDict(4096, 256)
	if err != nil {
		t.Fatalf("newDict error %s", err)
	}
	hc := newHChain(dict, 4, 256)
	dict.Write([]byte("ballack"))
	dict.Discard(2)
	p := make([]byte, 4)
	var (
		h0, h1 uint32
		ok     bool
	)
	if h0, ok = hashAt(dict, p, 0); !ok {
		t.Fatalf("hashAt index %d failed", 0)
	}
	if h1, ok = hashAt(dict, p, 1); !ok {
		t.Fatalf("hashAt index %d failed", 1)
	}
	hc.put(h0, pointer(0))
	hc.put(h1, pointer(1))
	hc.resize(512)
	ptrs := make([]ptr, 4)
	n := hc.get(h0, ptrs)
	if n != 1 {
		t.Fatalf("hc.get(%#08x, p) returned %d; want %d", h0, n, 1)
	}
	if ptrs[0] != pointer(0) {
		t.Fatalf("ptrs[0] is %d; want %d", ptrs[0], pointer(0))
	}
}
*/

const example = `LZMA decoder test example
=========================
! LZMA ! Decoder ! TEST !
=========================
! TEST ! LZMA ! Decoder !
=========================
---- Test Line 1 --------
=========================
---- Test Line 2 --------
=========================
=== End of test file ====
=========================
`

/*
func TestHCFinder(t *testing.T) {
	const depth = 20
	f, err := newHCFinder(4, 4096, 4096, 200, depth)
	if err != nil {
		t.Fatalf("newHCFinder error %s", err)
	}
	dict := f.Dict()
	n, err := io.WriteString(dict, example)
	if err != nil {
		t.Fatalf("io.WriteString(dict, example) error %s", err)
	}
	if n != len(example) {
		t.Fatalf("io.WriteString(dict, example) returned %d; want %d",
			n, len(example))
	}
	matches := make([]operation, depth)
	t.Log("# Example")
	t.Log(example)
	for dict.Buffered() > 0 {
		pos := dict.Pos()
		end := pos + 10
		if end > int64(len(example)) {
			end = int64(len(example))
		}
		t.Logf("pos %d %q...\n", pos, example[pos:end])
		n := f.FindMatches(matches)
		if n == 0 {
			t.Log("no matches")
		}
		for i, m := range matches[:n] {
			t.Logf("matches[%d] = %v -> %q", i, m,
				example[pos:pos+int64(m.len)])
		}
		if n == 0 {
			f.Skip(1)
			continue
		}
		best := matches[n-1]
		f.Skip(int(best.len))
	}
	for i, ht := range f.ht {
		t.Logf("len(f.ht[%d].table) %d", i, len(ht.table))
	}
	t.Logf("len(f.hc.table) %d", len(f.hc.table))
}
*/
