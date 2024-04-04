// arithmetic provides arithmetic operations for Uint objects.
// This includes basic binary operations such as addition, subtraction, multiplication, division, and modulo operations
// as well as overflow checks, and negation. These functions are essential for numeric
// calculations using 256-bit unsigned integers.
package uint256

import (
	"math/bits"
)

// Add sets z to the sum x+y
func (z *Uint) Add(x, y *Uint) *Uint {
	var carry uint64
	z.arr[0], carry = bits.Add64(x.arr[0], y.arr[0], 0)
	z.arr[1], carry = bits.Add64(x.arr[1], y.arr[1], carry)
	z.arr[2], carry = bits.Add64(x.arr[2], y.arr[2], carry)
	z.arr[3], _ = bits.Add64(x.arr[3], y.arr[3], carry)
	return z
}

// AddOverflow sets z to the sum x+y, and returns z and whether overflow occurred
func (z *Uint) AddOverflow(x, y *Uint) (*Uint, bool) {
	var carry uint64
	z.arr[0], carry = bits.Add64(x.arr[0], y.arr[0], 0)
	z.arr[1], carry = bits.Add64(x.arr[1], y.arr[1], carry)
	z.arr[2], carry = bits.Add64(x.arr[2], y.arr[2], carry)
	z.arr[3], carry = bits.Add64(x.arr[3], y.arr[3], carry)
	return z, carry != 0
}

// Sub sets z to the difference x-y
func (z *Uint) Sub(x, y *Uint) *Uint {
	var carry uint64
	z.arr[0], carry = bits.Sub64(x.arr[0], y.arr[0], 0)
	z.arr[1], carry = bits.Sub64(x.arr[1], y.arr[1], carry)
	z.arr[2], carry = bits.Sub64(x.arr[2], y.arr[2], carry)
	z.arr[3], _ = bits.Sub64(x.arr[3], y.arr[3], carry)
	return z
}

// SubOverflow sets z to the difference x-y and returns z and true if the operation underflowed
func (z *Uint) SubOverflow(x, y *Uint) (*Uint, bool) {
	var carry uint64
	z.arr[0], carry = bits.Sub64(x.arr[0], y.arr[0], 0)
	z.arr[1], carry = bits.Sub64(x.arr[1], y.arr[1], carry)
	z.arr[2], carry = bits.Sub64(x.arr[2], y.arr[2], carry)
	z.arr[3], carry = bits.Sub64(x.arr[3], y.arr[3], carry)
	return z, carry != 0
}

// Neg returns -x mod 2^256.
func (z *Uint) Neg(x *Uint) *Uint {
	return z.Sub(new(Uint), x)
}

// commented out for possible overflow
// Mul sets z to the product x*y
func (z *Uint) Mul(x, y *Uint) *Uint {
	var (
		res              Uint
		carry            uint64
		res1, res2, res3 uint64
	)

	carry, res.arr[0] = bits.Mul64(x.arr[0], y.arr[0])
	carry, res1 = umulHop(carry, x.arr[1], y.arr[0])
	carry, res2 = umulHop(carry, x.arr[2], y.arr[0])
	res3 = x.arr[3]*y.arr[0] + carry

	carry, res.arr[1] = umulHop(res1, x.arr[0], y.arr[1])
	carry, res2 = umulStep(res2, x.arr[1], y.arr[1], carry)
	res3 = res3 + x.arr[2]*y.arr[1] + carry

	carry, res.arr[2] = umulHop(res2, x.arr[0], y.arr[2])
	res3 = res3 + x.arr[1]*y.arr[2] + carry

	res.arr[3] = res3 + x.arr[0]*y.arr[3]

	return z.Set(&res)
}

// MulOverflow sets z to the product x*y, and returns z and  whether overflow occurred
func (z *Uint) MulOverflow(x, y *Uint) (*Uint, bool) {
	p := umul(x, y)
	copy(z.arr[:], p[:4])
	return z, (p[4] | p[5] | p[6] | p[7]) != 0
}

// commented out for possible overflow
// Div sets z to the quotient x/y for returns z.
// If y == 0, z is set to 0
func (z *Uint) Div(x, y *Uint) *Uint {
	if y.IsZero() || y.Gt(x) {
		return z.Clear()
	}
	if x.Eq(y) {
		return z.SetOne()
	}
	// Shortcut some cases
	if x.IsUint64() {
		return z.SetUint64(x.Uint64() / y.Uint64())
	}

	// At this point, we know
	// x/y ; x > y > 0

	var quot Uint
	udivrem(quot.arr[:], x.arr[:], y)
	return z.Set(&quot)
}

