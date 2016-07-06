package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type value interface {
	String() string
	Set(string) error
}

type option struct {
	name         string
	value        value
	defaultValue string
}

func (o *option) Reset() {
	if err := o.value.Set(o.defaultValue); err != nil {
		panic(err)
	}
}

type optionSet struct {
	options map[string]*option
	parsed  bool
}

type int64Value int64

func newInt64Value(v int64, p *int64) *int64Value {
	*p = v
	return (*int64Value)(p)
}

func (i *int64Value) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 64)
	*i = int64Value(v)
	return err
}

func (i *int64Value) String() string { return fmt.Sprintf("%v", *i) }

func (s *optionSet) Int64Var(p *int64, name string, value int64) {
	v := newInt64Value(value, p)
	s.Var(v, name)
}

func (s *optionSet) Int64(name string, value int64) *int64 {
	p := new(int64)
	s.Int64Var(p, name, value)
	return p
}

type intValue int

func newIntValue(v int, p *int) *intValue {
	*p = v
	return (*intValue)(p)
}

func (i *intValue) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 64)
	*i = intValue(v)
	return err
}

func (i *intValue) String() string { return fmt.Sprintf("%v", *i) }

func (s *optionSet) IntVar(p *int, name string, value int) {
	v := newIntValue(value, p)
	s.Var(v, name)
}

func (s *optionSet) Int(name string, value int) *int {
	p := new(int)
	s.IntVar(p, name, value)
	return p
}

const (
	cKiB = 1 << ((iota + 1) * 10)
	cMiB
	cGiB
)

var (
	kibRegexp = regexp.MustCompile(`(?:KiB|Ki|k|kB|K|KB)$`)
	mibRegexp = regexp.MustCompile(`(?:MiB|Mi|m|M|MB)$`)
	gibRegexp = regexp.MustCompile(`(?:GiB|Gi|g|G|GB)$`)
)

type int64ExtValue int64

func newInt64ExtValue(v int64, p *int64) *int64ExtValue {
	*p = v
	return (*int64ExtValue)(p)
}

func (i *int64ExtValue) String() string {
	n := int64(*i)
	if n%cGiB == 0 {
		return fmt.Sprintf("%dGiB", n/cGiB)
	}
	if n%cMiB == 0 {
		return fmt.Sprintf("%dMiB", n/cMiB)
	}
	if n%cKiB == 0 {
		return fmt.Sprintf("%dKiB", n/cKiB)
	}
	return fmt.Sprintf("%d", n)
}

func (i *int64ExtValue) Set(s string) error {
	t := s
	a := 1
	if loc := kibRegexp.FindStringIndex(s); loc != nil {
		t = s[:loc[0]]
		a = cKiB
	} else if loc := mibRegexp.FindStringIndex(s); loc != nil {
		t = s[:loc[0]]
		a = cMiB
	} else if loc := gibRegexp.FindStringIndex(s); loc != nil {
		t = s[:loc[0]]
		a = cGiB
	}
	n, err := strconv.ParseInt(t, 0, 64)
	*i = int64ExtValue(n * int64(a))
	return err

}

func (s *optionSet) Int64ExtVar(p *int64, name string, value int64) {
	v := newInt64ExtValue(value, p)
	s.Var(v, name)
}

func (s *optionSet) Int64Ext(name string, value int64) *int64 {
	p := new(int64)
	s.Int64ExtVar(p, name, value)
	return p
}

type intExtValue int

func newIntExtValue(v int, p *int) *intExtValue {
	*p = v
	return (*intExtValue)(p)
}

func (i *intExtValue) String() string {
	n := int(*i)
	if n%cGiB == 0 {
		return fmt.Sprintf("%dGiB", n/cGiB)
	}
	if n%cMiB == 0 {
		return fmt.Sprintf("%dMiB", n/cMiB)
	}
	if n%cKiB == 0 {
		return fmt.Sprintf("%dKiB", n/cKiB)
	}
	return fmt.Sprintf("%d", n)
}

func (i *intExtValue) Set(s string) error {
	t := s
	a := 1
	if loc := kibRegexp.FindStringIndex(s); loc != nil {
		t = s[:loc[0]]
		a = cKiB
	} else if loc := mibRegexp.FindStringIndex(s); loc != nil {
		t = s[:loc[0]]
		a = cMiB
	} else if loc := gibRegexp.FindStringIndex(s); loc != nil {
		t = s[:loc[0]]
		a = cGiB
	}
	n, err := strconv.ParseInt(t, 0, 64)
	*i = intExtValue(n * int64(a))
	return err

}

func (s *optionSet) IntExtVar(p *int, name string, value int) {
	v := newIntExtValue(value, p)
	s.Var(v, name)
}

func (s *optionSet) IntExt(name string, value int) *int {
	p := new(int)
	s.IntExtVar(p, name, value)
	return p
}

type stringValue string

func newStringValue(value string, p *string) *stringValue {
	*p = value
	return (*stringValue)(p)
}

func (s *stringValue) String() string { return string(*s) }

func (s *stringValue) Set(str string) error {
	*s = stringValue(str)
	return nil
}

func (s *optionSet) StringVar(p *string, name string, value string) {
	v := newStringValue(value, p)
	s.Var(v, name)
}

func (s *optionSet) String(name string, value string) *string {
	p := new(string)
	s.StringVar(p, name, value)
	return p
}

func (s *optionSet) Parse(t string) error {
	if s.Parsed() {
		return errors.New("option set already parsed")
	}
	for _, a := range strings.Split(t, ",") {
		e := strings.SplitN(a, "=", 2)
		name, value := e[0], e[1]
		o, ok := s.options[name]
		if !ok {
			s.Reset()
			return fmt.Errorf("option %q not supported", name)
		}
		if err := o.value.Set(value); err != nil {
			s.Reset()
			return fmt.Errorf("option %s: %s", name, err)
		}
	}
	s.parsed = true
	return nil
}

func (s *optionSet) Reset() {
	for _, o := range s.options {
		o.Reset()
	}
	s.parsed = false
}

func (s *optionSet) Parsed() bool {
	return s.parsed
}

func (s *optionSet) Var(v value, name string) {
	if _, exists := s.options[name]; exists {
		panic(fmt.Errorf("option %s already exists", name))
	}
	if s.options == nil {
		s.options = make(map[string]*option)
	}
	s.options[name] = &option{name, v, v.String()}
}
