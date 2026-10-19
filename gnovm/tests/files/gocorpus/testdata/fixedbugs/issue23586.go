// errorcheck

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test that we type-check deferred/go functions even
// if they are not called (a common error). Specifically,
// we don't want to see errors such as import or variable
// declared and not used.

package p

import (
	"fmt"
	"math"
)

func f() {
	var i int
	defer func() { fmt.Println() } // ERROR "must be function call"
	go func() { _ = math.Sin(0) }  // ERROR "must be function call"
	go func() { _ = i}             // ERROR "must be function call"
}

// GnoError:
// line 15: "fmt" imported and not used
// line 16: "math" imported and not used
// line 20: declared and not used: i
// line 21: expression in defer must be function call
// line 22: expression in go must be function call
// line 23: expression in go must be function call

// Verdict: KnownIssue errorcheck: #6344 over-strict: lines 15, 16, 20
