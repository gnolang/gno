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

// GnoIncomplete: covered 1 of 4 markers; Gno bailed before the rest — a runnable variant is needed to exercise them

// GnoError:
// line 10: expected 2 return values

// GoTypeCheckError:
// line 10: not enough return values
// 	have (number)
// 	want (int, error)
