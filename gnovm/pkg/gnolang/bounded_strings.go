package gnolang

import (
	"fmt"
	"math"
	"strconv"

	"github.com/cockroachdb/apd/v3"
)

// Bounded printers — single source of truth for size-bounded
// rendering of Gno values, exceptions, and stacktraces.
//
// These are used ONLY on the validator-side panic-recovery path
// (when Machine.BoundedPanicRender is set). The existing Sprint /
// String / ProtectedString methods are unchanged and continue to
// be used by filetests, REPL, gno run, debug output, and any other
// trusted-context consumer.
//
// DESIGN INVARIANTS:
//
// 1. Internal helpers take NO *Machine parameter. User-defined
//    .String() / .Error() methods on Gno values are NEVER invoked
//    from the bounded path — those would re-enter the VM and could
//    be unbounded themselves. Structural render only.
//
// 2. Output bytes ≤ caller-supplied max (modulo a 3-byte "..."
//    suffix on truncation).
//
// 3. Transient memory during rendering is also bounded: leaf
//    renderers length-pre-check before calling Quote/String on
//    types whose intermediate allocation is proportional to value
//    size (StringValue, BigintValue, BigdecValue).
//
// 4. Composite walks (Array/Slice/Map/Struct) cap iteration count
//    AND pass per-element budget down. After the cap, an elision
//    marker like "<...K more>" is emitted.
//
// 5. Recursion depth is capped at MaxValueRenderDepth.

const (
	// BoundedRenderBytes is the per-call output cap for bounded
	// renderings (output size ≤ this; +3 for "..." suffix).
	BoundedRenderBytes = 1024

	// MaxStacktraceFrames is the per-stacktrace frame cap. Must be
	// kept ≤ maxStacktraceSize (the m.Stacktrace() trim).
	MaxStacktraceFrames = 16

	// MaxValueRenderDepth bounds composite-recursion depth.
	MaxValueRenderDepth = 8

	// MaxCompositeChildren bounds rendered children per composite
	// (slice/array/map/struct).
	MaxCompositeChildren = 32

	// MaxByteArrayBytes is the cap on hex-preview length for
	// ArrayValue.Data byte arrays.
	MaxByteArrayBytes = 256
)

const ellipsisLen = 3 // for "..." suffix

// writeBoundedString appends s to w, capped at rem bytes. If s
// exceeds rem, writes a truncated prefix + "..." (or just the
// prefix if rem < ellipsisLen).
func writeBoundedString(w *boundedBuf, s string, rem int) {
	if len(s) <= rem {
		w.WriteString(s)
		return
	}
	if rem < ellipsisLen {
		w.WriteString(s[:rem])
		return
	}
	w.WriteString(s[:rem-ellipsisLen])
	w.WriteString("...")
}

// boundedBuf is a length-bounded io.Writer-like buffer used by the
// bounded printers. Once cap is reached, additional writes are
// no-ops; .Truncated() reports overflow.
type boundedBuf struct {
	buf       []byte
	cap       int
	truncated bool
}

func newBoundedBuf(n int) *boundedBuf {
	if n < 0 {
		n = 0
	}
	return &boundedBuf{cap: n}
}

// Write implements io.Writer. Always returns (len(p), nil) so
// callers like fmt.Fprintf don't error on truncated writes.
// Truncation is silent; query .Truncated() to detect.
func (b *boundedBuf) Write(p []byte) (int, error) {
	avail := b.cap - len(b.buf)
	if avail <= 0 {
		b.truncated = true
		return len(p), nil
	}
	if len(p) <= avail {
		b.buf = append(b.buf, p...)
		return len(p), nil
	}
	b.buf = append(b.buf, p[:avail]...)
	b.truncated = true
	return len(p), nil
}

func (b *boundedBuf) WriteString(s string) (int, error) {
	avail := b.cap - len(b.buf)
	if avail <= 0 {
		b.truncated = true
		return len(s), nil
	}
	if len(s) <= avail {
		b.buf = append(b.buf, s...)
		return len(s), nil
	}
	b.buf = append(b.buf, s[:avail]...)
	b.truncated = true
	return len(s), nil
}

