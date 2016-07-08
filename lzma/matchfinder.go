// Copyright 2014-2016 Ulrich Kunitz. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lzma

import "errors"

// MatchFinder identifies an algorithm to find matches in the
// dictionary.
type MatchFinder byte

// Supported matcher algorithms.
const (
	HashTable4 MatchFinder = iota
	BinaryTree
)

// maStrings are used by the String method.
var maStrings = map[MatchFinder]string{
	HashTable4: "HashTable4",
	BinaryTree: "BinaryTree",
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
	if _, ok := maStrings[a]; !ok {
		return errUnsupportedMatchFinder
	}
	return nil
}

func (a MatchFinder) new(dictCap int) (m matcher, err error) {
	switch a {
	case HashTable4:
		return newHashTable(dictCap, 4)
	case BinaryTree:
		return newBinTree(dictCap)
	}
	return nil, errUnsupportedMatchFinder
}
