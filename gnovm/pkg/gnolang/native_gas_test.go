package gnolang

import (
	"math"
	"os"
	"regexp"
	"strconv"
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
	for range 10 {
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

func TestModExpWork(t *testing.T) {
	// units(expLen) * (modExpFloorWords + ceil(modLen/8)^2), where units() is
	// the modular-multiplication count of whichever expNN routine runs.
	const (
		gen  = modExpGenericPerBit * 8 // generic loop, per exponent byte
		mont = modExpMontPerWord       // Montgomery, per exponent word
	)
	cases := []struct {
		expLen, modLen int64
		want           int64
	}{
		{0, 256, 0}, // exponent 0: Exp returns immediately
		{256, 0, 0}, // empty modulus: X_modExp short-circuits
		{0, 0, 0},   //
		{1, 8, gen * (modExpFloorWords + 1)},
		{1, 1, gen * (modExpFloorWords + 1)}, // sub-word modulus still costs one word
		{1, 9, gen * (modExpFloorWords + 4)}, // 9 bytes rounds up to 2 words

		// Either side of the one-word exponent crossover. big.Int runs the
		// generic loop at 8 bytes and Montgomery at 9. The two must not cross:
		// an 8-byte exponent measures more expensive than a 9-byte one, so
		// charging it less would be a real undercharge.
		{7, 256, 7 * gen * (modExpFloorWords + 32*32)},
		{8, 256, 8 * gen * (modExpFloorWords + 32*32)},
		{9, 256, (modExpMontSetup + 2*mont) * (modExpFloorWords + 32*32)},
		{16, 256, (modExpMontSetup + 2*mont) * (modExpFloorWords + 32*32)},
		{17, 256, (modExpMontSetup + 3*mont) * (modExpFloorWords + 32*32)},

		// The shape the production slope is anchored to. Whole multiples of a
		// word are unaffected by the rounding.
		{256, 256, (modExpMontSetup + 32*mont) * (modExpFloorWords + 32*32)},
		{1024, 1024, (modExpMontSetup + 128*mont) * (modExpFloorWords + 128*128)},

		// Operand lengths reach the gas layer before crypto/modexp rejects
		// them, so the metric must saturate rather than wrap.
		{500_000_000, 500_000_000, modExpWorkSaturation},
		{1, 500_000_000, modExpWorkSaturation},
		{math.MaxInt64, math.MaxInt64, modExpWorkSaturation},
		// A huge exponent against a 1-byte modulus stays under the saturation
		// point: the quadratic term is 1, so only the floor term applies. The
		// exponent is clamped to maxOperandLenForWork (1<<28) first — flattening
		// above that is safe because X_modExp rejects anything past 1024 bytes
		// and returns without doing the work at all.
		{
			500_000_000, 1,
			(modExpMontSetup + ((1 << 28) / 8 * mont)) * (modExpFloorWords + 1),
		},
	}
	for _, c := range cases {
		if got := modExpWork(c.expLen, c.modLen); got != c.want {
			t.Errorf("modExpWork(%d, %d) = %d, want %d", c.expLen, c.modLen, got, c.want)
		}
	}
}

// TestModExpWorkNoCrossoverInversion pins the property the crossover exists to
// provide: cost must never fall as the exponent grows by one byte. The metric
// this replaced charged 8*len(exp) uniformly, which priced an 8-byte exponent
// below a 9-byte one even though big.Int's generic loop makes 8 bytes the more
// expensive of the two.
func TestModExpWorkNoCrossoverInversion(t *testing.T) {
	for modLen := int64(1); modLen <= 1024; modLen *= 2 {
		for expLen := int64(1); expLen < 64; expLen++ {
			prev := modExpWork(expLen-1, modLen)
			if got := modExpWork(expLen, modLen); got < prev {
				t.Fatalf("modExpWork(%d, %d) = %d < modExpWork(%d, %d) = %d",
					expLen, modLen, got, expLen-1, modLen, prev)
			}
		}
	}
}

// TestModExpWorkSlopeProductFitsInt64 guards the unguarded `gi.Slope * N`
// multiply in chargeNativeGas: the saturation clamp is only safe if the shipped
// slope times the clamp still fits. Uses the largest slope any row could
// plausibly carry after a re-fit rather than today's value, so raising the slope
// trips this before it trips consensus.
func TestModExpWorkSlopeProductFitsInt64(t *testing.T) {
	const maxPlausibleSlope = 1 << 16
	if modExpWorkSaturation > math.MaxInt64/maxPlausibleSlope {
		t.Fatalf("modExpWorkSaturation %d * slope %d overflows int64 — lower the "+
			"clamp or route the multiply through overflow.Mul",
			int64(modExpWorkSaturation), maxPlausibleSlope)
	}
}

// TestModExpWorkMonotonic: the metric must never decrease as either operand
// grows, or a larger call could be charged less than a smaller one.
func TestModExpWorkMonotonic(t *testing.T) {
	lens := []int64{1, 2, 7, 8, 9, 32, 64, 255, 256, 1024, 1 << 20, 1 << 30}
	for i, exp := range lens {
		for j, mod := range lens {
			got := modExpWork(exp, mod)
			if i > 0 {
				if prev := modExpWork(lens[i-1], mod); got < prev {
					t.Fatalf("non-monotonic in exp at mod=%d: %d(exp=%d) < %d(exp=%d)",
						mod, got, exp, prev, lens[i-1])
				}
			}
			if j > 0 {
				if prev := modExpWork(exp, lens[j-1]); got < prev {
					t.Fatalf("non-monotonic in mod at exp=%d: %d(mod=%d) < %d(mod=%d)",
						exp, got, mod, prev, lens[j-1])
				}
			}
		}
	}
}

func TestChargeNativeGas_ModExpWork(t *testing.T) {
	// Slope 1024 => 1 gas per unit of work, so Cycles reads back the metric.
	// SlopeIdx 0 names the exponent; the modulus is read from param 1.
	cleanup := registerTestNative(t, &NativeGasInfo{
		Base:  7,
		Slope: 1024, SlopeIdx: 0, SlopeKind: SizeModExpWork,
	})
	defer cleanup()

	cases := []struct {
		expLen, modLen int
		want           int64
	}{
		{0, 0, 7},
		{1, 8, 7 + (modExpGenericPerBit*8)*(modExpFloorWords+1)},
		{256, 256, 7 + (modExpMontSetup+32*modExpMontPerWord)*(modExpFloorWords+1024)},
		// Asymmetry is priced: a big exponent against a small modulus is no
		// longer free, which is what the old len(modulus)-only slope missed.
		{1024, 32, 7 + (modExpMontSetup+128*modExpMontPerWord)*(modExpFloorWords+16)},
	}
	for _, c := range cases {
		m := stubMachine([]int{c.expLen, c.modLen})
		_ = m.chargeNativeGas(&FuncValue{NativePkg: testNativePkg, NativeName: testNativeFn})
		if m.Cycles != c.want {
			t.Errorf("expLen=%d modLen=%d: got %d cycles, want %d", c.expLen, c.modLen, m.Cycles, c.want)
		}
	}
}

// TestChargeNativeGas_ModExpWorkMissingPair guards the pair read: a native
// registered with SizeModExpWork but only one parameter present must charge
// Base rather than index out of range.
func TestChargeNativeGas_ModExpWorkMissingPair(t *testing.T) {
	cleanup := registerTestNative(t, &NativeGasInfo{
		Base:  11,
		Slope: 1024, SlopeIdx: 0, SlopeKind: SizeModExpWork,
	})
	defer cleanup()

	m := stubMachine([]int{256})
	_ = m.chargeNativeGas(&FuncValue{NativePkg: testNativePkg, NativeName: testNativeFn})
	if m.Cycles != 11 {
		t.Fatalf("got %d cycles, want 11 (base only)", m.Cycles)
	}
}

func TestChargeNativeGas_PreCallTwoSlopesOnDistinctParams(t *testing.T) {
	// Mimic crypto/merkle.innerHash: two independent unbounded byte params,
	// each carrying the same per-byte rate so the charge tracks the sum of
	// their lengths. Metering only one would leave the other free — an
	// caller just moves the payload into the unmetered operand.
	cleanup := registerTestNative(t, &NativeGasInfo{
		Base:  100,
		Slope: 1024, SlopeIdx: 0, SlopeKind: SizeLenBytes,
		Slope2: 1024, Slope2Idx: 1, Slope2Kind: SizeLenBytes,
	})
	defer cleanup()

	cases := []struct {
		left, right int
		want        int64
	}{
		{0, 0, 100},
		{32, 32, 100 + 64},
		{4096, 0, 100 + 4096}, // all payload on the left
		{0, 4096, 100 + 4096}, // ...and on the right: same charge
		{1 << 20, 0, 100 + 1<<20},
		{1 << 19, 1 << 19, 100 + 1<<20},
	}
	for _, c := range cases {
		m := stubMachine([]int{c.left, c.right})
		_ = m.chargeNativeGas(&FuncValue{NativePkg: testNativePkg, NativeName: testNativeFn})
		if m.Cycles != c.want {
			t.Errorf("left=%d right=%d: got %d, want %d", c.left, c.right, m.Cycles, c.want)
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

// TestModExpConstantsMatchFitter keeps the runtime metric and the calibration
// fitter computing the same quantity. gen_native_table.py has to reduce each
// (expLen, modLen) bench point to a work value before it can fit a slope over
// them, so it carries its own copy of these constants. If the two drift, the
// fitter reduces every bench to a work value the runtime never charges on and
// the slope it emits is scaled by the ratio between them — a silently mispriced
// consensus row that no other test would catch, since the Go-side tests check
// the row against modExpWork and never against the fitter.
func TestModExpConstantsMatchFitter(t *testing.T) {
	t.Parallel()

	const fitter = "../../cmd/calibrate/gen_native_table.py"
	src, err := os.ReadFile(fitter)
	if err != nil {
		t.Fatalf("read %s: %v", fitter, err)
	}
	for _, c := range []struct {
		pyName string
		goVal  int64
	}{
		{"MODEXP_FLOOR_WORDS", modExpFloorWords},
		{"MODEXP_WORD_BYTES", modExpWordBytes},
		{"MODEXP_MONT_SETUP", modExpMontSetup},
		{"MODEXP_MONT_PER_WORD", modExpMontPerWord},
		{"MODEXP_GENERIC_PER_BIT", modExpGenericPerBit},
	} {
		re := regexp.MustCompile(`(?m)^` + c.pyName + `\s*=\s*(\d+)\s*$`)
		m := re.FindSubmatch(src)
		if m == nil {
			t.Errorf("%s: %s not found — the fitter must define it, or this test "+
				"can no longer tell whether the two agree", fitter, c.pyName)
			continue
		}
		got, err := strconv.ParseInt(string(m[1]), 10, 64)
		if err != nil {
			t.Errorf("%s: %s = %q, not an integer", fitter, c.pyName, m[1])
			continue
		}
		if got != c.goVal {
			t.Errorf("%s has %s = %d, but the runtime uses %d — the fitter would "+
				"reduce the calibration benches to a metric production never charges",
				fitter, c.pyName, got, c.goVal)
		}
	}
}
