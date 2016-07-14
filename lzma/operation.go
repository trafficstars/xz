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

// operation represents an operation in the encoded data stream.
type operation struct {
	distance uint32
	len      uint16
	tag      opTag
	c        byte
}

var nop = operation{}

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

func match(distance uint32, n int) operation {
	return operation{tag: matchTag, distance: distance, len: uint16(n)}
}

func lit(c byte) operation {
	return operation{tag: litTag, len: 1, c: c}
}
