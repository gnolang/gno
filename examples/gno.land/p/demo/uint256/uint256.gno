// Ported from https://github.com/holiman/uint256
package uint256

import (
	"errors"
	"math/bits"
	"strconv"
)

const (
	MaxUint64 = 1<<64 - 1
	uintSize  = 32 << (^uint(0) >> 63)
)

func Zero() *Uint {
	return NewUint(0)
}

func One() *Uint {
	return NewUint(1)
}

func (x *Uint) Min(y *Uint) *Uint {
	if x.Lt(y) {
		return x
	}
	return y
}

// Uint is represented as an array of 4 uint64, in little-endian order,
// so that Uint[3] is the most significant, and Uint[0] is the least significant
type Uint struct {
	arr [4]uint64
}

// NewUint returns a new initialized Uint.
func NewUint(val uint64) *Uint {
	z := &Uint{arr: [4]uint64{val, 0, 0, 0}}
	return z
}

// Uint64 returns the lower 64-bits of z
func (z *Uint) Uint64() uint64 {
	return z.arr[0]
}

// Uint64WithOverflow returns the lower 64-bits of z and bool whether overflow occurred
func (z *Uint) Uint64WithOverflow() (uint64, bool) {
	return z.arr[0], (z.arr[1] | z.arr[2] | z.arr[3]) != 0
}

func (z *Uint) Int32() int32 {
	x := z.arr[0]
	if x > 0x7fffffff {
		panic("U256 Int32 overflow")
	}
	return int32(x)
}

// SetUint64 sets z to the value x
func (z *Uint) SetUint64(x uint64) *Uint {
	z.arr[3], z.arr[2], z.arr[1], z.arr[0] = 0, 0, 0, x
	return z
}

// LtUint64 returns true if z is smaller than n
func (z *Uint) LtUint64(n uint64) bool {
	return z.arr[0] < n && (z.arr[1]|z.arr[2]|z.arr[3]) == 0
}

// GtUint64 returns true if z is larger than n
func (z *Uint) GtUint64(n uint64) bool {
	return z.arr[0] > n || (z.arr[1]|z.arr[2]|z.arr[3]) != 0
}

