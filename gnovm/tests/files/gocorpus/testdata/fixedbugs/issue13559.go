// errorcheck

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Verify that error messages print meaningful values
// for various extreme floating-point constants.

package p

// failure case in issue
const _ int64 = 1e-10000 // ERROR "1e\-10000 truncated|.* truncated to int64|truncated"

const (
	_ int64 = 1e10000000 // ERROR "integer too large|truncated to int64|truncated"
	_ int64 = 1e1000000  // ERROR "integer too large|truncated to int64|truncated"
	_ int64 = 1e100000   // ERROR "integer too large|truncated to int64|truncated"
	_ int64 = 1e10000    // ERROR "integer too large|truncated to int64|truncated"
	_ int64 = 1e1000     // ERROR "integer too large|truncated to int64|truncated"
	_ int64 = 1e100      // ERROR "10000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000 overflows|truncated to int64|truncated"
	_ int64 = 1e10
	_ int64 = 1e1
	_ int64 = 1e0
	_ int64 = 1e-1       // ERROR "0\.1 truncated|.* truncated to int64|truncated"
	_ int64 = 1e-10      // ERROR "1e\-10 truncated|.* truncated to int64|truncated"
	_ int64 = 1e-100     // ERROR "1e\-100 truncated|.* truncated to int64|truncated"
	_ int64 = 1e-1000    // ERROR "1e\-1000 truncated|.* truncated to int64|truncated"
	_ int64 = 1e-10000   // ERROR "1e\-10000 truncated|.* truncated to int64|truncated"
	_ int64 = 1e-100000  // ERROR "1e\-100000 truncated|.* truncated to int64|truncated"
	_ int64 = 1e-1000000 // ERROR "1e\-1000000 truncated|.* truncated to int64|truncated"
)

const (
	_ int64 = -1e10000000 // ERROR "integer too large|truncated to int64|truncated"
	_ int64 = -1e1000000  // ERROR "integer too large|truncated to int64|truncated"
	_ int64 = -1e100000   // ERROR "integer too large|truncated to int64|truncated"
	_ int64 = -1e10000    // ERROR "integer too large|truncated to int64|truncated"
	_ int64 = -1e1000     // ERROR "integer too large|truncated to int64|truncated"
	_ int64 = -1e100      // ERROR "\-10000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000 overflows|truncated to int64|truncated"
	_ int64 = -1e10
	_ int64 = -1e1
	_ int64 = -1e0
	_ int64 = -1e-1       // ERROR "\-0\.1 truncated|.* truncated to int64|truncated"
	_ int64 = -1e-10      // ERROR "\-1e\-10 truncated|.* truncated to int64|truncated"
	_ int64 = -1e-100     // ERROR "\-1e\-100 truncated|.* truncated to int64|truncated"
	_ int64 = -1e-1000    // ERROR "\-1e\-1000 truncated|.* truncated to int64|truncated"
	_ int64 = -1e-10000   // ERROR "\-1e\-10000 truncated|.* truncated to int64|truncated"
	_ int64 = -1e-100000  // ERROR "\-1e\-100000 truncated|.* truncated to int64|truncated"
	_ int64 = -1e-1000000 // ERROR "\-1e\-1000000 truncated|.* truncated to int64|truncated"
)

const (
	_ int64 = 1.23456789e10000000 // ERROR "integer too large|truncated to int64|truncated"
	_ int64 = 1.23456789e1000000  // ERROR "integer too large|truncated to int64|truncated"
	_ int64 = 1.23456789e100000   // ERROR "integer too large|truncated to int64|truncated"
	_ int64 = 1.23456789e10000    // ERROR "integer too large|truncated to int64|truncated"
	_ int64 = 1.23456789e1000     // ERROR "integer too large|truncated to int64|truncated"
	_ int64 = 1.23456789e100      // ERROR "12345678900000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000 overflows|truncated to int64|truncated"
	_ int64 = 1.23456789e10
	_ int64 = 1.23456789e1        // ERROR "12\.3457 truncated|.* truncated to int64|truncated"
	_ int64 = 1.23456789e0        // ERROR "1\.23457 truncated|.* truncated to int64|truncated"
	_ int64 = 1.23456789e-1       // ERROR "0\.123457 truncated|.* truncated to int64|truncated"
	_ int64 = 1.23456789e-10      // ERROR "1\.23457e\-10 truncated|.* truncated to int64|truncated"
	_ int64 = 1.23456789e-100     // ERROR "1\.23457e\-100 truncated|.* truncated to int64|truncated"
	_ int64 = 1.23456789e-1000    // ERROR "1\.23457e\-1000 truncated|.* truncated to int64|truncated"
	_ int64 = 1.23456789e-10000   // ERROR "1\.23457e\-10000 truncated|.* truncated to int64|truncated"
	_ int64 = 1.23456789e-100000  // ERROR "1\.23457e\-100000 truncated|.* truncated to int64|truncated"
	_ int64 = 1.23456789e-1000000 // ERROR "1\.23457e\-1000000 truncated|.* truncated to int64|truncated"
)

