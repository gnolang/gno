// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test various valid and invalid struct assignments and conversions.
// Does not compile.

package conversions2

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
	s = t // ERRORx `cannot use .* in assignment`
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
	s = t // ERRORx `cannot use .* in assignment`
	s = u // ERRORx `cannot use .* in assignment`
	s = S(s)
	s = S(t)
	s = S(u)
	t = u // ERRORx `cannot use .* in assignment`
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
	s = t // ERRORx `cannot use .* in assignment`
	s = u // ERRORx `cannot use .* in assignment`
	s = S(s)
	s = S(t)
	s = S(u)
	t = u // ERRORx `cannot use .* in assignment`
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
	s = t // ERRORx `cannot use .* in assignment`
	s = u // ERRORx `cannot use .* in assignment`
	s = S(s)
	s = S(t)
	s = S(u)
	t = u // ERRORx `cannot use .* in assignment`
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
	s = t // ERRORx `cannot use .* in assignment`
	s = u // ERRORx `cannot use .* in assignment`
	s = S(s)
	s = S(t /* ERROR "cannot convert" */ )
	s = S(u /* ERROR "cannot convert" */ )
	t = u   // ERRORx `cannot use .* in assignment`
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
	s = t // ERRORx `cannot use .* in assignment`
	s = u // ERRORx `cannot use .* in assignment`
	s = S(s)
	s = S(t)
	s = S(u /* ERROR "cannot convert" */ )
	t = u   // ERRORx `cannot use .* in assignment`
	t = T(u /* ERROR "cannot convert" */ )
}

// conversions between pointers to structs

func _() {
	type S struct{}
	type T struct{}
	var s *S
	var t *T
	var u *struct{}
	s = s
	s = t // ERRORx `cannot use .* in assignment`
	s = u // ERRORx `cannot use .* in assignment`
	s = (*S)(s)
	s = (*S)(t)
	s = (*S)(u)
	t = u // ERRORx `cannot use .* in assignment`
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
	s = t // ERRORx `cannot use .* in assignment`
	s = u // ERRORx `cannot use .* in assignment`
	s = (*S)(s)
	s = (*S)(t)
	s = (*S)(u)
	t = u // ERRORx `cannot use .* in assignment`
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
	s = t // ERRORx `cannot use .* in assignment`
	s = u // ERRORx `cannot use .* in assignment`
	s = (*S)(s)
	s = (*S)(t)
	s = (*S)(u)
	t = u // ERRORx `cannot use .* in assignment`
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
	s = t // ERRORx `cannot use .* in assignment`
	s = u // ERRORx `cannot use .* in assignment`
	s = (*S)(s)
	s = (*S)(t)
	s = (*S)(u)
	t = u // ERRORx `cannot use .* in assignment`
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
	s = t // ERRORx `cannot use .* in assignment`
	s = u // ERRORx `cannot use .* in assignment`
	s = (*S)(s)
	s = (*S)(t /* ERROR "cannot convert" */ )
	s = (*S)(u /* ERROR "cannot convert" */ )
	t = u      // ERRORx `cannot use .* in assignment`
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
	s = t // ERRORx `cannot use .* in assignment`
	s = u // ERRORx `cannot use .* in assignment`
	s = (*S)(s)
	s = (*S)(t)
	s = (*S)(u /* ERROR "cannot convert" */ )
	t = u      // ERRORx `cannot use .* in assignment`
	t = (*T)(u /* ERROR "cannot convert" */ )
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
	s = t // ERRORx `cannot use .* in assignment`
	s = u // ERRORx `cannot use .* in assignment`
	s = (*S)(s)
	s = (*S)(t)
	s = (*S)(u /* ERROR "cannot convert" */ )
	t = u      // ERRORx `cannot use .* in assignment`
	t = (*T)(u /* ERROR "cannot convert" */ )
}

