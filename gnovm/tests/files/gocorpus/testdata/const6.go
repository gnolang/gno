// errorcheck

// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Ideal vs non-ideal bool. See issue 3915, 3923.

package p

type mybool bool
type mybool1 bool

var (
	x, y int = 1, 2
	c1 bool = x < y
	c2 mybool = x < y
	c3 mybool = c2 == (x < y)
	c4 mybool = c2 == (1 < 2)
	c5 mybool = 1 < 2
	c6 mybool1 = x < y
	c7 = c1 == c2 // ERROR "mismatched types|incompatible types"
	c8 = c2 == c6 // ERROR "mismatched types|incompatible types"
	c9 = c1 == c6 // ERROR "mismatched types|incompatible types"
	_ = c2 && (x < y)
	_ = c2 && (1 < 2)
	_ = c1 && c2 // ERROR "mismatched types|incompatible types"
	_ = c2 && c6 // ERROR "mismatched types|incompatible types"
	_ = c1 && c6 // ERROR "mismatched types|incompatible types"
)

// GnoError:
// line 22: invalid operation: (mismatched types bool and gno.land/p/filetest/p.mybool)
// line 23: invalid operation: (mismatched types gno.land/p/filetest/p.mybool and gno.land/p/filetest/p.mybool1)
// line 24: invalid operation: (mismatched types bool and gno.land/p/filetest/p.mybool1)
// line 27: invalid operation: c1<VPBlock(2,4)> && c2<VPBlock(2,5)> (mismatched types bool and gno.land/p/filetest/p.mybool)
// line 28: invalid operation: c2<VPBlock(2,5)> && c6<VPBlock(2,9)> (mismatched types gno.land/p/filetest/p.mybool and gno.land/p/filetest/p.mybool1)
// line 29: invalid operation: c1<VPBlock(2,4)> && c6<VPBlock(2,9)> (mismatched types bool and gno.land/p/filetest/p.mybool1)

// GoTypeCheckError:
// line 22: invalid operation: c1 == c2 (mismatched types bool and mybool)
// line 23: invalid operation: c2 == c6 (mismatched types mybool and mybool1)
// line 24: invalid operation: c1 == c6 (mismatched types bool and mybool1)
// line 27: invalid operation: c1 && c2 (mismatched types bool and mybool)
// line 28: invalid operation: c2 && c6 (mismatched types mybool and mybool1)
// line 29: invalid operation: c1 && c6 (mismatched types bool and mybool1)
