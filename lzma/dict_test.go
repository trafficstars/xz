package lzma

import "testing"

func TestPrefixLen(t *testing.T) {
	tests := []struct {
		a, b []byte
		k    int
	}{
		{[]byte("abcde"), []byte("abc"), 3},
		{[]byte("abc"), []byte("uvw"), 0},
		{[]byte(""), []byte("uvw"), 0},
		{[]byte("abcde"), []byte("abcuvw"), 3},
	}
	for _, c := range tests {
		k := prefixLen(c.a, c.b)
		if k != c.k {
			t.Errorf("prefixLen(%q,%q) returned %d; want %d",
				c.a, c.b, k, c.k)
		}
		k = prefixLen(c.b, c.a)
		if k != c.k {
			t.Errorf("prefixLen(%q,%q) returned %d; want %d",
				c.b, c.a, k, c.k)
		}
	}
}
