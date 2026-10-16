package big

import (
	"math/big"
)

// Wire format: a *big.Int is represented as (neg bool, abs []byte) where
// abs is the big-endian unsigned magnitude with no leading zero bytes
// (empty/nil for zero). See gnovm/adr/pr5678_math_big_stdlib.md for the
// design rationale.
//
// All Gno-side setters maintain this canonical form. toBig defends
// against a non-canonical input anyway by stripping leading zeros and
// forcing neg=false when the magnitude is empty.

func X_add(aNeg bool, a []byte, bNeg bool, b []byte) (bool, []byte) {
	return fromBig(new(big.Int).Add(toBig(aNeg, a), toBig(bNeg, b)))
}

func X_sub(aNeg bool, a []byte, bNeg bool, b []byte) (bool, []byte) {
	return fromBig(new(big.Int).Sub(toBig(aNeg, a), toBig(bNeg, b)))
}

func X_mul(aNeg bool, a []byte, bNeg bool, b []byte) (bool, []byte) {
	return fromBig(new(big.Int).Mul(toBig(aNeg, a), toBig(bNeg, b)))
}

func X_quoRem(xNeg bool, x []byte, yNeg bool, y []byte) (bool, []byte, bool, []byte) {
	q, r := new(big.Int), new(big.Int)
	q.QuoRem(toBig(xNeg, x), toBig(yNeg, y), r)
	qNeg, qAbs := fromBig(q)
	rNeg, rAbs := fromBig(r)
	return qNeg, qAbs, rNeg, rAbs
}

func X_divMod(xNeg bool, x []byte, yNeg bool, y []byte) (bool, []byte, bool, []byte) {
	q, m := new(big.Int), new(big.Int)
	q.DivMod(toBig(xNeg, x), toBig(yNeg, y), m)
	qNeg, qAbs := fromBig(q)
	mNeg, mAbs := fromBig(m)
	return qNeg, qAbs, mNeg, mAbs
}

func X_setString(s string, base int) (bool, []byte, bool) {
	v, ok := new(big.Int).SetString(s, base)
	if !ok {
		return false, nil, false
	}
	neg, abs := fromBig(v)
	return neg, abs, true
}

func X_text(neg bool, abs []byte, base int) string {
	return toBig(neg, abs).Text(base)
}

func toBig(neg bool, abs []byte) *big.Int {
	// Skip leading zeros; an all-zero or empty abs decodes to zero.
	i := 0
	for i < len(abs) && abs[i] == 0 {
		i++
	}
	abs = abs[i:]
	x := new(big.Int).SetBytes(abs)
	// neg is honored only for non-zero magnitudes — there is no negative
	// zero in *big.Int and the wire format mirrors that.
	if neg && len(abs) > 0 {
		x.Neg(x)
	}
	return x
}

func fromBig(x *big.Int) (bool, []byte) {
	sign := x.Sign()
	if sign == 0 {
		return false, nil
	}
	// Bytes returns the magnitude regardless of sign — no Abs needed.
	return sign < 0, x.Bytes()
}
