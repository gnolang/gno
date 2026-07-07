// errorcheck

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test that compiler directives are ignored if they
// don't start at the beginning of the line.

package p

//line issue18393.go:20
import 42 // error on line 20


/* //line not at start of line: ignored */ //line issue18393.go:30
var x     // error on line 24, not 30


// ERROR "import path must be a string"



// ERROR "syntax error: unexpected newline, expecting type|expected type"

// GnoStaticIncomplete: covered 0 of 2 markers (Gno preprocess: 0, go/types guard: 0); Gno's own preprocess flags none (lenient); the rest are caught by neither — a runnable variant may exercise more

// GnoOverStrictError:
// line 19: import path must be a string (and 1 more errors)

// KnownIssue:
// TODO: explain the Gno bug (Gno rejects lines gc + go/types accept)
