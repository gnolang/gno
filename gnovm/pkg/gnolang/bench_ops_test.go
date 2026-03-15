package gnolang

import (
	"fmt"
	"io"
	"math"
	"math/big"
	"strings"
	"testing"

	"github.com/cockroachdb/apd/v3"

	bm "github.com/gnolang/gno/gnovm/pkg/benchops"
	"github.com/gnolang/gno/tm2/pkg/store"
)

// bench_ops_test.go: Go-level microbenchmarks for GnoVM op handlers.
//
// Each benchmark sets up the Machine stack, uses bm.SwitchOpCode to
// isolate timing to just the doOpXxx() call, and checks the result.

const (
	bmSetup  = byte(0x01) // dummy op code for setup phases
	bmTarget = byte(0x02) // op code for the measured operation
)

// benchAllocMeter tracks allocation gas across the benchmark.
// Set by benchMachine(), read by reportBenchops().
var benchAllocMeter store.GasMeter

func benchMachine() *Machine {
	benchAllocMeter = store.NewGasMeter(math.MaxInt64)
	alloc := NewAllocator(math.MaxInt64)
	m := NewMachineWithOptions(MachineOptions{
		PkgPath: "bench",
		Output:  io.Discard,
		Alloc:   alloc,
	})
	// Set alloc gas meter AFTER construction, so it doesn't get
	// overwritten by NewMachineWithOptions (which sets alloc.gasMeter
	// to opts.GasMeter). We intentionally don't set opts.GasMeter
	// to avoid enabling CPU gas metering via incrCPU.
	alloc.SetGasMeter(benchAllocMeter)
	return m
}

func reportBenchops(b *testing.B) {
	b.Helper()
	bm.StopOpCode()
	count := bm.OpCount(bmTarget)
	if count > 0 {
		avgNs := float64(bm.OpAccumDur(bmTarget).Nanoseconds()) / float64(count)
		b.ReportMetric(avgNs, "ns/op(pure)")
	}
	if benchAllocMeter != nil && b.N > 0 {
		// One-time setup allocs are amortized to ~0 across b.N iterations.
		allocGas := float64(benchAllocMeter.GasConsumed()) / float64(b.N)
		b.ReportMetric(allocGas, "alloc-gas/op")
	}
}

// ---------------------------------------------------------------------------
// doOpAdd: PopExpr, PopValue(rv), PeekValue(lv); lv = lv + rv
// addAssign switches on type: IntType does lv.SetInt(lv.GetInt()+rv.GetInt())
// StringType allocates via alloc.NewString(lv+rv)
// ---------------------------------------------------------------------------

