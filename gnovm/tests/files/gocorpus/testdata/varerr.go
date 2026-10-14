// errorcheck

// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Verify that a couple of illegal variable declarations are caught by the compiler.
// Does not compile.

package main

func main() {
	_ = asdf	// ERROR "undefined.*asdf"

	new = 1	// ERROR "use of builtin new not in function call|invalid left hand side|must be called"
}

// GnoError:
// line 13: name asdf not declared
// line 15: cannot assign to uverse new

// GoTypeCheckError:
// line 13
// line 15

// GnoWordingDiffers:
// line 13: gc expects: undefined.*asdf; gno: name asdf not declared
// line 15: gc expects: use of builtin new not in function call|invalid left hand side|must be called; gno: cannot assign to uverse new

// Verdict: Clean errorcheck
