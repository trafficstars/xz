package buz

import "testing"

func TestHash(t *testing.T) {
	data := []byte("balla")
	h4 := MakeHash(4)
	n := h4.Compute(data[:4])
	if n != 3 {
		t.Fatalf("h4.Compute returned %d; want %d", n, 3)
	}
	h5 := MakeHash(5)
	m := h5.Compute(data)
	if m != 4 {
		t.Fatalf("h5.Compute returned %d; want %d", n, 4)
	}
	for i := 0; i < n; i++ {
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
