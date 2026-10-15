// errorcheck

// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test "cannot assign" errors

package main

func main() {
	var s string = "hello"
	s[1:2] = "a" // ERROR "cannot assign to .* (\(strings are immutable\))?"
	s[3] = "b"   // ERROR "cannot assign to .* (\(strings are immutable\))?"

	const n int = 1
	const cs string = "hello"
	n = 2        // ERROR "cannot assign to .* (\(declared const\))?"
	cs = "hi"    // ERROR "cannot assign to .* (\(declared const\))?"
	true = false // ERROR "cannot assign to .* (\(declared const\))?"

	var m map[int]struct{ n int }
	m[0].n = 7 // ERROR "cannot assign to struct field .* in map$"

	1 = 7         // ERROR "cannot assign to 1"
	"hi" = 7      // ERROR `cannot assign to "hi"`
	nil = 7       // ERROR "cannot assign to nil"
	len("") = 7   // ERROR `cannot assign to len\(""\)`
	[]int{} = nil // ERROR "cannot assign to \[\]int\{\}"

	var x int = 7
	x + 1 = 7 // ERROR "cannot assign to x \+ 1"
}

// GnoError:
// line 13: cannot assign to s<VPBlock(1,0)>[(const (1 int)):(const (2 int))]
// line 14: cannot assign to s[3] (neither addressable nor a map index expression)
// line 18: cannot assign to const n
// line 19: cannot assign to const cs
// line 20: cannot assign to uverse true
// line 23: cannot assign to struct field m[0].n in map
// line 25: cannot assign to 1
// line 26: cannot assign to "hi"
// line 27: cannot assign to uverse nil
// line 28: cannot assign to len("")
// line 29: cannot assign to (const-type []int){}
// line 32: cannot assign to x + 1

// GnoWordingDiffers:
// line 27: gc expects: cannot assign to nil; gno: cannot assign to uverse nil
// line 29: gc expects: cannot assign to \[\]int\{\}; gno: cannot assign to (const-type []int){}

// Verdict: Clean errorcheck