func (b *boundedBuf) WriteByte(c byte) error {
	if len(b.buf) >= b.cap {
		b.truncated = true
		return nil
	}
	b.buf = append(b.buf, c)
	return nil
}

// Remaining returns the number of bytes left before the cap (clamped
// at 0).
func (b *boundedBuf) Remaining() int {
	r := b.cap - len(b.buf)
	if r < 0 {
		return 0
	}
	return r
}

// String returns the buffer contents, with a "..." suffix appended
// if any write was truncated.
func (b *boundedBuf) String() string {
	if b.truncated {
		return string(b.buf) + "..."
	}
	return string(b.buf)
}

// Truncated reports whether any write exceeded the cap.
func (b *boundedBuf) Truncated() bool {
	return b.truncated
}

// ----------------------------------------
// Public API

// BoundedSprintTV renders tv into ≤ max bytes (with "..." suffix
// on truncation).
//
// For primitive types (StringValue, BigintValue, BigdecValue, etc.)
// this length-pre-checks before any expensive render to avoid
// transient memory blowup proportional to value size. These leaf
// renderers do NOT invoke any user-defined .String()/.Error()
// methods.
//
// For composite types (struct, array, etc.), if m is non-nil it
// dispatches to tv.Sprint(m) and truncates the result. tv.Sprint(m)
// is gas-metered (user code runs through m.Eval), so bounded by
// the gas budget the adversary paid to construct the value. This
// preserves friendly rendering of user-defined Error()/String()
// methods.
//
// If m is nil, composite render is purely structural — no user
// methods invoked. Use m=nil only when you can't trust the gas
// budget (e.g., re-entry from a deep recovery path).
func BoundedSprintTV(tv TypedValue, m *Machine, lim int) string {
	// Composite/object types with non-nil m: defer to gas-metered
	// tv.Sprint(m) which dispatches to user-defined Error()/String()
	// methods through m.Eval. Result is truncated to fit.
	//
	// Decision is here (not inside boundedSprintTV) so internal
	// helpers stay m-free per the design invariant.
	if m != nil && tv.T != nil && tv.V != nil {
		if _, isPrim := baseOf(tv.T).(PrimitiveType); !isPrim {
			if s, ok := boundedUserSprint(tv, m, lim); ok {
				return s
			}
			// User Sprint exceeded the transient alloc cap or
			// otherwise panicked — fall through to structural
			// render below.
		}
	}
	w := newBoundedBuf(lim)
	boundedSprintTV(w, tv, 0)
	return w.String()
}

// boundedUserSprint runs tv.Sprint(m) under a tight transient
// allocator cap so a user-defined String()/Error() that allocates
// proportional to the (gas-paid) value size still can't blow up
// peak heap during recovery rendering.
//
// Returns (rendered, true) on success; ("", false) if the user
// method panicked (alloc cap, OOG, or other) — the caller should
// fall through to structural render.
//
// All panics are swallowed (including OOG): we're already on the
// recovery path for some original panic, and the gas meter is
// charged before any panic propagates, so swallowing OOG here
// loses no gas accounting. The trade-off is error classification:
// an adversarial panic value whose Error() OOGs surfaces as
// "VM panic: <structural>" rather than ErrOutOfGas. The original
// panic info is preserved, which is the more useful signal.
func boundedUserSprint(tv TypedValue, m *Machine, lim int) (s string, ok bool) {
	if m.Alloc == nil {
		// No allocator means no per-allocation cap; let it run.
		// (Filetests, REPL, etc. don't take this branch because
		// the m=nil filter at BoundedSprintTV's entry skips us.)
		s = tv.Sprint(m)
		ok = true
		return
	}
	// Snapshot and tighten. boundedRenderAllocCap allows enough
	// headroom for nested struct rendering at our lim cap; far
	// less than block-gas worth.
	const boundedRenderAllocCap = 64 * 1024
	savedMax := m.Alloc.maxBytes
	savedBytes := m.Alloc.bytes
	m.Alloc.maxBytes = savedBytes + boundedRenderAllocCap
	defer func() {
		m.Alloc.maxBytes = savedMax
		m.Alloc.bytes = savedBytes
		if r := recover(); r != nil {
			s = ""
			ok = false
		}
	}()
	rendered := tv.Sprint(m)
	if len(rendered) <= lim {
		s, ok = rendered, true
		return
	}
	if lim < ellipsisLen {
		s, ok = rendered[:lim], true
		return
	}
	s, ok = rendered[:lim-ellipsisLen]+"...", true
	return
}

