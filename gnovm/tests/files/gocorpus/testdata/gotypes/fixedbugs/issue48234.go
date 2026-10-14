// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

var _ = interface{
	m()
	m /* ERROR "duplicate method" */ ()
}(nil)

// GnoError:
// line 9: m redeclared in this block
// 	previous declaration at issue48234.go:9:2

// GoTypeCheckError:
// line 8
// line 9

// GnoWordingDiffers:
// line 9: gc expects: duplicate method; gno: m redeclared in this block
// 	previous declaration at issue48234.go:9:2

// Verdict: Clean errorcheck
