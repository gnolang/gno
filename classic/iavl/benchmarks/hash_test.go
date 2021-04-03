package benchmarks

import (
	"crypto"
	"fmt"
	"hash"
	"testing"

	_ "crypto/sha256"

	_ "golang.org/x/crypto/ripemd160"
	_ "golang.org/x/crypto/sha3"
)

func BenchmarkHash(b *testing.B) {
	hashers := []struct {
		name   string
		size   int
		hasher hash.Hash
	}{
		{"ripemd160", 64, crypto.RIPEMD160.New()},
		{"ripemd160", 512, crypto.RIPEMD160.New()},
		{"sha2-256", 64, crypto.SHA256.New()},
		{"sha2-256", 512, crypto.SHA256.New()},
		{"sha3-256", 64, crypto.SHA3_256.New()},
		{"sha3-256", 512, crypto.SHA3_256.New()},
	}

	for _, h := range hashers {
		prefix := fmt.Sprintf("%s-%d", h.name, h.size)
		b.Run(prefix, func(sub *testing.B) {
			benchHasher(sub, h.hasher, h.size)
		})
	}
}

func benchHasher(b *testing.B, hasher hash.Hash, size int) {
	// create all random bytes before to avoid timing this
	inputs := randBytes(b.N + size + 1)

	for i := 0; i < b.N; i++ {
		hasher.Reset()
		// grab a slice of size bytes from random string
		hasher.Write(inputs[i : i+size])
		hasher.Sum(nil)
	}
}
