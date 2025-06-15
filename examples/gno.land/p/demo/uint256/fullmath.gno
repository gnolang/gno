package uint256

import (
	"gno.land/p/demo/ufmt"
)

// MulDiv calculates floor(a * b / denominator) with full precision by
// performing a 512×256-bit division. It returns an error if the denominator is
// zero or if the intermediate results exceed the 256-bit range.
func MulDiv(a, b, denominator *Uint) (*Uint, error) {
	prod0 := Zero()
	prod1 := Zero()

	{
		// mm is the modulo multiplication result using a mask of all 1s.
		mm := new(Uint).MulMod(a, b, new(Uint).Not(Zero()))
		prod0 = new(Uint).Mul(a, b)

		ltBool := mm.Lt(prod0)
		ltUint := Zero()
		if ltBool {
			ltUint = One()
		}
		prod1 = new(Uint).Sub(new(Uint).Sub(mm, prod0), ltUint)
	}

	// Handle non-overflow cases, 256 by 256 division.
	if prod1.IsZero() {
		if !denominator.Gt(Zero()) {
			return nil, ErrDivisionByZero
		}
		result := new(Uint).Div(prod0, denominator)
		return result, nil
	}

	// Ensure the result is less than 2**256.
	// Also prevents denominator == 0.
	if !denominator.Gt(prod1) {
		return nil, ufmt.Errorf("denominator (%s) must be greater than prod1 (%s)",
			denominator.String(), prod1.String())
	}

	///////////////////////////////////////////////
	// 512 by 256 division.
	///////////////////////////////////////////////

	// Make division exact by subtracting the remainder from [prod1 prod0].
	remainder := new(Uint).MulMod(a, b, denominator)

	// Subtract 256-bit number from 512-bit number.
	gtBool := remainder.Gt(prod0)
	gtUint := Zero()
	if gtBool {
		gtUint = One()
	}
	prod1 = new(Uint).Sub(prod1, gtUint)
	prod0 = new(Uint).Sub(prod0, remainder)

	// Factor powers of two out of denominator.
	// Compute largest power of two divisor of denominator. Always >= 1.
	twos := new(Uint).And(new(Uint).Neg(denominator), denominator)

	// Divide denominator by the power of two.
	denomDiv := new(Uint).Div(denominator, twos)
	denominator = denomDiv

	// Divide [prod1 prod0] by the factors of two.
	prod0 = new(Uint).Div(prod0, twos)

	// Shift bits from prod1 into prod0.
	// For this we need to flip `twos` such that it is 2**256 / twos.
	// If twos is zero, then it becomes one.
	twos = new(Uint).Add(
		new(Uint).Div(
			new(Uint).Sub(Zero(), twos),
			twos,
		),
		One(),
	)
	prod0 = new(Uint).Or(prod0, new(Uint).Mul(prod1, twos))

	// Invert denominator mod 2**256.
	// Since denominator is now odd, it has an inverse modulo 2**256
	// such that denominator * inv = 1 mod 2**256.
	// Compute the inverse using a seed that is correct for four bits.
	inv := new(Uint).Mul(NewUint(3), denominator)
	inv = new(Uint).Xor(inv, NewUint(2))

	// Use Newton-Raphson iteration to improve the precision.
	// Each iteration doubles the correct bits.
	inv = new(Uint).Mul(inv, new(Uint).Sub(NewUint(2), new(Uint).Mul(denominator, inv))) // inverse mod 2**8
	inv = new(Uint).Mul(inv, new(Uint).Sub(NewUint(2), new(Uint).Mul(denominator, inv))) // inverse mod 2**16
	inv = new(Uint).Mul(inv, new(Uint).Sub(NewUint(2), new(Uint).Mul(denominator, inv))) // inverse mod 2**32
	inv = new(Uint).Mul(inv, new(Uint).Sub(NewUint(2), new(Uint).Mul(denominator, inv))) // inverse mod 2**64
	inv = new(Uint).Mul(inv, new(Uint).Sub(NewUint(2), new(Uint).Mul(denominator, inv))) // inverse mod 2**128
	inv = new(Uint).Mul(inv, new(Uint).Sub(NewUint(2), new(Uint).Mul(denominator, inv))) // inverse mod 2**256

	// Because the division is now exact, we can obtain the final result
	// by multiplying prod0 with the modular inverse of denominator.
	result := new(Uint).Mul(prod0, inv)
	return result, nil
}

// MulDivRoundingUp calculates ceil(a * b / denominator) with full precision.
// If a * b is not an exact multiple of the denominator, the result is incremented by one.
// It returns an error if denominator is zero or if the intermediate results exceed the 256-bit range.
func MulDivRoundingUp(a, b, denominator *Uint) (*Uint, error) {
	result, err := MulDiv(a, b, denominator)
	if err != nil {
		return nil, err
	}

	remainder := new(Uint).MulMod(a, b, denominator)
	if remainder.Gt(Zero()) {
		// Here we ensure that result + 1 does not overflow 256 bits.
		maxUint, err := FromDecimal(twoPow256Sub1)
		if err != nil {
			return nil, err
		}
		if !result.Lt(maxUint) {
			return nil, err
		}
		result = new(Uint).Add(result, One())
	}

	return result, nil
}

// DivRoundingUp performs division of x by y and rounds up the result.
// It returns an error if y is zero.
func DivRoundingUp(x, y *Uint) (*Uint, error) {
	if y.IsZero() {
		return nil, ErrDivisionByZero
	}

	div := new(Uint).Div(x, y)
	mod := new(Uint).Mod(x, y)

	// Add one if there is a remainder.
	result := new(Uint).Add(div, gt(mod, Zero()))
	return result, nil
}

func gt(x, y *Uint) *Uint {
	if x.Gt(y) {
		return One()
	}
	return Zero()
}
