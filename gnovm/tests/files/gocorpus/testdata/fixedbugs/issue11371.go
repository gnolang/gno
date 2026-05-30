// errorcheck

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 11371 (cmd/compile: meaningless error message "truncated to
// integer")

package issue11371

const a int = 1.1        // ERROR "constant 1.1 truncated to integer|floating-point constant truncated to integer|truncated to int|truncated"
const b int = 1e20       // ERROR "overflows int|integer constant overflow|truncated to int|truncated"
const c int = 1 + 1e-70  // ERROR "constant truncated to integer|truncated to int|truncated"
const d int = 1 - 1e-70  // ERROR "constant truncated to integer|truncated to int|truncated"
const e int = 1.00000001 // ERROR "constant truncated to integer|truncated to int|truncated"
const f int = 0.00000001 // ERROR "constant 1e-08 truncated to integer|floating-point constant truncated to integer|truncated to int|truncated"

// GnoError:
// line 12: cannot convert untyped bigdec to integer -- 1.1 not an exact integer
// line 13: bigint overflows target kind
// line 14: cannot convert untyped bigdec to integer -- 1.0000000000000000000000000000000000000000000000000000000000000000000001 not an exact integer
// line 15: cannot convert untyped bigdec to integer -- 0.9999999999999999999999999999999999999999999999999999999999999999999999 not an exact integer
// line 16: cannot convert untyped bigdec to integer -- 1.00000001 not an exact integer
// line 17: cannot convert untyped bigdec to integer -- 1E-8 not an exact integer

// GoTypeCheckError:
// line 12: cannot use 1.1 (untyped float constant) as int value in constant declaration (truncated)
// line 13: cannot use 1e20 (untyped float constant 1e+20) as int value in constant declaration (truncated)
// line 14: cannot use 1.1 (untyped float constant) as int value in constant declaration (truncated)
// line 15: cannot use 1 - 1e-70 (untyped float constant 1) as int value in constant declaration (truncated)
// line 16: cannot use 1.00000001 (untyped float constant) as int value in constant declaration (truncated)
// line 17: cannot use 0.00000001 (untyped float constant 1e-08) as int value in constant declaration (truncated)
