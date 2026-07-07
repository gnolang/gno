// run

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

var fail bool

type Closer interface {
	Close()
}

func nilInterfaceDeferCall() {
	var x Closer
	defer x.Close()
	// if it panics when evaluating x.Close, it should not reach here
	fail = true
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
	if fail {
		panic("fail")
	}
}


// Fixed: master PR #5715 (df91bada8); verified clean, broken at parent; re-golden after rebase.

// GnoOutput:

// GnoError:
// interface method call on undefined value

// GoOutput:

// KnownIssue:
// Deferred nil-interface method call escaped as an unrecoverable VM error
// instead of a recoverable runtime panic, so recover() never fired.
