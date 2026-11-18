package bls12381

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"

	bls "github.com/consensys/gnark-crypto/ecc/bls12-381"
	"github.com/consensys/gnark-crypto/ecc/bls12-381/fp"
	"github.com/consensys/gnark-crypto/ecc/bls12-381/fr"
	"github.com/stretchr/testify/require"
)

// g1Bytes returns the compressed G1 generator, a valid public key.
func g1Bytes() []byte {
	g1 := g1Generator()
	b := g1.Bytes()
	return b[:]
}

// signers returns n random secret keys, their concatenated compressed public
// keys, and the compressed aggregate signature of msg by all of them.
func signers(tb testing.TB, n int, msg []byte) (pubkeys, sig []byte) {
	tb.Helper()

	g1 := g1Generator()
	h, err := bls.HashToG2(msg, ethDST)
	require.NoError(tb, err)

	var aggSig bls.G2Jac
	for range n {
		var sk fr.Element
		_, err := sk.SetRandom()
		require.NoError(tb, err)
		var skInt big.Int
		sk.BigInt(&skInt)

		var pk bls.G1Affine
		pk.ScalarMultiplication(&g1, &skInt)
		pkb := pk.Bytes()
		pubkeys = append(pubkeys, pkb[:]...)

		var s bls.G2Affine
		s.ScalarMultiplication(&h, &skInt)
		aggSig.AddMixed(&s)
	}
	var aggAff bls.G2Affine
	aggAff.FromJacobian(&aggSig)
	sb := aggAff.Bytes()
	return pubkeys, sb[:]
}

func TestFastAggregateVerify(t *testing.T) {
	t.Parallel()

	msg := make([]byte, 32)
	_, _ = rand.Read(msg)
	pubkeys, sig := signers(t, 8, msg)
	var infSig bls.G2Affine
	infSigBytes := infSig.Bytes()

	require.True(t, X_fastAggregateVerify(pubkeys, msg, sig))

	t.Run("wrong message", func(t *testing.T) {
		t.Parallel()
		require.False(t, X_fastAggregateVerify(pubkeys, []byte("other"), sig))
	})
	t.Run("missing signer", func(t *testing.T) {
		t.Parallel()
		require.False(t, X_fastAggregateVerify(pubkeys[pubKeySize:], msg, sig))
	})
	t.Run("malformed lengths", func(t *testing.T) {
		t.Parallel()
		require.False(t, X_fastAggregateVerify(nil, msg, sig))
		require.False(t, X_fastAggregateVerify(pubkeys[1:], msg, sig))
		require.False(t, X_fastAggregateVerify(pubkeys, msg, sig[1:]))
	})
	t.Run("infinity pubkey", func(t *testing.T) {
		t.Parallel()
		var inf bls.G1Affine
		ib := inf.Bytes()
		require.False(t, X_fastAggregateVerify(append(append([]byte{}, pubkeys...), ib[:]...), msg, sig))
	})
	t.Run("identity aggregate", func(t *testing.T) {
		t.Parallel()
		// Keys that sum to the identity, each a valid non-identity point,
		// with the identity signature: e(0, H(m)) == e(-G1, 0) for every m.
		// Once with a random key, once with the generator.
		for _, key := range [][]byte{pubkeys[:pubKeySize], g1Bytes()} {
			var p, neg bls.G1Affine
			_, err := p.SetBytes(key)
			require.NoError(t, err)
			neg.Neg(&p)
			nb := neg.Bytes()
			pair := append(append([]byte{}, key...), nb[:]...)
			require.False(t, X_fastAggregateVerify(pair, msg, infSigBytes[:]))
			require.False(t, X_fastAggregateVerify(pair, []byte("any message"), infSigBytes[:]))
		}
	})
	t.Run("infinity signature", func(t *testing.T) {
		t.Parallel()
		require.False(t, X_fastAggregateVerify(pubkeys, msg, infSigBytes[:]))
	})
	t.Run("garbage signature", func(t *testing.T) {
		t.Parallel()
		bad := append([]byte{}, sig...)
		bad[10] ^= 0xff
		require.False(t, X_fastAggregateVerify(pubkeys, msg, bad))
	})
}

