// errorcheck

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Check that we don't ignore EOF.

package p

var f = func() { // ERROR "unexpected EOF|expected .*}.*"

// GnoStaticIncomplete: covered 0 of 1 markers (Gno preprocess: 0, go/types guard: 0); Gno's own preprocess flags none (lenient); the rest are caught by neither — a runnable variant may exercise more

// GnoOverStrictError:
// line 13: expected '(', found main

// KnownIssue:
// TODO: explain the Gno bug (Gno rejects lines gc + go/types accept)
