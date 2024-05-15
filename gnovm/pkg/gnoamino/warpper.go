// Most of the methods here are inspired by `gnonative.go` but are now included here for specific reasons:
//   - Methods in `gnonative` are meant for general use and may be removed later, while here they are focused
//     on marshaling and unmarshaling only.
//   - These methods should be simpler as they don't require the full support of reflection, enabling a more
//     detailed implementation.
//   - We expect familiarity with the TypedValue type, making its handling easier.
//   - We will probably implement Gas comsumtion within those methods.
package gnoamino

import (
	"errors"
	"fmt"
	"reflect"
	"unicode"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
)

var (
	ErrRecursivePointer = errors.New("recusive pointer aren't supported")
	ErrUnsuportedType   = errors.New("unsuported type")
)

// Not thread safe
type wrapperTypedValue struct {
	*gnolang.TypedValue

	Allocator *gnolang.Allocator
	// Store     gnolang.Store
}

var tInterface = reflect.TypeOf(new(interface{})).Elem()

// NOTE: tv.TypedValue.T need to be filled in order to be able guess the type
func (tva wrapperTypedValue) TypeDesc() reflect.Type {
	if tva.T == nil { // Unable to guess Go Type
		return tInterface
	}

	return tva.GoType()
}

func (tva wrapperTypedValue) GnoValue() *gnolang.TypedValue {
	return tva.TypedValue
}

func (tva wrapperTypedValue) MarshalAmino() (interface{}, error) {
	return tva.GoValue().Interface(), nil
}

func (tva *wrapperTypedValue) UnmarshalAmino(i interface{}) error {
	visited := map[uintptr]struct{}{} // keep track of visited ptraddr
	rv := reflect.ValueOf(i)
	tva.unmarshalValue(rv, visited)
	return nil
}

func (tva *wrapperTypedValue) GoValue() (ret reflect.Value) {
	return gno2GoValue(tva.TypedValue, reflect.Value{})
}

func (tva *wrapperTypedValue) GoType() reflect.Type {
	visited := map[uintptr]struct{}{} // keep track of visited ptraddr
	return gno2GoType(tva.T, visited)
}

func (tva *wrapperTypedValue) newWith(tv *gnolang.TypedValue) *wrapperTypedValue {
	return &wrapperTypedValue{
		TypedValue: tv,
		Allocator:  tva.Allocator,
		// Store:      tva.Store,
	}
}

// ----------------------------------------
// Gno to Go conversion

