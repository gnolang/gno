// errorcheck

// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 7214: No duplicate key error for maps with interface{} key type

package p

var _ = map[interface{}]int{2: 1, 2: 1} // ERROR "duplicate key"
var _ = map[interface{}]int{int(2): 1, int16(2): 1}
var _ = map[interface{}]int{int16(2): 1, int16(2): 1} // ERROR "duplicate key"

type S string

var _ = map[interface{}]int{"a": 1, "a": 1} // ERROR "duplicate key"
var _ = map[interface{}]int{"a": 1, S("a"): 1}
var _ = map[interface{}]int{S("a"): 1, S("a"): 1} // ERROR "duplicate key"

type I interface {
	f()
}

type N int

func (N) f() {}

var _ = map[I]int{N(0): 1, N(2): 1}
var _ = map[I]int{N(2): 1, N(2): 1} // ERROR "duplicate key"

// GnoError:
// line 11: duplicate key (2 int) in map literal
// line 13: duplicate key (2 int16) in map literal
// line 17: duplicate key ("a" string) in map literal
// line 19: duplicate key ("a" gno.land/p/filetest/p.S) in map literal
// line 30: duplicate key (2 gno.land/p/filetest/p.N) in map literal

// GoTypeCheckError:
// line 11: duplicate key 2 in map literal
// line 13: duplicate key 2 in map literal
// line 17: duplicate key "a" in map literal
// line 19: duplicate key "a" in map literal
// line 30: duplicate key 2 in map literal
