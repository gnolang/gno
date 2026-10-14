// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func main[T /* ERROR "func main must have no type parameters" */ any]() {}

// GoTypeCheckError:
// line 7

// GoTypesOnly:
// line 7: gc expects: func main must have no type parameters

// Verdict: Clean errorcheck
