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
// line 79: operator >= not defined on: InterfaceKind
// line 82: operator >= not defined on: InterfaceKind
// line 84: invalid operation: (mismatched types func() and main.I)
// line 85: invalid operation: (mismatched types func() and main.I)
// line 86: operator >= not defined on: InterfaceKind
// line 87: invalid operation: (mismatched types func() and main.I)
// line 88: invalid operation: (mismatched types func() and main.I)
// line 89: operator >= not defined on: InterfaceKind
// line 93: operator >= not defined on: InterfaceKind
// line 96: operator >= not defined on: InterfaceKind
// line 100: operator >= not defined on: InterfaceKind
// line 103: operator >= not defined on: InterfaceKind
// line 105: name _ not defined in fileset with files [issue9370.go]
// line 106: name _ not defined in fileset with files [issue9370.go]
// line 107: name _ not defined in fileset with files [issue9370.go]
// line 108: name _ not defined in fileset with files [issue9370.go]
// line 109: name _ not defined in fileset with files [issue9370.go]
// line 110: name _ not defined in fileset with files [issue9370.go]
// line 112: name _ not defined in fileset with files [issue9370.go]
// line 113: name _ not defined in fileset with files [issue9370.go]
// line 114: name _ not defined in fileset with files [issue9370.go]
// line 115: name _ not defined in fileset with files [issue9370.go]
// line 116: name _ not defined in fileset with files [issue9370.go]
// line 117: name _ not defined in fileset with files [issue9370.go]
// line 119: name _ not defined in fileset with files [issue9370.go]
// line 121: operator ^ not defined on: InterfaceKind
// line 122: operator ^ not defined on: InterfaceKind
// line 123: operator ^ not defined on: InterfaceKind
// line 124: operator ^ not defined on: InterfaceKind

// GoTypeCheckError:
// line 36
// line 39
// line 43
// line 46
// line 50
// line 53
// line 56
// line 57
// line 58
// line 59
// line 60
// line 61
// line 65
// line 68
// line 70
// line 71
// line 72
// line 73
// line 74
// line 75
// line 77
// line 78
// line 79
// line 80
// line 81
// line 82
// line 84
// line 85
// line 86
// line 87
// line 88
// line 89
// line 91
// line 92
// line 93
// line 94
// line 95
// line 96
// line 98
// line 99
// line 100
// line 101
// line 102
// line 103
// line 105
// line 106
// line 107
// line 108
// line 109
// line 110
// line 112
// line 113
// line 114
// line 115
// line 116
// line 117
// line 119
// line 121
// line 122
// line 123
// line 124

// GoTypesOnly:
// line 77: gc expects: invalid operation.*not defined|invalid operation
// line 78: gc expects: invalid operation.*not defined|invalid operation
// line 80: gc expects: invalid operation.*not defined|invalid operation
// line 81: gc expects: invalid operation.*not defined|invalid operation
// line 91: gc expects: invalid operation.*not defined|invalid operation
// line 92: gc expects: invalid operation.*not defined|invalid operation
// line 94: gc expects: invalid operation.*not defined|invalid operation
// line 95: gc expects: invalid operation.*not defined|invalid operation
// line 98: gc expects: invalid operation.*not defined|invalid operation
// line 99: gc expects: invalid operation.*not defined|invalid operation
// line 101: gc expects: invalid operation.*not defined|invalid operation
// line 102: gc expects: invalid operation.*not defined|invalid operation

// Verdict: Clean errorcheck
