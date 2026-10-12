// errorcheck

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func main() {
	for ; ; x := 1 { // ERROR "cannot declare in post statement"
		_ = x
		break
	}
}

// GoTypeCheckError:
// line 10

// GoTypesOnly:
// line 10: gc expects: cannot declare in post statement

// Verdict: Clean errorcheck
