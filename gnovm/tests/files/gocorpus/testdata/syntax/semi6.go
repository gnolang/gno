// errorcheck

// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type T1	// ERROR "newline in type declaration"

type T2 /* // ERROR "(semicolon.*|EOF) in type declaration" */

// GnoError:
// line 9: expected type, found newline (and 1 more errors)
// line 11: expected type, found newline