func (tva *wrapperTypedValue) unmarshalValue(rv reflect.Value, visited map[uintptr]struct{}) {
	if addr, ok := isPointer(rv); ok {
		if _, ok := visited[addr]; ok {
			panic(ErrRecursivePointer)
		}
		visited[addr] = struct{}{}
		defer func() { delete(visited, addr) }()
	}

	if tva.T == nil {
		panic("unable to unmarshal TypedValue with no type")
	}

	// special case if t == Float32Type or Float64Type
	switch ct := gnolang.BaseOf(tva.T).(type) {
	case gnolang.PrimitiveType:
		switch ct {
		case gnolang.BoolType, gnolang.UntypedBoolType:
			tva.SetBool(rv.Bool())
		case gnolang.StringType, gnolang.UntypedStringType:
			tva.V = tva.Allocator.NewString(rv.String())
		case gnolang.IntType:
			tva.SetInt(int(rv.Int()))
		case gnolang.Int8Type:
			tva.SetInt8(int8(rv.Int()))
		case gnolang.Int16Type:
			tva.SetInt16(int16(rv.Int()))
		case gnolang.Int32Type, gnolang.UntypedRuneType:
			tva.SetInt32(int32(rv.Int()))
		case gnolang.Int64Type:
			tva.SetInt64(rv.Int())
		case gnolang.UintType:
			tva.SetUint(uint(rv.Uint()))
		case gnolang.Uint8Type:
			tva.SetUint8(uint8(rv.Uint()))
		case gnolang.Uint16Type:
			tva.SetUint16(uint16(rv.Uint()))
		case gnolang.Uint32Type:
			tva.SetUint32(uint32(rv.Uint()))
		case gnolang.Uint64Type:
			tva.SetUint64(rv.Uint())
		case gnolang.Float32Type:
			tva.SetFloat32(float32(rv.Float()))
		case gnolang.Float64Type:
			tva.SetFloat64(rv.Float())
		case gnolang.BigintType, gnolang.UntypedBigintType:
			panic("unsupported Gno type: (Untyped)BigIntType")
		case gnolang.BigdecType, gnolang.UntypedBigdecType:
			panic("unsupported Gno type: (Untyped)BigDecType")
		default:
			panic("unknown Gno type: " + ct.Kind().String())
		}

		return

	case *gnolang.PointerType:
		var val gnolang.TypedValue
		val.T = ct.Elt

		tva.newWith(&val).unmarshalValue(rv.Elem(), visited)
		tva.V = gnolang.PointerValue{TV: &val} // heap tva.Allocator

	case *gnolang.ArrayType:
		rvl := rv.Len()
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			av := tva.Allocator.NewDataArray(rvl)
			data := av.Data
			reflect.Copy(reflect.ValueOf(data), rv)
			tva.V = av
		} else {
			av := tva.Allocator.NewListArray(rvl)
			list := av.List
			for i := 0; i < rvl; i++ {
				list[i].T = ct.Elt
				tva.newWith(&list[i]).unmarshalValue(rv.Index(i), visited)
			}
			tva.V = av
		}

	case *gnolang.SliceType:
		rvl := rv.Len()
		rvc := rv.Cap()
		list := make([]gnolang.TypedValue, rvl, rvc)
		for i := 0; i < rvl; i++ {
			list[i].T = ct.Elt
			tva.newWith(&list[i]).unmarshalValue(rv.Index(i), visited)
		}
		tva.V = tva.Allocator.NewSliceFromList(list)

	case *gnolang.StructType:
		fs := tva.Allocator.NewStructFields(len(ct.Fields))
		for i, field := range ct.Fields {
			name := string(field.Name)
			fs[i].T = field.Type

			if !isUpper(name) {
				continue
			}

			fv := rv.FieldByName(string(field.Name))
			tva.newWith(&fs[i]).unmarshalValue(fv, visited)
		}
		tva.V = tva.Allocator.NewStruct(fs)

	case *gnolang.MapType, *gnolang.InterfaceType, *gnolang.PackageType,
		*gnolang.FuncType, *gnolang.DeclaredType, *gnolang.TypeType:
		panic(fmt.Errorf("%w: %s", ErrUnsuportedType, ct.TypeID()))

	default:
		panic(fmt.Sprintf("unexpected type %v with base %v", tva.T, gnolang.BaseOf(tva.T)))
	}

}

