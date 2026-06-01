// run

// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"runtime"
	"strings"
)

type T struct {
	val int
}

func main() {
	defer expectError(22)
	var pT *T
	switch pT.val { // error should be here - line 22
	case 0:
		fmt.Println("0")
	case 1: // used to show up here instead
		fmt.Println("1")
	case 2:
		fmt.Println("2")
	}
	fmt.Println("finished")
}

func expectError(expectLine int) {
	if recover() == nil {
		panic("did not crash")
	}
	for i := 1;; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			panic("cannot find issue4562.go on stack")
		}
		if strings.HasSuffix(file, "issue4562.go") {
			if line != expectLine {
				panic(fmt.Sprintf("crashed at line %d, wanted line %d", line, expectLine))
			}
			break
		}
	}
}

// TypeCheckError:
// main/issue4562.go:38:32: undefined: runtime.Caller

// GnoOutput:

// GoOutput:
// panic: runtime error: invalid memory address or nil pointer dereference [recovered]
// 	panic: cannot find issue4562.go on stack
// [signal SIGSEGV: segmentation violation code=0x2 addr=0x0 pc=0x1023d71ec]
//
// goroutine 1 [running]:
// main.expectError(0x16)
// 	/tmp/claude-501/gno-filetest-go-2917163442/main.go:40 +0x108
// panic({0x102412d20?, 0x1024b50e0?})
// 	/Users/maxwell/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.25.9.darwin-arm64/src/runtime/panic.go:783 +0x120
// main.main()
// 	/tmp/claude-501/gno-filetest-go-2917163442/main.go:22 +0x3c
// exit status 2

// KnownDivergence: TODO: <category>: explain why this divergence is acceptable

// Unsupported: non-deterministic runtime output (signal SIGSEGV)
