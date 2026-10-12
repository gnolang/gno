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

// GoTypeCheckError:
// line 17
// line 18
// line 20
// line 21
// line 22
// line 23
// line 24
// line 30

// Verdict: Clean errorcheck
