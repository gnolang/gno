// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

type (
	_ [1 /* ERROR "invalid array length" */ << 100]int
	_ [1.0]int
	_ [1.1 /* ERROR "must be integer" */ ]int
)

// GnoError:
// line 8: bigint overflows target kind
// line 10: array length 1.1 (untyped float constant) must be integer

// GnoWordingDiffers:
// line 8: gc expects: invalid array length; gno: bigint overflows target kind

// Verdict: Clean errorcheck
