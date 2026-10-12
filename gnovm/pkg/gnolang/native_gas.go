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
	SizeModExpWork      NativeGasSize = 7 // modular-exponentiation work from an (exponent, modulus) pair at param[idx], param[idx+1]; PRE-CALL only
)

// Modular exponentiation is the one native shipping today whose cost is a
// genuine product of two parameters rather than a sum, so it cannot be
// expressed by Base + Slope*N1 + Slope2*N2 over two independent lengths.
// SizeModExpWork collapses the product into a single metric that the ordinary
// single-slope machinery can then price linearly.
//
// big.Int.Exp performs one modular squaring per exponent bit (plus windowed
// multiplies), each costing O(words(modulus)^2) word-multiplications on top of
// a per-iteration overhead that does not scale with the modulus. So:
//
//	work = bits(exponent) * (modExpFloorWords + words(modulus)^2)
//
// The floor term is what makes a single metric safe across the whole operand
// range: measured against Go's big.Int, the per-exponent-bit cost is ~82ns at a
// 32-byte modulus but ~1274ns at 256 bytes — a pure words^2 metric implies a
// coefficient varying ~5x across that range, which either undercharges small
// moduli or overcharges large ones. modExpFloorWords is that overhead expressed
// in units of words^2, so one coefficient covers both regimes.
//
// bits(exponent) is a charged quantity, not the operand's true bit length: it
// needs no read of the operand's bytes, and matches EIP-2565's use of exp_len
// for the high part of the exponent. A zero-padded exponent is therefore
// overcharged, which is the safe direction — and a zero-filled exponent is
// exactly the cheap-to-build DoS operand, so charging for its length is correct
// rather than merely conservative. It is NOT simply 8*len(exp): big.Int runs
// two different routines either side of a one-word exponent and they differ
// ~2.5x per bit, so one linear coefficient cannot bound both. modExpWork
// documents where each applies.
//
// What this metric deliberately does not model is the cost of converting the
// operands and building the result, which is linear in len(modulus) and runs
// even when there is no exponentiation at all. That belongs to a second,
// independent slope on the modulus parameter — see the crypto/modexp row in
// gnovm/stdlibs/native_gas.go. Leaving it to Base instead makes the charge for
// a zero-length exponent independent of the modulus, which is a hole: the call
// still allocates and fills len(modulus) bytes.
const (
	// modExpFloorWords is the per-exponent-bit overhead of big.Int.Exp
	// expressed in words(modulus)^2 units. Derived as the ratio of the
	// modulus-independent to the quadratic coefficient; being a ratio it
	// should carry across hardware better than an absolute ns figure.
	modExpFloorWords = 65

	// modExpWordBytes is one machine word of exponent, the threshold that
	// decides which of big.Int's two exponentiation routines runs. expNN
	// takes the windowed Montgomery path only when the exponent exceeds one
	// word (nat.go: `if len(y) > 1 && !slow`); at or below that it runs the
	// generic square-and-multiply loop, which reduces by full division on
	// every iteration rather than in Montgomery form.
	modExpWordBytes = 8

	// The Montgomery path's operation count, read off nat.go's
	// expNNMontgomery rather than fitted. Units are modular multiplications;
	// one Montgomery step is two of them (the product, then the REDC
	// reduction), so each count below is doubled.
	//
	//	powers table    2 + 14 loop iterations = 16 steps
	//	RR setup        one 2n-by-n division   ~  2 steps
	//	final convert                          =  1 step
	//	main loop       words(exp) x 16 windows x (4 squarings + 1 multiply)
	//
	// Being counts of a fixed algorithm rather than a curve fit, these need no
	// re-derivation on new hardware — only the slope converting units to
	// nanoseconds does. Over the recorded grid the Montgomery points hold to
	// 1.39x of each other across a 128x range of exponent sizes.
	modExpMontSetup   = 2 * (16 + 2 + 1)
	modExpMontPerWord = 2 * (16 * 5)

	// modExpGenericPerBit is the same count for the generic loop, which runs
	// one squaring, one multiply and one full division per exponent bit. The
	// first two are ~1.5 units; division is the rest. It is calibrated rather
	// than counted because nat.div's cost per word is set by hardware divide
	// latency, not by a multiplication count — the algorithmic estimate is
	// ~3.25 and the benches want 5.
	modExpGenericPerBit = 5

	// modExpWorkSaturation clamps the metric. Gas is charged before the
	// native runs, so this function sees operand lengths before
	// crypto/modexp rejects oversized ones — and maxAllocTx allows a 500MB
	// slice, whose product would wrap int64. Any value this large already
	// prices four orders of magnitude beyond the block gas limit, so clamping
	// can never make an expensive call look cheap.
	//
	// The binding constraint is not this function but chargeNativeGas's
	// unguarded `gi.Slope * N`. This clamp leaves room for a slope up to 2^19
	// before that product leaves int64, where a negative cost would refund gas
	// for the most expensive call possible.
	// TestModExpWorkSlopeProductFitsInt64 pins the headroom.
	modExpWorkSaturation = 1 << 44
)