func BenchmarkOpAdd_Int(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: IntType, N: i2n(42)})
		m.PushValue(TypedValue{T: IntType, N: i2n(8)})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpAdd()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.GetInt() != 50 {
			b.Fatalf("expected 50, got %d", res.GetInt())
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func benchOpAdd_String(b *testing.B, length int) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}
	s1 := strings.Repeat("a", length)
	s2 := strings.Repeat("b", length)
	sv1 := m.Alloc.NewString(s1)
	sv2 := m.Alloc.NewString(s2)

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: StringType, V: sv1})
		m.PushValue(TypedValue{T: StringType, V: sv2})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpAdd()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if len(res.GetString()) != 2*length {
			b.Fatalf("expected len %d, got %d", 2*length, len(res.GetString()))
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpAdd_String_10(b *testing.B)    { benchOpAdd_String(b, 10) }
func BenchmarkOpAdd_String_100(b *testing.B)   { benchOpAdd_String(b, 100) }
func BenchmarkOpAdd_String_1000(b *testing.B)  { benchOpAdd_String(b, 1000) }
func BenchmarkOpAdd_String_10000(b *testing.B) { benchOpAdd_String(b, 10000) }

func BenchmarkOpAdd_Float64(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}
	// Store float64 as uint64 bit pattern in N.
	v1 := math.Float64bits(3.14159)
	v2 := math.Float64bits(2.71828)

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: Float64Type, N: u2n(v1)})
		m.PushValue(TypedValue{T: Float64Type, N: u2n(v2)})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpAdd()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		got := math.Float64frombits(res.GetFloat64())
		if got < 5.8 || got > 5.9 {
			b.Fatalf("expected ~5.859, got %f", got)
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func benchOpAdd_BigInt(b *testing.B, bits int) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}
	v1 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(bits)), big.NewInt(1))
	v2 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(bits)), big.NewInt(1))
	bv1 := BigintValue{V: v1}
	bv2 := BigintValue{V: v2}
	expected := new(big.Int).Add(v1, v2)

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: UntypedBigintType, V: bv1})
		m.PushValue(TypedValue{T: UntypedBigintType, V: bv2})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpAdd()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.V.(BigintValue).V.Cmp(expected) != 0 {
			b.Fatal("unexpected result")
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpAdd_BigInt_64(b *testing.B)   { benchOpAdd_BigInt(b, 64) }
func BenchmarkOpAdd_BigInt_256(b *testing.B)  { benchOpAdd_BigInt(b, 256) }
func BenchmarkOpAdd_BigInt_1024(b *testing.B) { benchOpAdd_BigInt(b, 1024) }
func BenchmarkOpAdd_BigInt_4096(b *testing.B) { benchOpAdd_BigInt(b, 4096) }

func benchOpAdd_BigInt_Asym(b *testing.B, bitsA, bitsB int) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}
	v1 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(bitsA)), big.NewInt(1))
	v2 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(bitsB)), big.NewInt(1))
	bv1 := BigintValue{V: v1}
	bv2 := BigintValue{V: v2}
	expected := new(big.Int).Add(v1, v2)

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: UntypedBigintType, V: bv1})
		m.PushValue(TypedValue{T: UntypedBigintType, V: bv2})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpAdd()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.V.(BigintValue).V.Cmp(expected) != 0 {
			b.Fatal("unexpected result")
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpAdd_BigInt_64x4096(b *testing.B)  { benchOpAdd_BigInt_Asym(b, 64, 4096) }
func BenchmarkOpAdd_BigInt_256x4096(b *testing.B) { benchOpAdd_BigInt_Asym(b, 256, 4096) }

// ---------------------------------------------------------------------------
// doOpSub: PopExpr, PopValue(rv), PeekValue(lv); lv = lv - rv
// ---------------------------------------------------------------------------

func BenchmarkOpSub_Int(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: IntType, N: i2n(42)})
		m.PushValue(TypedValue{T: IntType, N: i2n(8)})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpSub()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.GetInt() != 34 {
			b.Fatalf("expected 34, got %d", res.GetInt())
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpSub_Float64(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}
	v1 := math.Float64bits(3.14159)
	v2 := math.Float64bits(2.71828)

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: Float64Type, N: u2n(v1)})
		m.PushValue(TypedValue{T: Float64Type, N: u2n(v2)})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpSub()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		got := math.Float64frombits(res.GetFloat64())
		if got < 0.4 || got > 0.5 {
			b.Fatalf("expected ~0.423, got %f", got)
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func benchOpSub_BigInt(b *testing.B, bits int) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}
	v1 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(bits)), big.NewInt(1))
	v2 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(bits)), big.NewInt(1))
	bv1 := BigintValue{V: v1}
	bv2 := BigintValue{V: v2}
	expected := new(big.Int).Sub(v1, v2)

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: UntypedBigintType, V: bv1})
		m.PushValue(TypedValue{T: UntypedBigintType, V: bv2})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpSub()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.V.(BigintValue).V.Cmp(expected) != 0 {
			b.Fatal("unexpected result")
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpSub_BigInt_64(b *testing.B)   { benchOpSub_BigInt(b, 64) }
func BenchmarkOpSub_BigInt_256(b *testing.B)  { benchOpSub_BigInt(b, 256) }
func BenchmarkOpSub_BigInt_1024(b *testing.B) { benchOpSub_BigInt(b, 1024) }
func BenchmarkOpSub_BigInt_4096(b *testing.B) { benchOpSub_BigInt(b, 4096) }

// ---------------------------------------------------------------------------
// doOpMul: PopExpr, PopValue(rv), PeekValue(lv); lv = lv * rv
// ---------------------------------------------------------------------------

func BenchmarkOpMul_Int(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: IntType, N: i2n(6)})
		m.PushValue(TypedValue{T: IntType, N: i2n(7)})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpMul()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.GetInt() != 42 {
			b.Fatalf("expected 42, got %d", res.GetInt())
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpMul_Float64(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}
	v1 := math.Float64bits(3.14159)
	v2 := math.Float64bits(2.71828)

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: Float64Type, N: u2n(v1)})
		m.PushValue(TypedValue{T: Float64Type, N: u2n(v2)})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpMul()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		got := math.Float64frombits(res.GetFloat64())
		if got < 8.5 || got > 8.6 {
			b.Fatalf("expected ~8.539, got %f", got)
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func benchOpMul_BigInt(b *testing.B, bits int) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}
	v1 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(bits)), big.NewInt(1))
	v2 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(bits)), big.NewInt(1))
	bv1 := BigintValue{V: v1}
	bv2 := BigintValue{V: v2}
	expected := new(big.Int).Mul(v1, v2)

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: UntypedBigintType, V: bv1})
		m.PushValue(TypedValue{T: UntypedBigintType, V: bv2})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpMul()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.V.(BigintValue).V.Cmp(expected) != 0 {
			b.Fatal("unexpected result")
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpMul_BigInt_64(b *testing.B)   { benchOpMul_BigInt(b, 64) }
func BenchmarkOpMul_BigInt_256(b *testing.B)  { benchOpMul_BigInt(b, 256) }
func BenchmarkOpMul_BigInt_1024(b *testing.B) { benchOpMul_BigInt(b, 1024) }
func BenchmarkOpMul_BigInt_4096(b *testing.B) { benchOpMul_BigInt(b, 4096) }

func benchOpMul_BigInt_Asym(b *testing.B, bitsA, bitsB int) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}
	v1 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(bitsA)), big.NewInt(1))
	v2 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(bitsB)), big.NewInt(1))
	bv1 := BigintValue{V: v1}
	bv2 := BigintValue{V: v2}
	expected := new(big.Int).Mul(v1, v2)

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: UntypedBigintType, V: bv1})
		m.PushValue(TypedValue{T: UntypedBigintType, V: bv2})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpMul()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.V.(BigintValue).V.Cmp(expected) != 0 {
			b.Fatal("unexpected result")
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpMul_BigInt_64x4096(b *testing.B)  { benchOpMul_BigInt_Asym(b, 64, 4096) }
func BenchmarkOpMul_BigInt_256x4096(b *testing.B) { benchOpMul_BigInt_Asym(b, 256, 4096) }
func BenchmarkOpMul_BigInt_64x1024(b *testing.B)  { benchOpMul_BigInt_Asym(b, 64, 1024) }
func BenchmarkOpMul_BigInt_256x1024(b *testing.B) { benchOpMul_BigInt_Asym(b, 256, 1024) }

// ---------------------------------------------------------------------------
// doOpQuo: PopExpr, PopValue(rv), PeekValue(lv); lv = lv / rv
// ---------------------------------------------------------------------------

func BenchmarkOpQuo_Int(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: IntType, N: i2n(42)})
		m.PushValue(TypedValue{T: IntType, N: i2n(6)})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpQuo()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.GetInt() != 7 {
			b.Fatalf("expected 7, got %d", res.GetInt())
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpQuo_Float64(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}
	v1 := math.Float64bits(10.0)
	v2 := math.Float64bits(3.0)

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: Float64Type, N: u2n(v1)})
		m.PushValue(TypedValue{T: Float64Type, N: u2n(v2)})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpQuo()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		got := math.Float64frombits(res.GetFloat64())
		if got < 3.3 || got > 3.4 {
			b.Fatalf("expected ~3.333, got %f", got)
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func benchOpQuo_BigInt(b *testing.B, bits int) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}
	// dividend: bits-wide; divisor: bits/2-wide (min 32)
	v1 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(bits)), big.NewInt(1))
	divisorBits := bits / 2
	if divisorBits < 32 {
		divisorBits = 32
	}
	v2 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(divisorBits)), big.NewInt(1))
	bv1 := BigintValue{V: v1}
	bv2 := BigintValue{V: v2}
	expected := new(big.Int).Quo(v1, v2)

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: UntypedBigintType, V: bv1})
		m.PushValue(TypedValue{T: UntypedBigintType, V: bv2})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpQuo()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.V.(BigintValue).V.Cmp(expected) != 0 {
			b.Fatal("unexpected result")
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpQuo_BigInt_64(b *testing.B)   { benchOpQuo_BigInt(b, 64) }
func BenchmarkOpQuo_BigInt_256(b *testing.B)  { benchOpQuo_BigInt(b, 256) }
func BenchmarkOpQuo_BigInt_1024(b *testing.B) { benchOpQuo_BigInt(b, 1024) }
func BenchmarkOpQuo_BigInt_4096(b *testing.B) { benchOpQuo_BigInt(b, 4096) }

func benchOpQuo_BigInt_Asym(b *testing.B, bitsA, bitsB int) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}
	// bitsA = dividend, bitsB = divisor
	v1 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(bitsA)), big.NewInt(1))
	v2 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(bitsB)), big.NewInt(1))
	bv1 := BigintValue{V: v1}
	bv2 := BigintValue{V: v2}
	expected := new(big.Int).Quo(v1, v2)

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: UntypedBigintType, V: bv1})
		m.PushValue(TypedValue{T: UntypedBigintType, V: bv2})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpQuo()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.V.(BigintValue).V.Cmp(expected) != 0 {
			b.Fatal("unexpected result")
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpQuo_BigInt_4096x64(b *testing.B)  { benchOpQuo_BigInt_Asym(b, 4096, 64) }
func BenchmarkOpQuo_BigInt_4096x256(b *testing.B) { benchOpQuo_BigInt_Asym(b, 4096, 256) }
func BenchmarkOpQuo_BigInt_1024x64(b *testing.B)  { benchOpQuo_BigInt_Asym(b, 1024, 64) }

// ---------------------------------------------------------------------------
// doOpRem: PopExpr, PopValue(rv), PeekValue(lv); lv = lv % rv
// ---------------------------------------------------------------------------

func BenchmarkOpRem_Int(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: IntType, N: i2n(42)})
		m.PushValue(TypedValue{T: IntType, N: i2n(5)})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpRem()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.GetInt() != 2 {
			b.Fatalf("expected 2, got %d", res.GetInt())
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func benchOpRem_BigInt(b *testing.B, bits int) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}
	// dividend: bits-wide; divisor: bits/2-wide (min 32)
	v1 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(bits)), big.NewInt(1))
	divisorBits := bits / 2
	if divisorBits < 32 {
		divisorBits = 32
	}
	v2 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(divisorBits)), big.NewInt(1))
	bv1 := BigintValue{V: v1}
	bv2 := BigintValue{V: v2}
	expected := new(big.Int).Rem(v1, v2)

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: UntypedBigintType, V: bv1})
		m.PushValue(TypedValue{T: UntypedBigintType, V: bv2})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpRem()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.V.(BigintValue).V.Cmp(expected) != 0 {
			b.Fatal("unexpected result")
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpRem_BigInt_64(b *testing.B)   { benchOpRem_BigInt(b, 64) }
func BenchmarkOpRem_BigInt_256(b *testing.B)  { benchOpRem_BigInt(b, 256) }
func BenchmarkOpRem_BigInt_1024(b *testing.B) { benchOpRem_BigInt(b, 1024) }
func BenchmarkOpRem_BigInt_4096(b *testing.B) { benchOpRem_BigInt(b, 4096) }

func benchOpRem_BigInt_Asym(b *testing.B, bitsA, bitsB int) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}
	// bitsA = dividend, bitsB = divisor
	v1 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(bitsA)), big.NewInt(1))
	v2 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(bitsB)), big.NewInt(1))
	bv1 := BigintValue{V: v1}
	bv2 := BigintValue{V: v2}
	expected := new(big.Int).Rem(v1, v2)

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: UntypedBigintType, V: bv1})
		m.PushValue(TypedValue{T: UntypedBigintType, V: bv2})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpRem()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.V.(BigintValue).V.Cmp(expected) != 0 {
			b.Fatal("unexpected result")
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpRem_BigInt_4096x64(b *testing.B)  { benchOpRem_BigInt_Asym(b, 4096, 64) }
func BenchmarkOpRem_BigInt_4096x256(b *testing.B) { benchOpRem_BigInt_Asym(b, 4096, 256) }

// ---------------------------------------------------------------------------
// doOpBand: PopExpr, PopValue(rv), PeekValue(lv); lv = lv & rv
// ---------------------------------------------------------------------------

func BenchmarkOpBand(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: IntType, N: i2n(0xFF)})
		m.PushValue(TypedValue{T: IntType, N: i2n(0x0F)})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpBand()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.GetInt() != 0x0F {
			b.Fatalf("expected 0x0F, got %d", res.GetInt())
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

// ---------------------------------------------------------------------------
// doOpBor: PopExpr, PopValue(rv), PeekValue(lv); lv = lv | rv
// ---------------------------------------------------------------------------

func BenchmarkOpBor(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: IntType, N: i2n(0xF0)})
		m.PushValue(TypedValue{T: IntType, N: i2n(0x0F)})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpBor()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.GetInt() != 0xFF {
			b.Fatalf("expected 0xFF, got %d", res.GetInt())
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

// ---------------------------------------------------------------------------
// doOpXor: PopExpr, PopValue(rv), PeekValue(lv); lv = lv ^ rv
// ---------------------------------------------------------------------------

func BenchmarkOpXor(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: IntType, N: i2n(0xFF)})
		m.PushValue(TypedValue{T: IntType, N: i2n(0x0F)})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpXor()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.GetInt() != 0xF0 {
			b.Fatalf("expected 0xF0, got %d", res.GetInt())
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

// ---------------------------------------------------------------------------
// doOpShl: PopExpr, PopValue(rv), PeekValue(lv); lv = lv << rv
// rv must be unsigned type.
// ---------------------------------------------------------------------------

func BenchmarkOpShl(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: IntType, N: i2n(1)})
		m.PushValue(TypedValue{T: UintType, N: u2n(10)})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpShl()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.GetInt() != 1024 {
			b.Fatalf("expected 1024, got %d", res.GetInt())
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

// ---------------------------------------------------------------------------
// doOpShr: PopExpr, PopValue(rv), PeekValue(lv); lv = lv >> rv
// rv must be unsigned type.
// ---------------------------------------------------------------------------

func BenchmarkOpShr(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: IntType, N: i2n(1024)})
		m.PushValue(TypedValue{T: UintType, N: u2n(10)})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpShr()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.GetInt() != 1 {
			b.Fatalf("expected 1, got %d", res.GetInt())
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

// ---------------------------------------------------------------------------
// doOpEql: PopExpr, PopValue(rv), PeekValue(lv); lv = (lv == rv)
// Result type is UntypedBoolType. isEql dispatches on type.
// ---------------------------------------------------------------------------

func BenchmarkOpEql_Int(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: IntType, N: i2n(42)})
		m.PushValue(TypedValue{T: IntType, N: i2n(42)})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpEql()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if !res.GetBool() {
			b.Fatal("expected true")
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func benchOpEql_String(b *testing.B, length int) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}
	s := strings.Repeat("x", length)
	sv := m.Alloc.NewString(s)

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: StringType, V: sv})
		m.PushValue(TypedValue{T: StringType, V: sv})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpEql()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if !res.GetBool() {
			b.Fatal("expected true")
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpEql_String_10(b *testing.B)    { benchOpEql_String(b, 10) }
func BenchmarkOpEql_String_100(b *testing.B)   { benchOpEql_String(b, 100) }
func BenchmarkOpEql_String_1000(b *testing.B)  { benchOpEql_String(b, 1000) }
func BenchmarkOpEql_String_10000(b *testing.B) { benchOpEql_String(b, 10000) }

// ---------------------------------------------------------------------------
// doOpNeq: PopExpr, PopValue(rv), PeekValue(lv); lv = (lv != rv)
// ---------------------------------------------------------------------------

func BenchmarkOpNeq(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: IntType, N: i2n(42)})
		m.PushValue(TypedValue{T: IntType, N: i2n(43)})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpNeq()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if !res.GetBool() {
			b.Fatal("expected true")
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

// ---------------------------------------------------------------------------
// doOpLss: PopExpr, PopValue(rv), PeekValue(lv); lv = (lv < rv)
// ---------------------------------------------------------------------------

func BenchmarkOpLss(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: IntType, N: i2n(5)})
		m.PushValue(TypedValue{T: IntType, N: i2n(10)})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpLss()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if !res.GetBool() {
			b.Fatal("expected true")
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func benchOpLss_String(b *testing.B, length int) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}
	s1 := strings.Repeat("a", length)
	s2 := strings.Repeat("a", length-1) + "b"
	sv1 := m.Alloc.NewString(s1)
	sv2 := m.Alloc.NewString(s2)

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: StringType, V: sv1})
		m.PushValue(TypedValue{T: StringType, V: sv2})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpLss()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if !res.GetBool() {
			b.Fatal("expected true")
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpLss_String_10(b *testing.B)    { benchOpLss_String(b, 10) }
func BenchmarkOpLss_String_100(b *testing.B)   { benchOpLss_String(b, 100) }
func BenchmarkOpLss_String_1000(b *testing.B)  { benchOpLss_String(b, 1000) }
func BenchmarkOpLss_String_10000(b *testing.B) { benchOpLss_String(b, 10000) }

// ---------------------------------------------------------------------------
// doOpLeq: PopExpr, PopValue(rv), PeekValue(lv); lv = (lv <= rv)
// ---------------------------------------------------------------------------

func BenchmarkOpLeq(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: IntType, N: i2n(10)})
		m.PushValue(TypedValue{T: IntType, N: i2n(10)})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpLeq()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if !res.GetBool() {
			b.Fatal("expected true")
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

// ---------------------------------------------------------------------------
// doOpGtr: PopExpr, PopValue(rv), PeekValue(lv); lv = (lv > rv)
// ---------------------------------------------------------------------------

func BenchmarkOpGtr(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: IntType, N: i2n(10)})
		m.PushValue(TypedValue{T: IntType, N: i2n(5)})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpGtr()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if !res.GetBool() {
			b.Fatal("expected true")
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

// ---------------------------------------------------------------------------
// doOpGeq: PopExpr, PopValue(rv), PeekValue(lv); lv = (lv >= rv)
// ---------------------------------------------------------------------------

func BenchmarkOpGeq(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: IntType, N: i2n(10)})
		m.PushValue(TypedValue{T: IntType, N: i2n(10)})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpGeq()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if !res.GetBool() {
			b.Fatal("expected true")
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

// ---------------------------------------------------------------------------
// doOpLor: PopValue(rv), PeekValue(lv); lv = lv || rv
// No PopExpr — called after doOpBinary1 evaluates the right side.
// ---------------------------------------------------------------------------

func BenchmarkOpLor(b *testing.B) {
	m := benchMachine()
	defer m.Release()

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: BoolType, N: i2n(0)}) // false
		m.PushValue(TypedValue{T: BoolType, N: i2n(1)}) // true
		bm.SwitchOpCode(bmTarget)
		m.doOpLor()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if !res.GetBool() {
			b.Fatal("expected true")
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

// ---------------------------------------------------------------------------
// doOpLand: PopValue(rv), PeekValue(lv); lv = lv && rv
// ---------------------------------------------------------------------------

func BenchmarkOpLand(b *testing.B) {
	m := benchMachine()
	defer m.Release()

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: BoolType, N: i2n(1)}) // true
		m.PushValue(TypedValue{T: BoolType, N: i2n(1)}) // true
		bm.SwitchOpCode(bmTarget)
		m.doOpLand()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if !res.GetBool() {
			b.Fatal("expected true")
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

// ---------------------------------------------------------------------------
// doOpUneg: PopExpr(UnaryExpr), PeekValue(1); xv = -xv
// ---------------------------------------------------------------------------

func BenchmarkOpUneg(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	expr := &UnaryExpr{}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: IntType, N: i2n(42)})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpUneg()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.GetInt() != -42 {
			b.Fatalf("expected -42, got %d", res.GetInt())
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func benchOpUneg_BigInt(b *testing.B, bits int) {
	m := benchMachine()
	defer m.Release()
	expr := &UnaryExpr{}
	v := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(bits)), big.NewInt(1))
	bv := BigintValue{V: v}
	expected := new(big.Int).Neg(v)

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: UntypedBigintType, V: bv})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpUneg()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.V.(BigintValue).V.Cmp(expected) != 0 {
			b.Fatal("unexpected result")
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpUneg_BigInt_64(b *testing.B)   { benchOpUneg_BigInt(b, 64) }
func BenchmarkOpUneg_BigInt_256(b *testing.B)  { benchOpUneg_BigInt(b, 256) }
func BenchmarkOpUneg_BigInt_1024(b *testing.B) { benchOpUneg_BigInt(b, 1024) }
func BenchmarkOpUneg_BigInt_4096(b *testing.B) { benchOpUneg_BigInt(b, 4096) }

// ---------------------------------------------------------------------------
// doOpUnot: PopExpr(UnaryExpr), PeekValue(1); xv = !xv
// ---------------------------------------------------------------------------

func BenchmarkOpUnot(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	expr := &UnaryExpr{}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: BoolType, N: i2n(1)}) // true
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpUnot()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.GetBool() {
			b.Fatal("expected false")
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

// ---------------------------------------------------------------------------
// doOpUpos: PopExpr(UnaryExpr); no-op (+x == x)
// ---------------------------------------------------------------------------

func BenchmarkOpUpos(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	expr := &UnaryExpr{}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: IntType, N: i2n(42)})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpUpos()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.GetInt() != 42 {
			b.Fatalf("expected 42, got %d", res.GetInt())
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

// ---------------------------------------------------------------------------
// doOpUxor: PopExpr(UnaryExpr), PeekValue(1); xv = ^xv
// ---------------------------------------------------------------------------

func BenchmarkOpUxor(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	expr := &UnaryExpr{}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: IntType, N: i2n(0x0F)})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpUxor()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.GetInt() != ^int64(0x0F) {
			b.Fatalf("expected %d, got %d", ^int64(0x0F), res.GetInt())
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func benchOpUxor_BigInt(b *testing.B, bits int) {
	m := benchMachine()
	defer m.Release()
	expr := &UnaryExpr{}
	v := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(bits)), big.NewInt(1))
	bv := BigintValue{V: v}
	expected := new(big.Int).Not(v)

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: UntypedBigintType, V: bv})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpUxor()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.V.(BigintValue).V.Cmp(expected) != 0 {
			b.Fatal("unexpected result")
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpUxor_BigInt_64(b *testing.B)   { benchOpUxor_BigInt(b, 64) }
func BenchmarkOpUxor_BigInt_256(b *testing.B)  { benchOpUxor_BigInt(b, 256) }
func BenchmarkOpUxor_BigInt_1024(b *testing.B) { benchOpUxor_BigInt(b, 1024) }
func BenchmarkOpUxor_BigInt_4096(b *testing.B) { benchOpUxor_BigInt(b, 4096) }

// ---------------------------------------------------------------------------
// doOpSliceLit: PopExpr(CompositeLitExpr), pop N values, pop type, push slice.
// Parameterized by element count.
// ---------------------------------------------------------------------------

func benchOpSliceLit(b *testing.B, n int) {
	m := benchMachine()
	defer m.Release()

	st := m.Alloc.NewType(&SliceType{Elt: IntType, Vrd: false})
	elts := make([]KeyValueExpr, n)
	for i := range n {
		elts[i] = KeyValueExpr{Value: &ConstExpr{}}
	}
	litExpr := &CompositeLitExpr{Elts: elts}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(asValue(st))
		for i := range n {
			m.PushValue(TypedValue{T: IntType, N: i2n(int64(i))})
		}
		m.PushExpr(litExpr)
		bm.SwitchOpCode(bmTarget)
		m.doOpSliceLit()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		sv := res.V.(*SliceValue)
		if sv.Length != n {
			b.Fatalf("expected len %d, got %d", n, sv.Length)
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpSliceLit_1(b *testing.B)   { benchOpSliceLit(b, 1) }
func BenchmarkOpSliceLit_10(b *testing.B)  { benchOpSliceLit(b, 10) }
func BenchmarkOpSliceLit_100(b *testing.B)  { benchOpSliceLit(b, 100) }
func BenchmarkOpSliceLit_1000(b *testing.B) { benchOpSliceLit(b, 1000) }

// ---------------------------------------------------------------------------
// doOpArrayLit: PopExpr(CompositeLitExpr), PopValues(N), peek type at bottom, push array.
// Parameterized by element count.
// ---------------------------------------------------------------------------

func benchOpArrayLit(b *testing.B, n int) {
	m := benchMachine()
	defer m.Release()

	at := m.Alloc.NewType(&ArrayType{Elt: IntType, Len: n})
	elts := make([]KeyValueExpr, n)
	for i := range n {
		elts[i] = KeyValueExpr{Value: &ConstExpr{}}
	}
	litExpr := &CompositeLitExpr{Elts: elts}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(asValue(at))
		for i := range n {
			m.PushValue(TypedValue{T: IntType, N: i2n(int64(i))})
		}
		m.PushExpr(litExpr)
		bm.SwitchOpCode(bmTarget)
		m.doOpArrayLit()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		av := res.V.(*ArrayValue)
		if len(av.List) != n {
			b.Fatalf("expected len %d, got %d", n, len(av.List))
		}
		if av.List[0].GetInt() != 0 {
			b.Fatalf("expected first element 0, got %d", av.List[0].GetInt())
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpArrayLit_1(b *testing.B)   { benchOpArrayLit(b, 1) }
func BenchmarkOpArrayLit_10(b *testing.B)  { benchOpArrayLit(b, 10) }
func BenchmarkOpArrayLit_100(b *testing.B)  { benchOpArrayLit(b, 100) }
func BenchmarkOpArrayLit_1000(b *testing.B) { benchOpArrayLit(b, 1000) }

// ---------------------------------------------------------------------------
// doOpStructLit: PopExpr(CompositeLitExpr), PopValues(nFields), peek type, push struct.
// Parameterized by field count.
// ---------------------------------------------------------------------------

func benchOpStructLit(b *testing.B, nFields int) {
	m := benchMachine()
	defer m.Release()

	fields := make([]FieldType, nFields)
	for i := range nFields {
		fields[i] = FieldType{
			Name: Name("f" + string(rune('a'+i))),
			Type: IntType,
		}
	}
	st := m.Alloc.NewType(&StructType{
		PkgPath: "bench",
		Fields:  fields,
	})
	elts := make([]KeyValueExpr, nFields)
	for i := range nFields {
		elts[i] = KeyValueExpr{Value: &ConstExpr{}}
	}
	litExpr := &CompositeLitExpr{Elts: elts}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(asValue(st))
		for i := range nFields {
			m.PushValue(TypedValue{T: IntType, N: i2n(int64(i + 1))})
		}
		m.PushExpr(litExpr)
		bm.SwitchOpCode(bmTarget)
		m.doOpStructLit()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		sv := res.V.(*StructValue)
		if len(sv.Fields) != nFields {
			b.Fatalf("expected %d fields, got %d", nFields, len(sv.Fields))
		}
		if sv.Fields[0].GetInt() != 1 {
			b.Fatalf("expected first field 1, got %d", sv.Fields[0].GetInt())
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpStructLit_1(b *testing.B)    { benchOpStructLit(b, 1) }
func BenchmarkOpStructLit_10(b *testing.B)   { benchOpStructLit(b, 10) }
func BenchmarkOpStructLit_100(b *testing.B)  { benchOpStructLit(b, 100) }
func BenchmarkOpStructLit_1000(b *testing.B) { benchOpStructLit(b, 1000) }

// ---------------------------------------------------------------------------
// doOpStructLit (named fields): S{X: 1, Y: 2}.
// Named field path calls defaultStructFields, allocates fsset []bool,
// and does per-field Copy + index lookup.
// Parameterized by number of fields.
// ---------------------------------------------------------------------------

func benchOpStructLitNamed(b *testing.B, nFields int) {
	m := benchMachine()
	defer m.Release()

	fields := make([]FieldType, nFields)
	for i := range nFields {
		fields[i] = FieldType{
			Name: Name("f" + string(rune('a'+i))),
			Type: IntType,
		}
	}
	st := m.Alloc.NewType(&StructType{
		PkgPath: "bench",
		Fields:  fields,
	})
	elts := make([]KeyValueExpr, nFields)
	for i := range nFields {
		elts[i] = KeyValueExpr{
			Key: &NameExpr{
				Name: fields[i].Name,
				Path: ValuePath{Depth: 0, Index: uint16(i)},
			},
			Value: &ConstExpr{},
		}
	}
	litExpr := &CompositeLitExpr{Elts: elts}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(asValue(st))
		for i := range nFields {
			m.PushValue(TypedValue{T: IntType, N: i2n(int64(i + 1))})
		}
		m.PushExpr(litExpr)
		bm.SwitchOpCode(bmTarget)
		m.doOpStructLit()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		sv := res.V.(*StructValue)
		if len(sv.Fields) != nFields {
			b.Fatalf("expected %d fields, got %d", nFields, len(sv.Fields))
		}
		if sv.Fields[0].GetInt() != 1 {
			b.Fatalf("expected first field 1, got %d", sv.Fields[0].GetInt())
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpStructLitNamed_1(b *testing.B)    { benchOpStructLitNamed(b, 1) }
func BenchmarkOpStructLitNamed_10(b *testing.B)   { benchOpStructLitNamed(b, 10) }
func BenchmarkOpStructLitNamed_100(b *testing.B)  { benchOpStructLitNamed(b, 100) }
func BenchmarkOpStructLitNamed_1000(b *testing.B) { benchOpStructLitNamed(b, 1000) }

// ---------------------------------------------------------------------------
// doOpMapLit: PopExpr(CompositeLitExpr), PopValues(N*2), peek type, push map.
// Parameterized by entry count.
// ---------------------------------------------------------------------------

func benchOpMapLit(b *testing.B, n int) {
	m := benchMachine()
	defer m.Release()

	mt := m.Alloc.NewType(&MapType{Key: IntType, Value: IntType})
	elts := make([]KeyValueExpr, n)
	for i := range n {
		elts[i] = KeyValueExpr{Key: &ConstExpr{}, Value: &ConstExpr{}}
	}
	litExpr := &CompositeLitExpr{Elts: elts}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(asValue(mt))
		for i := range n {
			m.PushValue(TypedValue{T: IntType, N: i2n(int64(i))})      // key
			m.PushValue(TypedValue{T: IntType, N: i2n(int64(i * 10))}) // value
		}
		m.PushExpr(litExpr)
		bm.SwitchOpCode(bmTarget)
		m.doOpMapLit()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		mv := res.V.(*MapValue)
		if mv.List.Size != n {
			b.Fatalf("expected %d entries, got %d", n, mv.List.Size)
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpMapLit_1(b *testing.B)    { benchOpMapLit(b, 1) }
func BenchmarkOpMapLit_10(b *testing.B)   { benchOpMapLit(b, 10) }
func BenchmarkOpMapLit_100(b *testing.B)  { benchOpMapLit(b, 100) }
func BenchmarkOpMapLit_1000(b *testing.B) { benchOpMapLit(b, 1000) }

// ---------------------------------------------------------------------------
// doOpIndex1: PopExpr, PopValue(index), PeekValue(container); *xv = result
// Parameterized by container type (array, slice, map).
// ---------------------------------------------------------------------------

func BenchmarkOpIndex1_Array(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	expr := &IndexExpr{}

	at := &ArrayType{Elt: IntType, Len: 10}
	av := defaultArrayValue(m.Alloc, at)
	for i := range 10 {
		av.List[i] = TypedValue{T: IntType, N: i2n(int64(i * 10))}
	}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: at, V: av})
		m.PushValue(TypedValue{T: IntType, N: i2n(3)})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpIndex1()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.GetInt() != 30 {
			b.Fatalf("expected 30, got %d", res.GetInt())
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpIndex1_ByteArray(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	expr := &IndexExpr{}

	at := &ArrayType{Elt: Uint8Type, Len: 100}
	av := m.Alloc.NewDataArray(100)
	for i := range av.Data {
		av.Data[i] = byte(i % 256)
	}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: at, V: av})
		m.PushValue(TypedValue{T: IntType, N: i2n(42)})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpIndex1()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.GetUint8() != 42 {
			b.Fatalf("expected 42, got %d", res.GetUint8())
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpIndex1_Slice(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	expr := &IndexExpr{}

	st := &SliceType{Elt: IntType}
	baseArray := m.Alloc.NewListArray(10)
	for i := range 10 {
		baseArray.List[i] = TypedValue{T: IntType, N: i2n(int64(i * 10))}
	}
	sv := m.Alloc.NewSlice(baseArray, 0, 10, 10)

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: st, V: sv})
		m.PushValue(TypedValue{T: IntType, N: i2n(3)})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpIndex1()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.GetInt() != 30 {
			b.Fatalf("expected 30, got %d", res.GetInt())
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func benchOpIndex1MapHit(b *testing.B, size int) {
	m := benchMachine()
	defer m.Release()
	expr := &IndexExpr{}

	mt := &MapType{Key: IntType, Value: IntType}
	mv := &MapValue{}
	mv.MakeMap(size)
	for i := range size {
		kv := TypedValue{T: IntType, N: i2n(int64(i))}
		pv := mv.GetPointerForKey(m.Alloc, m.Store, kv)
		*pv.TV = TypedValue{T: IntType, N: i2n(int64(i * 10))}
	}
	// Look up a key near the middle.
	lookupKey := int64(size / 2)
	expected := lookupKey * 10

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: mt, V: mv})
		m.PushValue(TypedValue{T: IntType, N: i2n(lookupKey)})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpIndex1()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.GetInt() != expected {
			b.Fatalf("expected %d, got %d", expected, res.GetInt())
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpIndex1_MapHit_10(b *testing.B)     { benchOpIndex1MapHit(b, 10) }
func BenchmarkOpIndex1_MapHit_100(b *testing.B)    { benchOpIndex1MapHit(b, 100) }
func BenchmarkOpIndex1_MapHit_1000(b *testing.B)   { benchOpIndex1MapHit(b, 1000) }
func BenchmarkOpIndex1_MapHit_10000(b *testing.B)  { benchOpIndex1MapHit(b, 10000) }
func BenchmarkOpIndex1_MapHit_100000(b *testing.B) { benchOpIndex1MapHit(b, 100000) }

func BenchmarkOpIndex1_MapMiss(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	expr := &IndexExpr{}

	mt := &MapType{Key: IntType, Value: IntType}
	mv := &MapValue{}
	mv.MakeMap(10)
	for i := range 10 {
		kv := TypedValue{T: IntType, N: i2n(int64(i))}
		pv := mv.GetPointerForKey(m.Alloc, m.Store, kv)
		*pv.TV = TypedValue{T: IntType, N: i2n(int64(i * 10))}
	}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: mt, V: mv})
		m.PushValue(TypedValue{T: IntType, N: i2n(999)}) // key not in map
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpIndex1()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.GetInt() != 0 { // default int value
			b.Fatalf("expected 0 (default), got %d", res.GetInt())
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func benchOpIndex1_MapStringKey(b *testing.B, keyLen int) {
	m := benchMachine()
	defer m.Release()
	expr := &IndexExpr{}

	mt := &MapType{Key: StringType, Value: IntType}
	mv := &MapValue{}
	mv.MakeMap(10)
	for i := range 10 {
		k := strings.Repeat("x", keyLen-1) + string(rune('A'+i))
		kv := TypedValue{T: StringType, V: m.Alloc.NewString(k)}
		pv := mv.GetPointerForKey(m.Alloc, m.Store, kv)
		*pv.TV = TypedValue{T: IntType, N: i2n(int64(i))}
	}
	lookupKey := m.Alloc.NewString(strings.Repeat("x", keyLen-1) + string(rune('A'+5)))

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: mt, V: mv})
		m.PushValue(TypedValue{T: StringType, V: lookupKey})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpIndex1()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.GetInt() != 5 {
			b.Fatalf("expected 5, got %d", res.GetInt())
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpIndex1_MapStringKey_10(b *testing.B)    { benchOpIndex1_MapStringKey(b, 10) }
func BenchmarkOpIndex1_MapStringKey_100(b *testing.B)   { benchOpIndex1_MapStringKey(b, 100) }
func BenchmarkOpIndex1_MapStringKey_1000(b *testing.B)  { benchOpIndex1_MapStringKey(b, 1000) }
func BenchmarkOpIndex1_MapStringKey_10000(b *testing.B) { benchOpIndex1_MapStringKey(b, 10000) }

// ---------------------------------------------------------------------------
// doOpSelector: PopExpr(SelectorExpr), PeekValue(1); *xv = field value
// Parameterized by field count.
// ---------------------------------------------------------------------------

func benchOpSelector(b *testing.B, nFields int, fieldIdx int) {
	m := benchMachine()
	defer m.Release()

	fields := make([]FieldType, nFields)
	for i := range nFields {
		fields[i] = FieldType{
			Name: Name("f" + string(rune('a'+i))),
			Type: IntType,
		}
	}
	st := &StructType{PkgPath: "bench", Fields: fields}

	fieldValues := make([]TypedValue, nFields)
	for i := range nFields {
		fieldValues[i] = TypedValue{T: IntType, N: i2n(int64(i + 1))}
	}
	sv := m.Alloc.NewStruct(fieldValues)

	selExpr := &SelectorExpr{
		Path: ValuePath{
			Type:  VPField,
			Depth: 0,
			Index: uint16(fieldIdx),
			Name:  fields[fieldIdx].Name,
		},
	}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: st, V: sv})
		m.PushExpr(selExpr)
		bm.SwitchOpCode(bmTarget)
		m.doOpSelector()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.GetInt() != int64(fieldIdx+1) {
			b.Fatalf("expected %d, got %d", fieldIdx+1, res.GetInt())
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpSelector_1field(b *testing.B)     { benchOpSelector(b, 1, 0) }
func BenchmarkOpSelector_10fields(b *testing.B)   { benchOpSelector(b, 10, 9) }
func BenchmarkOpSelector_100fields(b *testing.B)  { benchOpSelector(b, 100, 99) }
func BenchmarkOpSelector_1000fields(b *testing.B) { benchOpSelector(b, 1000, 999) }

// ---------------------------------------------------------------------------
// Shift ops: parameterized by shift amount.
// doOpShl: PopExpr, PopValue(rv uint), PeekValue(lv); lv <<= rv
// ---------------------------------------------------------------------------

func benchOpShlParam(b *testing.B, shift uint64) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}
	expected := int64(1) << shift

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: IntType, N: i2n(1)})
		m.PushValue(TypedValue{T: UintType, N: u2n(shift)})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpShl()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.GetInt() != expected {
			b.Fatalf("expected %d, got %d", expected, res.GetInt())
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpShl_2(b *testing.B)  { benchOpShlParam(b, 2) }
func BenchmarkOpShl_32(b *testing.B) { benchOpShlParam(b, 32) }
func BenchmarkOpShl_62(b *testing.B) { benchOpShlParam(b, 62) }

// BigInt shift: allocates new big.Int, shift amount is the cost driver.
func benchOpShlBigInt(b *testing.B, shift uint64) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}
	bv := BigintValue{V: big.NewInt(1)}
	expected := new(big.Int).Lsh(big.NewInt(1), uint(shift))

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: UntypedBigintType, V: bv})
		m.PushValue(TypedValue{T: UintType, N: u2n(shift)})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpShl()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		got := res.V.(BigintValue).V
		if got.Cmp(expected) != 0 {
			b.Fatalf("expected %s, got %s", expected, got)
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpShl_BigInt_10(b *testing.B)   { benchOpShlBigInt(b, 10) }
func BenchmarkOpShl_BigInt_100(b *testing.B)  { benchOpShlBigInt(b, 100) }
func BenchmarkOpShl_BigInt_1000(b *testing.B) { benchOpShlBigInt(b, 1000) }

// ---------------------------------------------------------------------------
// doOpBandn: PopExpr, PopValue(rv), PeekValue(lv); lv = lv &^ rv
// ---------------------------------------------------------------------------

func BenchmarkOpBandn(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: IntType, N: i2n(0xFF)})
		m.PushValue(TypedValue{T: IntType, N: i2n(0x0F)})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpBandn()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.GetInt() != 0xF0 {
			b.Fatalf("expected 0xF0, got %d", res.GetInt())
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

// ---------------------------------------------------------------------------
// doOpIndex2: PopExpr, PeekValue(1)=index, PeekValue(2)=map; comma-ok pattern
// Returns value in xv slot and bool in iv slot.
// ---------------------------------------------------------------------------

func BenchmarkOpIndex2_MapHit(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	expr := &IndexExpr{}

	mt := &MapType{Key: IntType, Value: IntType}
	mv := &MapValue{}
	mv.MakeMap(10)
	for i := range 10 {
		kv := TypedValue{T: IntType, N: i2n(int64(i))}
		pv := mv.GetPointerForKey(m.Alloc, m.Store, kv)
		*pv.TV = TypedValue{T: IntType, N: i2n(int64(i * 10))}
	}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: mt, V: mv})
		m.PushValue(TypedValue{T: IntType, N: i2n(5)})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpIndex2()
		bm.SwitchOpCode(bmSetup)
		// xv is at position 2, iv (bool) at position 1
		boolRes := m.PeekValue(1)
		valRes := m.PeekValue(2)
		if !boolRes.GetBool() {
			b.Fatal("expected ok=true")
		}
		if valRes.GetInt() != 50 {
			b.Fatalf("expected 50, got %d", valRes.GetInt())
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpIndex2_MapMiss(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	expr := &IndexExpr{}

	mt := &MapType{Key: IntType, Value: IntType}
	mv := &MapValue{}
	mv.MakeMap(10)
	for i := range 10 {
		kv := TypedValue{T: IntType, N: i2n(int64(i))}
		pv := mv.GetPointerForKey(m.Alloc, m.Store, kv)
		*pv.TV = TypedValue{T: IntType, N: i2n(int64(i * 10))}
	}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: mt, V: mv})
		m.PushValue(TypedValue{T: IntType, N: i2n(999)})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpIndex2()
		bm.SwitchOpCode(bmSetup)
		boolRes := m.PeekValue(1)
		if boolRes.GetBool() {
			b.Fatal("expected ok=false")
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

// ---------------------------------------------------------------------------
// doOpSlice: PopExpr(SliceExpr), pop high/low, pop base; push slice result.
// ---------------------------------------------------------------------------

func BenchmarkOpSlice_Array(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	sliceExpr := &SliceExpr{
		Low:  &ConstExpr{},
		High: &ConstExpr{},
	}

	at := &ArrayType{Elt: IntType, Len: 100}
	av := defaultArrayValue(m.Alloc, at)
	for i := range 100 {
		av.List[i] = TypedValue{T: IntType, N: i2n(int64(i))}
	}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: at, V: av})   // base
		m.PushValue(TypedValue{T: IntType, N: i2n(10)}) // low
		m.PushValue(TypedValue{T: IntType, N: i2n(50)}) // high
		m.PushExpr(sliceExpr)
		bm.SwitchOpCode(bmTarget)
		m.doOpSlice()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		sv := res.V.(*SliceValue)
		if sv.Length != 40 {
			b.Fatalf("expected len 40, got %d", sv.Length)
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpSlice_ByteArray(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	sliceExpr := &SliceExpr{
		Low:  &ConstExpr{},
		High: &ConstExpr{},
	}

	at := &ArrayType{Elt: Uint8Type, Len: 100}
	av := m.Alloc.NewDataArray(100)
	for i := range av.Data {
		av.Data[i] = byte(i)
	}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: at, V: av})          // base
		m.PushValue(TypedValue{T: IntType, N: i2n(10)}) // low
		m.PushValue(TypedValue{T: IntType, N: i2n(50)}) // high
		m.PushExpr(sliceExpr)
		bm.SwitchOpCode(bmTarget)
		m.doOpSlice()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		sv := res.V.(*SliceValue)
		if sv.Length != 40 {
			b.Fatalf("expected len 40, got %d", sv.Length)
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpSlice_Slice(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	sliceExpr := &SliceExpr{
		Low:  &ConstExpr{},
		High: &ConstExpr{},
	}

	st := &SliceType{Elt: IntType}
	baseArray := m.Alloc.NewListArray(100)
	for i := range 100 {
		baseArray.List[i] = TypedValue{T: IntType, N: i2n(int64(i))}
	}
	sv := m.Alloc.NewSlice(baseArray, 0, 100, 100)

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: st, V: sv})           // base
		m.PushValue(TypedValue{T: IntType, N: i2n(10)}) // low
		m.PushValue(TypedValue{T: IntType, N: i2n(50)}) // high
		m.PushExpr(sliceExpr)
		bm.SwitchOpCode(bmTarget)
		m.doOpSlice()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		rsv := res.V.(*SliceValue)
		if rsv.Length != 40 {
			b.Fatalf("expected len 40, got %d", rsv.Length)
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func benchOpSlice_String(b *testing.B, length int) {
	m := benchMachine()
	defer m.Release()
	sliceExpr := &SliceExpr{
		Low:  &ConstExpr{},
		High: &ConstExpr{},
	}

	s := strings.Repeat("x", length)
	sv := m.Alloc.NewString(s)
	half := length / 2

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: StringType, V: sv})          // base
		m.PushValue(TypedValue{T: IntType, N: i2n(0)})         // low
		m.PushValue(TypedValue{T: IntType, N: i2n(int64(half))}) // high
		m.PushExpr(sliceExpr)
		bm.SwitchOpCode(bmTarget)
		m.doOpSlice()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if len(res.GetString()) != half {
			b.Fatalf("expected len %d, got %d", half, len(res.GetString()))
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpSlice_String_10(b *testing.B)    { benchOpSlice_String(b, 10) }
func BenchmarkOpSlice_String_100(b *testing.B)   { benchOpSlice_String(b, 100) }
func BenchmarkOpSlice_String_1000(b *testing.B)  { benchOpSlice_String(b, 1000) }
func BenchmarkOpSlice_String_10000(b *testing.B) { benchOpSlice_String(b, 10000) }

// ---------------------------------------------------------------------------
// doOpStar: PopValue, dereference pointer or get pointer-to type.
// ---------------------------------------------------------------------------

func BenchmarkOpStar(b *testing.B) {
	m := benchMachine()
	defer m.Release()

	pt := &PointerType{Elt: IntType}
	target := TypedValue{T: IntType, N: i2n(42)}
	pv := PointerValue{TV: &target}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: pt, V: pv})
		bm.SwitchOpCode(bmTarget)
		m.doOpStar()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.GetInt() != 42 {
			b.Fatalf("expected 42, got %d", res.GetInt())
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

// ---------------------------------------------------------------------------
// doOpTypeAssert1: PopExpr, PopValue(type), PeekValue(value); concrete assert.
// ---------------------------------------------------------------------------

func BenchmarkOpTypeAssert1(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	expr := &TypeAssertExpr{}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		// value to assert (concrete IntType)
		m.PushValue(TypedValue{T: IntType, N: i2n(42)})
		// type to assert against
		m.PushValue(asValue(IntType))
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpTypeAssert1()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.GetInt() != 42 {
			b.Fatalf("expected 42, got %d", res.GetInt())
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

// ---------------------------------------------------------------------------
// doOpTypeAssert2: PopExpr, PeekValue(1)=type, PeekValue(2)=value; comma-ok.
// ---------------------------------------------------------------------------

func BenchmarkOpTypeAssert2_Hit(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	expr := &TypeAssertExpr{}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: IntType, N: i2n(42)})
		m.PushValue(asValue(IntType))
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpTypeAssert2()
		bm.SwitchOpCode(bmSetup)
		boolRes := m.PeekValue(1)
		valRes := m.PeekValue(2)
		if !boolRes.GetBool() {
			b.Fatal("expected ok=true")
		}
		if valRes.GetInt() != 42 {
			b.Fatalf("expected 42, got %d", valRes.GetInt())
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpTypeAssert2_Miss(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	expr := &TypeAssertExpr{}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: IntType, N: i2n(42)})
		m.PushValue(asValue(StringType))
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpTypeAssert2()
		bm.SwitchOpCode(bmSetup)
		boolRes := m.PeekValue(1)
		if boolRes.GetBool() {
			b.Fatal("expected ok=false")
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

// ---------------------------------------------------------------------------
// doOpConvert: PopValue(value), PopValue(type); ConvertTo then push result.
// ---------------------------------------------------------------------------

func BenchmarkOpConvert_IntToInt64(b *testing.B) {
	m := benchMachine()
	defer m.Release()

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(asValue(Int64Type))
		m.PushValue(TypedValue{T: IntType, N: i2n(42)})
		bm.SwitchOpCode(bmTarget)
		m.doOpConvert()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.GetInt64() != 42 {
			b.Fatalf("expected 42, got %d", res.GetInt64())
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func benchOpConvert_StringToBytes(b *testing.B, length int) {
	m := benchMachine()
	defer m.Release()

	uint8SliceType := &SliceType{Elt: Uint8Type}
	sv := m.Alloc.NewString(strings.Repeat("x", length))

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(asValue(uint8SliceType))
		m.PushValue(TypedValue{T: StringType, V: sv})
		bm.SwitchOpCode(bmTarget)
		m.doOpConvert()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		rsv := res.V.(*SliceValue)
		if rsv.Length != length {
			b.Fatalf("expected len %d, got %d", length, rsv.Length)
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpConvert_StringToBytes_10(b *testing.B)    { benchOpConvert_StringToBytes(b, 10) }
func BenchmarkOpConvert_StringToBytes_100(b *testing.B)   { benchOpConvert_StringToBytes(b, 100) }
func BenchmarkOpConvert_StringToBytes_1000(b *testing.B)  { benchOpConvert_StringToBytes(b, 1000) }
func BenchmarkOpConvert_StringToBytes_10000(b *testing.B) { benchOpConvert_StringToBytes(b, 10000) }

// ---------------------------------------------------------------------------
// doOpInc/doOpDec: PopStmt(IncDecStmt), PopAsPointer(s.X); mutate in place.
// Requires block setup with a named variable.
// ---------------------------------------------------------------------------

// benchBlockVar creates a Machine with a block containing one IntType variable
// and returns the NameExpr needed to reference it via PopAsPointer.
func benchBlockVar(m *Machine) (*Block, *NameExpr) {
	blk, nxs := benchBlockVars(m, 1)
	return blk, nxs[0]
}

// benchBlockVars creates a block with n IntType variables named x0..xN-1.
func benchBlockVars(m *Machine, n int) (*Block, []*NameExpr) {
	values := make([]TypedValue, n)
	for i := range n {
		values[i] = TypedValue{T: IntType, N: i2n(0)}
	}
	blk := &Block{Values: values}
	m.Blocks = append(m.Blocks, blk)

	nxs := make([]*NameExpr, n)
	for i := range n {
		name := Name("x" + string(rune('0'+i)))
		nxs[i] = &NameExpr{
			Name: name,
			Path: ValuePath{
				Type:  VPBlock,
				Depth: 1,
				Index: uint16(i),
				Name:  name,
			},
		}
	}
	return blk, nxs
}

func BenchmarkOpInc_Int(b *testing.B) {
	m := benchMachine()
	defer m.Release()

	blk, nx := benchBlockVar(m)
	stmt := &IncDecStmt{X: nx}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		blk.Values[0] = TypedValue{T: IntType, N: i2n(0)}
		m.PushStmt(stmt)
		bm.SwitchOpCode(bmTarget)
		m.doOpInc()
		bm.SwitchOpCode(bmSetup)
		if blk.Values[0].GetInt() != 1 {
			b.Fatalf("expected 1, got %d", blk.Values[0].GetInt())
		}
	}
	reportBenchops(b)
}

func BenchmarkOpDec_Int(b *testing.B) {
	m := benchMachine()
	defer m.Release()

	blk, nx := benchBlockVar(m)
	stmt := &IncDecStmt{X: nx}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		blk.Values[0] = TypedValue{T: IntType, N: i2n(10)}
		m.PushStmt(stmt)
		bm.SwitchOpCode(bmTarget)
		m.doOpDec()
		bm.SwitchOpCode(bmSetup)
		if blk.Values[0].GetInt() != 9 {
			b.Fatalf("expected 9, got %d", blk.Values[0].GetInt())
		}
	}
	reportBenchops(b)
}

// ---------------------------------------------------------------------------
// doOpAddAssign: PopStmt, PopValue(rv), PopAsPointer(lhs); lv += rv.
// ---------------------------------------------------------------------------

func BenchmarkOpAddAssign_Int(b *testing.B) {
	m := benchMachine()
	defer m.Release()

	blk, nx := benchBlockVar(m)
	stmt := &AssignStmt{
		Lhs: []Expr{nx},
		Op:  ADD_ASSIGN,
	}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		blk.Values[0] = TypedValue{T: IntType, N: i2n(10)}
		m.PushValue(TypedValue{T: IntType, N: i2n(5)})
		m.PushStmt(stmt)
		bm.SwitchOpCode(bmTarget)
		m.doOpAddAssign()
		bm.SwitchOpCode(bmSetup)
		if blk.Values[0].GetInt() != 15 {
			b.Fatalf("expected 15, got %d", blk.Values[0].GetInt())
		}
	}
	reportBenchops(b)
}

// ---------------------------------------------------------------------------
// doOpSubAssign: PopStmt, PopValue(rv), PopAsPointer(lhs); lv -= rv.
// ---------------------------------------------------------------------------

func BenchmarkOpSubAssign_Int(b *testing.B) {
	m := benchMachine()
	defer m.Release()

	blk, nx := benchBlockVar(m)
	stmt := &AssignStmt{
		Lhs: []Expr{nx},
		Op:  SUB_ASSIGN,
	}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		blk.Values[0] = TypedValue{T: IntType, N: i2n(10)}
		m.PushValue(TypedValue{T: IntType, N: i2n(3)})
		m.PushStmt(stmt)
		bm.SwitchOpCode(bmTarget)
		m.doOpSubAssign()
		bm.SwitchOpCode(bmSetup)
		if blk.Values[0].GetInt() != 7 {
			b.Fatalf("expected 7, got %d", blk.Values[0].GetInt())
		}
	}
	reportBenchops(b)
}

// ---------------------------------------------------------------------------
// doOpMulAssign: PopStmt, PopValue(rv), PopAsPointer(lhs); lv *= rv.
// ---------------------------------------------------------------------------

func BenchmarkOpMulAssign_Int(b *testing.B) {
	m := benchMachine()
	defer m.Release()

	blk, nx := benchBlockVar(m)
	stmt := &AssignStmt{
		Lhs: []Expr{nx},
		Op:  MUL_ASSIGN,
	}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		blk.Values[0] = TypedValue{T: IntType, N: i2n(6)}
		m.PushValue(TypedValue{T: IntType, N: i2n(7)})
		m.PushStmt(stmt)
		bm.SwitchOpCode(bmTarget)
		m.doOpMulAssign()
		bm.SwitchOpCode(bmSetup)
		if blk.Values[0].GetInt() != 42 {
			b.Fatalf("expected 42, got %d", blk.Values[0].GetInt())
		}
	}
	reportBenchops(b)
}

// ---------------------------------------------------------------------------
// doOpQuoAssign, doOpRemAssign, doOpBandAssign, etc.
// ---------------------------------------------------------------------------

func BenchmarkOpQuoAssign_Int(b *testing.B) {
	m := benchMachine()
	defer m.Release()

	blk, nx := benchBlockVar(m)
	stmt := &AssignStmt{Lhs: []Expr{nx}, Op: QUO_ASSIGN}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		blk.Values[0] = TypedValue{T: IntType, N: i2n(42)}
		m.PushValue(TypedValue{T: IntType, N: i2n(6)})
		m.PushStmt(stmt)
		bm.SwitchOpCode(bmTarget)
		m.doOpQuoAssign()
		bm.SwitchOpCode(bmSetup)
		if blk.Values[0].GetInt() != 7 {
			b.Fatalf("expected 7, got %d", blk.Values[0].GetInt())
		}
	}
	reportBenchops(b)
}

func BenchmarkOpRemAssign_Int(b *testing.B) {
	m := benchMachine()
	defer m.Release()

	blk, nx := benchBlockVar(m)
	stmt := &AssignStmt{Lhs: []Expr{nx}, Op: REM_ASSIGN}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		blk.Values[0] = TypedValue{T: IntType, N: i2n(42)}
		m.PushValue(TypedValue{T: IntType, N: i2n(5)})
		m.PushStmt(stmt)
		bm.SwitchOpCode(bmTarget)
		m.doOpRemAssign()
		bm.SwitchOpCode(bmSetup)
		if blk.Values[0].GetInt() != 2 {
			b.Fatalf("expected 2, got %d", blk.Values[0].GetInt())
		}
	}
	reportBenchops(b)
}

func BenchmarkOpBandAssign_Int(b *testing.B) {
	m := benchMachine()
	defer m.Release()

	blk, nx := benchBlockVar(m)
	stmt := &AssignStmt{Lhs: []Expr{nx}, Op: BAND_ASSIGN}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		blk.Values[0] = TypedValue{T: IntType, N: i2n(0xFF)}
		m.PushValue(TypedValue{T: IntType, N: i2n(0x0F)})
		m.PushStmt(stmt)
		bm.SwitchOpCode(bmTarget)
		m.doOpBandAssign()
		bm.SwitchOpCode(bmSetup)
		if blk.Values[0].GetInt() != 0x0F {
			b.Fatalf("expected 0x0F, got %d", blk.Values[0].GetInt())
		}
	}
	reportBenchops(b)
}

func BenchmarkOpBorAssign_Int(b *testing.B) {
	m := benchMachine()
	defer m.Release()

	blk, nx := benchBlockVar(m)
	stmt := &AssignStmt{Lhs: []Expr{nx}, Op: BOR_ASSIGN}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		blk.Values[0] = TypedValue{T: IntType, N: i2n(0xF0)}
		m.PushValue(TypedValue{T: IntType, N: i2n(0x0F)})
		m.PushStmt(stmt)
		bm.SwitchOpCode(bmTarget)
		m.doOpBorAssign()
		bm.SwitchOpCode(bmSetup)
		if blk.Values[0].GetInt() != 0xFF {
			b.Fatalf("expected 0xFF, got %d", blk.Values[0].GetInt())
		}
	}
	reportBenchops(b)
}

func BenchmarkOpXorAssign_Int(b *testing.B) {
	m := benchMachine()
	defer m.Release()

	blk, nx := benchBlockVar(m)
	stmt := &AssignStmt{Lhs: []Expr{nx}, Op: XOR_ASSIGN}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		blk.Values[0] = TypedValue{T: IntType, N: i2n(0xFF)}
		m.PushValue(TypedValue{T: IntType, N: i2n(0x0F)})
		m.PushStmt(stmt)
		bm.SwitchOpCode(bmTarget)
		m.doOpXorAssign()
		bm.SwitchOpCode(bmSetup)
		if blk.Values[0].GetInt() != 0xF0 {
			b.Fatalf("expected 0xF0, got %d", blk.Values[0].GetInt())
		}
	}
	reportBenchops(b)
}

func BenchmarkOpShlAssign_Int(b *testing.B) {
	m := benchMachine()
	defer m.Release()

	blk, nx := benchBlockVar(m)
	stmt := &AssignStmt{Lhs: []Expr{nx}, Op: SHL_ASSIGN}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		blk.Values[0] = TypedValue{T: IntType, N: i2n(1)}
		m.PushValue(TypedValue{T: UintType, N: u2n(10)})
		m.PushStmt(stmt)
		bm.SwitchOpCode(bmTarget)
		m.doOpShlAssign()
		bm.SwitchOpCode(bmSetup)
		if blk.Values[0].GetInt() != 1024 {
			b.Fatalf("expected 1024, got %d", blk.Values[0].GetInt())
		}
	}
	reportBenchops(b)
}

func BenchmarkOpShrAssign_Int(b *testing.B) {
	m := benchMachine()
	defer m.Release()

	blk, nx := benchBlockVar(m)
	stmt := &AssignStmt{Lhs: []Expr{nx}, Op: SHR_ASSIGN}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		blk.Values[0] = TypedValue{T: IntType, N: i2n(1024)}
		m.PushValue(TypedValue{T: UintType, N: u2n(10)})
		m.PushStmt(stmt)
		bm.SwitchOpCode(bmTarget)
		m.doOpShrAssign()
		bm.SwitchOpCode(bmSetup)
		if blk.Values[0].GetInt() != 1 {
			b.Fatalf("expected 1, got %d", blk.Values[0].GetInt())
		}
	}
	reportBenchops(b)
}

func BenchmarkOpBandnAssign_Int(b *testing.B) {
	m := benchMachine()
	defer m.Release()

	blk, nx := benchBlockVar(m)
	stmt := &AssignStmt{Lhs: []Expr{nx}, Op: BAND_NOT_ASSIGN}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		blk.Values[0] = TypedValue{T: IntType, N: i2n(0xFF)}
		m.PushValue(TypedValue{T: IntType, N: i2n(0x0F)})
		m.PushStmt(stmt)
		bm.SwitchOpCode(bmTarget)
		m.doOpBandnAssign()
		bm.SwitchOpCode(bmSetup)
		if blk.Values[0].GetInt() != 0xF0 {
			b.Fatalf("expected 0xF0, got %d", blk.Values[0].GetInt())
		}
	}
	reportBenchops(b)
}

