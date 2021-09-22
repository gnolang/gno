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

func (v DataByteValue) String() string {
	return fmt.Sprintf("(%0X)", (v.GetByte()))
}

func (v *ArrayValue) String() string {
	ss := make([]string, len(v.List))
	for i, e := range v.List {
		ss[i] = e.String()
	}
	return "array[" + strings.Join(ss, ",") + "]"
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
		return fmt.Sprintf("slice[0x%X]", vbase.Data[v.Offset:v.Offset+v.Length])
	}
}

func (v PointerValue) String() string {
	// NOTE: cannot do below, due to recursion problems.
	// TODO: create a different String2(...) function.
	// return fmt.Sprintf("&%s", v.TypedValue.String())
	return fmt.Sprintf("&%p (*%s)", v.TV, v.TV.T.String())
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

func (v nativeValue) String() string {
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
	return fmt.Sprintf("ref(%v)",
		v.ObjectID)
}
