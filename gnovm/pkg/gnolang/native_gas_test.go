package gnolang

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/store/types"
)

// recordingMeter captures consumed gas amounts per ConsumeGas call.
type recordingMeter struct {
	consumed types.Gas
}

func (r *recordingMeter) GasConsumed() types.Gas        { return r.consumed }
func (r *recordingMeter) GasConsumedToLimit() types.Gas { return r.consumed }
func (r *recordingMeter) Remaining() types.Gas          { return 1 << 30 }
func (r *recordingMeter) Limit() types.Gas              { return 1 << 30 }
func (r *recordingMeter) ConsumeGas(amount types.Gas, _ string) {
	r.consumed += amount
}
func (r *recordingMeter) RefundGas(amount types.Gas, _ string) {
	r.consumed -= amount
}
func (r *recordingMeter) IsPastLimit() bool { return false }
func (r *recordingMeter) IsOutOfGas() bool  { return false }

const (
	testNativePkg = "x_test_native"
	testNativeFn  = Name("fn")
)

// registerTestNative installs a temporary entry into the package-global
// nativeGasIndex and returns a cleanup func. Tests that need a "no entry"
// state should NOT call this; they get the panic / no-meter behavior.
func registerTestNative(t testing.TB, gi *NativeGasInfo) func() {
	t.Helper()
	key := testNativePkg + "\x00" + string(testNativeFn)
	if _, exists := nativeGasIndex[key]; exists {
		t.Fatalf("test native key %q already registered — fix test cleanup", key)
	}
	nativeGasIndex[key] = gi
	return func() { delete(nativeGasIndex, key) }
}

// stubMachine builds a Machine with a single block. paramLens is the
// length of each block-slot string (used by SizeLenBytes/String).
func stubMachine(paramLens []int) *Machine {
	m := &Machine{GasMeter: &recordingMeter{}}
	blk := &Block{Values: make([]TypedValue, len(paramLens))}
	for i, n := range paramLens {
		blk.Values[i] = TypedValue{T: StringType, V: StringValue(string(make([]byte, n)))}
	}
	m.Blocks = []*Block{blk}
	return m
}

func TestChargeNativeGas_Flat(t *testing.T) {
	cleanup := registerTestNative(t, &NativeGasInfo{Base: 100, SlopeIdx: -1, SlopeKind: SizeFlat})
	defer cleanup()
	m := stubMachine([]int{0})
	gi := m.chargeNativeGas(&FuncValue{NativePkg: testNativePkg, NativeName: testNativeFn})
	if gi != nil {
		t.Fatalf("flat with no post-charge: chargeNativeGas should return nil, got %+v", gi)
	}
	if m.Cycles != 100 {
		t.Fatalf("flat: got %d cycles, want 100", m.Cycles)
	}
}

func TestChargeNativeGas_LinearScalesWithInput(t *testing.T) {
	cleanup := registerTestNative(t, &NativeGasInfo{
		Base: 45, Slope: 390, SlopeIdx: 0, SlopeKind: SizeLenBytes,
	})
	defer cleanup()

	cases := []struct {
		n    int
		want int64
	}{
		{0, 45},                // base only
		{1024, 45 + 390},       // +1 KiB → +slope
		{2048, 45 + 780},       // +2 KiB → +2*slope
		{65536, 45 + 390*64},   // +64 KiB
	}
	for _, c := range cases {
		m := stubMachine([]int{c.n})
		_ = m.chargeNativeGas(&FuncValue{NativePkg: testNativePkg, NativeName: testNativeFn})
		if m.Cycles != c.want {
			t.Errorf("n=%d: got %d cycles, want %d", c.n, m.Cycles, c.want)
		}
	}

	// Sanity: gas must monotonically increase with input.
	prev := int64(-1)
	for _, n := range []int{0, 1, 64, 1024, 16384, 65536, 1 << 20} {
		m := stubMachine([]int{n})
		_ = m.chargeNativeGas(&FuncValue{NativePkg: testNativePkg, NativeName: testNativeFn})
		if m.Cycles < prev {
			t.Fatalf("non-monotonic: n=%d cycles=%d < prev=%d", n, m.Cycles, prev)
		}
		prev = m.Cycles
	}
}

func TestChargeNativeGas_PostCallReturnLen(t *testing.T) {
	// Mimic bankerGetCoins: pre=flat 100, post=20*N/1024 on the return at
	// stack offset 2 (a slice with length 1024 → +20 cost).
	cleanup := registerTestNative(t, &NativeGasInfo{
		Base: 100, SlopeIdx: -1, SlopeKind: SizeFlat,
		PostBase: 50, PostSlope: 20480, PostSlopeIdx: 2, PostSlopeKind: SizeReturnLen,
	})
	defer cleanup()
	m := stubMachine(nil)
	gi := m.chargeNativeGas(&FuncValue{NativePkg: testNativePkg, NativeName: testNativeFn})
	if gi == nil {
		t.Fatal("expected non-nil gi for native with post-call charge")
	}
	if m.Cycles != 100 {
		t.Fatalf("pre-call: got %d, want 100", m.Cycles)
	}
	// Simulate nativeBody pushing returns: top of stack is "amounts" (any
	// TV at offset 1), bottom is "denoms" (slice of len 1024 at offset 2).
	denoms := TypedValue{T: StringType, V: StringValue(string(make([]byte, 1024)))}
	amounts := TypedValue{T: StringType, V: StringValue("")}
	m.PushValue(denoms)
	m.PushValue(amounts)
	m.chargeNativeGasPost(gi)
	// 50 + 20480*1024/1024 = 50 + 20480 = 20530, plus pre-call 100 = 20630
	if want := int64(100 + 50 + 20480); m.Cycles != want {
		t.Fatalf("post-call: got %d, want %d", m.Cycles, want)
	}
}