// BoundedSprintException renders e.Value (the head exception's
// value) into ≤ max bytes. Does NOT walk e.Previous — chained
// panics are rendered head-only in the bounded path.
func BoundedSprintException(e *Exception, m *Machine, lim int) string {
	if e == nil {
		return "<nil>"
	}
	return BoundedSprintTV(e.Value, m, lim)
}

// BoundedStacktrace renders s into ≤ max bytes. Frame format:
//
//	[defer ]<funcName-or-(anonymous)> at <pkgpath>/<file>:<line>
//
// FuncName comes from StacktraceCall.FuncName (pre-rendered at
// capture time). Caps frames at MaxStacktraceFrames; emits a
// "... K of N frames elided ..." marker when frames are dropped.
func BoundedStacktrace(s Stacktrace, lim int) string {
	w := newBoundedBuf(lim)
	totalCalls := len(s.Calls)
	visit := totalCalls
	if visit > MaxStacktraceFrames {
		visit = MaxStacktraceFrames
	}
	for i := 0; i < visit; i++ {
		call := s.Calls[i]
		var line int
		if i == 0 {
			line = s.LastLine
		} else {
			// Defensive: don't panic if Calls[i-1].CallExpr is nil
			// (synthetic frames or future schema).
			if cx := s.Calls[i-1].CallExpr; cx != nil {
				line = cx.GetLine()
			}
		}
		boundedSprintFrame(w, call, line)
	}
	// Combined elision count: original m.Stacktrace() trims to
	// maxStacktraceSize; if more frames pre-existed, they are
	// reflected in s.NumFramesElided. Plus our visit-cap drops
	// (totalCalls - visit) frames.
	elided := s.NumFramesElided + (totalCalls - visit)
	if elided > 0 {
		fmt.Fprintf(w, "... %d frame(s) elided ...\n", elided)
	}
	return w.String()
}

// BoundedExceptionStacktrace renders the head exception's value +
// stacktrace within max bytes. Does NOT walk e.Previous; bounded
// path renders head-only.
func BoundedExceptionStacktrace(m *Machine, lim int) string {
	if m == nil || m.Exception == nil {
		return ""
	}
	w := newBoundedBuf(lim)
	w.WriteString("panic: ")
	// Reserve roughly half for the value, half for the stacktrace.
	half := lim / 2
	if half < 64 {
		half = lim // tiny budgets — give it all to the value
	}
	w.WriteString(BoundedSprintTV(m.Exception.Value, m, half))
	w.WriteString("\n")
	w.WriteString(BoundedStacktrace(m.Exception.Stacktrace, w.Remaining()))
	return w.String()
}

// ----------------------------------------
// Internal helpers — NO *Machine parameter.

// boundedSprintTV dispatches on the typed value's kind: primitives
// route through length-pre-checked leaf renderers; composites
// render structurally via boundedSprintValue. NEVER invokes
// user-defined .String()/.Error() methods — the m=non-nil branch
// is handled by the public BoundedSprintTV entry point so internal
// helpers stay m-free.
func boundedSprintTV(w *boundedBuf, tv TypedValue, depth int) {
	if w.Remaining() <= 0 {
		return
	}
	if tv.T == nil {
		if tv.V == nil {
			w.WriteString("undefined")
			return
		}
		boundedSprintValue(w, tv.V, depth)
		return
	}
	if _, isPrim := baseOf(tv.T).(PrimitiveType); isPrim {
		boundedSprintPrimitiveTV(w, tv)
		return
	}
	if tv.V == nil {
		w.WriteString("nil")
		return
	}
	boundedSprintValue(w, tv.V, depth)
}

