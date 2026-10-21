// errorcheck

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func f0()
func f1(_ int)
func f2(_, _ int)
func f2ddd(_, _ int, _ ...int)

func f() {
	var x int
	f0(1)              // ERROR "too many arguments in call to f0\n\thave \(number\)\n\twant \(\)"
	f0(x)              // ERROR "too many arguments in call to f0\n\thave \(int\)\n\twant \(\)"
	f1()               // ERROR "not enough arguments in call to f1\n\thave \(\)\n\twant \(int\)"
	f1(1, 2)           // ERROR "too many arguments in call to f1\n\thave \(number, number\)\n\twant \(int\)"
	f2(1)              // ERROR "not enough arguments in call to f2\n\thave \(number\)\n\twant \(int, int\)"
	f2(1, "foo", true) // ERROR "too many arguments in call to f2\n\thave \(number, string, bool\)\n\twant \(int, int\)"
	f2ddd(1)           // ERROR "not enough arguments in call to f2ddd\n\thave \(number\)\n\twant \(int, int, \.\.\.int\)"
	f2ddd(1, 2)
	f2ddd(1, 2, 3)
}

// GnoError:
// line 9: function f0 does not have a body but is not natively defined (did you build after pulling from the repository?)
// line 10: function f1 does not have a body but is not natively defined (did you build after pulling from the repository?)
// line 11: function f2 does not have a body but is not natively defined (did you build after pulling from the repository?)
// line 12: function f2ddd does not have a body but is not natively defined (did you build after pulling from the repository?)
// line 16: wrong argument count in call to f0<VPBlock(3,0)>; want 0 got 1
// line 17: wrong argument count in call to f0<VPBlock(3,0)>; want 0 got 1
// line 18: wrong argument count in call to f1<VPBlock(3,1)>; want 1 got 0
// line 19: wrong argument count in call to f1<VPBlock(3,1)>; want 1 got 2
// line 20: wrong argument count in call to f2<VPBlock(3,2)>; want 2 got 1
// line 21: wrong argument count in call to f2<VPBlock(3,2)>; want 2 got 3
// line 22: not enough arguments in call to f2ddd<VPBlock(3,3)>; want 2 (besides variadic) got 1

// GnoWordingDiffers:
// line 16: gc expects: too many arguments in call to f0\n\thave \(number\)\n\twant \(\); gno: wrong argument count in call to f0<VPBlock(3,0)>; want 0 got 1
// line 17: gc expects: too many arguments in call to f0\n\thave \(int\)\n\twant \(\); gno: wrong argument count in call to f0<VPBlock(3,0)>; want 0 got 1
// line 18: gc expects: not enough arguments in call to f1\n\thave \(\)\n\twant \(int\); gno: wrong argument count in call to f1<VPBlock(3,1)>; want 1 got 0
// line 19: gc expects: too many arguments in call to f1\n\thave \(number, number\)\n\twant \(int\); gno: wrong argument count in call to f1<VPBlock(3,1)>; want 1 got 2
// line 20: gc expects: not enough arguments in call to f2\n\thave \(number\)\n\twant \(int, int\); gno: wrong argument count in call to f2<VPBlock(3,2)>; want 2 got 1
// line 21: gc expects: too many arguments in call to f2\n\thave \(number, string, bool\)\n\twant \(int, int\); gno: wrong argument count in call to f2<VPBlock(3,2)>; want 2 got 3
// line 22: gc expects: not enough arguments in call to f2ddd\n\thave \(number\)\n\twant \(int, int, \.\.\.int\); gno: not enough arguments in call to f2ddd<VPBlock(3,3)>; want 2 (besides variadic) got 1

// Verdict: Clean errorcheck
