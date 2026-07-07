// run

// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type I interface{ M() }

type T struct {
	x int
}

func (T) M() {}

var pt *T

func f() (r int) {
	defer func() { recover() }()

	var i I = pt
	defer i.M()
	r = 1
	return
}

func main() {
	if got := f(); got != 1 {
		panic(got)
	}
}


// Fixing: PR #5737 (fix/defer12, call-time dispatch); verified clean on branch, broken on master; re-golden after merge.

// GnoOutput:

// GnoError:
// 0

// GoOutput:

// KnownIssue:
// defer i.M() on an interface holding nil *T panics when the defer is
// registered instead of at call time, so r = 1 never runs and f() returns
// 0 (Go dispatches interface-bound method values at call time; the panic
// then fires inside the deferred call and is recovered).
