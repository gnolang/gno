// errorcheck

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

type T interface{ M() }

func F() T

var _ = F().(*X) // ERROR "impossible type assertion:( F\(\).\(\*X\))?\n\t\*X does not implement T \(missing method M\)"

type X struct{}

// GnoError:
// line 11: function F does not have a body but is not natively defined (did you build after pulling from the repository?)
// line 13: impossible type assertion: *gno.land/p/filetest/p.X does not implement gno.land/p/filetest/p.T (missing method M)

// GnoWordingDiffers:
// line 13: gc expects: impossible type assertion:( F\(\).\(\*X\))?\n\t\*X does not implement T \(missing method M\); gno: impossible type assertion: *gno.land/p/filetest/p.X does not implement gno.land/p/filetest/p.T (missing method M)

// Verdict: Clean errorcheck
