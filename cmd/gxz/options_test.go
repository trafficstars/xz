package main

import "testing"

func TestOptionSet_Parse(t *testing.T) {
	var os optionSet
	a := os.Int64("a", 1)
	b := os.Int64("b", 2)
	if *a != 1 {
		t.Fatalf("*a is %d; want %d", *a, 1)
	}
	if *b != 2 {
		t.Fatalf("*b is %d; want %d", *b, 2)
	}
	if os.Parsed() {
		t.Fatalf("os.Parsed() is %t before parsing; want %t",
			os.Parsed(), false)
	}
	if err := os.Parse("a=3"); err != nil {
		t.Fatalf("os.Parse(%q) error %s", "a=3", err)
	}
	if !os.Parsed() {
		t.Fatalf("os.Parsed() is %t after parsing; want %t",
			os.Parsed(), true)
	}
	if *a != 3 {
		t.Fatalf("*a is %d; want %d", *a, 3)
	}
	if *b != 2 {
		t.Fatalf("*b is %d; want %d", *b, 2)
	}

	os.Reset()
	if os.Parsed() {
		t.Fatalf("os.Parsed() is %t before parsing; want %t",
			os.Parsed(), false)
	}
	if err := os.Parse("a=5,b=4"); err != nil {
		t.Fatalf("os.Parse(%q) error %s", "a=3", err)
	}
	if *a != 5 {
		t.Fatalf("*a is %d; want %d", *a, 5)
	}
	if *b != 4 {
		t.Fatalf("*b is %d; want %d", *b, 4)
	}
}

func TestOptionSet_Int64Ext(t *testing.T) {
	var os optionSet
	a := os.Int64Ext("a", 4*eKiB)
	if *a != 4*eKiB {
		t.Fatalf("*a is %d; want %d", *a, 4*eKiB)
	}
	if err := os.Parse("a=2MiB"); err != nil {
		t.Fatalf("os.Parse(%q) error %s", "a=2MiB", err)
	}
	if *a != 2*eMiB {
		t.Fatalf("*a is %d; want %d", *a, 2*eMiB)
	}
}
