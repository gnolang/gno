package gnolang

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/overflow"
	"github.com/gnolang/gno/tm2/pkg/store"
)

type Printer struct {
	GasMeter store.GasMeter
}

func NewPrinter(gasMeter store.GasMeter) *Printer {
	return &Printer{
		GasMeter: gasMeter,
	}
}

func (p *Printer) incrCPU(size int64) {
	if p == nil {
		return
	}

	if p.GasMeter != nil {
		gasCPU := overflow.Mulp(size, GasFactorCPU)
		p.GasMeter.ConsumeGas(gasCPU, "")
	}
}

func (p *Printer) Sprintf(format string, args ...interface{}) string {
	ss := fmt.Sprintf(format, args...)

	p.incrCPU(int64(OpCharPrint * len(ss)))

	return ss
}

func (p *Printer) Sprint(args ...interface{}) string {
	ss := fmt.Sprint(args...)

	p.incrCPU(int64(OpCharPrint * len(ss)))

	return ss
}

type protectedStringer interface {
	ProtectedString(*Printer, *seenValues) string
}

const (
	// defaultSeenValuesSize indicates the maximum anticipated depth of the stack when printing a Value type.
	defaultSeenValuesSize = 32

	// nestedLimit indicates the maximum nested level when printing a deeply recursive value.
	// if this increases significantly a map should be used instead
	nestedLimit = 10
)

type seenValues struct {
	values []Value
	nc     int // nested counter, to limit recursivity
}

func (sv *seenValues) Put(v Value) {
	sv.values = append(sv.values, v)
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
		nc:     nestedLimit,
	}
}

func (sv StringValue) String(printer *Printer) string {
	return printer.Sprint(strconv.Quote(string(sv)))
}

func (biv BigintValue) String(printer *Printer) string {
	return printer.Sprint(biv.V.String())
}

func (bdv BigdecValue) String(printer *Printer) string {
	return printer.Sprint(bdv.V.String())
}

func (dbv DataByteValue) String(printer *Printer) string {
	return printer.Sprintf("(%0X)", (dbv.GetByte()))
}

func (av *ArrayValue) String(printer *Printer) string {
	return av.ProtectedString(printer, newSeenValues())
}

func (av *ArrayValue) ProtectedString(printer *Printer, seen *seenValues) string {
	if i := seen.IndexOf(av); i != -1 {
		return fmt.Sprintf("ref@%d", i)
	}

	seen.nc--
	if seen.nc < 0 {
		return "..."
	}
	seen.Put(av)
	defer seen.Pop()

	ss := make([]string, len(av.List))
	if av.Data == nil {
		for i, e := range av.List {
			ss[i] = e.ProtectedString(printer, seen)
		}
		// NOTE: we may want to unify the representation,
		// but for now tests expect this to be different.
		// This may be helpful for testing implementation behavior.
		return printer.Sprint("array[" + strings.Join(ss, ",") + "]")
	}
	if len(av.Data) > 256 {
		return printer.Sprintf("array[0x%X...]", av.Data[:256])
	}
	return printer.Sprintf("array[0x%X]", av.Data)
}

func (sv *SliceValue) String(printer *Printer) string {
	return sv.ProtectedString(printer, newSeenValues())
}

func (sv *SliceValue) ProtectedString(printer *Printer, seen *seenValues) string {
	if sv.Base == nil {
		return printer.Sprint("nil-slice")
	}

	if i := seen.IndexOf(sv); i != -1 {
		return printer.Sprintf("ref@%d", i)
	}

	if ref, ok := sv.Base.(RefValue); ok {
		return printer.Sprintf("slice[%v]", ref)
	}

	seen.Put(sv)
	defer seen.Pop()

	vbase := sv.Base.(*ArrayValue)
	if vbase.Data == nil {
		ss := make([]string, sv.Length)
		for i, e := range vbase.List[sv.Offset : sv.Offset+sv.Length] {
			ss[i] = e.ProtectedString(printer, seen)
		}
		return "slice[" + strings.Join(ss, ",") + "]"
	}
	if sv.Length > 256 {
		return printer.Sprintf("slice[0x%X...(%d)]", vbase.Data[sv.Offset:sv.Offset+256], sv.Length)
	}
	return fmt.Sprintf("slice[0x%X]", vbase.Data[sv.Offset:sv.Offset+sv.Length])
}

