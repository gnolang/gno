// errorcheck

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

// Use a different line number for each token so we can
// check that the error message appears at the correct
// position.
var _ = struct{}{ /*line :20:1*/foo /*line :21:1*/: /*line :22:1*/0 }







// ERROR "unknown field foo"

// GnoIncomplete: covered 0 of 1 markers; Gno bailed before the rest — a runnable variant is needed to exercise them

// KnownIssue:
// line 19: struct type struct{} has no field foo