// GnoError:
// line 23: cannot use gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:17:1-31:2].T as gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:17:1-31:2].S without explicit conversion
// line 43: cannot use gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:33:1-51:2].T as gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:33:1-51:2].S without explicit conversion
// line 44: cannot use struct{x int} as struct{x int}: struct tags differ
// line 48: cannot use struct{x int} as struct{x int}: struct tags differ
// line 64: cannot use gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:53:1-72:2].T as gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:53:1-72:2].S without explicit conversion
// line 65: cannot use struct{x gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:53:1-72:2].E} as struct{x gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:53:1-72:2].E}: struct tags differ
// line 69: cannot use struct{x gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:53:1-72:2].E} as struct{x gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:53:1-72:2].E}: struct tags differ
// line 92: cannot use gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:74:1-100:2].T as gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:74:1-100:2].S without explicit conversion
// line 93: cannot use struct{x struct{x int}} as struct{x struct{x int}}: struct tags differ
// line 97: cannot use struct{x struct{x int}} as struct{x struct{x int}}: struct tags differ
// line 118: cannot use gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:102:1-126:2].T as gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:102:1-126:2].S without explicit conversion
// line 119: cannot use struct{x gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:102:1-126:2].E2} as struct{x gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:102:1-126:2].E1}
// line 121: cannot convert t<VPBlock(1,5)> (of type gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:102:1-126:2].T) to type gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:102:1-126:2].S
// line 122: cannot convert u<VPBlock(1,6)> (of type struct{x gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:102:1-126:2].E2}) to type gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:102:1-126:2].S
// line 123: cannot use struct{x gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:102:1-126:2].E2} as struct{x gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:102:1-126:2].E2}: struct tags differ
// line 143: cannot use gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:128:1-151:2].T as gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:128:1-151:2].S without explicit conversion
// line 144: cannot use struct{f func(gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:128:1-151:2].E)} as struct{f func(struct{x int})}
// line 147: cannot convert u<VPBlock(1,5)> (of type struct{f func(gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:128:1-151:2].E)}) to type gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:128:1-151:2].S
// line 148: cannot use struct{f func(gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:128:1-151:2].E)} as struct{f func(struct{x int})}
// line 149: cannot convert u<VPBlock(1,5)> (of type struct{f func(gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:128:1-151:2].E)}) to type gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:128:1-151:2].T
// line 161: cannot use gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:155:1-169:2].T as gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:155:1-169:2].S without explicit conversion
// line 181: cannot use gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:171:1-189:2].T as gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:171:1-189:2].S without explicit conversion
// line 182: cannot use struct{x int} as struct{x int}: struct tags differ
// line 186: cannot use struct{x int} as struct{x int}: struct tags differ
// line 202: cannot use gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:191:1-210:2].T as gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:191:1-210:2].S without explicit conversion
// line 203: cannot use struct{x gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:191:1-210:2].E} as struct{x gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:191:1-210:2].E}: struct tags differ
// line 207: cannot use struct{x gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:191:1-210:2].E} as struct{x gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:191:1-210:2].E}: struct tags differ
// line 230: cannot use gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:212:1-238:2].T as gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:212:1-238:2].S without explicit conversion
// line 231: cannot use struct{x struct{x int}} as struct{x struct{x int}}: struct tags differ
// line 235: cannot use struct{x struct{x int}} as struct{x struct{x int}}: struct tags differ
// line 256: cannot use gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:240:1-264:2].T as gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:240:1-264:2].S without explicit conversion
// line 257: cannot use struct{x gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:240:1-264:2].E2} as struct{x gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:240:1-264:2].E1}
// line 259: cannot convert t<VPBlock(1,5)> (of type *gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:240:1-264:2].T) to type *gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:240:1-264:2].S
// line 260: cannot convert u<VPBlock(1,6)> (of type *struct{x gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:240:1-264:2].E2}) to type *gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:240:1-264:2].S
// line 261: cannot use struct{x gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:240:1-264:2].E2} as struct{x gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:240:1-264:2].E2}: struct tags differ
// line 281: cannot use gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:266:1-289:2].T as gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:266:1-289:2].S without explicit conversion
// line 282: cannot use struct{f func(gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:266:1-289:2].E)} as struct{f func(struct{x int})}
// line 285: cannot convert u<VPBlock(1,5)> (of type *struct{f func(gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:266:1-289:2].E)}) to type *gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:266:1-289:2].S
// line 286: cannot use struct{f func(gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:266:1-289:2].E)} as struct{f func(struct{x int})}
// line 287: cannot convert u<VPBlock(1,5)> (of type *struct{f func(gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:266:1-289:2].E)}) to type *gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:266:1-289:2].T
// line 306: cannot use gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:291:1-314:2].T as gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:291:1-314:2].S without explicit conversion
// line 307: cannot use struct{f func(gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:291:1-314:2].E)} as struct{f func(*struct{x int})}
// line 310: cannot convert u<VPBlock(1,5)> (of type *struct{f func(gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:291:1-314:2].E)}) to type *gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:291:1-314:2].S
// line 311: cannot use struct{f func(gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:291:1-314:2].E)} as struct{f func(*struct{x int})}
// line 312: cannot convert u<VPBlock(1,5)> (of type *struct{f func(gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:291:1-314:2].E)}) to type *gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:291:1-314:2].T

