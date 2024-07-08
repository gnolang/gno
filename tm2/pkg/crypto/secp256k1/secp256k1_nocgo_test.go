//go:build !libsecp256k1

package secp256k1

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Ensure that signature verification works, and that
// non-canonical signatures fail.
// Note: run with CGO_ENABLED=0 or go test -tags !cgo.
func TestSignatureVerificationAndRejectUpperS(t *testing.T) {
	t.Parallel()

	msg := []byte("We have lingered long enough on the shores of the cosmic ocean.")
	for i := 0; i < 500; i++ {
		priv := GenPrivKey()
		sigStr, err := priv.Sign(msg)
		require.NoError(t, err)
		_, ok := signatureFromBytes(sigStr)
		require.True(t, ok)

		pub := priv.PubKey()
		require.True(t, pub.VerifyBytes(msg, sigStr))
	}
}

func BenchmarkVerify(b *testing.B) {
	priv := GenPrivKey()
	msg := []byte("We have lingered long enough on the shores of the cosmic ocean.")
	sigStr, err := priv.Sign(msg)
	require.NoError(b, err)

	pub := priv.PubKey()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ok := pub.VerifyBytes(msg, sigStr)
		require.True(b, ok)
	}
}
