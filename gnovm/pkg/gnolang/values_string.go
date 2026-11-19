package gnolang

import (
	"bytes"
	"fmt"
	"io"
	"math/big"
	"strconv"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/store/types"
)

const (
	// defaultSeenValuesSize indicates the maximum anticipated depth of the stack when printing a Value type.
	defaultSeenValuesSize = 32

	// nestedLimit indicates the maximum nested level when printing a deeply recursive value.
	// if this increases significantly a map should be used instead
	nestedLimit = 10
)

type seenValues struct {
	values []Value
}

func (sv *seenValues) Put(v Value) bool {
	if len(sv.values) >= nestedLimit {
		return false
	}

	sv.values = append(sv.values, v)
	return true
}

func (sv *seenValues) IndexOf(v Value) int {
	for i, vv := range sv.values {
		if vv == v {
			return i
		}
	}

	return -1
}

// Pop should be called by using a defer after each Put.
// Consider why this is necessary:
//   - we are printing an array of structs
//   - each invocation of struct.ProtectedString adds the value to the seenValues
//   - without calling Pop before exiting struct.ProtectedString, the next call to
//     struct.ProtectedString in the array.ProtectedString loop will not result in the value
//     being printed if the value has already been print
//   - this is NOT recursion and SHOULD be printed
func (sv *seenValues) Pop() {
	sv.values = sv.values[:len(sv.values)-1]
}

func newSeenValues() *seenValues {
	return &seenValues{
		values: make([]Value, 0, defaultSeenValuesSize),
	}
}

func (sv StringValue) String() string {
	return strconv.Quote(string(sv))
}

func (biv BigintValue) String() string {
	return biv.V.String()
}

func (bdv BigdecValue) String() string {
	if bdv.F != nil {
		return bigdecFloatString(bdv.F)
	}
	if bdv.V == nil {
		return "0.0"
	}
	s := bdv.V.FloatString(10)
	// Trim trailing zeros after the decimal point, but keep at least one
	// decimal digit so bigdec values are visually distinct from integers.
	if strings.ContainsRune(s, '.') {
		s = strings.TrimRight(s, "0")
		// Keep at least one digit after the decimal point.
		if s[len(s)-1] == '.' {
			s += "0"
		}
	}
	return s
}

// bigdecFloatString renders f. Within ratOverflowBits it prints exactly. Past
// it, exact digits cost time growing faster than |exp|, so it prints six
// approximate significant digits instead, in a few dozen 64-bit operations
// whatever the exponent:
//
//  1. Write f = mant * 2^exp and guess the decimal exponent e = exp*log10(2).
//  2. Divide by 10^e as 2^e times 5^e. The 2^e is an exponent shift, and 5^e
//     comes from repeated squaring. 10^e itself can overflow big.Float's
//     exponent at the extremes; 5^e cannot.
//  3. Multiply or divide by 10 until the mantissa lands in [1, 10), since the
//     guess can be off by one.
//  4. Print six digits, carrying a round-up to "10" into the exponent.
func bigdecFloatString(f *big.Float) string {
	var mant big.Float
	exp := f.MantExp(&mant) // f = mant * 2^exp, 0.5 <= |mant| < 1
	if exp >= -ratOverflowBits && exp <= ratOverflowBits {
		return f.Text('g', -1)
	}
	e := (int64(exp) * 1292913986) >> 32 // exp * log10(2), in 32-bit fixed point
	m := new(big.Float).SetPrec(64).Set(mant.Abs(&mant))
	m.SetMantExp(m, exp-int(e))
	p5 := new(big.Float).SetPrec(64).SetInt64(1)
	b := new(big.Float).SetPrec(64).SetInt64(5)
	for n := max(e, -e); ; b.Mul(b, b) {
		if n&1 == 1 {
			p5.Mul(p5, b)
		}
		if n >>= 1; n == 0 {
			break
		}
	}
	if e > 0 {
		m.Quo(m, p5)
	} else {
		m.Mul(m, p5)
	}
	one, ten := big.NewFloat(1), big.NewFloat(10)
	for ; m.Cmp(ten) >= 0; e++ {
		m.Quo(m, ten)
	}
	for ; m.Cmp(one) < 0; e-- {
		m.Mul(m, ten)
	}
	digits := m.Text('g', 6)
	if digits == "10" { // m rounded up to the next power of ten
		digits, e = "1", e+1
	}
	sign := ""
	if f.Sign() < 0 {
		sign = "-"
	}
	return fmt.Sprintf("%s%se%+d", sign, digits, e)
}

