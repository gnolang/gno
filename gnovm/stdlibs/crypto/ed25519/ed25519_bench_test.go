package ed25519

import (
	"crypto/ed25519"
	"fmt"
	"testing"
)

func BenchmarkEd25519Verify(b *testing.B) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		b.Fatal(err)
	}
	sizes := []int{64, 1024, 65536, 1048576}
	for _, size := range sizes {
		msg := make([]byte, size)
		sig := ed25519.Sign(priv, msg)
		b.Run(fmt.Sprintf("msgsize=%d", size), func(b *testing.B) {
			b.SetBytes(int64(size))
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				ed25519.Verify(pub, msg, sig)
			}
		})
	}
}
