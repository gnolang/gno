package gnolang

import (
	"fmt"
	"io"
	"math"
	"reflect"
	"strconv"
	"sync"

	"github.com/gnolang/gno/tm2/pkg/store"
)

// protectedWriter is the dispatch contract for the streaming
// counterpart of ProtectedString. The recursion threads (w, seen),
// each parameter doing one job: w handles where bytes go and how
// they're accounted; seen handles cycle detection.
type protectedWriter interface {
	WriteProtected(w *meteredWriter, seen *seenValues)
}

// meteredWriterBufSize is the size of meteredWriter's internal buffer. Gas
// is charged once per flush, so output cost is proportional to flushes
// (~bytes/meteredWriterBufSize) rather than to the number of individual
// Write calls.
const meteredWriterBufSize = 1024

// streamOutputGasPerByte is the gas charged per byte of formatted output
// flushed by meteredWriter. It is calibrated SEPARATELY from allocGas:
// allocGas prices a Go heap allocation (malloc + zero-fill, a concave curve),
// whereas output cost is ~linear in bytes — the formatter runs
// strconv.Append*/copy work proportional to the bytes produced, in native Go
// that no gno-opcode (incrCPU) gas covers. Pricing output through allocGas
// (its 1 KiB chunk ≈ 0.24 gas/byte) therefore under-charged the formatter by
// ~16x.
//
// Reference convention: 1 gas = 1 ns on the calibration box (see
// gnovm/cmd/calibrate). Measured ~3.0 ns/byte asymptotically for the
// format+flush path (BenchmarkStreamOutputProduce_*), scaled to reference via
// the 1 KiB-alloc anchor (241 gas / ~196 ns local ≈ 1.23x) → ~3.7 gas/byte,
// rounded up to 4 for a small margin. See values_string_gas_test.go.
const streamOutputGasPerByte = 4

// streamOutputGas is the gas charged for flushing n bytes of formatted output.
func streamOutputGas(n int) int64 {
	return int64(n) * streamOutputGasPerByte
}

// meteredWriter is a bufio.Writer-style buffer that meters output gas
// once per flush instead of once per write. Bytes accumulate in buf;
// when it fills (or a formatter needs headroom it can't fit) the buffer
// is flushed to parent and gas is charged for the flushed bytes —
// streamOutputGas(n), the separately-calibrated per-output-byte cost — via
// the gas meter directly. Charging per flush (not per write) keeps the
// accounting independent of how many small WriteByte/WriteString calls fed
// the buffer, and removes the need for a separate scratch array, since
// strconv.Append* formatters write straight into the buffer tail after
// reserving space.
//
// The writer does NOT hold an *Allocator and never allocates: output is
// a transient sink (bytes.Buffer / strings.Builder / io.Discard) the GC
// never owns, so it must not count against — or trigger — the per-tx
// allocator budget. The only side effect is the gas charge for the CPU
// of producing the bytes.
//
// Write methods never return an error: the only sink is parent, which
// in every production path is a bytes.Buffer / strings.Builder / the
// machine output writer. A parent write failure is treated as fatal and
// panics, keeping the recursive formatters free of error plumbing.
type meteredWriter struct {
	parent   io.Writer
	gasMeter store.GasMeter
	buf      [meteredWriterBufSize]byte
	n        int
}

// meteredWriterPool recycles the (kilobyte-sized) meteredWriter structs so a
// fresh Sprint doesn't heap-allocate its buffer each time.
var meteredWriterPool = sync.Pool{New: func() any { return &meteredWriter{} }}

// newUnmeteredWriter borrows a meteredWriter wrapping w that charges NO gas —
// for the debug / query / test rendering paths (String / ProtectedString, or
// Sprint with no machine) where output is deliberately not metered. The caller
// owns the writer and must Release() it once done (after the final Flush).
func newUnmeteredWriter(w io.Writer) *meteredWriter {
	mw := meteredWriterPool.Get().(*meteredWriter)
	mw.parent = w
	mw.n = 0
	mw.gasMeter = nil
	return mw
}

