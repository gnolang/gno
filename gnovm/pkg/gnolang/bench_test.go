package gnolang

import (
	"fmt"
	"strings"
	"testing"
)

var sink any = nil

var pkgIDPaths = []string{
	"encoding/json",
	"math/bits",
	"github.com/gnolang/gno/gnovm/pkg/gnolang",
	"a",
	" ",
	"",
	"github.com/gnolang/gno/gnovm/pkg/gnolang/vendor/pkg/github.com/gnolang/vendored",
}

func BenchmarkPkgIDFromPkgPath(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, path := range pkgIDPaths {
			sink = PkgIDFromPkgPath(path)
		}
	}

	if sink == nil {
		b.Fatal("Benchmark did not run!")
	}
	sink = nil
}

var benchmarkSliceSink TypedValue

func BenchmarkStringSliceAlloc(b *testing.B) {
	sizes := []int{5, 20, 40, 80, 160, 320}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Len%d", size), func(b *testing.B) {
			input := makeBenchmarkString(size)
			low := size / 4
			high := low + size/2
			if high > size {
				high = size
			}

			b.ReportAllocs()

			var totalAllocBytes int64
			for i := 0; i < b.N; i++ {
				alloc := NewAllocator(1024 * 1024)

				tv := TypedValue{
					T: StringType,
					V: alloc.NewString(input),
				}

				_, bytesBefore := alloc.Status()

				result := tv.GetSlice(alloc, low, high)
				benchmarkSliceSink = result

				_, bytesAfter := alloc.Status()
				totalAllocBytes += bytesAfter - bytesBefore
			}

			avgAllocBytes := float64(totalAllocBytes) / float64(b.N)
			b.ReportMetric(avgAllocBytes, "alloc_bytes/op")
		})
	}
}

func makeBenchmarkString(size int) string {
	const alphabet = "abcdefghijklmnopqrstuvwxyz"
	repeats := (size + len(alphabet) - 1) / len(alphabet)
	return strings.Repeat(alphabet, repeats)[:size]
}