const (
	_ int64 = -1.23456789e10000000 // ERROR "integer too large|truncated to int64|truncated"
	_ int64 = -1.23456789e1000000  // ERROR "integer too large|truncated to int64|truncated"
	_ int64 = -1.23456789e100000   // ERROR "integer too large|truncated to int64|truncated"
	_ int64 = -1.23456789e10000    // ERROR "integer too large|truncated to int64|truncated"
	_ int64 = -1.23456789e1000     // ERROR "integer too large|truncated to int64|truncated"
	_ int64 = -1.23456789e100      // ERROR "\-12345678900000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000 overflows|truncated to int64|truncated"
	_ int64 = -1.23456789e10
	_ int64 = -1.23456789e1        // ERROR "\-12\.3457 truncated|.* truncated to int64|truncated"
	_ int64 = -1.23456789e0        // ERROR "\-1\.23457 truncated|.* truncated to int64|truncated"
	_ int64 = -1.23456789e-1       // ERROR "\-0\.123457 truncated|.* truncated to int64|truncated"
	_ int64 = -1.23456789e-10      // ERROR "\-1\.23457e\-10 truncated|.* truncated to int64|truncated"
	_ int64 = -1.23456789e-100     // ERROR "\-1\.23457e\-100 truncated|.* truncated to int64|truncated"
	_ int64 = -1.23456789e-1000    // ERROR "\-1\.23457e\-1000 truncated|.* truncated to int64|truncated"
	_ int64 = -1.23456789e-10000   // ERROR "\-1\.23457e\-10000 truncated|.* truncated to int64|truncated"
	_ int64 = -1.23456789e-100000  // ERROR "\-1\.23457e\-100000 truncated|.* truncated to int64|truncated"
	_ int64 = -1.23456789e-1000000 // ERROR "\-1\.23457e\-1000000 truncated|.* truncated to int64|truncated"
)

