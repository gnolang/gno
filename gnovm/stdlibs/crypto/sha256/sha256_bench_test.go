package sha256

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

func BenchmarkSha256Sum256(b *testing.B) {
	sizes := []int{64, 1024, 65536, 1048576}
	for _, size := range sizes {
		data := make([]byte, size)
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			b.SetBytes(int64(size))
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				sha256.Sum256(data)
			}
		})
	}
}
