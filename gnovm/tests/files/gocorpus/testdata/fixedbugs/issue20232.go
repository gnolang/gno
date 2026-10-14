// errorcheck -d=panic

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

const x = 6e5518446744 // ERROR "malformed constant: 6e5518446744"
const _ = x * x
const _ = 1e-1000000000
const _ = 1e+1000000000 // ERROR "malformed constant: 1e\+1000000000"

// GnoError:
// line 9: invalid decimal constant: 6e5518446744

// GoTypeCheckError:
// line 9
// line 12

// GoTypesOnly:
// line 12: gc expects: malformed constant: 1e\+1000000000

// GnoWordingDiffers:
// line 9: gc expects: malformed constant: 6e5518446744; gno: invalid decimal constant: 6e5518446744

// Verdict: Clean errorcheck
