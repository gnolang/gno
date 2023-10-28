package gnolang

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const recursed string = "<recursed>"

func (v StringValue) String() string {
	return strconv.Quote(string(v))
}

func (v BigintValue) String() string {
	return v.V.String()
}

func (v BigdecValue) String() string {
	return v.V.String()
}

func (v DataByteValue) String() string {
	return fmt.Sprintf("(%0X)", (v.GetByte()))
}

func (v *ArrayValue) String() string {
	return v.ProtectedString(map[Value]struct{}{})
}

func (v *ArrayValue) ProtectedString(seen map[Value]struct{}) string {
	if _, ok := seen[v]; ok {
		return recursed
	}

	seen[v] = struct{}{}
	ss := make([]string, len(v.List))
	if v.Data == nil {
		for i, e := range v.List {
			ss[i] = e.ProtectedString(seen)
		}
		// NOTE: we may want to unify the representation,
		// but for now tests expect this to be different.
		// This may be helpful for testing implementation behavior.
		return "array[" + strings.Join(ss, ",") + "]"
	}
	if len(v.Data) > 256 {
		return fmt.Sprintf("array[0x%X...]", v.Data[:256])
	}
	return fmt.Sprintf("array[0x%X]", v.Data)
}

func (v *SliceValue) String() string {
	return v.ProtectedString(map[Value]struct{}{})
}

func (v *SliceValue) ProtectedString(seen map[Value]struct{}) string {
	if v.Base == nil {
		return "nil-slice"
	}

	if _, ok := seen[v]; ok {
		return recursed
	}

	if ref, ok := v.Base.(RefValue); ok {
		return fmt.Sprintf("slice[%v]", ref)
	}

	seen[v] = struct{}{}
	vbase := v.Base.(*ArrayValue)
	if vbase.Data == nil {
		ss := make([]string, v.Length)
		for i, e := range vbase.List[v.Offset : v.Offset+v.Length] {
			ss[i] = e.ProtectedString(seen)
		}
		return "slice[" + strings.Join(ss, ",") + "]"
	}
	if v.Length > 256 {
		return fmt.Sprintf("slice[0x%X...(%d)]", vbase.Data[v.Offset:v.Offset+256], v.Length)
	}
	return fmt.Sprintf("slice[0x%X]", vbase.Data[v.Offset:v.Offset+v.Length])
}

func (v PointerValue) String() string {
	return v.ProtectedString(map[Value]struct{}{})
}

func (v PointerValue) ProtectedString(seen map[Value]struct{}) string {
	if _, ok := seen[v]; ok {
		return recursed
	}

	seen[v] = struct{}{}
	return fmt.Sprintf("&%s", v.TV.ProtectedString(seen))
}

func (v *StructValue) String() string {
	return v.ProtectedString(map[Value]struct{}{})
}

func (v *StructValue) ProtectedString(seen map[Value]struct{}) string {
	if _, ok := seen[v]; ok {
		return recursed
	}

	seen[v] = struct{}{}
	ss := make([]string, len(v.Fields))
	for i, f := range v.Fields {
		ss[i] = f.ProtectedString(seen)
	}
	return "struct{" + strings.Join(ss, ",") + "}"
}

