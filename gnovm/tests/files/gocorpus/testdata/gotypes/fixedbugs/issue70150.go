// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func _() {
	var values []int
	vf(values /* ERROR "(variable of type []int) as string value" */)
	vf(values...) /* ERROR "have ([]int...)\n\twant (string, ...int)" */
	vf("ab", "cd", values /* ERROR "have (string, string, []int...)\n\twant (string, ...int)" */ ...)
}

func vf(method string, values ...int) {
}

// GnoError:
// line 9: cannot use []int as string
// line 10: not enough arguments in call to vf<VPBlock(3,1)>; want 2 (including variadic) got 1
// line 11: not enough arguments in call to vf<VPBlock(3,1)>; want 2 (including variadic) got 3

// GnoWordingDiffers:
// line 9: gc expects: (variable of type []int) as string value; gno: cannot use []int as string
// line 10: gc expects: have ([]int...)
// 	want (string, ...int); gno: not enough arguments in call to vf<VPBlock(3,1)>; want 2 (including variadic) got 1
// line 11: gc expects: have (string, string, []int...)
// 	want (string, ...int); gno: not enough arguments in call to vf<VPBlock(3,1)>; want 2 (including variadic) got 3

// Verdict: Clean errorcheck
