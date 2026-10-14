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

// GnoError:
// line 25: cannot use main[main/convert2.go:18:1-32:2].T as main[main/convert2.go:18:1-32:2].S without explicit conversion
// line 45: cannot use main[main/convert2.go:34:1-52:2].T as main[main/convert2.go:34:1-52:2].S without explicit conversion
// line 46: cannot use struct{x int} as struct{x int}: struct tags differ
// line 50: cannot use struct{x int} as struct{x int}: struct tags differ
// line 66: cannot use main[main/convert2.go:54:1-73:2].T as main[main/convert2.go:54:1-73:2].S without explicit conversion
// line 67: cannot use struct{x main[main/convert2.go:54:1-73:2].E} as struct{x main[main/convert2.go:54:1-73:2].E}: struct tags differ
// line 71: cannot use struct{x main[main/convert2.go:54:1-73:2].E} as struct{x main[main/convert2.go:54:1-73:2].E}: struct tags differ
// line 94: cannot use main[main/convert2.go:75:1-101:2].T as main[main/convert2.go:75:1-101:2].S without explicit conversion
// line 95: cannot use struct{x struct{x int}} as struct{x struct{x int}}: struct tags differ
// line 99: cannot use struct{x struct{x int}} as struct{x struct{x int}}: struct tags differ
// line 120: cannot use main[main/convert2.go:103:1-127:2].T as main[main/convert2.go:103:1-127:2].S without explicit conversion
// line 121: cannot use struct{x main[main/convert2.go:103:1-127:2].E2} as struct{x main[main/convert2.go:103:1-127:2].E1}
// line 123: cannot convert t<VPBlock(1,5)> (of type main[main/convert2.go:103:1-127:2].T) to type main[main/convert2.go:103:1-127:2].S
// line 124: cannot convert u<VPBlock(1,6)> (of type struct{x main[main/convert2.go:103:1-127:2].E2}) to type main[main/convert2.go:103:1-127:2].S
// line 125: cannot use struct{x main[main/convert2.go:103:1-127:2].E2} as struct{x main[main/convert2.go:103:1-127:2].E2}: struct tags differ
// line 145: cannot use main[main/convert2.go:129:1-152:2].T as main[main/convert2.go:129:1-152:2].S without explicit conversion
// line 146: cannot use struct{f func(main[main/convert2.go:129:1-152:2].E)} as struct{f func(struct{x int})}
// line 149: cannot convert u<VPBlock(1,5)> (of type struct{f func(main[main/convert2.go:129:1-152:2].E)}) to type main[main/convert2.go:129:1-152:2].S
// line 150: cannot use struct{f func(main[main/convert2.go:129:1-152:2].E)} as struct{f func(struct{x int})}
// line 151: cannot convert u<VPBlock(1,5)> (of type struct{f func(main[main/convert2.go:129:1-152:2].E)}) to type main[main/convert2.go:129:1-152:2].T
// line 163: cannot use main[main/convert2.go:156:1-170:2].T as main[main/convert2.go:156:1-170:2].S without explicit conversion
// line 183: cannot use main[main/convert2.go:172:1-190:2].T as main[main/convert2.go:172:1-190:2].S without explicit conversion
// line 184: cannot use struct{x int} as struct{x int}: struct tags differ
// line 188: cannot use struct{x int} as struct{x int}: struct tags differ
// line 204: cannot use main[main/convert2.go:192:1-211:2].T as main[main/convert2.go:192:1-211:2].S without explicit conversion
// line 205: cannot use struct{x main[main/convert2.go:192:1-211:2].E} as struct{x main[main/convert2.go:192:1-211:2].E}: struct tags differ
// line 209: cannot use struct{x main[main/convert2.go:192:1-211:2].E} as struct{x main[main/convert2.go:192:1-211:2].E}: struct tags differ
// line 232: cannot use main[main/convert2.go:213:1-239:2].T as main[main/convert2.go:213:1-239:2].S without explicit conversion
// line 233: cannot use struct{x struct{x int}} as struct{x struct{x int}}: struct tags differ
// line 237: cannot use struct{x struct{x int}} as struct{x struct{x int}}: struct tags differ
// line 258: cannot use main[main/convert2.go:241:1-265:2].T as main[main/convert2.go:241:1-265:2].S without explicit conversion
// line 259: cannot use struct{x main[main/convert2.go:241:1-265:2].E2} as struct{x main[main/convert2.go:241:1-265:2].E1}
// line 261: cannot convert t<VPBlock(1,5)> (of type *main[main/convert2.go:241:1-265:2].T) to type *main[main/convert2.go:241:1-265:2].S
// line 262: cannot convert u<VPBlock(1,6)> (of type *struct{x main[main/convert2.go:241:1-265:2].E2}) to type *main[main/convert2.go:241:1-265:2].S
// line 263: cannot use struct{x main[main/convert2.go:241:1-265:2].E2} as struct{x main[main/convert2.go:241:1-265:2].E2}: struct tags differ
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
// line 319: cannot convert s<VPBlock(1,0)> (of type []uint8) to type [4]uint8
// line 320: cannot convert s<VPBlock(1,0)> (of type []uint8) to type *[4]uint8
// line 323: cannot convert s<VPBlock(1,0)> (of type []uint8) to type main[main/convert2.go:317:1-329:2].A
// line 324: cannot convert s<VPBlock(1,0)> (of type []uint8) to type *main[main/convert2.go:317:1-329:2].A
// line 327: cannot convert s<VPBlock(1,0)> (of type []uint8) to type main[main/convert2.go:317:1-329:2].P
// line 328: cannot convert s<VPBlock(1,0)> (of type []uint8) to type *main[main/convert2.go:317:1-329:2].P

// GoTypeCheckError:
// line 25
// line 45
// line 46
// line 50
// line 66
// line 67
// line 71
// line 94
// line 95
// line 99
// line 120
// line 121
// line 123
// line 124
// line 125
// line 145
// line 146
// line 149
// line 150
// line 151
// line 163
// line 164
// line 168
// line 183
// line 184
// line 188
// line 204
// line 205
// line 209
// line 232
// line 233
// line 237
// line 258
// line 259
// line 261
// line 262
// line 263
// line 283
// line 284
// line 287
// line 288
// line 289
// line 308
// line 309
// line 312
// line 313
// line 314
// line 319
// line 323
// line 328

// GoTypesOnly:
// line 164: gc expects: cannot use .* in assignment|incompatible type
// line 168: gc expects: cannot use .* in assignment|incompatible type

// GnoOverStrictError:
// line 320: cannot convert s<VPBlock(1,0)> (of type []uint8) to type *[4]uint8
// line 324: cannot convert s<VPBlock(1,0)> (of type []uint8) to type *main[main/convert2.go:317:1-329:2].A
// line 327: cannot convert s<VPBlock(1,0)> (of type []uint8) to type main[main/convert2.go:317:1-329:2].P

// Verdict: KnownIssue errorcheck: 3 over-strict line(s), 0 uncaught marker(s)