func (pv PointerValue) String(printer *Printer) string {
	return pv.ProtectedString(printer, newSeenValues())
}

func (pv PointerValue) ProtectedString(printer *Printer, seen *seenValues) string {
	if i := seen.IndexOf(pv); i != -1 {
		return printer.Sprintf("ref@%d", i)
	}

	seen.Put(pv)
	defer seen.Pop()

	// Handle nil TV's, avoiding a nil pointer deref below.
	if pv.TV == nil {
		return printer.Sprint("&<nil>")
	}

	return printer.Sprintf("&%s", pv.TV.ProtectedString(printer, seen))
}

func (sv *StructValue) String(printer *Printer) string {
	return sv.ProtectedString(printer, newSeenValues())
}

func (sv *StructValue) ProtectedString(printer *Printer, seen *seenValues) string {
	if i := seen.IndexOf(sv); i != -1 {
		return fmt.Sprintf("ref@%d", i)
	}

	seen.Put(sv)
	defer seen.Pop()

	ss := make([]string, len(sv.Fields))
	for i, f := range sv.Fields {
		ss[i] = f.ProtectedString(printer, seen)
	}
	return "struct{" + strings.Join(ss, ",") + "}"
}

func (fv *FuncValue) String(printer *Printer) string {
	name := string(fv.Name)
	if fv.Type == nil {
		return printer.Sprintf("incomplete-func ?%s(?)?", name)
	}
	if name == "" {
		return printer.Sprintf("%s{...}", fv.Type.String(printer))
	}
	return name
}

func (bmv *BoundMethodValue) String(printer *Printer) string {
	name := bmv.Func.Name
	var (
		recvT   string = "?"
		params  string = "?"
		results string = "(?)"
	)
	if ft, ok := bmv.Func.Type.(*FuncType); ok {
		recvT = ft.Params[0].Type.String(printer)
		params = FieldTypeList(ft.Params).StringForFunc(printer)
		if len(results) > 0 {
			results = FieldTypeList(ft.Results).StringForFunc(printer)
			results = "(" + results + ")"
		}
	}
	return printer.Sprintf("<%s>.%s(%s)%s",
		recvT, name, params, results)
}

func (mv *MapValue) String(printer *Printer) string {
	return mv.ProtectedString(printer, newSeenValues())
}

func (mv *MapValue) ProtectedString(printer *Printer, seen *seenValues) string {
	if mv.List == nil {
		return printer.Sprint("zero-map")
	}

	if i := seen.IndexOf(mv); i != -1 {
		return printer.Sprintf("ref@%d", i)
	}

	seen.Put(mv)
	defer seen.Pop()

	ss := make([]string, 0, mv.GetLength())
	next := mv.List.Head
	for next != nil {
		ss = append(ss,
			next.Key.ProtectedString(printer, seen)+":"+
				next.Value.ProtectedString(printer, seen))
		next = next.Next
	}
	return printer.Sprint("map{" + strings.Join(ss, ",") + "}")
}

func (tv TypeValue) String(printer *Printer) string {
	return fmt.Sprintf("typeval{%s}",
		tv.Type.String(printer))
}

func (pv *PackageValue) String(printer *Printer) string {
	return fmt.Sprintf("package(%s %s)", pv.PkgName, pv.PkgPath)
}

func (b *Block) String(printer *Printer) string {
	return b.StringIndented(printer, "    ")
}

