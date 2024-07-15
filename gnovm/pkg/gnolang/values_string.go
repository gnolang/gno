package gnolang

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const stringByteLimit = 1024

type protectedWriter interface {
	ProtectedWrite(*limitedValueStringWriter, *seenValues) error
}

var errStringLimitExceeded = fmt.Errorf("string limit exceeded")

type limitedValueStringWriter struct {
	limit      int
	quoteDepth int
	builder    strings.Builder
}

func newLimitedStringValueWriter(limit int) *limitedValueStringWriter {
	return &limitedValueStringWriter{
		limit: limit,
	}
}

func (w *limitedValueStringWriter) BeginQuotation() {
	slashes := strings.Repeat(`\\`, w.quoteDepth)
	w.builder.WriteString(slashes + `"`)
	w.quoteDepth++
}

func (w *limitedValueStringWriter) EndQuotation() {
	w.quoteDepth--
	slashes := strings.Repeat(`\\`, w.quoteDepth)
	w.builder.WriteString(slashes + `"`)
}

func (w *limitedValueStringWriter) WriteValueString(s string) error {
	var limitExceeded bool
	if w.builder.Len()+len(s) > w.limit {
		s = s[:w.limit-w.builder.Len()] + "..."
		limitExceeded = true
	}

	if _, err := w.builder.WriteString(s); err != nil {
		return err
	}

	if limitExceeded {
		return errStringLimitExceeded
	}

	return nil
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
	w := newLimitedStringValueWriter(stringByteLimit)
	w.BeginQuotation()
	v.ProtectedWrite(w, nil)
	w.EndQuotation()
	return w.builder.String()
}

func (v StringValue) ProtectedWrite(w *limitedValueStringWriter, _ *seenValues) error {
	quoted := string(v)

	// If `quoteDepth` is not zero, then we are already in a quoted string, so quote this
	// string to escape any inner quotes. Also remove the outer quotes after escaping the
	// inner quotes because the beginning outer quote has already been written and the
	// ending outer quote will be written afterwards.
	for i := 1; i <= w.quoteDepth; i++ {
		quoted = strconv.Quote(quoted)
		quoted = quoted[1 : len(quoted)-1]
	}

	return w.WriteValueString(quoted)
}

func (bv BigintValue) String() string {
	w := newLimitedStringValueWriter(stringByteLimit)
	bv.ProtectedWrite(w, nil)
	return w.builder.String()
}

func (bv BigintValue) ProtectedWrite(w *limitedValueStringWriter, _ *seenValues) error {
	return w.WriteValueString(bv.V.String())
}

func (bv BigdecValue) String() string {
	w := newLimitedStringValueWriter(stringByteLimit)
	bv.ProtectedWrite(w, nil)
	return w.builder.String()
}

func (bv BigdecValue) ProtectedWrite(w *limitedValueStringWriter, _ *seenValues) error {
	return w.WriteValueString(bv.V.String())
}

func (dbv DataByteValue) String() string {
	w := newLimitedStringValueWriter(stringByteLimit)
	dbv.ProtectedWrite(w, nil)
	return w.builder.String()
}

func (dbv DataByteValue) ProtectedWrite(w *limitedValueStringWriter, _ *seenValues) error {
	return w.WriteValueString(fmt.Sprintf("(%0X)", dbv.GetByte()))
}

func (av *ArrayValue) String() string {
	w := newLimitedStringValueWriter(stringByteLimit)
	av.ProtectedWrite(w, newSeenValues())
	return w.builder.String()
}

func (av *ArrayValue) ProtectedWrite(w *limitedValueStringWriter, seen *seenValues) error {
	if seen.Contains(av) {
		return w.WriteValueString(fmt.Sprintf("%p", av))
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

		return w.WriteValueString(fmt.Sprintf("array[0x%X%s]", av.Data[:bounds], suffix))
	}

	if err := w.WriteValueString("array["); err != nil {
		return err
	}

	for i, e := range av.List {
		if err := e.ProtectedWrite(w, seen); err != nil {
			return err
		}

		if i < len(av.List)-1 {
			if err := w.WriteValueString(","); err != nil {
				return err
			}
		}
	}

	return w.WriteValueString("]")
}

func (sv *SliceValue) String() string {
	w := newLimitedStringValueWriter(stringByteLimit)
	sv.ProtectedWrite(w, newSeenValues())
	return w.builder.String()
}

