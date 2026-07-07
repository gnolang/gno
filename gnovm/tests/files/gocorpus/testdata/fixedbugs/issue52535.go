// compile

// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// gofrontend crashed converting unnamed bool type to any.

package p

func F() {
	m := make(map[int]int)
	var ok any
	_, ok = m[0]
	_ = ok
}

// GnoPreprocessError:
// line 14: want bool type got interface {}

// GoBuildError:
// line 13: predeclared any requires go1.18 or later (-lang was set to go1.17

// KnownDivergence:
// compile-error-wording: both Gno and Go reject; wording/stage differ
