// errorcheck

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Verify that comparisons of slice/map/func values against converted nil
// values are properly rejected.

package p

func bug() {
	type S []byte
	type M map[int]int
	type F func()

	var s S
	var m M
	var f F

	_ = s == S(nil) // ERROR "compare.*to nil|operator \=\= not defined for .|cannot compare"
	_ = S(nil) == s // ERROR "compare.*to nil|operator \=\= not defined for .|cannot compare"
	switch s {
	case S(nil): // ERROR "compare.*to nil|operator \=\= not defined for .|cannot compare"
	}

	_ = m == M(nil) // ERROR "compare.*to nil|operator \=\= not defined for .|cannot compare"
	_ = M(nil) == m // ERROR "compare.*to nil|operator \=\= not defined for .|cannot compare"
	switch m {
	case M(nil): // ERROR "compare.*to nil|operator \=\= not defined for .|cannot compare"
	}

	_ = f == F(nil) // ERROR "compare.*to nil|operator \=\= not defined for .|cannot compare"
	_ = F(nil) == f // ERROR "compare.*to nil|operator \=\= not defined for .|cannot compare"
	switch f {
	case F(nil): // ERROR "compare.*to nil|operator \=\= not defined for .|cannot compare"
	}
}

// GnoError:
// line 21: gno.land/p/filetest/p[gno.land/p/filetest/p/issue13480.go:13:1-39:2].S can only be compared to nil
// line 22: gno.land/p/filetest/p[gno.land/p/filetest/p/issue13480.go:13:1-39:2].S can only be compared to nil
// line 27: gno.land/p/filetest/p[gno.land/p/filetest/p/issue13480.go:13:1-39:2].M can only be compared to nil
// line 28: gno.land/p/filetest/p[gno.land/p/filetest/p/issue13480.go:13:1-39:2].M can only be compared to nil
// line 33: gno.land/p/filetest/p[gno.land/p/filetest/p/issue13480.go:13:1-39:2].F can only be compared to nil
// line 34: gno.land/p/filetest/p[gno.land/p/filetest/p/issue13480.go:13:1-39:2].F can only be compared to nil

// GoTypeCheckError:
// line 21: invalid operation: s == S(nil) (slice can only be compared to nil)
// line 22: invalid operation: S(nil) == s (slice can only be compared to nil)
// line 24: invalid case S(nil) in switch on s (slice can only be compared to nil)
// line 27: invalid operation: m == M(nil) (map can only be compared to nil)
// line 28: invalid operation: M(nil) == m (map can only be compared to nil)
// line 30: invalid case M(nil) in switch on m (map can only be compared to nil)
// line 33: invalid operation: f == F(nil) (func can only be compared to nil)
// line 34: invalid operation: F(nil) == f (func can only be compared to nil)
// line 36: invalid case F(nil) in switch on f (func can only be compared to nil)
