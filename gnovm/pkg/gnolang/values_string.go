package gnolang

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/store/types"
)

type protectedStringer interface {
	ProtectedString(*seenValues) string
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

func (sv StringValue) String(m *Machine) string {
	return strconv.Quote(string(sv))
}

func (biv BigintValue) String(m *Machine) string {
	return biv.V.String()
}

func (bdv BigdecValue) String(m *Machine) string {
	return bdv.V.String()
}

func (dbv DataByteValue) String(m *Machine) string {
	return fmt.Sprintf("(%0X)", (dbv.GetByte()))
}

func (av *ArrayValue) String(m *Machine) string {
	return av.ProtectedString(m, newSeenValues())
}

const CPUCYCLES = "CPUCycles"

func (av *ArrayValue) ProtectedString(m *Machine, seen *seenValues) string {
	defer func() {
		// 7 characters for the array itself
		m.GasMeter.ConsumeGas(types.Gas(7*OpCharPrint), CPUCYCLES)
	}()

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
			m.GasMeter.ConsumeGas(OpCPUAssign, CPUCYCLES)
			ss[i] = e.ProtectedString(m, seen)
		}
		// NOTE: we may want to unify the representation,
		// but for now tests expect this to be different.
		// This may be helpful for testing implementation behavior.
		m.GasMeter.ConsumeGas(OpCPUAssign, CPUCYCLES)
		return "array[" + strings.Join(ss, ",") + "]"
	}
	if len(av.Data) > 256 {
		return fmt.Sprintf("array[0x%X...]", av.Data[:256])
	}
	return fmt.Sprintf("array[0x%X]", av.Data)
}

func (sv *SliceValue) String(m *Machine) string {
	return sv.ProtectedString(m, newSeenValues())
}

func (sv *SliceValue) ProtectedString(m *Machine, seen *seenValues) string {
	defer func() {
		// 7 characters for the slice itself
		m.GasMeter.ConsumeGas(types.Gas(7*OpCharPrint), CPUCYCLES)
	}()

	if sv.Base == nil {
		return "nil-slice"
	}

	if i := seen.IndexOf(sv); i != -1 {
		res := fmt.Sprintf("ref@%d", i)
		m.GasMeter.ConsumeGas(types.Gas(len(res)*OpCharPrint), CPUCYCLES)
		return res
	}

	if ref, ok := sv.Base.(RefValue); ok {
		res := fmt.Sprintf("slice[%v]", ref)
		m.GasMeter.ConsumeGas(types.Gas(len(res)*OpCharPrint), CPUCYCLES)
		return res
	}

	seen.Put(sv)
	defer seen.Pop()

	vbase := sv.Base.(*ArrayValue)
	if vbase.Data == nil {
		m.GasMeter.ConsumeGas(int64(sv.Length*OpCPUAssign), CPUCYCLES)
		ss := make([]string, sv.Length)
		for i, e := range vbase.List[sv.Offset : sv.Offset+sv.Length] {
			ss[i] = e.ProtectedString(m, seen)
		}
		return "slice[" + strings.Join(ss, ",") + "]"
	}

	if sv.Length > 256 {
		return fmt.Sprintf("slice[0x%X...(%d)]", vbase.Data[sv.Offset:sv.Offset+256], sv.Length)
	}
	return fmt.Sprintf("slice[0x%X]", vbase.Data[sv.Offset:sv.Offset+sv.Length])
}

func (pv PointerValue) String(m *Machine) string {
	return pv.ProtectedString(m, newSeenValues())
}

func (pv PointerValue) ProtectedString(m *Machine, seen *seenValues) string {
	if i := seen.IndexOf(pv); i != -1 {
		res := fmt.Sprintf("ref@%d", i)
		m.GasMeter.ConsumeGas(types.Gas(len(res)*OpCharPrint), CPUCYCLES)
		return res
	}

	seen.Put(pv)
	defer seen.Pop()

	// Handle nil TV's, avoiding a nil pointer deref below.
	if pv.TV == nil {
		res := "&<nil>"
		m.GasMeter.ConsumeGas(types.Gas(len(res)*OpCharPrint), CPUCYCLES)
		return res
	}

	res := fmt.Sprintf("&%s", pv.TV.ProtectedString(m, seen))
	m.GasMeter.ConsumeGas(types.Gas(len(res)*OpCharPrint), CPUCYCLES)
	return res
}