// ---------------------------------------------------------------------------
// doOpDefine: PopStmt, PopValues(n), LastBlock; define variables.
// Parameterized by number of variables.
// ---------------------------------------------------------------------------

func benchOpDefine(b *testing.B, n int) {
	m := benchMachine()
	defer m.Release()

	blk, nxs := benchBlockVars(m, n)
	lhs := make([]Expr, n)
	for i := range n {
		lhs[i] = &NameExpr{
			Name: nxs[i].Name,
			Type: NameExprTypeDefine,
			Path: nxs[i].Path,
		}
	}
	stmt := &AssignStmt{Lhs: lhs, Op: DEFINE}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		for i := range n {
			blk.Values[i] = TypedValue{} // reset
			m.PushValue(TypedValue{T: IntType, N: i2n(int64(i + 1))})
		}
		m.PushStmt(stmt)
		bm.SwitchOpCode(bmTarget)
		m.doOpDefine()
		bm.SwitchOpCode(bmSetup)
		for i := range n {
			if blk.Values[i].GetInt() != int64(i+1) {
				b.Fatalf("var %d: expected %d, got %d", i, i+1, blk.Values[i].GetInt())
			}
		}
	}
	reportBenchops(b)
}

func BenchmarkOpDefine_1(b *testing.B)   { benchOpDefine(b, 1) }
func BenchmarkOpDefine_10(b *testing.B)  { benchOpDefine(b, 10) }
func BenchmarkOpDefine_100(b *testing.B)  { benchOpDefine(b, 100) }
func BenchmarkOpDefine_1000(b *testing.B) { benchOpDefine(b, 1000) }

