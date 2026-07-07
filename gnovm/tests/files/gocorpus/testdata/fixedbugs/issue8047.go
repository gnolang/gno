// run

// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 8047.  Stack copier shouldn't crash if there
// is a nil defer.

package main

func stackit(n int) {
	if n == 0 {
		return
	}
	stackit(n - 1)
}

func main() {
	defer func() {
		// catch & ignore panic from nil defer below
		err := recover()
		if err == nil {
			panic("defer of nil func didn't panic")
		}
	}()
	defer ((func())(nil))()
	stackit(1000)
}


// Fixed: master PR #5722 (49af0f55c); verified clean, broken at parent; re-golden after rebase.

// GnoOutput:

// GnoError:
// interface conversion: gnolang.Value is nil, not *gnolang.FuncValue

// GoOutput:

// KnownIssue:
// Deferring a typed-nil func value crashed the VM host-side: the nil
// Value was asserted to *FuncValue without a nil check, instead of a
// recoverable "nil function" panic at call time.
