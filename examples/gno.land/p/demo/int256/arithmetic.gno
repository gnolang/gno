package int256

import "gno.land/p/demo/uint256"

func (z *Int) Add(x, y *Int) *Int {
	z.value.Add(&x.value, &y.value)
	return z
}

func (z *Int) AddUint256(x *Int, y *uint256.Uint) *Int {
	z.value.Add(&x.value, y)
	return z
}

func (z *Int) Sub(x, y *Int) *Int {
	z.value.Sub(&x.value, &y.value)
	return z
}

func (z *Int) SubUint256(x *Int, y *uint256.Uint) *Int {
	z.value.Sub(&x.value, y)
	return z
}

func (z *Int) Mul(x, y *Int) *Int {
	z.value.Mul(&x.value, &y.value)
	return z
}

/* ----------------------------------------------------------------------------
// Division and modulus operations
// ----------------------------------------------------------------------------
//
// This package provides three different division and modulus operations:
//  - Div and Rem: Truncated division (T-division)
//  - Quo and Mod: Floored division (F-division)
//  - DivE and ModE: Euclidean division (E-division)
//
// * Truncated division (Div, Rem) is the most common implementation in modern processors
// and programming languages. It rounds quotients towards zero and the remainder
// always has the same sign as the dividend.
//
// * Floored division (Quo, Mod) always rounds quotients towards negative infinity.
// This ensures that the modulus is always non-negative for a positive divisor,
// which can be useful in certain algorithms.
//
// * Euclidean division (DivE, ModE) ensures that the remainder is always non-negative,
// regardless of the signs of the dividend and divisor. This has several mathematical
// advantages:
//  1. It satisfies the unique division with remainder theorem.
//  2. It preserves division and modulus properties for negative divisors.
//  3. It allows for optimizations in divisions by powers of two.
// [+] Currently, ModE and Mod are shared the same implementation.
//
// ## Performance considerations:
//  - For most operations, the performance difference between these division types is negligible.
//  - Euclidean division may require an extra comparison and potentially an addition,
//    which could impact performance in extremely performance-critical scenarios.
//  - For divisions by powers of two, Euclidean division can be optimized to use
//    bitwise operations, potentially offering better performance.
//
// ## Usage guidelines:
//  - Use Div and Rem for general-purpose division that matches most common expectations.
//  - Use Quo and Mod when you need a non-negative remainder for positive divisors,
//    or when implementing algorithms that assume floored division.
//  - Use DivE and ModE when you need the mathematical properties of Euclidean division,
//    or when working with algorithms that specifically require it.
//
// Note: When working with negative numbers, be aware of the differences in behavior
// between these division types, especially at the boundaries of integer ranges.
//
// ## References
// 1. Daan Leijen, “Division and Modulus for Computer Scientists” (https://www.microsoft.com/en-us/research/wp-content/uploads/2016/02/divmodnote-letter.pdf)
/* ---------------------------------------------------------------------------- */

// Div performs integer division z = x / y and returns z.
// If y == 0, it panics with a "division by zero" error.
//
// This function handles signed division using two's complement representation:
//  1. Determine the sign of the quotient based on the signs of x and y.
//  2. Perform unsigned division on the absolute values.
//  3. Adjust the result's sign if necessary.
//
// Example visualization for 8-bit integers (scaled down from 256-bit for simplicity):
//
// Let x = -6 (11111010 in two's complement) and y = 3 (00000011)
//
// Step 2: Determine signs
//
//	x: negative (MSB is 1)
//	y: positive (MSB is 0)
//
// Step 3: Calculate absolute values
//
//	|x| = 6:  11111010 -> 00000110
//	     NOT: 00000101
//	     +1:  00000110
//
//	|y| = 3:  00000011 (already positive)
//
// Step 4: Unsigned division
//
//	6 / 3 = 2:  00000010
//
// Step 5: Adjust sign (x and y have different signs)
//
//	-2:  00000010 -> 11111110
//	     NOT: 11111101
//	     +1:  11111110
//
// Note: This implementation rounds towards zero, as is standard in Go.
func (z *Int) Div(x, y *Int) *Int {
	// Step 1: Check for division by zero
	if y.value.IsZero() {
		panic("division by zero")
	}

	// Step 2: Determine the signs of x and y
	xSign := x.Sign()
	ySign := y.Sign()

	// Step 3: Calculate the absolute values of x and y
	xAbs, xSign := abs(x)
	yAbs, ySign := abs(y)

	// perform unsigned division on the absolute values
	z.value.Div(&xAbs, &yAbs)

	// Step 5: Adjust the sign of the result
	// if x and y have different signs, the result must be negative
	if xSign != ySign {
		// convert the positive result to negative using two's complement
		setNegative(&z.value, &z.value)
	}

	return z
}