// IsUint64 reports whether z can be represented as a uint64.
func (z *Uint) IsUint64() bool {
	return (z.arr[1] | z.arr[2] | z.arr[3]) == 0
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


// Clear sets z to 0
func (z *Uint) Clear() *Uint {
	z.arr[3], z.arr[2], z.arr[1], z.arr[0] = 0, 0, 0, 0
	return z
}

// SetAllOne sets all the bits of z to 1
func (z *Uint) SetAllOne() *Uint {
	z.arr[3], z.arr[2], z.arr[1], z.arr[0] = MaxUint64, MaxUint64, MaxUint64, MaxUint64
	return z
}

// Neg returns -x mod 2**256.
func (z *Uint) Neg(x *Uint) *Uint {
	return z.Sub(new(Uint), x)
}

// Not sets z = ^x and returns z.
func (z *Uint) Not(x *Uint) *Uint {
	z.arr[3], z.arr[2], z.arr[1], z.arr[0] = ^x.arr[3], ^x.arr[2], ^x.arr[1], ^x.arr[0]
	return z
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

// Set sets z to x and returns z.
func (z *Uint) Set(x *Uint) *Uint {
	*z = *x

	return z
}

// SetOne sets z to 1
func (z *Uint) SetOne() *Uint {
	z.arr[3], z.arr[2], z.arr[1], z.arr[0] = 0, 0, 0, 1
	return z
}

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

// SubOverflow sets z to the difference x-y and returns z and true if the operation underflowed
func (z *Uint) SubOverflow(x, y *Uint) (*Uint, bool) {
	var carry uint64
	z.arr[0], carry = bits.Sub64(x.arr[0], y.arr[0], 0)
	z.arr[1], carry = bits.Sub64(x.arr[1], y.arr[1], carry)
	z.arr[2], carry = bits.Sub64(x.arr[2], y.arr[2], carry)
	z.arr[3], carry = bits.Sub64(x.arr[3], y.arr[3], carry)
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

// Lsh sets z = x << n and returns z.
func (z *Uint) Lsh(x *Uint, n uint) *Uint {
	// n % 64 == 0
	if n&0x3f == 0 {
		switch n {
		case 0:
			return z.Set(x)
		case 64:
			return z.lsh64(x)
		case 128:
			return z.lsh128(x)
		case 192:
			return z.lsh192(x)
		default:
			return z.Clear()
		}
	}
	var a, b uint64
	// Big swaps first
	switch {
	case n > 192:
		if n > 256 {
			return z.Clear()
		}
		z.lsh192(x)
		n -= 192
		goto sh192
	case n > 128:
		z.lsh128(x)
		n -= 128
		goto sh128
	case n > 64:
		z.lsh64(x)
		n -= 64
		goto sh64
	default:
		z.Set(x)
	}

	// remaining shifts
	a = z.arr[0] >> (64 - n)
	z.arr[0] = z.arr[0] << n

sh64:
	b = z.arr[1] >> (64 - n)
	z.arr[1] = (z.arr[1] << n) | a

sh128:
	a = z.arr[2] >> (64 - n)
	z.arr[2] = (z.arr[2] << n) | b

sh192:
	z.arr[3] = (z.arr[3] << n) | a

	return z
}

// Rsh sets z = x >> n and returns z.
func (z *Uint) Rsh(x *Uint, n uint) *Uint {
	// n % 64 == 0
	if n&0x3f == 0 {
		switch n {
		case 0:
			return z.Set(x)
		case 64:
			return z.rsh64(x)
		case 128:
			return z.rsh128(x)
		case 192:
			return z.rsh192(x)
		default:
			return z.Clear()
		}
	}
	var a, b uint64
	// Big swaps first
	switch {
	case n > 192:
		if n > 256 {
			return z.Clear()
		}
		z.rsh192(x)
		n -= 192
		goto sh192
	case n > 128:
		z.rsh128(x)
		n -= 128
		goto sh128
	case n > 64:
		z.rsh64(x)
		n -= 64
		goto sh64
	default:
		z.Set(x)
	}

	// remaining shifts
	a = z.arr[3] << (64 - n)
	z.arr[3] = z.arr[3] >> n

sh64:
	b = z.arr[2] << (64 - n)
	z.arr[2] = (z.arr[2] >> n) | a

sh128:
	a = z.arr[1] << (64 - n)
	z.arr[1] = (z.arr[1] >> n) | b

sh192:
	z.arr[0] = (z.arr[0] >> n) | a

	return z
}

// SRsh (Signed/Arithmetic right shift)
// considers z to be a signed integer, during right-shift
// and sets z = x >> n and returns z.
func (z *Uint) SRsh(x *Uint, n uint) *Uint {
	// If the MSB is 0, SRsh is same as Rsh.
	if !x.isBitSet(255) {
		return z.Rsh(x, n)
	}
	if n%64 == 0 {
		switch n {
		case 0:
			return z.Set(x)
		case 64:
			return z.srsh64(x)
		case 128:
			return z.srsh128(x)
		case 192:
			return z.srsh192(x)
		default:
			return z.SetAllOne()
		}
	}
	var a uint64 = MaxUint64 << (64 - n%64)
	// Big swaps first
	switch {
	case n > 192:
		if n > 256 {
			return z.SetAllOne()
		}
		z.srsh192(x)
		n -= 192
		goto sh192
	case n > 128:
		z.srsh128(x)
		n -= 128
		goto sh128
	case n > 64:
		z.srsh64(x)
		n -= 64
		goto sh64
	default:
		z.Set(x)
	}

	// remaining shifts
	z.arr[3], a = (z.arr[3]>>n)|a, z.arr[3]<<(64-n)

sh64:
	z.arr[2], a = (z.arr[2]>>n)|a, z.arr[2]<<(64-n)

sh128:
	z.arr[1], a = (z.arr[1]>>n)|a, z.arr[1]<<(64-n)

sh192:
	z.arr[0] = (z.arr[0] >> n) | a

	return z
}

func (z *Uint) lsh64(x *Uint) *Uint {
	z.arr[3], z.arr[2], z.arr[1], z.arr[0] = x.arr[2], x.arr[1], x.arr[0], 0
	return z
}

func (z *Uint) lsh128(x *Uint) *Uint {
	z.arr[3], z.arr[2], z.arr[1], z.arr[0] = x.arr[1], x.arr[0], 0, 0
	return z
}

func (z *Uint) lsh192(x *Uint) *Uint {
	z.arr[3], z.arr[2], z.arr[1], z.arr[0] = x.arr[0], 0, 0, 0
	return z
}

func (z *Uint) rsh64(x *Uint) *Uint {
	z.arr[3], z.arr[2], z.arr[1], z.arr[0] = 0, x.arr[3], x.arr[2], x.arr[1]
	return z
}

func (z *Uint) rsh128(x *Uint) *Uint {
	z.arr[3], z.arr[2], z.arr[1], z.arr[0] = 0, 0, x.arr[3], x.arr[2]
	return z
}

func (z *Uint) rsh192(x *Uint) *Uint {
	z.arr[3], z.arr[2], z.arr[1], z.arr[0] = 0, 0, 0, x.arr[3]
	return z
}

func (z *Uint) srsh64(x *Uint) *Uint {
	z.arr[3], z.arr[2], z.arr[1], z.arr[0] = MaxUint64, x.arr[3], x.arr[2], x.arr[1]
	return z
}

func (z *Uint) srsh128(x *Uint) *Uint {
	z.arr[3], z.arr[2], z.arr[1], z.arr[0] = MaxUint64, MaxUint64, x.arr[3], x.arr[2]
	return z
}

func (z *Uint) srsh192(x *Uint) *Uint {
	z.arr[3], z.arr[2], z.arr[1], z.arr[0] = MaxUint64, MaxUint64, MaxUint64, x.arr[3]
	return z
}

// Or sets z = x | y and returns z.
func (z *Uint) Or(x, y *Uint) *Uint {
	z.arr[0] = x.arr[0] | y.arr[0]
	z.arr[1] = x.arr[1] | y.arr[1]
	z.arr[2] = x.arr[2] | y.arr[2]
	z.arr[3] = x.arr[3] | y.arr[3]
	return z
}

// And sets z = x & y and returns z.
func (z *Uint) And(x, y *Uint) *Uint {
	z.arr[0] = x.arr[0] & y.arr[0]
	z.arr[1] = x.arr[1] & y.arr[1]
	z.arr[2] = x.arr[2] & y.arr[2]
	z.arr[3] = x.arr[3] & y.arr[3]
	return z
}

// Xor sets z = x ^ y and returns z.
func (z *Uint) Xor(x, y *Uint) *Uint {
	z.arr[0] = x.arr[0] ^ y.arr[0]
	z.arr[1] = x.arr[1] ^ y.arr[1]
	z.arr[2] = x.arr[2] ^ y.arr[2]
	z.arr[3] = x.arr[3] ^ y.arr[3]
	return z
}

// Byte sets z to the value of the byte at position n,
// with 'z' considered as a big-endian 32-byte integer
// if 'n' > 32, f is set to 0
// Example: f = '5', n=31 => 5
func (z *Uint) Byte(n *Uint) *Uint {
	// in z, z.arr[0] is the least significant
	if number, overflow := n.Uint64WithOverflow(); !overflow {
		if number < 32 {
			number := z.arr[4-1-number/8]
			offset := (n.arr[0] & 0x7) << 3 // 8*(n.d % 8)
			z.arr[0] = (number & (0xff00000000000000 >> offset)) >> (56 - offset)
			z.arr[3], z.arr[2], z.arr[1] = 0, 0, 0
			return z
		}
	}

	return z.Clear()
}

// MarshalJSON implements json.Marshaler.
// MarshalJSON marshals using the 'decimal string' representation. This is _not_ compatible
// with big.Uint: big.Uint marshals into JSON 'native' numeric format.
//
// The JSON  native format is, on some platforms, (e.g. javascript), limited to 53-bit large
// integer space. Thus, U256 uses string-format, which is not compatible with
// big.int (big.Uint refuses to unmarshal a string representation).
func (z *Uint) MarshalJSON() ([]byte, error) {
	return []byte(`"` + z.Dec() + `"`), nil
}

// UnmarshalJSON implements json.Unmarshaler. UnmarshalJSON accepts either
// - Quoted string: either hexadecimal OR decimal
// - Not quoted string: only decimal
func (z *Uint) UnmarshalJSON(input []byte) error {
	if len(input) < 2 || input[0] != '"' || input[len(input)-1] != '"' {
		// if not quoted, it must be decimal
		return z.fromDecimal(string(input))
	}
	return z.UnmarshalText(input[1 : len(input)-1])
}

// MarshalText implements encoding.TextMarshaler
// MarshalText marshals using the decimal representation (compatible with big.Uint)
func (z *Uint) MarshalText() ([]byte, error) {
	return []byte(z.Dec()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler. This method
// can unmarshal either hexadecimal or decimal.
// - For hexadecimal, the input _must_ be prefixed with 0x or 0X
func (z *Uint) UnmarshalText(input []byte) error {
	if len(input) >= 2 && input[0] == '0' && (input[1] == 'x' || input[1] == 'X') {
		return z.fromHex(string(input))
	}
	return z.fromDecimal(string(input))
}

const (
	hextable  = "0123456789abcdef"
	bintable  = "\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\x00\x01\x02\x03\x04\x05\x06\a\b\t\xff\xff\xff\xff\xff\xff\xff\n\v\f\r\x0e\x0f\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\n\v\f\r\x0e\x0f\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff\xff"
	badNibble = 0xff
)

// fromHex is the internal implementation of parsing a hex-string.
func (z *Uint) fromHex(hex string) error {
	if err := checkNumberS(hex); err != nil {
		return err
	}
	if len(hex) > 66 {
		return ErrBig256Range
	}
	z.Clear()
	end := len(hex)
	for i := 0; i < 4; i++ {
		start := end - 16
		if start < 2 {
			start = 2
		}
		for ri := start; ri < end; ri++ {
			nib := bintable[hex[ri]]
			if nib == badNibble {
				return ErrSyntax
			}
			z.arr[i] = z.arr[i] << 4
			z.arr[i] += uint64(nib)
		}
		end = start
	}
	return nil
}

// FromDecimal is a convenience-constructor to create an Uint from a
// decimal (base 10) string. Numbers larger than 256 bits are not accepted.
func FromDecimal(decimal string) (*Uint, error) {
	var z Uint
	if err := z.SetFromDecimal(decimal); err != nil {
		return nil, err
	}
	return &z, nil
}

// MustFromDecimal is a convenience-constructor to create an Int from a
// decimal (base 10) string.
// Returns a new Int and panics if any error occurred.
func MustFromDecimal(decimal string) *Uint {
	var z Uint
	if err := z.SetFromDecimal(decimal); err != nil {
		panic(err)
	}
	return &z
}

const twoPow256Sub1 = "115792089237316195423570985008687907853269984665640564039457584007913129639935"

// SetFromDecimal sets z from the given string, interpreted as a decimal number.
// OBS! This method is _not_ strictly identical to the (*big.Uint).SetString(..., 10) method.
// Notable differences:
// - This method does not accept underscore input, e.g. "100_000",
// - This method does not accept negative zero as valid, e.g "-0",
//   - (this method does not accept any negative input as valid))
func (z *Uint) SetFromDecimal(s string) (err error) {
	// Remove max one leading +
	if len(s) > 0 && s[0] == '+' {
		s = s[1:]
	}
	// Remove any number of leading zeroes
	if len(s) > 0 && s[0] == '0' {
		var i int
		var c rune
		for i, c = range s {
			if c != '0' {
				break
			}
		}
		s = s[i:]
	}
	if len(s) < len(twoPow256Sub1) {
		return z.fromDecimal(s)
	}
	if len(s) == len(twoPow256Sub1) {
		if s > twoPow256Sub1 {
			return ErrBig256Range
		}
		return z.fromDecimal(s)
	}
	return ErrBig256Range
}

var (
	ErrEmptyString      = errors.New("empty hex string")
	ErrSyntax           = errors.New("invalid hex string")
	ErrMissingPrefix    = errors.New("hex string without 0x prefix")
	ErrEmptyNumber      = errors.New("hex string \"0x\"")
	ErrLeadingZero      = errors.New("hex number with leading zero digits")
	ErrBig256Range      = errors.New("hex number > 256 bits")
	ErrBadBufferLength  = errors.New("bad ssz buffer length")
	ErrBadEncodedLength = errors.New("bad ssz encoded length")
)

func checkNumberS(input string) error {
	l := len(input)
	if l == 0 {
		return ErrEmptyString
	}
	if l < 2 || input[0] != '0' ||
		(input[1] != 'x' && input[1] != 'X') {
		return ErrMissingPrefix
	}
	if l == 2 {
		return ErrEmptyNumber
	}
	if len(input) > 3 && input[2] == '0' {
		return ErrLeadingZero
	}
	return nil
}

// multipliers holds the values that are needed for fromDecimal
var multipliers = [5]*Uint{
	nil, // represents first round, no multiplication needed
	{[4]uint64{10000000000000000000, 0, 0, 0}},                                     // 10 ^ 19
	{[4]uint64{687399551400673280, 5421010862427522170, 0, 0}},                     // 10 ^ 38
	{[4]uint64{5332261958806667264, 17004971331911604867, 2938735877055718769, 0}}, // 10 ^ 57
	{[4]uint64{0, 8607968719199866880, 532749306367912313, 1593091911132452277}},   // 10 ^ 76
}

// fromDecimal is a helper function to only ever be called via SetFromDecimal
// this function takes a string and chunks it up, calling ParseUint on it up to 5 times
// these chunks are then multiplied by the proper power of 10, then added together.
func (z *Uint) fromDecimal(bs string) error {
	// first clear the input
	z.Clear()
	// the maximum value of uint64 is 18446744073709551615, which is 20 characters
	// one less means that a string of 19 9's is always within the uint64 limit
	var (
		num       uint64
		err       error
		remaining = len(bs)
	)
	if remaining == 0 {
		return errors.New("EOF")
	}
	// We proceed in steps of 19 characters (nibbles), from least significant to most significant.
	// This means that the first (up to) 19 characters do not need to be multiplied.
	// In the second iteration, our slice of 19 characters needs to be multipleied
	// by a factor of 10^19. Et cetera.
	for i, mult := range multipliers {
		if remaining <= 0 {
			return nil // Done
		} else if remaining > 19 {
			num, err = parseUint(bs[remaining-19:remaining], 10, 64)
		} else {
			// Final round
			num, err = parseUint(bs, 10, 64)
		}
		if err != nil {
			return err
		}
		// add that number to our running total
		if i == 0 {
			z.SetUint64(num)
		} else {
			base := NewUint(num)
			z.Add(z, base.Mul(base, mult))
		}
		// Chop off another 19 characters
		if remaining > 19 {
			bs = bs[0 : remaining-19]
		}
		remaining -= 19
	}
	return nil
}

// lower(c) is a lower-case letter if and only if
// c is either that lower-case letter or the equivalent upper-case letter.
// Instead of writing c == 'x' || c == 'X' one can write lower(c) == 'x'.
// Note that lower of non-letters can produce other non-letters.
func lower(c byte) byte {
	return c | ('x' - 'X')
}

// ParseUint is like ParseUint but for unsigned numbers.
//
// A sign prefix is not permitted.
func parseUint(s string, base int, bitSize int) (uint64, error) {
	const fnParseUint = "ParseUint"

	if s == "" {
		return 0, errors.New("syntax error: ParseUint empty string")
	}

	base0 := base == 0

	s0 := s
	switch {
	case 2 <= base && base <= 36:
		// valid base; nothing to do

	case base == 0:
		// Look for octal, hex prefix.
		base = 10
		if s[0] == '0' {
			switch {
			case len(s) >= 3 && lower(s[1]) == 'b':
				base = 2
				s = s[2:]
			case len(s) >= 3 && lower(s[1]) == 'o':
				base = 8
				s = s[2:]
			case len(s) >= 3 && lower(s[1]) == 'x':
				base = 16
				s = s[2:]
			default:
				base = 8
				s = s[1:]
			}
		}

	default:
		return 0, errors.New("invalid base")
	}

	if bitSize == 0 {
		bitSize = uintSize
	} else if bitSize < 0 || bitSize > 64 {
		return 0, errors.New("invalid bit size")
	}

	// Cutoff is the smallest number such that cutoff*base > maxUint64.
	// Use compile-time constants for common cases.
	var cutoff uint64
	switch base {
	case 10:
		cutoff = MaxUint64/10 + 1
	case 16:
		cutoff = MaxUint64/16 + 1
	default:
		cutoff = MaxUint64/uint64(base) + 1
	}

	maxVal := uint64(1)<<uint(bitSize) - 1

	underscores := false
	var n uint64
	for _, c := range []byte(s) {
		var d byte
		switch {
		case c == '_' && base0:
			underscores = true
			continue
		case '0' <= c && c <= '9':
			d = c - '0'
		case 'a' <= lower(c) && lower(c) <= 'z':
			d = lower(c) - 'a' + 10
		default:
			return 0, errors.New("syntax error")
		}

		if d >= byte(base) {
			return 0, errors.New("syntax error")
		}

		if n >= cutoff {
			// n*base overflows
			return maxVal, errors.New("range error")
		}
		n *= uint64(base)

		n1 := n + uint64(d)
		if n1 < n || n1 > maxVal {
			// n+d overflows
			return maxVal, errors.New("range error")
		}
		n = n1
	}

	if underscores && !underscoreOK(s0) {
		return 0, errors.New("syntax error")
	}

	return n, nil
}

// underscoreOK reports whether the underscores in s are allowed.
// Checking them in this one function lets all the parsers skip over them simply.
// Underscore must appear only between digits or between a base prefix and a digit.
func underscoreOK(s string) bool {
	// saw tracks the last character (class) we saw:
	// ^ for beginning of number,
	// 0 for a digit or base prefix,
	// _ for an underscore,
	// ! for none of the above.
	saw := '^'
	i := 0

	// Optional sign.
	if len(s) >= 1 && (s[0] == '-' || s[0] == '+') {
		s = s[1:]
	}

	// Optional base prefix.
	hex := false
	if len(s) >= 2 && s[0] == '0' && (lower(s[1]) == 'b' || lower(s[1]) == 'o' || lower(s[1]) == 'x') {
		i = 2
		saw = '0' // base prefix counts as a digit for "underscore as digit separator"
		hex = lower(s[1]) == 'x'
	}

	// Number proper.
	for ; i < len(s); i++ {
		// Digits are always okay.
		if '0' <= s[i] && s[i] <= '9' || hex && 'a' <= lower(s[i]) && lower(s[i]) <= 'f' {
			saw = '0'
			continue
		}
		// Underscore must follow digit.
		if s[i] == '_' {
			if saw != '0' {
				return false
			}
			saw = '_'
			continue
		}
		// Underscore must also be followed by digit.
		if saw == '_' {
			return false
		}
		// Saw non-digit, non-underscore.
		saw = '!'
	}
	return saw != '_'
}

// Dec returns the decimal representation of z.
func (z *Uint) Dec() string {
	if z.IsZero() {
		return "0"
	}
	if z.IsUint64() {
		return strconv.FormatUint(z.Uint64(), 10)
	}

	// The max uint64 value being 18446744073709551615, the largest
	// power-of-ten below that is 10000000000000000000.
	// When we do a DivMod using that number, the remainder that we
	// get back is the lower part of the output.
	//
	// The ascii-output of remainder will never exceed 19 bytes (since it will be
	// below 10000000000000000000).
	//
	// Algorithm example using 100 as divisor
	//
	// 12345 % 100 = 45   (rem)
	// 12345 / 100 = 123  (quo)
	// -> output '45', continue iterate on 123
	var (
		// out is 98 bytes long: 78 (max size of a string without leading zeroes,
		// plus slack so we can copy 19 bytes every iteration).
		// We init it with zeroes, because when strconv appends the ascii representations,
		// it will omit leading zeroes.
		out     = []byte("00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
		divisor = NewUint(10000000000000000000) // 20 digits
		y       = new(Uint).Set(z)              // copy to avoid modifying z
		pos     = len(out)                      // position to write to
		buf     = make([]byte, 0, 19)           // buffer to write uint64:s to
	)
	for {
		// Obtain Q and R for divisor
		var quot Uint
		rem := udivrem(quot.arr[:], y.arr[:], divisor)
		y.Set(&quot) // Set Q for next loop
		// Convert the R to ascii representation
		buf = strconv.AppendUint(buf[:0], rem.Uint64(), 10)
		// Copy in the ascii digits
		copy(out[pos-len(buf):], buf)
		if y.IsZero() {
			break
		}
		// Move 19 digits left
		pos -= 19
	}
	// skip leading zeroes by only using the 'used size' of buf
	return string(out[pos-len(buf):])
}

func (z *Uint) ToString() string {
	if z == nil {
		panic("U256 ToString() nil")
	}

	return z.Dec()
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

// Clone creates a new Int identical to z
func (z *Uint) Clone() *Uint {
	var x Uint
	x.arr[0] = z.arr[0]
	x.arr[1] = z.arr[1]
	x.arr[2] = z.arr[2]
	x.arr[3] = z.arr[3]

	return &x
}

func (z *Uint) IsNil() bool {
	return z == nil
}

func (z *Uint) NilToZero() *Uint {
	if z == nil {
		z = NewUint(0)
	}

	return z
}

func (z *Uint) AndNot(x, y *Uint) *Uint {
	z.arr[0] = x.arr[0] &^ y.arr[0]
	z.arr[1] = x.arr[1] &^ y.arr[1]
	z.arr[2] = x.arr[2] &^ y.arr[2]
	z.arr[3] = x.arr[3] &^ y.arr[3]
	return z
}


// SetBytes interprets buf as the bytes of a big-endian unsigned
// integer, sets z to that value, and returns z.
// If buf is larger than 32 bytes, the last 32 bytes is used.
func (z *Uint) SetBytes(buf []byte) *Uint {
	switch l := len(buf); l {
	case 0:
		z.Clear()
	case 1:
		z.SetBytes1(buf)
	case 2:
		z.SetBytes2(buf)
	case 3:
		z.SetBytes3(buf)
	case 4:
		z.SetBytes4(buf)
	case 5:
		z.SetBytes5(buf)
	case 6:
		z.SetBytes6(buf)
	case 7:
		z.SetBytes7(buf)
	case 8:
		z.SetBytes8(buf)
	case 9:
		z.SetBytes9(buf)
	case 10:
		z.SetBytes10(buf)
	case 11:
		z.SetBytes11(buf)
	case 12:
		z.SetBytes12(buf)
	case 13:
		z.SetBytes13(buf)
	case 14:
		z.SetBytes14(buf)
	case 15:
		z.SetBytes15(buf)
	case 16:
		z.SetBytes16(buf)
	case 17:
		z.SetBytes17(buf)
	case 18:
		z.SetBytes18(buf)
	case 19:
		z.SetBytes19(buf)
	case 20:
		z.SetBytes20(buf)
	case 21:
		z.SetBytes21(buf)
	case 22:
		z.SetBytes22(buf)
	case 23:
		z.SetBytes23(buf)
	case 24:
		z.SetBytes24(buf)
	case 25:
		z.SetBytes25(buf)
	case 26:
		z.SetBytes26(buf)
	case 27:
		z.SetBytes27(buf)
	case 28:
		z.SetBytes28(buf)
	case 29:
		z.SetBytes29(buf)
	case 30:
		z.SetBytes30(buf)
	case 31:
		z.SetBytes31(buf)
	default:
		z.SetBytes32(buf[l-32:])
	}
	return z
}
