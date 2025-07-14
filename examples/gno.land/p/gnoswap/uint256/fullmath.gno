// REF: https://github.com/Uniswap/v3-core/blob/main/contracts/libraries/FullMath.sol

// Package fullmath implements Uniswap V3's FullMath library.
//
// This library provides advanced fixed-point math operations that are essential
// for Uniswap V3's tick math and liquidity calculations. It enables precise
// calculations of (a * b / denominator) with full 512-bit intermediate precision.
//
// NOTE: Unlike other arithmetic functions in the uint256 package that return errors,
// functions in this file panic on invalid inputs to maintain behavioral compatibility
// with the original Solidity implementation which uses require() statements.
//
// This design choice is intentional because:
// 1. These functions are typically used in hot paths where error handling would add overhead
// 2. Invalid inputs (like zero denominator) represent programming errors, not runtime conditions
// 3. Staying close to the Solidity implementation makes protocol porting more reliable
//
// If you need error-returning versions, wrap these functions with appropriate error handling.
package uint256

import (
	"gno.land/p/demo/ufmt"
)

// MulDiv calculates (a * b) / denominator with full 512-bit intermediate precision.
// This function is crucial for maintaining precision in Uniswap V3's tick math.
//
// The implementation uses Remco Bloemen's multiplication and division algorithm
// which computes the 512-bit product and divides it by the denominator.
//
// Parameters:
//   - a: First multiplicand
//   - b: Second multiplicand
//   - denominator: The divisor (must be greater than 0)
//
// Returns:
//   - The result of (a * b) / denominator, rounded down
//
// Panics:
//   - If denominator is 0
//   - If the result would overflow 256 bits (i.e., if a * b / denominator > 2^256 - 1)
func MulDiv(a, b, denominator *Uint) *Uint {
	// Compute the 512-bit product of a * b
	p := umul(a, b)

	// Split the 512-bit product into high and low 256-bit parts
	lo := &Uint{arr: [4]uint64{p[0], p[1], p[2], p[3]}}
	hi := &Uint{arr: [4]uint64{p[4], p[5], p[6], p[7]}}

	// If the high part is zero, we can use simple 256-bit division
	if hi.IsZero() {
		if denominator.IsZero() {
			panic(ufmt.Sprintf("denominator must be greater than 0. got %s", denominator.ToString()))
		}
		result := new(Uint).Div(lo, denominator)
		return result
	}

	// Ensure the result will fit in 256 bits
	// If hi >= denominator, then (hi * 2^256 + lo) / denominator >= 2^256
	if denominator.Lte(hi) {
		panic(ufmt.Sprintf("overflow: denominator(%s) must be greater than hi(%s)", denominator.ToString(), hi.ToString()))
	}

	// Compute the remainder of (a * b) mod denominator
	remainder := new(Uint).MulMod(a, b, denominator)

	// Subtract remainder from [hi, lo] to make it divisible by denominator
	// If remainder > lo, we need to borrow from hi
	if remainder.Gt(lo) {
		hi = new(Uint).Sub(hi, one)
	}
	lo = new(Uint).Sub(lo, remainder)

	// Factor out powers of 2 from denominator
	// This optimization leverages the fact that division by powers of 2 is just bit shifting
	twos := new(Uint).And(new(Uint).Neg(denominator), denominator)
	denominator = new(Uint).Div(denominator, twos)
	lo = new(Uint).Div(lo, twos)

	// Find the multiplicative inverse of the odd denominator modulo 2^256
	// Using Newton-Raphson iteration to find the inverse
	// Start with a good initial guess: inverse = 3 * denominator XOR 2
	inv := Zero()
	inv = new(Uint).Mul(three, denominator)
	inv = new(Uint).Xor(inv, two)

	// Newton-Raphson iteration: inv = inv * (2 - denominator * inv)
	// Each iteration doubles the precision
	inv = new(Uint).Mul(inv, new(Uint).Sub(two, new(Uint).Mul(denominator, inv))) // inverse mod 2^8
	inv = new(Uint).Mul(inv, new(Uint).Sub(two, new(Uint).Mul(denominator, inv))) // inverse mod 2^16
	inv = new(Uint).Mul(inv, new(Uint).Sub(two, new(Uint).Mul(denominator, inv))) // inverse mod 2^32
	inv = new(Uint).Mul(inv, new(Uint).Sub(two, new(Uint).Mul(denominator, inv))) // inverse mod 2^64
	inv = new(Uint).Mul(inv, new(Uint).Sub(two, new(Uint).Mul(denominator, inv))) // inverse mod 2^128
	inv = new(Uint).Mul(inv, new(Uint).Sub(two, new(Uint).Mul(denominator, inv))) // inverse mod 2^256

	// Compute the quotient
	// 2^256 / twos = (2^256 - 1) / twos + 1
	// negTwos = -twos = 2^256 - twos (mod 2^256)
	// divResult = negTwos / twos = (2^256 - twos) / twos = 2^256/twos - 1
	// twos = divResult + 1 = 2^256/twos
	negTwos := new(Uint).Sub(Zero(), twos)
	divResult := new(Uint).Div(negTwos, twos)
	twos = new(Uint).Add(divResult, one)

	lo = new(Uint).Or(lo, new(Uint).Mul(hi, twos))

	result := new(Uint).Mul(lo, inv)
	return result
}

// MulDivRoundingUp calculates ceil((a * b) / denominator).
// This is useful for ensuring that rounding errors favor the protocol
// rather than users, which is important for security.
//
// Parameters:
//   - a: First multiplicand
//   - b: Second multiplicand
//   - denominator: The divisor (must be greater than 0)
//
// Returns:
//   - The result of (a * b) / denominator, rounded up
//
// Panics:
//   - If denominator is 0
//   - If the result would overflow 256 bits
func MulDivRoundingUp(a, b, denominator *Uint) *Uint {
	result := MulDiv(a, b, denominator)

	// Check if there's a remainder
	mulModResult := new(Uint).MulMod(a, b, denominator)

	// If there's no remainder, return the result as-is
	if mulModResult.IsZero() {
		return result
	}

	// Add 1 to round up, but check for overflow
	if result.Eq(MustFromDecimal(MAX_UINT256)) {
		panic(ufmt.Sprintf("overflow: result(%s) + 1 would exceed MAX_UINT256", result.ToString()))
	}

	return new(Uint).Add(result, one)
}

// DivRoundingUp performs division with rounding up.
// This is a simpler operation that doesn't require the full precision of MulDiv.
//
// Parameters:
//   - x: Dividend
//   - y: Divisor (must be greater than 0)
//
// Returns:
//   - The result of x / y, rounded up
//
// Panics:
//   - If y is 0
func DivRoundingUp(x, y *Uint) *Uint {
	if y.IsZero() {
		panic("division by zero")
	}
	div := new(Uint).Div(x, y)
	mod := new(Uint).Mod(x, y)
	z := new(Uint).Add(div, gt(mod, Zero()))
	return z
}

// gt returns one if x is greater than y, zero otherwise.
// This is a branchless helper for conditional increment.
//
// Parameters:
//   - x: First number to compare
//   - y: Second number to compare
//
// Returns:
//   - one if x > y, zero otherwise
func gt(x, y *Uint) *Uint {
	if x.Gt(y) {
		return one
	}
	return Zero()
}