// newMeteredWriter borrows a meteredWriter wrapping w so that flushed output
// charges gas against m's gas meter. m must be non-nil: a nil machine on a
// metered path is a bug (use newUnmeteredWriter when there is legitimately no
// machine), and requiring it here stops a caller from silently skipping gas by
// passing nil. m.GasMeter may itself be nil — query / test machines are built
// without a meter — in which case gas accounting is skipped. The caller owns
// the writer and must Release() it once done (after the final Flush).
func newMeteredWriter(w io.Writer, m *Machine) *meteredWriter {
	if m == nil {
		panic("newMeteredWriter: nil machine; use newUnmeteredWriter for unmetered rendering")
	}
	mw := meteredWriterPool.Get().(*meteredWriter)
	mw.parent = w
	mw.n = 0
	mw.gasMeter = m.GasMeter
	return mw
}

// Release returns mw to the pool. Safe to call after a panic mid-format:
// newMeteredWriter resets n, so a half-filled recycled buffer is harmless.
func (mw *meteredWriter) Release() {
	mw.parent = nil
	mw.gasMeter = nil
	meteredWriterPool.Put(mw)
}

// Flush writes the buffered bytes to parent and charges output gas for
// them (which may panic OutOfGasError, propagating through the recursion
// to the SDK's doRecover). A parent write error is fatal.
//
// n is reset to 0 BEFORE charging gas: if ConsumeGas panics OutOfGas, the
// panic unwinds through callers whose deferred cleanup (e.g. Fprint's
// `defer func(){ mw.Flush(); mw.Release() }`) calls Flush again. With n
// already cleared that second Flush is a no-op, so the tripping chunk is
// charged exactly once rather than twice. The buffered bytes (buf[:n]) are
// dropped on OOG — the same outcome as before, since the parent.Write was
// never reached on the failing flush.
func (mw *meteredWriter) Flush() {
	if mw.n == 0 {
		return
	}
	n := mw.n
	mw.n = 0
	if mw.gasMeter != nil {
		mw.gasMeter.ConsumeGas(streamOutputGas(n), "stream output")
	}
	if _, err := mw.parent.Write(mw.buf[:n]); err != nil {
		panic(fmt.Sprintf("meteredWriter: parent write failed: %v", err))
	}
}

// reserve flushes if fewer than need bytes remain free. need must be
// <= meteredWriterBufSize so the post-flush buffer can hold it. If a
// caller ever reserves more than the buffer holds, the strconv.Append*
// that follows would silently reallocate (writing outside mw.buf and
// desyncing mw.n) — fail loudly instead.
func (mw *meteredWriter) reserve(need int) {
	if need > len(mw.buf) {
		panic("meteredWriter.reserve: need exceeds buffer size")
	}
	if mw.n+need > len(mw.buf) {
		mw.Flush()
	}
}

// WriteByte buffers a single byte. It returns an error only to satisfy
// io.ByteWriter's canonical signature (and go vet's stdmethods check);
// the error is always nil and callers ignore it.
func (mw *meteredWriter) WriteByte(b byte) error {
	mw.reserve(1)
	mw.buf[mw.n] = b
	mw.n++
	return nil
}

func (mw *meteredWriter) WriteString(s string) {
	for len(s) > 0 {
		if mw.n == len(mw.buf) {
			mw.Flush()
		}
		c := copy(mw.buf[mw.n:], s)
		mw.n += c
		s = s[c:]
	}
}

func (mw *meteredWriter) WriteBytes(p []byte) {
	for len(p) > 0 {
		if mw.n == len(mw.buf) {
			mw.Flush()
		}
		c := copy(mw.buf[mw.n:], p)
		mw.n += c
		p = p[c:]
	}
}

