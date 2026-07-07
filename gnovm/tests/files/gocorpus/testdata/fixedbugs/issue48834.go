// errorcheck

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func _() (int, error) {
	return 1 // ERROR "not enough (arguments to return|return values)\n\thave \(number\)\n\twant \(int, error\)"
}

func _() (int, error) {
	var x int
	return x // ERROR "not enough (arguments to return|return values)\n\thave \(int\)\n\twant \(int, error\)"
}

func _() int {
	return 1, 2 // ERROR "too many (arguments to return|return values)\n\thave \(number, number\)\n\twant \(int\)"
}

func _() {
	return 1 // ERROR "too many (arguments to return|return values)\n\thave \(number\)\n\twant \(\)"
}

// GnoError:
// line 9: 2: [function "._0" does not terminate]
// line 10: expected 2 return values
// line 11: expected declaration, found '}' (and 1 more errors)
// line 13: 2: [function "._0" does not terminate]
// line 15: expected 2 return values
// line 16: expected declaration, found '}'
// line 18: 2: [function "._1" does not terminate]
// line 19: expected 1 return values
// line 20: expected declaration, found '}'
// line 23: expected 0 return values

// GoTypeCheckError:
// line 10: not enough return values
// 	have (number)
// 	want (int, error)
// line 15: not enough return values
// 	have (int)
// 	want (int, error)
// line 19: too many return values
// 	have (number, number)
// 	want (int)
// line 23: too many return values
// 	have (number)
// 	want ()

// GnoOverStrictError:
// line 9: 2: [function "._0" does not terminate]
// line 11: expected declaration, found '}' (and 1 more errors)
// line 13: 2: [function "._0" does not terminate]
// line 16: expected declaration, found '}'
// line 18: 2: [function "._1" does not terminate]
// line 20: expected declaration, found '}'