// Quo sets z to the quotient x/y for y != 0 and returns z.
// It implements truncated division (or T-division).
//
// The function performs the following steps:
//  1. Check for division by zero
//  2. Determine the signs of x and y
//  3. Calculate the absolute values of x and y
//  4. Perform unsigned division and get the remainder
//  5. Adjust the quotient for truncation towards zero
//  6. Adjust the sign of the result
//
// Example visualization for 8-bit integers (scaled down from 256-bit for simplicity):
//
// Let x = -7 (11111001 in two's complement) and y = 3 (00000011)
//
// Step 2: Determine signs
//
//	x: negative (MSB is 1)
//	y: positive (MSB is 0)
//
// Step 3: Calculate absolute values
//
//	|x| = 7:  11111001 -> 00000111
//	     NOT: 00000110
//	     +1:  00000111
//
//	|y| = 3:  00000011 (already positive)
//
// Step 4: Unsigned division
//
//	7 / 3 = 2 remainder 1
//	q = 2:  00000010
//	r = 1:  00000001
//
// Step 5: Adjust for truncation
//
//	Signs are different and r != 0, so add 1 to q
//	q = 3:  00000011
//
// Step 6: Adjust sign (x and y have different signs)
//
//	-3:  00000011 -> 11111101
//	     NOT: 11111100
//	     +1:  11111101
//
// Final result: -3 (11111101 in two's complement)
//
// Note: This implementation ensures correct truncation towards zero for negative dividends.
func (z *Int) Quo(x, y *Int) *Int {
	// Step 1: Check for division by zero
	if y.value.IsZero() {
		panic("division by zero")
	}

	// Step 2: Determine the signs of x and y
	xSign := x.Sign()
	ySign := y.Sign()

	// Step 3: Calculate the absolute values of x and y
	var xAbs, yAbs, q, r uint256.Uint
	xAbs, xSign = abs(x)
	yAbs, ySign = abs(y)

	// Step 4: Perform unsigned division and get the remainder
	//  q = xAbs / yAbs
	//  r = xAbs % yAbs
	//
	// Use Euclidean division to always yields a non-negative remainder, which is
	// crucial for correct truncation towards zero in subsequent steps.
	// By using this method here, we can easily adjust the quotient in _Step 5_
	// to achieve truncated division semantics.
	q.DivMod(&xAbs, &yAbs, &r)

	// Step 5: Adjust the quotient for truncation towards zero.
	//
	// If x and y have different signs and there's a non-zero remainder,
	// we need to round towards zero by adding 1 to the quotient magnitude.
	if (xSign < 0) != (ySign < 0) && !r.IsZero() {
		q.Add(&q, one)
	}

	// Step 6: Adjust the sign of the result
	if xSign != ySign {
		setNegative(&q, &q)
	}

	z.value.Set(&q)
	return z
}

// Rem sets z to the remainder x%y for y != 0 and returns z.
//
// The function performs the following steps:
//  1. Check for division by zero
//  2. Determine the signs of x and y
//  3. Calculate the absolute values of x and y
//  4. Perform unsigned division and get the remainder
//  5. Adjust the sign of the remainder
//
// Example visualization for 8-bit integers (scaled down from 256-bit for simplicity):
//
// Let x = -7 (11111001 in two's complement) and y = 3 (00000011)
//
// Step 2: Determine signs
//
//	x: negative (MSB is 1)
//	y: positive (MSB is 0)
//
// Step 3: Calculate absolute values
//
//	|x| = 7:  11111001 -> 00000111
//	     NOT: 00000110
//	     +1:  00000111
//
//	|y| = 3:  00000011 (already positive)
//
// Step 4: Unsigned division
//
//	7 / 3 = 2 remainder 1
//	q = 2:  00000010 (not used in result)
//	r = 1:  00000001
//
// Step 5: Adjust sign of remainder (x is negative)
//
//	-1:  00000001 -> 11111111
//	     NOT: 11111110
//	     +1:  11111111
//
// Final result: -1 (11111111 in two's complement)
//
// Note: The sign of the remainder is always the same as the sign of the dividend (x).
func (z *Int) Rem(x, y *Int) *Int {
	// Step 1: Check for division by zero
	if y.value.IsZero() {
		panic("division by zero")
	}

	// Step 2: Determine the signs of x and y
	xSign := x.Sign()
	ySign := y.Sign()

	// Step 3: Calculate the absolute values of x and y
	var xAbs, yAbs, q, r uint256.Uint
	xAbs, xSign = abs(x)
	yAbs, ySign = abs(y)

	// Step 4: Perform unsigned division and get the remainder
	q.DivMod(&xAbs, &yAbs, &r)

	// Step 5: Adjust the sign of the remainder
	if xSign < 0 {
		if !r.IsZero() {
			// convert positive remainder to negative
			setNegative(&r, &r)
		}
	}

	z.value.Set(&r)
	return z
}

