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

// signers returns n random secret keys, their concatenated compressed public
// keys, and the compressed aggregate signature of msg by all of them.
func signers(t testing.TB, n int, msg []byte) (pubkeys, sig []byte) {
	t.Helper()

	_, _, g1, _ := bls.Generators()
	h, err := bls.HashToG2(msg, ethDST)
	require.NoError(t, err)

	var aggSig bls.G2Jac
	for range n {
		var sk fr.Element
		_, err := sk.SetRandom()
		require.NoError(t, err)
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

	require.True(t, X_fastAggregateVerify(pubkeys, msg, sig))

	t.Run("wrong message", func(t *testing.T) {
		require.False(t, X_fastAggregateVerify(pubkeys, []byte("other"), sig))
	})
	t.Run("missing signer", func(t *testing.T) {
		require.False(t, X_fastAggregateVerify(pubkeys[pubKeySize:], msg, sig))
	})
	t.Run("malformed lengths", func(t *testing.T) {
		require.False(t, X_fastAggregateVerify(nil, msg, sig))
		require.False(t, X_fastAggregateVerify(pubkeys[1:], msg, sig))
		require.False(t, X_fastAggregateVerify(pubkeys, msg, sig[1:]))
	})
	t.Run("infinity pubkey", func(t *testing.T) {
		var inf bls.G1Affine
		ib := inf.Bytes()
		require.False(t, X_fastAggregateVerify(append(append([]byte{}, pubkeys...), ib[:]...), msg, sig))
	})
	t.Run("garbage signature", func(t *testing.T) {
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
func committee(t testing.TB, n int, msg []byte, signs func(int) bool) (pubkeys [][]byte, aggregate, sig []byte) {
	t.Helper()
	_, _, g1, _ := bls.Generators()
	h, err := bls.HashToG2(msg, ethDST)
	require.NoError(t, err)
	var aggPk bls.G1Jac
	var aggSig bls.G2Jac
	for i := range n {
		var sk fr.Element
		_, err := sk.SetRandom()
		require.NoError(t, err)
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
		require.False(t, X_fastAggregateVerifyExcluding(aggregate, excluded[pubKeySize:], msg, sig))
		require.False(t, X_fastAggregateVerifyExcluding(aggregate, append(append([]byte{}, excluded...), pubkeys[0]...), msg, sig))
	})
	t.Run("wrong message", func(t *testing.T) {
		require.False(t, X_fastAggregateVerifyExcluding(aggregate, excluded, []byte("other"), sig))
	})
	t.Run("nobody signed", func(t *testing.T) {
		var all []byte
		for _, pk := range pubkeys {
			all = append(all, pk...)
		}
		require.False(t, X_fastAggregateVerifyExcluding(aggregate, all, msg, sig))
	})
	t.Run("malformed", func(t *testing.T) {
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
