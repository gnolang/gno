package drand

import (
	"encoding/hex"
	"testing"

	"gno.land/p/nt/uassert/v0"
)

func TestHashToG1Vectors(t *testing.T) {
	for _, v := range hashVectors {
		got := hex.EncodeToString(HashToG1(mustHex(v.msg)))
		uassert.Equal(t, v.point, got, "msg "+v.msg)
	}
}

func TestVerifyBeacons(t *testing.T) {
	for _, v := range beaconVectors {
		sig := mustHex(v.sig)
		uassert.True(t, Verify(v.round, sig), "valid beacon rejected")
		r := Randomness(sig)
		uassert.Equal(t, v.randomness, hex.EncodeToString(r[:]))
	}
}

func TestVerifyRejects(t *testing.T) {
	v := beaconVectors[len(beaconVectors)-1]
	sig := mustHex(v.sig)

	uassert.False(t, Verify(v.round+1, sig), "wrong round")
	uassert.False(t, Verify(0, sig), "round 0")
	uassert.False(t, Verify(v.round, sig[:63]), "short sig")
	uassert.False(t, Verify(v.round, append(sig, 0)), "long sig")

	// Another round's valid signature.
	other := mustHex(beaconVectors[0].sig)
	uassert.False(t, Verify(v.round, other), "signature of another round")

	// Bit flip: off the curve (or a different point).
	bad := append([]byte(nil), sig...)
	bad[40] ^= 1
	uassert.False(t, Verify(v.round, bad), "tampered sig")

	// Point at infinity.
	uassert.False(t, Verify(v.round, make([]byte, 64)), "infinity")

	// Non-canonical x coordinate (x + p).
	x := feFromCanonicalUnchecked(sig[:32])
	nc := append(rawAdd(x, pLimbs).bytes(), sig[32:]...)
	uassert.False(t, Verify(v.round, nc), "non-canonical x")

	// -sig is on the curve but is not the signature.
	neg := append(append([]byte(nil), sig[:32]...), feNeg(feFromCanonicalUnchecked(sig[32:])).bytes()...)
	uassert.False(t, Verify(v.round, neg), "negated sig")
}

func TestRoundTime(t *testing.T) {
	uassert.Equal(t, uint64(1), RoundAt(GenesisTime))
	uassert.Equal(t, GenesisTime, TimeOf(1))
	uassert.Equal(t, uint64(0), RoundAt(GenesisTime-1))
	for _, r := range []uint64{1, 2, 1000, 21167346} {
		uassert.Equal(t, r, RoundAt(TimeOf(r)))
		uassert.Equal(t, r, RoundAt(TimeOf(r)+Period-1))
		uassert.Equal(t, r+1, RoundAt(TimeOf(r)+Period))
	}
}

func TestDeriveAndRand(t *testing.T) {
	r := Randomness(mustHex(beaconVectors[0].sig))
	a, b := Derive(r, "a"), Derive(r, "b")
	uassert.True(t, a != b, "salts must give different values")
	uassert.True(t, a == Derive(r, "a"), "Derive must be deterministic")

	x, y := NewRand(a), NewRand(a)
	for i := 0; i < 10; i++ {
		uassert.Equal(t, x.IntN(1000), y.IntN(1000))
	}
}

func TestFieldOps(t *testing.T) {
	a := mustFe("2a1b3c4d5e6f708192a3b4c5d6e7f8091a2b3c4d5e6f708192a3b4c5d6e7f809")
	b := mustFe("0fedcba9876543210fedcba9876543210fedcba9876543210fedcba987654321")
	uassert.True(t, feSub(feAdd(a, b), b) == a, "add/sub")
	uassert.True(t, feAdd(a, feNeg(a)).isZero(), "neg")
	uassert.True(t, feMul(a, feInv(a)) == feOne, "inv")
	s, ok := feSqrt(feSquare(a))
	uassert.True(t, ok, "square has a root")
	uassert.True(t, feSquare(s) == feSquare(a), "sqrt")
	_, ok = feSqrt(fe{})
	uassert.False(t, ok, "zero is reported as non-square")
	// p-1 is not a square since p = 3 mod 4.
	_, ok = feSqrt(feNeg(feOne))
	uassert.False(t, ok, "-1 is not a square")
	// 64-byte reduction.
	big := make([]byte, 64)
	for i := range big {
		big[i] = 0xff
	}
	r := feFromBytes(big)
	uassert.True(t, r.lt(pLimbs), "reduced")
}
