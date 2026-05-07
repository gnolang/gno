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
func registerTestNative(tb testing.TB, gi *NativeGasInfo) func() {
	tb.Helper()
	key := testNativePkg + "\x00" + string(testNativeFn)
	if _, exists := nativeGasIndex[key]; exists {
		tb.Fatalf("test native key %q already registered — fix test cleanup", key)
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
		{0, 45},              // base only
		{1024, 45 + 390},     // +1 KiB → +slope
		{2048, 45 + 780},     // +2 KiB → +2*slope
		{65536, 45 + 390*64}, // +64 KiB
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

func TestChargeNativeGas_PanicOnUncalibratedStdlib(t *testing.T) {
	// Stdlib native (non-empty NativePkg) without a registered gas
	// entry, with a real GasMeter installed → panic. This is the
	// forcing function ensuring every new native ships with a gas
	// entry. Test/no-meter Machines bypass the panic (see next test).
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
	// No GasMeter installed → silent no-op for uncalibrated natives.
	// Lets unit tests build minimal Machines without registering gas.
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

// stubMachineWithSliceParam builds a Machine with one block slot
// holding a []string SliceValue whose inner elements each have length
// `innerLen`. Used to exercise SizeSliceTotalBytes and SizeLenSlice on
// the same param.
func stubMachineWithSliceParam(count, innerLen int) *Machine {
	m := &Machine{GasMeter: &recordingMeter{}}
	av := &ArrayValue{List: make([]TypedValue, count)}
	for i := range av.List {
		av.List[i] = TypedValue{T: StringType, V: StringValue(string(make([]byte, innerLen)))}
	}
	sv := &SliceValue{Base: av, Offset: 0, Length: count, Maxcap: count}
	blk := &Block{Values: []TypedValue{{T: &SliceType{Elt: StringType}, V: sv}}}
	m.Blocks = []*Block{blk}
	return m
}

func TestChargeNativeGas_SliceTotalBytes(t *testing.T) {
	// Slope1 on count, Slope2 on total inner bytes — both at SlopeIdx 0.
	// per-element slope = 1024 (=> 1 ns/element), per-byte slope = 1024
	// (=> 1 ns/byte). For count=4, innerLen=10 → cost = base + 4 + 40.
	cleanup := registerTestNative(t, &NativeGasInfo{
		Base:  100,
		Slope: 1024, SlopeIdx: 0, SlopeKind: SizeLenSlice,
		Slope2: 1024, Slope2Idx: 0, Slope2Kind: SizeSliceTotalBytes,
	})
	defer cleanup()

	cases := []struct {
		count, innerLen int
		want            int64
	}{
		{0, 0, 100},                       // base only
		{4, 10, 100 + 4 + 40},             // 4 elements, 40 bytes total
		{16, 100, 100 + 16 + 1600},        // larger
		{2, 50_000, 100 + 2 + 100_000},    // bytes-dominated
		{128, 1024, 100 + 128 + 128*1024}, // large in both dims
	}
	for _, c := range cases {
		m := stubMachineWithSliceParam(c.count, c.innerLen)
		_ = m.chargeNativeGas(&FuncValue{NativePkg: testNativePkg, NativeName: testNativeFn})
		if m.Cycles != c.want {
			t.Errorf("count=%d innerLen=%d: got %d, want %d", c.count, c.innerLen, m.Cycles, c.want)
		}
	}
}

func TestChargeNativeGas_PostCallTwoSlopes(t *testing.T) {
	// Mimic getSysParamStrings: post-call charges per element AND per
	// total inner bytes on the returned []string at stack offset 2.
	cleanup := registerTestNative(t, &NativeGasInfo{
		Base: 50, SlopeIdx: -1, SlopeKind: SizeFlat,
		PostBase:  30,
		PostSlope: 1024, PostSlopeIdx: 2, PostSlopeKind: SizeReturnLen,
		PostSlope2: 1024, PostSlope2Idx: 2, PostSlope2Kind: SizeSliceTotalBytes,
	})
	defer cleanup()

	m := stubMachine(nil)
	gi := m.chargeNativeGas(&FuncValue{NativePkg: testNativePkg, NativeName: testNativeFn})
	if gi == nil {
		t.Fatal("expected non-nil gi for post-charge native")
	}
	if m.Cycles != 50 {
		t.Fatalf("pre-call: got %d, want 50", m.Cycles)
	}

	// Push 8 returns of 32 bytes each at offset 2 (slice), plus a dummy
	// at offset 1 (any TV). Expected post: 30 + 8 + 8*32 = 294.
	count, innerLen := 8, 32
	av := &ArrayValue{List: make([]TypedValue, count)}
	for i := range av.List {
		av.List[i] = TypedValue{T: StringType, V: StringValue(string(make([]byte, innerLen)))}
	}
	sv := &SliceValue{Base: av, Offset: 0, Length: count, Maxcap: count}
	returnSlice := TypedValue{T: &SliceType{Elt: StringType}, V: sv}
	dummy := TypedValue{T: StringType, V: StringValue("")}
	m.PushValue(returnSlice)
	m.PushValue(dummy)
	m.chargeNativeGasPost(gi)
	if want := int64(50 + 30 + count + count*innerLen); m.Cycles != want {
		t.Fatalf("post-call: got %d, want %d", m.Cycles, want)
	}
}

func TestChargeNativeGas_SliceTotalBytes_NilAndOffset(t *testing.T) {
	// Empty / nil slice TV → SizeSliceTotalBytes returns 0 cleanly.
	cleanup := registerTestNative(t, &NativeGasInfo{
		Base:  10,
		Slope: 1024, SlopeIdx: 0, SlopeKind: SizeSliceTotalBytes,
	})
	defer cleanup()

	// Nil V → 0 inner bytes.
	m := &Machine{GasMeter: &recordingMeter{}}
	m.Blocks = []*Block{{Values: []TypedValue{{T: &SliceType{Elt: StringType}, V: nil}}}}
	_ = m.chargeNativeGas(&FuncValue{NativePkg: testNativePkg, NativeName: testNativeFn})
	if m.Cycles != 10 {
		t.Fatalf("nil slice V: got %d, want 10 (base only)", m.Cycles)
	}

	// Sliced array with offset — only the [Offset:Length] window counts.
	av := &ArrayValue{List: []TypedValue{
		{T: StringType, V: StringValue("aa")},     // 2 — outside window
		{T: StringType, V: StringValue("bbbbb")},  // 5 — in window
		{T: StringType, V: StringValue("ccc")},    // 3 — in window
		{T: StringType, V: StringValue("dddddd")}, // 6 — outside window
	}}
	sv := &SliceValue{Base: av, Offset: 1, Length: 2, Maxcap: 3}
	m2 := &Machine{GasMeter: &recordingMeter{}}
	m2.Blocks = []*Block{{Values: []TypedValue{{T: &SliceType{Elt: StringType}, V: sv}}}}
	_ = m2.chargeNativeGas(&FuncValue{NativePkg: testNativePkg, NativeName: testNativeFn})
	if want := int64(10 + 5 + 3); m2.Cycles != want {
		t.Fatalf("offset window: got %d, want %d (only inner bbbbb+ccc count)", m2.Cycles, want)
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