// Write satisfies io.Writer so a *meteredWriter can be passed where an
// io.Writer is expected (notably (*TypedValue).Fprint's dispatch).
// The error is always nil — buffering never fails; a parent failure
// panics in Flush.
func (mw *meteredWriter) Write(p []byte) (int, error) {
	mw.WriteBytes(p)
	return len(p), nil
}

// The Write{Int,Uint,Bool,Float} helpers append strconv output straight
// into the buffer tail after reserving worst-case space, so no scratch
// slice escapes to the heap.

func (mw *meteredWriter) WriteInt(i int64) {
	mw.reserve(20) // len("-9223372036854775808")
	mw.n = len(strconv.AppendInt(mw.buf[:mw.n], i, 10))
}

func (mw *meteredWriter) WriteUint(u uint64) {
	mw.reserve(20) // len("18446744073709551615")
	mw.n = len(strconv.AppendUint(mw.buf[:mw.n], u, 10))
}

func (mw *meteredWriter) WriteBool(b bool) {
	mw.reserve(5) // len("false")
	mw.n = len(strconv.AppendBool(mw.buf[:mw.n], b))
}

func (mw *meteredWriter) WriteFloat(f float64, bitSize int) {
	mw.reserve(32) // generous upper bound for 'g' formatting
	mw.n = len(strconv.AppendFloat(mw.buf[:mw.n], f, 'g', -1, bitSize))
}

// WriteQuote writes the Go-quoted form of s (fmt %q-equivalent), matching
// ProtectedString's strconv.Quote post-step.
//
// AppendQuote appends straight into the buffer tail: in the common case the
// quoted form fits and it writes in place (no allocation). It only grows — and
// thus allocates a new backing array — when the result exceeds the buffer
// size; detect that via len(b) and copy the produced bytes through WriteBytes
// (which flushes as needed). Output is strconv.AppendQuote's, so byte-identical.
func (mw *meteredWriter) WriteQuote(s string) {
	b := strconv.AppendQuote(mw.buf[:mw.n], s)
	if len(b) <= len(mw.buf) {
		mw.n = len(b) // appended in place; no allocation
		return
	}
	mw.WriteBytes(b[mw.n:]) // didn't fit: copy the quoted bytes out
}

// appendHexUpper appends the uppercase hex encoding of src to dst,
// matching fmt.Sprintf("%X", src).
func appendHexUpper(dst, src []byte) []byte {
	const hexDigits = "0123456789ABCDEF"
	for _, b := range src {
		dst = append(dst, hexDigits[b>>4], hexDigits[b&0x0F])
	}
	return dst
}