// ---------------------------------------------------------------------------
// doOpAssign: PopStmt, PopValues(n), PopAsPointer for each lhs.
// Parameterized by number of variables.
// ---------------------------------------------------------------------------

func benchOpAssign(b *testing.B, n int) {
	m := benchMachine()
	defer m.Release()

	blk, nxs := benchBlockVars(m, n)
	lhs := make([]Expr, n)
	for i := range n {
		lhs[i] = nxs[i]
	}
	stmt := &AssignStmt{Lhs: lhs, Op: ASSIGN}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		for i := range n {
			blk.Values[i] = TypedValue{T: IntType, N: i2n(0)}
			m.PushValue(TypedValue{T: IntType, N: i2n(int64(i + 10))})
		}
		m.PushStmt(stmt)
		bm.SwitchOpCode(bmTarget)
		m.doOpAssign()
		bm.SwitchOpCode(bmSetup)
		for i := range n {
			if blk.Values[i].GetInt() != int64(i+10) {
				b.Fatalf("var %d: expected %d, got %d", i, i+10, blk.Values[i].GetInt())
			}
		}
	}
	reportBenchops(b)
}

func BenchmarkOpAssign_1(b *testing.B)   { benchOpAssign(b, 1) }
func BenchmarkOpAssign_10(b *testing.B)  { benchOpAssign(b, 10) }
func BenchmarkOpAssign_100(b *testing.B)  { benchOpAssign(b, 100) }
func BenchmarkOpAssign_1000(b *testing.B) { benchOpAssign(b, 1000) }

// ===========================================================================
// Pessimistic type variants — BigInt bitwise, Float64/BigInt comparisons,
// Shr BigInt, Inc/Dec Float64/BigInt, long string Convert.
// ===========================================================================

// --- BigInt bitwise ops (all allocate big.NewInt(0).Op) ---

func benchOpBand_BigInt(b *testing.B, bits int) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}
	v1 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(bits)), big.NewInt(1))
	v2 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(bits)), big.NewInt(1))
	bv1 := BigintValue{V: v1}
	bv2 := BigintValue{V: v2}
	expected := new(big.Int).And(v1, v2)

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: UntypedBigintType, V: bv1})
		m.PushValue(TypedValue{T: UntypedBigintType, V: bv2})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpBand()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.V.(BigintValue).V.Cmp(expected) != 0 {
			b.Fatal("unexpected result")
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpBand_BigInt_64(b *testing.B)   { benchOpBand_BigInt(b, 64) }
func BenchmarkOpBand_BigInt_256(b *testing.B)  { benchOpBand_BigInt(b, 256) }
func BenchmarkOpBand_BigInt_1024(b *testing.B) { benchOpBand_BigInt(b, 1024) }
func BenchmarkOpBand_BigInt_4096(b *testing.B) { benchOpBand_BigInt(b, 4096) }

func benchOpBor_BigInt(b *testing.B, bits int) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}
	v1 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(bits)), big.NewInt(1))
	v2 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(bits)), big.NewInt(1))
	bv1 := BigintValue{V: v1}
	bv2 := BigintValue{V: v2}
	expected := new(big.Int).Or(v1, v2)

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: UntypedBigintType, V: bv1})
		m.PushValue(TypedValue{T: UntypedBigintType, V: bv2})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpBor()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.V.(BigintValue).V.Cmp(expected) != 0 {
			b.Fatal("unexpected result")
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpBor_BigInt_64(b *testing.B)   { benchOpBor_BigInt(b, 64) }
func BenchmarkOpBor_BigInt_256(b *testing.B)  { benchOpBor_BigInt(b, 256) }
func BenchmarkOpBor_BigInt_1024(b *testing.B) { benchOpBor_BigInt(b, 1024) }
func BenchmarkOpBor_BigInt_4096(b *testing.B) { benchOpBor_BigInt(b, 4096) }

func benchOpXor_BigInt(b *testing.B, bits int) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}
	v1 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(bits)), big.NewInt(1))
	v2 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(bits)), big.NewInt(1))
	bv1 := BigintValue{V: v1}
	bv2 := BigintValue{V: v2}
	expected := new(big.Int).Xor(v1, v2)

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: UntypedBigintType, V: bv1})
		m.PushValue(TypedValue{T: UntypedBigintType, V: bv2})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpXor()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.V.(BigintValue).V.Cmp(expected) != 0 {
			b.Fatal("unexpected result")
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpXor_BigInt_64(b *testing.B)   { benchOpXor_BigInt(b, 64) }
func BenchmarkOpXor_BigInt_256(b *testing.B)  { benchOpXor_BigInt(b, 256) }
func BenchmarkOpXor_BigInt_1024(b *testing.B) { benchOpXor_BigInt(b, 1024) }
func BenchmarkOpXor_BigInt_4096(b *testing.B) { benchOpXor_BigInt(b, 4096) }

func benchOpBandn_BigInt(b *testing.B, bits int) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}
	v1 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(bits)), big.NewInt(1))
	v2 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(bits)), big.NewInt(1))
	bv1 := BigintValue{V: v1}
	bv2 := BigintValue{V: v2}
	expected := new(big.Int).AndNot(v1, v2)

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: UntypedBigintType, V: bv1})
		m.PushValue(TypedValue{T: UntypedBigintType, V: bv2})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpBandn()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.V.(BigintValue).V.Cmp(expected) != 0 {
			b.Fatal("unexpected result")
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpBandn_BigInt_64(b *testing.B)   { benchOpBandn_BigInt(b, 64) }
func BenchmarkOpBandn_BigInt_256(b *testing.B)  { benchOpBandn_BigInt(b, 256) }
func BenchmarkOpBandn_BigInt_1024(b *testing.B) { benchOpBandn_BigInt(b, 1024) }
func BenchmarkOpBandn_BigInt_4096(b *testing.B) { benchOpBandn_BigInt(b, 4096) }

// --- Shr BigInt ---

// --- Float64 and BigInt comparisons (softfloat / big.Cmp) ---

func BenchmarkOpEql_Float64(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}
	v1 := math.Float64bits(3.14159)
	v2 := math.Float64bits(2.71828)

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: Float64Type, N: u2n(v1)})
		m.PushValue(TypedValue{T: Float64Type, N: u2n(v2)})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpEql()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.GetBool() {
			b.Fatal("expected false")
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func benchOpEql_BigInt(b *testing.B, bits int) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}
	v1 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(bits)), big.NewInt(1))
	v2 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(bits)), big.NewInt(1))
	// Make them differ so result is false.
	v2.Sub(v2, big.NewInt(1))
	bv1 := BigintValue{V: v1}
	bv2 := BigintValue{V: v2}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: UntypedBigintType, V: bv1})
		m.PushValue(TypedValue{T: UntypedBigintType, V: bv2})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpEql()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.GetBool() {
			b.Fatal("expected false")
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpEql_BigInt_64(b *testing.B)   { benchOpEql_BigInt(b, 64) }
func BenchmarkOpEql_BigInt_256(b *testing.B)  { benchOpEql_BigInt(b, 256) }
func BenchmarkOpEql_BigInt_1024(b *testing.B) { benchOpEql_BigInt(b, 1024) }
func BenchmarkOpEql_BigInt_4096(b *testing.B) { benchOpEql_BigInt(b, 4096) }

func BenchmarkOpLss_Float64(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}
	v1 := math.Float64bits(2.71828)
	v2 := math.Float64bits(3.14159)

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: Float64Type, N: u2n(v1)})
		m.PushValue(TypedValue{T: Float64Type, N: u2n(v2)})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpLss()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if !res.GetBool() {
			b.Fatal("expected true")
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func benchOpLss_BigInt(b *testing.B, bits int) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}
	v1 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(bits)), big.NewInt(1))
	v2 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(bits)), big.NewInt(1))
	// Make v1 < v2.
	v1.Sub(v1, big.NewInt(1))
	bv1 := BigintValue{V: v1}
	bv2 := BigintValue{V: v2}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: UntypedBigintType, V: bv1})
		m.PushValue(TypedValue{T: UntypedBigintType, V: bv2})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpLss()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if !res.GetBool() {
			b.Fatal("expected true")
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpLss_BigInt_64(b *testing.B)   { benchOpLss_BigInt(b, 64) }
func BenchmarkOpLss_BigInt_256(b *testing.B)  { benchOpLss_BigInt(b, 256) }
func BenchmarkOpLss_BigInt_1024(b *testing.B) { benchOpLss_BigInt(b, 1024) }
func BenchmarkOpLss_BigInt_4096(b *testing.B) { benchOpLss_BigInt(b, 4096) }

func BenchmarkOpGtr_Float64(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}
	v1 := math.Float64bits(3.14159)
	v2 := math.Float64bits(2.71828)

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: Float64Type, N: u2n(v1)})
		m.PushValue(TypedValue{T: Float64Type, N: u2n(v2)})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpGtr()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if !res.GetBool() {
			b.Fatal("expected true")
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

// --- Inc/Dec Float64 and BigInt ---

func BenchmarkOpInc_Float64(b *testing.B) {
	m := benchMachine()
	defer m.Release()

	blk, nx := benchBlockVar(m)
	stmt := &IncDecStmt{X: nx}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		blk.Values[0] = TypedValue{T: Float64Type, N: u2n(math.Float64bits(1.0))}
		m.PushStmt(stmt)
		bm.SwitchOpCode(bmTarget)
		m.doOpInc()
		bm.SwitchOpCode(bmSetup)
		got := math.Float64frombits(blk.Values[0].GetFloat64())
		if got < 1.9 || got > 2.1 {
			b.Fatalf("expected ~2.0, got %f", got)
		}
	}
	reportBenchops(b)
}

func benchOpInc_BigInt(b *testing.B, bits int) {
	m := benchMachine()
	defer m.Release()

	blk, nx := benchBlockVar(m)
	stmt := &IncDecStmt{X: nx}
	v := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(bits)), big.NewInt(1))
	expected := new(big.Int).Add(v, big.NewInt(1))

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		blk.Values[0] = TypedValue{T: UntypedBigintType, V: BigintValue{V: new(big.Int).Set(v)}}
		m.PushStmt(stmt)
		bm.SwitchOpCode(bmTarget)
		m.doOpInc()
		bm.SwitchOpCode(bmSetup)
		got := blk.Values[0].V.(BigintValue).V
		if got.Cmp(expected) != 0 {
			b.Fatal("unexpected result")
		}
	}
	reportBenchops(b)
}

func BenchmarkOpInc_BigInt_64(b *testing.B)   { benchOpInc_BigInt(b, 64) }
func BenchmarkOpInc_BigInt_256(b *testing.B)  { benchOpInc_BigInt(b, 256) }
func BenchmarkOpInc_BigInt_1024(b *testing.B) { benchOpInc_BigInt(b, 1024) }
func BenchmarkOpInc_BigInt_4096(b *testing.B) { benchOpInc_BigInt(b, 4096) }

func BenchmarkOpDec_Float64(b *testing.B) {
	m := benchMachine()
	defer m.Release()

	blk, nx := benchBlockVar(m)
	stmt := &IncDecStmt{X: nx}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		blk.Values[0] = TypedValue{T: Float64Type, N: u2n(math.Float64bits(10.0))}
		m.PushStmt(stmt)
		bm.SwitchOpCode(bmTarget)
		m.doOpDec()
		bm.SwitchOpCode(bmSetup)
		got := math.Float64frombits(blk.Values[0].GetFloat64())
		if got < 8.9 || got > 9.1 {
			b.Fatalf("expected ~9.0, got %f", got)
		}
	}
	reportBenchops(b)
}

func benchOpDec_BigInt(b *testing.B, bits int) {
	m := benchMachine()
	defer m.Release()

	blk, nx := benchBlockVar(m)
	stmt := &IncDecStmt{X: nx}
	v := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(bits)), big.NewInt(1))
	expected := new(big.Int).Sub(v, big.NewInt(1))

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		blk.Values[0] = TypedValue{T: UntypedBigintType, V: BigintValue{V: new(big.Int).Set(v)}}
		m.PushStmt(stmt)
		bm.SwitchOpCode(bmTarget)
		m.doOpDec()
		bm.SwitchOpCode(bmSetup)
		got := blk.Values[0].V.(BigintValue).V
		if got.Cmp(expected) != 0 {
			b.Fatal("unexpected result")
		}
	}
	reportBenchops(b)
}

func BenchmarkOpDec_BigInt_64(b *testing.B)   { benchOpDec_BigInt(b, 64) }
func BenchmarkOpDec_BigInt_256(b *testing.B)  { benchOpDec_BigInt(b, 256) }
func BenchmarkOpDec_BigInt_1024(b *testing.B) { benchOpDec_BigInt(b, 1024) }
func BenchmarkOpDec_BigInt_4096(b *testing.B) { benchOpDec_BigInt(b, 4096) }

// ---------------------------------------------------------------------------
// BigDec benchmarks parameterized by number of decimal digits
// ---------------------------------------------------------------------------

