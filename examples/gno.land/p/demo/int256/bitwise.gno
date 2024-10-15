package int256

import (
	"gno.land/p/demo/uint256"
)

// Or sets z = x | y and returns z.
func (z *Int) Or(x, y *Int) *Int {
	if x.neg == y.neg {
		if x.neg {
			// (-x) | (-y) == ^(x-1) | ^(y-1) == ^((x-1) & (y-1)) == -(((x-1) & (y-1)) + 1)
			x1 := new(uint256.Uint).Sub(x.abs, one)
			y1 := new(uint256.Uint).Sub(y.abs, one)
			z.abs = z.abs.Add(z.abs.And(x1, y1), one)
			z.neg = true // z cannot be zero if x and y are negative
			return z
		}

		// x | y == x | y
		z.abs = z.abs.Or(x.abs, y.abs)
		z.neg = false
		return z
	}

	// x.neg != y.neg
	if x.neg {
		x, y = y, x // | is symmetric
	}

	// x | (-y) == x | ^(y-1) == ^((y-1) &^ x) == -(^((y-1) &^ x) + 1)
	y1 := new(uint256.Uint).Sub(y.abs, one)
	z.abs = z.abs.Add(z.abs.AndNot(y1, x.abs), one)
	z.neg = true // z cannot be zero if one of x or y is negative

	return z
}

// And sets z = x & y and returns z.
func (z *Int) And(x, y *Int) *Int {
	if x.neg == y.neg {
		if x.neg {
			// (-x) & (-y) == ^(x-1) & ^(y-1) == ^((x-1) | (y-1)) == -(((x-1) | (y-1)) + 1)
			x1 := new(uint256.Uint).Sub(x.abs, one)
			y1 := new(uint256.Uint).Sub(y.abs, one)
			z.abs = z.abs.Add(z.abs.Or(x1, y1), one)
			z.neg = true // z cannot be zero if x and y are negative
			return z
		}

		// x & y == x & y
		z.abs = z.abs.And(x.abs, y.abs)
		z.neg = false
		return z
	}

	// x.neg != y.neg
	// REF: https://cs.opensource.google/go/go/+/refs/tags/go1.22.1:src/math/big/int.go;l=1192-1202;drc=d57303e65f00b84b528ee682747dbe1fd3316d30
	if x.neg {
		x, y = y, x // & is symmetric
	}

	// x & (-y) == x & ^(y-1) == x &^ (y-1)
	y1 := new(uint256.Uint).Sub(y.abs, uint256.One())
	z.abs = z.abs.AndNot(x.abs, y1)
	z.neg = false
	return z
}

// Rsh sets z = x >> n and returns z.
// OBS: Different from original implementation it was using math.Big
func (z *Int) Rsh(x *Int, n uint) *Int {
	if !x.neg {
		z.abs.Rsh(x.abs, n)
		z.neg = x.neg
		return z
	}

	// REF: https://cs.opensource.google/go/go/+/refs/tags/go1.22.1:src/math/big/int.go;l=1118-1126;drc=d57303e65f00b84b528ee682747dbe1fd3316d30
	t := NewInt(0).Sub(FromUint256(x.abs), NewInt(1))
	t = t.Rsh(t, n)

	_tmp := t.Add(t, NewInt(1))
	z.abs = _tmp.Abs()
	z.neg = true

	return z
}

// Lsh sets z = x << n and returns z.
func (z *Int) Lsh(x *Int, n uint) *Int {
	z.abs.Lsh(x.abs, n)
	z.neg = x.neg
	return z
}