// writeProtectedSprint is the bare-form recursion mirror of
// (*TypedValue).ProtectedSprint — same dispatch, same cycle handling,
// but writes bytes to w instead of returning a string.
//
// Output is byte-identical to ProtectedSprint(seen, considerDeclaredType)
// over the corpus in TestSprintMatchesGolden.
func writeProtectedSprint(w *meteredWriter, tv TypedValue, seen *seenValues, considerDeclaredType bool) {
	if i := seen.IndexOf(tv.V); i != -1 {
		w.WriteString("ref@")
		w.WriteInt(int64(i))
		return
	}

	// Declared type — delegate to the wrapped (type-labeled) form.
	if _, ok := tv.T.(*DeclaredType); ok && considerDeclaredType {
		tv.WriteProtected(w, seen)
		return
	}

	// RefValue early-return — matches the special case in ProtectedSprint
	// added to prevent recursive print loops.
	if tv.V != nil {
		if v, ok := tv.V.(RefValue); ok {
			w.WriteString(v.String())
			return
		}
	}

	switch bt := baseOf(tv.T).(type) {
	case PrimitiveType:
		switch bt {
		case UntypedBoolType, BoolType:
			w.WriteBool(tv.GetBool())
		case UntypedStringType, StringType:
			w.WriteString(tv.GetString())
		case IntType:
			w.WriteInt(tv.GetInt())
		case Int8Type:
			w.WriteInt(int64(tv.GetInt8()))
		case Int16Type:
			w.WriteInt(int64(tv.GetInt16()))
		case UntypedRuneType, Int32Type:
			w.WriteInt(int64(tv.GetInt32()))
		case Int64Type:
			w.WriteInt(tv.GetInt64())
		case UintType:
			w.WriteUint(tv.GetUint())
		case Uint8Type:
			w.WriteUint(uint64(tv.GetUint8()))
		case DataByteType:
			w.WriteUint(uint64(tv.GetDataByte()))
		case Uint16Type:
			w.WriteUint(uint64(tv.GetUint16()))
		case Uint32Type:
			w.WriteUint(uint64(tv.GetUint32()))
		case Uint64Type:
			w.WriteUint(tv.GetUint64())
		case Float32Type:
			w.WriteFloat(float64(math.Float32frombits(tv.GetFloat32())), 32)
		case Float64Type:
			w.WriteFloat(math.Float64frombits(tv.GetFloat64()), 64)
		case UntypedBigintType:
			w.WriteString(tv.V.(BigintValue).V.String())
		case UntypedBigdecType:
			w.WriteString(tv.V.(BigdecValue).String())
		default:
			panic("should not happen")
		}
	case *PointerType:
		if tv.V == nil {
			w.WriteString("typed-nil")
			return
		}
		tv.V.(PointerValue).WriteProtected(w, seen)
	case *FuncType:
		switch fv := tv.V.(type) {
		case nil:
			w.WriteString(nilStr)
			w.WriteString(" ")
			w.WriteString(tv.T.String())
		case *FuncValue, *BoundMethodValue:
			w.WriteString(fv.(fmt.Stringer).String())
		default:
			panic(fmt.Sprintf("unexpected func type %v", reflect.TypeOf(tv.V)))
		}
	case *InterfaceType:
		if debug {
			if tv.DebugHasValue() {
				panic("should not happen")
			}
		}
		w.WriteString(nilStr)
	case *DeclaredType:
		panic("should not happen")
	case *PackageType:
		w.WriteString(tv.V.(*PackageValue).String())
	case *ChanType:
		panic("not yet implemented")
	case *TypeType:
		w.WriteString(tv.V.(TypeValue).String())
	default:
		// The remaining types may have a nil value.
		if tv.V == nil {
			w.WriteString("(")
			w.WriteString(nilStr)
			w.WriteString(" ")
			w.WriteString(tv.T.String())
			w.WriteString(")")
			return
		}
		// *ArrayType, *SliceType, *StructType, *MapType
		switch v := tv.V.(type) {
		case protectedWriter:
			v.WriteProtected(w, seen)
		case fmt.Stringer:
			// *NativeType etc.
			w.WriteString(v.String())
		default:
			if debug {
				panic(fmt.Sprintf("unexpected type %s", tv.T.String()))
			}
			panic("should not happen")
		}
	}
}

// Fprint writes the formatted form of tv to w. It is the streaming
// counterpart of (*TypedValue).Sprint(m). When w is not already a
// *meteredWriter, it is wrapped (and flushed before returning); when it
// already is one — the uversePrint case — it is used directly and the
// caller owns the flush, avoiding double-wrapping.
func (tv *TypedValue) Fprint(w io.Writer, m *Machine) {
	mw, ok := w.(*meteredWriter)
	if !ok {
		if m == nil {
			mw = newUnmeteredWriter(w) // debug/test Sprint with no machine
		} else {
			mw = newMeteredWriter(w, m)
		}
		// Deferred Flush: covers all of Fprint's early-return paths uniformly,
		// and the caller reads w only after we return (so flush-on-exit is in
		// time). Contrast protectedStringOf, which reads its buffer in-function
		// and must flush explicitly.
		defer func() {
			mw.Flush()
			mw.Release()
		}()
	}

	// undefined → "undefined", matching Sprint(m)'s short-circuit.
	if tv == nil || tv.T == nil {
		mw.WriteString(undefinedStr)
		return
	}

	// Stringer / Error dispatch — invokes the gno-side method via m.Eval
	// and writes the resulting string. The intermediate Go string is
	// allocated outside the writer's accounting (bounded indirectly by
	// the gno call's own gas/alloc budget); see ADR for the gap note.
	if IsImplementedBy(gStringerType, tv.T) && !tv.IsNilInterface() {
		res := m.Eval(Call(Sel(&ConstExpr{TypedValue: *tv}, "String")))
		mw.WriteString(res[0].GetString())
		return
	}
	if IsImplementedBy(gErrorType, tv.T) {
		res := m.Eval(Call(Sel(&ConstExpr{TypedValue: *tv}, "Error")))
		mw.WriteString(res[0].GetString())
		return
	}

	writeProtectedSprint(mw, *tv, newSeenValues(), true)
}

