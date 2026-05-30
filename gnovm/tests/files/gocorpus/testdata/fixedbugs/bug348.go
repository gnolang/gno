// run

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"runtime"
	"strings"
)

func f() {
	var x *string
	
	for _, i := range *x {  // THIS IS LINE 17
		println(i)
	}
}

func g() {
}

func main() {
	defer func() {
		for i := 0;; i++ {
			pc, file, line, ok := runtime.Caller(i)
			if !ok {
				print("BUG: bug348: cannot find caller\n")
				return
			}
			if !strings.Contains(file, "bug348.go") || runtime.FuncForPC(pc).Name() != "main.f" {
				// walk past runtime frames
				continue
			}
			if line != 17 {
				print("BUG: bug348: panic at ", file, ":", line, " in ", runtime.FuncForPC(pc).Name(), "\n")
				return
			}
			recover()
			return
		}
	}()
	f()
}

// TypeCheckError:
// main/bug348.go:28:34: undefined: runtime.Caller; main/bug348.go:33:55: undefined: runtime.FuncForPC; main/bug348.go:38:70: undefined: runtime.FuncForPC

// GnoOutput:

// GoOutput:
// BUG: bug348: cannot find caller
// panic: runtime error: invalid memory address or nil pointer dereference
// [signal SIGSEGV: segmentation violation code=0x2 addr=0x0 pc=0x10079b108]
//
// goroutine 1 [running]:
// main.f(...)
// 	/tmp/claude-501/gno-filetest-go-2717917391/main.go:17
// main.main()
// 	/tmp/claude-501/gno-filetest-go-2717917391/main.go:45 +0x38
// exit status 2

// Divergence: TODO: <category>: explain why this divergence is acceptable

// Unsupported: non-deterministic runtime output (signal SIGSEGV)
