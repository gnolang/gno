// Package big implements arbitrary-precision arithmetic (big numbers).
//
// This package is a subset of Go's math/big. Only the Int type is
// currently provided. Float and Rat are planned follow-ups.
//
// The semantics, naming, and method signatures mirror Go's math/big so
// existing Go code that uses *big.Int can be ported with minimal change:
//
//	z := new(big.Int).SetInt64(1)
//	z.Lsh(z, 256)        // not yet implemented; see godoc for the supported subset
//	z.Add(z, big.NewInt(1))
//	println(z.String())
//
// All arithmetic methods take pointer arguments and return the receiver,
// allowing chaining. As in Go's math/big, the receiver may alias any
// argument: z.Add(z, y) is well-defined.
package big

import (
	"encoding/binary"
	"math/bits"
	"strconv"
)

// divCheck panics on the Gno side for y == 0 so the division-by-zero
// error is a normal recoverable Gno panic instead of a native Go panic
// surfaced through the genstd bridge (which is not catchable by a
// realm-level recover()). The message matches Go's math/big exactly.
func divCheck(y *Int) {
	if len(y.abs) == 0 {
		panic("division by zero")
	}
}

// An Int represents a signed multi-precision integer. The zero value
// for Int represents the value 0.
//
// Operations always take pointer arguments (*Int) rather than Int values,
// and each unique Int value requires its own unique *Int pointer. To
// "copy" an Int value, an existing (or newly allocated) Int must be set
// to the new value using the Int.Set method; shallow copies of Ints are
// not supported and may lead to errors.
type Int struct {
	// neg is true when the integer is strictly negative.
	// abs is the big-endian unsigned magnitude, with no leading zero
	// bytes. The empty/nil slice represents zero, in which case neg
	// must be false (there is a single representation for zero).
	neg bool
	abs []byte
}

// NewInt allocates and returns a new Int set to x.
func NewInt(x int64) *Int {
	return new(Int).SetInt64(x)
}

// Set sets z to x and returns z.
func (z *Int) Set(x *Int) *Int {
	if z != x {
		z.abs = append(z.abs[:0], x.abs...)
		// Re-normalize defensively: every other setter enforces the
		// invariant len(abs) == 0 ⇒ neg == false. Set honors it too even
		// if x was constructed by an internal path that bypassed it.
		z.neg = x.neg && len(z.abs) > 0
	}
	return z
}

// SetInt64 sets z to x and returns z.
func (z *Int) SetInt64(x int64) *Int {
	neg := false
	var u uint64
	if x < 0 {
		neg = true
		// uint64(-x) is correct even for x == math.MinInt64
		// because -MinInt64 wraps to MinInt64 and the uint64
		// conversion reinterprets the bit pattern as 1<<63.
		u = uint64(-x)
	} else {
		u = uint64(x)
	}
	z.abs = appendUint64(z.abs[:0], u)
	z.neg = neg && len(z.abs) > 0
	return z
}

// SetUint64 sets z to x and returns z.
func (z *Int) SetUint64(x uint64) *Int {
	z.abs = appendUint64(z.abs[:0], x)
	z.neg = false
	return z
}

func appendUint64(buf []byte, x uint64) []byte {
	if x == 0 {
		return buf
	}
	var tmp [8]byte
	binary.BigEndian.PutUint64(tmp[:], x)
	i := 0
	for i < 7 && tmp[i] == 0 {
		i++
	}
	return append(buf, tmp[i:]...)
}

// IsUint64 reports whether x can be represented as a uint64.
func (x *Int) IsUint64() bool {
	return !x.neg && len(x.abs) <= 8
}

// IsInt64 reports whether x can be represented as an int64.
func (x *Int) IsInt64() bool {
	if len(x.abs) < 8 {
		return true
	}
	if len(x.abs) > 8 {
		return false
	}
	hi := x.abs[0]
	if x.neg {
		// MinInt64 is exactly 0x8000000000000000 with sign bit.
		if hi < 0x80 {
			return true
		}
		if hi > 0x80 {
			return false
		}
		for i := 1; i < 8; i++ {
			if x.abs[i] != 0 {
				return false
			}
		}
		return true
	}
	return hi < 0x80
}

// Uint64 returns the uint64 representation of x.
// If x cannot be represented in a uint64, the result is undefined.
func (x *Int) Uint64() uint64 {
	// Low 64 bits of the magnitude, like Go's low64: wider values
	// truncate instead of scanning the whole slice.
	b := x.abs
	if len(b) > 8 {
		b = b[len(b)-8:]
	}
	var v uint64
	for _, c := range b {
		v = v<<8 | uint64(c)
	}
	return v
}