// boundedSprintPrimitiveTV renders a primitive TypedValue using
// the m-independent GetXxx accessors. Strings and big numbers go
// through their bounded leaf renderers.
func boundedSprintPrimitiveTV(w *boundedBuf, tv TypedValue) {
	pt, _ := baseOf(tv.T).(PrimitiveType)
	switch pt {
	case BoolType, UntypedBoolType:
		fmt.Fprintf(w, "%t", tv.GetBool())
	case StringType, UntypedStringType:
		// Match TypedValue.Sprint: emit raw bytes (no quotes, no
		// escape). Pre-truncate to avoid huge intermediate.
		writeBoundedString(w, tv.GetString(), w.Remaining())
	case IntType:
		fmt.Fprintf(w, "%d", tv.GetInt())
	case Int8Type:
		fmt.Fprintf(w, "%d", tv.GetInt8())
	case Int16Type:
		fmt.Fprintf(w, "%d", tv.GetInt16())
	case Int32Type, UntypedRuneType:
		fmt.Fprintf(w, "%d", tv.GetInt32())
	case Int64Type:
		fmt.Fprintf(w, "%d", tv.GetInt64())
	case UintType:
		fmt.Fprintf(w, "%d", tv.GetUint())
	case Uint8Type:
		fmt.Fprintf(w, "%d", tv.GetUint8())
	case Uint16Type:
		fmt.Fprintf(w, "%d", tv.GetUint16())
	case Uint32Type:
		fmt.Fprintf(w, "%d", tv.GetUint32())
	case Uint64Type:
		fmt.Fprintf(w, "%d", tv.GetUint64())
	case Float32Type:
		fmt.Fprintf(w, "%g", math.Float32frombits(tv.GetFloat32()))
	case Float64Type:
		fmt.Fprintf(w, "%g", math.Float64frombits(tv.GetFloat64()))
	case UntypedBigintType:
		// Guard against tv.V == nil: GetBigInt does a typed
		// assertion on tv.V and would panic on a zero-value slot.
		if tv.V == nil {
			w.WriteString("<nil>")
		} else if bi := tv.GetBigInt(); bi != nil {
			boundedSprintBigInt(w, bi)
		} else {
			w.WriteString("<nil>")
		}
	case UntypedBigdecType:
		if tv.V == nil {
			w.WriteString("<nil>")
		} else if bd := tv.GetBigDec(); bd != nil {
			boundedSprintBigDec(w, bd)
		} else {
			w.WriteString("<nil>")
		}
	default:
		fmt.Fprintf(w, "<%v>", pt)
	}
}

// boundedSprintValue renders a Value via type-switch on its
// concrete Go type. Default arm renders type name only — never
// dispatches to an unknown type's String() method.
func boundedSprintValue(w *boundedBuf, v Value, depth int) {
	if w.Remaining() <= 0 {
		return
	}
	if v == nil {
		w.WriteString("<nil>")
		return
	}
	switch x := v.(type) {
	case StringValue:
		boundedSprintStringValue(w, x)
	case BigintValue:
		boundedSprintBigInt(w, x.V)
	case BigdecValue:
		boundedSprintBigDec(w, x.V)
	case *ArrayValue:
		boundedSprintArrayValue(w, x, depth)
	case *SliceValue:
		boundedSprintSliceValue(w, x, depth)
	case *StructValue:
		boundedSprintStructValue(w, x, depth)
	case *MapValue:
		boundedSprintMapValue(w, x, depth)
	case *FuncValue:
		boundedSprintFuncValue(w, x)
	case *BoundMethodValue:
		boundedSprintBoundMethodValue(w, x)
	case PointerValue:
		boundedSprintPointerValue(w, x, depth)
	case *PackageValue:
		fmt.Fprintf(w, "<package %s>", x.PkgPath)
	case TypeValue:
		// Avoid recursive Type.String(); show kind only.
		if x.Type == nil {
			w.WriteString("<type nil>")
		} else {
			fmt.Fprintf(w, "<type %s>", x.Type.Kind())
		}
	default:
		fmt.Fprintf(w, "<%T>", v)
	}
}

