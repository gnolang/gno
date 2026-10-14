// compile -d=ssa/check/on

// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func main() {
_:
_:
}

// GnoPreprocessError:
// line 12: label _ defined and not used

// Verdict: KnownIssue compile: TODO: explain the Gno bug (Gno rejects code gc + go/types accept)