// Int64 returns the int64 representation of x.
// If x cannot be represented in an int64, the result is undefined.
func (x *Int) Int64() int64 {
	v := int64(x.Uint64())
	if x.neg {
		// Inverse of the SetInt64 trick: for x == MinInt64, -v wraps
		// back to MinInt64 in two's complement, which is what we want.
		v = -v
	}
	return v
}

// Sign returns:
//
//	-1 if x <  0
//	 0 if x == 0
//	+1 if x >  0
func (x *Int) Sign() int {
	if len(x.abs) == 0 {
		return 0
	}
	if x.neg {
		return -1
	}
	return 1
}

// Neg sets z to -x and returns z.
func (z *Int) Neg(x *Int) *Int {
	if z != x {
		z.Set(x)
	}
	z.neg = len(z.abs) > 0 && !z.neg
	return z
}

// Abs sets z to |x| (the absolute value of x) and returns z.
func (z *Int) Abs(x *Int) *Int {
	if z != x {
		z.Set(x)
	}
	z.neg = false
	return z
}

// BitLen returns the length of the absolute value of x in bits. The
// bit length of 0 is 0.
func (x *Int) BitLen() int {
	if len(x.abs) == 0 {
		return 0
	}
	return (len(x.abs)-1)*8 + bits.Len8(x.abs[0])
}

// Bytes returns the absolute value of x as a big-endian byte slice.
// To use a fixed length slice, or a preallocated one, use FillBytes.
func (x *Int) Bytes() []byte {
	out := make([]byte, len(x.abs))
	copy(out, x.abs)
	return out
}

// SetBytes interprets buf as the bytes of a big-endian unsigned
// integer, sets z to that value, and returns z.
func (z *Int) SetBytes(buf []byte) *Int {
	// Skip leading zeros so the abs invariant is maintained.
	i := 0
	for i < len(buf) && buf[i] == 0 {
		i++
	}
	z.abs = append(z.abs[:0], buf[i:]...)
	z.neg = false
	return z
}

// CmpAbs compares the absolute values of x and y and returns:
//
//	-1 if |x| <  |y|
//	 0 if |x| == |y|
//	+1 if |x| >  |y|
func (x *Int) CmpAbs(y *Int) int {
	return cmpAbsBytes(x.abs, y.abs)
}

func cmpAbsBytes(a, b []byte) int {
	if len(a) != len(b) {
		if len(a) < len(b) {
			return -1
		}
		return 1
	}
	for i := range a {
		if a[i] != b[i] {
			if a[i] < b[i] {
				return -1
			}
			return 1
		}
	}
	return 0
}

// Cmp compares x and y and returns:
//
//	-1 if x <  y
//	 0 if x == y
//	+1 if x >  y
func (x *Int) Cmp(y *Int) int {
	switch {
	case x.neg == y.neg:
		c := cmpAbsBytes(x.abs, y.abs)
		if x.neg {
			return -c
		}
		return c
	case x.neg:
		return -1
	default:
		return 1
	}
}

// String returns the decimal representation of x as generated by x.Text(10).
func (x *Int) String() string {
	return x.Text(10)
}

// Text returns the string representation of x in the given base. Base
// must be between 2 and 62, inclusive. The result uses the lower-case
// letters 'a' to 'z' for digit values 10 to 35, and the upper-case
// letters 'A' to 'Z' for digit values 36 to 61. No prefix (such as "0x")
// is added to the string. If x is a nil pointer it returns "<nil>".
func (x *Int) Text(base int) string {
	// Like Go, the nil check comes before base validation:
	// (*Int)(nil).Text(1) is "<nil>", not a panic.
	if x == nil {
		return "<nil>"
	}
	if base < 2 || base > 62 {
		panic("invalid base")
	}
	return text(x.neg, x.abs, base)
}

// SetString sets z to the value of s, interpreted in the given base,
// and returns z and a boolean indicating success. The entire string
// (not just a prefix) must be valid for success. If SetString fails,
// the value of z is undefined but the returned value is nil.
//
// The base argument must be 0 or a value between 2 and 62, inclusive;
// otherwise SetString panics. For base 0, the number prefix determines
// the actual conversion base: a prefix of "0b" or "0B" selects base 2,
// "0", "0o" or "0O" selects base 8, and "0x" or "0X" selects base 16.
// Otherwise, the selected base is 10 and no prefix is accepted.
func (z *Int) SetString(s string, base int) (*Int, bool) {
	if base != 0 && (base < 2 || base > 62) {
		panic("invalid number base " + strconv.Itoa(base))
	}
	neg, abs, ok := setString(s, base)
	if !ok {
		return nil, false
	}
	z.neg = neg && len(abs) > 0
	z.abs = abs
	return z, true
}

// Add sets z to the sum x+y and returns z.
func (z *Int) Add(x, y *Int) *Int {
	z.neg, z.abs = add(x.neg, x.abs, y.neg, y.abs)
	return z
}

