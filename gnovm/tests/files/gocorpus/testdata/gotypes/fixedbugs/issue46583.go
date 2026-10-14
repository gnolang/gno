// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

type T1 struct{}
func (t T1) m(int) {}
var f1 func(T1)

type T2 struct{}
func (t T2) m(x int) {}
var f2 func(T2)

type T3 struct{}
func (T3) m(int) {}
var f3 func(T3)

type T4 struct{}
func (T4) m(x int) {}
var f4 func(T4)

func _() {
	f1 = T1 /* ERROR "func(T1, int)" */ .m
	f2 = T2 /* ERROR "func(t T2, x int)" */ .m
	f3 = T3 /* ERROR "func(T3, int)" */ .m
	f4 = T4 /* ERROR "func(_ T4, x int)" */ .m
}

// GnoError:
// line 24: cannot use func(gno.land/p/filetest/p.T1, int) as func(gno.land/p/filetest/p.T1)
// line 25: cannot use func(gno.land/p/filetest/p.T2, int) as func(gno.land/p/filetest/p.T2)
// line 26: cannot use func(gno.land/p/filetest/p.T3, int) as func(gno.land/p/filetest/p.T3)
// line 27: cannot use func(gno.land/p/filetest/p.T4, int) as func(gno.land/p/filetest/p.T4)

// GoTypeCheckError:
// line 24
// line 25
// line 26
// line 27

// GnoWordingDiffers:
// line 24: gc expects: func(T1, int); gno: cannot use func(gno.land/p/filetest/p.T1, int) as func(gno.land/p/filetest/p.T1)
// line 25: gc expects: func(t T2, x int); gno: cannot use func(gno.land/p/filetest/p.T2, int) as func(gno.land/p/filetest/p.T2)
// line 26: gc expects: func(T3, int); gno: cannot use func(gno.land/p/filetest/p.T3, int) as func(gno.land/p/filetest/p.T3)
// line 27: gc expects: func(_ T4, x int); gno: cannot use func(gno.land/p/filetest/p.T4, int) as func(gno.land/p/filetest/p.T4)

// Verdict: Clean errorcheck
