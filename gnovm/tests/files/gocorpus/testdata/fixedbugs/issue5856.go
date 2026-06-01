// run

// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

func main() {
	f()
	panic("deferred function not run")
}

var x = 1

func f() {
	if x == 0 {
		return
	}
	defer g()
	panic("panic")
}

func g() {
	_, file, line, _ := runtime.Caller(2)
	if !strings.HasSuffix(file, "issue5856.go") || line != 28 {
		fmt.Printf("BUG: defer called from %s:%d, want issue5856.go:28\n", file, line)
		os.Exit(1)
	}
	os.Exit(0)
}

// TypeCheckError:
// main/issue5856.go:32:30: undefined: runtime.Caller; main/issue5856.go:35:6: undefined: os.Exit; main/issue5856.go:37:5: undefined: os.Exit

// GnoOutput:

// GoOutput:
// BUG: defer called from /tmp/claude-501/gno-filetest-go-3656518913/main.go:28, want issue5856.go:28
// exit status 1

// KnownDivergence: TODO: <category>: explain why this divergence is acceptable

// Unsupported: non-deterministic runtime output (gno-filetest-go-)
