// errorcheck

// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Verify that concrete/interface comparisons are
// typechecked correctly by the compiler.

package main

type I interface {
	Method()
}

type C int

func (C) Method() {}

type G func()

func (G) Method() {}

var (
	e interface{}
	i I
	c C
	n int
	f func()
	g G
)

var (
	_ = e == c
	_ = e != c
	_ = e >= c // ERROR "invalid operation.*not defined|invalid comparison|cannot compare"
	_ = c == e
	_ = c != e
	_ = c >= e // ERROR "invalid operation.*not defined|invalid comparison|cannot compare"

	_ = i == c
	_ = i != c
	_ = i >= c // ERROR "invalid operation.*not defined|invalid comparison|cannot compare"
	_ = c == i
	_ = c != i
	_ = c >= i // ERROR "invalid operation.*not defined|invalid comparison|cannot compare"

	_ = e == n
	_ = e != n
	_ = e >= n // ERROR "invalid operation.*not defined|invalid comparison|cannot compare"
	_ = n == e
	_ = n != e
	_ = n >= e // ERROR "invalid operation.*not defined|invalid comparison|cannot compare"

	// i and n are not assignable to each other
	_ = i == n // ERROR "invalid operation.*mismatched types|incompatible types"
	_ = i != n // ERROR "invalid operation.*mismatched types|incompatible types"
	_ = i >= n // ERROR "invalid operation.*mismatched types|incompatible types"
	_ = n == i // ERROR "invalid operation.*mismatched types|incompatible types"
	_ = n != i // ERROR "invalid operation.*mismatched types|incompatible types"
	_ = n >= i // ERROR "invalid operation.*mismatched types|incompatible types"

	_ = e == 1
	_ = e != 1
	_ = e >= 1 // ERROR "invalid operation.*not defined|invalid comparison"
	_ = 1 == e
	_ = 1 != e
	_ = 1 >= e // ERROR "invalid operation.*not defined|invalid comparison"

	_ = i == 1 // ERROR "invalid operation.*mismatched types|incompatible types|cannot convert"
	_ = i != 1 // ERROR "invalid operation.*mismatched types|incompatible types|cannot convert"
	_ = i >= 1 // ERROR "invalid operation.*mismatched types|incompatible types|cannot convert"
	_ = 1 == i // ERROR "invalid operation.*mismatched types|incompatible types|cannot convert"
	_ = 1 != i // ERROR "invalid operation.*mismatched types|incompatible types|cannot convert"
	_ = 1 >= i // ERROR "invalid operation.*mismatched types|incompatible types|cannot convert"

	_ = e == f // ERROR "invalid operation.*not defined|invalid operation"
	_ = e != f // ERROR "invalid operation.*not defined|invalid operation"
	_ = e >= f // ERROR "invalid operation.*not defined|invalid comparison"
	_ = f == e // ERROR "invalid operation.*not defined|invalid operation"
	_ = f != e // ERROR "invalid operation.*not defined|invalid operation"
	_ = f >= e // ERROR "invalid operation.*not defined|invalid comparison"

	_ = i == f // ERROR "invalid operation.*mismatched types|incompatible types"
	_ = i != f // ERROR "invalid operation.*mismatched types|incompatible types"
	_ = i >= f // ERROR "invalid operation.*mismatched types|incompatible types"
	_ = f == i // ERROR "invalid operation.*mismatched types|incompatible types"
	_ = f != i // ERROR "invalid operation.*mismatched types|incompatible types"
	_ = f >= i // ERROR "invalid operation.*mismatched types|incompatible types"

	_ = e == g // ERROR "invalid operation.*not defined|invalid operation"
	_ = e != g // ERROR "invalid operation.*not defined|invalid operation"
	_ = e >= g // ERROR "invalid operation.*not defined|invalid comparison"
	_ = g == e // ERROR "invalid operation.*not defined|invalid operation"
	_ = g != e // ERROR "invalid operation.*not defined|invalid operation"
	_ = g >= e // ERROR "invalid operation.*not defined|invalid comparison"

	_ = i == g // ERROR "invalid operation.*not defined|invalid operation"
	_ = i != g // ERROR "invalid operation.*not defined|invalid operation"
	_ = i >= g // ERROR "invalid operation.*not defined|invalid comparison"
	_ = g == i // ERROR "invalid operation.*not defined|invalid operation"
	_ = g != i // ERROR "invalid operation.*not defined|invalid operation"
	_ = g >= i // ERROR "invalid operation.*not defined|invalid comparison"

	_ = _ == e // ERROR "cannot use .*_.* as value"
	_ = _ == i // ERROR "cannot use .*_.* as value"
	_ = _ == c // ERROR "cannot use .*_.* as value"
	_ = _ == n // ERROR "cannot use .*_.* as value"
	_ = _ == f // ERROR "cannot use .*_.* as value"
	_ = _ == g // ERROR "cannot use .*_.* as value"

	_ = e == _ // ERROR "cannot use .*_.* as value"
	_ = i == _ // ERROR "cannot use .*_.* as value"
	_ = c == _ // ERROR "cannot use .*_.* as value"
	_ = n == _ // ERROR "cannot use .*_.* as value"
	_ = f == _ // ERROR "cannot use .*_.* as value"
	_ = g == _ // ERROR "cannot use .*_.* as value"

	_ = _ == _ // ERROR "cannot use .*_.* as value"

	_ = e ^ c // ERROR "invalid operation.*mismatched types|incompatible types"
	_ = c ^ e // ERROR "invalid operation.*mismatched types|incompatible types"
	_ = 1 ^ e // ERROR "invalid operation.*mismatched types|incompatible types"
	_ = e ^ 1 // ERROR "invalid operation.*mismatched types|incompatible types"
	_ = 1 ^ c
	_ = c ^ 1
)

