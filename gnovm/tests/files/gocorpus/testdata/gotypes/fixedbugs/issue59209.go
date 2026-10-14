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
// line 10: cannot convert untyped bigdec to integer -- 1.1 not an exact integer

// GoTypeCheckError:
// line 8
// line 10

// GnoWordingDiffers:
// line 8: gc expects: invalid array length; gno: bigint overflows target kind
// line 10: gc expects: must be integer; gno: cannot convert untyped bigdec to integer -- 1.1 not an exact integer

// Verdict: Clean errorcheck
