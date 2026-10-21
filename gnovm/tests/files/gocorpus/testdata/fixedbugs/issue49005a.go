// errorcheck

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

type T interface{ M() }

func F() T

var _ = F().(*X) // ERROR "undefined: X"

// GnoError:
// line 11: function F does not have a body but is not natively defined (did you build after pulling from the repository?)
// line 13: name not defined: X

// GnoWordingDiffers:
// line 13: gc expects: undefined: X; gno: name not defined: X

// Verdict: Clean errorcheck
