// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

const s float32 = 0
var _ = 0 << s

// GnoError:
// line 8: invalid operation: invalid shift count: (const (0 float32))

// Verdict: KnownIssue compile: #6344 an untyped float constant shift count representable as uint is rejected
