package drand

import (
	"crypto/modexp"
	"encoding/binary"
	"math/bits"
)

// fe is an element of the BN254 base field, as four little-endian 64-bit
// limbs, always fully reduced (< p).
//
// All the heavy arithmetic (squaring, inverse, square root, reducing a wide
// integer) goes to the native modexp; gno code only does limb additions.
type fe [4]uint64

var (
	pLimbs = fe{0x3c208c16d87cfd47, 0x97816a916871ca8d, 0xb85045b68181585d, 0x30644e72e131a029}
	pBytes = mustHex("30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47")

	// Exponents: p-2 (inverse), (p+1)/4 (square root, p = 3 mod 4).
	expInverse = mustHex("30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd45")
	expSqrt    = mustHex("0c19139cb84c680a6e14116da060561765e05aa45a1c72a34f082305b61f3f52")

	expOne   = []byte{1}
	expTwo   = []byte{2}
	expThree = []byte{3}
)

func feFromUint(v uint64) fe { return fe{v, 0, 0, 0} }

// feFromBytes reduces a big-endian integer of up to 64 bytes mod p.
func feFromBytes(b []byte) fe {
	return feFromCanonicalUnchecked(modexp.ModExp(b, expOne, pBytes))
}

// feFromCanonical decodes exactly 32 big-endian bytes and reports whether the
// value is already reduced.
func feFromCanonical(b []byte) (fe, bool) {
	e := feFromCanonicalUnchecked(b)
	return e, e.lt(pLimbs)
}

func feFromCanonicalUnchecked(b []byte) fe {
	be := binary.BigEndian
	return fe{be.Uint64(b[24:]), be.Uint64(b[16:]), be.Uint64(b[8:]), be.Uint64(b[0:])}
}

// bytes returns the 32-byte big-endian encoding.
func (a fe) bytes() []byte {
	out := make([]byte, 32)
	be := binary.BigEndian
	be.PutUint64(out[0:], a[3])
	be.PutUint64(out[8:], a[2])
	be.PutUint64(out[16:], a[1])
	be.PutUint64(out[24:], a[0])
	return out
}

func (a fe) isZero() bool { return a[0]|a[1]|a[2]|a[3] == 0 }

func (a fe) isOdd() bool { return a[0]&1 == 1 }

func (a fe) lt(b fe) bool {
	for i := 3; i >= 0; i-- {
		if a[i] != b[i] {
			return a[i] < b[i]
		}
	}
	return false
}

func feAdd(a, b fe) fe {
	r := rawAdd(a, b) // p < 2^254, so a+b never overflows 256 bits.
	if !r.lt(pLimbs) {
		r = rawSub(r, pLimbs)
	}
	return r
}

func feSub(a, b fe) fe {
	if a.lt(b) {
		return rawSub(rawAdd(a, pLimbs), b)
	}
	return rawSub(a, b)
}

func feNeg(a fe) fe {
	if a.isZero() {
		return a
	}
	return rawSub(pLimbs, a)
}

func rawAdd(a, b fe) fe {
	var r fe
	var c uint64
	r[0], c = bits.Add64(a[0], b[0], 0)
	r[1], c = bits.Add64(a[1], b[1], c)
	r[2], c = bits.Add64(a[2], b[2], c)
	r[3], _ = bits.Add64(a[3], b[3], c)
	return r
}

func rawSub(a, b fe) fe {
	var r fe
	var c uint64
	r[0], c = bits.Sub64(a[0], b[0], 0)
	r[1], c = bits.Sub64(a[1], b[1], c)
	r[2], c = bits.Sub64(a[2], b[2], c)
	r[3], _ = bits.Sub64(a[3], b[3], c)
	return r
}

// feMul uses quarter-squares, ab = ((a+b)/2)^2 - ((a-b)/2)^2, so the only
// heavy work is two squarings done by the native modexp. A schoolbook
// 256x256 product interpreted by the VM costs several times more.
func feMul(a, b fe) fe {
	return feSub(feSquare(feHalf(feAdd(a, b))), feSquare(feHalf(feSub(a, b))))
}

// feHalf returns a/2 mod p.
func feHalf(a fe) fe {
	if a.isOdd() {
		a = rawAdd(a, pLimbs) // < 2^255, no overflow
	}
	a[0] = a[0]>>1 | a[1]<<63
	a[1] = a[1]>>1 | a[2]<<63
	a[2] = a[2]>>1 | a[3]<<63
	a[3] >>= 1
	return a
}

func feSquare(a fe) fe { return feExp(a, expTwo) }

func feExp(a fe, e []byte) fe {
	return feFromCanonicalUnchecked(modexp.ModExp(a.bytes(), e, pBytes))
}

// feInv returns a^-1, and 0 for 0 (same convention as RFC 9380 inv0).
func feInv(a fe) fe { return feExp(a, expInverse) }

// feSqrt returns a square root of a and whether a is a non-zero square.
// Zero is reported as a non-square, matching drand/kyber's legendre check.
func feSqrt(a fe) (fe, bool) {
	if a.isZero() {
		return a, false
	}
	r := feExp(a, expSqrt)
	return r, feSquare(r) == a
}