// GnoError:
// line 36: invalid operation: e >= c (operator >= not defined on interface)
// line 39: invalid operation: c >= e (operator >= not defined on interface)
// line 43: invalid operation: i >= c (operator >= not defined on interface)
// line 46: invalid operation: c >= i (operator >= not defined on interface)
// line 50: invalid operation: e >= n (operator >= not defined on interface)
// line 53: invalid operation: n >= e (operator >= not defined on interface)
// line 56: invalid operation: (mismatched types int and main.I)
// line 57: invalid operation: (mismatched types int and main.I)
// line 58: invalid operation: i >= n (mismatched types I and int)
// line 59: invalid operation: (mismatched types int and main.I)
// line 60: invalid operation: (mismatched types int and main.I)
// line 61: invalid operation: n >= i (mismatched types int and I)
// line 65: invalid operation: e >= 1 (operator >= not defined on interface)
// line 68: invalid operation: 1 >= e (operator >= not defined on interface)
// line 70: invalid operation: (mismatched types <untyped> bigint and main.I)
// line 71: invalid operation: (mismatched types <untyped> bigint and main.I)
// line 72: cannot convert 1 (untyped int constant) to type interface{Method()}
// line 73: invalid operation: (mismatched types <untyped> bigint and main.I)
// line 74: invalid operation: (mismatched types <untyped> bigint and main.I)
// line 75: cannot convert 1 (untyped int constant) to type interface{Method()}
// line 77: invalid operation: e == f (func can only be compared to nil)
// line 78: invalid operation: e != f (func can only be compared to nil)
// line 79: invalid operation: e >= f (operator >= not defined on interface)
// line 80: invalid operation: f == e (func can only be compared to nil)
// line 81: invalid operation: f != e (func can only be compared to nil)
// line 82: invalid operation: f >= e (operator >= not defined on func)
// line 84: invalid operation: (mismatched types func() and main.I)
// line 85: invalid operation: (mismatched types func() and main.I)
// line 86: invalid operation: i >= f (mismatched types I and func())
// line 87: invalid operation: (mismatched types func() and main.I)
// line 88: invalid operation: (mismatched types func() and main.I)
// line 89: invalid operation: f >= i (mismatched types func() and I)
// line 91: invalid operation: e == g (func can only be compared to nil)
// line 92: invalid operation: e != g (func can only be compared to nil)
// line 93: invalid operation: e >= g (operator >= not defined on interface)
// line 94: invalid operation: g == e (func can only be compared to nil)
// line 95: invalid operation: g != e (func can only be compared to nil)
// line 96: invalid operation: g >= e (operator >= not defined on func)
// line 98: invalid operation: i == g (func can only be compared to nil)
// line 99: invalid operation: i != g (func can only be compared to nil)
// line 100: invalid operation: i >= g (operator >= not defined on interface)
// line 101: invalid operation: g == i (func can only be compared to nil)
// line 102: invalid operation: g != i (func can only be compared to nil)
// line 103: invalid operation: g >= i (operator >= not defined on func)
// line 105: cannot use _ as value or type
// line 106: cannot use _ as value or type
// line 107: cannot use _ as value or type
// line 108: cannot use _ as value or type
// line 109: cannot use _ as value or type
// line 110: cannot use _ as value or type
// line 112: cannot use _ as value or type
// line 113: cannot use _ as value or type
// line 114: cannot use _ as value or type
// line 115: cannot use _ as value or type
// line 116: cannot use _ as value or type
// line 117: cannot use _ as value or type
// line 119: cannot use _ as value or type
// line 121: invalid operation: e ^ c (mismatched types interface{} and C)
// line 122: invalid operation: c ^ e (mismatched types C and interface{})
// line 123: invalid operation: 1 ^ e (mismatched types int and interface{})
// line 124: invalid operation: e ^ 1 (mismatched types interface{} and int)