// Sub sets z to the difference x-y and returns z.
func (z *Int) Sub(x, y *Int) *Int {
	z.neg, z.abs = sub(x.neg, x.abs, y.neg, y.abs)
	return z
}

// Mul sets z to the product x*y and returns z.
func (z *Int) Mul(x, y *Int) *Int {
	z.neg, z.abs = mul(x.neg, x.abs, y.neg, y.abs)
	return z
}

// Quo sets z to the quotient x/y for y != 0. If y == 0, a division-by-zero
// run-time panic occurs. Quo implements truncated division (like Go); see
// QuoRem for more details.
func (z *Int) Quo(x, y *Int) *Int {
	divCheck(y)
	qNeg, q, _, _ := quoRem(x.neg, x.abs, y.neg, y.abs)
	z.neg = qNeg
	z.abs = q
	return z
}

// Rem sets z to the remainder x%y for y != 0. If y == 0, a division-by-zero
// run-time panic occurs. Rem implements truncated modulus (like Go); see
// QuoRem for more details.
func (z *Int) Rem(x, y *Int) *Int {
	divCheck(y)
	_, _, rNeg, r := quoRem(x.neg, x.abs, y.neg, y.abs)
	z.neg = rNeg
	z.abs = r
	return z
}

// QuoRem sets z to the quotient x/y and r to the remainder x%y and
// returns the pair (z, r) for y != 0. If y == 0, a division-by-zero
// run-time panic occurs.
//
// QuoRem implements T-division and modulus (like Go):
//
//	q = x/y      with the result truncated to zero
//	r = x - y*q
//
// The two output receivers z and r must be distinct *Int values (i.e.
// z != r); passing the same pointer for both panics, since the natural
// "last write wins" semantics would silently overwrite the quotient
// with the remainder.
func (z *Int) QuoRem(x, y, r *Int) (*Int, *Int) {
	divCheck(y)
	if z == r {
		panic("big: QuoRem requires distinct *Int for quotient and remainder")
	}
	qNeg, q, rNeg, rem := quoRem(x.neg, x.abs, y.neg, y.abs)
	z.neg = qNeg
	z.abs = q
	r.neg = rNeg
	r.abs = rem
	return z, r
}

// Div sets z to the quotient x/y for y != 0. If y == 0, a division-by-zero
// run-time panic occurs. Div implements Euclidean division (unlike Go);
// see DivMod for more details.
func (z *Int) Div(x, y *Int) *Int {
	divCheck(y)
	qNeg, q, _, _ := divMod(x.neg, x.abs, y.neg, y.abs)
	z.neg = qNeg
	z.abs = q
	return z
}

// Mod sets z to the modulus x%y for y != 0. If y == 0, a division-by-zero
// run-time panic occurs. Mod implements Euclidean modulus (unlike Go);
// see DivMod for more details.
func (z *Int) Mod(x, y *Int) *Int {
	divCheck(y)
	_, _, rNeg, r := divMod(x.neg, x.abs, y.neg, y.abs)
	z.neg = rNeg
	z.abs = r
	return z
}

// DivMod sets z to the quotient x div y and m to the modulus x mod y
// and returns the pair (z, m) for y != 0. If y == 0, a division-by-zero
// run-time panic occurs.
//
// DivMod implements Euclidean division and modulus (unlike Go):
//
//	q = x div y  such that
//	m = x - y*q  with 0 <= m < |y|
//
// The two output receivers z and m must be distinct *Int values (i.e.
// z != m); passing the same pointer for both panics, since the natural
// "last write wins" semantics would silently overwrite the quotient
// with the modulus.
func (z *Int) DivMod(x, y, m *Int) (*Int, *Int) {
	divCheck(y)
	if z == m {
		panic("big: DivMod requires distinct *Int for quotient and modulus")
	}
	qNeg, q, mNeg, mod := divMod(x.neg, x.abs, y.neg, y.abs)
	z.neg = qNeg
	z.abs = q
	m.neg = mNeg
	m.abs = mod
	return z, m
}

// Native bridge stubs. Implementations live in int.go.

func add(aNeg bool, a []byte, bNeg bool, b []byte) (bool, []byte) // injected

func sub(aNeg bool, a []byte, bNeg bool, b []byte) (bool, []byte) // injected

func mul(aNeg bool, a []byte, bNeg bool, b []byte) (bool, []byte) // injected

func quoRem(xNeg bool, x []byte, yNeg bool, y []byte) (bool, []byte, bool, []byte) // injected

func divMod(xNeg bool, x []byte, yNeg bool, y []byte) (bool, []byte, bool, []byte) // injected

func setString(s string, base int) (bool, []byte, bool) // injected

func text(neg bool, abs []byte, base int) string // injected