func (sv *StructValue) String(m *Machine) string {
	return sv.ProtectedString(m, newSeenValues())
}

func (sv *StructValue) ProtectedString(m *Machine, seen *seenValues) string {
	if i := seen.IndexOf(sv); i != -1 {
		res := fmt.Sprintf("ref@%d", i)
		m.GasMeter.ConsumeGas(types.Gas(len(res)*OpCharPrint), CPUCYCLES)
		return res
	}

	seen.Put(sv)
	defer seen.Pop()

	ss := make([]string, len(sv.Fields))
	for i, f := range sv.Fields {
		ss[i] = f.ProtectedString(m, seen)
	}

	// 8 characters for the struct
	// the fields will be accounted for in their own function call
	m.GasMeter.ConsumeGas(types.Gas(8*OpCharPrint), CPUCYCLES)
	return "struct{" + strings.Join(ss, ",") + "}"
}

func (fv *FuncValue) String(m *Machine) string {
	name := string(fv.Name)
	if fv.Type == nil {
		return fmt.Sprintf("incomplete-func ?%s(?)?", name)
	}
	if name == "" {
		return fmt.Sprintf("%s{...}", fv.Type.String())
	}
	return name
}

func (bmv *BoundMethodValue) String(m *Machine) string {
	name := bmv.Func.Name
	var (
		recvT   string = "?"
		params  string = "?"
		results string = "(?)"
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

func (mv *MapValue) String(m *Machine) string {
	return mv.ProtectedString(m, newSeenValues())
}

func (mv *MapValue) ProtectedString(m *Machine, seen *seenValues) string {
	if mv.List == nil {
		res := "zero-map"
		m.GasMeter.ConsumeGas(types.Gas(len(res)*OpCharPrint), CPUCYCLES)
		return res
	}

	if i := seen.IndexOf(mv); i != -1 {
		res := fmt.Sprintf("ref@%d", i)
		m.GasMeter.ConsumeGas(types.Gas(len(res)*OpCharPrint), CPUCYCLES)
		return res
	}

	seen.Put(mv)
	defer seen.Pop()

	ss := make([]string, 0, mv.GetLength())
	next := mv.List.Head
	for next != nil {
		m.GasMeter.ConsumeGas(OpCharPrint, CPUCYCLES)
		ss = append(ss,
			next.Key.ProtectedString(m, seen)+":"+
				next.Value.ProtectedString(m, seen))
		next = next.Next
	}

	m.GasMeter.ConsumeGas(types.Gas(5*OpCharPrint), CPUCYCLES)
	return "map{" + strings.Join(ss, ",") + "}"
}

func (tv TypeValue) String(m *Machine) string {
	ptr := ""
	if reflect.TypeOf(tv.Type).Kind() == reflect.Ptr {
		ptr = fmt.Sprintf(" (%p)", tv.Type)
	}
	/*
		mthds := ""
		if d, ok := tv.Type.(*DeclaredType); ok {
			mthds = fmt.Sprintf(" %v", d.Methods)
		}
	*/
	return fmt.Sprintf("typeval{%s%s}",
		tv.Type.String(), ptr)
}

func (pv *PackageValue) String(m *Machine) string {
	return fmt.Sprintf("package(%s %s)", pv.PkgName, pv.PkgPath)
}

func (rv RefValue) String(m *Machine) string {
	if rv.PkgPath == "" {
		return fmt.Sprintf("ref(%v)",
			rv.ObjectID)
	}
	return fmt.Sprintf("ref(%s)",
		rv.PkgPath)
}

func (hiv *HeapItemValue) String(m *Machine) string {
	return fmt.Sprintf("heapitem(%v)",
		hiv.Value)
}

// ----------------------------------------
// *TypedValue.Sprint

// for print() and println().
func (tv *TypedValue) Sprint(m *Machine) string {
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

	return tv.ProtectedSprint(m, newSeenValues(), true)
}

func (tv *TypedValue) ProtectedSprint(m *Machine, seen *seenValues, considerDeclaredType bool) string {
	if i := seen.IndexOf(tv.V); i != -1 {
		res := fmt.Sprintf("ref@%d", i)
		m.GasMeter.ConsumeGas(types.Gas(len(res)*OpCharPrint), CPUCYCLES)
		return res
	}

	// print declared type
	if _, ok := tv.T.(*DeclaredType); ok && considerDeclaredType {
		return tv.ProtectedString(m, seen)
	}

	// This is a special case that became necessary after adding `ProtectedString()` methods to
	// reliably prevent recursive print loops.
	if tv.V != nil {
		if v, ok := tv.V.(RefValue); ok {
			res := v.String(m)
			m.GasMeter.ConsumeGas(types.Gas(len(res)*OpCharPrint), CPUCYCLES)
			return res
		}
	}

	// otherwise, default behavior.
	switch bt := baseOf(tv.T).(type) {
	case PrimitiveType:
		switch bt {
		case UntypedBoolType, BoolType:
			res := fmt.Sprintf("%t", tv.GetBool())
			m.GasMeter.ConsumeGas(types.Gas(len(res)*OpCharPrint), CPUCYCLES)
			return res
		case UntypedStringType, StringType:
			res := tv.GetString()
			m.GasMeter.ConsumeGas(types.Gas(len(res)*OpCharPrint), CPUCYCLES)
			return res
		case IntType:
			res := fmt.Sprintf("%d", tv.GetInt())
			m.GasMeter.ConsumeGas(types.Gas(len(res)*OpCharPrint), CPUCYCLES)
			return res
		case Int8Type:
			res := fmt.Sprintf("%d", tv.GetInt8())
			m.GasMeter.ConsumeGas(types.Gas(len(res)*OpCharPrint), CPUCYCLES)
			return res
		case Int16Type:
			res := fmt.Sprintf("%d", tv.GetInt16())
			m.GasMeter.ConsumeGas(types.Gas(len(res)*OpCharPrint), CPUCYCLES)
			return res
		case UntypedRuneType, Int32Type:
			res := fmt.Sprintf("%d", tv.GetInt32())
			m.GasMeter.ConsumeGas(types.Gas(len(res)*OpCharPrint), CPUCYCLES)
			return res
		case Int64Type:
			res := fmt.Sprintf("%d", tv.GetInt64())
			m.GasMeter.ConsumeGas(types.Gas(len(res)*OpCharPrint), CPUCYCLES)
			return res
		case UintType:
			res := fmt.Sprintf("%d", tv.GetUint())
			m.GasMeter.ConsumeGas(types.Gas(len(res)*OpCharPrint), CPUCYCLES)
			return res
		case Uint8Type:
			res := fmt.Sprintf("%d", tv.GetUint8())
			m.GasMeter.ConsumeGas(types.Gas(len(res)*OpCharPrint), CPUCYCLES)
			return res
		case DataByteType:
			res := fmt.Sprintf("%d", tv.GetDataByte())
			m.GasMeter.ConsumeGas(types.Gas(len(res)*OpCharPrint), CPUCYCLES)
			return res
		case Uint16Type:
			res := fmt.Sprintf("%d", tv.GetUint16())
			m.GasMeter.ConsumeGas(types.Gas(len(res)*OpCharPrint), CPUCYCLES)
			return res
		case Uint32Type:
			res := fmt.Sprintf("%d", tv.GetUint32())
			m.GasMeter.ConsumeGas(types.Gas(len(res)*OpCharPrint), CPUCYCLES)
			return res
		case Uint64Type:
			res := fmt.Sprintf("%d", tv.GetUint64())
			m.GasMeter.ConsumeGas(types.Gas(len(res)*OpCharPrint), CPUCYCLES)
			return res
		case Float32Type:
			res := fmt.Sprintf("%v", math.Float32frombits(tv.GetFloat32()))
			m.GasMeter.ConsumeGas(types.Gas(len(res)*OpCharPrint), CPUCYCLES)
			return res
		case Float64Type:
			res := fmt.Sprintf("%v", math.Float64frombits(tv.GetFloat64()))
			m.GasMeter.ConsumeGas(types.Gas(len(res)*OpCharPrint), CPUCYCLES)
			return res
		case UntypedBigintType:
			res := tv.V.(BigintValue).V.String()
			m.GasMeter.ConsumeGas(types.Gas(len(res)*OpCharPrint), CPUCYCLES)
			return res
		case UntypedBigdecType:
			res := tv.V.(BigdecValue).V.String()
			m.GasMeter.ConsumeGas(types.Gas(len(res)*OpCharPrint), CPUCYCLES)
			return res
		default:
			panic("should not happen")
		}
	case *PointerType:
		if tv.V == nil {
			res := "invalid-pointer"
			m.GasMeter.ConsumeGas(types.Gas(len(res)*OpCharPrint), CPUCYCLES)
			return res
		}
		res := tv.V.(PointerValue).ProtectedString(m, seen)
		m.GasMeter.ConsumeGas(types.Gas(len(res)*OpCharPrint), CPUCYCLES)
		return res
	case *FuncType:
		switch fv := tv.V.(type) {
		case nil:
			ft := tv.T.String()
			res := nilStr + " " + ft
			m.GasMeter.ConsumeGas(types.Gas(len(res)*OpCharPrint), CPUCYCLES)
			return res
		case *FuncValue, *BoundMethodValue:
			res := fv.String(m)
			m.GasMeter.ConsumeGas(types.Gas(len(res)*OpCharPrint), CPUCYCLES)
			return res
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
		m.GasMeter.ConsumeGas(types.Gas(len(nilStr)*OpCharPrint), CPUCYCLES)
		return nilStr
	case *DeclaredType:
		panic("should not happen")
	case *PackageType:
		res := tv.V.(*PackageValue).String(m)
		m.GasMeter.ConsumeGas(types.Gas(len(res)*OpCharPrint), CPUCYCLES)
		return res
	case *ChanType:
		panic("not yet implemented")
	case *TypeType:
		res := tv.V.(TypeValue).String(m)
		m.GasMeter.ConsumeGas(types.Gas(len(res)*OpCharPrint), CPUCYCLES)
		return res
	default:
		// The remaining types may have a nil value.
		if tv.V == nil {
			res := "(" + nilStr + " " + tv.T.String() + ")"
			m.GasMeter.ConsumeGas(types.Gas(len(res)*OpCharPrint), CPUCYCLES)
			return res
		}

		// *ArrayType, *SliceType, *StructType, *MapType
		if ps, ok := tv.V.(protectedStringer); ok {
			res := ps.ProtectedString(seen)
			m.GasMeter.ConsumeGas(types.Gas(len(res)*OpCharPrint), CPUCYCLES)
			return res
		} else if s, ok := tv.V.(GasStringer); ok {
			// *NativeType
			res := s.String(m)
			m.GasMeter.ConsumeGas(types.Gas(len(res)*OpCharPrint), CPUCYCLES)
			return res
		}

		if debug {
			panic(fmt.Sprintf(
				"unexpected type %s",
				tv.T.String()))
		} else {
			panic("should not happen")
		}
	}
}

type GasStringer interface {
	String(m *Machine) string
}

// ----------------------------------------
// TypedValue.String()

// For gno debugging/testing.
func (tv TypedValue) String(m *Machine) string {
	return tv.ProtectedString(m, newSeenValues())
}

func (tv TypedValue) ProtectedString(m *Machine, seen *seenValues) string {
	m.GasMeter.ConsumeGas(OpCPUAssign, CPUCYCLES)

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
		vs = tv.ProtectedSprint(m, seen, false)
		if base := baseOf(tv.T); base == StringType || base == UntypedStringType {
			vs = strconv.Quote(vs)
		}
	}

	ts := tv.T.String()
	return fmt.Sprintf("(%s %s)", vs, ts) // TODO improve
}
