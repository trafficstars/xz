package buz

import "testing"

func TestHash(t *testing.T) {
	data := []byte("balla")
	h4 := MakeHash(4)
	h4.Compute(data[:4])
	if len(h4) != 3 {
		t.Fatalf("len(h4) is %d; want %d", len(h4), 3)
	}
	h5 := MakeHash(5)
	h5.Compute(data)
	if len(h5) != 4 {
		t.Fatalf("len(h5) is %d; want %d", len(h5), 4)
	}
	for i := 0; i < len(h4); i++ {
		if h4[i] != h5[i] {
			t.Fatalf("h4[%d] != h5[%d]", i, i)
		}
	}
	if h5[3] == h5[2] {
		t.Fatalf("h5[%d] == h5[%d]", 3, 2)
	}
	for i, h := range h5 {
		t.Logf("h5[%d] %#08x", i, h)
	}
}

func TestSingleHash(t *testing.T) {
	s := "ball"
	data := []byte(s)
	h4 := MakeHash(4)
	h4.Compute(data)
	w := h4[2]
	h := SingleHash(data)
	if h != w {
		t.Fatalf("SingleHash([]byte(%q)) is %#08x; want %#08x", s, h, w)
	}
}
