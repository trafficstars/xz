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
	a := os.Int64Ext("a", 4*cKiB)
	if *a != 4*cKiB {
		t.Fatalf("*a is %d; want %d", *a, 4*cKiB)
	}
	if err := os.Parse("a=2MiB"); err != nil {
		t.Fatalf("os.Parse(%q) error %s", "a=2MiB", err)
	}
	if *a != 2*cMiB {
		t.Fatalf("*a is %d; want %d", *a, 2*cMiB)
	}
}

func TestOptionSet_String(t *testing.T) {
	var os optionSet
	s := os.String("s", "default")
	if *s != "default" {
		t.Fatalf("*s is %q; want %q", *s, "default")
	}
	if err := os.Parse("s=foo"); err != nil {
		t.Fatalf("os.Parse(%q) error %s", "s=foo", err)
	}
	if *s != "foo" {
		t.Fatalf("*s is %q; want %q", *s, "foo")
	}
}

func TestOptionSet_Action(t *testing.T) {
	var os optionSet
	os.Int("a", 3)
	var act bool
	os.SetAction("a", func(*option) { act = true })
	if err := os.Parse("a=0"); err != nil {
		t.Fatalf("os.Parse(%q) error %s", "a=0", err)
	}
	if !act {
		t.Fatalf("action on a didn't fire")
	}
}
