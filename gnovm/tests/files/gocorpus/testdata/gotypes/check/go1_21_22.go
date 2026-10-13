// -lang=go1.21

// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Check Go language version-specific errors.

//go:build go1.22

package p

func f() {
	for _ = range /* ok because of upgrade to 1.22 */ 10 {
	}
}

// GnoPreprocessError:
// line 14: 3: range iteration requires map, string, array, slice, or pointer to array

// GoTypeCheckError:
// line 14

// KnownDivergence:
// compile-error-wording: both Gno and Go reject; wording/stage differ
