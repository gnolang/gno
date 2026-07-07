// errorcheck

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Verify that type switch statements with duplicate cases are detected
// by the compiler.
// Does not compile.

package main

import "fmt"

func f4(e interface{}) {
	switch e.(type) {
	case int:
	case int: // ERROR "duplicate case int in type switch"
	case int64:
	case error:
	case error: // ERROR "duplicate case error in type switch"
	case fmt.Stringer:
	case fmt.Stringer: // ERROR "duplicate case fmt.Stringer in type switch"
	case struct {
		i int "tag1"
	}:
	case struct {
		i int "tag2"
	}:
	case struct { // ERROR "duplicate case struct { i int .tag1. } in type switch|duplicate case"
		i int "tag1"
	}:
	}
}

// GnoError:
// line 16: 3: duplicate type int in type switch
// line 17: expected '}', found 'case'
// line 18: expected '}', found 'case'
// line 19: expected '}', found 'case'
// line 20: expected '}', found 'case'
// line 21: expected '}', found 'case'
// line 22: expected '}', found 'case'
// line 23: expected '}', found 'case'
// line 24: expected '}', found 'case'
// line 25: expected ';', found int (and 1 more errors)
// line 26: expected ';', found ':'
// line 27: expected '}', found 'case'
// line 28: expected ';', found int (and 1 more errors)
// line 29: expected ';', found ':'
// line 30: expected '}', found 'case'
// line 31: expected ';', found int (and 1 more errors)

// GoTypeCheckError:
// line 18: duplicate case int in type switch
// line 21: duplicate case error in type switch
// line 23: duplicate case fmt.Stringer in type switch
// line 30: duplicate case struct{i int "tag1"} in type switch

// GnoOverStrictError:
// line 16: 3: duplicate type int in type switch
// line 17: expected '}', found 'case'
// line 19: expected '}', found 'case'
// line 20: expected '}', found 'case'
// line 22: expected '}', found 'case'
// line 24: expected '}', found 'case'
// line 25: expected ';', found int (and 1 more errors)
// line 26: expected ';', found ':'
// line 27: expected '}', found 'case'
// line 28: expected ';', found int (and 1 more errors)
// line 29: expected ';', found ':'
// line 31: expected ';', found int (and 1 more errors)
