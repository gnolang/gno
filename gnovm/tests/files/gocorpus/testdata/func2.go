// compile

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test function signatures.
// Compiled but not run.

package main

type t1 int
type t2 int
type t3 int

func f1(t1, t2, t3)
func f2(t1, t2, t3 bool)
func f3(t1, t2, x t3)
func f4(t1, *t3)
func (x *t1) f5(y []t2) (t1, *t3)
func f6() (int, *string)
func f7(*t2, t3)
func f8(os int) int

func f9(os int) int {
	return os
}
func f10(err error) error {
	return err
}
func f11(t1 string) string {
	return t1
}

// GnoPreprocessError:
// line 16: function f1 does not have a body but is not natively defined (did you build after pulling from the repository?)

// GoBuildError:
// line 16: missing function body
// ./main.go:17:6: missing function body
// ./main.go:18:6: missing function body
// ./main.go:19:6: missing function body
// ./main.go:20:6: missing function body
// ./main.go:21:6: missing function body
// ./main.go:22:6: missing function body
// ./main.go:23:6: missing function body

// KnownDivergence:
// compile-error-wording: both Gno and Go reject; wording/stage differ
