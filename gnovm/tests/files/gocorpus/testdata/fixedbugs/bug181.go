// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type T *struct {
	T;	// ERROR "embed.*pointer"
}

// GoTypeCheckError:
// line 10

// GoTypesOnly:
// line 10: gc expects: embed.*pointer

// Verdict: Clean errorcheck
