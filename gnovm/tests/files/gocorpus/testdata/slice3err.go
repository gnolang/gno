// errorcheck

// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

var array *[10]int
var slice []int
var str string
var i, j, k int

func f() {
	// check what missing arguments are allowed
	_ = array[:]
	_ = array[i:]
	_ = array[:j]
	_ = array[i:j]
	_ = array[::] // ERROR "middle index required in 3-index slice|invalid slice indices" "final index required in 3-index slice"
	_ = array[i::] // ERROR "middle index required in 3-index slice|invalid slice indices" "final index required in 3-index slice"
	_ = array[:j:] // ERROR "final index required in 3-index slice|invalid slice indices"
	_ = array[i:j:] // ERROR "final index required in 3-index slice|invalid slice indices"
	_ = array[::k] // ERROR "middle index required in 3-index slice|invalid slice indices"
	_ = array[i::k] // ERROR "middle index required in 3-index slice|invalid slice indices"
	_ = array[:j:k]
	_ = array[i:j:k]
	
	_ = slice[:]
	_ = slice[i:]
	_ = slice[:j]
	_ = slice[i:j]
	_ = slice[::] // ERROR "middle index required in 3-index slice|invalid slice indices" "final index required in 3-index slice"
	_ = slice[i::] // ERROR "middle index required in 3-index slice|invalid slice indices" "final index required in 3-index slice"
	_ = slice[:j:] // ERROR "final index required in 3-index slice|invalid slice indices"
	_ = slice[i:j:] // ERROR "final index required in 3-index slice|invalid slice indices"
	_ = slice[::k] // ERROR "middle index required in 3-index slice|invalid slice indices"
	_ = slice[i::k] // ERROR "middle index required in 3-index slice|invalid slice indices"
	_ = slice[:j:k]
	_ = slice[i:j:k]
	
	_ = str[:]
	_ = str[i:]
	_ = str[:j]
	_ = str[i:j]
	_ = str[::] // ERROR "3-index slice of string" "middle index required in 3-index slice" "final index required in 3-index slice"
	_ = str[i::] // ERROR "3-index slice of string" "middle index required in 3-index slice" "final index required in 3-index slice"
	_ = str[:j:] // ERROR "3-index slice of string" "final index required in 3-index slice"
	_ = str[i:j:] // ERROR "3-index slice of string" "final index required in 3-index slice"
	_ = str[::k] // ERROR "3-index slice of string" "middle index required in 3-index slice"
	_ = str[i::k] // ERROR "3-index slice of string" "middle index required in 3-index slice"
	_ = str[:j:k] // ERROR "3-index slice of string"
	_ = str[i:j:k] // ERROR "3-index slice of string"

	// check invalid indices
	_ = array[1:2]
	_ = array[2:1] // ERROR "invalid slice index|invalid slice indices|inverted slice"
	_ = array[2:2]
	_ = array[i:1]
	_ = array[1:j]
	_ = array[1:2:3]
	_ = array[1:3:2] // ERROR "invalid slice index|invalid slice indices|inverted slice"
	_ = array[2:1:3] // ERROR "invalid slice index|invalid slice indices|inverted slice"
	_ = array[2:3:1] // ERROR "invalid slice index|invalid slice indices|inverted slice"
	_ = array[3:1:2] // ERROR "invalid slice index|invalid slice indices|inverted slice"
	_ = array[3:2:1] // ERROR "invalid slice index|invalid slice indices|inverted slice"
	_ = array[i:1:2]
	_ = array[i:2:1] // ERROR "invalid slice index|invalid slice indices|inverted slice"
	_ = array[1:j:2]
	_ = array[2:j:1] // ERROR "invalid slice index|invalid slice indices"
	_ = array[1:2:k]
	_ = array[2:1:k] // ERROR "invalid slice index|invalid slice indices|inverted slice"
	
	_ = slice[1:2]
	_ = slice[2:1] // ERROR "invalid slice index|invalid slice indices|inverted slice"
	_ = slice[2:2]
	_ = slice[i:1]
	_ = slice[1:j]
	_ = slice[1:2:3]
	_ = slice[1:3:2] // ERROR "invalid slice index|invalid slice indices|inverted slice"
	_ = slice[2:1:3] // ERROR "invalid slice index|invalid slice indices|inverted slice"
	_ = slice[2:3:1] // ERROR "invalid slice index|invalid slice indices|inverted slice"
	_ = slice[3:1:2] // ERROR "invalid slice index|invalid slice indices|inverted slice"
	_ = slice[3:2:1] // ERROR "invalid slice index|invalid slice indices|inverted slice"
	_ = slice[i:1:2]
	_ = slice[i:2:1] // ERROR "invalid slice index|invalid slice indices|inverted slice"
	_ = slice[1:j:2]
	_ = slice[2:j:1] // ERROR "invalid slice index|invalid slice indices"
	_ = slice[1:2:k]
	_ = slice[2:1:k] // ERROR "invalid slice index|invalid slice indices|inverted slice"
	
	_ = str[1:2]
	_ = str[2:1] // ERROR "invalid slice index|invalid slice indices|inverted slice"
	_ = str[2:2]
	_ = str[i:1]
	_ = str[1:j]

	// check out of bounds indices on array
	_ = array[11:11] // ERROR "out of bounds"
	_ = array[11:12] // ERROR "out of bounds"
	_ = array[11:] // ERROR "out of bounds"
	_ = array[:11] // ERROR "out of bounds"
	_ = array[1:11] // ERROR "out of bounds"
	_ = array[1:11:12] // ERROR "out of bounds"
	_ = array[1:2:11] // ERROR "out of bounds"
	_ = array[1:11:3] // ERROR "out of bounds|invalid slice index"
	_ = array[11:2:3] // ERROR "out of bounds|inverted slice|invalid slice index"
	_ = array[11:12:13] // ERROR "out of bounds"

	// slice bounds not checked
	_ = slice[11:11]
	_ = slice[11:12]
	_ = slice[11:]
	_ = slice[:11]
	_ = slice[1:11]
	_ = slice[1:11:12]
	_ = slice[1:2:11]
	_ = slice[1:11:3] // ERROR "invalid slice index|invalid slice indices"
	_ = slice[11:2:3] // ERROR "invalid slice index|invalid slice indices|inverted slice"
	_ = slice[11:12:13]
}

