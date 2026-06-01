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

// GnoStaticIncomplete: covered 0 of 1 markers (Gno preprocess: 0, go/types guard: 0); Gno's own preprocess flags none (lenient); the rest are caught by neither — a runnable variant may exercise more

// KnownIssue:
// line 19: struct type struct{} has no field foo
