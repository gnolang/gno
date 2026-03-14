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
	lengths := []int{0, 8, 32, 1024}

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
