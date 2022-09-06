package gno

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

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
	ss := make([]string, len(v.List))
	if v.Data == nil {
		for i, e := range v.List {
			ss[i] = e.String()
		}
		// NOTE: we may want to unify the representation,
		// but for now tests expect this to be different.
		// This may be helpful for testing implementation behavior.
		return "array[" + strings.Join(ss, ",") + "]"
	} else {
		if len(v.Data) > 256 {
			return fmt.Sprintf("array[0x%X...]", v.Data[:256])
		} else {
			return fmt.Sprintf("array[0x%X]", v.Data)
		}
	}
}

func (v *SliceValue) String() string {
	if v.Base == nil {
		return "nil-slice"
	}
	if ref, ok := v.Base.(RefValue); ok {
		return fmt.Sprintf("slice[%v]", ref)
	}
	vbase := v.Base.(*ArrayValue)
	if vbase.Data == nil {
		ss := make([]string, v.Length)
		for i, e := range vbase.List[v.Offset : v.Offset+v.Length] {
			ss[i] = e.String()
		}
		return "slice[" + strings.Join(ss, ",") + "]"
	} else {
		if v.Length > 256 {
			return fmt.Sprintf("slice[0x%X...(%d)]", vbase.Data[v.Offset:v.Offset+256], v.Length)
		} else {
			return fmt.Sprintf("slice[0x%X]", vbase.Data[v.Offset:v.Offset+v.Length])
		}
	}
}

func (v PointerValue) String() string {
	// NOTE: cannot do below, due to recursion problems.
	// TODO: create a different String2(...) function.
	// return fmt.Sprintf("&%s", v.TypedValue.String())
	return fmt.Sprintf("&%p.(*%s)", v.TV, v.TV.T.String())
}

func (v *StructValue) String() string {
	ss := make([]string, len(v.Fields))
	for i, f := range v.Fields {
		ss[i] = f.String()
	}
	return "struct{" + strings.Join(ss, ",") + "}"
}

func (v *FuncValue) String() string {
	name := ""
	if v.Name != "" {
		name = string(v.Name)
	}
	if v.Type == nil {
		return fmt.Sprintf("incomplete-func ?%s(?)?", name)
	}
	return name
}

func (v *BoundMethodValue) String() string {
	name := v.Func.Name
	var recvT string
	var params string
	var results string
	if ft, ok := v.Func.Type.(*FuncType); ok {
		recvT = ft.Params[0].Type.String()
		params = FieldTypeList(ft.Params).StringWithCommas()
		if len(results) > 0 {
			results = FieldTypeList(ft.Results).StringWithCommas()
			results = "(" + results + ")"
		}
	} else {
		recvT = "?"
		params = "?"
		results = "(?)"
	}
	return fmt.Sprintf("<%s>.%s(%s)%s",
		recvT, name, params, results)
}

func (v *MapValue) String() string {
	if v.List == nil {
		return "zero-map"
	}
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
	} else {
		return fmt.Sprintf("ref(%s)",
			v.PkgPath)
	}
}

//----------------------------------------
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
	// otherwise, default behavior.
	switch bt := baseOf(tv.T).(type) {
	case PrimitiveType:
		switch bt {
		case UntypedBoolType, BoolType:
			return fmt.Sprintf("%t", tv.GetBool())
		case UntypedStringType, StringType:
			return string(tv.GetString())
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
		} else {
			return tv.V.(PointerValue).String()
		}
	case *ArrayType:
		return tv.V.(*ArrayValue).String()
	case *SliceType:
		return tv.V.(*SliceValue).String()
	case *StructType:
		return tv.V.(*StructValue).String()
	case *MapType:
		return tv.V.(*MapValue).String()
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
		return "nil"
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

//----------------------------------------
// TypedValue.String()

// For gno debugging/testing.
func (tv TypedValue) String() string {
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
		default:
			vs = "nil"
		}
	} else {
		vs = fmt.Sprintf("%v", tv.V)
	}
	ts := tv.T.String()
	return fmt.Sprintf("(%s %s)", vs, ts) // TODO improve
}
