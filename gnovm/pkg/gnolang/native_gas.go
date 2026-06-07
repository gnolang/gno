package gnolang

import "fmt"

// Per-native gas charging. See gnovm/cmd/calibrate/gen_native_table.py and
// gnovm/cmd/calibrate/native_gas_formulas.md for the calibration pipeline.
//
// The table is registered into nativeGasIndex at init time by stdlibs (and
// any other package that ships natives) via RegisterNativeGas. This mirrors
// the pattern used for OpCPU* constants — gas info lives globally next to
// the dispatcher, no per-Machine plumbing needed. Each native call does
// one map lookup; the alternative (caching on FuncValue) bumps allocator
// gas via increased FuncValue size.

// NativeGasSize names the rule for extracting N from a native's parameters
// (pre-call) or return values (post-call).
type NativeGasSize uint8

const (
	SizeFlat            NativeGasSize = 0 // no slope, no N
	SizeLenBytes        NativeGasSize = 1 // len(param[idx]) — []byte
	SizeLenString       NativeGasSize = 2 // len(param[idx]) — string
	SizeLenSlice        NativeGasSize = 3 // len(param[idx]) — slice of any
	SizeNumCallFrames   NativeGasSize = 4 // m.NumCallFrames(); idx ignored
	SizeReturnLen       NativeGasSize = 5 // len(return[idx]); POST-CALL only (legacy name kept; functionally identical to SizeLenSlice but reads the return stack)
	SizeSliceTotalBytes NativeGasSize = 6 // sum of inner element lengths for a []string or []byte-slice; works pre- and post-call
)

// NativeGasInfo is the per-function gas descriptor.
//
// Pre-call charge:  Base + Slope*N1/1024 + Slope2*N2/1024
//
//	(read off the call block before nativeBody)
//
// Post-call charge: PostBase + PostSlope*M1/1024 + PostSlope2*M2/1024
//
//	(read off the return stack after nativeBody)
//
// The two pre-call slopes are independent additive components (mirrors
// the `base + slopeP * P + slopeC * C` shape used for some CPU ops in
// op_gas_formulas.md). Typical use: Slope on len(slice) for per-element
// loop overhead, Slope2 on SizeSliceTotalBytes for per-byte marshal cost
// in []string params (chain.emit, chain/params.SetStrings, etc.).
//
// Bases are calibrated end-to-end through the dispatcher (Gno↔Go reflect
// + X_ work + return push). The /1024 mirrors machine.go:incrCPUBigInt's
// slopePerKb convention so sub-1 ns/byte slopes survive integer math.
type NativeGasInfo struct {
	Base      int64
	Slope     int64 // per 1024 units of N1
	SlopeIdx  int8  // -1 for flat; ignored when SlopeKind == SizeNumCallFrames
	SlopeKind NativeGasSize

	// Optional second pre-call slope, summed independently.
	// Zero Slope2 = unused.
	Slope2     int64
	Slope2Idx  int8
	Slope2Kind NativeGasSize

	// Optional post-call charge. Zero PostBase + zero PostSlope +
	// zero PostSlope2 = no post-charge (skipped via the gi-nil
	// shortcut returned by chargeNativeGas). PostSlopeIdx is the
	// stack offset from the top of m.Values (1 = topmost = last-pushed
	// return).
	PostBase      int64
	PostSlope     int64
	PostSlopeIdx  int8
	PostSlopeKind NativeGasSize

	// Optional second post-call slope, summed independently.
	PostSlope2     int64
	PostSlope2Idx  int8
	PostSlope2Kind NativeGasSize
}

// hasPost reports whether gi requires a post-call charge.
func (gi *NativeGasInfo) hasPost() bool {
	return gi.PostBase != 0 || gi.PostSlope != 0 || gi.PostSlope2 != 0
}

// nativeGasIndex maps "pkgPath\x00name" → calibrated descriptor. Populated
// at init time by stdlibs (or any package shipping natives) via
// RegisterNativeGas. Read-only after init.
var nativeGasIndex = map[string]*NativeGasInfo{}

// RegisterNativeGas installs a calibrated gas descriptor for a (pkgPath,
// name) native. Must be called before any Machine runs (i.e. from init()
// of the package shipping the native).
func RegisterNativeGas(pkgPath string, name Name, info *NativeGasInfo) {
	key := pkgPath + "\x00" + string(name)
	if _, exists := nativeGasIndex[key]; exists {
		panic(fmt.Sprintf("duplicate native gas registration for %s.%s", pkgPath, name))
	}
	nativeGasIndex[key] = info
}