// makeBigDec creates a BigdecValue with approximately `digits` decimal digits.
// Uses strings.Repeat to build a number like "1234567890123..." of the given length.
func makeBigDec(digits int) BigdecValue {
	s := strings.Repeat("1234567890", (digits/10)+1)[:digits]
	d, _, err := apd.NewFromString(s)
	if err != nil {
		panic(err)
	}
	return BigdecValue{V: d}
}

// --- doOpAdd BigDec ---

func benchOpAdd_BigDec(b *testing.B, digits int) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}
	bv1 := makeBigDec(digits)
	bv2 := makeBigDec(digits)

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: UntypedBigdecType, V: bv1})
		m.PushValue(TypedValue{T: UntypedBigdecType, V: bv2})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpAdd()
		bm.SwitchOpCode(bmSetup)
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpAdd_BigDec_10(b *testing.B)    { benchOpAdd_BigDec(b, 10) }
func BenchmarkOpAdd_BigDec_100(b *testing.B)   { benchOpAdd_BigDec(b, 100) }
func BenchmarkOpAdd_BigDec_1000(b *testing.B)  { benchOpAdd_BigDec(b, 1000) }
func BenchmarkOpAdd_BigDec_10000(b *testing.B) { benchOpAdd_BigDec(b, 10000) }

// --- doOpSub BigDec ---

func benchOpSub_BigDec(b *testing.B, digits int) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}
	bv1 := makeBigDec(digits)
	bv2 := makeBigDec(digits)

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: UntypedBigdecType, V: bv1})
		m.PushValue(TypedValue{T: UntypedBigdecType, V: bv2})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpSub()
		bm.SwitchOpCode(bmSetup)
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpSub_BigDec_10(b *testing.B)    { benchOpSub_BigDec(b, 10) }
func BenchmarkOpSub_BigDec_100(b *testing.B)   { benchOpSub_BigDec(b, 100) }
func BenchmarkOpSub_BigDec_1000(b *testing.B)  { benchOpSub_BigDec(b, 1000) }
func BenchmarkOpSub_BigDec_10000(b *testing.B) { benchOpSub_BigDec(b, 10000) }

// --- doOpMul BigDec ---

func benchOpMul_BigDec(b *testing.B, digits int) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}
	bv1 := makeBigDec(digits)
	bv2 := makeBigDec(digits)

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: UntypedBigdecType, V: bv1})
		m.PushValue(TypedValue{T: UntypedBigdecType, V: bv2})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpMul()
		bm.SwitchOpCode(bmSetup)
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpMul_BigDec_10(b *testing.B)    { benchOpMul_BigDec(b, 10) }
func BenchmarkOpMul_BigDec_100(b *testing.B)   { benchOpMul_BigDec(b, 100) }
func BenchmarkOpMul_BigDec_1000(b *testing.B)  { benchOpMul_BigDec(b, 1000) }
func BenchmarkOpMul_BigDec_10000(b *testing.B) { benchOpMul_BigDec(b, 10000) }

// --- doOpQuo BigDec ---

func benchOpQuo_BigDec(b *testing.B, digits int) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}
	bv1 := makeBigDec(digits)
	// Divisor: smaller but non-trivial
	divisorDigits := digits / 2
	if divisorDigits < 5 {
		divisorDigits = 5
	}
	bv2 := makeBigDec(divisorDigits)

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: UntypedBigdecType, V: bv1})
		m.PushValue(TypedValue{T: UntypedBigdecType, V: bv2})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpQuo()
		bm.SwitchOpCode(bmSetup)
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpQuo_BigDec_10(b *testing.B)    { benchOpQuo_BigDec(b, 10) }
func BenchmarkOpQuo_BigDec_100(b *testing.B)   { benchOpQuo_BigDec(b, 100) }
func BenchmarkOpQuo_BigDec_1000(b *testing.B)  { benchOpQuo_BigDec(b, 1000) }
func BenchmarkOpQuo_BigDec_10000(b *testing.B) { benchOpQuo_BigDec(b, 10000) }

// --- doOpUneg BigDec ---

func benchOpUneg_BigDec(b *testing.B, digits int) {
	m := benchMachine()
	defer m.Release()
	expr := &UnaryExpr{}
	bv := makeBigDec(digits)

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: UntypedBigdecType, V: bv})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpUneg()
		bm.SwitchOpCode(bmSetup)
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpUneg_BigDec_10(b *testing.B)    { benchOpUneg_BigDec(b, 10) }
func BenchmarkOpUneg_BigDec_100(b *testing.B)   { benchOpUneg_BigDec(b, 100) }
func BenchmarkOpUneg_BigDec_1000(b *testing.B)  { benchOpUneg_BigDec(b, 1000) }
func BenchmarkOpUneg_BigDec_10000(b *testing.B) { benchOpUneg_BigDec(b, 10000) }

// --- doOpInc BigDec ---

func benchOpInc_BigDec(b *testing.B, digits int) {
	m := benchMachine()
	defer m.Release()

	blk, nx := benchBlockVar(m)
	stmt := &IncDecStmt{X: nx}
	bv := makeBigDec(digits)

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		blk.Values[0] = TypedValue{T: UntypedBigdecType, V: bv}
		m.PushStmt(stmt)
		bm.SwitchOpCode(bmTarget)
		m.doOpInc()
		bm.SwitchOpCode(bmSetup)
	}
	reportBenchops(b)
}

func BenchmarkOpInc_BigDec_10(b *testing.B)    { benchOpInc_BigDec(b, 10) }
func BenchmarkOpInc_BigDec_100(b *testing.B)   { benchOpInc_BigDec(b, 100) }
func BenchmarkOpInc_BigDec_1000(b *testing.B)  { benchOpInc_BigDec(b, 1000) }
func BenchmarkOpInc_BigDec_10000(b *testing.B) { benchOpInc_BigDec(b, 10000) }

// --- doOpDec BigDec ---

func benchOpDec_BigDec(b *testing.B, digits int) {
	m := benchMachine()
	defer m.Release()

	blk, nx := benchBlockVar(m)
	stmt := &IncDecStmt{X: nx}
	bv := makeBigDec(digits)

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		blk.Values[0] = TypedValue{T: UntypedBigdecType, V: bv}
		m.PushStmt(stmt)
		bm.SwitchOpCode(bmTarget)
		m.doOpDec()
		bm.SwitchOpCode(bmSetup)
	}
	reportBenchops(b)
}

func BenchmarkOpDec_BigDec_10(b *testing.B)    { benchOpDec_BigDec(b, 10) }
func BenchmarkOpDec_BigDec_100(b *testing.B)   { benchOpDec_BigDec(b, 100) }
func BenchmarkOpDec_BigDec_1000(b *testing.B)  { benchOpDec_BigDec(b, 1000) }
func BenchmarkOpDec_BigDec_10000(b *testing.B) { benchOpDec_BigDec(b, 10000) }

// --- Slice with 3-index expression (low:high:max) ---

