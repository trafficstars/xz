// Copyright 2014-2016 Ulrich Kunitz. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lzma

import (
	"errors"
	"io"
)

// dict provides the dictionary for match finder. It includes additional
// space for the data to be matched.
type dict struct {
	buf      buffer
	head     int64
	capacity int
}

// newEncoderDict creates the encoder dictionary. The argument bufSize
// defines the size of the additional buffer.
func newDict(dictCap, bufSize int) (d *dict, err error) {
	if !(1 <= dictCap && int64(dictCap) <= MaxDictCap) {
		return nil, errors.New(
			"lzma: dictionary capacity out of range")
	}
	if bufSize < 1 {
		return nil, errors.New(
			"lzma: buffer size must be larger than zero")
	}
	d = &dict{
		buf:      *newBuffer(dictCap + bufSize),
		capacity: dictCap,
	}
	return d, nil
}

// Len returns the data available in the encoder dictionary.
func (d *dict) Len() int {
	n := d.buf.Available()
	if int64(n) > d.head {
		return int(d.head)
	}
	return n
}

// dictLen returns the actual length of data in the dictionary.
func (d *dict) dictLen() int {
	if d.head < int64(d.capacity) {
		return int(d.head)
	}
	return d.capacity
}

// Available returns the number of bytes that can be written by a
// following Write call.
func (d *dict) Available() int {
	return d.buf.Available() - d.dictLen()
}

// Write writes data into the dictionary buffer. Note that the position
// of the dictionary head will not be moved. If there is not enough
// space in the buffer ErrNoSpace will be returned.
func (d *dict) Write(p []byte) (n int, err error) {
	m := d.Available()
	if len(p) > m {
		p = p[:m]
		err = ErrNoSpace
	}
	var e error
	if n, e = d.buf.Write(p); e != nil {
		err = e
	}
	return n, err
}

// Pos returns the position of the head.
func (d *dict) Pos() int64 { return d.head }

// Returns the byte as the head. It makes no checks, whether there is
// actually buffered data.
func (d *dict) HeadByte() byte {
	return d.buf.data[d.buf.rear]
}

// ByteAt returns the byte at the given distance.
func (d *dict) ByteAt(distance int) byte {
	if !(0 < distance && distance <= d.Len()) {
		return 0
	}
	i := d.buf.rear - distance
	if i < 0 {
		i += len(d.buf.data)
	}
	return d.buf.data[i]
}

// CopyN copies the last n bytes from the dictionary into the provided
// writer. This is used for copying uncompressed data into an
// uncompressed segment.
func (d *dict) CopyN(w io.Writer, n int) (written int, err error) {
	if n <= 0 {
		return 0, nil
	}
	m := d.Len()
	if n > m {
		n = m
		err = ErrNoSpace
	}
	i := d.buf.rear - n
	var e error
	if i < 0 {
		i += len(d.buf.data)
		if written, e = w.Write(d.buf.data[i:]); e != nil {
			return written, e
		}
		i = 0
	}
	var k int
	k, e = w.Write(d.buf.data[i:d.buf.rear])
	written += k
	if e != nil {
		err = e
	}
	return written, err
}

// Buffered returns the number of bytes in the buffer.
func (d *dict) Buffered() int { return d.buf.Buffered() }

// Peek behaves like read but doesn't advance the reading position.
// While not all the requested may be provided, the function never
// returns an error.
func (d *dict) Peek(p []byte) (n int, err error) { return d.buf.Peek(p) }

// Discard skips the next n bytes to read from the buffer. The actual
// numbers of bytes will be returned.
func (d *dict) Discard(n int) (discarded int, err error) {
	discarded, err = d.buf.Discard(n)
	d.head += int64(discarded)
	return discarded, err
}

// Read provides the standard read method for the dictionary. The
// function might return less data than requested, but it will never
// return an error.
func (d *dict) Read(p []byte) (n int, err error) {
	n, err = d.buf.Read(p)
	d.head += int64(n)
	return n, err
}

// prefixLen returns the length of the common prefix of a and b.
func prefixLen(a, b []byte) int {
	if len(a) > len(b) {
		a, b = b, a
	}
	for i, c := range a {
		if b[i] != c {
			return i
		}
	}
	return len(a)
}

// matchLen returns the length of the common prefix for the given
// distance from the rear and the byte slice p.
func (d *dict) matchLen(distance int, p []byte) int {
	var n int
	b := &d.buf
	i := b.rear - distance
	if i < 0 {
		if n = prefixLen(p, b.data[len(b.data)+i:]); n < -i {
			return n
		}
		p = p[n:]
		i = 0
	}
	n += prefixLen(p, b.data[i:])
	return n
}
