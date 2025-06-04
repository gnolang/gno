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

type StringBuilderWithGasMeter struct {
	GasMeter store.GasMeter
	strings.Builder
}

func NewStringBuilderWithGasMeter(gasMeter store.GasMeter) *StringBuilderWithGasMeter {
	return &StringBuilderWithGasMeter{
		GasMeter: gasMeter,
	}
}

func (p *StringBuilderWithGasMeter) incrCPU(size int64) {
	if p == nil {
		return
	}

	if p.GasMeter != nil {
		gasCPU := overflow.Mulp(size, GasFactorCPU)
		p.GasMeter.ConsumeGas(gasCPU, "")
	}
}

func (p *StringBuilderWithGasMeter) WriteString(s string) (int, error) {
	p.incrCPU(int64(OpCharPrint * len(s)))
	return p.Builder.WriteString(s)
}

type protectedStringer interface {
	ProtectedString(Builder, *seenValues) string
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

func (sv StringValue) String(builder Builder) Builder {
	builder.WriteString(strconv.Quote(string(sv)))
	return builder
}

func (biv BigintValue) String(builder Builder) Builder {
	builder.WriteString(biv.V.String())
	return builder
}

func (bdv BigdecValue) String(builder Builder) Builder {
	builder.WriteString(bdv.V.String())

	return builder
}

func (dbv DataByteValue) String(builder Builder) Builder {
	builder.WriteString(fmt.Sprintf("(%0X)", (dbv.GetByte())))

	return builder
}

func (av *ArrayValue) String(builder Builder) Builder {
	av.ProtectedString(builder, newSeenValues())
	return builder
}

func (av *ArrayValue) ProtectedString(builder Builder, seen *seenValues) Builder {
	if i := seen.IndexOf(av); i != -1 {
		builder.WriteString(fmt.Sprintf("ref@%d", i))
		return builder
	}

	seen.nc--
	if seen.nc < 0 {
		builder.WriteString("...")
		return builder
	}
	seen.Put(av)
	defer seen.Pop()

	if av.Data == nil {
		builder.WriteString("array[")
		for i, e := range av.List {
			e.ProtectedString(builder, seen)
			if i < len(av.List)-1 {
				builder.WriteString(",")
			}
		}
		builder.WriteString("]")
		// NOTE: we may want to unify the representation,
		// but for now tests expect this to be different.
		// This may be helpful for testing implementation behavior.
		return builder
	}
	if len(av.Data) > 256 {
		builder.WriteString(fmt.Sprintf("array[0x%X...]", av.Data[:256]))

		return builder
	}
	builder.WriteString(fmt.Sprintf("array[0x%X]", av.Data))
	return builder
}

func (sv *SliceValue) String(builder Builder) Builder {
	return sv.ProtectedString(builder, newSeenValues())
}

func (sv *SliceValue) ProtectedString(builder Builder, seen *seenValues) Builder {
	if sv.Base == nil {
		builder.WriteString(fmt.Sprint("nil-slice"))
		return builder
	}

	if i := seen.IndexOf(sv); i != -1 {
		builder.WriteString(fmt.Sprintf("ref@%d", i))
		return builder
	}

	if ref, ok := sv.Base.(RefValue); ok {
		builder.WriteString("slice[")
		ref.String(builder)
		builder.WriteString("]")
		return builder
	}

	seen.Put(sv)
	defer seen.Pop()

	vbase := sv.Base.(*ArrayValue)
	if vbase.Data == nil {
		builder.WriteString("slice[")
		for i, e := range vbase.List[sv.Offset : sv.Offset+sv.Length] {
			e.ProtectedString(builder, seen)
			if i < sv.Length-1 {
				builder.WriteString(",")
			}
		}
		builder.WriteString("]")

		return builder
	}
	if sv.Length > 256 {
		builder.WriteString(fmt.Sprintf("slice[0x%X...(%d)]", vbase.Data[sv.Offset:sv.Offset+256], sv.Length))

		return builder
	}
	builder.WriteString(fmt.Sprintf("slice[0x%X]", vbase.Data[sv.Offset:sv.Offset+sv.Length]))

	return builder
}

func (pv PointerValue) String(builder Builder) Builder {
	return pv.ProtectedString(builder, newSeenValues())
}

func (pv PointerValue) ProtectedString(builder Builder, seen *seenValues) Builder {
	if i := seen.IndexOf(pv); i != -1 {
		builder.WriteString(fmt.Sprintf("ref@%d", i))
		return builder
	}

	seen.Put(pv)
	defer seen.Pop()

	// Handle nil TV's, avoiding a nil pointer deref below.
	if pv.TV == nil {
		builder.WriteString("&<nil>")
		return builder
	}

	builder.WriteString("&")
	pv.TV.ProtectedString(builder, seen)
	return builder
}

func (sv *StructValue) String(builder Builder) Builder {
	return sv.ProtectedString(builder, newSeenValues())
}

func (sv *StructValue) ProtectedString(builder Builder, seen *seenValues) Builder {
	if i := seen.IndexOf(sv); i != -1 {
		builder.WriteString(fmt.Sprintf("ref@%d", i))

		return builder
	}

	seen.Put(sv)
	defer seen.Pop()

	builder.WriteString("struct{")
	for i, f := range sv.Fields {
		f.ProtectedString(builder, seen)
		if i < len(sv.Fields)-1 {
			builder.WriteString(",")
		}
	}
	builder.WriteString("}")

	return builder
}

func (fv *FuncValue) String(builder Builder) Builder {
	name := string(fv.Name)
	if fv.Type == nil {
		builder.WriteString(fmt.Sprintf("incomplete-func ?%s(?)?", name))
		return builder
	}
	if name == "" {
		fv.Type.String(builder)
		builder.WriteString("{...}")
		return builder
	}

	builder.WriteString(name)

	return builder
}

func (bmv *BoundMethodValue) String(builder Builder) Builder {
	name := bmv.Func.Name

	if ft, ok := bmv.Func.Type.(*FuncType); ok {
		builder.WriteString("<")
		ft.Params[0].Type.String(builder)
		builder.WriteString(">.")
		builder.WriteString(string(name) + "(")
		FieldTypeList(ft.Params).StringForFunc(builder)
		builder.WriteString(")")

		builder.WriteString("(")
		FieldTypeList(ft.Results).StringForFunc(builder)
		builder.WriteString(")")

		return builder
	}

	builder.WriteString(fmt.Sprintf("<?>.?%s(?)(?)",
		name))
	return builder
}

func (mv *MapValue) String(builder Builder) Builder {
	return mv.ProtectedString(builder, newSeenValues())
}

func (mv *MapValue) ProtectedString(builder Builder, seen *seenValues) Builder {
	if mv.List == nil {
		builder.WriteString("zero-map")
		return builder
	}

	if i := seen.IndexOf(mv); i != -1 {
		builder.WriteString(fmt.Sprintf("ref@%d", i))
		return builder
	}

	seen.Put(mv)
	defer seen.Pop()

	builder.WriteString("map{")
	next := mv.List.Head
	for next != nil {
		next.Key.ProtectedString(builder, seen)
		builder.WriteString(":")
		next.Value.ProtectedString(builder, seen)
		next = next.Next
		if next != nil {
			builder.WriteString(",")
		}
	}
	builder.WriteString("}")
	return builder
}

func (tv TypeValue) String(builder Builder) Builder {
	builder.WriteString("typeval{")
	tv.Type.String(builder)
	builder.WriteString("}")

	return builder
}

func (pv *PackageValue) String(builder Builder) Builder {
	builder.WriteString(fmt.Sprintf("package(%s %s)", pv.PkgName, pv.PkgPath))

	return builder
}

func (b *Block) String(builder Builder) Builder {
	return b.StringIndented(builder, "    ")
}

func (b *Block) StringIndented(builder Builder, indent string) Builder {
	source := toString(b.Source)
	if len(source) > 32 {
		source = source[:32] + "..."
	}
	builder.WriteString(fmt.Sprintf("%sBlock(ID:%v,Addr:%p,Source:%s,Parent:%p)",
		indent, b.ObjectInfo.ID, b, source, b.Parent)) // XXX Parent may be RefValue{}.
	if b.Source != nil {
		if _, ok := b.Source.(RefNode); ok {
			builder.WriteString(fmt.Sprintf("\n%s(RefNode names not shown)", indent))
		} else {
			types := b.Source.GetStaticBlock().Types
			for i, n := range b.Source.GetBlockNames() {
				if len(b.Values) <= i {
					builder.WriteString(fmt.Sprintf("\n%s%s: undefined static:%s", indent, n, types[i].String(builder)))
				} else {
					builder.WriteString(fmt.Sprintf("\n%s%s: %s static:%s",
						indent, n, b.Values[i].String(builder), types[i].String(builder)))
				}
			}
		}
	}
	return builder
}

func (rv RefValue) String(builder Builder) Builder {
	if rv.PkgPath == "" {
		builder.WriteString(fmt.Sprintf("ref(%v)",
			rv.ObjectID))

		return builder
	}

	builder.WriteString(fmt.Sprintf("ref(%s)",
		rv.PkgPath))

	return builder
}

func (hiv *HeapItemValue) String(builder Builder) Builder {
	builder.WriteString("heapitem(")
	hiv.Value.String(builder)
	builder.WriteString(")")
	return builder
}

// ----------------------------------------
// *TypedValue.Sprint

// for print() and println().
func (tv *TypedValue) Sprint(builder Builder, m *Machine) Builder {
	// if undefined, just "undefined".
	if tv == nil || tv.T == nil {
		builder.WriteString(undefinedStr)
		return builder
	}

	// if implements .String(), return it.
	if IsImplementedBy(gStringerType, tv.T) && !tv.IsNilInterface() {
		res := m.Eval(Call(Sel(&ConstExpr{TypedValue: *tv}, "String")))
		builder.WriteString(res[0].GetString())

		return builder
	}
	// if implements .Error(), return it.
	if IsImplementedBy(gErrorType, tv.T) {
		res := m.Eval(Call(Sel(&ConstExpr{TypedValue: *tv}, "Error")))
		builder.WriteString(res[0].GetString())
		return builder
	}

	return tv.ProtectedSprint(builder, newSeenValues(), true)
}

func (tv *TypedValue) ProtectedSprint(builder Builder, seen *seenValues, considerDeclaredType bool) Builder {
	if i := seen.IndexOf(tv.V); i != -1 {
		builder.WriteString(fmt.Sprintf("ref@%d", i))
		return builder
	}

	// print declared type
	if _, ok := tv.T.(*DeclaredType); ok && considerDeclaredType {
		return tv.ProtectedString(builder, seen)
	}

	// This is a special case that became necessary after adding `ProtectedString()` methods to
	// reliably prevent recursive print loops.
	if tv.V != nil {
		if v, ok := tv.V.(RefValue); ok {
			v.String(builder)
			return builder
		}
	}

	// otherwise, default behavior.
	switch bt := baseOf(tv.T).(type) {
	case PrimitiveType:
		switch bt {
		case UntypedBoolType, BoolType:
			builder.WriteString(fmt.Sprintf("%t", tv.GetBool()))

			return builder
		case UntypedStringType, StringType:
			builder.WriteString(tv.GetString())

			return builder
		case IntType:
			builder.WriteString(fmt.Sprintf("%d", tv.GetInt()))

			return builder
		case Int8Type:
			builder.WriteString(fmt.Sprintf("%d", tv.GetInt8()))

			return builder
		case Int16Type:
			builder.WriteString(fmt.Sprintf("%d", tv.GetInt16()))

			return builder
		case UntypedRuneType, Int32Type:
			builder.WriteString(fmt.Sprintf("%d", tv.GetInt32()))

			return builder
		case Int64Type:
			builder.WriteString(fmt.Sprintf("%d", tv.GetInt64()))

			return builder
		case UintType:
			builder.WriteString(fmt.Sprintf("%d", tv.GetUint()))
			return builder
		case Uint8Type:
			builder.WriteString(fmt.Sprintf("%d", tv.GetUint8()))
			return builder
		case DataByteType:
			builder.WriteString(fmt.Sprintf("%d", tv.GetDataByte()))

			return builder
		case Uint16Type:
			builder.WriteString(fmt.Sprintf("%d", tv.GetUint16()))

			return builder
		case Uint32Type:
			builder.WriteString(fmt.Sprintf("%d", tv.GetUint32()))

			return builder
		case Uint64Type:
			builder.WriteString(fmt.Sprintf("%d", tv.GetUint64()))

			return builder
		case Float32Type:
			builder.WriteString(fmt.Sprintf("%v", math.Float32frombits(tv.GetFloat32())))
			return builder
		case Float64Type:
			builder.WriteString(fmt.Sprintf("%v", math.Float64frombits(tv.GetFloat64())))
			return builder
		case UntypedBigintType:
			builder.WriteString(tv.V.(BigintValue).V.String())

			return builder
		case UntypedBigdecType:
			builder.WriteString(tv.V.(BigdecValue).V.String())

			return builder
		default:
			panic("should not happen")
		}
	case *PointerType:
		if tv.V == nil {
			builder.WriteString("typed-nil")

			return builder
		}
		roPre, roPost := "", ""
		if tv.IsReadonly() {
			roPre, roPost = "readonly(", ")"
		}
		builder.WriteString(roPre)
		tv.V.(PointerValue).ProtectedString(builder, seen)
		builder.WriteString(roPost)
		return builder
	case *FuncType:
		switch fv := tv.V.(type) {
		case nil:
			builder.WriteString(nilStr + " ")
			tv.T.String(builder)
			return builder
		case *FuncValue, *BoundMethodValue:
			fv.String(builder)

			return builder
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
		builder.WriteString(nilStr)

		return builder
	case *DeclaredType:
		panic("should not happen")
	case *PackageType:
		tv.V.(*PackageValue).String(builder)

		return builder
	case *ChanType:
		panic("not yet implemented")
	case *TypeType:
		tv.V.(TypeValue).String(builder)

		return builder
	default:
		// The remaining types may have a nil value.
		if tv.V == nil {
			builder.WriteString("(" + nilStr + " ")
			tv.T.String(builder)
			builder.WriteString(")")
			return builder
		}
		// Value may be N_Readonly
		roPre, roPost := "", ""
		if tv.IsReadonly() {
			roPre, roPost = "readonly(", ")"
		}
		// *ArrayType, *SliceType, *StructType, *MapType
		if ps, ok := tv.V.(protectedStringer); ok {
			builder.WriteString(roPre)
			ps.ProtectedString(builder, seen)
			builder.WriteString(roPost)
			return builder
		}

		// *NativeType
		builder.WriteString(roPre)
		tv.V.String(builder)
		builder.WriteString(roPost)
		return builder
	}
}

// ----------------------------------------
// TypedValue.String()

// For gno debugging/testing.
func (tv TypedValue) String(builder Builder) Builder {
	return tv.ProtectedString(builder, newSeenValues())
}

func (tv TypedValue) ProtectedString(builder Builder, seen *seenValues) Builder {
	if tv.IsUndefined() {
		builder.WriteString("(undefined)")
		return builder
	}
	if tv.V == nil {
		switch baseOf(tv.T) {
		case BoolType, UntypedBoolType:
			builder.WriteString(fmt.Sprintf("%t", tv.GetBool()))
		case StringType, UntypedStringType:
			builder.WriteString(fmt.Sprintf("%s", tv.GetString()))
		case IntType:
			builder.WriteString(fmt.Sprintf("%d", tv.GetInt()))
		case Int8Type:
			builder.WriteString(fmt.Sprintf("%d", tv.GetInt8()))
		case Int16Type:
			builder.WriteString(fmt.Sprintf("%d", tv.GetInt16()))
		case Int32Type, UntypedRuneType:
			builder.WriteString(fmt.Sprintf("%d", tv.GetInt32()))
		case Int64Type:
			builder.WriteString(fmt.Sprintf("%d", tv.GetInt64()))
		case UintType:
			builder.WriteString(fmt.Sprintf("%d", tv.GetUint()))
		case Uint8Type:
			builder.WriteString(fmt.Sprintf("%d", tv.GetUint8()))
		case DataByteType:
			builder.WriteString(fmt.Sprintf("%d", tv.GetDataByte()))
		case Uint16Type:
			builder.WriteString(fmt.Sprintf("%d", tv.GetUint16()))
		case Uint32Type:
			builder.WriteString(fmt.Sprintf("%d", tv.GetUint32()))
		case Uint64Type:
			builder.WriteString(fmt.Sprintf("%d", tv.GetUint64()))
		case Float32Type:
			builder.WriteString(fmt.Sprintf("%v", math.Float32frombits(tv.GetFloat32())))
		case Float64Type:
			builder.WriteString(fmt.Sprintf("%v", math.Float64frombits(tv.GetFloat64())))
		// Complex types that require recusion protection.
		default:
			builder.WriteString(nilStr)
		}
	} else {
		if base := baseOf(tv.T); base == StringType || base == UntypedStringType {
			tv.ProtectedSprint(builder, seen, false)
		} else {
			tv.ProtectedSprint(builder, seen, true)
		}
	}
	builder.WriteString(" ")
	tv.T.String(builder)
	builder.WriteString(")")
	return builder
}
