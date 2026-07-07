// errorcheck

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// We have a limit of 1GB for stack frames.
// Make sure we include the callee args section.

package main

type Big = [400e6]byte

func f() { // GC_ERROR "stack frame too large"
	// Note: This test relies on the fact that we currently always
	// spill function-results to the stack, even if they're so
	// large that we would normally heap allocate them. If we ever
	// improve the backend to spill temporaries to the heap, this
	// test will probably need updating to find some new way to
	// construct an overly large stack frame.
	g(h(), h())
}

func g(Big, Big)
func h() Big

// GnoStaticIncomplete: covered 0 of 1 markers (Gno preprocess: 0, go/types guard: 0); Gno's own preprocess flags none (lenient); the rest are caught by neither — a runnable variant may exercise more

// GnoOverStrictError:
// line 24: function g does not have a body but is not natively defined (did you build after pulling from the repository?)

// KnownIssue:
// TODO: explain the Gno bug (Gno rejects lines gc + go/types accept)
