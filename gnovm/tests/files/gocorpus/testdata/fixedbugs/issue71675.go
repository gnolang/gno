// run
// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package main

//go:noinline
func i() {
	for range yieldInts {
		defer func() {
			println("I")
			recover()
		}()
	}
	// This panic causes dead code elimination of the return block.
	// The compiler should nonetheless emit a deferreturn.
	panic("i panic")
}

//go:noinline
func h() {
	defer func() {
		println("H first")
	}()
	for range yieldInts {
		defer func() {
			println("H second")
		}()
	}
	defer func() {
		println("H third")
	}()
	for range yieldIntsPanic {
		defer func() {
			println("h recover:called")
			recover()
		}()
	}
}

//go:noinline
func yieldInts(yield func(int) bool) {
	if !yield(0) {
		return
	}
}

//go:noinline
func g() {
	defer func() {
		println("G first")
	}()
	for range yieldIntsPanic {
		defer func() {
			println("g recover:called")
			recover()
		}()
	}
}

//go:noinline
func yieldIntsPanic(yield func(int) bool) {
	if !yield(0) {
		return
	}
	panic("yield stop")
}

//go:noinline
func next(i int) int {
	if i == 0 {
		panic("next stop")
	}
	return i + 1
}

//go:noinline
func f() {
	defer func() {
		println("F first")
	}()
	for i := 0; i < 1; i = next(i) {
		defer func() {
			println("f recover:called")
			recover()
		}()
	}
}
func main() {
	f()
	println("f returned")
	g()
	println("g returned")
	h()
	println("h returned")
	i()
	println("i returned")

}

// KnownDivergence:
// Go1.17 pinned.

// GnoOutput:

// GnoError:
// main/issue71675.go:9:2-14:3: range iteration requires map, string, array, slice, or pointer to array; got FuncKind

// GoOutput:
// f recover:called
// F first
// f returned
// g recover:called
// G first
// g returned
// h recover:called
// H third
// H second
// H first
// h returned
// I
// i returned
