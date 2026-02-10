package int256

import (
	"gno.land/p/onbloc/uint256"
)

const divisionByZeroError = "division by zero"

// Add adds two int256 values and saves the result in z.
func (z *Int) Add(x, y *Int) *Int {
	z.value.Add(&x.value, &y.value)
	return z
}

// AddUint256 adds int256 and uint256 values and saves the result in z.
func (z *Int) AddUint256(x *Int, y *uint256.Uint) *Int {
	z.value.Add(&x.value, y)
	return z
}

// Sub subtracts two int256 values and saves the result in z.
func (z *Int) Sub(x, y *Int) *Int {
	z.value.Sub(&x.value, &y.value)
	return z
}

// SubUint256 subtracts uint256 and int256 values and saves the result in z.
func (z *Int) SubUint256(x *Int, y *uint256.Uint) *Int {
	z.value.Sub(&x.value, y)
	return z
}

// Mul multiplies two int256 values and saves the result in z.
//
// It considers the signs of the operands to determine the sign of the result.
func (z *Int) Mul(x, y *Int) *Int {
	xAbs, xSign := x.Abs(), x.Sign()
	yAbs, ySign := y.Abs(), y.Sign()

	z.value.Mul(xAbs, yAbs)

	if xSign != ySign {
		z.value.Neg(&z.value)
	}

	return z
}

// Abs returns the absolute value of z.
func (z *Int) Abs() *uint256.Uint {
	if z.Sign() >= 0 {
		return &z.value
	}

	var absValue uint256.Uint
	absValue.Sub(uint0, &z.value).Neg(&z.value)

	return &absValue
}

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
	if y.IsZero() {
		panic(divisionByZeroError)
	}

	// Step 2, 3: Calculate the absolute values of x and y
	xAbs, xSign := x.Abs(), x.Sign()
	yAbs, ySign := y.Abs(), y.Sign()

	// Step 4: Perform unsigned division on the absolute values
	z.value.Div(xAbs, yAbs)

	// Step 5: Adjust the sign of the result
	// if x and y have different signs, the result must be negative
	if xSign != ySign {
		z.value.Neg(&z.value)
	}

	return z
}

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
//	7 / 3 = 2:  00000010
//
// Step 5: Adjust sign (x and y have different signs)
//
//	-2:  00000010 -> 11111110
//	     NOT: 11111101
//	     +1:  11111110
//
// Final result: -2 (11111110 in two's complement)
//
// Note: This implementation rounds towards zero, as is standard in Go.
func (z *Int) Quo(x, y *Int) *Int {
	// Step 1: Check for division by zero
	if y.IsZero() {
		panic(divisionByZeroError)
	}

	// Step 2, 3: Calculate the absolute values of x and y
	xAbs, xSign := x.Abs(), x.Sign()
	yAbs, ySign := y.Abs(), y.Sign()

	// perform unsigned division on the absolute values
	z.value.Div(xAbs, yAbs)

	// Step 5: Adjust the sign of the result
	// if x and y have different signs, the result must be negative
	if xSign != ySign {
		z.value.Neg(&z.value)
	}

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
	if y.IsZero() {
		panic(divisionByZeroError)
	}

	// Step 2, 3
	xAbs, xSign := x.Abs(), x.Sign()
	yAbs := y.Abs()

	// Step 4: Perform unsigned division and get the remainder
	var q, r uint256.Uint
	q.DivMod(xAbs, yAbs, &r)

	// Step 5: Adjust the sign of the remainder
	if xSign < 0 {
		r.Neg(&r)
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
	if y.IsZero() {
		panic(divisionByZeroError)
	}

	// Compute the truncated division quotient
	z.Quo(x, y)

	// Compute the remainder
	r := new(Int).Rem(x, y)

	// If the remainder is negative, adjust the quotient
	if r.Sign() < 0 {
		if y.Sign() > 0 {
			z.Sub(z, NewInt(1))
		} else {
			z.Add(z, NewInt(1))
		}
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
	if y.IsZero() {
		panic(divisionByZeroError)
	}

	// Perform T-division to get the remainder
	z.Rem(x, y)

	// Adjust the remainder if necessary
	if z.Sign() >= 0 {
		return z
	}
	if y.Sign() > 0 {
		return z.Add(z, y)
	}

	return z.Sub(z, y)
}

// Sets z to the sum x + y, where z and x are uint256s and y is an int256.
//
// If the y is positive, it adds y.value to x. otherwise, it subtracts y.Abs() from x.
func AddDelta(z, x *uint256.Uint, y *Int) {
	if y.Sign() >= 0 {
		z.Add(x, &y.value)
	} else {
		z.Sub(x, y.Abs())
	}
}

// Sets z to the sum x + y, where z and x are uint256s and y is an int256.
//
// This function returns true if the addition overflows, false otherwise.
func AddDeltaOverflow(z, x *uint256.Uint, y *Int) bool {
	var overflow bool
	if y.Sign() >= 0 {
		_, overflow = z.AddOverflow(x, &y.value)
	} else {
		var absY uint256.Uint
		absY.Sub(uint0, &y.value) // absY = -y.value
		_, overflow = z.SubOverflow(x, &absY)
	}

	return overflow
}
