// errorcheck

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func g() int
func h(int)

var b bool

func f() {
	did := g()
	if !did && b { // ERROR "invalid operation"
		h(x) // ERROR "undefined"
	}
}

// GnoError:
// line 9: function g does not have a body but is not natively defined (did you build after pulling from the repository?)
// line 10: function h does not have a body but is not natively defined (did you build after pulling from the repository?)
// line 16: operator ! not defined on: IntKind
// line 17: name x not declared

// GnoWordingDiffers:
// line 16: gc expects: invalid operation; gno: operator ! not defined on: IntKind
// line 17: gc expects: undefined; gno: name x not declared

// Verdict: Clean errorcheck
