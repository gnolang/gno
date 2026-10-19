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
// line 36: operator >= not defined on: InterfaceKind
// line 39: operator >= not defined on: InterfaceKind
// line 43: operator >= not defined on: InterfaceKind
// line 46: operator >= not defined on: InterfaceKind
// line 50: operator >= not defined on: InterfaceKind
// line 53: operator >= not defined on: InterfaceKind
// line 56: invalid operation: (mismatched types int and main.I)
// line 57: invalid operation: (mismatched types int and main.I)
// line 58: operator >= not defined on: InterfaceKind
// line 59: invalid operation: (mismatched types int and main.I)
// line 60: invalid operation: (mismatched types int and main.I)
// line 61: operator >= not defined on: InterfaceKind
// line 65: operator >= not defined on: InterfaceKind
// line 68: operator >= not defined on: InterfaceKind
// line 70: invalid operation: (mismatched types <untyped> bigint and main.I)
// line 71: invalid operation: (mismatched types <untyped> bigint and main.I)
// line 72: operator >= not defined on: InterfaceKind
// line 73: invalid operation: (mismatched types <untyped> bigint and main.I)
// line 74: invalid operation: (mismatched types <untyped> bigint and main.I)
// line 75: operator >= not defined on: InterfaceKind
// line 77: func() can only be compared to nil
// line 78: func() can only be compared to nil
// line 79: operator >= not defined on: InterfaceKind
// line 80: func() can only be compared to nil
// line 81: func() can only be compared to nil
// line 82: operator >= not defined on: InterfaceKind
// line 84: func() can only be compared to nil
// line 85: func() can only be compared to nil
// line 86: operator >= not defined on: InterfaceKind
// line 87: func() can only be compared to nil
// line 88: func() can only be compared to nil
// line 89: operator >= not defined on: InterfaceKind
// line 91: main.G can only be compared to nil
// line 92: main.G can only be compared to nil
// line 93: operator >= not defined on: InterfaceKind
// line 94: main.G can only be compared to nil
// line 95: main.G can only be compared to nil
// line 96: operator >= not defined on: InterfaceKind
// line 98: main.G can only be compared to nil
// line 99: main.G can only be compared to nil
// line 100: operator >= not defined on: InterfaceKind
// line 101: main.G can only be compared to nil
// line 102: main.G can only be compared to nil
// line 103: operator >= not defined on: InterfaceKind
// line 105: name not defined: _
// line 106: name not defined: _
// line 107: name not defined: _
// line 108: name not defined: _
// line 109: name not defined: _
// line 110: name not defined: _
// line 112: name not defined: _
// line 113: name not defined: _
// line 114: name not defined: _
// line 115: name not defined: _
// line 116: name not defined: _
// line 117: name not defined: _
// line 119: name not defined: _
// line 121: operator ^ not defined on: InterfaceKind
// line 122: operator ^ not defined on: InterfaceKind
// line 123: operator ^ not defined on: InterfaceKind
// line 124: operator ^ not defined on: InterfaceKind

