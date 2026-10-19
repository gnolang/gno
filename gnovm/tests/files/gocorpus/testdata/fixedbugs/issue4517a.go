// errorcheck

// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

var init = 1 // ERROR "cannot declare init - must be func"

// Verdict: KnownIssue errorcheck: #6344 uncaught markers: lines 9
