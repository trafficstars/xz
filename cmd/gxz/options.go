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

const (
	eKiB = 1 << ((iota + 1) * 10)
	eMiB
	eGiB
)

type int64ExtValue int64

func newInt64ExtValue(v int64, p *int64) *int64ExtValue {
	*p = v
	return (*int64ExtValue)(p)
}

func (i *int64ExtValue) String() string {
	n := int64(*i)
	if n%eGiB == 0 {
		return fmt.Sprintf("%dGiB", n/eGiB)
	}
	if n%eMiB == 0 {
		return fmt.Sprintf("%dMiB", n/eMiB)
	}
	if n%eKiB == 0 {
		return fmt.Sprintf("%dKiB", n/eKiB)
	}
	return fmt.Sprintf("%d", n)
}

var (
	kibRegexp = regexp.MustCompile(`(?:KiB|Ki|k|kB|K|KB)$`)
	mibRegexp = regexp.MustCompile(`(?:MiB|Mi|m|M|MB)$`)
	gibRegexp = regexp.MustCompile(`(?:GiB|Gi|g|G|GB)$`)
)

func (i *int64ExtValue) Set(s string) error {
	t := s
	a := 1
	if loc := kibRegexp.FindStringIndex(s); loc != nil {
		t = s[:loc[0]]
		a = eKiB
	} else if loc := mibRegexp.FindStringIndex(s); loc != nil {
		t = s[:loc[0]]
		a = eMiB
	} else if loc := gibRegexp.FindStringIndex(s); loc != nil {
		t = s[:loc[0]]
		a = eGiB
	}
	n, err := strconv.ParseInt(t, 0, 64)
	*i = int64ExtValue(n * int64(a))
	return err

}

func (s *optionSet) Int64Var(p *int64, name string, value int64) {
	v := newInt64Value(value, p)
	s.Var(v, name)
}

func (s *optionSet) Int64(name string, value int64) *int64 {
	p := new(int64)
	s.Int64Var(p, name, value)
	return p
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
