package main

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// value provides an interface to a value store.
type value interface {
	String() string
	Set(string) error
}

// action should be executed if an option is updated.
type action func(o *option)

// option keeps information about a specific option.
type option struct {
	name         string
	value        value
	defaultValue string
	action       action
}

// Reset sets the option to its default value.
func (o *option) Reset() {
	if err := o.value.Set(o.defaultValue); err != nil {
		panic(err)
	}
}

// optionSet supports the parsing of options in the format
// opt1=1,opt2=abc.
type optionSet struct {
	options map[string]*option
	parsed  bool
}

// int64Value stores an int64 option value.
type int64Value int64

// newInt64Value creates a new int64 option value.
func newInt64Value(v int64, p *int64) *int64Value {
	*p = v
	return (*int64Value)(p)
}

// Set sets the int64 value.
func (i *int64Value) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 64)
	*i = int64Value(v)
	return err
}

// String returns a represantion of in64 value.
func (i *int64Value) String() string { return fmt.Sprintf("%v", *i) }

// Int64Var adds an int64 variable to the option set.
func (s *optionSet) Int64Var(p *int64, name string, value int64) {
	v := newInt64Value(value, p)
	s.Var(v, name)
}

// Int64 creates an int64 value and adds it to the option set.
func (s *optionSet) Int64(name string, value int64) *int64 {
	p := new(int64)
	s.Int64Var(p, name, value)
	return p
}

// intValue represents an int value in an option.
type intValue int

// newIntValue creates a new int value.
func newIntValue(v int, p *int) *intValue {
	*p = v
	return (*intValue)(p)
}

// Set sets the int value.
func (i *intValue) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 64)
	*i = intValue(v)
	return err
}

// String returns a string representation of the int value.
func (i *intValue) String() string { return fmt.Sprintf("%v", *i) }

// IntVar adds a new int option to the option set.
func (s *optionSet) IntVar(p *int, name string, value int) {
	v := newIntValue(value, p)
	s.Var(v, name)
}

// Int addes a new int option to the option set.
func (s *optionSet) Int(name string, value int) *int {
	p := new(int)
	s.IntVar(p, name, value)
	return p
}

// Constants for Kibibyte, Mebibyte and Gibibyte sizes.
const (
	cKiB = 1 << ((iota + 1) * 10)
	cMiB
	cGiB
)

// Regular expressions for extended option values.
var (
	kibRegexp = regexp.MustCompile(`(?:KiB|Ki|k|kB|K|KB)$`)
	mibRegexp = regexp.MustCompile(`(?:MiB|Mi|m|M|MB)$`)
	gibRegexp = regexp.MustCompile(`(?:GiB|Gi|g|G|GB)$`)
)

// int64ExtValue represents an extended intt64 value.
type int64ExtValue int64

// newInt64ExtValue creates a new extended int64 value.
func newInt64ExtValue(v int64, p *int64) *int64ExtValue {
	*p = v
	return (*int64ExtValue)(p)
}

// String returns a string representation of the int64 extended value.
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

// Set sets the int64 extended value.
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

// Int64ExtVar adds an extended int64 option to the option set.
func (s *optionSet) Int64ExtVar(p *int64, name string, value int64) {
	v := newInt64ExtValue(value, p)
	s.Var(v, name)
}

// Int64Ext adds an extended int64 option to the option set.
func (s *optionSet) Int64Ext(name string, value int64) *int64 {
	p := new(int64)
	s.Int64ExtVar(p, name, value)
	return p
}

// intExtValue stores the extended int value.
type intExtValue int

// newIntExtValue creates a new extended int value.
func newIntExtValue(v int, p *int) *intExtValue {
	*p = v
	return (*intExtValue)(p)
}

// String returns a string representatiin of the extended int.
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

// Set modifies the extended int value.
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

// IntExtVar adds an extended int option to the option set.
func (s *optionSet) IntExtVar(p *int, name string, value int) {
	v := newIntExtValue(value, p)
	s.Var(v, name)
}

// IntExt adds an extended int option to the option set.
func (s *optionSet) IntExt(name string, value int) *int {
	p := new(int)
	s.IntExtVar(p, name, value)
	return p
}

// stringValue represents a string value in an option.
type stringValue string

// newStringValue creates a new string option value.
func newStringValue(value string, p *string) *stringValue {
	*p = value
	return (*stringValue)(p)
}

// String returns the string representation of the value.
func (s *stringValue) String() string { return string(*s) }

// Set sets the string value to a new string. The function never returns
// an error.
func (s *stringValue) Set(str string) error {
	*s = stringValue(str)
	return nil
}

// StringVar adds a new string option to the option set.
func (s *optionSet) StringVar(p *string, name string, value string) {
	v := newStringValue(value, p)
	s.Var(v, name)
}

// String adds a new string option to the option set.
func (s *optionSet) String(name string, value string) *string {
	p := new(string)
	s.StringVar(p, name, value)
	return p
}