// GnoWordingDiffers:
// line 36: gc expects: invalid operation.*not defined|invalid comparison|cannot compare; gno: operator >= not defined on: InterfaceKind
// line 39: gc expects: invalid operation.*not defined|invalid comparison|cannot compare; gno: operator >= not defined on: InterfaceKind
// line 43: gc expects: invalid operation.*not defined|invalid comparison|cannot compare; gno: operator >= not defined on: InterfaceKind
// line 46: gc expects: invalid operation.*not defined|invalid comparison|cannot compare; gno: operator >= not defined on: InterfaceKind
// line 50: gc expects: invalid operation.*not defined|invalid comparison|cannot compare; gno: operator >= not defined on: InterfaceKind
// line 53: gc expects: invalid operation.*not defined|invalid comparison|cannot compare; gno: operator >= not defined on: InterfaceKind
// line 58: gc expects: invalid operation.*mismatched types|incompatible types; gno: operator >= not defined on: InterfaceKind
// line 61: gc expects: invalid operation.*mismatched types|incompatible types; gno: operator >= not defined on: InterfaceKind
// line 65: gc expects: invalid operation.*not defined|invalid comparison; gno: operator >= not defined on: InterfaceKind
// line 68: gc expects: invalid operation.*not defined|invalid comparison; gno: operator >= not defined on: InterfaceKind
// line 72: gc expects: invalid operation.*mismatched types|incompatible types|cannot convert; gno: operator >= not defined on: InterfaceKind
// line 75: gc expects: invalid operation.*mismatched types|incompatible types|cannot convert; gno: operator >= not defined on: InterfaceKind
// line 77: gc expects: invalid operation.*not defined|invalid operation; gno: func() can only be compared to nil
// line 78: gc expects: invalid operation.*not defined|invalid operation; gno: func() can only be compared to nil
// line 79: gc expects: invalid operation.*not defined|invalid comparison; gno: operator >= not defined on: InterfaceKind
// line 80: gc expects: invalid operation.*not defined|invalid operation; gno: func() can only be compared to nil
// line 81: gc expects: invalid operation.*not defined|invalid operation; gno: func() can only be compared to nil
// line 82: gc expects: invalid operation.*not defined|invalid comparison; gno: operator >= not defined on: InterfaceKind
// line 84: gc expects: invalid operation.*mismatched types|incompatible types; gno: func() can only be compared to nil
// line 85: gc expects: invalid operation.*mismatched types|incompatible types; gno: func() can only be compared to nil
// line 86: gc expects: invalid operation.*mismatched types|incompatible types; gno: operator >= not defined on: InterfaceKind
// line 87: gc expects: invalid operation.*mismatched types|incompatible types; gno: func() can only be compared to nil
// line 88: gc expects: invalid operation.*mismatched types|incompatible types; gno: func() can only be compared to nil
// line 89: gc expects: invalid operation.*mismatched types|incompatible types; gno: operator >= not defined on: InterfaceKind
// line 91: gc expects: invalid operation.*not defined|invalid operation; gno: main.G can only be compared to nil
// line 92: gc expects: invalid operation.*not defined|invalid operation; gno: main.G can only be compared to nil
// line 93: gc expects: invalid operation.*not defined|invalid comparison; gno: operator >= not defined on: InterfaceKind
// line 94: gc expects: invalid operation.*not defined|invalid operation; gno: main.G can only be compared to nil
// line 95: gc expects: invalid operation.*not defined|invalid operation; gno: main.G can only be compared to nil
// line 96: gc expects: invalid operation.*not defined|invalid comparison; gno: operator >= not defined on: InterfaceKind
// line 98: gc expects: invalid operation.*not defined|invalid operation; gno: main.G can only be compared to nil
// line 99: gc expects: invalid operation.*not defined|invalid operation; gno: main.G can only be compared to nil
// line 100: gc expects: invalid operation.*not defined|invalid comparison; gno: operator >= not defined on: InterfaceKind
// line 101: gc expects: invalid operation.*not defined|invalid operation; gno: main.G can only be compared to nil
// line 102: gc expects: invalid operation.*not defined|invalid operation; gno: main.G can only be compared to nil
// line 103: gc expects: invalid operation.*not defined|invalid comparison; gno: operator >= not defined on: InterfaceKind
// line 105: gc expects: cannot use .*_.* as value; gno: name not defined: _
// line 106: gc expects: cannot use .*_.* as value; gno: name not defined: _
// line 107: gc expects: cannot use .*_.* as value; gno: name not defined: _
// line 108: gc expects: cannot use .*_.* as value; gno: name not defined: _
// line 109: gc expects: cannot use .*_.* as value; gno: name not defined: _
// line 110: gc expects: cannot use .*_.* as value; gno: name not defined: _
// line 112: gc expects: cannot use .*_.* as value; gno: name not defined: _
// line 113: gc expects: cannot use .*_.* as value; gno: name not defined: _
// line 114: gc expects: cannot use .*_.* as value; gno: name not defined: _
// line 115: gc expects: cannot use .*_.* as value; gno: name not defined: _
// line 116: gc expects: cannot use .*_.* as value; gno: name not defined: _
// line 117: gc expects: cannot use .*_.* as value; gno: name not defined: _
// line 119: gc expects: cannot use .*_.* as value; gno: name not defined: _
// line 121: gc expects: invalid operation.*mismatched types|incompatible types; gno: operator ^ not defined on: InterfaceKind
// line 122: gc expects: invalid operation.*mismatched types|incompatible types; gno: operator ^ not defined on: InterfaceKind
// line 123: gc expects: invalid operation.*mismatched types|incompatible types; gno: operator ^ not defined on: InterfaceKind
// line 124: gc expects: invalid operation.*mismatched types|incompatible types; gno: operator ^ not defined on: InterfaceKind

// Verdict: Clean errorcheck