// writeRefOrPut emits the "ref@N" cycle-output or installs v in the
// seen-stack via Put, returning true when it handled the value (the
// caller has nothing more to do).
func writeRefOrPut(w *meteredWriter, seen *seenValues, v Value) (handled bool) {
	if i := seen.IndexOf(v); i != -1 {
		w.WriteString("ref@")
		w.WriteInt(int64(i))
		return true
	}
	if !seen.Put(v) {
		w.WriteString("...")
		return true
	}
	return false
}

// writeValueList writes the recursive aggregate form "label[e0,e1,...,eN]"
// shared by ArrayValue and SliceValue.
func writeValueList(w *meteredWriter, seen *seenValues, label string, elems []TypedValue) {
	w.WriteString(label)
	w.WriteByte('[')
	for i := range elems {
		if i > 0 {
			w.WriteByte(',')
		}
		elems[i].WriteProtected(w, seen)
	}
	w.WriteByte(']')
}

func (av *ArrayValue) WriteProtected(w *meteredWriter, seen *seenValues) {
	if writeRefOrPut(w, seen, av) {
		return
	}
	defer seen.Pop()

	if av.Data == nil {
		writeValueList(w, seen, "array", av.List)
		return
	}

	// Byte-data path: "array[0x..]" or "array[0x..first256...]" for >256 bytes.
	w.WriteString("array[0x")
	if len(av.Data) > 256 {
		w.WriteBytes(appendHexUpper(nil, av.Data[:256]))
		w.WriteString("...]")
		return
	}
	w.WriteBytes(appendHexUpper(nil, av.Data))
	w.WriteByte(']')
}

func (sv *SliceValue) WriteProtected(w *meteredWriter, seen *seenValues) {
	if sv.Base == nil {
		w.WriteString("nil-slice")
		return
	}

	if i := seen.IndexOf(sv); i != -1 {
		w.WriteString("ref@")
		w.WriteInt(int64(i))
		return
	}

	if ref, ok := sv.Base.(RefValue); ok {
		// "slice[%v]" where %v uses RefValue.String()
		w.WriteString("slice[")
		w.WriteString(ref.String())
		w.WriteByte(']')
		return
	}

	if !seen.Put(sv) {
		w.WriteString("...")
		return
	}
	defer seen.Pop()

	vbase := sv.Base.(*ArrayValue)
	if vbase.Data == nil {
		writeValueList(w, seen, "slice", vbase.List[sv.Offset:sv.Offset+sv.Length])
		return
	}

	// Byte-data path mirrors fmt.Sprintf("slice[0x%X...(%d)]" / "slice[0x%X]").
	w.WriteString("slice[0x")
	if sv.Length > 256 {
		w.WriteBytes(appendHexUpper(nil, vbase.Data[sv.Offset:sv.Offset+256]))
		w.WriteString("...(")
		w.WriteInt(int64(sv.Length))
		w.WriteString(")]")
		return
	}
	w.WriteBytes(appendHexUpper(nil, vbase.Data[sv.Offset:sv.Offset+sv.Length]))
	w.WriteByte(']')
}