func (b *Block) StringIndented(printer *Printer, indent string) string {
	source := toString(b.Source)
	if len(source) > 32 {
		source = source[:32] + "..."
	}
	lines := make([]string, 0, 3)
	lines = append(lines,
		printer.Sprintf("Block(ID:%v,Addr:%p,Source:%s,Parent:%p)",
			b.ObjectInfo.ID, b, source, b.Parent)) // XXX Parent may be RefValue{}.
	if b.Source != nil {
		if _, ok := b.Source.(RefNode); ok {
			lines = append(lines,
				printer.Sprintf("%s(RefNode names not shown)", indent))
		} else {
			types := b.Source.GetStaticBlock().Types
			for i, n := range b.Source.GetBlockNames() {
				if len(b.Values) <= i {
					lines = append(lines,
						printer.Sprintf("%s%s: undefined static:%s", indent, n, types[i]))
				} else {
					lines = append(lines,
						printer.Sprintf("%s%s: %s static:%s",
							indent, n, b.Values[i].String(printer), types[i]))
				}
			}
		}
	}
	return strings.Join(lines, "\n")
}

func (rv RefValue) String(printer *Printer) string {
	if rv.PkgPath == "" {
		return printer.Sprintf("ref(%v)",
			rv.ObjectID)
	}
	return printer.Sprintf("ref(%s)",
		rv.PkgPath)
}

func (hiv *HeapItemValue) String(printer *Printer) string {
	return printer.Sprintf("heapitem(%v)",
		hiv.Value)
}

// ----------------------------------------
// *TypedValue.Sprint

// for print() and println().
func (tv *TypedValue) Sprint(printer *Printer, m *Machine) string {
	// if undefined, just "undefined".
	if tv == nil || tv.T == nil {
		return undefinedStr
	}

	// if implements .String(), return it.
	if IsImplementedBy(gStringerType, tv.T) && !tv.IsNilInterface() {
		res := m.Eval(Call(Sel(&ConstExpr{TypedValue: *tv}, "String")))
		return res[0].GetString()
	}
	// if implements .Error(), return it.
	if IsImplementedBy(gErrorType, tv.T) {
		res := m.Eval(Call(Sel(&ConstExpr{TypedValue: *tv}, "Error")))
		return res[0].GetString()
	}

	return tv.ProtectedSprint(printer, newSeenValues(), true)
}

func (tv *TypedValue) ProtectedSprint(printer *Printer, seen *seenValues, considerDeclaredType bool) string {
	if i := seen.IndexOf(tv.V); i != -1 {
		return printer.Sprintf("ref@%d", i)
	}

	// print declared type
	if _, ok := tv.T.(*DeclaredType); ok && considerDeclaredType {
		return tv.ProtectedString(printer, seen)
	}

	// This is a special case that became necessary after adding `ProtectedString()` methods to
	// reliably prevent recursive print loops.
	if tv.V != nil {
		if v, ok := tv.V.(RefValue); ok {
			return v.String(printer)
		}
	}

	// otherwise, default behavior.
	switch bt := baseOf(tv.T).(type) {
	case PrimitiveType:
		switch bt {
		case UntypedBoolType, BoolType:
			return fmt.Sprintf("%t", tv.GetBool())
		case UntypedStringType, StringType:
			return tv.GetString()
		case IntType:
			return fmt.Sprintf("%d", tv.GetInt())
		case Int8Type:
			return fmt.Sprintf("%d", tv.GetInt8())
		case Int16Type:
			return fmt.Sprintf("%d", tv.GetInt16())
		case UntypedRuneType, Int32Type:
			return fmt.Sprintf("%d", tv.GetInt32())
		case Int64Type:
			return fmt.Sprintf("%d", tv.GetInt64())
		case UintType:
			return fmt.Sprintf("%d", tv.GetUint())
		case Uint8Type:
			return fmt.Sprintf("%d", tv.GetUint8())
		case DataByteType:
			return fmt.Sprintf("%d", tv.GetDataByte())
		case Uint16Type:
			return fmt.Sprintf("%d", tv.GetUint16())
		case Uint32Type:
			return fmt.Sprintf("%d", tv.GetUint32())
		case Uint64Type:
			return fmt.Sprintf("%d", tv.GetUint64())
		case Float32Type:
			return fmt.Sprintf("%v", math.Float32frombits(tv.GetFloat32()))
		case Float64Type:
			return fmt.Sprintf("%v", math.Float64frombits(tv.GetFloat64()))
		case UntypedBigintType:
			return tv.V.(BigintValue).V.String()
		case UntypedBigdecType:
			return tv.V.(BigdecValue).V.String()
		default:
			panic("should not happen")
		}
	case *PointerType:
		if tv.V == nil {
			return "typed-nil"
		}
		roPre, roPost := "", ""
		if tv.IsReadonly() {
			roPre, roPost = "readonly(", ")"
		}
		return roPre + tv.V.(PointerValue).ProtectedString(printer, seen) + roPost
	case *FuncType:
		switch fv := tv.V.(type) {
		case nil:
			ft := tv.T.String(printer)
			return nilStr + " " + ft
		case *FuncValue, *BoundMethodValue:
			return fv.String(printer)
		default:
			panic(fmt.Sprintf(
				"unexpected func type %v",
				reflect.TypeOf(tv.V)))
		}
	case *InterfaceType:
		if debug {
			if tv.DebugHasValue() {
				panic("should not happen")
			}
		}
		return nilStr
	case *DeclaredType:
		panic("should not happen")
	case *PackageType:
		return tv.V.(*PackageValue).String(printer)
	case *ChanType:
		panic("not yet implemented")
	case *TypeType:
		return tv.V.(TypeValue).String(printer)
	default:
		// The remaining types may have a nil value.
		if tv.V == nil {
			return "(" + nilStr + " " + tv.T.String(printer) + ")"
		}
		// Value may be N_Readonly
		roPre, roPost := "", ""
		if tv.IsReadonly() {
			roPre, roPost = "readonly(", ")"
		}
		// *ArrayType, *SliceType, *StructType, *MapType
		if ps, ok := tv.V.(protectedStringer); ok {
			return roPre + ps.ProtectedString(printer, seen) + roPost
		}

		// *NativeType
		return roPre + tv.V.String(printer) + roPost

	}
}