func (dbv DataByteValue) String() string {
	return fmt.Sprintf("(%0X)", (dbv.GetByte()))
}

// protectedStringOf renders v through the writer-based formatter and
// returns the accumulated bytes. The meteredWriter buffers internally and
// flushes a single chunk into the bytes.Buffer, so no per-shape size
// hint is needed (the debug paths pass no machine, so no gas is charged).
func protectedStringOf(v protectedWriter, seen *seenValues) string {
	var b bytes.Buffer
	_ = writeProtectedTo(&b, v, seen) // a bytes.Buffer never fails a write
	return b.String()
}

// writeProtectedTo streams v's protected form into w. See streamTo.
func writeProtectedTo(w io.Writer, v protectedWriter, seen *seenValues) error {
	return streamTo(w, func(mw *meteredWriter) { v.WriteProtected(mw, seen) })
}

// streamTo runs render on an unmetered writer over w. A write error from w
// aborts the render (Flush panics *ParentWriteError) and is returned.
func streamTo(w io.Writer, render func(*meteredWriter)) (err error) {
	mw := newUnmeteredWriter(w)
	defer mw.Release()
	defer recoverParentWriteError(&err)
	render(mw)
	// Flush explicitly, not deferred, so w holds the whole render on return.
	// Streaming callers that hand the buffer to a downstream consumer
	// (uversePrint, Fprint) defer Flush instead.
	mw.Flush()
	return nil
}

func (av *ArrayValue) String() string {
	return av.ProtectedString(newSeenValues())
}

func (av *ArrayValue) ProtectedString(seen *seenValues) string {
	return protectedStringOf(av, seen)
}

func (sv *SliceValue) String() string {
	return sv.ProtectedString(newSeenValues())
}

func (sv *SliceValue) ProtectedString(seen *seenValues) string {
	return protectedStringOf(sv, seen)
}

func (pv PointerValue) String() string {
	return pv.ProtectedString(newSeenValues())
}

func (pv PointerValue) ProtectedString(seen *seenValues) string {
	return protectedStringOf(pv, seen)
}

func (sv *StructValue) String() string {
	return sv.ProtectedString(newSeenValues())
}

func (sv *StructValue) ProtectedString(seen *seenValues) string {
	return protectedStringOf(sv, seen)
}

func (fv *FuncValue) String() string {
	name := string(fv.Name)
	if fv.Type == nil {
		return fmt.Sprintf("incomplete-func ?%s(?)?", name)
	}
	if name == "" {
		return fmt.Sprintf("%s{...}", fv.Type.String())
	}
	return name
}

func (bmv *BoundMethodValue) String() string {
	if bmv.IsLazy() {
		// Unresolved interface bind: the concrete method is determined at call
		// time; render from the saved operand type + selector name.
		return fmt.Sprintf("<%s>.%s(?)(?)", bmv.Receiver.T.String(), bmv.Method)
	}
	name := bmv.Func.Name
	var (
		recvT   = "?"
		params  = "?"
		results = "(?)"
	)
	if ft, ok := bmv.Func.Type.(*FuncType); ok {
		recvT = ft.Params[0].Type.String()
		params = FieldTypeList(ft.Params).StringForFunc()
		if len(results) > 0 {
			results = FieldTypeList(ft.Results).StringForFunc()
			results = "(" + results + ")"
		}
	}
	return fmt.Sprintf("<%s>.%s(%s)%s",
		recvT, name, params, results)
}

