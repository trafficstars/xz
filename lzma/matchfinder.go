// Copyright 2014-2016 Ulrich Kunitz. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lzma

import "errors"

// MatchFinder identifies the algorithm to be used for finding matches.
type MatchFinder byte

// Match finders supported by NewWriterCfg.
const (
	HashChain2 MatchFinder = 0x02
	HashChain3 MatchFinder = 0x03
	HashChain4 MatchFinder = 0x04
	BinTree2   MatchFinder = 0x12
	BinTree3   MatchFinder = 0x13
	BinTree4   MatchFinder = 0x14
)

// maStrings are used by the String method.
var maStrings = map[MatchFinder]string{
	HashChain2: "HashChain2",
	HashChain3: "HashChain3",
	HashChain4: "HashChain4",
	BinTree2:   "BinTree2",
	BinTree3:   "BinTree3",
	BinTree4:   "BinTree4",
}

// String returns a string representation of the Matcher.
func (a MatchFinder) String() string {
	if s, ok := maStrings[a]; ok {
		return s
	}
	return "unknown"
}

var errUnsupportedMatchFinder = errors.New(
	"lzma: unsupported match finder value")

// verify checks whether the matcher value is supported.
func (a MatchFinder) verify() error {
	if a == 0 {
		return nil
	}
	if _, ok := maStrings[a]; !ok {
		return errUnsupportedMatchFinder
	}
	return nil
}

// new creates the matcher based on the matchfinder value.
func (a MatchFinder) new(dictCap int) (m matcher, err error) {
	/* TODO: make this a WriterConfig and Writer2Config operation */
	switch a {
	case 0, HashChain4:
		return newHCFinder(4, dictCap, 4096, 200, 20)
	case HashChain3:
		return newHCFinder(3, dictCap, 4096, 200, 20)
	case HashChain2:
		return newHCFinder(2, dictCap, 4096, 200, 20)
	}
	return nil, errUnsupportedMatchFinder
}