// chargeNativeGas charges the pre-call cost for a native call. Returns
// the *NativeGasInfo when a post-call charge is also required (caller
// invokes chargeNativeGasPost after nativeBody returns); returns nil
// otherwise so the dispatcher can skip the post pass with a cheap
// nil-check.
//
// Behavior:
//   - Uverse builtins (no NativePkg, e.g. append/len/print) charge the
//     historical OpCPUCallNativeBody flat. Variable-cost ones like print
//     also self-charge (see uversePrint). TODO: extend the calibration
//     table to cover uverse natives too.
//   - Calibrated stdlibs charge Base + Slope*N1/1024 + Slope2*N2/1024.
//   - Stdlibs with no calibrated entry panic when a real GasMeter is
//     attached. This forces every new native to come with a benchmark.
//     Test/no-meter Machines silently fall through (no charge).
func (m *Machine) chargeNativeGas(fv *FuncValue) *NativeGasInfo {
	if fv.NativePkg == "" {
		m.incrCPU(OpCPUCallNativeBody)
		return nil
	}
	gi := nativeGasIndex[fv.NativePkg+"\x00"+string(fv.NativeName)]
	if gi == nil {
		if m.GasMeter == nil {
			// Test/no-meter Machine — silently no-op rather than panic
			// so unit tests that build minimal Machines without a gas
			// meter can still call natives.
			return nil
		}
		// Forcing function: every native must register a gas entry at
		// init time. Production stdlibs do this in
		// gnovm/stdlibs/native_gas.go; test stdlibs in
		// gnovm/tests/stdlibs/native_gas.go. A new native missing from
		// either trips this panic at first invocation, surfacing the
		// gap immediately rather than silently undercharging.
		panic(fmt.Sprintf("native %s.%s has no calibrated gas entry — register one in gnovm/stdlibs/native_gas.go (or tests/stdlibs/native_gas.go for test-only)",
			fv.NativePkg, fv.NativeName))
	}
	cost := gi.Base
	if gi.Slope != 0 {
		cost += gi.Slope * m.nativeSizeFromBlock(gi.SlopeKind, gi.SlopeIdx) / 1024
	}
	if gi.Slope2 != 0 {
		cost += gi.Slope2 * m.nativeSizeFromBlock(gi.Slope2Kind, gi.Slope2Idx) / 1024
	}
	m.incrCPU(cost)
	if !gi.hasPost() {
		return nil
	}
	return gi
}

// chargeNativeGasPost charges the post-call cost using gi (returned by
// chargeNativeGas). Reads return values off m.Values. Caller must guard
// with `if gi != nil` to skip cleanly when no post-charge is configured.
func (m *Machine) chargeNativeGasPost(gi *NativeGasInfo) {
	cost := gi.PostBase
	if gi.PostSlope != 0 {
		cost += gi.PostSlope * m.nativeSizeFromStack(gi.PostSlopeKind, gi.PostSlopeIdx) / 1024
	}
	if gi.PostSlope2 != 0 {
		cost += gi.PostSlope2 * m.nativeSizeFromStack(gi.PostSlope2Kind, gi.PostSlope2Idx) / 1024
	}
	m.incrCPU(cost)
}

// nativeSizeFromBlock extracts N from the call block (pre-call params).
func (m *Machine) nativeSizeFromBlock(kind NativeGasSize, idx int8) int64 {
	if kind == SizeNumCallFrames {
		return int64(m.NumCallFrames())
	}
	if idx < 0 {
		return 0
	}
	tv := &m.LastBlock().Values[idx]
	return nativeSizeOf(tv, kind, m.Store)
}

// nativeSizeFromStack extracts N from the return stack (post-call returns).
func (m *Machine) nativeSizeFromStack(kind NativeGasSize, idx int8) int64 {
	if kind == SizeNumCallFrames {
		return int64(m.NumCallFrames())
	}
	if idx < 0 {
		return 0
	}
	tv := m.PeekValue(int(idx))
	return nativeSizeOf(tv, kind, m.Store)
}

// nativeSizeOf computes the metric for a single TypedValue under the
// given kind. SizeReturnLen is treated identically to SizeLenSlice (the
// distinction is only documentary — which side of the call the kind is
// expected to be used).
func nativeSizeOf(tv *TypedValue, kind NativeGasSize, store Store) int64 {
	switch kind {
	case SizeSliceTotalBytes:
		return sumSliceInnerLen(tv, store)
	default: // SizeLenBytes / SizeLenString / SizeLenSlice / SizeReturnLen
		return int64(tv.GetLength())
	}
}

// sumSliceInnerLen sums the lengths of inner elements in a slice or
// array TypedValue. Used to compute total payload bytes of e.g.
// []string params for chain.emit and chain/params.SetStrings.
//
// For data-backed []byte arrays (av.Data != nil), returns the byte
// count directly — consistent with what GetLength reports.
func sumSliceInnerLen(tv *TypedValue, store Store) int64 {
	var list []TypedValue
	switch v := tv.V.(type) {
	case nil:
		return 0
	case *ArrayValue:
		if v.Data != nil {
			return int64(len(v.Data))
		}
		list = v.List
	case *SliceValue:
		base := v.GetBase(store)
		if base == nil {
			return 0
		}
		if base.Data != nil {
			return int64(v.Length)
		}
		end := v.Offset + v.Length
		if end > len(base.List) {
			end = len(base.List)
		}
		list = base.List[v.Offset:end]
	default:
		return 0
	}
	var total int64
	for i := range list {
		total += int64(list[i].GetLength())
	}
	return total
}