// ----------------------------------------
// Leaf renderers (size-bounded by pre-checks).

// boundedSprintStringValue renders sv as a Go-quoted string,
// bounded by w.Remaining() bytes. To avoid allocating the full
// strconv.Quote of an arbitrarily-large source first, the source
// is pre-truncated to a size that, after Quote expansion, fits
// the cap.
func boundedSprintStringValue(w *boundedBuf, sv StringValue) {
	rem := w.Remaining()
	if rem <= 0 {
		return
	}
	s := string(sv)
	// strconv.Quote can expand by up to ~6× for non-ASCII /
	// non-printable bytes (`\u00XX` = 6 bytes per source byte
	// worst case). Pre-truncate to bound the intermediate.
	const expandFactor = 6
	preTruncCap := rem / expandFactor
	if preTruncCap < 1 {
		preTruncCap = 1
	}
	srcWasTruncated := false
	if len(s) > preTruncCap {
		s = s[:preTruncCap]
		srcWasTruncated = true
	}
	q := strconv.Quote(s)
	if srcWasTruncated {
		// Replace the closing `"` with `..."` to indicate truncation.
		if len(q) > 0 && q[len(q)-1] == '"' {
			q = q[:len(q)-1] + `..."`
		} else {
			q += "..."
		}
	}
	// Quote may produce more than rem bytes (rare given pre-trunc) —
	// writeBoundedString truncates with "..." if so.
	writeBoundedString(w, q, rem)
}

// boundedSprintBigInt renders a *big.Int with a length pre-check
// to avoid allocating a giant decimal string for huge numbers.
func boundedSprintBigInt(w *boundedBuf, bi interface {
	BitLen() int
	String() string
}) {
	rem := w.Remaining()
	if rem <= 0 {
		return
	}
	if bi == nil {
		w.WriteString("<nil>")
		return
	}
	// 1 decimal digit ≈ 3.32 bits. Use ×3 conservatively.
	if bi.BitLen() > rem*3 {
		fmt.Fprintf(w, "<bigint, bits=%d>", bi.BitLen())
		return
	}
	writeBoundedString(w, bi.String(), rem)
}

// boundedSprintBigDec renders a *apd.Decimal. Pre-checks the
// coefficient's bit-length before allocating the full decimal
// string — bd.String() is O(coeff size) and would otherwise let
// an attacker who grew the coefficient via runtime arithmetic
// (e.g. apd's unlimited-precision Add at op_binary.go) burn
// unmetered CPU/memory on this rendering path.
//
// fmtE/fmtF output length is dominated by the coefficient's
// decimal-digit count (≈ BitLen / 3.32). The Exponent adds ≤ ~12
// bytes (sign + int32 digits + "E"); the fmtF zero-pad path is
// itself capped by apd's adjExponentLimit rule, so no Exponent
// term is needed in the gate.
func boundedSprintBigDec(w *boundedBuf, bd *apd.Decimal) {
	rem := w.Remaining()
	if rem <= 0 {
		return
	}
	if bd == nil {
		w.WriteString("<nil>")
		return
	}
	// Zero coefficient — render directly. apd's fmtF zero-pad path
	// allocates up to ~|Exponent| bytes (capped at ~2000 by apd's
	// adjExponentLimit) for negative-Exponent zero values; we
	// sidestep it since "0" carries the same numeric information.
	if bd.Coeff.BitLen() == 0 {
		w.WriteString("0")
		return
	}
	// 1 decimal digit ≈ 3.32 bits. Use ×3 conservatively
	// (matches boundedSprintBigInt).
	if bd.Coeff.BitLen() > rem*3 {
		fmt.Fprintf(w, "<bigdec, bits=%d>", bd.Coeff.BitLen())
		return
	}
	writeBoundedString(w, bd.String(), rem)
}

// ----------------------------------------
// Composite renderers.

