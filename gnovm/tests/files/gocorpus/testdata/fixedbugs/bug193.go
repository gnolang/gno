// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func main() {
	s := uint(10)
	ss := 1 << s
	y1 := float64(ss)
	y2 := float64(1 << s) // ERROR "shift"
	y3 := string(1 << s)  // ERROR "shift"
	_, _, _, _, _ = s, ss, y1, y2, y3
}

// GnoError:
// line 13: operator << not defined on: Float64Kind
// line 14: operator << not defined on: StringKind

// GoTypeCheckError:
// line 13: invalid operation: shifted operand 1 (type float64) must be integer
// line 14: invalid operation: shifted operand 1 (type string) must be integer
