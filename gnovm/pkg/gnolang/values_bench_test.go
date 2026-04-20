package gnolang

import (
	"fmt"
	"testing"

	storetypes "github.com/gnolang/gno/tm2/pkg/store/types"
)

func newBenchStoreWithGas() (Store, storetypes.GasMeter) {
	alloc := NewAllocator(1 << 62)
	gm := storetypes.NewGasMeter(1 << 62)
	ds := NewStore(alloc, nil, nil)
	tx := ds.BeginTransaction(nil, nil, gm)
	tx.GetAllocator().SetGasMeter(gm)
	return tx, gm
}

func BenchmarkComputeMapKey_String(b *testing.B) {
	st, gm := newBenchStoreWithGas()
	tv := typedString("hello")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = tv.ComputeMapKey(st, false)
	}
	b.StopTimer()

	b.ReportMetric(float64(gm.GasConsumed())/float64(b.N), "gas/op")
}

// BenchmarkComputeMapKey_StringLen varies the string length to isolate the
// per-byte slope for StringKind in ComputeMapKey (the big switch at
// values.go:StringKind case, which does `bz = append(bz, v...)` on the
// underlying bytes — O(N)).
func BenchmarkComputeMapKey_StringLen(b *testing.B) {
	lengths := []int{0, 1, 8, 32, 128, 1024, 16384, 1 << 20, 1 << 24}

	for _, n := range lengths {
		n := n
		b.Run(fmt.Sprintf("len=%d", n), func(b *testing.B) {
			st, gm := newBenchStoreWithGas()
			s := make([]byte, n)
			for i := range s {
				s[i] = byte('a' + i%26)
			}
			tv := typedString(string(s))

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = tv.ComputeMapKey(st, false)
			}
			b.StopTimer()

			b.ReportMetric(float64(gm.GasConsumed())/float64(b.N), "gas/op")
		})
	}
}

func BenchmarkComputeMapKey_Int(b *testing.B) {
	st, gm := newBenchStoreWithGas()
	tv := typedInt(123)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = tv.ComputeMapKey(st, false)
	}
	b.StopTimer()

	b.ReportMetric(float64(gm.GasConsumed())/float64(b.N), "gas/op")
}

func BenchmarkComputeMapKey_Bytes(b *testing.B) {
	lengths := []int{0, 8, 32, 128, 1024, 1 << 20, 1 << 24}

	for _, n := range lengths {
		n := n // capture
		b.Run(fmt.Sprintf("len=%d", n), func(b *testing.B) {
			st, gm := newBenchStoreWithGas()

			arrayValue := ArrayValue{Data: make([]byte, n)}
			tv := TypedValue{
				T: &ArrayType{Len: n, Elt: Uint8Type},
				V: &arrayValue,
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = tv.ComputeMapKey(st, false)
			}
			b.StopTimer()

			b.ReportMetric(float64(gm.GasConsumed())/float64(b.N), "gas/op")
		})
	}
}

func BenchmarkComputeMapKey_IntArray(b *testing.B) {
	lengths := []int{0, 1, 8, 32, 128, 1024, 8192}

	for _, n := range lengths {
		n := n
		b.Run(fmt.Sprintf("len=%d", n), func(b *testing.B) {
			st, gm := newBenchStoreWithGas()

			arr := ArrayValue{
				List: make([]TypedValue, n),
			}
			for i := 0; i < n; i++ {
				arr.List[i] = typedInt(i)
			}

			tv := TypedValue{
				T: &ArrayType{Len: n, Elt: IntType},
				V: &arr,
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = tv.ComputeMapKey(st, false)
			}
			b.StopTimer()

			b.ReportMetric(float64(gm.GasConsumed())/float64(b.N), "gas/op")
		})
	}
}

// BenchmarkComputeMapKey_Struct varies field count on a struct-of-int key.
// Calibrates the per-element slope for StructKind (same inner work as array
// element recursion — separate fit confirms the slope is shared).
func BenchmarkComputeMapKey_Struct(b *testing.B) {
	counts := []int{0, 1, 4, 16, 64, 256, 1024}

	for _, n := range counts {
		n := n
		b.Run(fmt.Sprintf("fields=%d", n), func(b *testing.B) {
			st, gm := newBenchStoreWithGas()

			fields := make([]FieldType, n)
			for i := range fields {
				fields[i] = FieldType{Name: Name(fmt.Sprintf("f%d", i)), Type: IntType}
			}
			fv := make([]TypedValue, n)
			for i := range fv {
				fv[i] = typedInt(i)
			}
			sv := &StructValue{Fields: fv}
			tv := TypedValue{
				T: &StructType{Fields: fields},
				V: sv,
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = tv.ComputeMapKey(st, false)
			}
			b.StopTimer()

			b.ReportMetric(float64(gm.GasConsumed())/float64(b.N), "gas/op")
		})
	}
}

// BenchmarkComputeMapKey_NestedArray recurses a constant-element-count
// array nested `depth` levels deep. The inner-most level has a single int.
// Each recursion level calls ComputeMapKey once — so total calls = depth+1.
// Validates that the per-call base gas scales linearly with recursion depth
// (DoS shape from advisory GHSA-m7rp-96x5-hvpx).
func BenchmarkComputeMapKey_NestedArray(b *testing.B) {
	depths := []int{1, 2, 4, 8, 16, 32}

	for _, d := range depths {
		d := d
		b.Run(fmt.Sprintf("depth=%d", d), func(b *testing.B) {
			st, gm := newBenchStoreWithGas()

			// Build nested [1][1]...[1]int of given depth.
			var tv TypedValue
			tv = typedInt(0)
			for i := 0; i < d; i++ {
				tv = TypedValue{
					T: &ArrayType{Len: 1, Elt: tv.T},
					V: &ArrayValue{List: []TypedValue{tv}},
				}
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = tv.ComputeMapKey(st, false)
			}
			b.StopTimer()

			b.ReportMetric(float64(gm.GasConsumed())/float64(b.N), "gas/op")
		})
	}
}
