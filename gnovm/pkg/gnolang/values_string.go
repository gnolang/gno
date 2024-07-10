package gnolang

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
)

const stringByteLimit = 1024

type protectedWriter interface {
	ProtectedWrite(io.StringWriter, *seenValues) error
}

var errStringLimitExceeded = fmt.Errorf("string limit exceeded")

type limitedStringWriter struct {
	limit   int
	builder strings.Builder
}

func newLimitedStringWriter(limit int) *limitedStringWriter {
	return &limitedStringWriter{
		limit: limit,
	}
}

func (w *limitedStringWriter) WriteString(s string) (int, error) {
	var limiteExceeded bool
	if w.builder.Len()+len(s) > w.limit {
		s = s[:w.limit-w.builder.Len()] + "..."
		limiteExceeded = true
	}

	n, err := w.builder.WriteString(s)
	if err != nil {
		return n, err
	}

	if limiteExceeded {
		return n, errStringLimitExceeded
	}

	return n, nil
}

// This indicates the maximum anticipated depth of the stack when printing a Value type.
const defaultSeenValuesSize = 32

type seenValues struct {
	values []Value
}

func (sv *seenValues) Put(v Value) {
	sv.values = append(sv.values, v)
}

func (sv *seenValues) Contains(v Value) bool {
	for _, vv := range sv.values {
		if vv == v {
			return true
		}
	}

	return false
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
	return &seenValues{values: make([]Value, 0, defaultSeenValuesSize)}
}

func (v StringValue) String() string {
	w := newLimitedStringWriter(stringByteLimit)
	v.ProtectedWrite(w, nil)
	return w.builder.String()
}

func (v StringValue) ProtectedWrite(w io.StringWriter, _ *seenValues) error {
	_, err := w.WriteString(strconv.Quote(string(v)))
	return err
}

func (bv BigintValue) String() string {
	w := newLimitedStringWriter(stringByteLimit)
	bv.ProtectedWrite(w, nil)
	return w.builder.String()
}

func (bv BigintValue) ProtectedWrite(w io.StringWriter, _ *seenValues) error {
	_, err := w.WriteString(bv.V.String())
	return err
}

func (bv BigdecValue) String() string {
	w := newLimitedStringWriter(stringByteLimit)
	bv.ProtectedWrite(w, nil)
	return w.builder.String()
}

func (bv BigdecValue) ProtectedWrite(w io.StringWriter, _ *seenValues) error {
	_, err := w.WriteString(bv.V.String())
	return err
}

func (dbv DataByteValue) String() string {
	w := newLimitedStringWriter(stringByteLimit)
	dbv.ProtectedWrite(w, nil)
	return w.builder.String()
}

func (dbv DataByteValue) ProtectedWrite(w io.StringWriter, _ *seenValues) error {
	_, err := w.WriteString(fmt.Sprintf("(%0X)", dbv.GetByte()))
	return err
}

func (av *ArrayValue) String() string {
	w := newLimitedStringWriter(stringByteLimit)
	av.ProtectedWrite(w, newSeenValues())
	return w.builder.String()
}

func (av *ArrayValue) ProtectedWrite(w io.StringWriter, seen *seenValues) error {
	if seen.Contains(av) {
		_, err := w.WriteString(fmt.Sprintf("%p", av))
		return err
	}

	seen.Put(av)
	defer seen.Pop()

	if av.Data != nil {
		var suffix string
		bounds := len(av.Data)
		if bounds > 256 {
			bounds = 256
			suffix = "..."
		}

		_, err := w.WriteString(fmt.Sprintf("array[0x%X%s]", av.Data[:bounds], suffix))
		return err
	}

	if _, err := w.WriteString("array["); err != nil {
		return err
	}

	for i, e := range av.List {
		if err := e.ProtectedWrite(w, seen); err != nil {
			return err
		}

		if i < len(av.List)-1 {
			if _, err := w.WriteString(","); err != nil {
				return err
			}
		}
	}

	_, err := w.WriteString("]")
	return err
}

func (sv *SliceValue) String() string {
	w := newLimitedStringWriter(stringByteLimit)
	sv.ProtectedWrite(w, newSeenValues())
	return w.builder.String()
}

