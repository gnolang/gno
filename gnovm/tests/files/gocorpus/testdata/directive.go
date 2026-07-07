// errorcheck

// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Verify that misplaced directives are diagnosed.

//go:noinline // ERROR "misplaced compiler directive"

//go:noinline // ERROR "misplaced compiler directive"
package main

//go:nosplit
func f1() {}

//go:nosplit
//go:noinline
func f2() {}

//go:noinline // ERROR "misplaced compiler directive"

//go:noinline // ERROR "misplaced compiler directive"
var x int

//go:noinline // ERROR "misplaced compiler directive"
const c = 1

//go:noinline // ERROR "misplaced compiler directive"
type T int

type (
	//go:noinline // ERROR "misplaced compiler directive"
	T2 int
	//go:noinline // ERROR "misplaced compiler directive"
	T3 int
)

//go:noinline
func f() {
	x := 1

	{
		_ = x
	}
	//go:noinline // ERROR "misplaced compiler directive"
	var y int
	_ = y

	//go:noinline // ERROR "misplaced compiler directive"
	const c = 1

	_ = func() {}

	//go:noinline // ERROR "misplaced compiler directive"
	type T int
}

// Error:
// main:0:0: name main not declared

// GnoOutput:

// GnoError:
// main:0:0: name main not declared

// GoOutput:
// # gnofiletest
// ./main.go:9:3: misplaced compiler directive
// ./main.go:11:3: misplaced compiler directive
// ./main.go:21:3: misplaced compiler directive
// ./main.go:23:3: misplaced compiler directive
// ./main.go:26:3: misplaced compiler directive
// ./main.go:29:3: misplaced compiler directive
// ./main.go:33:4: misplaced compiler directive
// ./main.go:35:4: misplaced compiler directive
// ./main.go:46:4: misplaced compiler directive
// ./main.go:50:4: misplaced compiler directive
// ./main.go:50:4: too many errors

// KnownIssue:
// TODO: explain the Gno bug (Gno errors where Go runs clean)
