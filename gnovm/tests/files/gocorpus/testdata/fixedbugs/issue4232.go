// errorcheck

// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// issue 4232
// issue 7200

package p

func f() {
	var a [10]int
	_ = a[-1]  // ERROR "invalid array index -1|index out of bounds|must not be negative"
	_ = a[-1:] // ERROR "invalid slice index -1|index out of bounds|must not be negative"
	_ = a[:-1] // ERROR "invalid slice index -1|index out of bounds|must not be negative"
	_ = a[10]  // ERROR "invalid array index 10|index .*out of bounds"
	_ = a[9:10]
	_ = a[10:10]
	_ = a[9:12]            // ERROR "invalid slice index 12|index .*out of bounds"
	_ = a[11:12]           // ERROR "invalid slice index 11|index .*out of bounds"
	_ = a[1<<100 : 1<<110] // ERROR "overflows int|integer constant overflow|invalid slice index 1 << 100|index out of bounds"

	var s []int
	_ = s[-1]  // ERROR "invalid slice index -1|index .*out of bounds|must not be negative"
	_ = s[-1:] // ERROR "invalid slice index -1|index .*out of bounds|must not be negative"
	_ = s[:-1] // ERROR "invalid slice index -1|index .*out of bounds|must not be negative"
	_ = s[10]
	_ = s[9:10]
	_ = s[10:10]
	_ = s[9:12]
	_ = s[11:12]
	_ = s[1<<100 : 1<<110] // ERROR "overflows int|integer constant overflow|invalid slice index 1 << 100|index out of bounds"

	const c = "foofoofoof"
	_ = c[-1]  // ERROR "invalid string index -1|index out of bounds|must not be negative"
	_ = c[-1:] // ERROR "invalid slice index -1|index out of bounds|must not be negative"
	_ = c[:-1] // ERROR "invalid slice index -1|index out of bounds|must not be negative"
	_ = c[10]  // ERROR "invalid string index 10|index .*out of bounds"
	_ = c[9:10]
	_ = c[10:10]
	_ = c[9:12]            // ERROR "invalid slice index 12|index .*out of bounds"
	_ = c[11:12]           // ERROR "invalid slice index 11|index .*out of bounds"
	_ = c[1<<100 : 1<<110] // ERROR "overflows int|integer constant overflow|invalid slice index 1 << 100|index out of bounds"

	var t string
	_ = t[-1]  // ERROR "invalid string index -1|index out of bounds|must not be negative"
	_ = t[-1:] // ERROR "invalid slice index -1|index out of bounds|must not be negative"
	_ = t[:-1] // ERROR "invalid slice index -1|index out of bounds|must not be negative"
	_ = t[10]
	_ = t[9:10]
	_ = t[10:10]
	_ = t[9:12]
	_ = t[11:12]
	_ = t[1<<100 : 1<<110] // ERROR "overflows int|integer constant overflow|invalid slice index 1 << 100|index out of bounds"
}

// GnoError:
// line 14: invalid argument: index must not be negative: (-1 <untyped> bigint)
// line 15: invalid argument: index must not be negative: (-1 <untyped> bigint)
// line 16: invalid argument: index must not be negative: (-1 <untyped> bigint)
// line 22: bigint overflows target kind
// line 25: invalid argument: index must not be negative: (-1 <untyped> bigint)
// line 26: invalid argument: index must not be negative: (-1 <untyped> bigint)
// line 27: invalid argument: index must not be negative: (-1 <untyped> bigint)
// line 33: bigint overflows target kind
// line 36: invalid argument: index must not be negative: (-1 <untyped> bigint)
// line 37: invalid argument: index must not be negative: (-1 <untyped> bigint)
// line 38: invalid argument: index must not be negative: (-1 <untyped> bigint)
// line 44: bigint overflows target kind
// line 47: invalid argument: index must not be negative: (-1 <untyped> bigint)
// line 48: invalid argument: index must not be negative: (-1 <untyped> bigint)
// line 49: invalid argument: index must not be negative: (-1 <untyped> bigint)
// line 55: bigint overflows target kind

// GoTypeCheckError:
// line 14: invalid argument: index -1 (constant of type int) must not be negative
// line 15: invalid argument: index -1 (constant of type int) must not be negative
// line 16: invalid argument: index -1 (constant of type int) must not be negative
// line 17: invalid argument: index 10 out of bounds [0:10]
// line 20: invalid argument: index 12 out of bounds [0:11]
// line 21: invalid argument: index 11 out of bounds [0:11]
// line 22: 1 << 100 (untyped int constant 1267650600228229401496703205376) overflows int
// line 25: invalid argument: index -1 (constant of type int) must not be negative
// line 26: invalid argument: index -1 (constant of type int) must not be negative
// line 27: invalid argument: index -1 (constant of type int) must not be negative
// line 33: 1 << 100 (untyped int constant 1267650600228229401496703205376) overflows int
// line 36: invalid argument: index -1 (constant of type int) must not be negative
// line 37: invalid argument: index -1 (constant of type int) must not be negative
// line 38: invalid argument: index -1 (constant of type int) must not be negative
// line 39: invalid argument: index 10 out of bounds [0:10]
// line 42: invalid argument: index 12 out of bounds [0:11]
// line 43: invalid argument: index 11 out of bounds [0:11]
// line 44: 1 << 100 (untyped int constant 1267650600228229401496703205376) overflows int
// line 47: invalid argument: index -1 (constant of type int) must not be negative
// line 48: invalid argument: index -1 (constant of type int) must not be negative
// line 49: invalid argument: index -1 (constant of type int) must not be negative
// line 55: 1 << 100 (untyped int constant 1267650600228229401496703205376) overflows int
