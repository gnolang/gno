// compile

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 28390/28430: Function call arguments were not
// converted correctly under some circumstances.

package main

func g(_ interface{}, e error)
func h() (int, error)

func f() {
	g(h())
}

// GnoPreprocessError:
// line 12: function g does not have a body but is not natively defined (did you build after pulling from the repository?)

// GoBuildError:
// line 12: missing function body
// ./main.go:13:6: missing function body

// KnownDivergence:
// compile-error-wording: both Gno and Go reject; wording/stage differ