// Parses the string and changes the option values accordingly.
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
		if o.action != nil {
			o.action(o)
		}
	}
	s.parsed = true
	return nil
}

// Reset sets all options back to their default values.
func (s *optionSet) Reset() {
	for _, o := range s.options {
		o.Reset()
	}
	s.parsed = false
}

// Parsed returns whether the options have already been parsed.
func (s *optionSet) Parsed() bool {
	return s.parsed
}

// Var adds the give value using the given name.
func (s *optionSet) Var(v value, name string) {
	if _, exists := s.options[name]; exists {
		panic(fmt.Errorf("option %s already exists", name))
	}
	if s.options == nil {
		s.options = make(map[string]*option)
	}
	s.options[name] = &option{name: name, value: v,
		defaultValue: v.String()}
}

// SetAction sets an action for an option.
func (s *optionSet) SetAction(name string, a action) {
	o, ok := s.options[name]
	if !ok {
		panic(fmt.Errorf("option %q not found in SetAction", name))
	}
	o.action = a
}

type lzmaOptions struct {
	preset  int
	dictCap int64
	lc      int
	lp      int
	pb      int
	mf      string
	mode    string
	niceLen int
	depth   int
}

func newLZMAOptions() *lzmaOptions {
	return new(lzmaOptions).applyPreset(6)
}

// preset sets the options to a specfic preset.
func (o *lzmaOptions) applyPreset(preset int) *lzmaOptions {
	if !(0 <= preset && preset <= 9) {
		panic(fmt.Errorf("preset %d unsupported", preset))
	}
	*o = lzmaOptions{
		preset:  preset,
		dictCap: 1 << lzmaDictCapExps[preset],
		lc:      3,
		lp:      0,
		pb:      2,
	}
	// match finder
	switch preset {
	case 0:
		o.mf = "hc3"
	case 1, 2, 3:
		o.mf = "hc4"
	default:
		o.mf = "bt4"
	}
	// mode
	if preset <= 3 {
		o.mode = "fast"
	} else {
		o.mode = "normal"
	}
	// niceLen
	switch preset {
	case 0, 1:
		o.niceLen = 128
	case 2, 3:
		o.niceLen = 273
	case 4:
		o.niceLen = 16
	case 5:
		o.niceLen = 32
	default:
		o.niceLen = 64
	}
	// depth
	switch preset {
	case 0:
		o.depth = 4
	case 1:
		o.depth = 8
	case 2:
		o.depth = 24
	case 3:
		o.depth = 48
	}
	return o
}

func (o *lzmaOptions) Set(s string) error {
	var opts optionSet
	opts.IntVar(&o.preset, "preset", o.preset)
	opts.Int64ExtVar(&o.dictCap, "dict", o.dictCap)
	opts.IntVar(&o.lc, "lc", o.lc)
	opts.IntVar(&o.lp, "lp", o.lp)
	opts.IntVar(&o.pb, "pb", o.pb)
	opts.StringVar(&o.mf, "mf", o.mf)
	opts.StringVar(&o.mode, "mode", o.mode)
	opts.IntVar(&o.niceLen, "nice", o.niceLen)
	opts.IntVar(&o.depth, "depth", o.depth)
	opts.SetAction("preset", func(*option) { o.applyPreset(o.preset) })
	if err := opts.Parse(s); err != nil {
		return err
	}
	return nil
}

func (o *lzmaOptions) Update() {
	panic("unexpected update")
}

func (o *lzmaOptions) Get() interface{} { return *o }

func (o *lzmaOptions) String() string {
	var buf bytes.Buffer
	if o.preset != 6 {
		fmt.Fprintf(&buf, "preset=%d,", o.preset)
	}
	p := new(lzmaOptions).applyPreset(o.preset)
	if o.dictCap != p.dictCap {
		fmt.Fprintf(&buf, "dict=%s,",
			(*int64ExtValue)(&o.dictCap).String())
	}
	if o.lc != p.lc {
		fmt.Fprintf(&buf, "lc=%d,", o.lc)
	}
	if o.lp != p.lp {
		fmt.Fprintf(&buf, "lp=%d,", o.lp)
	}
	if o.pb != p.pb {
		fmt.Fprintf(&buf, "pb=%d,", o.pb)
	}
	if o.mf != p.mf {
		fmt.Fprintf(&buf, "mf=%s,", o.mf)
	}
	if o.mode != p.mode {
		fmt.Fprintf(&buf, "mode=%s,", o.mode)
	}
	if o.niceLen != p.niceLen {
		fmt.Fprintf(&buf, "nice=%d,", o.niceLen)
	}
	if o.depth != p.depth {
		fmt.Fprintf(&buf, "depth=%d,", o.depth)
	}
	s := buf.String()
	if len(s) > 0 && s[len(s)-1] == ',' {
		s = s[:len(s)-1]
	}
	return s
}