func BenchmarkFastAggregateVerify(b *testing.B) {
	for _, n := range []int{1, 64, 256, 512} {
		b.Run(fmt.Sprint(n), func(b *testing.B) {
			msg := make([]byte, 32)
			pubkeys, sig := signers(b, n, msg)
			b.ResetTimer()
			for range b.N {
				if !X_fastAggregateVerify(pubkeys, msg, sig) {
					b.Fatal("verify failed")
				}
			}
		})
	}
}

// committee returns n random secret keys' compressed public keys, their
// compressed aggregate, and the aggregate signature of msg by the members
// for which signs(i) is true.
func committee(tb testing.TB, n int, msg []byte, signs func(int) bool) (pubkeys [][]byte, aggregate, sig []byte) {
	tb.Helper()
	g1 := g1Generator()
	h, err := bls.HashToG2(msg, ethDST)
	require.NoError(tb, err)
	var aggPk bls.G1Jac
	var aggSig bls.G2Jac
	for i := range n {
		var sk fr.Element
		_, err := sk.SetRandom()
		require.NoError(tb, err)
		var skInt big.Int
		sk.BigInt(&skInt)
		var pk bls.G1Affine
		pk.ScalarMultiplication(&g1, &skInt)
		pkb := pk.Bytes()
		pubkeys = append(pubkeys, pkb[:])
		aggPk.AddMixed(&pk)
		if signs(i) {
			var s bls.G2Affine
			s.ScalarMultiplication(&h, &skInt)
			aggSig.AddMixed(&s)
		}
	}
	var a bls.G1Affine
	a.FromJacobian(&aggPk)
	ab := a.Bytes()
	var s bls.G2Affine
	s.FromJacobian(&aggSig)
	sb := s.Bytes()
	return pubkeys, ab[:], sb[:]
}

func TestFastAggregateVerifyExcluding(t *testing.T) {
	t.Parallel()

	msg := []byte("signing root")
	absent := map[int]bool{2: true, 5: true, 11: true}
	pubkeys, aggregate, sig := committee(t, 16, msg, func(i int) bool { return !absent[i] })
	var excluded, signers []byte
	for i, pk := range pubkeys {
		if absent[i] {
			excluded = append(excluded, pk...)
		} else {
			signers = append(signers, pk...)
		}
	}

	require.True(t, X_fastAggregateVerifyExcluding(aggregate, excluded, msg, sig))
	// Same answer as verifying over the signers' keys.
	require.True(t, X_fastAggregateVerify(signers, msg, sig))

	t.Run("wrong excluded set", func(t *testing.T) {
		t.Parallel()
		require.False(t, X_fastAggregateVerifyExcluding(aggregate, excluded[pubKeySize:], msg, sig))
		require.False(t, X_fastAggregateVerifyExcluding(aggregate, append(append([]byte{}, excluded...), pubkeys[0]...), msg, sig))
	})
	t.Run("wrong message", func(t *testing.T) {
		t.Parallel()
		require.False(t, X_fastAggregateVerifyExcluding(aggregate, excluded, []byte("other"), sig))
	})
	t.Run("nobody signed", func(t *testing.T) {
		t.Parallel()
		var all []byte
		for _, pk := range pubkeys {
			all = append(all, pk...)
		}
		require.False(t, X_fastAggregateVerifyExcluding(aggregate, all, msg, sig))
	})
	t.Run("nobody signed, identity signature", func(t *testing.T) {
		t.Parallel()
		// Excluding everyone leaves the identity as the signers' key, which
		// would verify the identity signature over any message:
		// verifyAggregate's identity check alone refuses it.
		var all []byte
		for _, pk := range pubkeys {
			all = append(all, pk...)
		}
		var inf bls.G2Affine
		infSig := inf.Bytes()
		require.False(t, X_fastAggregateVerifyExcluding(aggregate, all, []byte("any message"), infSig[:]))
	})
	t.Run("malformed", func(t *testing.T) {
		t.Parallel()
		require.False(t, X_fastAggregateVerifyExcluding(aggregate[1:], excluded, msg, sig))
		require.False(t, X_fastAggregateVerifyExcluding(aggregate, excluded[1:], msg, sig))
		var inf bls.G1Affine
		ib := inf.Bytes()
		require.False(t, X_fastAggregateVerifyExcluding(ib[:], excluded, msg, sig))
		require.False(t, X_fastAggregateVerifyExcluding(aggregate, append(append([]byte{}, excluded...), ib[:]...), msg, sig))
	})
}

