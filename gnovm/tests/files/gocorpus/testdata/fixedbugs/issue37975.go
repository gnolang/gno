// run

// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Make sure runtime.panicmakeslice* are called.

package main

import "strings"

func main() {
	// Test typechecking passes if len is valid
	// but cap is out of range for len's type.
	var x byte
	_ = make([]int, x, 300)

	capOutOfRange := func() {
		i := 2
		s := make([]int, i, 1)
		s[0] = 1
	}
	lenOutOfRange := func() {
		i := -1
		s := make([]int, i, 3)
		s[0] = 1
	}

	tests := []struct {
		f        func()
		panicStr string
	}{
		{capOutOfRange, "cap out of range"},
		{lenOutOfRange, "len out of range"},
	}

	for _, tc := range tests {
		shouldPanic(tc.panicStr, tc.f)
	}

}

func shouldPanic(str string, f func()) {
	defer func() {
		err := recover()
		runtimeErr := err.(error).Error()
		if !strings.Contains(runtimeErr, str) {
			panic("got panic " + runtimeErr + ", want " + str)
		}
	}()

	f()
}


// Fixing: PR #5732 (fix/5667, typedRuntimeError); verified clean on branch, broken on master; re-golden after merge.

// GnoOutput:

// GnoError:
// runtime error: makeslice: cap out of range
// 	string doesn't implement interface {Error func() string} (missing method Error)

// GoOutput:

// KnownIssue:
// The makeslice panic messages already match Go, but the panic value is a
// bare string, so recover().(error) fails ("string doesn't implement
// interface"). Same root cause as fixedbugs/issue19040.go.