func boundedSprintArrayValue(w *boundedBuf, av *ArrayValue, depth int) {
	if av == nil {
		w.WriteString("<nil>")
		return
	}
	if depth >= MaxValueRenderDepth {
		w.WriteString("[...]")
		return
	}
	// Byte-array path.
	if av.Data != nil {
		n := len(av.Data)
		rem := w.Remaining()
		// "0x" prefix + 2 hex chars per byte + "...total N bytes" suffix
		const suffixOverhead = 24 // generous — len("...total NNNNNN bytes") ≈ 22
		previewBytes := MaxByteArrayBytes
		if previewBytes > n {
			previewBytes = n
		}
		if previewBytes*2+2+suffixOverhead > rem {
			previewBytes = (rem - 2 - suffixOverhead) / 2
			if previewBytes < 0 {
				previewBytes = 0
			}
		}
		w.WriteString("0x")
		for i := 0; i < previewBytes; i++ {
			fmt.Fprintf(w, "%02X", av.Data[i])
		}
		if n > previewBytes {
			fmt.Fprintf(w, "...total %d bytes", n)
		}
		return
	}
	boundedSprintList(w, av.List, depth, "[", "]")
}

func boundedSprintSliceValue(w *boundedBuf, sv *SliceValue, depth int) {
	if sv == nil {
		w.WriteString("<nil>")
		return
	}
	if depth >= MaxValueRenderDepth {
		w.WriteString("[...]")
		return
	}
	// Pull the underlying list slice.
	if av, ok := sv.Base.(*ArrayValue); ok {
		if av.Data != nil {
			// data slice — render as byte array
			boundedSprintArrayValue(w, av, depth)
			return
		}
		// index range
		from := sv.Offset
		to := sv.Offset + sv.Length
		if to > len(av.List) {
			to = len(av.List)
		}
		if from < 0 {
			from = 0
		}
		boundedSprintList(w, av.List[from:to], depth, "[", "]")
		return
	}
	w.WriteString("<slice>")
}

func boundedSprintList(w *boundedBuf, list []TypedValue, depth int, open, closer string) {
	n := len(list)
	if n == 0 {
		w.WriteString(open)
		w.WriteString(closer)
		return
	}
	visit := n
	if visit > MaxCompositeChildren {
		visit = MaxCompositeChildren
	}
	w.WriteString(open)
	for i := 0; i < visit; i++ {
		if i > 0 {
			w.WriteString(", ")
		}
		nLeft := visit - i
		budget := w.Remaining() / nLeft
		if budget <= 0 {
			fmt.Fprintf(w, "<...%d more>", visit-i)
			w.WriteString(closer)
			return
		}
		sub := newBoundedBuf(budget)
		boundedSprintTV(sub, list[i], depth+1)
		w.WriteString(sub.String())
	}
	if n > visit {
		fmt.Fprintf(w, ", <...%d more>", n-visit)
	}
	w.WriteString(closer)
}

func boundedSprintStructValue(w *boundedBuf, sv *StructValue, depth int) {
	if sv == nil {
		w.WriteString("<nil>")
		return
	}
	if depth >= MaxValueRenderDepth {
		w.WriteString("{...}")
		return
	}
	boundedSprintList(w, sv.Fields, depth, "{", "}")
}

func boundedSprintMapValue(w *boundedBuf, mv *MapValue, depth int) {
	if mv == nil {
		w.WriteString("<nil>")
		return
	}
	if depth >= MaxValueRenderDepth {
		w.WriteString("{...}")
		return
	}
	if mv.List == nil || mv.List.Size == 0 {
		w.WriteString("{}")
		return
	}
	n := mv.List.Size
	visit := n
	if visit > MaxCompositeChildren {
		visit = MaxCompositeChildren
	}
	w.WriteString("{")
	cur := mv.List.Head
	for i := 0; i < visit && cur != nil; i++ {
		if i > 0 {
			w.WriteString(", ")
		}
		nLeft := visit - i
		// Each entry is "key:value" — split budget in 2 for key and value.
		budget := w.Remaining() / nLeft
		if budget <= 0 {
			fmt.Fprintf(w, "<...%d more>", visit-i)
			w.WriteString("}")
			return
		}
		half := budget / 2
		if half < 1 {
			half = 1
		}
		ksub := newBoundedBuf(half)
		boundedSprintTV(ksub, cur.Key, depth+1)
		w.WriteString(ksub.String())
		w.WriteString(":")
		vsub := newBoundedBuf(budget - half - 1)
		boundedSprintTV(vsub, cur.Value, depth+1)
		w.WriteString(vsub.String())
		cur = cur.Next
	}
	if n > visit {
		fmt.Fprintf(w, ", <...%d more>", n-visit)
	}
	w.WriteString("}")
}

