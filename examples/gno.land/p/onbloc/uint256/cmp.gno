// cmp (or, comparisons) includes methods for comparing Uint instances.
// These comparison functions cover a range of operations including equality checks, less than/greater than
// evaluations, and specialized comparisons such as signed greater than. These are fundamental for logical
// decision making based on Uint values.
package uint256

import (
	"math/bits"
)

// Cmp compares z and x and returns:
//
//	-1 if z <  x
//	 0 if z == x
//	+1 if z >  x
func (z *Uint) Cmp(x *Uint) (r int) {
	// z < x <=> z - x < 0 i.e. when subtraction overflows.
	d0, carry := bits.Sub64(z.arr[0], x.arr[0], 0)
	d1, carry := bits.Sub64(z.arr[1], x.arr[1], carry)
	d2, carry := bits.Sub64(z.arr[2], x.arr[2], carry)
	d3, carry := bits.Sub64(z.arr[3], x.arr[3], carry)
	if carry == 1 {
		return -1
	}
	if d0|d1|d2|d3 == 0 {
		return 0
	}
	return 1
}

// IsZero returns true if z == 0
func (z *Uint) IsZero() bool {
	return (z.arr[0] | z.arr[1] | z.arr[2] | z.arr[3]) == 0
}

// Sign returns:
//
//	-1 if z <  0
//	 0 if z == 0
//	+1 if z >  0
//
// Where z is interpreted as a two's complement signed number
func (z *Uint) Sign() int {
	if z.IsZero() {
		return 0
	}
	if z.arr[3] < 0x8000000000000000 {
		return 1
	}
	return -1
}

// LtUint64 returns true if z is smaller than n
func (z *Uint) LtUint64(n uint64) bool {
	return z.arr[0] < n && (z.arr[1]|z.arr[2]|z.arr[3]) == 0
}

// GtUint64 returns true if z is larger than n
func (z *Uint) GtUint64(n uint64) bool {
	return z.arr[0] > n || (z.arr[1]|z.arr[2]|z.arr[3]) != 0
}

// Lt returns true if z < x
func (z *Uint) Lt(x *Uint) bool {
	// z < x <=> z - x < 0 i.e. when subtraction overflows.
	_, carry := bits.Sub64(z.arr[0], x.arr[0], 0)
	_, carry = bits.Sub64(z.arr[1], x.arr[1], carry)
	_, carry = bits.Sub64(z.arr[2], x.arr[2], carry)
	_, carry = bits.Sub64(z.arr[3], x.arr[3], carry)

	return carry != 0
}

// Gt returns true if z > x
func (z *Uint) Gt(x *Uint) bool {
	return x.Lt(z)
}

// Lte returns true if z <= x
func (z *Uint) Lte(x *Uint) bool {
	cond1 := z.Lt(x)
	cond2 := z.Eq(x)

	if cond1 || cond2 {
		return true
	}
	return false
}

// Gte returns true if z >= x
func (z *Uint) Gte(x *Uint) bool {
	cond1 := z.Gt(x)
	cond2 := z.Eq(x)

	if cond1 || cond2 {
		return true
	}
	return false
}

// Eq returns true if z == x
func (z *Uint) Eq(x *Uint) bool {
	return (z.arr[0] == x.arr[0]) && (z.arr[1] == x.arr[1]) && (z.arr[2] == x.arr[2]) && (z.arr[3] == x.arr[3])
}

// Neq returns true if z != x
func (z *Uint) Neq(x *Uint) bool {
	return !z.Eq(x)
}

// Sgt interprets z and x as signed integers, and returns
// true if z > x
func (z *Uint) Sgt(x *Uint) bool {
	zSign := z.Sign()
	xSign := x.Sign()

	switch {
	case zSign >= 0 && xSign < 0:
		return true
	case zSign < 0 && xSign >= 0:
		return false
	default:
		return z.Gt(x)
	}
}

// Max selects the maximum value between elements, then stores the result in z.
// 3 if Max(1, 3, 2)
func (z *Uint) Max(x *Uint, ys...*Uint) *Uint {
	lenys := len(ys)
	
	var best *Uint = x
	if lenys == 0 {
		goto copyResult
	}
	for n := 0; n < lenys; n++ {
		if best == nil || ys[n] != nil && ys[n].Gt(best) {
			best = ys[n]
		}
	}
copyResult:
	if best != nil {
		copy(z.arr[:], best.arr[:])
	}
	return z
}

func (z *Uint) Min(x *Uint, ys...*Uint) *Uint {
	lenys := len(ys)
	
	var best *Uint = x
	if lenys == 0 {
		goto copyResult
	}
	for n := 0; n < lenys; n++ {
		if best == nil || ys[n] != nil && ys[n].Lt(best) {
			best = ys[n]
		}
	}
copyResult:
	if best != nil {
		copy(z.arr[:], best.arr[:])
	}
	return z
}