// GnoError:
// line 13: cannot convert untyped bigdec to integer -- 1E-10000 not an exact integer
// line 16: invalid decimal constant: 1e10000000
// line 17: invalid decimal constant: 1e1000000
// line 18: bigint overflows target kind
// line 19: bigint overflows target kind
// line 20: bigint overflows target kind
// line 21: bigint overflows target kind
// line 25: cannot convert untyped bigdec to integer -- 0.1 not an exact integer
// line 26: cannot convert untyped bigdec to integer -- 1E-10 not an exact integer
// line 27: cannot convert untyped bigdec to integer -- 1E-100 not an exact integer
// line 28: cannot convert untyped bigdec to integer -- 1E-1000 not an exact integer
// line 29: cannot convert untyped bigdec to integer -- 1E-10000 not an exact integer
// line 30: cannot convert untyped bigdec to integer -- 1E-100000 not an exact integer
// line 31: invalid decimal constant: 1e-1000000
// line 35: invalid decimal constant: 1e10000000
// line 36: invalid decimal constant: 1e1000000
// line 37: bigint underflows target kind
// line 38: bigint underflows target kind
// line 39: bigint underflows target kind
// line 40: bigint underflows target kind
// line 44: cannot convert untyped bigdec to integer -- -0.1 not an exact integer
// line 45: cannot convert untyped bigdec to integer -- -1E-10 not an exact integer
// line 46: cannot convert untyped bigdec to integer -- -1E-100 not an exact integer
// line 47: cannot convert untyped bigdec to integer -- -1E-1000 not an exact integer
// line 48: cannot convert untyped bigdec to integer -- -1E-10000 not an exact integer
// line 49: cannot convert untyped bigdec to integer -- -1E-100000 not an exact integer
// line 50: invalid decimal constant: 1e-1000000
// line 54: invalid decimal constant: 1.23456789e10000000
// line 55: invalid decimal constant: 1.23456789e1000000
// line 56: bigint overflows target kind
// line 57: bigint overflows target kind
// line 58: bigint overflows target kind
// line 59: bigint overflows target kind
// line 61: cannot convert untyped bigdec to integer -- 12.3456789 not an exact integer
// line 62: cannot convert untyped bigdec to integer -- 1.23456789 not an exact integer
// line 63: cannot convert untyped bigdec to integer -- 0.123456789 not an exact integer
// line 64: cannot convert untyped bigdec to integer -- 1.23456789E-10 not an exact integer
// line 65: cannot convert untyped bigdec to integer -- 1.23456789E-100 not an exact integer
// line 66: cannot convert untyped bigdec to integer -- 1.23456789E-1000 not an exact integer
// line 67: cannot convert untyped bigdec to integer -- 1.23456789E-10000 not an exact integer
// line 68: invalid decimal constant: 1.23456789e-100000
// line 69: invalid decimal constant: 1.23456789e-1000000
// line 73: invalid decimal constant: 1.23456789e10000000
// line 74: invalid decimal constant: 1.23456789e1000000
// line 75: bigint underflows target kind
// line 76: bigint underflows target kind
// line 77: bigint underflows target kind
// line 78: bigint underflows target kind
// line 80: cannot convert untyped bigdec to integer -- -12.3456789 not an exact integer
// line 81: cannot convert untyped bigdec to integer -- -1.23456789 not an exact integer
// line 82: cannot convert untyped bigdec to integer -- -0.123456789 not an exact integer
// line 83: cannot convert untyped bigdec to integer -- -1.23456789E-10 not an exact integer
// line 84: cannot convert untyped bigdec to integer -- -1.23456789E-100 not an exact integer
// line 85: cannot convert untyped bigdec to integer -- -1.23456789E-1000 not an exact integer
// line 86: cannot convert untyped bigdec to integer -- -1.23456789E-10000 not an exact integer
// line 87: invalid decimal constant: 1.23456789e-100000
// line 88: invalid decimal constant: 1.23456789e-1000000