// MulMod calculates the modulo-m multiplication of x and y and
// returns z.
// If m == 0, z is set to 0 (OBS: differs from the big.Int)
func (z *Uint) MulMod(x, y, m *Uint) *Uint {
	if x.IsZero() || y.IsZero() || m.IsZero() {
		return z.Clear()
	}
	p := umul(x, y)

	if m.arr[3] != 0 {
		mu := Reciprocal(m)
		r := reduce4(p, m, mu)
		return z.Set(&r)
	}

	var (
		pl Uint
		ph Uint
	)

	pl = Uint{arr: [4]uint64{p[0], p[1], p[2], p[3]}}
	ph = Uint{arr: [4]uint64{p[4], p[5], p[6], p[7]}}

	// If the multiplication is within 256 bits use Mod().
	if ph.IsZero() {
		return z.Mod(&pl, m)
	}

	var quot [8]uint64
	rem := udivrem(quot[:], p[:], m)
	return z.Set(&rem)
}

// Mod sets z to the modulus x%y for y != 0 and returns z.
// If y == 0, z is set to 0 (OBS: differs from the big.Uint)
func (z *Uint) Mod(x, y *Uint) *Uint {
	if x.IsZero() || y.IsZero() {
		return z.Clear()
	}
	switch x.Cmp(y) {
	case -1:
		// x < y
		copy(z.arr[:], x.arr[:])
		return z
	case 0:
		// x == y
		return z.Clear() // They are equal
	}

	// At this point:
	// x != 0
	// y != 0
	// x > y

	// Shortcut trivial case
	if x.IsUint64() {
		return z.SetUint64(x.Uint64() % y.Uint64())
	}

	var quot Uint
	*z = udivrem(quot.arr[:], x.arr[:], y)
	return z
}

// DivMod sets z to the quotient x div y and m to the modulus x mod y and returns the pair (z, m) for y != 0.
// If y == 0, both z and m are set to 0 (OBS: differs from the big.Int)
func (z *Uint) DivMod(x, y, m *Uint) (*Uint, *Uint) {
	if y.IsZero() {
		return z.Clear(), m.Clear()
	}
	var quot Uint
	*m = udivrem(quot.arr[:], x.arr[:], y)
	*z = quot
	return z, m
}

// Exp sets z = base**exponent mod 2**256, and returns z.
func (z *Uint) Exp(base, exponent *Uint) *Uint {
	res := Uint{arr: [4]uint64{1, 0, 0, 0}}
	multiplier := *base
	expBitLen := exponent.BitLen()

	curBit := 0
	word := exponent.arr[0]
	for ; curBit < expBitLen && curBit < 64; curBit++ {
		if word&1 == 1 {
			res.Mul(&res, &multiplier)
		}
		multiplier.squared()
		word >>= 1
	}

	word = exponent.arr[1]
	for ; curBit < expBitLen && curBit < 128; curBit++ {
		if word&1 == 1 {
			res.Mul(&res, &multiplier)
		}
		multiplier.squared()
		word >>= 1
	}

	word = exponent.arr[2]
	for ; curBit < expBitLen && curBit < 192; curBit++ {
		if word&1 == 1 {
			res.Mul(&res, &multiplier)
		}
		multiplier.squared()
		word >>= 1
	}

	word = exponent.arr[3]
	for ; curBit < expBitLen && curBit < 256; curBit++ {
		if word&1 == 1 {
			res.Mul(&res, &multiplier)
		}
		multiplier.squared()
		word >>= 1
	}
	return z.Set(&res)
}

func (z *Uint) squared() {
	var (
		res                    Uint
		carry0, carry1, carry2 uint64
		res1, res2             uint64
	)

	carry0, res.arr[0] = bits.Mul64(z.arr[0], z.arr[0])
	carry0, res1 = umulHop(carry0, z.arr[0], z.arr[1])
	carry0, res2 = umulHop(carry0, z.arr[0], z.arr[2])

	carry1, res.arr[1] = umulHop(res1, z.arr[0], z.arr[1])
	carry1, res2 = umulStep(res2, z.arr[1], z.arr[1], carry1)

	carry2, res.arr[2] = umulHop(res2, z.arr[0], z.arr[2])

	res.arr[3] = 2*(z.arr[0]*z.arr[3]+z.arr[1]*z.arr[2]) + carry0 + carry1 + carry2

	z.Set(&res)
}

