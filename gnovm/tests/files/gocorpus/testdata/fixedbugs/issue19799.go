// run

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"os"
	"runtime"
)

func foo(x int) int {
	return x + 1
}

func test() {
	defer func() {
		if r := recover(); r != nil {
			pcs := make([]uintptr, 10)
			n := runtime.Callers(0, pcs)
			pcs = pcs[:n]
			frames := runtime.CallersFrames(pcs)
			for {
				f, more := frames.Next()
				if f.Function == "main.foo" {
					println("did not expect to see call to foo in stack trace")
					os.Exit(1)
				}
				if !more {
					break
				}
			}
		}
	}()
	var v []int
	foo(v[0])
}

func bar(x ...int) int {
	return x[0] + 1
}

func testVariadic() {
	defer func() {
		if r := recover(); r != nil {
			pcs := make([]uintptr, 10)
			n := runtime.Callers(0, pcs)
			pcs = pcs[:n]
			frames := runtime.CallersFrames(pcs)
			for {
				f, more := frames.Next()
				if f.Function == "main.bar" {
					println("did not expect to see call to bar in stack trace")
					os.Exit(1)
				}
				if !more {
					break
				}
			}
		}
	}()
	var v []int
	bar(v[0])
}

func main() {
	test()
	testVariadic()
}

// TypeCheckError:
// main/issue19799.go:22:17: undefined: runtime.Callers; main/issue19799.go:24:22: undefined: runtime.CallersFrames; main/issue19799.go:29:9: undefined: os.Exit; main/issue19799.go:49:17: undefined: runtime.Callers; main/issue19799.go:51:22: undefined: runtime.CallersFrames; main/issue19799.go:56:9: undefined: os.Exit

// GnoOutput:

// GnoError:
// main/issue19799.go:21:18-25: name uintptr not declared

// GoOutput:

// Unsupported: uintptr type not supported in Gno
