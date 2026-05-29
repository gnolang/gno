// errorcheck

// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 4283: nil == nil can't be done as the type is unknown.

package p

func F1() bool {
	return nil == nil	// ERROR "invalid"
}

func F2() bool {
	return nil != nil	// ERROR "invalid"
}

// GnoError:
// line 12: <nil> is not comparable

// GoTypeCheckError:
// line 12: invalid operation: nil == nil (operator == not defined on untyped nil)
// line 16: invalid operation: nil != nil (operator != not defined on untyped nil)
