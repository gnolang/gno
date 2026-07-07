// compile

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package foo

type T interface {
	foo()
}

func f() (T, int)

func g(v interface{}) (interface{}, int) {
	var x int
	v, x = f()
	return v, x
}

// GnoPreprocessError:
// line 13: function f does not have a body but is not natively defined (did you build after pulling from the repository?)

// GoBuildError:
// line 13: missing function body

// KnownDivergence:
// compile-error-wording: both Gno and Go reject; wording/stage differ
