// errorcheck

// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func main() {
	for x; y; z	// ERROR "expected .*{.* after for clause|undefined"
	{
		z	// GCCGO_ERROR "undefined"

// GnoError:
// line 10: expected '{', found newline (and 1 more errors)
// line 12: expected '}', found 'EOF'

// GoTypeCheckError:
// line 10: expected '{', found newline (and 1 more errors)

// GnoOverStrictError:
// line 12: expected '}', found 'EOF'