func boundedSprintFuncValue(w *boundedBuf, fv *FuncValue) {
	if fv == nil {
		w.WriteString("<nil>")
		return
	}
	name := string(fv.Name)
	if name == "" {
		w.WriteString("<func>")
		return
	}
	fmt.Fprintf(w, "<func %s>", name)
}

func boundedSprintBoundMethodValue(w *boundedBuf, bmv *BoundMethodValue) {
	if bmv == nil || bmv.Func == nil {
		w.WriteString("<nil>")
		return
	}
	name := string(bmv.Func.Name)
	if name == "" {
		w.WriteString("<bound-method>")
		return
	}
	fmt.Fprintf(w, "<bound-method %s>", name)
}

func boundedSprintPointerValue(w *boundedBuf, pv PointerValue, depth int) {
	// Render shape only — no recursion into target value to avoid
	// pointer cycles.
	if pv.TV == nil {
		w.WriteString("<*nil>")
		return
	}
	if pv.TV.T == nil {
		w.WriteString("<*?>")
		return
	}
	fmt.Fprintf(w, "<*%s>", pv.TV.T.Kind())
}

// ----------------------------------------
// Frame name + line render.

// stacktraceFuncName produces a frame's display name including
// receiver type prefix for methods. Bounded by design: only uses
// *DeclaredType pkgpath+name (each ≤256 by memfile validation);
// falls back to Kind().String() for unusual receivers.
//
// Output examples:
//
//	"Inc"                            -- free function
//	""                               -- anonymous function (caller renders as "(anonymous)")
//	"gno.land/r/x.Counter.Inc"       -- value-receiver method
//	"(*gno.land/r/x.Counter).Inc"    -- pointer-receiver method
func stacktraceFuncName(fr *Frame) string {
	if fr.Func == nil {
		return ""
	}
	name := string(fr.Func.Name)
	if !fr.Receiver.IsDefined() {
		return name
	}
	rt := fr.Receiver.T
	if rt == nil {
		return name
	}
	if pt, ok := rt.(*PointerType); ok {
		if dt, ok := pt.Elt.(*DeclaredType); ok {
			return "(*" + dt.PkgPath + "." + string(dt.Name) + ")." + name
		}
		return "(*<" + pt.Elt.Kind().String() + ">)." + name
	}
	if dt, ok := rt.(*DeclaredType); ok {
		return dt.PkgPath + "." + string(dt.Name) + "." + name
	}
	return "<" + rt.Kind().String() + ">." + name
}

// boundedSprintFrame writes a single frame line:
//
//	[defer ]<funcName-or-(anonymous)> at <pkgpath>/<file>:<line>
func boundedSprintFrame(w *boundedBuf, sc StacktraceCall, line int) {
	if sc.IsDefer {
		w.WriteString("defer ")
	}
	if sc.FuncName == "" {
		w.WriteString("(anonymous)")
	} else {
		w.WriteString(sc.FuncName)
	}
	w.WriteString(" at ")
	w.WriteString(sc.FuncLoc.PkgPath)
	if sc.FuncLoc.File != "" {
		w.WriteByte('/')
		w.WriteString(sc.FuncLoc.File)
	}
	if line == -1 {
		w.WriteString(":native\n")
	} else {
		fmt.Fprintf(w, ":%d\n", line)
	}
}

// truncateForLog exposes the same shape as the keeper's truncate
// for tests that want to verify the suffix behavior.
func truncateForLog(s string, lim int) string {
	if len(s) <= lim {
		return s
	}
	if lim < 3 {
		return s[:lim]
	}
	return s[:lim-3] + "..."
}
