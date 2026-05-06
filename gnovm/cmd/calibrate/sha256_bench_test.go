package calibrate

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

func BenchmarkSHA256(b *testing.B) {
	for _, size := range []int{32, 64, 128, 256} {
		data := make([]byte, size)
		for i := range data {
			data[i] = byte(i)
		}
		b.Run(fmt.Sprintf("bytes=%d", size), func(b *testing.B) {
			b.SetBytes(int64(size))
			for i := 0; i < b.N; i++ {
				sha256.Sum256(data)
			}
		})
	}
}
