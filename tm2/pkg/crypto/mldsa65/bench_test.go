package mldsa65

import (
	"io"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/crypto/internal/benchmarking"
	"github.com/gnolang/gno/tm2/pkg/crypto/secp256k1"
)

func BenchmarkKeyGeneration(b *testing.B) {
	benchmarkKeygenWrapper := func(reader io.Reader) crypto.PrivKey {
		return genPrivKey(reader)
	}
	benchmarking.BenchmarkKeyGeneration(b, benchmarkKeygenWrapper)
}

func BenchmarkSigning(b *testing.B) {
	priv := GenPrivKey()
	benchmarking.BenchmarkSigning(b, priv)
}

func BenchmarkVerification(b *testing.B) {
	priv := GenPrivKey()
	benchmarking.BenchmarkVerification(b, priv)
}

// BenchmarkVerifyBytes compares VerifyBytes across the three account key
// types on the same 128-byte message. It is the measurement behind
// auth.DefaultSigVerifyCostMLDSA65.
func BenchmarkVerifyBytes(b *testing.B) {
	msg := make([]byte, 128)
	for i := range msg {
		msg[i] = byte(i)
	}
	for _, c := range []struct {
		name string
		priv crypto.PrivKey
	}{
		{"ed25519", ed25519.GenPrivKeyFromSecret([]byte("bench"))},
		{"secp256k1", secp256k1.GenPrivKeySecp256k1([]byte("bench"))},
		{"mldsa65", GenPrivKeyFromSecret([]byte("bench"))},
	} {
		pub := c.priv.PubKey()
		sig, err := c.priv.Sign(msg)
		if err != nil {
			b.Fatal(err)
		}
		b.Run(c.name, func(b *testing.B) {
			for b.Loop() {
				if !pub.VerifyBytes(msg, sig) {
					b.Fatal("signature did not verify")
				}
			}
		})
	}
}
