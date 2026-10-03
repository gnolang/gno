package evmnet

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/big"
	"testing"

	"golang.org/x/crypto/sha3"
)

// This file is a plain math/big implementation of hash-to-G1 (RFC 9380
// expand_message_xmd + Shallue-van de Woestijne, Z = 1) written the same way
// gno.land/p/drand/v0 does it in gno. Checking it against kyber pins down the constants
// and the exact branch rules the gno port must follow.

var (
	fp, _ = new(big.Int).SetString("30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47", 16)
	bigB  = big.NewInt(3)
)

func mod(x *big.Int) *big.Int { return x.Mod(x, fp) }

func svdwConstants() (c1, c2, c3, c4 *big.Int) {
	// Z = 1, A = 0, B = 3.
	c1 = big.NewInt(4)                                                     // g(Z) = Z^3 + B
	c2 = mod(new(big.Int).Neg(new(big.Int).ModInverse(big.NewInt(2), fp))) // -Z / 2
	c3 = new(big.Int).ModSqrt(mod(big.NewInt(-12)), fp)                    // sqrt(-g(Z) * 3Z^2)
	if c3.Bit(0) == 1 {
		c3 = mod(new(big.Int).Neg(c3))
	}
	inv3 := new(big.Int).ModInverse(big.NewInt(3), fp)
	c4 = mod(new(big.Int).Mul(big.NewInt(-16), inv3)) // -4 g(Z) / 3Z^2
	return
}

func gx(x *big.Int) *big.Int {
	y := new(big.Int).Exp(x, big.NewInt(3), fp)
	return mod(y.Add(y, bigB))
}

// sqrtIfSquare mirrors kyber: zero is treated as a non-square.
func sqrtIfSquare(v *big.Int) (*big.Int, bool) {
	if v.Sign() == 0 {
		return nil, false
	}
	e := new(big.Int).Add(fp, big.NewInt(1))
	e.Rsh(e, 2)
	y := new(big.Int).Exp(v, e, fp)
	if new(big.Int).Exp(y, big.NewInt(2), fp).Cmp(v) != 0 {
		return nil, false
	}
	return y, true
}

func mapToPointRef(u *big.Int) (x, y *big.Int) {
	c1, c2, c3, c4 := svdwConstants()
	one := big.NewInt(1)
	tv1 := mod(new(big.Int).Mul(new(big.Int).Mul(u, u), c1))
	tv2 := mod(new(big.Int).Add(one, tv1))
	tv1 = mod(new(big.Int).Sub(one, tv1))
	tv3 := new(big.Int).Exp(mod(new(big.Int).Mul(tv1, tv2)), new(big.Int).Sub(fp, big.NewInt(2)), fp)
	tv5 := mod(new(big.Int).Mul(new(big.Int).Mul(new(big.Int).Mul(u, tv1), tv3), c3))
	x1 := mod(new(big.Int).Sub(c2, tv5))
	x2 := mod(new(big.Int).Add(c2, tv5))
	tv8 := mod(new(big.Int).Mul(new(big.Int).Mul(tv2, tv2), tv3))
	x3 := mod(new(big.Int).Add(one, mod(new(big.Int).Mul(new(big.Int).Mul(tv8, tv8), c4))))

	for _, cand := range []*big.Int{x1, x2, x3} {
		if r, ok := sqrtIfSquare(gx(cand)); ok {
			x, y = cand, r
			break
		}
	}
	if y == nil { // unreachable for SVDW, x3 is always a square
		panic("no square")
	}
	if u.Bit(0) != y.Bit(0) {
		y = mod(new(big.Int).Neg(y))
	}
	return x, y
}

func expandMsgXmd(dst, msg []byte, outLen int) []byte {
	h := sha3.NewLegacyKeccak256
	dstPrime := append(append([]byte{}, dst...), byte(len(dst)))
	b0h := h()
	b0h.Write(make([]byte, 136)) // keccak256 block size
	b0h.Write(msg)
	b0h.Write([]byte{byte(outLen >> 8), byte(outLen), 0})
	b0h.Write(dstPrime)
	b0 := b0h.Sum(nil)
	out := []byte{}
	prev := make([]byte, 32)
	for i := 1; len(out) < outLen; i++ {
		in := make([]byte, 32)
		for j := range in {
			in[j] = b0[j] ^ prev[j]
		}
		if i == 1 {
			in = b0
		}
		bh := h()
		bh.Write(in)
		bh.Write([]byte{byte(i)})
		bh.Write(dstPrime)
		prev = bh.Sum(nil)
		out = append(out, prev...)
	}
	return out[:outLen]
}

func addAffine(x1, y1, x2, y2 *big.Int) (*big.Int, *big.Int) {
	var l *big.Int
	if x1.Cmp(x2) == 0 {
		if y1.Cmp(y2) != 0 {
			return big.NewInt(0), big.NewInt(0)
		}
		num := mod(new(big.Int).Mul(big.NewInt(3), new(big.Int).Mul(x1, x1)))
		den := new(big.Int).ModInverse(mod(new(big.Int).Mul(big.NewInt(2), y1)), fp)
		l = mod(num.Mul(num, den))
	} else {
		num := mod(new(big.Int).Sub(y2, y1))
		den := new(big.Int).ModInverse(mod(new(big.Int).Sub(x2, x1)), fp)
		l = mod(num.Mul(num, den))
	}
	x3 := mod(new(big.Int).Sub(new(big.Int).Sub(new(big.Int).Mul(l, l), x1), x2))
	y3 := mod(new(big.Int).Sub(new(big.Int).Mul(l, new(big.Int).Sub(x1, x3)), y1))
	return x3, y3
}

func hashToG1Ref(msg []byte) []byte {
	b := expandMsgXmd([]byte(DST), msg, 96)
	u0 := mod(new(big.Int).SetBytes(b[:48]))
	u1 := mod(new(big.Int).SetBytes(b[48:]))
	x0, y0 := mapToPointRef(u0)
	x1, y1 := mapToPointRef(u1)
	x, y := addAffine(x0, y0, x1, y1)
	out := make([]byte, 64)
	x.FillBytes(out[:32])
	y.FillBytes(out[32:])
	return out
}

func TestHashToG1MatchesKyber(t *testing.T) {
	msgs := [][]byte{{}, []byte("abc"), Message(1), Message(liveRound)}
	for i := 0; i < 200; i++ {
		s := sha256.Sum256([]byte(fmt.Sprint(i)))
		msgs = append(msgs, s[:])
	}
	for _, m := range msgs {
		if got, want := hashToG1Ref(m), HashToG1(m); !bytes.Equal(got, want) {
			t.Fatalf("msg %x:\n ref   %x\n kyber %x", m, got, want)
		}
	}
}

func TestPrintSVDWConstants(t *testing.T) {
	c1, c2, c3, c4 := svdwConstants()
	for i, c := range []*big.Int{c1, c2, c3, c4} {
		t.Logf("c%d = %064x", i+1, c)
	}
}
