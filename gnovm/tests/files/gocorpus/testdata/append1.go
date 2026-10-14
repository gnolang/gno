// errorcheck

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Verify that append arguments requirements are enforced by the
// compiler.

package main

func main() {

	s := make([]int, 8)

	_ = append()           // ERROR "missing arguments to append|not enough arguments for append"
	_ = append(s...)       // ERROR "cannot use ... on first argument|not enough arguments in call to append"
	_ = append(s, 2, s...) // ERROR "too many arguments to append|too many arguments in call to append"

	_ = append(s, make([]int, 0))     // ERROR "cannot use make\(\[\]int, 0\) \(value of type \[\]int\) as int value in argument to append"
	_ = append(s, make([]int, -1)...) // ERROR "negative len argument in make|index -1.* must not be negative"
}

// GnoError:
// line 17: not enough arguments in call to (const (append func(<X>{}, ...<X.Elem()>{}) <X>{})); want 2 (including variadic) got 1
// line 18: not enough arguments in call to (const (append func(<X>{}, ...<X.Elem()>{}) <X>{})); want 2 (including variadic) got 3
// line 20: cannot use []int as int
// line 21: invalid argument: index (-1 <untyped> bigint) must not be negative

// GoTypeCheckError:
// line 16
// line 17
// line 18
// line 20
// line 21

// GoTypesOnly:
// line 16: gc expects: missing arguments to append|not enough arguments for append

// GnoWordingDiffers:
// line 17: gc expects: cannot use ... on first argument|not enough arguments in call to append; gno: not enough arguments in call to (const (append func(<X>{}, ...<X.Elem()>{}) <X>{})); want 2 (including variadic) got 1
// line 18: gc expects: too many arguments to append|too many arguments in call to append; gno: not enough arguments in call to (const (append func(<X>{}, ...<X.Elem()>{}) <X>{})); want 2 (including variadic) got 3
// line 20: gc expects: cannot use make\(\[\]int, 0\) \(value of type \[\]int\) as int value in argument to append; gno: cannot use []int as int
// line 21: gc expects: negative len argument in make|index -1.* must not be negative; gno: invalid argument: index (-1 <untyped> bigint) must not be negative

// Verdict: Clean errorcheck
