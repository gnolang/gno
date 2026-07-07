// errorcheck -complete

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func F() // ERROR "missing function body"

//go:noescape
func f() {} // ERROR "can only use //go:noescape with external func implementations"

// GnoError:
// line 9: function F does not have a body but is not natively defined (did you build after pulling from the repository?)

// UncaughtError:
// line 12: uncaught; gc expects: can only use //go:noescape with external func implementations
