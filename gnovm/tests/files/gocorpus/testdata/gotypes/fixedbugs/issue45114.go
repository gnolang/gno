// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

var s uint
var _ = string(1 /* ERRORx `shifted operand 1 .* must be integer` */ << s)

// GnoError:
// line 8: operator << not defined on: StringKind

// GnoWordingDiffers:
// line 8: gc expects: shifted operand 1 .* must be integer; gno: operator << not defined on: StringKind

// Verdict: Clean errorcheck
