// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

// declarations of main
const _, main /* ERROR "cannot declare main" */ , _ = 0, 1, 2
type main /* ERROR "cannot declare main" */ struct{}
var _, main /* ERROR "cannot declare main" */ int

// GnoError:
// line 9: main redeclared in this block
// 	previous declaration at decls5.go:8:10
// line 10: main redeclared in this block
// 	previous declaration at decls5.go:8:10

// GoTypeCheckError:
// line 8
// line 9
// line 10

// ParseBlocked:
// line 8: gc expects: cannot declare main

// GnoWordingDiffers:
// line 9: gc expects: cannot declare main; gno: main redeclared in this block
// 	previous declaration at decls5.go:8:10
// line 10: gc expects: cannot declare main; gno: main redeclared in this block
// 	previous declaration at decls5.go:8:10

// Verdict: Clean errorcheck
