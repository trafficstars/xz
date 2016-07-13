package lzma

import "testing"

func TestClp2u32(t *testing.T) {
	tests := []struct{ x, y uint32 }{
		{0, 0},
		{1 << 0, 1 << 0},
		{3, 4},
		{1<<10 + 32, 1 << 11},
		{1<<29 + 512, 1 << 30},
		{1 << 30, 1 << 30},
	}
	for _, c := range tests {
		y := clp2u32(c.x)
		if y != c.y {
			t.Errorf("clp2u32(%d) is %d; want %d", c.x, y, c.y)
		}
	}
}
