// errorcheck

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

var main = func() {} // ERROR "must be func|redefinition"

// GoTypeCheckError:
// line 9

// GoTypesOnly:
// line 9: gc expects: must be func|redefinition

// Verdict: Clean errorcheck
