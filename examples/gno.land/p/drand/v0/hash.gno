package drand

import (
	"crypto/bn254"
	"crypto/keccak256"
)

// DST is the hash-to-curve domain separation tag of the evmnet scheme.
const DST = "BLS_SIG_BN254G1_XMD:KECCAK-256_SVDW_RO_NUL_"

const keccakRate = 136 // keccak256 block size in bytes

// Shallue-van de Woestijne constants for BN254 G1 (A = 0, B = 3, Z = 1),
// RFC 9380 section 6.6.1, cross-checked against drand/kyber by
// contribs/gnodrand/internal/evmnet/svdw_test.go. c1 = g(Z) = 4 is applied as two doublings.
var (
	feOne  = feFromUint(1)
	curveB = feFromUint(3)
	svdwC2 = mustFe("183227397098d014dc2822db40c0ac2ecbc0b548b438e5469e10460b6c3e7ea3") // -Z/2
	svdwC3 = mustFe("00000000000000016789af3a83522eb353c98fc6b36d713d5d8d1cc5dffffffa") // sqrt(-g(Z)*3Z^2), sgn0 = 0
	svdwC4 = mustFe("10216f7ba065e00de81ac1e7808072c9dd2b2385cd7b438469602eb24829a9bd") // -4g(Z)/3Z^2
)

// HashToG1 hashes msg onto G1 the way drand evmnet does (RFC 9380
// hash_to_curve with expand_message_xmd/keccak256 and the SVDW map) and
// returns the 64-byte affine encoding x|y.
func HashToG1(msg []byte) []byte {
	u := expandMsgXmd(msg, 96)
	p0 := mapToPoint(feFromBytes(u[:48]))
	p1 := mapToPoint(feFromBytes(u[48:]))
	in := make([]byte, 0, bn254.G1AddInputSize)
	in = append(in, p0...)
	in = append(in, p1...)
	out := bn254.G1Add(in)
	if out == nil {
		panic("drand: hash-to-curve produced an invalid point")
	}
	return out
}

// dstPrime is DST || I2OSP(len(DST), 1).
var dstPrime = append([]byte(DST), byte(len(DST)))

// expandMsgXmd implements RFC 9380 section 5.3.1 with keccak256 and DST.
func expandMsgXmd(msg []byte, outLen int) []byte {
	b0in := make([]byte, 0, keccakRate+len(msg)+3+len(dstPrime))
	b0in = append(b0in, make([]byte, keccakRate)...)
	b0in = append(b0in, msg...)
	b0in = append(b0in, byte(outLen>>8), byte(outLen), 0)
	b0in = append(b0in, dstPrime...)
	b0 := keccak256.Sum256(b0in)

	out := make([]byte, 0, outLen+32)
	var prev [32]byte // b_0 xor zero is b_0, so b_1 needs no special case
	for i := 1; len(out) < outLen; i++ {
		in := make([]byte, 0, 32+1+len(dstPrime))
		for j := 0; j < 32; j++ {
			in = append(in, b0[j]^prev[j])
		}
		in = append(in, byte(i))
		in = append(in, dstPrime...)
		prev = keccak256.Sum256(in)
		out = append(out, prev[:]...)
	}
	return out[:outLen]
}

// curveG returns x^3 + 3.
func curveG(x fe) fe {
	return feAdd(feExp(x, expThree), curveB)
}

// mapToPoint is the SVDW map, as in drand/kyber, returning the 64-byte
// affine encoding. Candidates x2 and x3 are only computed when needed.
func mapToPoint(u fe) []byte {
	tv1 := feSquare(u)
	tv1 = feAdd(tv1, tv1)
	tv1 = feAdd(tv1, tv1) // * c1 (= 4)
	tv2 := feAdd(feOne, tv1)
	tv1 = feSub(feOne, tv1)
	tv3 := feInv(feMul(tv1, tv2))
	tv5 := feMul(feMul(feMul(u, tv1), tv3), svdwC3)
	x := feSub(svdwC2, tv5) // x1
	y, ok := feSqrt(curveG(x))
	if !ok {
		x = feAdd(svdwC2, tv5) // x2
		y, ok = feSqrt(curveG(x))
	}
	if !ok {
		tv8 := feMul(feSquare(tv2), tv3)
		x = feAdd(feOne, feMul(feSquare(tv8), svdwC4)) // x3
		y, ok = feSqrt(curveG(x))
	}
	if !ok {
		panic("drand: svdw found no square") // unreachable: g(x3) is always square
	}
	if u.isOdd() != y.isOdd() {
		y = feNeg(y)
	}
	out := make([]byte, 0, 64)
	out = append(out, x.bytes()...)
	return append(out, y.bytes()...)
}