func BenchmarkOpSlice_3Index(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	sliceExpr := &SliceExpr{
		Low:  &ConstExpr{},
		High: &ConstExpr{},
		Max:  &ConstExpr{},
	}

	st := &SliceType{Elt: IntType}
	baseArray := m.Alloc.NewListArray(100)
	for i := range 100 {
		baseArray.List[i] = TypedValue{T: IntType, N: i2n(int64(i))}
	}
	sv := m.Alloc.NewSlice(baseArray, 0, 100, 100)

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: st, V: sv})           // base
		m.PushValue(TypedValue{T: IntType, N: i2n(10)}) // low
		m.PushValue(TypedValue{T: IntType, N: i2n(50)}) // high
		m.PushValue(TypedValue{T: IntType, N: i2n(80)}) // max
		m.PushExpr(sliceExpr)
		bm.SwitchOpCode(bmTarget)
		m.doOpSlice()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		rsv := res.V.(*SliceValue)
		if rsv.Length != 40 {
			b.Fatalf("expected len 40, got %d", rsv.Length)
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

// --- Composite lits with string elements (Copy allocates) ---

func BenchmarkOpArrayLit_10_String(b *testing.B) {
	m := benchMachine()
	defer m.Release()

	at := m.Alloc.NewType(&ArrayType{Elt: StringType, Len: 10})
	elts := make([]KeyValueExpr, 10)
	for i := range 10 {
		elts[i] = KeyValueExpr{Value: &ConstExpr{}}
	}
	litExpr := &CompositeLitExpr{Elts: elts}
	strs := make([]StringValue, 10)
	for i := range 10 {
		strs[i] = m.Alloc.NewString(strings.Repeat("x", 20))
	}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(asValue(at))
		for i := range 10 {
			m.PushValue(TypedValue{T: StringType, V: strs[i]})
		}
		m.PushExpr(litExpr)
		bm.SwitchOpCode(bmTarget)
		m.doOpArrayLit()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		av := res.V.(*ArrayValue)
		if len(av.List) != 10 {
			b.Fatalf("expected 10 elements, got %d", len(av.List))
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

// ===========================================================================
// Additional parameterizations for existing benchmarks (audit gap #1-8)
// ===========================================================================

// --- isEql with ArrayKind: recursive O(N) element comparison ---

func benchOpEql_Array(b *testing.B, n int) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}

	at := &ArrayType{Elt: IntType, Len: n}
	av1 := defaultArrayValue(m.Alloc, at)
	av2 := defaultArrayValue(m.Alloc, at)
	for i := range n {
		av1.List[i] = TypedValue{T: IntType, N: i2n(int64(i))}
		av2.List[i] = TypedValue{T: IntType, N: i2n(int64(i))}
	}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: at, V: av1})
		m.PushValue(TypedValue{T: at, V: av2})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpEql()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if !res.GetBool() {
			b.Fatal("expected true")
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpEql_Array_1(b *testing.B)    { benchOpEql_Array(b, 1) }
func BenchmarkOpEql_Array_10(b *testing.B)   { benchOpEql_Array(b, 10) }
func BenchmarkOpEql_Array_100(b *testing.B)  { benchOpEql_Array(b, 100) }
func BenchmarkOpEql_Array_1000(b *testing.B) { benchOpEql_Array(b, 1000) }

// --- isEql with ArrayKind (Data path): bytes.Equal O(n) comparison ---

func benchOpEql_ByteArray(b *testing.B, n int) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}

	at := &ArrayType{Elt: Uint8Type, Len: n}
	av1 := m.Alloc.NewDataArray(n)
	av2 := m.Alloc.NewDataArray(n)
	for i := 0; i < n; i++ {
		av1.Data[i] = byte(i % 256)
		av2.Data[i] = byte(i % 256)
	}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: at, V: av1})
		m.PushValue(TypedValue{T: at, V: av2})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpEql()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if !res.GetBool() {
			b.Fatal("expected true")
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpEql_ByteArray_1(b *testing.B)    { benchOpEql_ByteArray(b, 1) }
func BenchmarkOpEql_ByteArray_10(b *testing.B)   { benchOpEql_ByteArray(b, 10) }
func BenchmarkOpEql_ByteArray_100(b *testing.B)  { benchOpEql_ByteArray(b, 100) }
func BenchmarkOpEql_ByteArray_1000(b *testing.B) { benchOpEql_ByteArray(b, 1000) }

// --- isEql with StructKind: recursive O(fields) comparison ---

func benchOpEql_Struct(b *testing.B, nFields int) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}

	fields := make([]FieldType, nFields)
	for i := range nFields {
		fields[i] = FieldType{
			Name: Name("f" + string(rune('a'+i%26)) + string(rune('0'+i/26))),
			Type: IntType,
		}
	}
	st := &StructType{PkgPath: "bench", Fields: fields}

	fv1 := make([]TypedValue, nFields)
	fv2 := make([]TypedValue, nFields)
	for i := range nFields {
		fv1[i] = TypedValue{T: IntType, N: i2n(int64(i))}
		fv2[i] = TypedValue{T: IntType, N: i2n(int64(i))}
	}
	sv1 := m.Alloc.NewStruct(fv1)
	sv2 := m.Alloc.NewStruct(fv2)

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: st, V: sv1})
		m.PushValue(TypedValue{T: st, V: sv2})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpEql()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if !res.GetBool() {
			b.Fatal("expected true")
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpEql_Struct_1(b *testing.B)    { benchOpEql_Struct(b, 1) }
func BenchmarkOpEql_Struct_10(b *testing.B)   { benchOpEql_Struct(b, 10) }
func BenchmarkOpEql_Struct_100(b *testing.B)  { benchOpEql_Struct(b, 100) }
func BenchmarkOpEql_Struct_1000(b *testing.B) { benchOpEql_Struct(b, 1000) }

// --- Shl BigInt near maxBigintShift limit ---

func BenchmarkOpShl_BigInt_10000(b *testing.B) { benchOpShlBigInt(b, 10000) }

// --- Shr BigInt with large shifts (no limit!) ---

func benchOpShrBigInt(b *testing.B, shift uint64) {
	m := benchMachine()
	defer m.Release()
	expr := &BinaryExpr{}
	v := new(big.Int).Lsh(big.NewInt(1), 10000) // large value
	bv := BigintValue{V: v}
	expected := new(big.Int).Rsh(v, uint(shift))

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: UntypedBigintType, V: bv})
		m.PushValue(TypedValue{T: UintType, N: u2n(shift)})
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpShr()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.V.(BigintValue).V.Cmp(expected) != 0 {
			b.Fatal("unexpected result")
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpShr_BigInt_10(b *testing.B)    { benchOpShrBigInt(b, 10) }
func BenchmarkOpShr_BigInt_100(b *testing.B)   { benchOpShrBigInt(b, 100) }
func BenchmarkOpShr_BigInt_1000(b *testing.B)  { benchOpShrBigInt(b, 1000) }
func BenchmarkOpShr_BigInt_10000(b *testing.B) { benchOpShrBigInt(b, 10000) }

// --- Convert String→[]rune (O(rune_count) per-rune alloc) ---

func benchOpConvert_StringToRunes(b *testing.B, length int) {
	m := benchMachine()
	defer m.Release()

	runeSliceType := &SliceType{Elt: Int32Type}
	sv := m.Alloc.NewString(strings.Repeat("a", length))

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(asValue(runeSliceType))
		m.PushValue(TypedValue{T: StringType, V: sv})
		bm.SwitchOpCode(bmTarget)
		m.doOpConvert()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		rsv := res.V.(*SliceValue)
		if rsv.Length != length {
			b.Fatalf("expected len %d, got %d", length, rsv.Length)
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpConvert_StringToRunes_1(b *testing.B)    { benchOpConvert_StringToRunes(b, 1) }
func BenchmarkOpConvert_StringToRunes_10(b *testing.B)   { benchOpConvert_StringToRunes(b, 10) }
func BenchmarkOpConvert_StringToRunes_100(b *testing.B)  { benchOpConvert_StringToRunes(b, 100) }
func BenchmarkOpConvert_StringToRunes_1000(b *testing.B) { benchOpConvert_StringToRunes(b, 1000) }

// --- SliceLit2 sparse: maxVal amplification ---

func benchOpSliceLit2_Sparse(b *testing.B, maxIdx int) {
	m := benchMachine()
	defer m.Release()

	st := &SliceType{Elt: IntType}
	// Two keyed elements: index 0 and index maxIdx
	elts := []KeyValueExpr{
		{Key: &ConstExpr{}, Value: &ConstExpr{}},
		{Key: &ConstExpr{}, Value: &ConstExpr{}},
	}
	litExpr := &CompositeLitExpr{Elts: elts}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(asValue(st))
		// Push key-value pairs: (0, 1) and (maxIdx, 2)
		m.PushValue(TypedValue{T: IntType, N: i2n(0)})
		m.PushValue(TypedValue{T: IntType, N: i2n(1)})
		m.PushValue(TypedValue{T: IntType, N: i2n(int64(maxIdx))})
		m.PushValue(TypedValue{T: IntType, N: i2n(2)})
		m.PushExpr(litExpr)
		bm.SwitchOpCode(bmTarget)
		m.doOpSliceLit2()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		rsv := res.V.(*SliceValue)
		if rsv.Length != maxIdx+1 {
			b.Fatalf("expected len %d, got %d", maxIdx+1, rsv.Length)
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpSliceLit2_Sparse_10(b *testing.B)    { benchOpSliceLit2_Sparse(b, 9) }
func BenchmarkOpSliceLit2_Sparse_100(b *testing.B)   { benchOpSliceLit2_Sparse(b, 99) }
func BenchmarkOpSliceLit2_Sparse_1000(b *testing.B)  { benchOpSliceLit2_Sparse(b, 999) }
func BenchmarkOpSliceLit2_Sparse_10000(b *testing.B) { benchOpSliceLit2_Sparse(b, 9999) }

// --- ArrayLit uint8: NewDataArray (flat byte alloc) vs non-uint8 ---

func benchOpArrayLit_Uint8(b *testing.B, n int) {
	m := benchMachine()
	defer m.Release()

	at := m.Alloc.NewType(&ArrayType{Elt: Uint8Type, Len: n})
	elts := make([]KeyValueExpr, n)
	for i := range n {
		elts[i] = KeyValueExpr{Value: &ConstExpr{}}
	}
	litExpr := &CompositeLitExpr{Elts: elts}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(asValue(at))
		for i := range n {
			m.PushValue(TypedValue{T: Uint8Type, N: u2n(uint64(i % 256))})
		}
		m.PushExpr(litExpr)
		bm.SwitchOpCode(bmTarget)
		m.doOpArrayLit()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		av := res.V.(*ArrayValue)
		if len(av.Data) != n {
			b.Fatalf("expected data len %d, got %d", n, len(av.Data))
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpArrayLit_Uint8_1(b *testing.B)    { benchOpArrayLit_Uint8(b, 1) }
func BenchmarkOpArrayLit_Uint8_10(b *testing.B)   { benchOpArrayLit_Uint8(b, 10) }
func BenchmarkOpArrayLit_Uint8_100(b *testing.B)  { benchOpArrayLit_Uint8(b, 100) }
func BenchmarkOpArrayLit_Uint8_1000(b *testing.B) { benchOpArrayLit_Uint8(b, 1000) }

// --- doOpEval NameExpr: block depth traversal ---

func benchOpEval_NameExpr(b *testing.B, depth int) {
	m := benchMachine()
	defer m.Release()

	// Build nested block chain: block[0] is outermost, block[depth] is innermost (LastBlock).
	// GetPointerTo loop: for i := 1; i < Depth; i++ { b = b.Parent }
	// So Depth = depth+1 means depth hops from LastBlock to block[0].
	blocks := make([]*Block, depth+1)
	for i := range depth + 1 {
		blocks[i] = &Block{Values: []TypedValue{{T: IntType, N: i2n(int64(i))}}}
		if i > 0 {
			blocks[i].Parent = blocks[i-1]
		}
	}
	// Target var is in blocks[0] (outermost), depth hops from LastBlock.
	blocks[0].Values[0] = TypedValue{T: IntType, N: i2n(99)}
	// Push only the innermost block — GetPointerTo traverses Parent chain.
	m.Blocks = append(m.Blocks, blocks[depth])

	nx := &NameExpr{
		Name: "x",
		Path: ValuePath{
			Type:  VPBlock,
			Depth: uint8(depth + 1), // Depth=1 means current block, +1 per parent hop
			Index: 0,
			Name:  "x",
		},
	}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushExpr(nx)
		bm.SwitchOpCode(bmTarget)
		m.doOpEval()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.GetInt() != 99 {
			b.Fatalf("expected 99, got %d", res.GetInt())
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpEval_NameExpr_Depth1(b *testing.B)   { benchOpEval_NameExpr(b, 1) }
func BenchmarkOpEval_NameExpr_Depth10(b *testing.B)  { benchOpEval_NameExpr(b, 10) }
func BenchmarkOpEval_NameExpr_Depth100(b *testing.B) { benchOpEval_NameExpr(b, 100) }

// --- doOpValueDecl: defaultTypedValue recursion for nested types ---

func benchOpValueDecl_Default(b *testing.B, nt Type) {
	m := benchMachine()
	defer m.Release()

	blk, nxs := benchBlockVars(m, 1)
	nameExpr := NameExpr{
		Name: nxs[0].Name,
		Type: NameExprTypeDefine,
		Path: nxs[0].Path,
	}
	stmt := &ValueDecl{
		NameExprs: []NameExpr{nameExpr},
		Type:      &ConstExpr{}, // non-nil triggers type pop
		Values:    nil,          // nil triggers defaultTypedValue
	}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		blk.Values[0] = TypedValue{}
		m.PushValue(asValue(nt))
		m.PushStmt(stmt)
		bm.SwitchOpCode(bmTarget)
		m.doOpValueDecl()
		bm.SwitchOpCode(bmSetup)
		if blk.Values[0].T == nil {
			b.Fatal("expected non-nil type")
		}
	}
	reportBenchops(b)
}

func BenchmarkOpValueDecl_DefaultInt(b *testing.B) {
	benchOpValueDecl_Default(b, IntType)
}

func BenchmarkOpValueDecl_DefaultArray100(b *testing.B) {
	benchOpValueDecl_Default(b, &ArrayType{Elt: IntType, Len: 100})
}

func BenchmarkOpValueDecl_DefaultStruct10(b *testing.B) {
	fields := make([]FieldType, 10)
	for i := range 10 {
		fields[i] = FieldType{Name: Name("f"), Type: IntType}
	}
	benchOpValueDecl_Default(b, &StructType{PkgPath: "bench", Fields: fields})
}

func BenchmarkOpValueDecl_DefaultArray1000(b *testing.B) {
	benchOpValueDecl_Default(b, &ArrayType{Len: 1000, Elt: IntType})
}

// --- Convert []rune→String (O(rune_count) re-encode) ---

func benchOpConvert_RunesToString(b *testing.B, length int) {
	m := benchMachine()
	defer m.Release()

	// Build a rune slice (list-backed, Int32Kind elements).
	list := make([]TypedValue, length)
	for i := range length {
		list[i] = TypedValue{T: Int32Type, N: i2n(int64('a' + i%26))}
	}
	sliceBase := m.Alloc.NewListArray(length)
	copy(sliceBase.List, list)
	sv := m.Alloc.NewSlice(sliceBase, 0, length, length)
	runeSliceType := &SliceType{Elt: Int32Type}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(asValue(StringType))
		m.PushValue(TypedValue{T: runeSliceType, V: sv})
		bm.SwitchOpCode(bmTarget)
		m.doOpConvert()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if len(res.GetString()) != length {
			b.Fatalf("expected len %d, got %d", length, len(res.GetString()))
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpConvert_RunesToString_1(b *testing.B)    { benchOpConvert_RunesToString(b, 1) }
func BenchmarkOpConvert_RunesToString_10(b *testing.B)   { benchOpConvert_RunesToString(b, 10) }
func BenchmarkOpConvert_RunesToString_100(b *testing.B)  { benchOpConvert_RunesToString(b, 100) }
func BenchmarkOpConvert_RunesToString_1000(b *testing.B) { benchOpConvert_RunesToString(b, 1000) }

// --- doOpEval BasicLitExpr: literal parsing cost ---

func benchOpEval_BasicLitInt(b *testing.B, value string) {
	m := benchMachine()
	defer m.Release()

	litExpr := &BasicLitExpr{Kind: INT, Value: value}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushExpr(litExpr)
		bm.SwitchOpCode(bmTarget)
		m.doOpEval()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.T != UntypedBigintType {
			b.Fatal("expected UntypedBigintType")
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpEval_BasicLitInt_Small(b *testing.B)  { benchOpEval_BasicLitInt(b, "42") }
func BenchmarkOpEval_BasicLitInt_Large(b *testing.B)  { benchOpEval_BasicLitInt(b, strings.Repeat("9", 100)) }
func BenchmarkOpEval_BasicLitInt_Hex(b *testing.B)    { benchOpEval_BasicLitInt(b, "0x"+strings.Repeat("FF", 50)) }

func BenchmarkOpEval_BasicLitString(b *testing.B) {
	m := benchMachine()
	defer m.Release()

	litExpr := &BasicLitExpr{Kind: STRING, Value: `"` + strings.Repeat("x", 100) + `"`}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushExpr(litExpr)
		bm.SwitchOpCode(bmTarget)
		m.doOpEval()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if len(res.GetString()) != 100 {
			b.Fatalf("expected len 100, got %d", len(res.GetString()))
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

// --- doOpTypeAssert1 interface: VerifyImplementedBy with many methods ---

// benchInterfaceAndImpl creates an InterfaceType with nMethods methods and a
// DeclaredType that implements nImpl of them. When nImpl == nMethods the type
// fully implements the interface.
func benchInterfaceAndImpl(alloc *Allocator, nMethods, nImpl int) (*InterfaceType, *DeclaredType, *StructValue) {
	methods := make([]FieldType, nMethods)
	for i := range nMethods {
		methods[i] = FieldType{
			Name: Name(fmt.Sprintf("M%d", i)),
			Type: &FuncType{Params: []FieldType{}, Results: []FieldType{}},
		}
	}
	iface := &InterfaceType{
		PkgPath: "bench",
		Methods: methods,
	}
	st := &StructType{PkgPath: "bench", Fields: []FieldType{}}
	dt := &DeclaredType{
		PkgPath: "bench",
		Name:    "S",
		Base:    st,
		Methods: make([]TypedValue, nImpl),
	}
	for i := range nImpl {
		ft := &FuncType{
			Params:  []FieldType{{Name: "self", Type: dt}},
			Results: []FieldType{},
		}
		fv := &FuncValue{
			Type:     ft,
			IsMethod: true,
			Source:   &FuncDecl{},
			Name:     methods[i].Name,
			PkgPath:  "bench",
			body:     []Stmt{},
		}
		dt.Methods[i] = TypedValue{T: ft, V: fv}
	}
	sv := alloc.NewStruct([]TypedValue{})
	return iface, dt, sv
}

func benchOpTypeAssert1_Interface(b *testing.B, nMethods int) {
	m := benchMachine()
	defer m.Release()
	expr := &TypeAssertExpr{}

	iface, dt, sv := benchInterfaceAndImpl(m.Alloc, nMethods, nMethods)

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: dt, V: sv})
		m.PushValue(asValue(iface))
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpTypeAssert1()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.T != dt {
			b.Fatal("expected declared type")
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpTypeAssert1_Interface_1(b *testing.B)    { benchOpTypeAssert1_Interface(b, 1) }
func BenchmarkOpTypeAssert1_Interface_10(b *testing.B)   { benchOpTypeAssert1_Interface(b, 10) }
func BenchmarkOpTypeAssert1_Interface_100(b *testing.B)  { benchOpTypeAssert1_Interface(b, 100) }

// --- doOpSelector VPInterface: interface method dispatch via findEmbeddedFieldType ---
// Cost is O(nMethods) due to method matching.

func benchOpSelector_VPInterface(b *testing.B, nMethods int) {
	m := benchMachine()
	defer m.Release()

	_, dt, sv := benchInterfaceAndImpl(m.Alloc, nMethods, nMethods)

	// Target the last method to maximize search cost.
	lastName := Name(fmt.Sprintf("M%d", nMethods-1))
	selExpr := &SelectorExpr{
		Path: ValuePath{
			Type:  VPInterface,
			Depth: 0,
			Index: 0,
			Name:  lastName,
		},
	}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: dt, V: sv})
		m.PushExpr(selExpr)
		bm.SwitchOpCode(bmTarget)
		m.doOpSelector()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if _, ok := res.V.(*BoundMethodValue); !ok {
			b.Fatal("expected BoundMethodValue")
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpSelector_VPInterface_1(b *testing.B)   { benchOpSelector_VPInterface(b, 1) }
func BenchmarkOpSelector_VPInterface_10(b *testing.B)  { benchOpSelector_VPInterface(b, 10) }
func BenchmarkOpSelector_VPInterface_100(b *testing.B) { benchOpSelector_VPInterface(b, 100) }

// --- doOpSelector VPValMethod: BoundMethodValue allocation ---

func BenchmarkOpSelector_VPValMethod(b *testing.B) {
	m := benchMachine()
	defer m.Release()

	// Create a DeclaredType with a value method.
	st := &StructType{
		PkgPath: "bench",
		Fields: []FieldType{
			{Name: "x", Type: IntType},
		},
	}
	dt := &DeclaredType{
		PkgPath: "bench",
		Name:    "MyStruct",
		Base:    st,
	}
	ft := &FuncType{
		Params: []FieldType{
			{Name: "self", Type: dt},
		},
		Results: []FieldType{{Type: IntType}},
	}
	fv := &FuncValue{
		Type:     ft,
		IsMethod: true,
		Source:   &FuncDecl{},
		Name:     "GetX",
		PkgPath:  "bench",
		body:     []Stmt{},
	}
	dt.Methods = []TypedValue{{T: ft, V: fv}}

	fieldValues := []TypedValue{{T: IntType, N: i2n(42)}}
	sv := m.Alloc.NewStruct(fieldValues)

	selExpr := &SelectorExpr{
		Path: ValuePath{
			Type:  VPValMethod,
			Depth: 0,
			Index: 0,
			Name:  "GetX",
		},
	}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: dt, V: sv})
		m.PushExpr(selExpr)
		bm.SwitchOpCode(bmTarget)
		m.doOpSelector()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if _, ok := res.V.(*BoundMethodValue); !ok {
			b.Fatal("expected BoundMethodValue")
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

// --- doOpFuncLit: closure capture cost O(nCaptures × depth) ---
// Each capture requires GetPointerToDirect which traverses O(Depth) parent blocks.

func benchOpFuncLit(b *testing.B, nCaptures int) {
	m := benchMachine()
	defer m.Release()

	// Build a block with HeapItemValues for capture.
	values := make([]TypedValue, nCaptures)
	for i := range nCaptures {
		values[i] = TypedValue{
			T: heapItemType{},
			V: m.Alloc.NewHeapItem(TypedValue{T: IntType, N: i2n(int64(i))}),
		}
	}
	blk := &Block{Values: values}
	m.Blocks = append(m.Blocks, blk)

	// Build HeapCaptures with Depth=1 (current block, 0 hops).
	captures := make(NameExprs, nCaptures)
	for i := range nCaptures {
		captures[i] = NameExpr{
			Path: ValuePath{
				Type:  VPBlock,
				Depth: 1,
				Index: uint16(i),
				Name:  Name("c"),
			},
		}
	}

	ft := &FuncType{Params: []FieldType{}, Results: []FieldType{}}
	flit := &FuncLitExpr{
		HeapCaptures: captures,
		Body:         []Stmt{},
	}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{V: TypeValue{Type: ft}})
		m.PushExpr(flit)
		bm.SwitchOpCode(bmTarget)
		m.doOpFuncLit()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if _, ok := res.V.(*FuncValue); !ok {
			b.Fatal("expected FuncValue")
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpFuncLit_Captures_0(b *testing.B)    { benchOpFuncLit(b, 0) }
func BenchmarkOpFuncLit_Captures_1(b *testing.B)    { benchOpFuncLit(b, 1) }
func BenchmarkOpFuncLit_Captures_10(b *testing.B)   { benchOpFuncLit(b, 10) }
func BenchmarkOpFuncLit_Captures_100(b *testing.B)  { benchOpFuncLit(b, 100) }
func BenchmarkOpFuncLit_Captures_1000(b *testing.B) { benchOpFuncLit(b, 1000) }

// --- doOpCall: block alloc + captures copy + param assignment ---
// benchFuncDeclNode creates a minimal FuncDecl usable as BlockNode Source.
func benchFuncDeclNode(numNames int, heapIdxs []int) *FuncDecl {
	fd := &FuncDecl{}
	fd.StaticBlock.NumNames = uint16(numNames)
	fd.StaticBlock.Names = make([]Name, numNames)
	fd.StaticBlock.HeapItems = make([]bool, numNames)
	for _, idx := range heapIdxs {
		fd.StaticBlock.HeapItems[idx] = true
	}
	fd.StaticBlock.Block.Source = fd
	fd.Body = []Stmt{} // empty body
	return fd
}

func benchOpCall(b *testing.B, nParams int, nCaptures int) {
	m := benchMachine()
	defer m.Release()

	// Build FuncType with nParams int params.
	params := make([]FieldType, nParams)
	for i := range nParams {
		params[i] = FieldType{Name: Name("p"), Type: IntType}
	}
	ft := &FuncType{Params: params, Results: []FieldType{}}

	// FuncDecl as source with slots for params + captures.
	numNames := nParams + nCaptures
	fd := benchFuncDeclNode(numNames, nil)

	// Build captures (HeapItemValues).
	captures := make([]TypedValue, nCaptures)
	for i := range nCaptures {
		captures[i] = TypedValue{
			T: heapItemType{},
			V: m.Alloc.NewHeapItem(TypedValue{T: IntType, N: i2n(int64(i))}),
		}
	}

	fv := &FuncValue{
		Type:      ft,
		IsClosure: true, // GetParent returns nil (avoids store lookup)
		Source:    fd,
		Captures:  captures,
		PkgPath:   "bench",
		body:      []Stmt{},
	}

	cx := &CallExpr{NumArgs: nParams}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		// Push args + func value for PushFrameCall's numValues calc.
		for i := range nParams {
			m.PushValue(TypedValue{T: IntType, N: i2n(int64(i))})
		}
		m.PushValue(TypedValue{T: ft, V: fv})
		m.PushFrameCall(cx, fv, TypedValue{}, false)
		bm.SwitchOpCode(bmTarget)
		m.doOpCall()
		bm.SwitchOpCode(bmSetup)
		// doOpCall pushes block + ops + stmts; clean up.
		m.Blocks = m.Blocks[:0]
		m.Ops = m.Ops[:0]
		m.Stmts = m.Stmts[:0]
		m.Values = m.Values[:0]
		m.Frames = m.Frames[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpCall_0Params_0Captures(b *testing.B)      { benchOpCall(b, 0, 0) }
func BenchmarkOpCall_1Params_0Captures(b *testing.B)      { benchOpCall(b, 1, 0) }
func BenchmarkOpCall_10Params_0Captures(b *testing.B)     { benchOpCall(b, 10, 0) }
func BenchmarkOpCall_100Params_0Captures(b *testing.B)    { benchOpCall(b, 100, 0) }
func BenchmarkOpCall_1000Params_0Captures(b *testing.B)   { benchOpCall(b, 1000, 0) }
func BenchmarkOpCall_0Params_1Captures(b *testing.B)      { benchOpCall(b, 0, 1) }
func BenchmarkOpCall_0Params_10Captures(b *testing.B)     { benchOpCall(b, 0, 10) }
func BenchmarkOpCall_0Params_100Captures(b *testing.B)    { benchOpCall(b, 0, 100) }
func BenchmarkOpCall_0Params_1000Captures(b *testing.B)   { benchOpCall(b, 0, 1000) }
func BenchmarkOpCall_10Params_10Captures(b *testing.B)    { benchOpCall(b, 10, 10) }

// --- doOpReturn: unwind stack + realm check ---

func BenchmarkOpReturn(b *testing.B) {
	m := benchMachine()
	defer m.Release()

	ft := &FuncType{Params: []FieldType{}, Results: []FieldType{{Type: IntType}}}
	fd := benchFuncDeclNode(0, nil)
	fv := &FuncValue{
		Type:      ft,
		IsClosure: true,
		Source:    fd,
		PkgPath:   "bench",
		body:      []Stmt{},
	}
	cx := &CallExpr{NumArgs: 0}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		// Set up a call frame as doOpReturn expects.
		m.PushValue(TypedValue{T: ft, V: fv}) // func value for PushFrameCall
		m.PushFrameCall(cx, fv, TypedValue{}, false)
		// Push a block (doOpReturn pops to frame's NumBlocks).
		m.Blocks = append(m.Blocks, &Block{})
		// Push result value.
		m.PushValue(TypedValue{T: IntType, N: i2n(42)})
		bm.SwitchOpCode(bmTarget)
		m.doOpReturn()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.GetInt() != 42 {
			b.Fatalf("expected 42, got %d", res.GetInt())
		}
		m.Values = m.Values[:0]
		m.Blocks = m.Blocks[:0]
		m.Frames = m.Frames[:0]
	}
	reportBenchops(b)
}

// --- doOpDefer: pop args + store in frame's Defers ---

func benchOpDefer(b *testing.B, nArgs int) {
	m := benchMachine()
	defer m.Release()

	// Build func type with nArgs params.
	params := make([]FieldType, nArgs)
	for i := range nArgs {
		params[i] = FieldType{Name: Name("a"), Type: IntType}
	}
	ft := &FuncType{Params: params, Results: []FieldType{}}
	fd := benchFuncDeclNode(nArgs, nil)
	fv := &FuncValue{
		Type:      ft,
		IsClosure: true,
		Source:    fd,
		PkgPath:   "bench",
		body:      []Stmt{},
	}

	// Need a call frame for MustPeekCallFrame(1).
	outerFt := &FuncType{Params: []FieldType{}, Results: []FieldType{}}
	outerFv := &FuncValue{
		Type:      outerFt,
		IsClosure: true,
		Source:    benchFuncDeclNode(0, nil),
		PkgPath:   "bench",
		body:      []Stmt{},
	}
	outerCx := &CallExpr{NumArgs: 0}
	m.PushValue(TypedValue{T: outerFt, V: outerFv})
	m.PushFrameCall(outerCx, outerFv, TypedValue{}, false)
	m.Blocks = append(m.Blocks, &Block{}) // block for outer frame

	ds := &DeferStmt{
		Call: CallExpr{
			NumArgs: nArgs,
			Args:    make([]Expr, nArgs),
		},
	}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		// Push args + func value.
		for i := range nArgs {
			m.PushValue(TypedValue{T: IntType, N: i2n(int64(i))})
		}
		m.PushValue(TypedValue{T: ft, V: fv})
		m.PushStmt(ds)
		bm.SwitchOpCode(bmTarget)
		m.doOpDefer()
		bm.SwitchOpCode(bmSetup)
		// Reset defers for next iteration.
		m.LastFrame().Defers = m.LastFrame().Defers[:0]
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpDefer_1Arg(b *testing.B)    { benchOpDefer(b, 1) }
func BenchmarkOpDefer_10Args(b *testing.B)  { benchOpDefer(b, 10) }
func BenchmarkOpDefer_100Args(b *testing.B) { benchOpDefer(b, 100) }

// --- OpForLoop: heap item copy at end of iteration ---
// Benchmarks the cost of copying HeapItemValues at the end of each loop
// iteration (Go 1.22 loopvars semantics).

func benchOpForLoopHeapCopy(b *testing.B, numInit int) {
	m := benchMachine()
	defer m.Release()

	// Build block with HeapItemValues in Values[0..numInit-1].
	values := make([]TypedValue, numInit)
	for i := range numInit {
		values[i] = TypedValue{
			T: heapItemType{},
			V: m.Alloc.NewHeapItem(TypedValue{T: IntType, N: i2n(int64(i))}),
		}
	}
	blk := &Block{Values: values}
	// Set bodyStmt to end-of-body state: NextBodyIndex == BodyLen.
	// With Cond=nil and Post=nil, the loop returns after heap copy.
	blk.bodyStmt = bodyStmt{
		Body:          []Stmt{},
		BodyLen:       0,
		NextBodyIndex: 0, // == BodyLen, triggers heap copy
		NumInit:       numInit,
		Cond:          nil,
		Post:          nil,
		NumOps:        0,
		NumValues:     0,
		NumExprs:      0,
		NumStmts:      0,
	}
	m.Blocks = append(m.Blocks, blk)

	// doOpExec(OpForLoop) needs PeekStmt(1) — push a dummy stmt.
	dummyStmt := blk.GetBodyStmt()
	m.PushStmt(dummyStmt)

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		// Reset block state to end-of-body.
		blk.bodyStmt.NextBodyIndex = 0 // == BodyLen (0)
		// Restore HeapItemValues (doOpExec replaces them).
		for i := range numInit {
			blk.Values[i].V = m.Alloc.NewHeapItem(TypedValue{T: IntType, N: i2n(int64(i))})
		}
		bm.SwitchOpCode(bmTarget)
		m.doOpExec(OpForLoop)
		bm.SwitchOpCode(bmSetup)
	}
	reportBenchops(b)
}

func BenchmarkOpForLoop_HeapCopy_0(b *testing.B)    { benchOpForLoopHeapCopy(b, 0) }
func BenchmarkOpForLoop_HeapCopy_1(b *testing.B)    { benchOpForLoopHeapCopy(b, 1) }
func BenchmarkOpForLoop_HeapCopy_10(b *testing.B)   { benchOpForLoopHeapCopy(b, 10) }
func BenchmarkOpForLoop_HeapCopy_100(b *testing.B)  { benchOpForLoopHeapCopy(b, 100) }
func BenchmarkOpForLoop_HeapCopy_1000(b *testing.B) { benchOpForLoopHeapCopy(b, 1000) }

// --- doOpIfCond: condition check + ExpandWith ---

func BenchmarkOpIfCond_TrueBranch(b *testing.B) {
	m := benchMachine()
	defer m.Release()

	// Create an IfCaseStmt (the Then branch) with a small body.
	thenCase := &IfCaseStmt{
		Body: []Stmt{&EmptyStmt{}},
	}
	thenCase.StaticBlock.NumNames = 0
	thenCase.StaticBlock.HeapItems = []bool{}
	thenCase.StaticBlock.Block.Source = thenCase

	ifStmt := &IfStmt{
		Then: *thenCase,
		Else: IfCaseStmt{Body: []Stmt{}},
	}

	// Need a block on the stack for ExpandWith.
	blk := &Block{Values: []TypedValue{}}
	m.Blocks = append(m.Blocks, blk)

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: BoolType, N: i2n(1)}) // true
		m.PushStmt(ifStmt)
		bm.SwitchOpCode(bmTarget)
		m.doOpIfCond()
		bm.SwitchOpCode(bmSetup)
		// doOpIfCond pushes OpBody + bodyStmt; clean up.
		m.Ops = m.Ops[:0]
		m.Stmts = m.Stmts[:0]
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

// --- doOpBinary1: LAND/LOR short-circuit dispatch ---

func BenchmarkOpBinary1_LAND_True(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	bx := &BinaryExpr{
		Op:    LAND,
		Right: &ConstExpr{TypedValue: TypedValue{T: BoolType, N: i2n(1)}},
	}
	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: BoolType, N: i2n(1)}) // true → must eval RHS
		m.PushExpr(bx)
		bm.SwitchOpCode(bmTarget)
		m.doOpBinary1()
		bm.SwitchOpCode(bmSetup)
		m.Ops = m.Ops[:0]
		m.Exprs = m.Exprs[:0]
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpBinary1_LAND_False(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	bx := &BinaryExpr{
		Op:    LAND,
		Right: &ConstExpr{TypedValue: TypedValue{T: BoolType, N: i2n(1)}},
	}
	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: BoolType, N: i2n(0)}) // false → short circuit
		m.PushExpr(bx)
		bm.SwitchOpCode(bmTarget)
		m.doOpBinary1()
		bm.SwitchOpCode(bmSetup)
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

// --- doOpRef: PopAsPointer2 + PointerType allocation ---

func BenchmarkOpRef(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	blk, nx := benchBlockVar(m)
	blk.Values[0] = TypedValue{T: IntType, N: i2n(42)}
	rx := &RefExpr{X: nx}
	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushExpr(rx)
		bm.SwitchOpCode(bmTarget)
		m.doOpRef()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if _, ok := res.T.(*PointerType); !ok {
			b.Fatal("expected PointerType")
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

// --- doOpCompositeLit: dispatch to sub-ops ---

func BenchmarkOpCompositeLit_Array(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	at := &ArrayType{Len: 3, Elt: IntType}
	elts := []KeyValueExpr{
		{Value: &ConstExpr{TypedValue: TypedValue{T: IntType, N: i2n(1)}}},
		{Value: &ConstExpr{TypedValue: TypedValue{T: IntType, N: i2n(2)}}},
		{Value: &ConstExpr{TypedValue: TypedValue{T: IntType, N: i2n(3)}}},
	}
	cle := &CompositeLitExpr{Elts: elts}
	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{V: TypeValue{Type: at}})
		m.PushExpr(cle)
		bm.SwitchOpCode(bmTarget)
		m.doOpCompositeLit()
		bm.SwitchOpCode(bmSetup)
		m.Ops = m.Ops[:0]
		m.Exprs = m.Exprs[:0]
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

// --- doOpTypeDecl: assign type to block ---

func BenchmarkOpTypeDecl(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	blk, _ := benchBlockVar(m)
	blk.Values[0] = TypedValue{}
	td := &TypeDecl{
		NameExpr: NameExpr{
			Name: "T",
			Path: ValuePath{Type: VPBlock, Depth: 1, Index: 0, Name: "T"},
		},
	}
	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(asValue(IntType))
		m.PushStmt(td)
		bm.SwitchOpCode(bmTarget)
		m.doOpTypeDecl()
		bm.SwitchOpCode(bmSetup)
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

// --- doOpReturnAfterCopy: copies results to block then returns ---

func BenchmarkOpReturnAfterCopy(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	ft := &FuncType{
		Params:  []FieldType{{Name: "x", Type: IntType}},
		Results: []FieldType{{Name: "r", Type: IntType}},
	}
	fd := benchFuncDeclNode(2, nil)
	fv := &FuncValue{
		Type:      ft,
		IsClosure: true,
		Source:    fd,
		PkgPath:   "bench",
		body:      []Stmt{},
	}
	cx := &CallExpr{NumArgs: 1}
	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: IntType, N: i2n(1)})
		m.PushValue(TypedValue{T: ft, V: fv})
		m.PushFrameCall(cx, fv, TypedValue{}, false)
		fblk := &Block{Values: []TypedValue{
			{T: IntType, N: i2n(1)}, // param
			{T: IntType, N: i2n(0)}, // result slot
		}}
		m.Blocks = append(m.Blocks, fblk)
		m.PushValue(TypedValue{T: IntType, N: i2n(99)}) // result
		bm.SwitchOpCode(bmTarget)
		m.doOpReturnAfterCopy()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.GetInt() != 99 {
			b.Fatalf("expected 99, got %d", res.GetInt())
		}
		m.Values = m.Values[:0]
		m.Blocks = m.Blocks[:0]
		m.Frames = m.Frames[:0]
	}
	reportBenchops(b)
}

// --- doOpPrecall: function value dispatch ---

func BenchmarkOpPrecall_FuncValue(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	ft := &FuncType{Params: []FieldType{{Name: "x", Type: IntType}}, Results: []FieldType{}}
	fd := benchFuncDeclNode(1, nil)
	fv := &FuncValue{
		Type:      ft,
		IsClosure: true,
		Source:    fd,
		PkgPath:   "bench",
		body:      []Stmt{},
	}
	cx := &CallExpr{NumArgs: 1}
	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: ft, V: fv})           // func
		m.PushValue(TypedValue{T: IntType, N: i2n(1)})  // arg
		m.PushExpr(cx)
		bm.SwitchOpCode(bmTarget)
		m.doOpPrecall()
		bm.SwitchOpCode(bmSetup)
		m.Ops = m.Ops[:0]
		m.Frames = m.Frames[:0]
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpPrecall_TypeConversion(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	cx := &CallExpr{NumArgs: 1, Args: []Expr{&ConstExpr{}}}
	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{V: TypeValue{Type: Int64Type}}) // type
		m.PushValue(TypedValue{T: IntType, N: i2n(42)})       // arg
		m.PushExpr(cx)
		bm.SwitchOpCode(bmTarget)
		m.doOpPrecall()
		bm.SwitchOpCode(bmSetup)
		m.Ops = m.Ops[:0]
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

// benchMethodSetup creates a DeclaredType with one value method "DoStuff(self S, y int)".
// Returns the FuncType, FuncValue, DeclaredType, and a StructValue receiver.
func benchMethodSetup(alloc *Allocator) (ft *FuncType, fv *FuncValue, dt *DeclaredType, sv *StructValue) {
	st := &StructType{
		PkgPath: "bench",
		Fields:  []FieldType{{Name: "x", Type: IntType}},
	}
	dt = &DeclaredType{
		PkgPath: "bench",
		Name:    "S",
		Base:    st,
	}
	ft = &FuncType{
		Params:  []FieldType{{Name: "self", Type: dt}, {Name: "y", Type: IntType}},
		Results: []FieldType{},
	}
	fd := benchFuncDeclNode(2, nil) // self + y
	fv = &FuncValue{
		Type:      ft,
		IsMethod:  true,
		IsClosure: true,
		Source:    fd,
		Name:      "DoStuff",
		PkgPath:   "bench",
		body:      []Stmt{},
	}
	dt.Methods = []TypedValue{{T: ft, V: fv}}
	sv = alloc.NewStruct([]TypedValue{{T: IntType, N: i2n(42)}})
	return
}

// --- doOpPrecall BoundMethodValue: method call dispatch ---

func BenchmarkOpPrecall_BoundMethod(b *testing.B) {
	m := benchMachine()
	defer m.Release()

	ft, fv, dt, sv := benchMethodSetup(m.Alloc)
	bmv := &BoundMethodValue{
		Func:     fv,
		Receiver: TypedValue{T: dt, V: sv},
	}
	cx := &CallExpr{NumArgs: 1}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: ft, V: bmv})            // bound method
		m.PushValue(TypedValue{T: IntType, N: i2n(1)})    // arg
		m.PushExpr(cx)
		bm.SwitchOpCode(bmTarget)
		m.doOpPrecall()
		bm.SwitchOpCode(bmSetup)
		m.Ops = m.Ops[:0]
		m.Frames = m.Frames[:0]
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

// --- doOpCall with receiver (method call) ---

func BenchmarkOpCall_Method(b *testing.B) {
	m := benchMachine()
	defer m.Release()

	ft, fv, dt, sv := benchMethodSetup(m.Alloc)
	recv := TypedValue{T: dt, V: sv}
	cx := &CallExpr{NumArgs: 1}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: IntType, N: i2n(7)}) // arg
		m.PushValue(TypedValue{T: ft, V: fv})          // func
		m.PushFrameCall(cx, fv, recv, false)
		bm.SwitchOpCode(bmTarget)
		m.doOpCall()
		bm.SwitchOpCode(bmSetup)
		m.Blocks = m.Blocks[:0]
		m.Ops = m.Ops[:0]
		m.Stmts = m.Stmts[:0]
		m.Values = m.Values[:0]
		m.Frames = m.Frames[:0]
	}
	reportBenchops(b)
}

// --- doOpIfCond false branch ---

func BenchmarkOpIfCond_FalseBranch(b *testing.B) {
	m := benchMachine()
	defer m.Release()

	elseCase := &IfCaseStmt{
		Body: []Stmt{&EmptyStmt{}},
	}
	elseCase.StaticBlock.NumNames = 0
	elseCase.StaticBlock.HeapItems = []bool{}
	elseCase.StaticBlock.Block.Source = elseCase

	thenCase := &IfCaseStmt{Body: []Stmt{&EmptyStmt{}}}
	thenCase.StaticBlock.NumNames = 0
	thenCase.StaticBlock.HeapItems = []bool{}
	thenCase.StaticBlock.Block.Source = thenCase

	ifStmt := &IfStmt{
		Then: *thenCase,
		Else: *elseCase,
	}

	blk := &Block{Values: []TypedValue{}}
	m.Blocks = append(m.Blocks, blk)

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: BoolType, N: i2n(0)}) // false
		m.PushStmt(ifStmt)
		bm.SwitchOpCode(bmTarget)
		m.doOpIfCond()
		bm.SwitchOpCode(bmSetup)
		m.Ops = m.Ops[:0]
		m.Stmts = m.Stmts[:0]
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

// --- doOpTypeAssert2 interface: parameterized by nMethods ---

func benchOpTypeAssert2_Interface(b *testing.B, nMethods int, shouldMatch bool) {
	m := benchMachine()
	defer m.Release()
	expr := &TypeAssertExpr{}

	nImpl := nMethods
	if !shouldMatch {
		nImpl = nMethods - 1
		if nImpl < 0 {
			nImpl = 0
		}
	}
	iface, dt, sv := benchInterfaceAndImpl(m.Alloc, nMethods, nImpl)

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: dt, V: sv})
		m.PushValue(asValue(iface))
		m.PushExpr(expr)
		bm.SwitchOpCode(bmTarget)
		m.doOpTypeAssert2()
		bm.SwitchOpCode(bmSetup)
		boolRes := m.PeekValue(1)
		if shouldMatch && !boolRes.GetBool() {
			b.Fatal("expected ok=true")
		}
		if !shouldMatch && boolRes.GetBool() {
			b.Fatal("expected ok=false")
		}
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpTypeAssert2_Interface_Hit_1(b *testing.B)   { benchOpTypeAssert2_Interface(b, 1, true) }
func BenchmarkOpTypeAssert2_Interface_Hit_10(b *testing.B)  { benchOpTypeAssert2_Interface(b, 10, true) }
func BenchmarkOpTypeAssert2_Interface_Hit_100(b *testing.B) { benchOpTypeAssert2_Interface(b, 100, true) }
func BenchmarkOpTypeAssert2_Interface_Miss_10(b *testing.B) { benchOpTypeAssert2_Interface(b, 10, false) }

// --- doOpReturnFromBlock: reads named results from block ---

func BenchmarkOpReturnFromBlock(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	ft := &FuncType{
		Params:  []FieldType{{Name: "x", Type: IntType}},
		Results: []FieldType{{Name: "r", Type: IntType}},
	}
	fd := benchFuncDeclNode(2, nil)
	fv := &FuncValue{
		Type:      ft,
		IsClosure: true,
		Source:    fd,
		PkgPath:   "bench",
		body:      []Stmt{},
	}
	cx := &CallExpr{NumArgs: 1}
	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: IntType, N: i2n(1)})
		m.PushValue(TypedValue{T: ft, V: fv})
		m.PushFrameCall(cx, fv, TypedValue{}, false)
		fblk := &Block{Values: []TypedValue{
			{T: IntType, N: i2n(1)},  // param
			{T: IntType, N: i2n(77)}, // named result
		}}
		m.Blocks = append(m.Blocks, fblk)
		bm.SwitchOpCode(bmTarget)
		m.doOpReturnFromBlock()
		bm.SwitchOpCode(bmSetup)
		res := m.PeekValue(1)
		if res.GetInt() != 77 {
			b.Fatalf("expected 77, got %d", res.GetInt())
		}
		m.Values = m.Values[:0]
		m.Blocks = m.Blocks[:0]
		m.Frames = m.Frames[:0]
	}
	reportBenchops(b)
}

// --- doOpReturnToBlock: assigns results back to function block ---

func BenchmarkOpReturnToBlock(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	ft := &FuncType{
		Params:  []FieldType{{Name: "x", Type: IntType}},
		Results: []FieldType{{Name: "r", Type: IntType}},
	}
	fd := benchFuncDeclNode(2, nil)
	fv := &FuncValue{
		Type:      ft,
		IsClosure: true,
		Source:    fd,
		PkgPath:   "bench",
		body:      []Stmt{},
	}
	cx := &CallExpr{NumArgs: 1}
	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: IntType, N: i2n(1)})
		m.PushValue(TypedValue{T: ft, V: fv})
		m.PushFrameCall(cx, fv, TypedValue{}, false)
		fblk := &Block{Values: []TypedValue{
			{T: IntType, N: i2n(1)}, // param
			{T: IntType, N: i2n(0)}, // result slot
		}}
		m.Blocks = append(m.Blocks, fblk)
		m.PushValue(TypedValue{T: IntType, N: i2n(55)}) // result on stack
		bm.SwitchOpCode(bmTarget)
		m.doOpReturnToBlock()
		bm.SwitchOpCode(bmSetup)
		if fblk.Values[1].GetInt() != 55 {
			b.Fatalf("expected 55, got %d", fblk.Values[1].GetInt())
		}
		m.Values = m.Values[:0]
		m.Blocks = m.Blocks[:0]
		m.Frames = m.Frames[:0]
	}
	reportBenchops(b)
}

