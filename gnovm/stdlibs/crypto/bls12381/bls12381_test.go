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
