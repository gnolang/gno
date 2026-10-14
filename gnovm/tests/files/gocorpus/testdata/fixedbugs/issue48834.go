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
// line 10: expected 2 return values; got 1
// line 15: expected 2 return values; got 1
// line 19: expected 1 return values; got 2
// line 23: expected 0 return values; got 1

// GoTypeCheckError:
// line 10
// line 15
// line 19
// line 23

// GnoWordingDiffers:
// line 10: gc expects: not enough (arguments to return|return values)\n\thave \(number\)\n\twant \(int, error\); gno: expected 2 return values; got 1
// line 15: gc expects: not enough (arguments to return|return values)\n\thave \(int\)\n\twant \(int, error\); gno: expected 2 return values; got 1
// line 19: gc expects: too many (arguments to return|return values)\n\thave \(number, number\)\n\twant \(int\); gno: expected 1 return values; got 2
// line 23: gc expects: too many (arguments to return|return values)\n\thave \(number\)\n\twant \(\); gno: expected 0 return values; got 1

// Verdict: Clean errorcheck