// --- doOpReturnCallDefers: processes defer chain ---

func benchOpReturnCallDefers(b *testing.B, nDefers int) {
	m := benchMachine()
	defer m.Release()
	// Outer function with defers.
	ft := &FuncType{Params: []FieldType{}, Results: []FieldType{}}
	fd := benchFuncDeclNode(0, nil)
	fv := &FuncValue{
		Type:      ft,
		IsClosure: true,
		Source:    fd,
		PkgPath:   "bench",
		body:      []Stmt{},
	}
	cx := &CallExpr{NumArgs: 0}
	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: ft, V: fv})
		m.PushFrameCall(cx, fv, TypedValue{}, false)
		m.Blocks = append(m.Blocks, &Block{})
		// Add defers.
		cfr := m.LastFrame()
		for range nDefers {
			cfr.PushDefer(Defer{
				Func:   fv,
				Args:   []TypedValue{},
				Source: &DeferStmt{Call: CallExpr{NumArgs: 0, Args: []Expr{}}},
				Parent: &Block{},
			})
		}
		m.PushOp(OpReturnCallDefers) // will be consumed by the op
		bm.SwitchOpCode(bmTarget)
		m.doOpReturnCallDefers()
		bm.SwitchOpCode(bmSetup)
		// doOpReturnCallDefers pops one defer and sets up the call.
		m.Ops = m.Ops[:0]
		m.Stmts = m.Stmts[:0]
		m.Blocks = m.Blocks[:0]
		m.Frames = m.Frames[:0]
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpReturnCallDefers_1(b *testing.B)   { benchOpReturnCallDefers(b, 1) }
func BenchmarkOpReturnCallDefers_10(b *testing.B)  { benchOpReturnCallDefers(b, 10) }
func BenchmarkOpReturnCallDefers_100(b *testing.B)  { benchOpReturnCallDefers(b, 100) }
func BenchmarkOpReturnCallDefers_1000(b *testing.B) { benchOpReturnCallDefers(b, 1000) }

// --- doOpPanic2: unwind to call frame ---

func BenchmarkOpPanic2(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	ft := &FuncType{Params: []FieldType{}, Results: []FieldType{}}
	fd := benchFuncDeclNode(0, nil)
	fv := &FuncValue{
		Type:      ft,
		IsClosure: true,
		Source:    fd,
		PkgPath:   "bench",
		body:      []Stmt{},
	}
	cx := &CallExpr{NumArgs: 0}
	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: ft, V: fv})
		m.PushFrameCall(cx, fv, TypedValue{}, false)
		m.Exception = &Exception{
			Value: TypedValue{T: StringType, V: m.Alloc.NewString("panic")},
		}
		bm.SwitchOpCode(bmTarget)
		m.doOpPanic2()
		bm.SwitchOpCode(bmSetup)
		// Pushes OpReturnCallDefers.
		m.Ops = m.Ops[:0]
		m.Frames = m.Frames[:0]
		m.Values = m.Values[:0]
		m.Exception = nil
	}
	reportBenchops(b)
}

// --- doOpExec OpBody: bodyStmt state machine ---

func BenchmarkOpBody(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	// Create a block with a bodyStmt containing a small body.
	blk := &Block{}
	blk.bodyStmt = bodyStmt{
		Body:          []Stmt{&EmptyStmt{}, &EmptyStmt{}, &EmptyStmt{}},
		BodyLen:       3,
		NextBodyIndex: -2,
	}
	m.Blocks = append(m.Blocks, blk)
	bs := blk.GetBodyStmt()
	m.PushStmt(bs)
	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		blk.bodyStmt.NextBodyIndex = -2
		bm.SwitchOpCode(bmTarget)
		m.doOpExec(OpBody) // processes init + dispatches first stmt
		bm.SwitchOpCode(bmSetup)
		// Dispatched to EXEC_SWITCH for first EmptyStmt.
		m.Ops = m.Ops[:0]
		m.Stmts = m.Stmts[:0]
		m.PushStmt(bs) // re-push for next iter
	}
	reportBenchops(b)
}

// --- OpRangeIter: array copy + iteration init ---

