// errorcheck

// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 6402: spurious 'use of untyped nil' error

package p

func f() uintptr {
	return nil // ERROR "cannot use nil as type uintptr in return argument|incompatible type|cannot use nil"
}

// GnoIncomplete: covered 0 of 1 markers; Gno bailed before the rest — a runnable variant is needed to exercise them
// GnoError:
// line 11: 2: name uintptr not defined in fileset with files [issue6402.go]
