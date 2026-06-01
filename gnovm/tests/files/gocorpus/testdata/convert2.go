// errorcheck

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test various valid and invalid struct assignments and conversions.
// Does not compile.

package main

type I interface {
	m()
}

// conversions between structs

func _() {
	type S struct{}
	type T struct{}
	var s S
	var t T
	var u struct{}
	s = s
	s = t // ERROR "cannot use .* in assignment|incompatible type"
	s = u
	s = S(s)
	s = S(t)
	s = S(u)
	t = u
	t = T(u)
}

func _() {
	type S struct{ x int }
	type T struct {
		x int "foo"
	}
	var s S
	var t T
	var u struct {
		x int "bar"
	}
	s = s
	s = t // ERROR "cannot use .* in assignment|incompatible type"
	s = u // ERROR "cannot use .* in assignment|incompatible type"
	s = S(s)
	s = S(t)
	s = S(u)
	t = u // ERROR "cannot use .* in assignment|incompatible type"
	t = T(u)
}

func _() {
	type E struct{ x int }
	type S struct{ x E }
	type T struct {
		x E "foo"
	}
	var s S
	var t T
	var u struct {
		x E "bar"
	}
	s = s
	s = t // ERROR "cannot use .* in assignment|incompatible type"
	s = u // ERROR "cannot use .* in assignment|incompatible type"
	s = S(s)
	s = S(t)
	s = S(u)
	t = u // ERROR "cannot use .* in assignment|incompatible type"
	t = T(u)
}

func _() {
	type S struct {
		x struct {
			x int "foo"
		}
	}
	type T struct {
		x struct {
			x int "bar"
		} "foo"
	}
	var s S
	var t T
	var u struct {
		x struct {
			x int "bar"
		} "bar"
	}
	s = s
	s = t // ERROR "cannot use .* in assignment|incompatible type"
	s = u // ERROR "cannot use .* in assignment|incompatible type"
	s = S(s)
	s = S(t)
	s = S(u)
	t = u // ERROR "cannot use .* in assignment|incompatible type"
	t = T(u)
}

func _() {
	type E1 struct {
		x int "foo"
	}
	type E2 struct {
		x int "bar"
	}
	type S struct{ x E1 }
	type T struct {
		x E2 "foo"
	}
	var s S
	var t T
	var u struct {
		x E2 "bar"
	}
	s = s
	s = t // ERROR "cannot use .* in assignment|incompatible type"
	s = u // ERROR "cannot use .* in assignment|incompatible type"
	s = S(s)
	s = S(t) // ERROR "cannot convert"
	s = S(u) // ERROR "cannot convert"
	t = u    // ERROR "cannot use .* in assignment|incompatible type"
	t = T(u)
}

func _() {
	type E struct{ x int }
	type S struct {
		f func(struct {
			x int "foo"
		})
	}
	type T struct {
		f func(struct {
			x int "bar"
		})
	}
	var s S
	var t T
	var u struct{ f func(E) }
	s = s
	s = t // ERROR "cannot use .* in assignment|incompatible type"
	s = u // ERROR "cannot use .* in assignment|incompatible type"
	s = S(s)
	s = S(t)
	s = S(u) // ERROR "cannot convert"
	t = u    // ERROR "cannot use .* in assignment|incompatible type"
	t = T(u) // ERROR "cannot convert"
}

// conversions between pointers to structs

func _() {
	type S struct{}
	type T struct{}
	var s *S
	var t *T
	var u *struct{}
	s = s
	s = t // ERROR "cannot use .* in assignment|incompatible type"
	s = u // ERROR "cannot use .* in assignment|incompatible type"
	s = (*S)(s)
	s = (*S)(t)
	s = (*S)(u)
	t = u // ERROR "cannot use .* in assignment|incompatible type"
	t = (*T)(u)
}

func _() {
	type S struct{ x int }
	type T struct {
		x int "foo"
	}
	var s *S
	var t *T
	var u *struct {
		x int "bar"
	}
	s = s
	s = t // ERROR "cannot use .* in assignment|incompatible type"
	s = u // ERROR "cannot use .* in assignment|incompatible type"
	s = (*S)(s)
	s = (*S)(t)
	s = (*S)(u)
	t = u // ERROR "cannot use .* in assignment|incompatible type"
	t = (*T)(u)
}

