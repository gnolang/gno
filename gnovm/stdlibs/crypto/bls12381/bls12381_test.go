package bls12381

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"

	bls "github.com/consensys/gnark-crypto/ecc/bls12-381"
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
