// compile

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

type C comparable

// GnoPreprocessError:
// line 9: name comparable not defined in fileset with files [issue47966.go]

// GoBuildError:
// line 9: predeclared comparable requires go1.18 or later (-lang was set to go1.17

// KnownDivergence:
// compile-error-wording: both Gno and Go reject; wording/stage differ
