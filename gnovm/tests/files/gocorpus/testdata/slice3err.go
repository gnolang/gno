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
// line 21: middle index required in 3-index slice (and 10 more errors)
// line 22: final index required in 3-index slice (and 10 more errors)
// line 23: final index required in 3-index slice (and 10 more errors)
// line 24: middle index required in 3-index slice (and 10 more errors)
// line 25: middle index required in 3-index slice (and 10 more errors)
// line 33: middle index required in 3-index slice (and 10 more errors)
// line 34: middle index required in 3-index slice (and 10 more errors)
// line 35: final index required in 3-index slice (and 9 more errors)
// line 36: final index required in 3-index slice (and 8 more errors)
// line 37: middle index required in 3-index slice (and 7 more errors)
// line 38: middle index required in 3-index slice (and 6 more errors)
// line 46: middle index required in 3-index slice (and 5 more errors)
// line 47: middle index required in 3-index slice (and 4 more errors)
// line 48: final index required in 3-index slice (and 3 more errors)
// line 49: final index required in 3-index slice (and 2 more errors)
// line 50: middle index required in 3-index slice (and 1 more errors)
// line 51: middle index required in 3-index slice

// GoTypeCheckError:
// line 20: middle index required in 3-index slice (and 10 more errors)
// line 21: middle index required in 3-index slice (and 10 more errors)
// line 22: final index required in 3-index slice (and 10 more errors)
// line 23: final index required in 3-index slice (and 10 more errors)
// line 24: middle index required in 3-index slice (and 10 more errors)
// line 25: middle index required in 3-index slice (and 10 more errors)
// line 33: middle index required in 3-index slice (and 10 more errors)
// line 34: middle index required in 3-index slice (and 10 more errors)
// line 35: final index required in 3-index slice (and 9 more errors)
// line 36: final index required in 3-index slice (and 8 more errors)
// line 37: middle index required in 3-index slice (and 7 more errors)
// line 38: middle index required in 3-index slice (and 6 more errors)
// line 46: middle index required in 3-index slice (and 5 more errors)
// line 47: middle index required in 3-index slice (and 4 more errors)
// line 48: final index required in 3-index slice (and 3 more errors)
// line 49: final index required in 3-index slice (and 2 more errors)
// line 50: middle index required in 3-index slice (and 1 more errors)
// line 51: middle index required in 3-index slice
// line 52: invalid operation: 3-index slice of string
// line 53: invalid operation: 3-index slice of string
// line 57: invalid slice indices: 1 < 2
// line 62: invalid slice indices: 2 < 3
// line 63: invalid slice indices: 1 < 2
// line 64: invalid slice indices: 1 < 2
// line 65: invalid slice indices: 1 < 3
// line 66: invalid slice indices: 2 < 3
// line 68: invalid slice indices: 1 < 2
// line 70: invalid slice indices: 1 < 2
// line 72: invalid slice indices: 1 < 2
// line 75: invalid slice indices: 1 < 2
// line 80: invalid slice indices: 2 < 3
// line 81: invalid slice indices: 1 < 2
// line 82: invalid slice indices: 1 < 2
// line 83: invalid slice indices: 1 < 3
// line 84: invalid slice indices: 2 < 3
// line 86: invalid slice indices: 1 < 2
// line 88: invalid slice indices: 1 < 2
// line 90: invalid slice indices: 1 < 2
// line 93: invalid slice indices: 1 < 2
// line 99: invalid argument: index 11 out of bounds [0:11]
// line 100: invalid argument: index 11 out of bounds [0:11]
// line 101: invalid argument: index 11 out of bounds [0:11]
// line 102: invalid argument: index 11 out of bounds [0:11]
// line 103: invalid argument: index 11 out of bounds [0:11]
// line 104: invalid argument: index 11 out of bounds [0:11]
// line 105: invalid argument: index 11 out of bounds [0:11]
// line 106: invalid argument: index 11 out of bounds [0:11]
// line 107: invalid argument: index 11 out of bounds [0:11]
// line 108: invalid argument: index 11 out of bounds [0:11]
// line 118: invalid slice indices: 3 < 11
// line 119: invalid slice indices: 2 < 11