func (mv *MapValue) String() string {
	return mv.ProtectedString(newSeenValues())
}

func (mv *MapValue) ProtectedString(seen *seenValues) string {
	return protectedStringOf(mv, seen)
}

func (tv TypeValue) String() string {
	return "typeval{" + BoundedTypeText(tv.Type, maxExprTypeTextBytes) + "}"
}

func (pv *PackageValue) String() string {
	return fmt.Sprintf("package(%s %s)", pv.PkgName, pv.PkgPath)
}

func (b *Block) String() string {
	return b.StringIndented("    ")
}

func (b *Block) StringIndented(indent string) string {
	source := toString(b.Source)
	if len(source) > 32 {
		source = source[:32] + "..."
	}
	lines := make([]string, 0, 3)
	lines = append(lines,
		fmt.Sprintf("Block(ID:%v,Addr:%p,Source:%s,Parent:%p)",
			b.ObjectInfo.ID, b, source, b.Parent)) // XXX Parent may be RefValue{}.
	if b.Source != nil {
		if _, ok := b.Source.(RefNode); ok {
			lines = append(lines,
				fmt.Sprintf("%s(RefNode names not shown)", indent))
		} else {
			types := b.Source.GetStaticBlock().Types
			for i, n := range b.Source.GetBlockNames() {
				if len(b.Values) <= i {
					lines = append(lines,
						fmt.Sprintf("%s%s: undefined static:%s", indent, n, types[i]))
				} else {
					lines = append(lines,
						fmt.Sprintf("%s%s: %s static:%s",
							indent, n, b.Values[i].String(), types[i]))
				}
			}
		}
	}
	return strings.Join(lines, "\n")
}

func (rv RefValue) String() string {
	if rv.PkgPath == "" {
		return fmt.Sprintf("ref(%v)",
			rv.ObjectID)
	}
	return fmt.Sprintf("ref(%s)",
		rv.PkgPath)
}

func (hiv *HeapItemValue) String() string {
	return fmt.Sprintf("heapitem(%v)",
		hiv.Value)
}

// ----------------------------------------
// *TypedValue.Sprint

// ImplError returns true if the TypedValue's type implements the error
// interface. gm meters the embedded-field BFS walk (nil = a no-op); see
// IsErrorType.
func (tv *TypedValue) ImplError(gm types.GasMeter) bool {
	return isImplementedBy(gm, gErrorType, tv.T)
}

// for print() and println().
func (tv *TypedValue) Sprint(m *Machine) string {
	var b bytes.Buffer
	tv.Fprint(&b, m)
	return b.String()
}

func (tv *TypedValue) ProtectedSprint(seen *seenValues, considerDeclaredType bool) string {
	var b bytes.Buffer
	mw := newUnmeteredWriter(&b)
	defer mw.Release()
	writeProtectedSprint(mw, *tv, seen, considerDeclaredType)
	mw.Flush() // explicit, not deferred — b is read below; see protectedStringOf.
	return b.String()
}

// ----------------------------------------
// TypedValue.String()

// For gno debugging/testing.
func (tv TypedValue) String() string {
	return tv.ProtectedString(newSeenValues())
}

func (tv TypedValue) ProtectedString(seen *seenValues) string {
	return protectedStringOf(tv, seen)
}

// WriteStringTo writes tv.String() into w without building the whole string. A
// write error from w stops the render and is returned, so a w with a byte cap
// bounds both the output and the heap of the render.
func (tv TypedValue) WriteStringTo(w io.Writer) error {
	return writeProtectedTo(w, tv, newSeenValues())
}

// WriteTypeTo writes t.String() without building the whole string.
func WriteTypeTo(w io.Writer, t Type) error {
	return streamTo(w, func(mw *meteredWriter) { writeType(mw, t) })
}

// typeString renders t with writeType, cut at maxTypeTextBytes. The composite
// Type.String() methods return it, so type text has one printer whether it is
// built or streamed.
func typeString(t Type) string {
	return BoundedTypeText(t, maxTypeTextBytes)
}
