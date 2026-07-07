// run

// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "runtime"

type Closer interface {
	Close()
}

func nilInterfaceDeferCall() {
	defer func() {
		// make sure a traceback happens with jmpdefer on the stack
		runtime.GC()
	}()
	var x Closer
	defer x.Close()
}

func shouldPanic(f func()) {
	defer func() {
		if recover() == nil {
			panic("did not panic")
		}
	}()
	f()
}

func main() {
	shouldPanic(nilInterfaceDeferCall)
}


// Fixed: master PR #5715 (df91bada8); verified clean, broken at parent; re-golden after rebase.

// GnoOutput:

// GnoError:
// interface method call on undefined value

// GoOutput:

// KnownIssue:
// Deferred nil-interface method call escaped as an unrecoverable VM error
// instead of a recoverable runtime panic, so recover() never fired. Same
// root cause as fixedbugs/issue15975.go and issue16760.go.
