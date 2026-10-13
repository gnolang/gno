// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func _() (int, UndefinedType /* ERROR "undefined: UndefinedType" */ , string)  {
	return 0 // ERROR "not enough return values\n\thave (number)\n\twant (int, unknown type, string)"
}

func _() (int, UndefinedType /* ERROR "undefined: UndefinedType" */ ) {
	return 0, 1, 2 // ERROR "too many return values\n\thave (number, number, number)\n\twant (int, unknown type)"
}

// test case from issue
func _() UndefinedType /* ERROR "undefined: UndefinedType" */ {
	return // ERROR "not enough return values\n\thave ()\n\twant (unknown type)"
}

// GnoError:
// line 7: 2: name UndefinedType not defined in fileset with files [issue58742.go]
// line 11: 2: name UndefinedType not defined in fileset with files [issue58742.go]
// line 16: 2: name UndefinedType not defined in fileset with files [issue58742.go]

// GoTypeCheckError:
// line 7
// line 8
// line 11
// line 12
// line 16
// line 17

// Clean: errorcheck