// udivrem divides u by d and produces both quotient and remainder.
// The quotient is stored in provided quot - len(u)-len(d)+1 words.
// It loosely follows the Knuth's division algorithm (sometimes referenced as "schoolbook" division) using 64-bit words.
// See Knuth, Volume 2, section 4.3.1, Algorithm D.
func udivrem(quot, u []uint64, d *Uint) (rem Uint) {
	var dLen int
	for i := len(d.arr) - 1; i >= 0; i-- {
		if d.arr[i] != 0 {
			dLen = i + 1
			break
		}
	}

	shift := uint(bits.LeadingZeros64(d.arr[dLen-1]))

	var dnStorage Uint
	dn := dnStorage.arr[:dLen]
	for i := dLen - 1; i > 0; i-- {
		dn[i] = (d.arr[i] << shift) | (d.arr[i-1] >> (64 - shift))
	}
	dn[0] = d.arr[0] << shift

	var uLen int
	for i := len(u) - 1; i >= 0; i-- {
		if u[i] != 0 {
			uLen = i + 1
			break
		}
	}

	if uLen < dLen {
		copy(rem.arr[:], u)
		return rem
	}

	var unStorage [9]uint64
	un := unStorage[:uLen+1]
	un[uLen] = u[uLen-1] >> (64 - shift)
	for i := uLen - 1; i > 0; i-- {
		un[i] = (u[i] << shift) | (u[i-1] >> (64 - shift))
	}
	un[0] = u[0] << shift

	// TODO: Skip the highest word of numerator if not significant.

	if dLen == 1 {
		r := udivremBy1(quot, un, dn[0])
		rem.SetUint64(r >> shift)
		return rem
	}

	udivremKnuth(quot, un, dn)

	for i := 0; i < dLen-1; i++ {
		rem.arr[i] = (un[i] >> shift) | (un[i+1] << (64 - shift))
	}
	rem.arr[dLen-1] = un[dLen-1] >> shift

	return rem
}

// umul computes full 256 x 256 -> 512 multiplication.
func umul(x, y *Uint) [8]uint64 {
	var (
		res                           [8]uint64
		carry, carry4, carry5, carry6 uint64
		res1, res2, res3, res4, res5  uint64
	)

	carry, res[0] = bits.Mul64(x.arr[0], y.arr[0])
	carry, res1 = umulHop(carry, x.arr[1], y.arr[0])
	carry, res2 = umulHop(carry, x.arr[2], y.arr[0])
	carry4, res3 = umulHop(carry, x.arr[3], y.arr[0])

	carry, res[1] = umulHop(res1, x.arr[0], y.arr[1])
	carry, res2 = umulStep(res2, x.arr[1], y.arr[1], carry)
	carry, res3 = umulStep(res3, x.arr[2], y.arr[1], carry)
	carry5, res4 = umulStep(carry4, x.arr[3], y.arr[1], carry)

	carry, res[2] = umulHop(res2, x.arr[0], y.arr[2])
	carry, res3 = umulStep(res3, x.arr[1], y.arr[2], carry)
	carry, res4 = umulStep(res4, x.arr[2], y.arr[2], carry)
	carry6, res5 = umulStep(carry5, x.arr[3], y.arr[2], carry)

	carry, res[3] = umulHop(res3, x.arr[0], y.arr[3])
	carry, res[4] = umulStep(res4, x.arr[1], y.arr[3], carry)
	carry, res[5] = umulStep(res5, x.arr[2], y.arr[3], carry)
	res[7], res[6] = umulStep(carry6, x.arr[3], y.arr[3], carry)

	return res
}

// umulStep computes (hi * 2^64 + lo) = z + (x * y) + carry.
func umulStep(z, x, y, carry uint64) (hi, lo uint64) {
	hi, lo = bits.Mul64(x, y)
	lo, carry = bits.Add64(lo, carry, 0)
	hi, _ = bits.Add64(hi, 0, carry)
	lo, carry = bits.Add64(lo, z, 0)
	hi, _ = bits.Add64(hi, 0, carry)
	return hi, lo
}

