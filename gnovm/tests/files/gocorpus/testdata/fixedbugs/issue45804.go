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

// GnoOverStrictError:
// line 9: function g does not have a body but is not natively defined (did you build after pulling from the repository?)

// GoTypeCheckError:
// line 16: invalid operation: operator ! not defined on did (variable of type int)
// line 17: undefined: x

// KnownIssue:
// TODO: explain the Gno bug (Gno rejects lines gc + go/types accept)
