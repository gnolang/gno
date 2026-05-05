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
	SizeFlat          NativeGasSize = 0 // no slope, no N
	SizeLenBytes      NativeGasSize = 1 // len(params[SlopeIdx]) — []byte
	SizeLenString     NativeGasSize = 2 // len(params[SlopeIdx]) — string
	SizeLenSlice      NativeGasSize = 3 // len(params[SlopeIdx]) — slice of any
	SizeNumCallFrames NativeGasSize = 4 // m.NumCallFrames(); SlopeIdx ignored
	SizeReturnLen     NativeGasSize = 5 // len(return at PostSlopeIdx) — POST-CALL only
)

// NativeGasInfo is the per-function gas descriptor.
//
// Pre-call charge:  Base  + Slope  * N / 1024  (read off block before nativeBody)
// Post-call charge: PostBase + PostSlope * N / 1024  (read off return stack after)
//
// Bases are calibrated end-to-end through the dispatcher (Gno↔Go reflect +
// X_ work + return push). The /1024 mirrors machine.go:incrCPUBigInt's
// slopePerKb convention so sub-1 ns/byte slopes survive integer math.
type NativeGasInfo struct {
	Base      int64
	Slope     int64 // per 1024 units of N
	SlopeIdx  int8  // -1 for flat
	SlopeKind NativeGasSize

	// Optional post-call charge. Zero PostBase + zero PostSlope = no
	// post-charge (skipped via the gi-nil-or-flat shortcut in
	// doOpCallNativeBody). PostSlopeIdx is the stack offset from the
	// top of m.Values (1 = topmost = last-pushed return). PostSlopeKind
	// must be SizeReturnLen if set.
	PostBase     int64
	PostSlope    int64
	PostSlopeIdx int8
	PostSlopeKind NativeGasSize
}

// hasPost reports whether gi requires a post-call charge.
func (gi *NativeGasInfo) hasPost() bool {
	return gi.PostBase != 0 || gi.PostSlope != 0
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
//   - Calibrated stdlibs charge Base + Slope*N/1024.
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
		// Uncalibrated native — fall back to the historical flat
		// charge. Many gno stdlib bindings (fmt.*, strings.*, errors.*,
		// etc.) aren't in the table yet; panicking would break every
		// realm that uses them. TODO: extend the calibration sweep to
		// cover these and re-introduce the panic as a build-time
		// forcing function once the table is comprehensive.
		m.incrCPU(OpCPUCallNativeBody)
		return nil
	}
	cost := gi.Base
	if gi.Slope != 0 {
		var n int64
		switch gi.SlopeKind {
		case SizeNumCallFrames:
			n = int64(m.NumCallFrames())
		default:
			if gi.SlopeIdx >= 0 {
				n = int64(m.LastBlock().Values[gi.SlopeIdx].GetLength())
			}
		}
		cost += gi.Slope * n / 1024
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
		var n int64
		switch gi.PostSlopeKind {
		case SizeReturnLen:
			n = int64(m.PeekValue(int(gi.PostSlopeIdx)).GetLength())
		}
		cost += gi.PostSlope * n / 1024
	}
	m.incrCPU(cost)
}