func (sv *StructValue) WriteProtected(w *meteredWriter, seen *seenValues) {
	if writeRefOrPut(w, seen, sv) {
		return
	}
	defer seen.Pop()

	w.WriteString("struct{")
	for i, f := range sv.Fields {
		if i > 0 {
			w.WriteByte(',')
		}
		f.WriteProtected(w, seen)
	}
	w.WriteByte('}')
}

func (mv *MapValue) WriteProtected(w *meteredWriter, seen *seenValues) {
	if mv.List == nil {
		w.WriteString("zero-map")
		return
	}

	if writeRefOrPut(w, seen, mv) {
		return
	}
	defer seen.Pop()

	w.WriteString("map{")
	for next, first := mv.List.Head, true; next != nil; next, first = next.Next, false {
		if !first {
			w.WriteByte(',')
		}
		next.Key.WriteProtected(w, seen)
		w.WriteByte(':')
		next.Value.WriteProtected(w, seen)
	}
	w.WriteByte('}')
}

func (pv PointerValue) WriteProtected(w *meteredWriter, seen *seenValues) {
	if writeRefOrPut(w, seen, pv) {
		return
	}
	defer seen.Pop()

	// Match ProtectedString's nil-TV branch ("&<nil>").
	if pv.TV == nil {
		w.WriteString("&<nil>")
		return
	}

	w.WriteByte('&')
	pv.TV.WriteProtected(w, seen)
}

// WriteProtected is the wrapped-form streaming counterpart of
// (TypedValue).ProtectedString — output is "(value type)". For
// primitives the value is rendered the same way ProtectedSprint
// would render it, except strings get quoted in the wrap form.
func (tv TypedValue) WriteProtected(w *meteredWriter, seen *seenValues) {
	if tv.IsUndefined() {
		w.WriteString("(undefined)")
		return
	}

	w.WriteByte('(')

	if tv.V == nil {
		// Mirror ProtectedString's V-nil primitive switch.
		switch baseOf(tv.T) {
		case BoolType, UntypedBoolType:
			w.WriteBool(tv.GetBool())
		case StringType, UntypedStringType:
			w.WriteString(tv.GetString())
		case IntType:
			w.WriteInt(tv.GetInt())
		case Int8Type:
			w.WriteInt(int64(tv.GetInt8()))
		case Int16Type:
			w.WriteInt(int64(tv.GetInt16()))
		case Int32Type, UntypedRuneType:
			w.WriteInt(int64(tv.GetInt32()))
		case Int64Type:
			w.WriteInt(tv.GetInt64())
		case UintType:
			w.WriteUint(tv.GetUint())
		case Uint8Type:
			w.WriteUint(uint64(tv.GetUint8()))
		case DataByteType:
			w.WriteUint(uint64(tv.GetDataByte()))
		case Uint16Type:
			w.WriteUint(uint64(tv.GetUint16()))
		case Uint32Type:
			w.WriteUint(uint64(tv.GetUint32()))
		case Uint64Type:
			w.WriteUint(tv.GetUint64())
		case Float32Type:
			w.WriteFloat(float64(math.Float32frombits(tv.GetFloat32())), 32)
		case Float64Type:
			w.WriteFloat(math.Float64frombits(tv.GetFloat64()), 64)
		default:
			// Complex types that require recursion protection.
			w.WriteString(nilStr)
		}
	} else if base := baseOf(tv.T); base == StringType || base == UntypedStringType {
		// V != nil string — quote it, matching ProtectedString's
		// strconv.Quote(vs) post-step.
		w.WriteQuote(tv.GetString())
	} else {
		// V != nil — recurse without re-considering the declared type.
		writeProtectedSprint(w, tv, seen, false)
	}

	w.WriteByte(' ')
	w.WriteString(tv.T.String())
	w.WriteByte(')')
}
