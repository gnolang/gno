// SPDX-License-Identifier: CC-BY-4.0
// REF: https://github.com/Uniswap/solidity-lib/blob/master/contracts/libraries/FullMath.sol
package uint256

func fullMul(
	x *Uint,
	y *Uint,
) (*Uint, *Uint) { // l, h
	mm := new(Uint).MulMod(x, y, MustFromDecimal(MAX_UINT256))

	l := new(Uint).Mul(x, y)
	h := new(Uint).Sub(mm, l)

	if mm.Lt(l) {
		h = new(Uint).Sub(h, One())
	}

	return l, h
}

func fullDiv(
	l *Uint,
	h *Uint,
	d *Uint,
) *Uint {
	// uint256 pow2 = d & -d;
	// d
	_negD := new(Uint).Neg(d)
	pow2 := new(Uint).And(d, _negD)
	d = new(Uint).Div(d, pow2)
	l = new(Uint).Div(l, pow2)

	_negPow2 := new(Uint).Neg(pow2)

	value1 := new(Uint).Div(_negPow2, pow2) // (-pow2) / pow2
	value2 := new(Uint).Add(value1, One())  // (-pow2) / pow2 + 1)
	value3 := new(Uint).Mul(h, value2)      // h * ((-pow2) / pow2 + 1);
	l = new(Uint).Add(l, value3)

	r := One()
	for i := 0; i < 8; i++ {
		value1 := new(Uint).Mul(d, r)               // d * r
		value2 := new(Uint).Sub(NewUint(2), value1) // 2 - ( d * r )
		r = new(Uint).Mul(r, value2)                // r *= 2 - d * r;
	}
	res := new(Uint).Mul(l, r)
	return res
}

func MulDiv(
	x *Uint,
	y *Uint,
	d *Uint,
) *Uint {
	l, h := fullMul(x, y)
	mm := new(Uint).MulMod(x, y, d)

	if mm.Gt(l) {
		h = new(Uint).Sub(h, One())
	}
	l = new(Uint).Sub(l, mm)

	if h.IsZero() {
		return new(Uint).Div(l, d)
	}

	if !(h.Lt(d)) {
		panic("FULLDIV_OVERFLOW")
	}

	return fullDiv(l, h, d)
}

func DivRoundingUp(
	x *Uint,
	y *Uint,
) *Uint {
	div := new(Uint).Div(x, y)

	mod := new(Uint).Mod(x, y)
	return new(Uint).Add(div, gt(mod, Zero()))
}

// HELPERs
func lt(x, y *Uint) *Uint {
	if x.Lt(y) {
		return One()
	} else {
		return Zero()
	}
}

func gt(x, y *Uint) *Uint {
	if x.Gt(y) {
		return One()
	} else {
		return Zero()
	}
}
