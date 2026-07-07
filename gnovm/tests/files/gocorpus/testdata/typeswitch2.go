// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Verify that various erroneous type switches are caught by the compiler.
// Does not compile.

package main

import "io"

func whatis(x interface{}) string {
	switch x.(type) {
	case int:
		return "int"
	case int: // ERROR "duplicate"
		return "int8"
	case io.Reader:
		return "Reader1"
	case io.Reader: // ERROR "duplicate"
		return "Reader2"
	case interface {
		r()
		w()
	}:
		return "rw"
	case interface {	// ERROR "duplicate"
		w()
		r()
	}:
		return "wr"

	}
	return ""
}

// GnoError:
// line 15: 3: duplicate type int in type switch
// line 16: expected '}', found 'case' (and 1 more errors)
// line 18: expected '}', found 'case' (and 1 more errors)
// line 20: expected '}', found 'case' (and 1 more errors)
// line 22: expected '}', found 'case' (and 1 more errors)
// line 24: expected '}', found 'case' (and 1 more errors)
// line 25: name r not declared
// line 26: name w not declared
// line 27: expected ';', found ':' (and 1 more errors)
// line 29: expected '}', found 'case' (and 1 more errors)
// line 32: expected ';', found ':' (and 1 more errors)
// line 36: expected declaration, found 'return'
// line 37: expected declaration, found '}'

// GoTypeCheckError:
// line 18: duplicate case int in type switch
// line 22: duplicate case io.Reader in type switch
// line 29: duplicate case interface{r()

// GnoOverStrictError:
// line 15: 3: duplicate type int in type switch
// line 16: expected '}', found 'case' (and 1 more errors)
// line 20: expected '}', found 'case' (and 1 more errors)
// line 24: expected '}', found 'case' (and 1 more errors)
// line 25: name r not declared
// line 26: name w not declared
// line 27: expected ';', found ':' (and 1 more errors)
// line 32: expected ';', found ':' (and 1 more errors)
// line 36: expected declaration, found 'return'
// line 37: expected declaration, found '}'