// modExpWork counts the modular multiplications big.Int.Exp will perform for
// the given operand lengths, in units of one multiplication at words(modulus).
// It saturates rather than wrapping.
//
// It is a count of a known algorithm, not a curve fitted to a benchmark. Every
// branch expNN takes is decided by the operand lengths alone, so the count is
// exact up to the cost model for a single multiplication — which is where
// modExpFloorWords comes in, and which is the only part needing re-derivation
// on new hardware. A fitted slope has to bound both of big.Int's routines with
// one coefficient and cannot: they differ ~2.5x per exponent bit, so any single
// line either underprices sub-word exponents or overcharges ordinary ones
// several fold.
//
// Clamping each factor before multiplying is what keeps this overflow-safe. A
// factor can only be clamped at a size where the metric already saturates, so
// clamping never lowers the result below what the true count would charge.
// maxOperandLenForWork also keeps (n+7) from wrapping — unclamped, that
// addition silently yields a negative word count at n=MaxInt64.
func modExpWork(expLen, modLen int64) int64 {
	if expLen <= 0 || modLen <= 0 {
		// A zero-length exponent means exponent 0 (Exp returns immediately);
		// a zero-length modulus short-circuits in X_modExp. Neither runs the
		// exponentiation loop, so neither owes anything on this metric — but
		// both still pay the row's per-modulus-byte slope, which is what
		// covers converting the operands and building the result.
		return 0
	}
	const maxOperandLenForWork = 1 << 28
	words := (min(modLen, maxOperandLenForWork) + 7) / 8
	unit := words*words + modExpFloorWords

	// One machine word of exponent decides the routine. At or below it expNN
	// runs the generic loop, one squaring + multiply + full division per
	// exponent bit. Above it expNNMontgomery runs, and its loop walks whole
	// words — `for i := len(y)-1; i >= 0; i--`, all 16 windows of each,
	// including a partly-filled top word — so a 9-byte exponent does the same
	// work as a 16-byte one and must not be charged less.
	var units int64
	if expLen <= modExpWordBytes {
		units = modExpGenericPerBit * 8 * expLen
	} else {
		ew := (min(expLen, maxOperandLenForWork) + modExpWordBytes - 1) / modExpWordBytes
		units = modExpMontSetup + modExpMontPerWord*ew
	}
	if units > modExpWorkSaturation/unit {
		return modExpWorkSaturation
	}
	return units * unit
}

// ModExpWork exposes the metric to gnovm/stdlibs, whose gas table owns the
// crypto/modexp row and whose guard test asserts against it. One definition
// means a change to the model cannot leave that test checking a different
// quantity than the runtime charges.
func ModExpWork(expLen, modLen int64) int64 { return modExpWork(expLen, modLen) }

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
	if kind == SizeModExpWork {
		// Reads a pair: exponent at idx, modulus at idx+1.
		vals := m.LastBlock().Values
		if int(idx)+1 >= len(vals) {
			return 0
		}
		return modExpWork(int64(vals[idx].GetLength()), int64(vals[idx+1].GetLength()))
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
	case SizeModExpWork:
		// Reached only via nativeSizeFromStack, i.e. from a post-call slope.
		// The metric needs a parameter pair, and the return stack has no such
		// pair; falling through to GetLength() would silently price on one
		// return value's byte length instead. Fail loudly rather than charge
		// the wrong quantity.
		panic("SizeModExpWork is pre-call only; it cannot be used as a post-call slope kind")
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
		end := min(v.Offset+v.Length, len(base.List))
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