// GoTypeCheckError:
// line 13: cannot use 1e-10000 (untyped float constant) as int64 value in constant declaration (truncated)
// line 16: cannot use 1e-10000 (untyped float constant) as int64 value in constant declaration (truncated)
// line 17: cannot use 1e1000000 (untyped float constant 1e+1000000) as int64 value in constant declaration (truncated)
// line 18: cannot use 1e100000 (untyped float constant 1e+100000) as int64 value in constant declaration (truncated)
// line 19: cannot use 1e10000 (untyped float constant 1e+10000) as int64 value in constant declaration (truncated)
// line 20: cannot use 1e1000 (untyped float constant 1e+1000) as int64 value in constant declaration (truncated)
// line 21: cannot use 1e100 (untyped float constant 1e+100) as int64 value in constant declaration (truncated)
// line 25: cannot use 1e-1 (untyped float constant 0.1) as int64 value in constant declaration (truncated)
// line 26: cannot use 1e-10 (untyped float constant) as int64 value in constant declaration (truncated)
// line 27: cannot use 1e-100 (untyped float constant) as int64 value in constant declaration (truncated)
// line 28: cannot use 1e-1000 (untyped float constant) as int64 value in constant declaration (truncated)
// line 29: cannot use 1e-10000 (untyped float constant) as int64 value in constant declaration (truncated)
// line 30: cannot use 1e-100000 (untyped float constant) as int64 value in constant declaration (truncated)
// line 31: cannot use 1e-1000000 (untyped float constant) as int64 value in constant declaration (truncated)
// line 35: cannot use -1e10000000 (untyped float constant -1e+10000000) as int64 value in constant declaration (truncated)
// line 36: cannot use -1e1000000 (untyped float constant -1e+1000000) as int64 value in constant declaration (truncated)
// line 37: cannot use -1e100000 (untyped float constant -1e+100000) as int64 value in constant declaration (truncated)
// line 38: cannot use -1e10000 (untyped float constant -1e+10000) as int64 value in constant declaration (truncated)
// line 39: cannot use -1e1000 (untyped float constant -1e+1000) as int64 value in constant declaration (truncated)
// line 40: cannot use -1e100 (untyped float constant -1e+100) as int64 value in constant declaration (truncated)
// line 44: cannot use -1e-1 (untyped float constant -0.1) as int64 value in constant declaration (truncated)
// line 45: cannot use -1e-10 (untyped float constant) as int64 value in constant declaration (truncated)
// line 46: cannot use -1e-100 (untyped float constant) as int64 value in constant declaration (truncated)
// line 47: cannot use -1e-1000 (untyped float constant) as int64 value in constant declaration (truncated)
// line 48: cannot use -1e-10000 (untyped float constant) as int64 value in constant declaration (truncated)
// line 49: cannot use -1e-100000 (untyped float constant) as int64 value in constant declaration (truncated)
// line 50: cannot use -1e-1000000 (untyped float constant) as int64 value in constant declaration (truncated)
// line 54: cannot use 1.23456789e10000000 (untyped float constant 1.23457e+10000000) as int64 value in constant declaration (truncated)
// line 55: cannot use 1.23456789e1000000 (untyped float constant 1.23457e+1000000) as int64 value in constant declaration (truncated)
// line 56: cannot use 1.23456789e100000 (untyped float constant 1.23457e+100000) as int64 value in constant declaration (truncated)
// line 57: cannot use 1.23456789e10000 (untyped float constant 1.23457e+10000) as int64 value in constant declaration (truncated)
// line 58: cannot use 1.23456789e1000 (untyped float constant 1.23457e+1000) as int64 value in constant declaration (truncated)
// line 59: cannot use 1.23456789e100 (untyped float constant 1.23457e+100) as int64 value in constant declaration (truncated)
// line 61: cannot use 1.23456789e1 (untyped float constant 12.3457) as int64 value in constant declaration (truncated)
// line 62: cannot use 1.23456789e0 (untyped float constant 1.23457) as int64 value in constant declaration (truncated)
// line 63: cannot use 1.23456789e-1 (untyped float constant 0.123457) as int64 value in constant declaration (truncated)
// line 64: cannot use 1.23456789e-10 (untyped float constant 1.23457e-10) as int64 value in constant declaration (truncated)
// line 65: cannot use 1.23456789e-100 (untyped float constant 1.23457e-100) as int64 value in constant declaration (truncated)
// line 66: cannot use 1.23456789e-1000 (untyped float constant 1.23457e-1000) as int64 value in constant declaration (truncated)
// line 67: cannot use 1.23456789e-10000 (untyped float constant 1.23457e-10000) as int64 value in constant declaration (truncated)
// line 68: cannot use 1.23456789e-100000 (untyped float constant 1.23457e-100000) as int64 value in constant declaration (truncated)
// line 69: cannot use 1.23456789e-1000000 (untyped float constant 1.23457e-1000000) as int64 value in constant declaration (truncated)
// line 73: cannot use -1.23456789e10000000 (untyped float constant -1.23457e+10000000) as int64 value in constant declaration (truncated)
// line 74: cannot use -1.23456789e1000000 (untyped float constant -1.23457e+1000000) as int64 value in constant declaration (truncated)
// line 75: cannot use -1.23456789e100000 (untyped float constant -1.23457e+100000) as int64 value in constant declaration (truncated)
// line 76: cannot use -1.23456789e10000 (untyped float constant -1.23457e+10000) as int64 value in constant declaration (truncated)
// line 77: cannot use -1.23456789e1000 (untyped float constant -1.23457e+1000) as int64 value in constant declaration (truncated)
// line 78: cannot use -1.23456789e100 (untyped float constant -1.23457e+100) as int64 value in constant declaration (truncated)
// line 80: cannot use -1.23456789e1 (untyped float constant -12.3457) as int64 value in constant declaration (truncated)
// line 81: cannot use -1.23456789e0 (untyped float constant -1.23457) as int64 value in constant declaration (truncated)
// line 82: cannot use -1.23456789e-1 (untyped float constant -0.123457) as int64 value in constant declaration (truncated)
// line 83: cannot use -1.23456789e-10 (untyped float constant -1.23457e-10) as int64 value in constant declaration (truncated)
// line 84: cannot use -1.23456789e-100 (untyped float constant -1.23457e-100) as int64 value in constant declaration (truncated)
// line 85: cannot use -1.23456789e-1000 (untyped float constant -1.23457e-1000) as int64 value in constant declaration (truncated)
// line 86: cannot use -1.23456789e-10000 (untyped float constant -1.23457e-10000) as int64 value in constant declaration (truncated)
// line 87: cannot use -1.23456789e-100000 (untyped float constant -1.23457e-100000) as int64 value in constant declaration (truncated)
// line 88: cannot use -1.23456789e-1000000 (untyped float constant -1.23457e-1000000) as int64 value in constant declaration (truncated)