func _() {
	type E struct{ x int }
	type S struct{ x E }
	type T struct {
		x E "foo"
	}
	var s *S
	var t *T
	var u *struct {
		x E "bar"
	}
	s = s
	s = t // ERROR "cannot use .* in assignment|incompatible type"
	s = u // ERROR "cannot use .* in assignment|incompatible type"
	s = (*S)(s)
	s = (*S)(t)
	s = (*S)(u)
	t = u // ERROR "cannot use .* in assignment|incompatible type"
	t = (*T)(u)
}

func _() {
	type S struct {
		x struct {
			x int "foo"
		}
	}
	type T struct {
		x struct {
			x int "bar"
		} "foo"
	}
	var s *S
	var t *T
	var u *struct {
		x struct {
			x int "bar"
		} "bar"
	}
	s = s
	s = t // ERROR "cannot use .* in assignment|incompatible type"
	s = u // ERROR "cannot use .* in assignment|incompatible type"
	s = (*S)(s)
	s = (*S)(t)
	s = (*S)(u)
	t = u // ERROR "cannot use .* in assignment|incompatible type"
	t = (*T)(u)
}

func _() {
	type E1 struct {
		x int "foo"
	}
	type E2 struct {
		x int "bar"
	}
	type S struct{ x E1 }
	type T struct {
		x E2 "foo"
	}
	var s *S
	var t *T
	var u *struct {
		x E2 "bar"
	}
	s = s
	s = t // ERROR "cannot use .* in assignment|incompatible type"
	s = u // ERROR "cannot use .* in assignment|incompatible type"
	s = (*S)(s)
	s = (*S)(t) // ERROR "cannot convert"
	s = (*S)(u) // ERROR "cannot convert"
	t = u       // ERROR "cannot use .* in assignment|incompatible type"
	t = (*T)(u)
}

func _() {
	type E struct{ x int }
	type S struct {
		f func(struct {
			x int "foo"
		})
	}
	type T struct {
		f func(struct {
			x int "bar"
		})
	}
	var s *S
	var t *T
	var u *struct{ f func(E) }
	s = s
	s = t // ERROR "cannot use .* in assignment|incompatible type"
	s = u // ERROR "cannot use .* in assignment|incompatible type"
	s = (*S)(s)
	s = (*S)(t)
	s = (*S)(u) // ERROR "cannot convert"
	t = u       // ERROR "cannot use .* in assignment|incompatible type"
	t = (*T)(u) // ERROR "cannot convert"
}

func _() {
	type E struct{ x int }
	type S struct {
		f func(*struct {
			x int "foo"
		})
	}
	type T struct {
		f func(*struct {
			x int "bar"
		})
	}
	var s *S
	var t *T
	var u *struct{ f func(E) }
	s = s
	s = t // ERROR "cannot use .* in assignment|incompatible type"
	s = u // ERROR "cannot use .* in assignment|incompatible type"
	s = (*S)(s)
	s = (*S)(t)
	s = (*S)(u) // ERROR "cannot convert"
	t = u       // ERROR "cannot use .* in assignment|incompatible type"
	t = (*T)(u) // ERROR "cannot convert"
}

func _() {
	var s []byte
	_ = ([4]byte)(s)
	_ = (*[4]byte)(s)

	type A [4]byte
	_ = (A)(s)
	_ = (*A)(s)

	type P *[4]byte
	_ = (P)(s)
	_ = (*P)(s) // ERROR "cannot convert"
}

// GnoStaticIncomplete: covered 31 of 48 markers (Gno preprocess: 31, go/types guard: 1); Gno bailed before the rest — a runnable variant may exercise more

