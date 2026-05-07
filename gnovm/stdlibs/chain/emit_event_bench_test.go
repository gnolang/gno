package chain

import (
	"fmt"
	"strings"
	"testing"
)

func BenchmarkAttrKeysAndValues(b *testing.B) {
	counts := []int{2, 10, 50, 200}
	for _, numPairs := range counts {
		attrs := make([]string, numPairs*2)
		for i := range attrs {
			attrs[i] = strings.Repeat("x", 100)
		}
		totalBytes := 0
		for _, a := range attrs {
			totalBytes += len(a)
		}
		b.Run(fmt.Sprintf("pairs=%d/bytes=%d", numPairs, totalBytes), func(b *testing.B) {
			b.SetBytes(int64(totalBytes))
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				attrKeysAndValues(attrs)
			}
		})
	}
}