// GnoWordingDiffers:
// line 23: gc expects: cannot use .* in assignment; gno: cannot use gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:17:1-31:2].T as gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:17:1-31:2].S without explicit conversion
// line 43: gc expects: cannot use .* in assignment; gno: cannot use gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:33:1-51:2].T as gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:33:1-51:2].S without explicit conversion
// line 44: gc expects: cannot use .* in assignment; gno: cannot use struct{x int} as struct{x int}: struct tags differ
// line 48: gc expects: cannot use .* in assignment; gno: cannot use struct{x int} as struct{x int}: struct tags differ
// line 64: gc expects: cannot use .* in assignment; gno: cannot use gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:53:1-72:2].T as gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:53:1-72:2].S without explicit conversion
// line 65: gc expects: cannot use .* in assignment; gno: cannot use struct{x gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:53:1-72:2].E} as struct{x gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:53:1-72:2].E}: struct tags differ
// line 69: gc expects: cannot use .* in assignment; gno: cannot use struct{x gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:53:1-72:2].E} as struct{x gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:53:1-72:2].E}: struct tags differ
// line 92: gc expects: cannot use .* in assignment; gno: cannot use gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:74:1-100:2].T as gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:74:1-100:2].S without explicit conversion
// line 93: gc expects: cannot use .* in assignment; gno: cannot use struct{x struct{x int}} as struct{x struct{x int}}: struct tags differ
// line 97: gc expects: cannot use .* in assignment; gno: cannot use struct{x struct{x int}} as struct{x struct{x int}}: struct tags differ
// line 118: gc expects: cannot use .* in assignment; gno: cannot use gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:102:1-126:2].T as gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:102:1-126:2].S without explicit conversion
// line 119: gc expects: cannot use .* in assignment; gno: cannot use struct{x gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:102:1-126:2].E2} as struct{x gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:102:1-126:2].E1}
// line 123: gc expects: cannot use .* in assignment; gno: cannot use struct{x gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:102:1-126:2].E2} as struct{x gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:102:1-126:2].E2}: struct tags differ
// line 143: gc expects: cannot use .* in assignment; gno: cannot use gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:128:1-151:2].T as gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:128:1-151:2].S without explicit conversion
// line 144: gc expects: cannot use .* in assignment; gno: cannot use struct{f func(gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:128:1-151:2].E)} as struct{f func(struct{x int})}
// line 148: gc expects: cannot use .* in assignment; gno: cannot use struct{f func(gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:128:1-151:2].E)} as struct{f func(struct{x int})}
// line 161: gc expects: cannot use .* in assignment; gno: cannot use gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:155:1-169:2].T as gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:155:1-169:2].S without explicit conversion
// line 181: gc expects: cannot use .* in assignment; gno: cannot use gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:171:1-189:2].T as gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:171:1-189:2].S without explicit conversion
// line 182: gc expects: cannot use .* in assignment; gno: cannot use struct{x int} as struct{x int}: struct tags differ
// line 186: gc expects: cannot use .* in assignment; gno: cannot use struct{x int} as struct{x int}: struct tags differ
// line 202: gc expects: cannot use .* in assignment; gno: cannot use gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:191:1-210:2].T as gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:191:1-210:2].S without explicit conversion
// line 203: gc expects: cannot use .* in assignment; gno: cannot use struct{x gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:191:1-210:2].E} as struct{x gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:191:1-210:2].E}: struct tags differ
// line 207: gc expects: cannot use .* in assignment; gno: cannot use struct{x gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:191:1-210:2].E} as struct{x gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:191:1-210:2].E}: struct tags differ
// line 230: gc expects: cannot use .* in assignment; gno: cannot use gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:212:1-238:2].T as gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:212:1-238:2].S without explicit conversion
// line 231: gc expects: cannot use .* in assignment; gno: cannot use struct{x struct{x int}} as struct{x struct{x int}}: struct tags differ
// line 235: gc expects: cannot use .* in assignment; gno: cannot use struct{x struct{x int}} as struct{x struct{x int}}: struct tags differ
// line 256: gc expects: cannot use .* in assignment; gno: cannot use gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:240:1-264:2].T as gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:240:1-264:2].S without explicit conversion
// line 257: gc expects: cannot use .* in assignment; gno: cannot use struct{x gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:240:1-264:2].E2} as struct{x gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:240:1-264:2].E1}
// line 261: gc expects: cannot use .* in assignment; gno: cannot use struct{x gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:240:1-264:2].E2} as struct{x gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:240:1-264:2].E2}: struct tags differ
// line 281: gc expects: cannot use .* in assignment; gno: cannot use gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:266:1-289:2].T as gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:266:1-289:2].S without explicit conversion
// line 282: gc expects: cannot use .* in assignment; gno: cannot use struct{f func(gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:266:1-289:2].E)} as struct{f func(struct{x int})}
// line 286: gc expects: cannot use .* in assignment; gno: cannot use struct{f func(gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:266:1-289:2].E)} as struct{f func(struct{x int})}
// line 306: gc expects: cannot use .* in assignment; gno: cannot use gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:291:1-314:2].T as gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:291:1-314:2].S without explicit conversion
// line 307: gc expects: cannot use .* in assignment; gno: cannot use struct{f func(gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:291:1-314:2].E)} as struct{f func(*struct{x int})}
// line 311: gc expects: cannot use .* in assignment; gno: cannot use struct{f func(gno.land/p/filetest/conversions2[gno.land/p/filetest/conversions2/conversions1.go:291:1-314:2].E)} as struct{f func(*struct{x int})}

// Verdict: KnownIssue errorcheck: #6344; uncaught markers: lines 162, 166