func (sv *SliceValue) ProtectedWrite(w *limitedValueStringWriter, seen *seenValues) error {
	if sv.Base == nil {
		return w.WriteValueString("nil-slice")
	}

	if seen.Contains(sv) {
		return w.WriteValueString(fmt.Sprintf("%p", sv))
	}

	if ref, ok := sv.Base.(RefValue); ok {
		return w.WriteValueString(fmt.Sprintf("slice[%v]", ref))
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

		return w.WriteValueString(fmt.Sprintf("slice[0x%X%s]", vbase.Data[sv.Offset:sv.Offset+bounds], suffix))
	}

	if err := w.WriteValueString("slice["); err != nil {
		return err
	}

	for i, e := range vbase.List[sv.Offset : sv.Offset+sv.Length] {
		if err := e.ProtectedWrite(w, seen); err != nil {
			return err
		}

		if i < sv.Length-1 {
			if err := w.WriteValueString(","); err != nil {
				return err
			}
		}
	}

	return w.WriteValueString("]")
}

func (pv PointerValue) String() string {
	w := newLimitedStringValueWriter(stringByteLimit)
	pv.ProtectedWrite(w, newSeenValues())
	return w.builder.String()
}

func (pv PointerValue) ProtectedWrite(w *limitedValueStringWriter, seen *seenValues) error {
	if seen.Contains(pv) {
		return w.WriteValueString(fmt.Sprintf("%p", &pv))
	}

	seen.Put(pv)
	defer seen.Pop()

	// Handle nil TV's, avoiding a nil pointer deref below.
	if pv.TV == nil {
		return w.WriteValueString("&<nil>")
	}

	if err := w.WriteValueString("&"); err != nil {
		return err
	}

	return pv.TV.ProtectedWrite(w, seen)
}

func (sv *StructValue) String() string {
	w := newLimitedStringValueWriter(stringByteLimit)
	sv.ProtectedWrite(w, newSeenValues())
	return w.builder.String()
}

func (sv *StructValue) ProtectedWrite(w *limitedValueStringWriter, seen *seenValues) error {
	if seen.Contains(sv) {
		return w.WriteValueString(fmt.Sprintf("%p", sv))
	}

	seen.Put(sv)
	defer seen.Pop()

	if err := w.WriteValueString("struct{"); err != nil {
		return err
	}

	for i, f := range sv.Fields {
		if err := f.ProtectedWrite(w, seen); err != nil {
			return err
		}

		if i < len(sv.Fields)-1 {
			if err := w.WriteValueString(","); err != nil {
				return err
			}
		}
	}

	return w.WriteValueString("}")
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
	w := newLimitedStringValueWriter(stringByteLimit)
	mv.ProtectedWrite(w, newSeenValues())
	return w.builder.String()
}

func (mv *MapValue) ProtectedWrite(w *limitedValueStringWriter, seen *seenValues) error {
	if mv.List == nil {
		return w.WriteValueString("zero-map")
	}

	if seen.Contains(mv) {
		return w.WriteValueString(fmt.Sprintf("%p", mv))
	}

	seen.Put(mv)
	defer seen.Pop()

	if err := w.WriteValueString("map{"); err != nil {
		return err
	}

	next := mv.List.Head
	for next != nil {
		if err := next.Key.ProtectedWrite(w, seen); err != nil {
			return err
		}

		if err := w.WriteValueString(":"); err != nil {
			return err
		}

		if err := next.Value.ProtectedWrite(w, seen); err != nil {
			return err
		}

		if next.Next != nil {
			if err := w.WriteValueString(","); err != nil {
				return err
			}
		}

		next = next.Next
	}

	return w.WriteValueString("}")
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

	w := newLimitedStringValueWriter(stringByteLimit)
	tv.NonPrimitiveProtectedWrite(w, newSeenValues(), true)
	return w.builder.String()
}

