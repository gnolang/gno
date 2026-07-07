// compile

// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

type I1 interface {
	int | any
}

type I2 interface {
	int | any
}

// GnoPreprocessError:
// line 10: operator | not defined on: TypeKind

// GoBuildError:
// line 10: predeclared any requires go1.18 or later (-lang was set to go1.17
// line 14: predeclared any requires go1.18 or later (-lang was set to go1.17

// KnownDivergence:
// compile-error-wording: both Gno and Go reject; wording/stage differ
