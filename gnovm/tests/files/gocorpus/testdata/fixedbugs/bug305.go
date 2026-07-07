// errorcheck

// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Use //line to set the line number of the next line to 20.
//line fixedbugs/bug305.go:20

package p

// Introduce an error which should be reported on line 24.
var a int = "bogus"

// Line 15 of file.
// 16
// 17
// 18
// 19
// 20
// 21
// 22
// 23
// ERROR "cannot|incompatible"

// GnoStaticIncomplete: covered 0 of 1 markers (Gno preprocess: 0, go/types guard: 0); Gno's own preprocess flags none (lenient); the rest are caught by neither — a runnable variant may exercise more

// GnoOverStrictError:
// line 23: cannot use untyped string as IntKind

// KnownIssue:
// TODO: explain the Gno bug (Gno rejects lines gc + go/types accept)