func (sv *SliceValue) ProtectedWrite(w io.StringWriter, seen *seenValues) error {
	if sv.Base == nil {
		_, err := w.WriteString("nil-slice")
		return err
	}

	if seen.Contains(sv) {
		_, err := w.WriteString(fmt.Sprintf("%p", sv))
		return err
	}

	if ref, ok := sv.Base.(RefValue); ok {
		_, err := w.WriteString(fmt.Sprintf("slice[%v]", ref))
		return err
	}

	seen.Put(sv)
	defer seen.Pop()

	vbase := sv.Base.(*ArrayValue)
	if vbase.Data != nil {
		var suffix string
		bounds := sv.Length
		if bounds > 256 {
			bounds = 256
			suffix = "..."
		}

		_, err := w.WriteString(fmt.Sprintf("slice[0x%X%s]", vbase.Data[sv.Offset:sv.Offset+bounds], suffix))
		return err
	}

	if _, err := w.WriteString("slice["); err != nil {
		return err
	}

	for i, e := range vbase.List[sv.Offset : sv.Offset+sv.Length] {
		if err := e.ProtectedWrite(w, seen); err != nil {
			return err
		}

		if i < sv.Length-1 {
			if _, err := w.WriteString(","); err != nil {
				return err
			}
		}
	}

	_, err := w.WriteString("]")
	return err
}

func (pv PointerValue) String() string {
	w := newLimitedStringWriter(stringByteLimit)
	pv.ProtectedWrite(w, newSeenValues())
	return w.builder.String()
}

func (pv PointerValue) ProtectedWrite(w io.StringWriter, seen *seenValues) error {
	if seen.Contains(pv) {
		_, err := w.WriteString(fmt.Sprintf("%p", &pv))
		return err
	}

	seen.Put(pv)
	defer seen.Pop()

	// Handle nil TV's, avoiding a nil pointer deref below.
	if pv.TV == nil {
		_, err := w.WriteString("&<nil>")
		return err
	}

	if _, err := w.WriteString("&"); err != nil {
		return err
	}

	return pv.TV.ProtectedWrite(w, seen)
}

func (sv *StructValue) String() string {
	w := newLimitedStringWriter(stringByteLimit)
	sv.ProtectedWrite(w, newSeenValues())
	return w.builder.String()
}

func (sv *StructValue) ProtectedWrite(w io.StringWriter, seen *seenValues) error {
	if seen.Contains(sv) {
		_, err := w.WriteString(fmt.Sprintf("%p", sv))
		return err
	}

	seen.Put(sv)
	defer seen.Pop()

	if _, err := w.WriteString("struct{"); err != nil {
		return err
	}

	for i, f := range sv.Fields {
		if err := f.ProtectedWrite(w, seen); err != nil {
			return err
		}

		if i < len(sv.Fields)-1 {
			if _, err := w.WriteString(","); err != nil {
				return err
			}
		}
	}

	_, err := w.WriteString("}")
	return err
}

func (fv *FuncValue) String() string {
	name := string(fv.Name)
	if fv.Type == nil {
		return fmt.Sprintf("incomplete-func ?%s(?)?", name)
	}
	return name
}

func (v *BoundMethodValue) String() string {
	name := v.Func.Name
	var (
		recvT   string = "?"
		params  string = "?"
		results string = "(?)"
	)
	if ft, ok := v.Func.Type.(*FuncType); ok {
		recvT = ft.Params[0].Type.String()
		params = FieldTypeList(ft.Params).StringWithCommas()
		if len(results) > 0 {
			results = FieldTypeList(ft.Results).StringWithCommas()
			results = "(" + results + ")"
		}
	}
	return fmt.Sprintf("<%s>.%s(%s)%s",
		recvT, name, params, results)
}

func (mv *MapValue) String() string {
	w := newLimitedStringWriter(stringByteLimit)
	mv.ProtectedWrite(w, newSeenValues())
	return w.builder.String()
}

