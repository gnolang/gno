// run

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Check the text of the panic that comes from
// a nil pointer passed to automatically generated method wrapper.

package main

import "fmt"

type T int

type I interface {
	F()
}

func (t T) F() {}

var (
	t *T
	i I = t
)

func main() {
	defer func() {
		got := recover().(error).Error()
		want := "value method main.T.F called using nil *T pointer"
		if got != want {
			fmt.Printf("panicwrap error text:\n\t%q\nwant:\n\t%q\n", got, want)
		}
	}()
	i.F()
}


// Fixing: PR #5732 (fix/5667, typedRuntimeError); verified clean on branch, broken on master; tracks issue #5667; re-golden after merge.

// GnoOutput:

// GnoError:
// runtime error: nil pointer dereference
// 	string doesn't implement interface {Error func() string} (missing method Error)

// GoOutput:

// KnownIssue:
// Runtime panics carry a bare string, so recover().(error) fails — Go
// runtime panics implement error (runtime.Error). The nil-receiver wrapper
// panic text also differs from gc's.
