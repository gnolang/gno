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

// GnoStaticIncomplete: covered 1 of 3 markers (Gno preprocess: 1, go/types guard: 1); Gno bailed before the rest — a runnable variant may exercise more

// GnoError:
// line 12: missing parameter name (and 2 more errors)

// GoTypeCheckError:
// line 12: missing parameter name (and 2 more errors)