func (mv *MapValue) ProtectedWrite(w io.StringWriter, seen *seenValues) error {
	if mv.List == nil {
		_, err := w.WriteString("zero-map")
		return err
	}

	if seen.Contains(mv) {
		_, err := w.WriteString(fmt.Sprintf("%p", mv))
		return err
	}

	seen.Put(mv)
	defer seen.Pop()

	if _, err := w.WriteString("map{"); err != nil {
		return err
	}

	next := mv.List.Head
	for next != nil {
		if err := next.Key.ProtectedWrite(w, seen); err != nil {
			return err
		}

		if _, err := w.WriteString(":"); err != nil {
			return err
		}

		if err := next.Value.ProtectedWrite(w, seen); err != nil {
			return err
		}

		if next.Next != nil {
			if _, err := w.WriteString(","); err != nil {
				return err
			}
		}

		next = next.Next
	}

	_, err := w.WriteString("}")
	return err
}

func (v TypeValue) String() string {
	var ptr string
	if reflect.TypeOf(v.Type).Kind() == reflect.Ptr {
		ptr = fmt.Sprintf(" (%p)", v.Type)
	}
	/*
		mthds := ""
		if d, ok := v.Type.(*DeclaredType); ok {
			mthds = fmt.Sprintf(" %v", d.Methods)
		}
	*/
	return fmt.Sprintf("typeval{%s%s}",
		v.Type.String(), ptr)
}

func (pv *PackageValue) String() string {
	return fmt.Sprintf("package(%s %s)", pv.PkgName, pv.PkgPath)
}

func (nv *NativeValue) String() string {
	return fmt.Sprintf("gonative{%v}",
		nv.Value.Interface())
	/*
		return fmt.Sprintf("gonative{%v (%s)}",
			v.Value.Interface(),
			v.Value.Type().String(),
		)
	*/
}

func (v RefValue) String() string {
	if v.PkgPath == "" {
		return fmt.Sprintf("ref(%v)",
			v.ObjectID)
	}
	return fmt.Sprintf("ref(%s)",
		v.PkgPath)
}

func (v *HeapItemValue) String() string {
	return fmt.Sprintf("heapitem(%v)",
		v.Value)
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
	if IsImplementedBy(gStringerType, tv.T) {
		// No worries here about the string being too large; the machine will exhaust
		// its cycles before that happens.
		res := m.Eval(Call(Sel(&ConstExpr{TypedValue: *tv}, "String")))
		return res[0].GetString()
	}
	// if implements .Error(), return it.
	if IsImplementedBy(gErrorType, tv.T) {
		res := m.Eval(Call(Sel(&ConstExpr{TypedValue: *tv}, "Error")))
		return res[0].GetString()
	}

	w := newLimitedStringWriter(stringByteLimit)
	tv.NonPrimitiveProtectedWrite(w, newSeenValues(), true)
	return w.builder.String()
}

