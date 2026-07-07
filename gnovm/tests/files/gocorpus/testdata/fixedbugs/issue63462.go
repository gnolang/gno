// compile

// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func f() {
	for b := "" < join([]string{}, "") && true; ; {
		_ = b
	}
}

//go:noinline
func join(elems []string, sep string) string {
	return ""
}

// GnoPreprocessError:
// line 10: 3: cannot convert b.loopvar<VPBlock(1,0)> (of type <untyped> bool) to type bool

// KnownIssue:
// TODO: explain the Gno bug (Gno rejects code gc + go/types accept)
