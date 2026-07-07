// compile

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func main() {
	Get([]string{"a", "b"})
}

func Get(ss []string) *[2]string {
	return (*[2]string)(ss)
}

// GnoPreprocessError:
// line 14: cannot convert ss<VPBlock(1,0)> (of type []string) to type *[2]string

// KnownIssue:
// TODO: explain the Gno bug (Gno rejects code gc + go/types accept)