func (tv *TypedValue) NonPrimitiveProtectedWrite(
	w *limitedValueStringWriter,
	seen *seenValues,
	considerDeclaredType bool,
) error {
	if seen.Contains(tv.V) {
		return w.WriteValueString(fmt.Sprintf("%p", tv))
	}

	// print declared type
	if _, ok := tv.T.(*DeclaredType); ok && considerDeclaredType {
		return tv.ProtectedWrite(w, seen)
	}

	// This is a special case that became necessary after adding `ProtectedString()` methods to
	// reliably prevent recursive print loops.
	if tv.V != nil {
		if v, ok := tv.V.(RefValue); ok {
			return w.WriteValueString(v.String())
		}
	}

	// otherwise, default behavior.
	switch bt := baseOf(tv.T).(type) {
	case PrimitiveType:
		switch bt {
		case UntypedBoolType, BoolType:
			return w.WriteValueString(fmt.Sprintf("%t", tv.GetBool()))
		case UntypedStringType, StringType:
			return tv.V.(StringValue).ProtectedWrite(w, seen)
		case IntType:
			return w.WriteValueString(fmt.Sprintf("%d", tv.GetInt()))
		case Int8Type:
			return w.WriteValueString(fmt.Sprintf("%d", tv.GetInt8()))
		case Int16Type:
			return w.WriteValueString(fmt.Sprintf("%d", tv.GetInt16()))
		case UntypedRuneType, Int32Type:
			return w.WriteValueString(fmt.Sprintf("%d", tv.GetInt32()))
		case Int64Type:
			return w.WriteValueString(fmt.Sprintf("%d", tv.GetInt64()))
		case UintType:
			return w.WriteValueString(fmt.Sprintf("%d", tv.GetUint()))
		case Uint8Type:
			return w.WriteValueString(fmt.Sprintf("%d", tv.GetUint8()))
		case DataByteType:
			return w.WriteValueString(fmt.Sprintf("%d", tv.GetDataByte()))
		case Uint16Type:
			return w.WriteValueString(fmt.Sprintf("%d", tv.GetUint16()))
		case Uint32Type:
			return w.WriteValueString(fmt.Sprintf("%d", tv.GetUint32()))
		case Uint64Type:
			return w.WriteValueString(fmt.Sprintf("%d", tv.GetUint64()))
		case Float32Type:
			return w.WriteValueString(fmt.Sprintf("%v", tv.GetFloat32()))
		case Float64Type:
			return w.WriteValueString(fmt.Sprintf("%v", tv.GetFloat64()))
		case UntypedBigintType, BigintType:
			return w.WriteValueString(tv.V.(BigintValue).V.String())
		case UntypedBigdecType, BigdecType:
			return w.WriteValueString(tv.V.(BigdecValue).V.String())
		default:
			panic(fmt.Sprintf("cannot print unknown primitive type %v", bt))
		}
	case *PointerType:
		if tv.V == nil {
			return w.WriteValueString("invalid-pointer")
		}

		return tv.V.(PointerValue).ProtectedWrite(w, seen)
	case *FuncType:
		switch fv := tv.V.(type) {
		case nil:
			ft := tv.T.String()
			return w.WriteValueString(nilStr + " " + ft)
		case *FuncValue, *BoundMethodValue:
			return w.WriteValueString(fv.String())
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
		return w.WriteValueString(nilStr)
	case *DeclaredType:
		panic("should not happen")
	case *PackageType:
		return w.WriteValueString(tv.V.(*PackageValue).String())
	case *ChanType:
		panic("not yet implemented")
	case *TypeType:
		return w.WriteValueString(tv.V.(TypeValue).String())
	default:
		// The remaining types may have a nil value.
		if tv.V == nil {
			return w.WriteValueString("(" + nilStr + " " + tv.T.String() + ")")
		}

		// *ArrayType, *SliceType, *StructType, *MapType
		if ps, ok := tv.V.(protectedWriter); ok {
			return ps.ProtectedWrite(w, seen)
		} else if s, ok := tv.V.(fmt.Stringer); ok {
			// *NativeType
			return w.WriteValueString(s.String())
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
	w := newLimitedStringValueWriter(stringByteLimit)
	tv.ProtectedWrite(w, newSeenValues())
	return w.builder.String()
}

func (tv TypedValue) ProtectedWrite(w *limitedValueStringWriter, seen *seenValues) error {
	if tv.IsUndefined() {
		return w.WriteValueString("(undefined)")
	}

	if tv.V == nil {
		var vs string
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

		return w.WriteValueString("(" + vs + " " + tv.T.String() + ")")
	}

	if err := w.WriteValueString("("); err != nil {
		return err
	}

	base := baseOf(tv.T)
	quoteString := base == StringType || base == UntypedStringType
	if quoteString {
		w.BeginQuotation()
	}

	if err := tv.NonPrimitiveProtectedWrite(w, seen, false); err != nil {
		return err
	}

	if quoteString {
		w.EndQuotation()
	}

	if err := w.WriteValueString(" " + tv.T.String() + ")"); err != nil {
		return err
	}

	return nil
}