// umulHop computes (hi * 2^64 + lo) = z + (x * y)
func umulHop(z, x, y uint64) (hi, lo uint64) {
	hi, lo = bits.Mul64(x, y)
	lo, carry := bits.Add64(lo, z, 0)
	hi, _ = bits.Add64(hi, 0, carry)
	return hi, lo
}

// udivremBy1 divides u by single normalized word d and produces both quotient and remainder.
// The quotient is stored in provided quot.
func udivremBy1(quot, u []uint64, d uint64) (rem uint64) {
	reciprocal := reciprocal2by1(d)
	rem = u[len(u)-1] // Set the top word as remainder.
	for j := len(u) - 2; j >= 0; j-- {
		quot[j], rem = udivrem2by1(rem, u[j], d, reciprocal)
	}
	return rem
}

// udivremKnuth implements the division of u by normalized multiple word d from the Knuth's division algorithm.
// The quotient is stored in provided quot - len(u)-len(d) words.
// Updates u to contain the remainder - len(d) words.
func udivremKnuth(quot, u, d []uint64) {
	dh := d[len(d)-1]
	dl := d[len(d)-2]
	reciprocal := reciprocal2by1(dh)

	for j := len(u) - len(d) - 1; j >= 0; j-- {
		u2 := u[j+len(d)]
		u1 := u[j+len(d)-1]
		u0 := u[j+len(d)-2]

		var qhat, rhat uint64
		if u2 >= dh { // Division overflows.
			qhat = ^uint64(0)
			// TODO: Add "qhat one to big" adjustment (not needed for correctness, but helps avoiding "add back" case).
		} else {
			qhat, rhat = udivrem2by1(u2, u1, dh, reciprocal)
			ph, pl := bits.Mul64(qhat, dl)
			if ph > rhat || (ph == rhat && pl > u0) {
				qhat--
				// TODO: Add "qhat one to big" adjustment (not needed for correctness, but helps avoiding "add back" case).
			}
		}

		// Multiply and subtract.
		borrow := subMulTo(u[j:], d, qhat)
		u[j+len(d)] = u2 - borrow
		if u2 < borrow { // Too much subtracted, add back.
			qhat--
			u[j+len(d)] += addTo(u[j:], d)
		}

		quot[j] = qhat // Store quotient digit.
	}
}

// isBitSet returns true if bit n-th is set, where n = 0 is LSB.
// The n must be <= 255.
func (z *Uint) isBitSet(n uint) bool {
	return (z.arr[n/64] & (1 << (n % 64))) != 0
}

// addTo computes x += y.
// Requires len(x) >= len(y).
func addTo(x, y []uint64) uint64 {
	var carry uint64
	for i := 0; i < len(y); i++ {
		x[i], carry = bits.Add64(x[i], y[i], carry)
	}
	return carry
}

// subMulTo computes x -= y * multiplier.
// Requires len(x) >= len(y).
func subMulTo(x, y []uint64, multiplier uint64) uint64 {
	var borrow uint64
	for i := 0; i < len(y); i++ {
		s, carry1 := bits.Sub64(x[i], borrow, 0)
		ph, pl := bits.Mul64(y[i], multiplier)
		t, carry2 := bits.Sub64(s, pl, 0)
		x[i] = t
		borrow = ph + carry1 + carry2
	}
	return borrow
}

// reciprocal2by1 computes <^d, ^0> / d.
func reciprocal2by1(d uint64) uint64 {
	reciprocal, _ := bits.Div64(^d, ^uint64(0), d)
	return reciprocal
}

// udivrem2by1 divides <uh, ul> / d and produces both quotient and remainder.
// It uses the provided d's reciprocal.
// Implementation ported from https://github.com/chfast/intx and is based on
// "Improved division by invariant integers", Algorithm 4.
func udivrem2by1(uh, ul, d, reciprocal uint64) (quot, rem uint64) {
	qh, ql := bits.Mul64(reciprocal, uh)
	ql, carry := bits.Add64(ql, ul, 0)
	qh, _ = bits.Add64(qh, uh, carry)
	qh++

	r := ul - qh*d

	if r > ql {
		qh--
		r += d
	}

	if r >= d {
		qh++
		r -= d
	}

	return qh, r
}
