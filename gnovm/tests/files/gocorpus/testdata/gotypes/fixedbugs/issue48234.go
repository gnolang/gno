// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

var _ = interface{
	m()
	m /* ERROR "duplicate method" */ ()
}(nil)

// GnoError:
// line 7: duplicate method m
// line 9: m redeclared in this block
// 	previous declaration at issue48234.go:9:2

// GnoWordingDiffers:
// line 9: gc expects: duplicate method; gno: m redeclared in this block
// 	previous declaration at issue48234.go:9:2

// Verdict: KnownIssue errorcheck: #6344 over-strict: lines 7