func (v *FuncValue) String() string {
	name := string(v.Name)
	if v.Type == nil {
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

func (v *MapValue) String() string {
	return v.ProtectedString(map[Value]struct{}{})
}

func (v *MapValue) ProtectedString(seen map[Value]struct{}) string {
	if v.List == nil {
		return "zero-map"
	}

	if _, ok := seen[v]; ok {
		return recursed
	}

	seen[v] = struct{}{}
	ss := make([]string, 0, v.GetLength())
	next := v.List.Head
	for next != nil {
		ss = append(ss,
			next.Key.String()+":"+
				next.Value.String())
		next = next.Next
	}
	return "map{" + strings.Join(ss, ",") + "}"
}

func (v TypeValue) String() string {
	ptr := ""
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

func (v *PackageValue) String() string {
	return fmt.Sprintf("package(%s %s)", v.PkgName, v.PkgPath)
}

func (v *NativeValue) String() string {
	return fmt.Sprintf("gonative{%v}",
		v.Value.Interface())
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

// ----------------------------------------
// *TypedValue.Sprint

// for print() and println().
func (tv *TypedValue) Sprint(m *Machine) string {
	// if undefined, just "undefined".
	if tv == nil || tv.T == nil {
		return "undefined"
	}
	// if implements .String(), return it.
	if IsImplementedBy(gStringerType, tv.T) {
		res := m.Eval(Call(Sel(&ConstExpr{TypedValue: *tv}, "String")))
		return res[0].GetString()
	}
	// if implements .Error(), return it.
	if IsImplementedBy(gErrorType, tv.T) {
		res := m.Eval(Call(Sel(&ConstExpr{TypedValue: *tv}, "Error")))
		return res[0].GetString()
	}

	return tv.ProtectedSprint(map[Value]struct{}{}, true)
}

func (tv *TypedValue) ProtectedSprint(seen map[Value]struct{}, considerDeclaredType bool) string {

	if _, ok := seen[tv.V]; ok {
		return recursed
	}

	// print declared type
	if _, ok := tv.T.(*DeclaredType); ok && considerDeclaredType {
		return tv.ProtectedString(seen)
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
		case Uint16Type:
			return fmt.Sprintf("%d", tv.GetUint16())
		case Uint32Type:
			return fmt.Sprintf("%d", tv.GetUint32())
		case Uint64Type:
			return fmt.Sprintf("%d", tv.GetUint64())
		case Float32Type:
			return fmt.Sprintf("%v", tv.GetFloat32())
		case Float64Type:
			return fmt.Sprintf("%v", tv.GetFloat64())
		case UntypedBigintType, BigintType:
			return tv.V.(BigintValue).V.String()
		case UntypedBigdecType, BigdecType:
			return tv.V.(BigdecValue).V.String()
		default:
			panic("should not happen")
		}
	case *PointerType:
		if tv.V == nil {
			return "invalid-pointer"
		}
		return tv.V.(PointerValue).ProtectedString(seen)
	case *ArrayType:
		return tv.V.(*ArrayValue).ProtectedString(seen)
	case *SliceType:
		return tv.V.(*SliceValue).ProtectedString(seen)
	case *StructType:
		return tv.V.(*StructValue).ProtectedString(seen)
	case *MapType:
		return tv.V.(*MapValue).ProtectedString(seen)
	case *FuncType:
		switch fv := tv.V.(type) {
		case nil:
			ft := tv.T.String()
			return "nil " + ft
		case *FuncValue:
			return fv.String()
		case *BoundMethodValue:
			return fv.String()
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
	case *TypeType:
		return tv.V.(TypeValue).String()
	case *DeclaredType:
		panic("should not happen")
	case *PackageType:
		return tv.V.(*PackageValue).String()
	case *ChanType:
		panic("not yet implemented")
		// return tv.V.(*ChanValue).String()
	case *NativeType:
		return fmt.Sprintf("%v",
			tv.V.(*NativeValue).Value.Interface())
	default:
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
	return tv.ProtectedString(map[Value]struct{}{})
}

func (tv TypedValue) ProtectedString(seen map[Value]struct{}) string {
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
			vs = fmt.Sprintf("%v", tv.GetFloat32())
		case Float64Type:
			vs = fmt.Sprintf("%v", tv.GetFloat64())
		// Complex types that require recusion protection.
		default:
			vs = nilStr
		}
	} else {
		// vs = fmt.Sprintf("%v", tv.V)
		vs = tv.ProtectedSprint(seen, false)
		if base := baseOf(tv.T); base == StringType || base == UntypedStringType {
			vs = strconv.Quote(vs)
		}
	}

	ts := tv.T.String()
	return fmt.Sprintf("(%s %s)", vs, ts) // TODO improve
}
