// run

// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"runtime"
	"strings"
)

func main() {
	f(nil)
}

func f(p *int32) {
	defer checkstack()
	v := *p         // panic should happen here, line 20
	sink = int64(v) // not here, line 21
}

var sink int64

func checkstack() {
	_ = recover()
	var buf [1024]byte
	n := runtime.Stack(buf[:], false)
	s := string(buf[:n])
	if strings.Contains(s, "issue27201.go:21 ") {
		panic("panic at wrong location")
	}
	if !strings.Contains(s, "issue27201.go:20 ") {
		panic("no panic at correct location")
	}
}

// TypeCheckError:
// main/issue27201.go:29:15: undefined: runtime.Stack

// GnoOutput:

// GoOutput:
// panic: runtime error: invalid memory address or nil pointer dereference [recovered]
// 	panic: no panic at correct location
// [signal SIGSEGV: segmentation violation code=0x2 addr=0x0 pc=0x1008a6ce4]
//
// goroutine 1 [running]:
// main.checkstack()
// 	/tmp/claude-501/gno-filetest-go-4285626519/main.go:35 +0xbc
// panic({0x1008d06c0?, 0x100939440?})
// 	/Users/maxwell/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.25.9.darwin-arm64/src/runtime/panic.go:783 +0x120
// main.f(0x1009382b8?)
// 	/tmp/claude-501/gno-filetest-go-4285626519/main.go:20 +0x34
// main.main()
// 	/tmp/claude-501/gno-filetest-go-4285626519/main.go:15 +0x20
// exit status 2

// Divergence: TODO: <category>: explain why this divergence is acceptable

// Unsupported: non-deterministic runtime output (signal SIGSEGV)
