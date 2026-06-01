// run

// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"runtime"
	"testing"
)

type T struct { int }

var globl *T

func F() {
	t := &T{}
	globl = t
}

func G() {
	t := &T{}
	_ = t
}

func main() {
	nf := testing.AllocsPerRun(100, F)
	ng := testing.AllocsPerRun(100, G)
	if int(nf) > 1 {
		fmt.Printf("AllocsPerRun(100, F) = %v, want 1\n", nf)
		os.Exit(1)
	}
	if int(ng) != 0 && (runtime.Compiler != "gccgo" || int(ng) != 1) {
		fmt.Printf("AllocsPerRun(100, G) = %v, want 0\n", ng)
		os.Exit(1)
	}
}

// TypeCheckError:
// main/issue4618.go:31:16: undefined: testing.AllocsPerRun; main/issue4618.go:32:16: undefined: testing.AllocsPerRun; main/issue4618.go:35:6: undefined: os.Exit; main/issue4618.go:37:30: undefined: runtime.Compiler; main/issue4618.go:39:6: undefined: os.Exit

// GnoOutput:

// GnoError:
// main/issue4618.go:31:8-28: name AllocsPerRun not declared

// GoOutput:

// KnownIssue:
// TODO: explain the Gno bug (Gno errors where Go runs clean)
