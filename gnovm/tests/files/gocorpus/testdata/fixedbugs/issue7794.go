// compile

// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func main() {
	var a [10]int
	const ca = len(a)
}

// GnoPreprocessError:
// line 10: declared and not used: a

// Verdict: KnownIssue compile: TODO: explain the Gno bug (Gno rejects code gc + go/types accept)