// Mod sets z to the modulus x%y for y != 0 and returns z.
// The result (z) has the same sign as the divisor y.
func (z *Int) Mod(x, y *Int) *Int {
	return z.ModE(x, y)
}

// DivE performs Euclidean division of x by y, setting z to the quotient and returning z.
// If y == 0, it panics with a "division by zero" error.
//
// Euclidean division satisfies the following properties:
//  1. The remainder is always non-negative: 0 <= x mod y < |y|
//  2. It follows the identity: x = y * (x div y) + (x mod y)
func (z *Int) DivE(x, y *Int) *Int {
	if y.value.IsZero() {
		panic("division by zero")
	}

	z.Div(x, y)

	// Get the remainder using T-division
	r := new(Int).Rem(x, y)

	// Adjust the quotient if necessary
	if r.Sign() < 0 {
		if y.Sign() > 0 {
			return z.Sub(z, NewInt(1))
		}
		return z.Add(z, NewInt(1))
	}

	return z
}

// ModE computes the Euclidean modulus of x by y, setting z to the result and returning z.
// If y == 0, it panics with a "division by zero" error.
//
// The Euclidean modulus is always non-negative and satisfies:
//
//	0 <= x mod y < |y|
//
// Example visualization for 8-bit integers (scaled down from 256-bit for simplicity):
//
// Case 1: Let x = -7 (11111001 in two's complement) and y = 3 (00000011)
//
// Step 1: Compute remainder (using Rem)
//
//	Result of Rem: -1 (11111111 in two's complement)
//
// Step 2: Adjust sign (result is negative, y is positive)
//
//	-1 + 3 = 2
//	11111111 + 00000011 = 00000010
//
// Final result: 2 (00000010)
//
// Case 2: Let x = -7 (11111001 in two's complement) and y = -3 (11111101 in two's complement)
//
// Step 1: Compute remainder (using Rem)
//
//	Result of Rem: -1 (11111111 in two's complement)
//
// Step 2: Adjust sign (result is negative, y is negative)
//
//	No adjustment needed
//
// Final result: -1 (11111111 in two's complement)
//
// Note: This implementation ensures that the result always has the same sign as y,
// which is different from the Rem operation.
func (z *Int) ModE(x, y *Int) *Int {
	if y.value.IsZero() {
		panic("division by zero")
	}

	// Perform T-division to get the remainder
	z.Rem(x, y)

	// Adjust the remainder if necessary
	if z.Sign() < 0 {
		if y.Sign() > 0 {
			return z.Add(z, y)
		}
		return z.Sub(z, y)
	}

	return z
}

func abs(x *Int) (uint256.Uint, int) {
	sign := x.Sign()
	var absVal uint256.Uint
	absVal.Set(&x.value)
	if sign < 0 {
		setNegative(&absVal, &x.value)
		sign = -1
	} else {
		sign = 1
	}
	return absVal, sign
}

func setNegative(z, x *uint256.Uint) {
	z.Set(x).Not(z).Add(z, one)
}

// not implemented yet

// DivUint256 sets z to the quotient x/y, where y is a uint256, and returns z
// If y == 0, z is set to 0
// func (z *Int) DivUint256(x *Int, y *uint256.Uint) *Int {
// 	z.abs.Div(x.abs, y)
// 	if z.abs.IsZero() {
// 		z.neg = false
// 	} else {
// 		z.neg = x.neg
// 	}
// 	return z
// }

// Sets z to the sum x + y, where z and x are uint256s and y is an int256.
// func AddDelta(z, x *uint256.Uint, y *Int) {
// 	if y.neg {
// 		z.Sub(x, y.abs)
// 	} else {
// 		z.Add(x, y.abs)
// 	}
// }

// Sets z to the sum x + y, where z and x are uint256s and y is an int256.
// func AddDeltaOverflow(z, x *uint256.Uint, y *Int) bool {
// 	var overflow bool
// 	if y.neg {
// 		_, overflow = z.SubOverflow(x, y.abs)
// 	} else {
// 		_, overflow = z.AddOverflow(x, y.abs)
// 	}
// 	return overflow
// }
