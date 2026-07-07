// errorcheck

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Verify that illegal composite literals are detected.
// Does not compile.

package main

var m map[int][3]int

func f() [3]int

func fp() *[3]int

var mp map[int]*[3]int

var (
	_ = [3]int{1, 2, 3}[:] // ERROR "slice of unaddressable value"
	_ = m[0][:]            // ERROR "slice of unaddressable value"
	_ = f()[:]             // ERROR "slice of unaddressable value"

	_ = 301[:]  // ERROR "cannot slice|attempt to slice object that is not"
	_ = 3.1[:]  // ERROR "cannot slice|attempt to slice object that is not"
	_ = true[:] // ERROR "cannot slice|attempt to slice object that is not"

	// these are okay because they are slicing a pointer to an array
	_ = (&[3]int{1, 2, 3})[:]
	_ = mp[0][:]
	_ = fp()[:]
)

type T struct {
	i    int
	f    float64
	s    string
	next *T
}

type TP *T
type Ti int

var (
	_ = &T{0, 0, "", nil}               // ok
	_ = &T{i: 0, f: 0, s: "", next: {}} // ERROR "missing type in composite literal|omit types within composite literal"
	_ = &T{0, 0, "", {}}                // ERROR "missing type in composite literal|omit types within composite literal"
	_ = TP{i: 0, f: 0, s: ""}           // ERROR "invalid composite literal type TP"
	_ = &Ti{}                           // ERROR "invalid composite literal type Ti|expected.*type for composite literal"
)

type M map[T]T

var (
	_ = M{{i: 1}: {i: 2}}
	_ = M{T{i: 1}: {i: 2}}
	_ = M{{i: 1}: T{i: 2}}
	_ = M{T{i: 1}: T{i: 2}}
)

type S struct{ s [1]*M1 }
type M1 map[S]int

var _ = M1{{s: [1]*M1{&M1{{}: 1}}}: 2}

// GnoError:
// line 14: function f does not have a body but is not natively defined (did you build after pulling from the repository?)
// line 16: function fp does not have a body but is not natively defined (did you build after pulling from the repository?)
// line 25: cannot slice variable of type <untyped> bigint
// line 26: cannot slice variable of type <untyped> bigdec
// line 27: cannot slice variable of type <untyped> bool
// line 47: types cannot be elided in composite literals for struct types
// line 48: types cannot be elided in composite literals for struct types
// line 49: unexpected composite lit type *main.T
// line 50: unexpected composite lit type int

// GoTypeCheckError:
// line 21: cannot slice unaddressable value [3]int{…} (value of type [3]int)
// line 22: cannot slice unaddressable value m[0] (map index expression of type [3]int)
// line 23: cannot slice unaddressable value f() (value of type [3]int)
// line 25: cannot slice 301 (untyped int constant)
// line 26: cannot slice 3.1 (untyped float constant)
// line 27: cannot slice true (untyped bool constant)
// line 47: missing type in composite literal
// line 48: missing type in composite literal
// line 49: invalid composite literal type TP
// line 50: invalid composite literal type Ti

// GnoOverStrictError:
// line 14: function f does not have a body but is not natively defined (did you build after pulling from the repository?)
// line 16: function fp does not have a body but is not natively defined (did you build after pulling from the repository?)
