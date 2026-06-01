// run

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test that an interface conversion error panics with an "interface
// conversion" run-time error. It was (incorrectly) panicking with a
// "nil pointer dereference."

package main

import (
	"fmt"
	"runtime"
	"strings"
)

type I interface {
	Get() int
}

func main() {
	defer func() {
		r := recover()
		if r == nil {
			panic("expected panic")
		}
		re, ok := r.(runtime.Error)
		if !ok {
			panic(fmt.Sprintf("got %T, expected runtime.Error", r))
		}
		if !strings.Contains(re.Error(), "interface conversion") {
			panic(fmt.Sprintf("got %q, expected interface conversion error", re.Error()))
		}
	}()
	e := (interface{})(0)
	if _, ok := e.(I); ok {
		panic("unexpected interface conversion success")
	}
	fmt.Println(e.(I))
	panic("unexpected interface conversion success")
}

// TypeCheckError:
// main/issue16130.go:29:24: undefined: runtime.Error

// GnoOutput:

// GnoError:
// main/issue16130.go:29:16-29: name Error not declared

// GoOutput:

// KnownIssue:
// TODO: explain the Gno bug (Gno errors where Go runs clean)