func benchOpRangeIter(b *testing.B, n int) {
	m := benchMachine()
	defer m.Release()
	// Build array value.
	elems := make([]TypedValue, n)
	for i := range n {
		elems[i] = TypedValue{T: IntType, N: i2n(int64(i))}
	}
	av := m.Alloc.NewListArray(n)
	copy(av.List, elems)
	at := &ArrayType{Len: n, Elt: IntType}
	arrayTV := TypedValue{T: at, V: av}

	// bodyStmt for the range (no key/value assignment, empty body).
	bs := &bodyStmt{
		Body:          []Stmt{},
		BodyLen:       0,
		NextBodyIndex: -2,
		Op:            ILLEGAL, // no assignment
	}
	// A dummy stmt for PopFrameAndReset's final PopStmt.
	dummyStmt := &EmptyStmt{}
	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		// Push dummy stmt, frame, block, array value, bodyStmt.
		m.PushStmt(dummyStmt)
		m.Frames = append(m.Frames, Frame{
			NumOps:    len(m.Ops),
			NumValues: len(m.Values),
			NumExprs:  len(m.Exprs),
			NumStmts:  len(m.Stmts),
			NumBlocks: len(m.Blocks),
		})
		m.PushValue(arrayTV)
		*bs = bodyStmt{
			Body:          []Stmt{},
			BodyLen:       0,
			NextBodyIndex: -2,
			Op:            ILLEGAL,
		}
		m.PushStmt(bs)
		bm.SwitchOpCode(bmTarget)
		m.doOpExec(OpRangeIter)
		bm.SwitchOpCode(bmSetup)
		// For n>0: copies array, iterates once (empty body), terminates.
		// Stacks should be restored by PopFrameAndReset.
		m.Values = m.Values[:0]
		m.Stmts = m.Stmts[:0]
		m.Frames = m.Frames[:0]
		m.Blocks = m.Blocks[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpRangeIter_1(b *testing.B)    { benchOpRangeIter(b, 1) }
func BenchmarkOpRangeIter_10(b *testing.B)   { benchOpRangeIter(b, 10) }
func BenchmarkOpRangeIter_100(b *testing.B)  { benchOpRangeIter(b, 100) }
func BenchmarkOpRangeIter_1000(b *testing.B) { benchOpRangeIter(b, 1000) }

// --- OpRangeIterString: UTF-8 decode per rune ---

func benchOpRangeIterString(b *testing.B, length int) {
	m := benchMachine()
	defer m.Release()
	s := strings.Repeat("a", length) // ASCII, 1 byte per rune
	sv := m.Alloc.NewString(s)
	strTV := TypedValue{T: StringType, V: sv}

	bs := &bodyStmt{
		Body:          []Stmt{},
		BodyLen:       0,
		NextBodyIndex: -2,
		Op:            ILLEGAL,
	}
	dummyStmt := &EmptyStmt{}
	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushStmt(dummyStmt)
		m.Frames = append(m.Frames, Frame{
			NumOps:    len(m.Ops),
			NumValues: len(m.Values),
			NumExprs:  len(m.Exprs),
			NumStmts:  len(m.Stmts),
			NumBlocks: len(m.Blocks),
		})
		m.PushValue(strTV)
		*bs = bodyStmt{
			Body:          []Stmt{},
			BodyLen:       0,
			NextBodyIndex: -2,
			Op:            ILLEGAL,
		}
		m.PushStmt(bs)
		bm.SwitchOpCode(bmTarget)
		m.doOpExec(OpRangeIterString)
		bm.SwitchOpCode(bmSetup)
		m.Values = m.Values[:0]
		m.Stmts = m.Stmts[:0]
		m.Frames = m.Frames[:0]
		m.Blocks = m.Blocks[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpRangeIterString_1(b *testing.B)    { benchOpRangeIterString(b, 1) }
func BenchmarkOpRangeIterString_10(b *testing.B)   { benchOpRangeIterString(b, 10) }
func BenchmarkOpRangeIterString_100(b *testing.B)  { benchOpRangeIterString(b, 100) }
func BenchmarkOpRangeIterString_1000(b *testing.B) { benchOpRangeIterString(b, 1000) }

// --- OpRangeIterMap: linked list traversal ---

func benchOpRangeIterMap(b *testing.B, n int) {
	m := benchMachine()
	defer m.Release()
	mt := &MapType{Key: IntType, Value: IntType}
	mv := &MapValue{
		List: &MapList{},
		vmap: make(map[MapKey]*MapListItem, n),
	}
	for i := range n {
		k := TypedValue{T: IntType, N: i2n(int64(i))}
		v := TypedValue{T: IntType, N: i2n(int64(i * 10))}
		ptr := mv.GetPointerForKey(m.Alloc, m.Store, k)
		ptr.TV.Assign(m.Alloc, v, false)
	}
	mapTV := TypedValue{T: mt, V: mv}

	bs := &bodyStmt{
		Body:          []Stmt{},
		BodyLen:       0,
		NextBodyIndex: -2,
		Op:            ILLEGAL,
	}
	dummyStmt := &EmptyStmt{}
	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushStmt(dummyStmt)
		m.Frames = append(m.Frames, Frame{
			NumOps:    len(m.Ops),
			NumValues: len(m.Values),
			NumExprs:  len(m.Exprs),
			NumStmts:  len(m.Stmts),
			NumBlocks: len(m.Blocks),
		})
		m.PushValue(mapTV)
		*bs = bodyStmt{
			Body:          []Stmt{},
			BodyLen:       0,
			NextBodyIndex: -2,
			Op:            ILLEGAL,
		}
		m.PushStmt(bs)
		bm.SwitchOpCode(bmTarget)
		m.doOpExec(OpRangeIterMap)
		bm.SwitchOpCode(bmSetup)
		m.Values = m.Values[:0]
		m.Stmts = m.Stmts[:0]
		m.Frames = m.Frames[:0]
		m.Blocks = m.Blocks[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpRangeIterMap_1(b *testing.B)    { benchOpRangeIterMap(b, 1) }
func BenchmarkOpRangeIterMap_10(b *testing.B)   { benchOpRangeIterMap(b, 10) }
func BenchmarkOpRangeIterMap_100(b *testing.B)  { benchOpRangeIterMap(b, 100) }
func BenchmarkOpRangeIterMap_1000(b *testing.B) { benchOpRangeIterMap(b, 1000) }

// --- doOpTypeSwitch: clause × case type iteration ---

func benchOpTypeSwitch(b *testing.B, nClauses int) {
	m := benchMachine()
	defer m.Release()
	// Build switch with nClauses type cases, match on last.
	clauses := make([]SwitchClauseStmt, nClauses)
	for i := range nClauses {
		dt := &DeclaredType{
			PkgPath: "bench",
			Name:    Name("T" + string(rune('0'+i))),
			Base:    &StructType{PkgPath: "bench"},
		}
		clauses[i] = SwitchClauseStmt{
			Cases: []Expr{&constTypeExpr{Type: dt}},
			Body:  []Stmt{&EmptyStmt{}},
		}
		clauses[i].StaticBlock.NumNames = 0
		clauses[i].StaticBlock.HeapItems = []bool{}
		clauses[i].StaticBlock.Block.Source = &clauses[i]
	}
	ss := &SwitchStmt{
		IsTypeSwitch: true,
		Clauses:      clauses,
	}
	// The value to switch on: matches the last clause.
	matchType := clauses[nClauses-1].Cases[0].(*constTypeExpr).Type
	blk := &Block{Values: []TypedValue{}}
	m.Blocks = append(m.Blocks, blk)

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: matchType})
		m.PushStmt(ss)
		bm.SwitchOpCode(bmTarget)
		m.doOpTypeSwitch()
		bm.SwitchOpCode(bmSetup)
		m.Ops = m.Ops[:0]
		m.Stmts = m.Stmts[:0]
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpTypeSwitch_1(b *testing.B)   { benchOpTypeSwitch(b, 1) }
func BenchmarkOpTypeSwitch_10(b *testing.B)  { benchOpTypeSwitch(b, 10) }
func BenchmarkOpTypeSwitch_100(b *testing.B)  { benchOpTypeSwitch(b, 100) }
func BenchmarkOpTypeSwitch_1000(b *testing.B) { benchOpTypeSwitch(b, 1000) }

// --- doOpTypeSwitch with interface case: IsImplementedBy cost ---

func benchOpTypeSwitch_Interface(b *testing.B, nMethods int) {
	m := benchMachine()
	defer m.Release()

	iface, dt, _ := benchInterfaceAndImpl(m.Alloc, nMethods, nMethods)

	// Single clause matching the interface.
	clause := SwitchClauseStmt{
		Cases: []Expr{&constTypeExpr{Type: iface}},
		Body:  []Stmt{&EmptyStmt{}},
	}
	clause.StaticBlock.NumNames = 0
	clause.StaticBlock.HeapItems = []bool{}
	clause.StaticBlock.Block.Source = &clause

	ss := &SwitchStmt{
		IsTypeSwitch: true,
		Clauses:      []SwitchClauseStmt{clause},
	}
	blk := &Block{Values: []TypedValue{}}
	m.Blocks = append(m.Blocks, blk)

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{T: dt})
		m.PushStmt(ss)
		bm.SwitchOpCode(bmTarget)
		m.doOpTypeSwitch()
		bm.SwitchOpCode(bmSetup)
		m.Ops = m.Ops[:0]
		m.Stmts = m.Stmts[:0]
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpTypeSwitch_Interface_1(b *testing.B)   { benchOpTypeSwitch_Interface(b, 1) }
func BenchmarkOpTypeSwitch_Interface_10(b *testing.B)  { benchOpTypeSwitch_Interface(b, 10) }
func BenchmarkOpTypeSwitch_Interface_100(b *testing.B)  { benchOpTypeSwitch_Interface(b, 100) }
func BenchmarkOpTypeSwitch_Interface_1000(b *testing.B) { benchOpTypeSwitch_Interface(b, 1000) }

// --- doOpSwitchClause: clause index iteration ---

func BenchmarkOpSwitchClause_DefaultMatch(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	// Default clause (no cases) matches immediately.
	clause := SwitchClauseStmt{
		Cases: nil, // default
		Body:  []Stmt{&EmptyStmt{}},
	}
	clause.StaticBlock.NumNames = 0
	clause.StaticBlock.HeapItems = []bool{}
	clause.StaticBlock.Block.Source = &clause
	ss := &SwitchStmt{Clauses: []SwitchClauseStmt{clause}}
	blk := &Block{Values: []TypedValue{}}
	m.Blocks = append(m.Blocks, blk)

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushStmt(ss)                                        // switch stmt
		m.PushValue(TypedValue{T: IntType, N: i2n(0)})        // clause index
		m.PushValue(TypedValue{T: IntType, N: i2n(0)})        // case index
		m.PushValue(TypedValue{T: IntType, N: i2n(42)})       // tag value
		bm.SwitchOpCode(bmTarget)
		m.doOpSwitchClause()
		bm.SwitchOpCode(bmSetup)
		m.Ops = m.Ops[:0]
		m.Stmts = m.Stmts[:0]
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

// --- doOpSwitchClauseCase: isEql comparison per case ---

func benchOpSwitchClauseCase(b *testing.B, match bool) {
	m := benchMachine()
	defer m.Release()
	clause := SwitchClauseStmt{
		Cases: []Expr{&ConstExpr{TypedValue: TypedValue{T: IntType, N: i2n(42)}}},
		Body:  []Stmt{&EmptyStmt{}},
	}
	clause.StaticBlock.NumNames = 0
	clause.StaticBlock.HeapItems = []bool{}
	clause.StaticBlock.Block.Source = &clause
	ss := &SwitchStmt{Clauses: []SwitchClauseStmt{clause}}
	blk := &Block{Values: []TypedValue{}}
	m.Blocks = append(m.Blocks, blk)

	tagVal := int64(42)
	if !match {
		tagVal = 99
	}
	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushStmt(ss)                                          // switch stmt
		m.PushValue(TypedValue{T: IntType, N: i2n(0)})          // clause index
		m.PushValue(TypedValue{T: IntType, N: i2n(0)})          // case index
		m.PushValue(TypedValue{T: IntType, N: i2n(tagVal)})     // tag value
		m.PushValue(TypedValue{T: IntType, N: i2n(42)})         // case value (evaluated)
		bm.SwitchOpCode(bmTarget)
		m.doOpSwitchClauseCase()
		bm.SwitchOpCode(bmSetup)
		m.Ops = m.Ops[:0]
		m.Stmts = m.Stmts[:0]
		m.Exprs = m.Exprs[:0]
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpSwitchClauseCase_Match(b *testing.B)  { benchOpSwitchClauseCase(b, true) }
func BenchmarkOpSwitchClauseCase_Miss(b *testing.B)   { benchOpSwitchClauseCase(b, false) }

// --- op_types.go: type construction ops ---

func BenchmarkOpFieldType(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	x := &FieldTypeExpr{NameExpr: NameExpr{Name: "x"}, Tag: nil}
	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(TypedValue{V: TypeValue{Type: IntType}})
		m.PushExpr(x)
		bm.SwitchOpCode(bmTarget)
		m.doOpFieldType()
		bm.SwitchOpCode(bmSetup)
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpArrayType(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	x := &ArrayTypeExpr{Len: &ConstExpr{}}
	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(asValue(IntType))                    // element type
		m.PushValue(TypedValue{T: IntType, N: i2n(10)})  // length
		m.PushExpr(x)
		bm.SwitchOpCode(bmTarget)
		m.doOpArrayType()
		bm.SwitchOpCode(bmSetup)
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpSliceType(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	x := &SliceTypeExpr{}
	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(asValue(IntType))
		m.PushExpr(x)
		bm.SwitchOpCode(bmTarget)
		m.doOpSliceType()
		bm.SwitchOpCode(bmSetup)
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func benchOpFuncType(b *testing.B, nParams, nResults int) {
	m := benchMachine()
	defer m.Release()

	params := make([]FieldTypeExpr, nParams)
	for i := range nParams {
		params[i] = FieldTypeExpr{NameExpr: NameExpr{Name: Name("p" + string(rune('a'+i)))}}
	}
	results := make([]FieldTypeExpr, nResults)
	for i := range nResults {
		results[i] = FieldTypeExpr{NameExpr: NameExpr{Name: Name("r" + string(rune('a'+i)))}}
	}
	x := &FuncTypeExpr{Params: params, Results: results}

	// Pre-build the FieldType values to push.
	fieldTVs := make([]TypedValue, nParams+nResults)
	for i := range nParams {
		ft := FieldType{Name: Name("p" + string(rune('a'+i))), Type: IntType}
		fieldTVs[i] = TypedValue{T: gTypeType, V: toTypeValue(ft)}
	}
	for i := range nResults {
		ft := FieldType{Name: Name("r" + string(rune('a'+i))), Type: IntType}
		fieldTVs[nParams+i] = TypedValue{T: gTypeType, V: toTypeValue(ft)}
	}

	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		for _, tv := range fieldTVs {
			m.PushValue(tv)
		}
		m.PushExpr(x)
		bm.SwitchOpCode(bmTarget)
		m.doOpFuncType()
		bm.SwitchOpCode(bmSetup)
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpFuncType_0Params_0Results(b *testing.B)    { benchOpFuncType(b, 0, 0) }
func BenchmarkOpFuncType_1Params_0Results(b *testing.B)    { benchOpFuncType(b, 1, 0) }
func BenchmarkOpFuncType_10Params_0Results(b *testing.B)   { benchOpFuncType(b, 10, 0) }
func BenchmarkOpFuncType_100Params_0Results(b *testing.B)  { benchOpFuncType(b, 100, 0) }
func BenchmarkOpFuncType_1000Params_0Results(b *testing.B) { benchOpFuncType(b, 1000, 0) }
func BenchmarkOpFuncType_0Params_1Results(b *testing.B)    { benchOpFuncType(b, 0, 1) }
func BenchmarkOpFuncType_0Params_10Results(b *testing.B)   { benchOpFuncType(b, 0, 10) }
func BenchmarkOpFuncType_0Params_100Results(b *testing.B)  { benchOpFuncType(b, 0, 100) }
func BenchmarkOpFuncType_0Params_1000Results(b *testing.B) { benchOpFuncType(b, 0, 1000) }

func BenchmarkOpMapType(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(asValue(IntType))    // key type
		m.PushValue(asValue(StringType)) // value type
		bm.SwitchOpCode(bmTarget)
		m.doOpMapType()
		bm.SwitchOpCode(bmSetup)
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func benchOpStructType(b *testing.B, nFields int) {
	m := benchMachine()
	defer m.Release()
	fields := make([]FieldTypeExpr, nFields)
	fieldTVs := make([]TypedValue, nFields)
	for i := range nFields {
		fields[i] = FieldTypeExpr{NameExpr: NameExpr{Name: Name("f" + string(rune('0'+i%10)))}}
		ft := FieldType{Name: fields[i].Name, Type: IntType}
		fieldTVs[i] = TypedValue{T: gTypeType, V: toTypeValue(ft)}
	}
	x := &StructTypeExpr{Fields: fields}
	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		for i := range nFields {
			m.PushValue(fieldTVs[i])
		}
		m.PushExpr(x)
		bm.SwitchOpCode(bmTarget)
		m.doOpStructType()
		bm.SwitchOpCode(bmSetup)
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpStructType_1(b *testing.B)   { benchOpStructType(b, 1) }
func BenchmarkOpStructType_10(b *testing.B)  { benchOpStructType(b, 10) }
func BenchmarkOpStructType_100(b *testing.B)  { benchOpStructType(b, 100) }
func BenchmarkOpStructType_1000(b *testing.B) { benchOpStructType(b, 1000) }

func benchOpInterfaceType(b *testing.B, nMethods int) {
	m := benchMachine()
	defer m.Release()
	methods := make([]FieldTypeExpr, nMethods)
	methodTVs := make([]TypedValue, nMethods)
	for i := range nMethods {
		methods[i] = FieldTypeExpr{NameExpr: NameExpr{Name: Name("M" + string(rune('a'+i%26)))}}
		mft := FieldType{
			Name: methods[i].Name,
			Type: &FuncType{Params: []FieldType{}, Results: []FieldType{}},
		}
		methodTVs[i] = TypedValue{T: gTypeType, V: toTypeValue(mft)}
	}
	x := &InterfaceTypeExpr{Methods: methods}
	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		for i := range nMethods {
			m.PushValue(methodTVs[i])
		}
		m.PushExpr(x)
		bm.SwitchOpCode(bmTarget)
		m.doOpInterfaceType()
		bm.SwitchOpCode(bmSetup)
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

func BenchmarkOpInterfaceType_1(b *testing.B)   { benchOpInterfaceType(b, 1) }
func BenchmarkOpInterfaceType_10(b *testing.B)  { benchOpInterfaceType(b, 10) }
func BenchmarkOpInterfaceType_100(b *testing.B)  { benchOpInterfaceType(b, 100) }
func BenchmarkOpInterfaceType_1000(b *testing.B) { benchOpInterfaceType(b, 1000) }

func BenchmarkOpChanType(b *testing.B) {
	m := benchMachine()
	defer m.Release()
	x := &ChanTypeExpr{Dir: SEND | RECV}
	bm.InitMeasure()
	bm.BeginOpCode(bmSetup)
	for range b.N {
		m.PushValue(asValue(IntType))
		m.PushExpr(x)
		bm.SwitchOpCode(bmTarget)
		m.doOpChanType()
		bm.SwitchOpCode(bmSetup)
		m.Values = m.Values[:0]
	}
	reportBenchops(b)
}

// ---------------------------------------------------------------------------
// Helper: encode int64/uint64 into [8]byte (little-endian, matching unsafe cast)
// ---------------------------------------------------------------------------

func i2n(v int64) [8]byte {
	var n [8]byte
	n[0] = byte(v)
	n[1] = byte(v >> 8)
	n[2] = byte(v >> 16)
	n[3] = byte(v >> 24)
	n[4] = byte(v >> 32)
	n[5] = byte(v >> 40)
	n[6] = byte(v >> 48)
	n[7] = byte(v >> 56)
	return n
}

func u2n(v uint64) [8]byte {
	return i2n(int64(v))
}
