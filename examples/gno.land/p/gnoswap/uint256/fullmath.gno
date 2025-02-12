// REF: https://github.com/Uniswap/v3-core/blob/main/contracts/libraries/FullMath.sol
package uint256

import (
	"gno.land/p/demo/ufmt"
)

const (
	MAX_UINT256 = "115792089237316195423570985008687907853269984665640564039457584007913129639935"
)

func MulDiv(
	a, b, denominator *Uint,
) *Uint {
	prod0 := Zero()
	prod1 := Zero()

	{
		mm := new(Uint).MulMod(a, b, new(Uint).Not(Zero()))
		prod0 = new(Uint).Mul(a, b)

		ltBool := mm.Lt(prod0)
		ltUint := Zero()
		if ltBool {
			ltUint = One()
		}
		prod1 = new(Uint).Sub(new(Uint).Sub(mm, prod0), ltUint)
	}

	// Handle non-overflow cases, 256 by 256 division
	if prod1.IsZero() {
		if !(denominator.Gt(Zero())) { // require(denominator > 0);
			panic(ufmt.Sprintf("uint256_MulDiv()__denominator(%s) > 0", denominator.ToString()))
		}

		result := new(Uint).Div(prod0, denominator)
		return result
	}

	// Make sure the result is less than 2**256.
	// Also prevents denominator == 0
	if !(denominator.Gt(prod1)) { // require(denominator > prod1)
		panic(ufmt.Sprintf("uint256_MulDiv()__denominator(%s) > prod1(%s)", denominator.ToString(), prod1.ToString()))
	}

	///////////////////////////////////////////////
	// 512 by 256 division.
	///////////////////////////////////////////////

	// Make division exact by subtracting the remainder from [prod1 prod0]
	// Compute remainder using mulmod
	remainder := Zero()
	remainder = new(Uint).MulMod(a, b, denominator)

	// Subtract 256 bit number from 512 bit number
	gtBool := remainder.Gt(prod0)
	gtUint := Zero()
	if gtBool {
		gtUint = One()
	}
	prod1 = new(Uint).Sub(prod1, gtUint)
	prod0 = new(Uint).Sub(prod0, remainder)

	// Factor powers of two out of denominator
	// Compute largest power of two divisor of denominator.
	// Always >= 1.
	twos := Zero()
	twos = new(Uint).And(new(Uint).Neg(denominator), denominator)

	// Divide denominator by power of two
	denominator = new(Uint).Div(denominator, twos)

	// Divide [prod1 prod0] by the factors of two
	prod0 = new(Uint).Div(prod0, twos)

	// Shift in bits from prod1 into prod0. For this we need
	// to flip `twos` such that it is 2**256 / twos.
	// If twos is zero, then it becomes one
	twos = new(Uint).Add(
		new(Uint).Div(
			new(Uint).Sub(Zero(), twos),
			twos,
		),
		One(),
	)
	prod0 = new(Uint).Or(prod0, new(Uint).Mul(prod1, twos))

	// Invert denominator mod 2**256
	// Now that denominator is an odd number, it has an inverse
	// modulo 2**256 such that denominator * inv = 1 mod 2**256.
	// Compute the inverse by starting with a seed that is correct
	// correct for four bits. That is, denominator * inv = 1 mod 2**4
	inv := Zero()
	inv = new(Uint).Mul(NewUint(3), denominator)
	inv = new(Uint).Xor(inv, NewUint(2))

	// Now use Newton-Raphson iteration to improve the precision.
	// Thanks to Hensel's lifting lemma, this also works in modular
	// arithmetic, doubling the correct bits in each step.

	inv = new(Uint).Mul(inv, new(Uint).Sub(NewUint(2), new(Uint).Mul(denominator, inv))) // inverse mod 2**8
	inv = new(Uint).Mul(inv, new(Uint).Sub(NewUint(2), new(Uint).Mul(denominator, inv))) // inverse mod 2**16
	inv = new(Uint).Mul(inv, new(Uint).Sub(NewUint(2), new(Uint).Mul(denominator, inv))) // inverse mod 2**32
	inv = new(Uint).Mul(inv, new(Uint).Sub(NewUint(2), new(Uint).Mul(denominator, inv))) // inverse mod 2**64
	inv = new(Uint).Mul(inv, new(Uint).Sub(NewUint(2), new(Uint).Mul(denominator, inv))) // inverse mod 2**128
	inv = new(Uint).Mul(inv, new(Uint).Sub(NewUint(2), new(Uint).Mul(denominator, inv))) // inverse mod 2**256

	// Because the division is now exact we can divide by multiplying
	// with the modular inverse of denominator. This will give us the
	// correct result modulo 2**256. Since the precoditions guarantee
	// that the outcome is less than 2**256, this is the final result.
	// We don't need to compute the high bits of the result and prod1
	// is no longer required.
	result := new(Uint).Mul(prod0, inv)
	return result
}

func MulDivRoundingUp(
	a, b, denominator *Uint,
) *Uint {
	result := MulDiv(a, b, denominator)

	if new(Uint).MulMod(a, b, denominator).Gt(Zero()) {
		if !(result.Lt(MustFromDecimal(MAX_UINT256))) { // require(result < MAX_UINT256)
			panic(ufmt.Sprintf("uint256_MulDivRoundingUp()__result(%s) < MAX_UINT256", result.ToString()))
		}

		result = new(Uint).Add(result, One())
	}

	return result
}

// UnsafeMath
// https://github.com/Uniswap/v3-core/blob/d8b1c635c275d2a9450bd6a78f3fa2484fef73eb/contracts/libraries/UnsafeMath.sol
func DivRoundingUp(
	x, y *Uint,
) *Uint {
	div := new(Uint).Div(x, y)
	mod := new(Uint).Mod(x, y)

	z := new(Uint).Add(div, gt(mod, Zero()))
	return z
}

func gt(x, y *Uint) *Uint {
	if x.Gt(y) {
		return One()
	}
	return Zero()
}
