// Copyright 2014-2016 Ulrich Kunitz. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lzma

import "errors"

// MatchFinder identifies the algorithm to be used for finding matches.
type MatchFinder byte

// Supported legacy match finders.
const (
	HashTable4 MatchFinder = iota
	BinaryTree
)

// Match finders supported by NewWriterCfg.
const (
	HashChain2 MatchFinder = 0x02
	HashChain3             = 0x03
	HashChain4             = 0x04
	BinTree2               = 0x12
	BinTree3               = 0x13
	BinTree4               = 0x14
)

// maStrings are used by the String method.
var maStrings = map[MatchFinder]string{
	HashTable4: "HashTable4",
	BinaryTree: "BinaryTree",
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
	switch a {
	case HashTable4:
		return newHashTable(dictCap, 4)
	case BinaryTree:
		return newBinTree(dictCap)
		// TODO
	}
	return nil, errUnsupportedMatchFinder
}
