package gnolang

import (
	"fmt"
	"testing"

	bm "github.com/gnolang/gno/gnovm/pkg/benchops"
	storetypes "github.com/gnolang/gno/tm2/pkg/store/types"
)

// ---------------------------------------------------------------------------
// ComputeMapKey: calibrate OpCPUComputeMapKey (per-call) and
// OpCPUSlopeComputeMapKeyByte (per-byte slope for the av.Data fast path and
// the StringType primitive path).
//
// Conventions: uses bm.SwitchOpCode to isolate timing to just the
// ComputeMapKey call, and reports ns/op(pure) alongside gas/op so the linear
// fit gives the per-byte slope in ns/byte. See bench_ops_test.go.
// ---------------------------------------------------------------------------

func newBenchStoreWithGas() (Store, storetypes.GasMeter) {
	alloc := NewAllocator(1 << 62)
	gm := storetypes.NewGasMeter(1 << 62)
	ds := NewStore(alloc, nil, nil)
	tx := ds.BeginTransaction(nil, nil, nil, gm)
	tx.GetAllocator().SetGasMeter(gm)
	return tx, gm
}

// benchMachineWithGas wires a gas meter into a benchMachine so
// ComputeMapKey's m.GasMeter.ConsumeGas path is exercised.
func benchMachineWithGas() (*Machine, storetypes.GasMeter) {
	m := benchMachine()
	gm := storetypes.NewGasMeter(1 << 62)
	m.GasMeter = gm
	return m, gm
}

func benchComputeMapKey(b *testing.B, tv TypedValue) {
	b.Helper()
	m, gm := benchMachineWithGas()
	defer m.Release()

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		bm.SwitchOpCode(bmTarget)
		_, _ = tv.ComputeMapKey(m, m.Store, false)
		bm.SwitchOpCode(bmSetup)
	}
	reportBenchops(b)
	b.ReportMetric(float64(gm.GasConsumed())/float64(b.N), "gas/op")
}

func BenchmarkComputeMapKey_String(b *testing.B) {
	benchComputeMapKey(b, typedString("hello"))
}

func BenchmarkComputeMapKey_Int(b *testing.B) {
	benchComputeMapKey(b, typedInt(123))
}

// Bytes: ArrayType with av.Data != nil — single ComputeMapKey call whose
// cost scales with len(av.Data). Linear fit against N gives ns/byte.
func BenchmarkComputeMapKey_Bytes(b *testing.B) {
	for _, n := range []int{0, 8, 32, 128, 1024, 1 << 20, 1 << 24} {
		b.Run(fmt.Sprintf("len=%d", n), func(b *testing.B) {
			tv := TypedValue{
				T: &ArrayType{Len: n, Elt: Uint8Type},
				V: &ArrayValue{Data: make([]byte, n)},
			}
			benchComputeMapKey(b, tv)
		})
	}
}

// LongString: StringType — same per-byte slope concern as Bytes, exercised
// through MapKeyBytes' append(bz, s...).
func BenchmarkComputeMapKey_LongString(b *testing.B) {
	for _, n := range []int{0, 8, 32, 128, 1024, 1 << 20} {
		b.Run(fmt.Sprintf("len=%d", n), func(b *testing.B) {
			s := string(make([]byte, n))
			tv := typedString(s)
			benchComputeMapKey(b, tv)
		})
	}
}

// IntArray: ArrayType with av.Data == nil — N+1 recursive ComputeMapKey
// calls. The per-call slope is (total ns) / (N+1).
func BenchmarkComputeMapKey_IntArray(b *testing.B) {
	for _, n := range []int{0, 8, 32, 1024} {
		b.Run(fmt.Sprintf("len=%d", n), func(b *testing.B) {
			arr := ArrayValue{List: make([]TypedValue, n)}
			for i := 0; i < n; i++ {
				arr.List[i] = typedInt(i)
			}
			tv := TypedValue{
				T: &ArrayType{Len: n, Elt: IntType},
				V: &arr,
			}
			benchComputeMapKey(b, tv)
		})
	}
}
