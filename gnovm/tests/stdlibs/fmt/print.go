package fmt

import (
	"unsafe"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
)

func X_typeString(v gnolang.TypedValue) string {
	if v.IsUndefined() {
		return "<nil>"
	}
	return v.T.String()
}

func X_valueOfInternal(v gnolang.TypedValue) (kind, declaredName string, bytes uint64, base gnolang.TypedValue) {
	if v.IsUndefined() {
		kind = "nil"
		return
	}
	if dt, ok := v.T.(*gnolang.DeclaredType); ok {
		declaredName = dt.String()
	}
	baseT := gnolang.BaseOf(v.T)
	base = gnolang.TypedValue{
		T: baseT,
		V: v.V,
		N: v.N,
	}
	switch baseT.Kind() {
	case gnolang.BoolKind:
		kind, bytes = "bool", v.GetUint64()&1
	case gnolang.StringKind:
		kind = "string"
	case gnolang.IntKind:
		kind, bytes = "int", v.GetUint64()
	case gnolang.Int8Kind:
		kind, bytes = "int8", v.GetUint64()&0xFF
	case gnolang.Int16Kind:
		kind, bytes = "int16", v.GetUint64()&0xFFFF
	case gnolang.Int32Kind:
		kind, bytes = "int32", v.GetUint64()&(1<<32-1)
	case gnolang.Int64Kind:
		kind, bytes = "int64", v.GetUint64()
	case gnolang.UintKind:
		kind, bytes = "uint", v.GetUint64()
	case gnolang.Uint8Kind:
		kind, bytes = "uint8", v.GetUint64()&0xFF
	case gnolang.Uint16Kind:
		kind, bytes = "uint16", v.GetUint64()&0xFFFF
	case gnolang.Uint32Kind:
		kind, bytes = "uint32", v.GetUint64()&(1<<32-1)
	case gnolang.Uint64Kind:
		kind, bytes = "uint64", v.GetUint64()
	case gnolang.Float32Kind:
		kind, bytes = "float32", v.GetUint64()&(1<<32-1)
	case gnolang.Float64Kind:
		kind, bytes = "float64", v.GetUint64()
	case gnolang.ArrayKind:
		kind = "array"
	case gnolang.SliceKind:
		kind = "slice"
	case gnolang.PointerKind:
		kind = "pointer"
	case gnolang.StructKind:
		kind = "struct"
	case gnolang.InterfaceKind:
		kind = "interface"
	case gnolang.FuncKind:
		kind = "func"
	case gnolang.MapKind:
		kind = "map"
	}
	return
}

func X_getAddr(m *gnolang.Machine, v gnolang.TypedValue) uint64 {
	switch v.T.Kind() {
	case gnolang.FuncKind, gnolang.MapKind, gnolang.SliceKind, gnolang.PointerKind:
	default:
		panic("invalid type")
	}
	switch v := v.V.(type) {
	case *gnolang.PointerValue:
		if v.TV == nil {
			return 0
		}
		return uint64(uintptr(unsafe.Pointer(v.TV))) ^
			uint64(uintptr(unsafe.Pointer(&v.Base))) ^
			uint64(v.Index) ^
			uint64(uintptr(unsafe.Pointer(v.Key)))
	case *gnolang.FuncValue:
		return uint64(uintptr(unsafe.Pointer(v)))
	case *gnolang.MapValue:
		return uint64(uintptr(unsafe.Pointer(v)))
	case *gnolang.SliceValue:
		return uint64(uintptr(unsafe.Pointer(v.GetBase(m.Store))))
	default:
		return 0
	}
}

func X_getPtrElem(v gnolang.TypedValue) gnolang.TypedValue {
	return v.V.(*gnolang.PointerValue).Deref()
}

var gSliceOfAny = &gnolang.SliceType{
	Elt: &gnolang.InterfaceType{},
}

func X_mapKeyValues(v gnolang.TypedValue) (keys, values gnolang.TypedValue) {
	if v.T.Kind() != gnolang.MapKind {
		panic("invalid arg to mapKeyValues")
	}

	mv := v.V.(*gnolang.MapValue)
	ks, vs := make([]gnolang.TypedValue, 0, mv.GetLength()), make([]gnolang.TypedValue, 0, mv.GetLength())
	for el := mv.List.Head; el != nil; el = el.Next {
		ks = append(ks, el.Key)
		vs = append(vs, el.Value)
	}

	// TODO: should sort keys withs something similar to fmtsort.
	keys.T = gSliceOfAny
	values.T = gSliceOfAny
	keys.V = &gnolang.SliceValue{
		Base: &gnolang.ArrayValue{
			List: ks,
		},
		Length: len(ks),
		Maxcap: len(ks),
	}
	values.V = &gnolang.SliceValue{
		Base: &gnolang.ArrayValue{
			List: vs,
		},
		Length: len(vs),
		Maxcap: len(vs),
	}
	return
}
