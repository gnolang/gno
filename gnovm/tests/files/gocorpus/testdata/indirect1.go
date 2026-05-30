// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Verify that illegal uses of indirection are caught by the compiler.
// Does not compile.

package main

var m0 map[string]int
var m1 *map[string]int
var m2 *map[string]int = &m0
var m3 map[string]int = map[string]int{"a": 1}
var m4 *map[string]int = &m3

var s0 string
var s1 *string
var s2 *string = &s0
var s3 string = "a"
var s4 *string = &s3

var a0 [10]int
var a1 *[10]int
var a2 *[10]int = &a0

var b0 []int
var b1 *[]int
var b2 *[]int = &b0
var b3 []int = []int{1, 2, 3}
var b4 *[]int = &b3

func f() {
	// this is spaced funny so that
	// the compiler will print a different
	// line number for each len call when
	// it decides there are type errors.
	x :=
		len(m0)+
		len(m1)+	// ERROR "illegal|invalid|must be"
		len(m2)+	// ERROR "illegal|invalid|must be"
		len(m3)+
		len(m4)+	// ERROR "illegal|invalid|must be"

		len(s0)+
		len(s1)+	// ERROR "illegal|invalid|must be"
		len(s2)+	// ERROR "illegal|invalid|must be"
		len(s3)+
		len(s4)+	// ERROR "illegal|invalid|must be"

		len(a0)+
		len(a1)+
		len(a2)+

		cap(a0)+
		cap(a1)+
		cap(a2)+

		len(b0)+
		len(b1)+	// ERROR "illegal|invalid|must be"
		len(b2)+	// ERROR "illegal|invalid|must be"
		len(b3)+
		len(b4)+	// ERROR "illegal|invalid|must be"

		cap(b0)+
		cap(b1)+	// ERROR "illegal|invalid|must be"
		cap(b2)+	// ERROR "illegal|invalid|must be"
		cap(b3)+
		cap(b4)	// ERROR "illegal|invalid|must be"
	_ = x
}

// GoTypeCheckError:
// line 41: invalid argument: m1 (variable of type *map[string]int) for built-in len
// line 42: invalid argument: m2 (variable of type *map[string]int) for built-in len
// line 44: invalid argument: m4 (variable of type *map[string]int) for built-in len
// line 47: invalid argument: s1 (variable of type *string) for built-in len
// line 48: invalid argument: s2 (variable of type *string) for built-in len
// line 50: invalid argument: s4 (variable of type *string) for built-in len
// line 61: invalid argument: b1 (variable of type *[]int) for built-in len
// line 62: invalid argument: b2 (variable of type *[]int) for built-in len
// line 64: invalid argument: b4 (variable of type *[]int) for built-in len
// line 67: invalid argument: b1 (variable of type *[]int) for built-in cap
// line 68: invalid argument: b2 (variable of type *[]int) for built-in cap
// line 70: invalid argument: b4 (variable of type *[]int) for built-in cap
