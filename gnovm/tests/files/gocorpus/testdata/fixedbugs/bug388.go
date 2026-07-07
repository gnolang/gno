// errorcheck

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 2231

package main
import "runtime"

func foo(runtime.UintType, i int) {  // ERROR "cannot declare name runtime.UintType|missing parameter name|undefined identifier"
	println(i, runtime.UintType) // GCCGO_ERROR "undefined identifier"
}

func qux() {
	var main.i	// ERROR "unexpected [.]|expected type"
	println(main.i)
}

func corge() {
	var foo.i int  // ERROR "unexpected [.]|expected type"
	println(foo.i)
}

func main() {
	foo(42,43)
	bar(1969)
}

// GnoError:
// line 12: missing parameter name (and 2 more errors)
// line 13: expected declaration, found println (and 4 more errors)
// line 14: expected declaration, found '}' (and 4 more errors)
// line 17: expected type, found '.' (and 1 more errors)
// line 18: unexpected selector expression type *gnolang.FuncType
// line 22: expected type, found '.'
// line 28: name bar not declared

// GoTypeCheckError:
// line 12: missing parameter name (and 2 more errors)
// line 17: expected type, found '.' (and 1 more errors)
// line 22: expected type, found '.'

// GnoOverStrictError:
// line 13: expected declaration, found println (and 4 more errors)
// line 14: expected declaration, found '}' (and 4 more errors)
// line 18: unexpected selector expression type *gnolang.FuncType
// line 28: name bar not declared
