// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main
type I int
type S struct { f map[I]int }
var v1 = S{ make(map[int]int) }		// ERROR "cannot|illegal|incompatible|wrong"
var v2 map[I]int = map[int]int{}	// ERROR "cannot|illegal|incompatible|wrong"
var v3 = S{ make(map[uint]int) }	// ERROR "cannot|illegal|incompatible|wrong"

// GnoError:
// line 10: cannot use map[int]int as map[main.I]int
// line 11: cannot use map[int]int as map[main.I]int
// line 12: cannot use map[uint]int as map[main.I]int

// GoTypeCheckError:
// line 10: cannot use make(map[int]int) (value of type map[int]int) as map[I]int value in struct literal
// line 11: cannot use map[int]int{} (value of type map[int]int) as map[I]int value in variable declaration
// line 12: cannot use make(map[uint]int) (value of type map[uint]int) as map[I]int value in struct literal
