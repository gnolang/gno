// errorcheck

// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

type T struct {
	f float64
}

var t T

func F() {
	_ = complex(1.0) // ERROR "invalid operation|not enough arguments"
	_ = complex(t.f) // ERROR "invalid operation|not enough arguments"
}

// GnoError:
// line 16: name complex not declared
// line 17: name complex not declared

// GoTypeCheckError:
// line 16: invalid operation: not enough arguments for complex(1.0) (expected 2, found 1)
// line 17: invalid operation: not enough arguments for complex(t.f) (expected 2, found 1)
