// errorcheck

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

type T interface{ M() }

func F() T

var _ = F().(*X) // ERROR "undefined: X"

// GnoIncomplete: covered 0 of 1 markers; Gno bailed before the rest — a runnable variant is needed to exercise them
// GnoError:
// line 11: function F does not have a body but is not natively defined (did you build after pulling from the repository?)