func (tv *TypedValue) NonPrimitiveProtectedWrite(
	w io.StringWriter,
	seen *seenValues,
	considerDeclaredType bool,
) error {
	if seen.Contains(tv.V) {
		_, err := w.WriteString(fmt.Sprintf("%p", tv))
		return err
	}

	// print declared type
	if _, ok := tv.T.(*DeclaredType); ok && considerDeclaredType {
		return tv.ProtectedWrite(w, seen)
	}

	// This is a special case that became necessary after adding `ProtectedString()` methods to
	// reliably prevent recursive print loops.
	if tv.V != nil {
		if v, ok := tv.V.(RefValue); ok {
			_, err := w.WriteString(v.String())
			return err
		}
	}

	// otherwise, default behavior.
	switch bt := baseOf(tv.T).(type) {
	case PrimitiveType:
		switch bt {
		case UntypedBoolType, BoolType:
			_, err := w.WriteString(fmt.Sprintf("%t", tv.GetBool()))
			return err
		case UntypedStringType, StringType:
			_, err := w.WriteString(tv.GetString())
			return err
		case IntType:
			_, err := w.WriteString(fmt.Sprintf("%d", tv.GetInt()))
			return err
		case Int8Type:
			_, err := w.WriteString(fmt.Sprintf("%d", tv.GetInt8()))
			return err
		case Int16Type:
			_, err := w.WriteString(fmt.Sprintf("%d", tv.GetInt16()))
			return err
		case UntypedRuneType, Int32Type:
			_, err := w.WriteString(fmt.Sprintf("%d", tv.GetInt32()))
			return err
		case Int64Type:
			_, err := w.WriteString(fmt.Sprintf("%d", tv.GetInt64()))
			return err
		case UintType:
			_, err := w.WriteString(fmt.Sprintf("%d", tv.GetUint()))
			return err
		case Uint8Type:
			_, err := w.WriteString(fmt.Sprintf("%d", tv.GetUint8()))
			return err
		case DataByteType:
			_, err := w.WriteString(fmt.Sprintf("%d", tv.GetDataByte()))
			return err
		case Uint16Type:
			_, err := w.WriteString(fmt.Sprintf("%d", tv.GetUint16()))
			return err
		case Uint32Type:
			_, err := w.WriteString(fmt.Sprintf("%d", tv.GetUint32()))
			return err
		case Uint64Type:
			_, err := w.WriteString(fmt.Sprintf("%d", tv.GetUint64()))
			return err
		case Float32Type:
			_, err := w.WriteString(fmt.Sprintf("%v", tv.GetFloat32()))
			return err
		case Float64Type:
			_, err := w.WriteString(fmt.Sprintf("%v", tv.GetFloat64()))
			return err
		case UntypedBigintType, BigintType:
			_, err := w.WriteString(tv.V.(BigintValue).V.String())
			return err
		case UntypedBigdecType, BigdecType:
			_, err := w.WriteString(tv.V.(BigdecValue).V.String())
			return err
		default:
			panic(fmt.Sprintf("cannot print unknown primitive type %v", bt))
		}
	case *PointerType:
		if tv.V == nil {
			_, err := w.WriteString("invalid-pointer")
			return err
		}

		return tv.V.(PointerValue).ProtectedWrite(w, seen)
	case *FuncType:
		switch fv := tv.V.(type) {
		case nil:
			ft := tv.T.String()
			_, err := w.WriteString(nilStr + " " + ft)
			return err
		case *FuncValue, *BoundMethodValue:
			_, err := w.WriteString(fv.String())
			return err
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
		_, err := w.WriteString(nilStr)
		return err
	case *DeclaredType:
		panic("should not happen")
	case *PackageType:
		_, err := w.WriteString(tv.V.(*PackageValue).String())
		return err
	case *ChanType:
		panic("not yet implemented")
	case *TypeType:
		_, err := w.WriteString(tv.V.(TypeValue).String())
		return err
	default:
		// The remaining types may have a nil value.
		if tv.V == nil {
			_, err := w.WriteString("(" + nilStr + " " + tv.T.String() + ")")
			return err
		}

		// *ArrayType, *SliceType, *StructType, *MapType
		if ps, ok := tv.V.(protectedWriter); ok {
			return ps.ProtectedWrite(w, seen)
		} else if s, ok := tv.V.(fmt.Stringer); ok {
			// *NativeType
			_, err := w.WriteString(s.String())
			return err
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

// ----------------------------------------
// TypedValue.String()

// For gno debugging/testing.
func (tv TypedValue) String() string {
	w := newLimitedStringWriter(stringByteLimit)
	tv.ProtectedWrite(w, newSeenValues())
	return w.builder.String()
}

func (tv TypedValue) ProtectedWrite(w io.StringWriter, seen *seenValues) error {
	if tv.IsUndefined() {
		_, err := w.WriteString("(undefined)")
		return err
	}

	var vs string
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
			vs = fmt.Sprintf("%v", tv.GetFloat32())
		case Float64Type:
			vs = fmt.Sprintf("%v", tv.GetFloat64())
		// Complex types that require recusion protection.
		default:
			vs = nilStr
		}
	} else {
		base := baseOf(tv.T)
		quoteString := base == StringType || base == UntypedStringType
		if quoteString {
			if _, err := w.WriteString("\""); err != nil {
				return err
			}
		}

		if err := tv.NonPrimitiveProtectedWrite(w, seen, false); err != nil {
			return err
		}

		if quoteString {
			if _, err := w.WriteString("\""); err != nil {
				return err
			}
		}
	}

	ts := tv.T.String()
	_, err := w.WriteString(fmt.Sprintf("(%s %s)", vs, ts))
	return err
}