// ----------------------------------------
// TypedValue.String()

// For gno debugging/testing.
func (tv TypedValue) String(printer *Printer) string {
	return tv.ProtectedString(printer, newSeenValues())
}

func (tv TypedValue) ProtectedString(printer *Printer, seen *seenValues) string {
	if tv.IsUndefined() {
		return "(undefined)"
	}
	vs := ""
	if tv.V == nil {
		switch baseOf(tv.T) {
		case BoolType, UntypedBoolType:
			vs = fmt.Sprintf("%t", tv.GetBool())
		case StringType, UntypedStringType:
			vs = fmt.Sprintf("%s", tv.GetString())
		case IntType:
			vs = fmt.Sprintf("%d", tv.GetInt())
		case Int8Type:
			vs = fmt.Sprintf("%d", tv.GetInt8())
		case Int16Type:
			vs = fmt.Sprintf("%d", tv.GetInt16())
		case Int32Type, UntypedRuneType:
			vs = fmt.Sprintf("%d", tv.GetInt32())
		case Int64Type:
			vs = fmt.Sprintf("%d", tv.GetInt64())
		case UintType:
			vs = fmt.Sprintf("%d", tv.GetUint())
		case Uint8Type:
			vs = fmt.Sprintf("%d", tv.GetUint8())
		case DataByteType:
			vs = fmt.Sprintf("%d", tv.GetDataByte())
		case Uint16Type:
			vs = fmt.Sprintf("%d", tv.GetUint16())
		case Uint32Type:
			vs = fmt.Sprintf("%d", tv.GetUint32())
		case Uint64Type:
			vs = fmt.Sprintf("%d", tv.GetUint64())
		case Float32Type:
			vs = fmt.Sprintf("%v", math.Float32frombits(tv.GetFloat32()))
		case Float64Type:
			vs = fmt.Sprintf("%v", math.Float64frombits(tv.GetFloat64()))
		// Complex types that require recusion protection.
		default:
			vs = nilStr
		}
	} else {
		vs = tv.ProtectedSprint(printer, seen, false)
		if base := baseOf(tv.T); base == StringType || base == UntypedStringType {
			vs = strconv.Quote(vs)
		}
	}

	ts := tv.T.String(printer)
	return fmt.Sprintf("(%s %s)", vs, ts) // TODO improve
}