func gno2GoType(t gnolang.Type, visited map[uintptr]struct{}) reflect.Type {
	if addr, ok := isPointer(reflect.ValueOf(t)); ok {
		if _, ok := visited[addr]; ok {
			panic(ErrRecursivePointer)
		}
		visited[addr] = struct{}{}
		defer func() { delete(visited, addr) }()
	}

	if t == gnolang.Float32Type {
		return reflect.TypeOf(float32(0.0))
	} else if t == gnolang.Float64Type {
		return reflect.TypeOf(float64(0.0))
	}

	switch ct := gnolang.BaseOf(t).(type) {
	case gnolang.PrimitiveType:
		switch ct {
		case gnolang.BoolType, gnolang.UntypedBoolType:
			return reflect.TypeOf(false)
		case gnolang.StringType, gnolang.UntypedStringType:
			return reflect.TypeOf("")
		case gnolang.IntType:
			return reflect.TypeOf(int(0))
		case gnolang.Int8Type:
			return reflect.TypeOf(int8(0))
		case gnolang.Int16Type:
			return reflect.TypeOf(int16(0))
		case gnolang.Int32Type, gnolang.UntypedRuneType:
			return reflect.TypeOf(int32(0))
		case gnolang.Int64Type:
			return reflect.TypeOf(int64(0))
		case gnolang.UintType:
			return reflect.TypeOf(uint(0))
		case gnolang.Uint8Type:
			return reflect.TypeOf(uint8(0))
		case gnolang.Uint16Type:
			return reflect.TypeOf(uint16(0))
		case gnolang.Uint32Type:
			return reflect.TypeOf(uint32(0))
		case gnolang.Uint64Type:
			return reflect.TypeOf(uint64(0))
		case gnolang.Float32Type:
			return reflect.TypeOf(float32(0))
		case gnolang.Float64Type:
			return reflect.TypeOf(float64(0))
		case gnolang.BigintType, gnolang.UntypedBigintType:
			panic("unsupported Gno type: (Untyped)BigIntType")
		case gnolang.BigdecType, gnolang.UntypedBigdecType:
			panic("unsupported Gno type: (Untyped)BigDecType")
		default:
			panic("unknown Gno type: " + ct.Kind().String())
		}

	case *gnolang.PointerType:
		et := gno2GoType(ct.Elem(), visited)
		return reflect.PointerTo(et)

	case *gnolang.ArrayType:
		ne := ct.Len
		et := gno2GoType(ct.Elem(), visited)
		return reflect.ArrayOf(ne, et)

	case *gnolang.SliceType:
		et := gno2GoType(ct.Elem(), visited)
		return reflect.SliceOf(et)

	case *gnolang.StructType:
		gfs := make([]reflect.StructField, 0, len(ct.Fields))
		for _, field := range ct.Fields {
			gft := gno2GoType(field.Type, visited) // return a reflect.Type

			fn := string(field.Name)
			pkgPath := ""
			if !isUpper(fn) {
				continue
			}

			gf := reflect.StructField{
				Name:      fn,
				PkgPath:   pkgPath,
				Type:      gft,
				Tag:       reflect.StructTag(field.Tag),
				Anonymous: field.Name == "",
				// Offset: dontcare
				// Index: dontcare
			}

			gfs = append(gfs, gf)
		}
		return reflect.StructOf(gfs)

	case *gnolang.MapType:
		kt := gno2GoType(ct.Key, visited)
		vt := gno2GoType(ct.Value, visited)
		return reflect.MapOf(kt, vt)

	case *gnolang.NativeType: // XXX(FIXME): remove me (?)
		return ct.Type

	case *gnolang.PackageType, *gnolang.InterfaceType, *gnolang.DeclaredType,
		*gnolang.TypeType, *gnolang.FuncType:
		panic(fmt.Errorf("%w: %s", ErrUnsuportedType, ct.TypeID()))

	default:
		panic(fmt.Sprintf("unexpected type %v with base %v", t, gnolang.BaseOf(t)))
	}
}

