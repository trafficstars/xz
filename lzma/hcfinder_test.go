package lzma

import "testing"

func TestHTable(t *testing.T) {
	var ht htable
	for u := uint32(0); u < 512; u++ {
		ht.put(u, pointer(int(u)))
	}
	if len(ht.table) != 1024 {
		t.Fatalf("len(ht.table) is %d; want %d", len(ht.table), 1024)
	}
	for u := uint32(0); u < 512; u++ {
		p, ok := ht.get(u)
		if !ok {
			t.Fatalf("get(%d) can't find a value", u)
		}
		w := pointer(int(u))
		if p != w {
			t.Fatalf("get(%d) returned %d; want %d", u, p, w)
		}
	}
}

func TestHChain(t *testing.T) {
	hc := newHChain(2, 256)
	hc.put(0, pointer(0))
	hc.put(0, pointer(1))
	hc.resize(512)
	ptrs := make([]ptr, 4)
	n := hc.get(0, ptrs)
	if n != 2 {
		t.Fatalf("hc.put(0, p) returned %d; want %d", n, 2)
	}
	ptrs = ptrs[0:n]
	t.Log(ptrs)
	for _, p := range ptrs {
		if !(0 <= p.index() && p.index() <= 1) {
			t.Fatalf("p.index() is %d; want 0 or 1", p.index())
		}
	}
}
