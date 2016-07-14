// Copyright 2014-2016 Ulrich Kunitz. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lzma

import (
	"errors"
	"fmt"
	"unicode"
)

// optag identifies the type of the operation.
type opTag byte

// Tags match and lit for an operation. The zero operation is nop, which
// is invalid in an LZMA stream.
const (
	matchTag opTag = iota + 1
	litTag
)

// operation represents an operation in the encoded data stream. Since
// we are using an uint32 for the distance and don't subtract 1, we
// cannot express an EOS Marker with operation. The decoder returns an
// error and the encoder doesn't use operation to write the EOS marker.
type operation struct {
	distance uint32
	len      uint16
	tag      opTag
	c        byte
}

// nop is no operation
var nop = operation{}

// verify checks whether the operation is valid. A nop is not a valid
// operation.
func (op operation) verify() error {
	switch op.tag {
	case matchTag:
		if !(minDistance <= op.distance && op.distance <= maxDistance) {
			return errors.New(
				"lzma: operation distance out of range")
		}
	case litTag:
		break
	default:
		return errors.New("lzma: wrong tag")
	}
	if !(1 <= op.len && op.len <= maxMatchLen) {
		return errors.New("lzma: operation length out of range")
	}
	return nil
}

// Strings prints a readable representation of an operation.
func (op operation) String() string {
	switch op.tag {
	case matchTag:
		return fmt.Sprintf("M{%d,%d}", op.distance, op.len)
	case litTag:
		var c byte
		if unicode.IsPrint(rune(op.c)) {
			c = op.c
		} else {
			c = '.'
		}
		return fmt.Sprintf("L{%c/%02x}", c, op.c)
	default:
		return "NOP"
	}
}

// match creates a match operation.
func match(distance uint32, n int) operation {
	return operation{tag: matchTag, distance: distance, len: uint16(n)}
}

// lit creates a literal operation.
func lit(c byte) operation {
	return operation{tag: litTag, len: 1, c: c}
}