func TestChargeNativeGas_NumCallFrames(t *testing.T) {
	cleanup := registerTestNative(t, &NativeGasInfo{
		Base: 100, Slope: 1024, SlopeIdx: -1, SlopeKind: SizeNumCallFrames,
	})
	defer cleanup()
	m := stubMachine(nil)
	// Add 10 call frames (Func != nil).
	for i := 0; i < 10; i++ {
		m.Frames = append(m.Frames, Frame{Func: &FuncValue{}})
	}
	m.chargeNativeGas(&FuncValue{NativePkg: testNativePkg, NativeName: testNativeFn})
	// 100 + 1024*10/1024 = 110
	if m.Cycles != 110 {
		t.Fatalf("NumCallFrames: got %d cycles, want 110", m.Cycles)
	}
}

func TestChargeNativeGas_PanicOnUncalibrated(t *testing.T) {
	// Stdlib native (non-empty NativePkg) without a registered entry,
	// real GasMeter present → must panic.
	m := stubMachine(nil)
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic on uncalibrated stdlib native")
		}
	}()
	m.chargeNativeGas(&FuncValue{NativePkg: "uncalibrated", NativeName: "fn"})
}

func TestChargeNativeGas_NoPanicWithoutMeter(t *testing.T) {
	// No GasMeter installed → don't panic on uncalibrated; just no-op.
	m := &Machine{Blocks: []*Block{{}}}
	m.chargeNativeGas(&FuncValue{NativePkg: "uncalibrated", NativeName: "fn"})
	if m.Cycles != 0 {
		t.Fatalf("no-meter path: got %d cycles, want 0", m.Cycles)
	}
}

func TestChargeNativeGas_FallbackForUverseBuiltin(t *testing.T) {
	// Empty NativePkg means uverse builtin (DefineNative path) — flat fallback.
	m := stubMachine(nil)
	m.chargeNativeGas(&FuncValue{}) // no NativePkg
	if m.Cycles != int64(OpCPUCallNativeBody) {
		t.Fatalf("uverse fallback: got %d cycles, want %d", m.Cycles, OpCPUCallNativeBody)
	}
}

// ---- Microbenchmarks for chargeNativeGas overhead ----

func benchStubMachine(b *testing.B, gi *NativeGasInfo, paramLen int, registerKey bool) (*Machine, *FuncValue) {
	b.Helper()
	m := &Machine{GasMeter: &recordingMeter{}}
	blk := &Block{Values: []TypedValue{
		{T: StringType, V: StringValue(string(make([]byte, paramLen)))},
	}}
	m.Blocks = []*Block{blk}
	fv := &FuncValue{NativePkg: testNativePkg, NativeName: testNativeFn}
	if registerKey && gi != nil {
		// Bench-time registration, cleared via b.Cleanup.
		key := testNativePkg + "\x00" + string(testNativeFn)
		nativeGasIndex[key] = gi
		b.Cleanup(func() { delete(nativeGasIndex, key) })
	}
	return m, fv
}

func BenchmarkChargeNativeGas_FallbackUverse(b *testing.B) {
	// Empty NativePkg → no map lookup, just flat incrCPU.
	m, _ := benchStubMachine(b, nil, 0, false)
	fv := &FuncValue{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.chargeNativeGas(fv)
	}
}

func BenchmarkChargeNativeGas_CalibratedFlat(b *testing.B) {
	gi := &NativeGasInfo{Base: 100, SlopeIdx: -1, SlopeKind: SizeFlat}
	m, fv := benchStubMachine(b, gi, 0, true)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.chargeNativeGas(fv)
	}
}

func BenchmarkChargeNativeGas_CalibratedLinear(b *testing.B) {
	gi := &NativeGasInfo{Base: 45, Slope: 390, SlopeIdx: 0, SlopeKind: SizeLenBytes}
	m, fv := benchStubMachine(b, gi, 1024, true)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.chargeNativeGas(fv)
	}
}

// Baseline: incrCPU alone (the floor of any charging path).
func BenchmarkChargeNativeGas_IncrCPUBaseline(b *testing.B) {
	m := &Machine{GasMeter: &recordingMeter{}}
	for i := 0; i < b.N; i++ {
		m.incrCPU(150)
	}
}
