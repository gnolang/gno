// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

// same const identifier declared twice should not be accepted
const none = 0  // GCCGO_ERROR "previous"
const none = 1;  // ERROR "redeclared|redef"

// GnoError:
// line 11: none redeclared in this block
// 	previous declaration at bug126.go:10:7

// GoTypeCheckError:
// line 11: none redeclared in this block
