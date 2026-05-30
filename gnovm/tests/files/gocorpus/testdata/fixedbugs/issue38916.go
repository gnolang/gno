// compile

// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func f(b bool, c complex128) func(complex128) complex128 {
	return func(p complex128) complex128 {
		b = (p+1i == 0) && b
		return (p + 2i) * (p + 3i - c)
	}
}

// KnownIssue:
// line 9: 2: name complex128 not defined in fileset with files [issue38916.go]
