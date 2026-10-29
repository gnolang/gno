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
// line 7: name not defined: UndefinedType
// line 8: expected 3 return values; got 1
// line 11: name not defined: UndefinedType
// line 12: expected 2 return values; got 3
// line 16: name not defined: UndefinedType
// line 17: expected 1 return values; got 0

// GnoWordingDiffers:
// line 7: gc expects: undefined: UndefinedType; gno: name not defined: UndefinedType
// line 8: gc expects: not enough return values
// 	have (number)
// 	want (int, unknown type, string); gno: expected 3 return values; got 1
// line 11: gc expects: undefined: UndefinedType; gno: name not defined: UndefinedType
// line 12: gc expects: too many return values
// 	have (number, number, number)
// 	want (int, unknown type); gno: expected 2 return values; got 3
// line 16: gc expects: undefined: UndefinedType; gno: name not defined: UndefinedType
// line 17: gc expects: not enough return values
// 	have ()
// 	want (unknown type); gno: expected 1 return values; got 0

// Verdict: Clean errorcheck