func BenchmarkFastAggregateVerifyExcluding(b *testing.B) {
	for _, missing := range []int{0, 8, 32, 128} {
		b.Run(fmt.Sprint(missing), func(b *testing.B) {
			msg := make([]byte, 32)
			pubkeys, aggregate, sig := committee(b, 512, msg, func(i int) bool { return i >= missing })
			var excluded []byte
			for _, pk := range pubkeys[:missing] {
				excluded = append(excluded, pk...)
			}
			b.ResetTimer()
			for range b.N {
				if !X_fastAggregateVerifyExcluding(aggregate, excluded, msg, sig) {
					b.Fatal("verify failed")
				}
			}
		})
	}
}

// keyFrom returns the compressed public key of secret s and its signature
// s*H(msg).
func keyFrom(tb testing.TB, s int64, msg []byte) (pk []byte, sig bls.G2Affine) {
	tb.Helper()
	g1 := g1Generator()
	h, err := bls.HashToG2(msg, ethDST)
	require.NoError(tb, err)
	var p bls.G1Affine
	p.ScalarMultiplication(&g1, big.NewInt(s))
	b := p.Bytes()
	sig.ScalarMultiplication(&h, big.NewInt(s))
	return b[:], sig
}

// sigBytes returns the compressed sum of signatures.
func sigBytes(points ...bls.G2Affine) []byte {
	var sum bls.G2Jac
	for i := range points {
		sum.AddMixed(&points[i])
	}
	var a bls.G2Affine
	a.FromJacobian(&sum)
	b := a.Bytes()
	return b[:]
}

// keySum returns the compressed sum of compressed keys.
func keySum(tb testing.TB, keys ...[]byte) []byte {
	tb.Helper()
	var sum bls.G1Jac
	for _, k := range keys {
		var p bls.G1Affine
		_, err := p.SetBytes(k)
		require.NoError(tb, err)
		sum.AddMixed(&p)
	}
	var a bls.G1Affine
	a.FromJacobian(&sum)
	b := a.Bytes()
	return b[:]
}

func cat(parts ...[]byte) []byte {
	var out []byte
	for _, p := range parts {
		out = append(out, p...)
	}
	return out
}

// offSubgroupG1 returns the encoding of a point on y^2 = x^3 + 4 outside the
// prime-order subgroup.
func offSubgroupG1(t *testing.T) []byte {
	t.Helper()
	var p bls.G1Affine
	var rhs, four fp.Element
	four.SetUint64(4)
	for i := uint64(1); ; i++ {
		p.X.SetUint64(i)
		rhs.Square(&p.X).Mul(&rhs, &p.X).Add(&rhs, &four)
		if p.Y.Sqrt(&rhs) == nil {
			continue
		}
		require.True(t, p.IsOnCurve())
		if !p.IsInSubGroup() {
			b := p.Bytes()
			return b[:]
		}
	}
}

// A key may sit on a sync committee more than once; summing it twice takes
// the curve's doubling branch.
func TestFastAggregateVerifyRepeatedKey(t *testing.T) {
	t.Parallel()
	msg := []byte("signing root")
	pk, s := keyFrom(t, 0x1234567, msg)
	q, sq := keyFrom(t, 0x89abcdef, msg)
	require.True(t, X_fastAggregateVerify(cat(pk, pk, q), msg, sigBytes(s, s, sq)))
	require.False(t, X_fastAggregateVerify(cat(pk, pk, q), msg, sigBytes(s, sq)))
}