// GnoError:
// line 25: cannot use main[main/convert2.go:18:1-32:2].T as main[main/convert2.go:18:1-32:2].S without explicit conversion
// line 45: cannot use main[main/convert2.go:34:1-52:2].T as main[main/convert2.go:34:1-52:2].S without explicit conversion
// line 66: cannot use main[main/convert2.go:54:1-73:2].T as main[main/convert2.go:54:1-73:2].S without explicit conversion
// line 94: cannot use main[main/convert2.go:75:1-101:2].T as main[main/convert2.go:75:1-101:2].S without explicit conversion
// line 120: cannot use main[main/convert2.go:103:1-127:2].T as main[main/convert2.go:103:1-127:2].S without explicit conversion
// line 121: cannot use struct{x main[main/convert2.go:103:1-127:2].E2} as struct{x main[main/convert2.go:103:1-127:2].E1}
// line 123: cannot convert t<VPBlock(1,5)> (of type main[main/convert2.go:103:1-127:2].T) to type main[main/convert2.go:103:1-127:2].S
// line 124: cannot convert u<VPBlock(1,6)> (of type struct{x main[main/convert2.go:103:1-127:2].E2}) to type main[main/convert2.go:103:1-127:2].S
// line 145: cannot use main[main/convert2.go:129:1-152:2].T as main[main/convert2.go:129:1-152:2].S without explicit conversion
// line 146: cannot use struct{f func(main[main/convert2.go:129:1-152:2].E)} as struct{f func(struct{x int})}
// line 149: cannot convert u<VPBlock(1,5)> (of type struct{f func(main[main/convert2.go:129:1-152:2].E)}) to type main[main/convert2.go:129:1-152:2].S
// line 150: cannot use struct{f func(main[main/convert2.go:129:1-152:2].E)} as struct{f func(struct{x int})}
// line 151: cannot convert u<VPBlock(1,5)> (of type struct{f func(main[main/convert2.go:129:1-152:2].E)}) to type main[main/convert2.go:129:1-152:2].T
// line 163: cannot use main[main/convert2.go:156:1-170:2].T as main[main/convert2.go:156:1-170:2].S without explicit conversion
// line 183: cannot use main[main/convert2.go:172:1-190:2].T as main[main/convert2.go:172:1-190:2].S without explicit conversion
// line 204: cannot use main[main/convert2.go:192:1-211:2].T as main[main/convert2.go:192:1-211:2].S without explicit conversion
// line 232: cannot use main[main/convert2.go:213:1-239:2].T as main[main/convert2.go:213:1-239:2].S without explicit conversion
// line 258: cannot use main[main/convert2.go:241:1-265:2].T as main[main/convert2.go:241:1-265:2].S without explicit conversion
// line 259: cannot use struct{x main[main/convert2.go:241:1-265:2].E2} as struct{x main[main/convert2.go:241:1-265:2].E1}
// line 261: cannot convert t<VPBlock(1,5)> (of type *main[main/convert2.go:241:1-265:2].T) to type *main[main/convert2.go:241:1-265:2].S
// line 262: cannot convert u<VPBlock(1,6)> (of type *struct{x main[main/convert2.go:241:1-265:2].E2}) to type *main[main/convert2.go:241:1-265:2].S
// line 283: cannot use main[main/convert2.go:267:1-290:2].T as main[main/convert2.go:267:1-290:2].S without explicit conversion
// line 284: cannot use struct{f func(main[main/convert2.go:267:1-290:2].E)} as struct{f func(struct{x int})}
// line 287: cannot convert u<VPBlock(1,5)> (of type *struct{f func(main[main/convert2.go:267:1-290:2].E)}) to type *main[main/convert2.go:267:1-290:2].S
// line 288: cannot use struct{f func(main[main/convert2.go:267:1-290:2].E)} as struct{f func(struct{x int})}
// line 289: cannot convert u<VPBlock(1,5)> (of type *struct{f func(main[main/convert2.go:267:1-290:2].E)}) to type *main[main/convert2.go:267:1-290:2].T
// line 308: cannot use main[main/convert2.go:292:1-315:2].T as main[main/convert2.go:292:1-315:2].S without explicit conversion
// line 309: cannot use struct{f func(main[main/convert2.go:292:1-315:2].E)} as struct{f func(*struct{x int})}
// line 312: cannot convert u<VPBlock(1,5)> (of type *struct{f func(main[main/convert2.go:292:1-315:2].E)}) to type *main[main/convert2.go:292:1-315:2].S
// line 313: cannot use struct{f func(main[main/convert2.go:292:1-315:2].E)} as struct{f func(*struct{x int})}
// line 314: cannot convert u<VPBlock(1,5)> (of type *struct{f func(main[main/convert2.go:292:1-315:2].E)}) to type *main[main/convert2.go:292:1-315:2].T

// GoTypeCheckError:
// line 25: cannot use t (variable of struct type T) as S value in assignment
