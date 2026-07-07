// compile

// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

var (
	x  int
	xs []int
)

func a([]int) (int, error)

func b() (int, error) {
	return a(append(xs, x))
}

func c(int, error) (int, error)

func d() (int, error) {
	return c(b())
}

// GnoPreprocessError:
// line 14: function a does not have a body but is not natively defined (did you build after pulling from the repository?)

// GoBuildError:
// line 14: missing function body
// ./main.go:21:6: missing function body

// KnownDivergence:
// compile-error-wording: both Gno and Go reject; wording/stage differ
