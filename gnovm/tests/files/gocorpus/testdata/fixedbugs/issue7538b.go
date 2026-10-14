// compile

// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 7538: blank (_) labels handled incorrectly

package p

func f() {
_:
}

// GnoPreprocessError:
// line 13: label _ defined and not used

// Verdict: KnownIssue compile: TODO: explain the Gno bug (Gno rejects code gc + go/types accept)