// rv must be addressable, or zero (invalid) (say if tv is referred to from a
// gno.PointerValue). In the latter case, an addressable one will be
// constructed and returned, otherwise returns rv.  if tv is undefined, rv must
// be valid.
func gno2GoValue(tva *gnolang.TypedValue, rv reflect.Value) (ret reflect.Value) {
	var rt reflect.Type
	bt := gnolang.BaseOf(tva.T)

	// XXX: rework this part
	if !rv.IsValid() {
		rt = gno2GoType(bt, map[uintptr]struct{}{})
		rv = reflect.New(rt).Elem()
		ret = rv
	} else if rv.Kind() == reflect.Interface {
		rt = gno2GoType(bt, map[uintptr]struct{}{})
		rv1 := rv
		rv2 := reflect.New(rt).Elem()
		rv = rv2
		defer func() {
			rv1.Set(rv2)
			ret = rv
		}()
	} else {
		ret = rv
		rt = rv.Type()
	}
	// XXX: ---

	switch ct := bt.(type) {
	case gnolang.PrimitiveType:
		switch ct {
		case gnolang.BoolType, gnolang.UntypedBoolType:
			rv.SetBool(tva.GetBool())
		case gnolang.StringType, gnolang.UntypedStringType:
			rv.SetString(tva.GetString())
		case gnolang.IntType:
			rv.SetInt(int64(tva.GetInt()))
		case gnolang.Int8Type:
			rv.SetInt(int64(tva.GetInt8()))
		case gnolang.Int16Type:
			rv.SetInt(int64(tva.GetInt16()))
		case gnolang.Int32Type, gnolang.UntypedRuneType:
			rv.SetInt(int64(tva.GetInt32()))
		case gnolang.Int64Type:
			rv.SetInt(tva.GetInt64())
		case gnolang.UintType:
			rv.SetUint(uint64(tva.GetUint()))
		case gnolang.Uint8Type:
			rv.SetUint(uint64(tva.GetUint8()))
		case gnolang.Uint16Type:
			rv.SetUint(uint64(tva.GetUint16()))
		case gnolang.Uint32Type:
			rv.SetUint(uint64(tva.GetUint32()))
		case gnolang.Uint64Type:
			rv.SetUint(tva.GetUint64())
		case gnolang.Float32Type:
			rv.SetFloat(float64(tva.GetFloat32()))
		case gnolang.Float64Type:
			rv.SetFloat(tva.GetFloat64())
		default:
			panic(fmt.Sprintf("unexpected primitive type %s", tva.T.String()))
		}

	case *gnolang.PointerType:
		// This doesn't take into account pointer relativity, or even
		// identical pointers -- every non-nil gno pointer type results in a
		// new addressable value in go.
		if tva.V == nil {
			// do nothing
		} else {
			rve := reflect.New(rv.Type().Elem()).Elem()
			rv2 := gno2GoValue(tva.V.(gnolang.PointerValue).TV, rve)
			rv.Set(rv2.Addr())
		}

	case *gnolang.ArrayType:
		// General case.
		av := tva.V.(*gnolang.ArrayValue)
		if av.Data == nil {
			for i := 0; i < ct.Len; i++ {
				etv := &av.List[i]
				if etv.IsUndefined() {
					continue
				}
				gno2GoValue(etv, rv.Index(i))
			}
		} else {
			for i := 0; i < ct.Len; i++ {
				val := av.Data[i]
				erv := rv.Index(i)
				erv.SetUint(uint64(val))
			}
		}

	case *gnolang.SliceType:
		st := rt
		// If uninitialized slice, return zero value.
		if tva.V == nil {
			return
		}
		// General case.
		sv := tva.V.(*gnolang.SliceValue)
		svo := sv.Offset
		svl := sv.Length
		svc := sv.Maxcap
		if sv.GetBase(nil).Data == nil {
			rv.Set(reflect.MakeSlice(st, svl, svc))
			for i := 0; i < svl; i++ {
				etv := &(sv.GetBase(nil).List[svo+i])
				if etv.IsUndefined() {
					continue
				}
				gno2GoValue(etv, rv.Index(i))
			}
		} else {
			data := make([]byte, svl, svc)
			copy(data[:svc], sv.GetBase(nil).Data[svo:svo+svc])
			rv.Set(reflect.ValueOf(data))
		}

	case *gnolang.StructType:
		// If uninitialized struct, return zero value.
		if tva.V == nil {
			return
		}
		// General case.
		sv := tva.V.(*gnolang.StructValue)
		for i := range ct.Fields {
			ftv := &(sv.Fields[i])
			if ftv.IsUndefined() {
				continue
			}

			if !isUpper(string(ct.Fields[i].Name)) {
				continue
			}

			gno2GoValue(ftv, rv.Field(i))
		}

	case *gnolang.MapType:
		// If uninitialized map, return zero value.
		if tva.V == nil {
			return
		}

		// General case.
		mv := tva.V.(*gnolang.MapValue)
		mt := rt
		rv.Set(reflect.MakeMapWithSize(mt, mv.List.Size))
		head := mv.List.Head
		vrt := mt.Elem()
		for head != nil {
			ktv, vtv := &head.Key, &head.Value
			krv := gno2GoValue(ktv, reflect.Value{})
			if vtv.IsUndefined() {
				vrv := reflect.New(vrt).Elem()
				rv.SetMapIndex(krv, vrv)
			} else {
				vrv := gno2GoValue(vtv, reflect.Value{})
				rv.SetMapIndex(krv, vrv)
			}
			head = head.Next
		}

	case *gnolang.NativeType:
		// If uninitialized native type, leave rv uninitialized.
		if tva.V == nil {
			return
		}

		// General case.
		rv.Set(tva.V.(*gnolang.NativeValue).Value)

	case *gnolang.FuncType, *gnolang.DeclaredType:
		// TODO: if tv.V.(*NativeValue), just return.
		// TODO: otherwise, set rv to wrapper.
		panic(fmt.Errorf("%w: %s", ErrUnsuportedType, ct.TypeID()))

	default:
		panic(fmt.Sprintf("unexpected type %s", tva.T.String()))
	}

	return
}

// true if the first rune is uppercase.
func isUpper(s string) bool {
	var first rune
	for _, c := range s {
		first = c
		break
	}
	return unicode.IsUpper(first)
}

func isPointer(rv reflect.Value) (uintptr, bool) {
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map,
		reflect.Pointer, reflect.Slice, reflect.UnsafePointer:
		return uintptr(rv.UnsafePointer()), true
	default:
		return 0, false
	}
}