// The decoder ignores bytes past a point's encoding, so only the natives'
// exact length checks refuse a trailing byte.
func TestTrailingByte(t *testing.T) {
	t.Parallel()
	msg := []byte("signing root")
	pk, s := keyFrom(t, 0x1234567, msg)
	q, _ := keyFrom(t, 0x89abcdef, msg)
	sig := sigBytes(s)
	require.True(t, X_fastAggregateVerify(pk, msg, sig))
	require.False(t, X_fastAggregateVerify(pk, msg, cat(sig, []byte{0})))
	aggregate := keySum(t, pk, q)
	require.True(t, X_fastAggregateVerifyExcluding(aggregate, q, msg, sig))
	require.False(t, X_fastAggregateVerifyExcluding(cat(aggregate, []byte{0}), q, msg, sig))
	require.False(t, X_fastAggregateVerifyExcluding(aggregate, q, msg, cat(sig, []byte{0})))
}

func TestOutsideSubgroup(t *testing.T) {
	t.Parallel()
	msg := make([]byte, 32)
	pubkeys, aggregate, sig := committee(t, 4, msg, func(int) bool { return true })
	bad := offSubgroupG1(t)
	require.False(t, X_fastAggregateVerify(cat(pubkeys[0], bad), msg, sig))
	require.False(t, X_fastAggregateVerifyExcluding(bad, nil, msg, sig))
	require.False(t, X_fastAggregateVerifyExcluding(aggregate, bad, msg, sig))
}

// A second encoding of a key fails the caller's Merkle proof, but one of a
// signature would get through, so both must be refused here.
func TestNonCanonical(t *testing.T) {
	t.Parallel()
	msg := []byte("signing root")
	pk, s := keyFrom(t, 0x1234567, msg)
	sig := sigBytes(s)
	modulus := fp.Modulus().FillBytes(make([]byte, fp.Bytes))

	xIsP := cat(modulus) // x = p, compressed flag set
	xIsP[0] |= 0x80
	infPayload := make([]byte, pubKeySize) // infinity flag with a nonzero x
	infPayload[0], infPayload[pubKeySize-1] = 0xc0, 0x01
	badMask := make([]byte, pubKeySize) // compressed, infinity and sign all set
	badMask[0] = 0xe0
	for name, key := range map[string][]byte{"x = p": xIsP, "infinity payload": infPayload, "invalid mask": badMask} {
		require.False(t, X_fastAggregateVerify(cat(pk, key), msg, sig), name)
	}

	sigXIsP := cat(modulus, make([]byte, fp.Bytes)) // x.A1 = p
	sigXIsP[0] |= 0x80
	sigInfPayload := make([]byte, signatureSize)
	sigInfPayload[0], sigInfPayload[signatureSize-1] = 0xc0, 0x01
	sigBadMask := make([]byte, signatureSize)
	sigBadMask[0] = 0xe0
	for name, bad := range map[string][]byte{"x = p": sigXIsP, "infinity payload": sigInfPayload, "invalid mask": sigBadMask} {
		require.False(t, X_fastAggregateVerify(pk, msg, bad), name)
	}
}

// A member holding two seats and absent from both is subtracted once per
// seat.
func TestFastAggregateVerifyExcludingRepeatedMember(t *testing.T) {
	t.Parallel()
	msg := []byte("signing root")
	k1, s1 := keyFrom(t, 0x1234567, msg)
	k2, _ := keyFrom(t, 0x89abcdef, msg)
	k3, s3 := keyFrom(t, 0x7654321, msg)
	aggregate := keySum(t, k1, k2, k2, k3)
	require.True(t, X_fastAggregateVerifyExcluding(aggregate, cat(k2, k2), msg, sigBytes(s1, s3)))
	require.False(t, X_fastAggregateVerifyExcluding(aggregate, k2, msg, sigBytes(s1, s3)))
}

// An identity aggregate minus -pk is pk: without the aggregate's identity
// check, one key holder would sign for it.
func TestFastAggregateVerifyExcludingIdentityAggregate(t *testing.T) {
	t.Parallel()
	msg := []byte("signing root")
	pk, s := keyFrom(t, 0x1234567, msg)
	var p, neg bls.G1Affine
	_, err := p.SetBytes(pk)
	require.NoError(t, err)
	neg.Neg(&p)
	nb := neg.Bytes()
	var inf bls.G1Affine
	ib := inf.Bytes()
	require.False(t, X_fastAggregateVerifyExcluding(ib[:], nb[:], msg, sigBytes(s)))
}