// GnoError:
// line 20: middle index required in 3-index slice (and 10 more errors)

// GoTypeCheckError:
// line 20

// UncaughtError:
// line 21: uncaught; gc expects: middle index required in 3-index slice|invalid slice indices | final index required in 3-index slice
// line 22: uncaught; gc expects: final index required in 3-index slice|invalid slice indices
// line 23: uncaught; gc expects: final index required in 3-index slice|invalid slice indices
// line 24: uncaught; gc expects: middle index required in 3-index slice|invalid slice indices
// line 25: uncaught; gc expects: middle index required in 3-index slice|invalid slice indices
// line 33: uncaught; gc expects: middle index required in 3-index slice|invalid slice indices | final index required in 3-index slice
// line 34: uncaught; gc expects: middle index required in 3-index slice|invalid slice indices | final index required in 3-index slice
// line 35: uncaught; gc expects: final index required in 3-index slice|invalid slice indices
// line 36: uncaught; gc expects: final index required in 3-index slice|invalid slice indices
// line 37: uncaught; gc expects: middle index required in 3-index slice|invalid slice indices
// line 38: uncaught; gc expects: middle index required in 3-index slice|invalid slice indices
// line 46: uncaught; gc expects: 3-index slice of string | middle index required in 3-index slice | final index required in 3-index slice
// line 47: uncaught; gc expects: 3-index slice of string | middle index required in 3-index slice | final index required in 3-index slice
// line 48: uncaught; gc expects: 3-index slice of string | final index required in 3-index slice
// line 49: uncaught; gc expects: 3-index slice of string | final index required in 3-index slice
// line 50: uncaught; gc expects: 3-index slice of string | middle index required in 3-index slice
// line 51: uncaught; gc expects: 3-index slice of string | middle index required in 3-index slice
// line 52: uncaught; gc expects: 3-index slice of string
// line 53: uncaught; gc expects: 3-index slice of string
// line 57: uncaught; gc expects: invalid slice index|invalid slice indices|inverted slice
// line 62: uncaught; gc expects: invalid slice index|invalid slice indices|inverted slice
// line 63: uncaught; gc expects: invalid slice index|invalid slice indices|inverted slice
// line 64: uncaught; gc expects: invalid slice index|invalid slice indices|inverted slice
// line 65: uncaught; gc expects: invalid slice index|invalid slice indices|inverted slice
// line 66: uncaught; gc expects: invalid slice index|invalid slice indices|inverted slice
// line 68: uncaught; gc expects: invalid slice index|invalid slice indices|inverted slice
// line 70: uncaught; gc expects: invalid slice index|invalid slice indices
// line 72: uncaught; gc expects: invalid slice index|invalid slice indices|inverted slice
// line 75: uncaught; gc expects: invalid slice index|invalid slice indices|inverted slice
// line 80: uncaught; gc expects: invalid slice index|invalid slice indices|inverted slice
// line 81: uncaught; gc expects: invalid slice index|invalid slice indices|inverted slice
// line 82: uncaught; gc expects: invalid slice index|invalid slice indices|inverted slice
// line 83: uncaught; gc expects: invalid slice index|invalid slice indices|inverted slice
// line 84: uncaught; gc expects: invalid slice index|invalid slice indices|inverted slice
// line 86: uncaught; gc expects: invalid slice index|invalid slice indices|inverted slice
// line 88: uncaught; gc expects: invalid slice index|invalid slice indices
// line 90: uncaught; gc expects: invalid slice index|invalid slice indices|inverted slice
// line 93: uncaught; gc expects: invalid slice index|invalid slice indices|inverted slice
// line 99: uncaught; gc expects: out of bounds
// line 100: uncaught; gc expects: out of bounds
// line 101: uncaught; gc expects: out of bounds
// line 102: uncaught; gc expects: out of bounds
// line 103: uncaught; gc expects: out of bounds
// line 104: uncaught; gc expects: out of bounds
// line 105: uncaught; gc expects: out of bounds
// line 106: uncaught; gc expects: out of bounds|invalid slice index
// line 107: uncaught; gc expects: out of bounds|inverted slice|invalid slice index
// line 108: uncaught; gc expects: out of bounds
// line 118: uncaught; gc expects: invalid slice index|invalid slice indices
// line 119: uncaught; gc expects: invalid slice index|invalid slice indices|inverted slice

// Verdict: KnownIssue errorcheck: 0 over-strict line(s), 50 uncaught marker(s)
