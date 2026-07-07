// run

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "strings"

var X interface{}

type T struct{}

func scopes() {
	p, ok := recover().(error)
	if ok && strings.Contains(p.Error(), "different scopes") {
		return
	}
	panic(p)
}

func F1() {
	type T struct{}
	X = T{}
}

func F2() {
	type T struct{}
	defer scopes()
	_ = X.(T)
}

func F3() {
	defer scopes()
	_ = X.(T)
}

func F4() {
	X = T{}
}

func main() {
	F1() // set X to F1's T
	F2() // check that X is not F2's T
	F3() // check that X is not package T
	F4() // set X to package T
	F2() // check that X is not F2's T
}


// Fixing: PR #5732 (fix/5667, typedRuntimeError); verified on branch (wording gap remains), broken on master; reclassify KnownDivergence after merge.

// GnoOutput:

// GnoError:
// undefined

// GoOutput:

// KnownIssue:
// Local-type identity is already correct (the assertion fails as it
// should); the bug is the panic value: runtime panics are bare strings, so
// recover().(error) yields nil and scopes() re-panics nil ("undefined").
// Same root cause as fixedbugs/issue19040.go. On PR #5732 the remaining gap
// is message wording only (no "different scopes") — KnownDivergence then.
