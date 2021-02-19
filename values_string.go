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
	return fmt.Sprintf("(%0X)", *(v.Ref))
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
	if v.Base.Data == nil {
		ss := make([]string, v.Length)
		for i, e := range v.Base.List[v.Offset : v.Offset+v.Length] {
			ss[i] = e.String()
		}
		return "slice[" + strings.Join(ss, ",") + "]"
	} else {
		return fmt.Sprintf("slice[0x%X]", v.Base.Data[v.Offset:v.Offset+v.Length])
	}
}

func (v PointerValue) String() string {
	return fmt.Sprintf("&%s", v.TypedValue.String())
}

func (v *StructValue) String() string {
	ss := make([]string, len(v.Fields))
	for i, f := range v.Fields {
		ss[i] = f.String()
	}
	return "struct{" + strings.Join(ss, ",") + "}"
}

func (v *FuncValue) String() string {
	recv := ""
	name := ""
	params := ""
	if v.Name != "" {
		name = string(v.Name)
	}
	if v.Type == nil {
		return fmt.Sprintf("incomplete-func ?%s(?)?", name)
	}
	if v.IsMethod {
		recv = "(" + v.Type.Params[0].String() + ") "
		params = FieldTypeList(v.Type.Params[1:]).StringWithCommas()
	} else {
		params = FieldTypeList(v.Type.Params).StringWithCommas()
	}
	results := ""
	if len(results) > 0 {
		results = FieldTypeList(v.Type.Results).StringWithCommas()
		results = "(" + results + ")"
	}
	return fmt.Sprintf("func %s%s(%s)%s",
		recv, name, params, results)
}

func (v BoundMethodValue) String() string {
	recvT := v.Func.Type.Params[0].Type.String()
	name := v.Func.Name
	params := FieldTypeList(v.Func.Type.Params).StringWithCommas()
	results := ""
	if len(results) > 0 {
		results = FieldTypeList(v.Func.Type.Results).StringWithCommas()
		results = "(" + results + ")"
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
	mthds := ""
	if d, ok := v.Type.(*DeclaredType); ok {
		mthds = fmt.Sprintf(" %v", d.Methods)
	}
	return fmt.Sprintf("typeval{%s%s%s}",
		v.Type.String(), ptr, mthds)
}

func (v *PackageValue) String() string {
	return fmt.Sprintf("package(%s)", v.PkgPath)
}

func (v nativeValue) String() string {
	return fmt.Sprintf("<go-native %v (%s)>",
		v.Value.Interface(),
		v.Value.String(),
	)
}

func (v blockValue) String() string {
	return fmt.Sprintf("block(%v)",
		v.Block)
}
