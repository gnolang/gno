// compile

// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

type Any any
type IntOrBool interface{ int | bool }

type I interface{ Any | IntOrBool }

var (
	X I = 42
	Y I = "xxx"
	Z I = true
)

type A interface{ *B | int }
type B interface{ A | any }

// GnoPreprocessError:
// line 10: operator | not defined on: TypeKind

// GoBuildError:
// line 9: predeclared any requires go1.18 or later (-lang was set to go1.17
// line 10: embedding interface element int | bool requires go1.18 or later (-lang was set to go1.17
// line 12: embedding interface element Any | IntOrBool requires go1.18 or later (-lang was set to go1.17
// line 20: embedding interface element *B | int requires go1.18 or later (-lang was set to go1.17
// line 21: predeclared any requires go1.18 or later (-lang was set to go1.17

// KnownDivergence:
// compile-error-wording: both Gno and Go reject; wording/stage differ
