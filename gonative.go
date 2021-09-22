package gno

import (
	"fmt"
	"math"
	"reflect"
)

// NOTE
//
// GoNative, *nativeType, and *nativeValue are experimental and subject to
// change.
//
// Go 1.15 reflect has a bug in creating new types with methods -- namely, it
// cannot, and so you cannot create types through reflection that obey any
// interface but the empty interface.

// NOTE: Go spec: Type values are comparable, such as with the == operator, so
// they can be used as map keys. Two Type values are equal if they represent
// identical types.
// NOTE: this is used for eager go2GnoType2(), not lazy go2Gnotype().
var go2GnoCache = map[reflect.Type]Type{}

//----------------------------------------
// Go to Gno conversion

// See go2GnoValue(); this is lazy.
func go2GnoType(rt reflect.Type) Type {
	if rt.PkgPath() != "" {
		return &nativeType{Type: rt}
	}
	return go2GnoBaseType(rt)
}

// like go2GnoType() but ignores name declaration.
// for native type unary/binary expression conversion.
// also used for untyped gno -> native conversion intermediary.
// XXX support unary conversions as we did for binary.
func go2GnoBaseType(rt reflect.Type) Type {
	switch rk := rt.Kind(); rk {
	case reflect.Bool:
		return BoolType
	case reflect.String:
		return StringType
	case reflect.Int:
		return IntType
	case reflect.Int8:
		return Int8Type
	case reflect.Int16:
		return Int16Type
	case reflect.Int32:
		return Int32Type
	case reflect.Int64:
		return Int64Type
	case reflect.Uint:
		return UintType
	case reflect.Uint8:
		return Uint8Type
	case reflect.Uint16:
		return Uint16Type
	case reflect.Uint32:
		return Uint32Type
	case reflect.Uint64:
		return Uint64Type
	case reflect.Float32:
		return Float32Type
	case reflect.Float64:
		return Float64Type
	case reflect.Array:
		return &nativeType{Type: rt}
	case reflect.Slice:
		return &nativeType{Type: rt}
	case reflect.Chan:
		return &nativeType{Type: rt}
	case reflect.Func:
		return &nativeType{Type: rt}
	case reflect.Interface:
		return &nativeType{Type: rt}
	case reflect.Map:
		return &nativeType{Type: rt}
	case reflect.Ptr:
		return &nativeType{Type: rt}
	case reflect.Struct:
		return &nativeType{Type: rt}
	case reflect.UnsafePointer:
		panic("not yet implemented")
	default:
		panic(fmt.Sprintf(
			"unexpected type %v", rt))
	}
}

// See go2GnoValue2(). Like go2GnoType() but also converts any top-level
// complex types (or pointers to them).  The result gets memoized in
// *nativeType.GnoType() for type inference in the preprocessor, as well as in
// the go2GnoCache lookup map to support recursive translations.
func go2GnoType2(rt reflect.Type) (t Type) {
	if gnot, ok := go2GnoCache[rt]; ok {
		return gnot
	}
	defer func() {
		if r := recover(); r != nil {
			panic(r) // do not run the below logic upon panic.
		}
		// regardless of rt kind, if rt.PkgPath() is set,
		// wrap t with declared type.
		pkgPath := rt.PkgPath()
		if pkgPath != "" {
			mtvs := []TypedValue(nil)
			if t.Kind() == InterfaceKind {
				// methods already set on t.Methods.
				// *DT.Methods not used in Go for interfaces.
			} else {
				prt := rt
				if rt.Kind() != reflect.Ptr {
					// NOTE: go reflect requires ptr kind for
					// methods with ptr receivers, whereas gno
					// methods are all declared on the
					// *DeclaredType.
					prt = reflect.PtrTo(rt)
				}
				nm := prt.NumMethod()
				mtvs = make([]TypedValue, nm)
				for i := 0; i < nm; i++ {
					mthd := prt.Method(i)
					ft := go2GnoFuncType(mthd.Type)
					fv := &FuncValue{
						Type:       ft,
						IsMethod:   true,
						SourceLoc:  Location{},
						Source:     nil,
						Name:       Name(mthd.Name),
						Body:       nil, // XXX
						Closure:    nil,
						PkgPath:    pkgPath,
						nativeBody: nil,
						pkg:        nil, // XXX
					}
					mtvs[i] = TypedValue{T: ft, V: fv}
				}
			}
			dt := &DeclaredType{
				PkgPath: pkgPath,
				Name:    Name(rt.Name()),
				Base:    t,
				Methods: mtvs,
			}
			dt.Seal()
			t = dt
		}
		// memoize t to cache.
		if debug {
			if gnot, ok := go2GnoCache[rt]; ok {
				if gnot.TypeID() != baseOf(t).TypeID() {
					panic("should not happen")
				}
			}
		}
		go2GnoCache[rt] = t // may overwrite
	}()
	switch rk := rt.Kind(); rk {
	case reflect.Bool:
		return BoolType
	case reflect.String:
		return StringType
	case reflect.Int:
		return IntType
	case reflect.Int8:
		return Int8Type
	case reflect.Int16:
		return Int16Type
	case reflect.Int32:
		return Int32Type
	case reflect.Int64:
		return Int64Type
	case reflect.Uint:
		return UintType
	case reflect.Uint8:
		return Uint8Type
	case reflect.Uint16:
		return Uint16Type
	case reflect.Uint32:
		return Uint32Type
	case reflect.Uint64:
		return Uint64Type
	case reflect.Float32:
		return Float32Type
	case reflect.Float64:
		return Float64Type
	case reflect.Array:
		// predefine gno type
		at := &ArrayType{}
		go2GnoCache[rt] = at
		// define gno type
		at.Len = rt.Len()
		at.Elt = go2GnoType(rt.Elem())
		at.Vrd = false
		return at
	case reflect.Slice:
		return &SliceType{
			Elt: go2GnoType(rt.Elem()),
			Vrd: false,
		}
	case reflect.Chan:
		// predefine gno type
		ct := &ChanType{}
		go2GnoCache[rt] = ct
		// define gno type
		chdir := toChanDir(rt.ChanDir())
		ct.Dir = chdir
		ct.Elt = go2GnoType(rt.Elem())
		return ct
	case reflect.Func:
		return go2GnoFuncType(rt)
	case reflect.Interface:
		// predefine gno type
		it := &InterfaceType{}
		go2GnoCache[rt] = it
		// define gno type
		nm := rt.NumMethod()
		fs := make([]FieldType, nm)
		for i := 0; i < nm; i++ {
			mthd := rt.Method(i)
			fs[i] = FieldType{
				Name: Name(mthd.Name),
				Type: go2GnoType2(mthd.Type),
			}
		}
		it.PkgPath = rt.PkgPath()
		it.Methods = fs
		return it
	case reflect.Map:
		// predefine gno type
		mt := &MapType{}
		go2GnoCache[rt] = mt
		// define gno type
		mt.Key = go2GnoType(rt.Key())
		mt.Value = go2GnoType(rt.Elem())
		return mt
	case reflect.Ptr:
		return &PointerType{
			// this is the only recursive call to go2GnoType2().
			Elt: go2GnoType2(rt.Elem()),
		}
	case reflect.Struct:
		// predefine gno type
		st := &StructType{}
		go2GnoCache[rt] = st
		// define gno type
		nf := rt.NumField()
		fs := make([]FieldType, nf)
		for i := 0; i < nf; i++ {
			sf := rt.Field(i)
			fs[i] = FieldType{
				Name: Name(sf.Name),
				Type: go2GnoType(sf.Type),
			}
		}
		st.PkgPath = rt.PkgPath()
		st.Fields = fs
		return st
	case reflect.UnsafePointer:
		panic("not yet implemented")
	default:
		panic("not yet implemented")
	}
}

// Default run-time representation of go-native values.  It is "lazy" in the
// sense that unnamed complex types like arrays and slices aren't translated
// to Gno canonical types except as *nativeType/*nativeValues, primarily for
// speed.  To force translation to Gno canonical types for unnamed complex
// types, call go2GnoValue2(), which is used by the implementation of
// ConvertTo().
// Unlike go2GnoValue2(), rv may be invalid.
func go2GnoValue(rv reflect.Value) (tv TypedValue) {
	if !rv.IsValid() {
		return
	}
	if rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return TypedValue{}
		} else {
			rv = rv.Elem()
		}
	}
	if rv.Type().PkgPath() != "" {
		rt := rv.Type()
		tv.T = &nativeType{Type: rt}
		tv.V = &nativeValue{Value: rv}
		return
	}
	tv.T = go2GnoType(rv.Type())
	switch rk := rv.Kind(); rk {
	case reflect.Bool:
		tv.SetBool(rv.Bool())
	case reflect.String:
		tv.V = StringValue(rv.String())
	case reflect.Int:
		tv.SetInt(int(rv.Int()))
	case reflect.Int8:
		tv.SetInt8(int8(rv.Int()))
	case reflect.Int16:
		tv.SetInt16(int16(rv.Int()))
	case reflect.Int32:
		tv.SetInt32(int32(rv.Int()))
	case reflect.Int64:
		tv.SetInt64(int64(rv.Int()))
	case reflect.Uint:
		tv.SetUint(uint(rv.Uint()))
	case reflect.Uint8:
		tv.SetUint8(uint8(rv.Uint()))
	case reflect.Uint16:
		tv.SetUint16(uint16(rv.Uint()))
	case reflect.Uint32:
		tv.SetUint32(uint32(rv.Uint()))
	case reflect.Uint64:
		tv.SetUint64(uint64(rv.Uint()))
	case reflect.Float32:
		fl := float32(rv.Float())
		u32 := math.Float32bits(fl)
		if sv, ok := tv.V.(*StructValue); ok { // update
			if debug {
				if len(sv.Fields) != 1 {
					panic("should not happen")
				}
				if sv.Fields[0].T != Uint32Type {
					panic("invalid Float32Type, expected Uint32 field")
				}
			}
			sv.Fields[0].SetUint32(u32)
		} else { // create
			ftv := TypedValue{T: Uint32Type}
			ftv.SetUint32(u32)
			tv.V = &StructValue{
				Fields: []TypedValue{ftv},
			}
		}
	case reflect.Float64:
		fl := rv.Float()
		u64 := math.Float64bits(fl)
		if sv, ok := tv.V.(*StructValue); ok { // update
			if debug {
				if len(sv.Fields) != 1 {
					panic("should not happen")
				}
				if sv.Fields[0].T != Uint64Type {
					panic("invalid Float64Type, expected Uint64 field")
				}
			}
			sv.Fields[0].SetUint64(u64)
		} else { // create
			ftv := TypedValue{T: Uint64Type}
			ftv.SetUint64(u64)
			tv.V = &StructValue{
				Fields: []TypedValue{ftv},
			}
		}
	case reflect.Array:
		tv.V = &nativeValue{rv}
	case reflect.Slice:
		tv.V = &nativeValue{rv}
	case reflect.Chan:
		tv.V = &nativeValue{rv}
	case reflect.Func:
		tv.V = &nativeValue{rv}
	case reflect.Interface:
		panic("should not happen")
	case reflect.Map:
		tv.V = &nativeValue{rv}
	case reflect.Ptr:
		tv.V = &nativeValue{rv}
	case reflect.Struct:
		tv.V = &nativeValue{rv}
	case reflect.UnsafePointer:
		panic("not yet implemented")
	default:
		panic("not yet implemented")
	}
	return
}

// Given rv which may have been updated by a go-native
// function, and the corresponding (original) input value tv,
// scan for changes and update tv recursively as needed.
// An additional side effect is that uninitialized input values
// become initialized.  Due to limitations of Go 1.15
// reflection, any child Gno declared types cannot change
// types.
func go2GnoValueUpdate(rlm *Realm, lvl int, tv *TypedValue, rv reflect.Value) {
	// Special case if nil:
	if tv.IsUndefined() {
		return // do nothing
	}
	// Special case if native type:
	if _, ok := tv.T.(*nativeType); ok {
		return // do nothing
	}
	// De-interface if interface.
	if rv.Kind() == reflect.Interface {
		rv = rv.Elem()
	}
	// General case:
	switch tvk := tv.T.Kind(); tvk {
	case BoolKind:
		if lvl != 0 {
			tv.SetBool(rv.Bool())
		}
	case StringKind:
		if lvl != 0 {
			tv.V = StringValue(rv.String())
		}
	case IntKind:
		if lvl != 0 {
			tv.SetInt(int(rv.Int()))
		}
	case Int8Kind:
		if lvl != 0 {
			tv.SetInt8(int8(rv.Int()))
		}
	case Int16Kind:
		if lvl != 0 {
			tv.SetInt16(int16(rv.Int()))
		}
	case Int32Kind:
		if lvl != 0 {
			tv.SetInt32(int32(rv.Int()))
		}
	case Int64Kind:
		if lvl != 0 {
			tv.SetInt64(int64(rv.Int()))
		}
	case UintKind:
		if lvl != 0 {
			tv.SetUint(uint(rv.Uint()))
		}
	case Uint8Kind:
		if lvl != 0 {
			tv.SetUint8(uint8(rv.Uint()))
		}
	case Uint16Kind:
		if lvl != 0 {
			tv.SetUint16(uint16(rv.Uint()))
		}
	case Uint32Kind:
		if lvl != 0 {
			tv.SetUint32(uint32(rv.Uint()))
		}
	case Uint64Kind:
		if lvl != 0 {
			tv.SetUint64(uint64(rv.Uint()))
		}
	case BigintKind:
		panic("not yet implemented")
	case ArrayKind:
		av := tv.V.(*ArrayValue)
		rvl := rv.Len()
		if debug {
			if rvl != baseOf(tv.T).(*ArrayType).Len {
				panic("go-native update error: array length mismmatch")
			}
		}
		if av.Data == nil {
			at := baseOf(tv.T).(*ArrayType)
			et := at.Elt
			for i := 0; i < rvl; i++ {
				erv := rv.Index(i)
				etv := &av.List[i]
				// XXX use Assign and Realm?
				if etv.T == nil && et.Kind() != InterfaceKind {
					etv.T = et
				}
				if etv.V == nil {
					etv.V = defaultValue(et)
				}
				go2GnoValueUpdate(rlm, lvl+1, etv, erv)
			}
		} else {
			for i := 0; i < rvl; i++ {
				erv := rv.Index(i)
				av.Data[i] = uint8(erv.Uint())
			}
		}
	case SliceKind:
		rvl := rv.Len()
		if rvl == 0 {
			if debug {
				if tv.V != nil && tv.V.(*SliceValue).GetLength() != 0 {
					panic("should not happen")
				}
			}
			return // do nothing
		}
		sv := tv.V.(*SliceValue)
		svo := sv.Offset
		svl := sv.Length
		if debug {
			if rvl != svl {
				panic("go-native update error: slice length mismmatch")
			}
		}
		if sv.GetBase(nil).Data == nil {
			st := baseOf(tv.T).(*SliceType)
			et := st.Elt
			for i := 0; i < rvl; i++ {
				erv := rv.Index(i)
				etv := &sv.GetBase(nil).List[svo+i]
				// XXX use Assign and Realm?
				if etv.T == nil && et.Kind() != InterfaceKind {
					etv.T = et
				}
				if etv.V == nil {
					etv.V = defaultValue(et)
				}
				go2GnoValueUpdate(rlm, lvl+1, etv, erv)
			}
		} else {
			for i := 0; i < rvl; i++ {
				erv := rv.Index(i)
				sv.GetBase(nil).Data[svo+i] = uint8(erv.Uint())
			}
		}
	case PointerKind:
		if tv.V == nil {
			return // do nothing
		}
		pv := tv.V.(PointerValue)
		etv := pv.TV
		erv := rv.Elem()
		go2GnoValueUpdate(rlm, lvl+1, etv, erv)
	case StructKind:
		st := baseOf(tv.T).(*StructType)
		sv := tv.V.(*StructValue)
		switch st.PkgPath {
		case float32PkgPath: // Special case if Float32.
			fl := float32(rv.Float())
			u32 := math.Float32bits(fl)
			if debug {
				if len(sv.Fields) != 1 {
					panic("should not happen")
				}
				if sv.Fields[0].T != Uint32Type {
					panic("invalid Float32Type, expected Uint32 field")
				}
			}
			sv.Fields[0].SetUint32(u32)
		case float64PkgPath: // Special case if Float64.
			fl := rv.Float()
			u64 := math.Float64bits(fl)
			if debug {
				if len(sv.Fields) != 1 {
					panic("should not happen")
				}
				if sv.Fields[0].T != Uint64Type {
					panic("invalid Float64Type, expected Uint64 field")
				}
			}
			sv.Fields[0].SetUint64(u64)
		default: // General case.
			for i := range st.Fields {
				ft := st.Fields[i].Type
				ftv := &sv.Fields[i]
				// XXX use Assign and Realm?
				if ftv.T == nil && ft.Kind() != InterfaceKind {
					ftv.T = ft
				}
				if ftv.V == nil {
					ftv.V = defaultValue(ft)
				}
				frv := rv.Field(i)
				go2GnoValueUpdate(rlm, lvl+1, ftv, frv)
			}
		}
	case PackageKind:
		panic("not yet implemented")
	case InterfaceKind:
		if debug {
			if !tv.IsNilInterface() {
				panic("should not happen")
			}
		}
	case ChanKind:
		panic("not yet implemented")
	case FuncKind:
		panic("not yet implemented")
	case MapKind:
		rvl := rv.Len()
		// If uninitialized map, return zero value.
		if tv.V == nil {
			if rvl != 0 {
				panic("not yet implemented")
			}
			return
		}
		// General case.
		mv := tv.V.(*MapValue)
		mvl := mv.List.Size
		// Copy map to new map for destructive iteration of items.
		rt := rv.Type()
		rv2 := reflect.MakeMapWithSize(rt, mvl)
		rvi := rv.MapRange()
		for rvi.Next() {
			k, v := rvi.Key(), rvi.Value()
			rv2.SetMapIndex(k, v)
		}
		// Iterate over mv (gno map) and update,
		// and also remove encountered items from rv2.
		head := mv.List.Head
		for head != nil {
			ktv, vtv := &head.Key, &head.Value
			// Update in place.
			krv := gno2GoValue(ktv, reflect.Value{})
			vrv := rv.MapIndex(krv)
			if vrv.IsZero() {
				// XXX remove key from mv
				panic("not yet implemented")
			} else {
				go2GnoValueUpdate(rlm, lvl+1, vtv, vrv)
			}
			// Delete from rv2
			rv2.SetMapIndex(krv, reflect.Value{})
			// Continue
			head = head.Next
		}
		// Add remaining items from rv2 to map.
		rv2i := rv2.MapRange()
		for rv2i.Next() {
			k, v := rv2i.Key(), rv2i.Value()
			ktv := go2GnoValue(k)
			vtv := go2GnoValue(v)
			ptr := mv.GetPointerForKey(nil, &ktv)
			if debug {
				if !ptr.TV.IsUndefined() {
					panic("should not happen")
				}
			}
			ptr.Assign2(nil, rlm, vtv, false) // document false
		}
	case TypeKind:
		panic("not yet implemented")
	default:
		panic("should not happen: unexpected gno kind")
	}
	return
}

// This function is like go2GnoValue() but less lazy (but still
// not recursive/eager). It is for converting Go types to Gno
// types upon an explicit conversion (via ConvertTo).  Panics on
// unexported/private fields.  Due to limitations of Go1.15, the
// namedness is dropped rather than converted.  This lets users
// convert go-native types to named or unnamed Gno types (sans
// private fields) via conversion.  The conversion is not
// recursive, and the extra conversion works on the top-level
// complex type/value, or a pointer to that type/value.  Some
// types that cannot be converted remain native.
// Unlike go2GnoValue(), rv must be valid.
func go2GnoValue2(rv reflect.Value) (tv TypedValue) {
	if debug {
		if !rv.IsValid() {
			panic("go2GnoValue2() requires valid rv")
		}
	}
	tv.T = go2GnoType2(rv.Type())
	switch rk := rv.Kind(); rk {
	case reflect.Bool:
		tv.SetBool(rv.Bool())
	case reflect.String:
		tv.V = StringValue(rv.String())
	case reflect.Int:
		tv.SetInt(int(rv.Int()))
	case reflect.Int8:
		tv.SetInt8(int8(rv.Int()))
	case reflect.Int16:
		tv.SetInt16(int16(rv.Int()))
	case reflect.Int32:
		tv.SetInt32(int32(rv.Int()))
	case reflect.Int64:
		tv.SetInt64(int64(rv.Int()))
	case reflect.Uint:
		tv.SetUint(uint(rv.Uint()))
	case reflect.Uint8:
		tv.SetUint8(uint8(rv.Uint()))
	case reflect.Uint16:
		tv.SetUint16(uint16(rv.Uint()))
	case reflect.Uint32:
		tv.SetUint32(uint32(rv.Uint()))
	case reflect.Uint64:
		tv.SetUint64(uint64(rv.Uint()))
	case reflect.Float32:
		fl := float32(rv.Float())
		u32 := math.Float32bits(fl)
		tv.SetUint32(u32)
	case reflect.Float64:
		fl := rv.Float()
		u64 := math.Float64bits(fl)
		tv.SetUint64(u64)
	case reflect.Array:
		rvl := rv.Len()
		list := make([]TypedValue, rvl)
		for i := 0; i < rvl; i++ {
			list[i] = go2GnoValue(rv.Index(i))
		}
		tv.V = &ArrayValue{
			List: list,
		}
	case reflect.Slice:
		rvl := rv.Len()
		rvc := rv.Cap()
		list := make([]TypedValue, rvl, rvc)
		for i := 0; i < rvl; i++ {
			list[i] = go2GnoValue(rv.Index(i))
		}
		tv.V = newSliceFromList(list)
	case reflect.Chan:
		panic("not yet implemented")
	case reflect.Func:
		// NOTE: the type may be a full gno type, either a
		// *FuncType or *DeclaredType.  The value may still be a
		// *nativeValue though, and the function can be called
		// regardless.
		tv.V = &nativeValue{
			Value: rv,
		}
	case reflect.Interface:
		panic("not yet implemented")
	case reflect.Map:
		panic("not yet implemented")
	case reflect.Ptr:
		tv.T = &PointerType{Elt: go2GnoType2(rv.Type().Elem())}
		val := go2GnoValue2(rv.Elem())
		tv.V = PointerValue{TV: &val} // heap alloc
	case reflect.Struct:
		panic("not yet implemented")
	case reflect.UnsafePointer:
		panic("not yet implemented")
	default:
		panic("not yet implemented")
	}
	return
}

// Converts native go function type to gno *FuncType,
// for the preprocessor to infer types of arguments.
func go2GnoFuncType(rt reflect.Type) *FuncType {
	// prewdefine func type
	ft := &FuncType{}
	go2GnoCache[rt] = ft
	// define func type
	hasVargs := rt.IsVariadic()
	ins := make([]FieldType, rt.NumIn())
	for i := 0; i < len(ins); i++ {
		it := go2GnoType(rt.In(i))
		if hasVargs && i == len(ins)-1 {
			it = &SliceType{
				Elt: it.Elem(),
				Vrd: true,
			}
		}
		ins[i] = FieldType{
			Name: "", // XXX dontcare?
			Type: it,
		}
	}
	outs := make([]FieldType, rt.NumOut())
	for i := 0; i < len(outs); i++ {
		ot := go2GnoType(rt.Out(i))
		outs[i] = FieldType{
			Name: "", // XXX dontcare?
			Type: ot,
		}
	}
	ft.Params = ins
	ft.Results = outs
	return ft
}

//----------------------------------------
// Gno to Go conversion

// NOTE: Recursive types are not supported, as named types are not
// supported.  See https://github.com/golang/go/issues/20013 and
// https://github.com/golang/go/issues/39717.
func gno2GoType(t Type) reflect.Type {
	// special case if t == Float32Type or Float64Type
	if t == Float32Type {
		return reflect.TypeOf(float32(0.0))
	} else if t == Float64Type {
		return reflect.TypeOf(float64(0.0))
	}
	switch ct := baseOf(t).(type) {
	case PrimitiveType:
		switch ct {
		case BoolType, UntypedBoolType:
			return reflect.TypeOf(false)
		case StringType, UntypedStringType:
			return reflect.TypeOf("")
		case IntType:
			return reflect.TypeOf(int(0))
		case Int8Type:
			return reflect.TypeOf(int8(0))
		case Int16Type:
			return reflect.TypeOf(int16(0))
		case Int32Type, UntypedRuneType:
			return reflect.TypeOf(int32(0))
		case Int64Type:
			return reflect.TypeOf(int64(0))
		case UintType:
			return reflect.TypeOf(uint(0))
		case Uint8Type:
			return reflect.TypeOf(uint8(0))
		case Uint16Type:
			return reflect.TypeOf(uint16(0))
		case Uint32Type:
			return reflect.TypeOf(uint32(0))
		case Uint64Type:
			return reflect.TypeOf(uint64(0))
		case BigintType, UntypedBigintType:
			panic("not yet implemented")
		default:
			panic("should not happen")
		}
	case *PointerType:
		et := gno2GoType(ct.Elem())
		return reflect.PtrTo(et)
	case *ArrayType:
		ne := ct.Len
		et := gno2GoType(ct.Elem())
		return reflect.ArrayOf(ne, et)
	case *SliceType:
		et := gno2GoType(ct.Elem())
		return reflect.SliceOf(et)
	case *StructType:
		gfs := make([]reflect.StructField, len(ct.Fields))
		for i, field := range ct.Fields {
			gft := gno2GoType(field.Type)
			fn := string(field.Name)
			pkgPath := ""
			if !isUpper(fn) {
				pkgPath = ct.PkgPath
			}
			gfs[i] = reflect.StructField{
				Name:      fn,
				PkgPath:   pkgPath,
				Type:      gft,
				Tag:       reflect.StructTag(field.Tag),
				Anonymous: field.Name == "",
				// Offset: dontcare
				// Index: dontcare
			}
		}
		return reflect.StructOf(gfs)
	case *MapType:
		kt := gno2GoType(ct.Key)
		vt := gno2GoType(ct.Value)
		return reflect.MapOf(kt, vt)
	case *FuncType:
		panic("not yet supported")
	case *InterfaceType:
		if ct.IsEmptyInterface() {
			// XXX move out
			rt := reflect.TypeOf((*interface{})(nil)).Elem()
			return rt
		} else {
			// NOTE: can this be implemented in go1.15? i think not.
			panic("not yet supported")
		}
	case *TypeType:
		panic("should not happen")
	case *DeclaredType:
		// NOTE: Go1.15 has issues with generating types and values using
		// reflect to declare types with methods.  When Go has fixed these
		// issues, we can revisit.  For now, all Gno objects passed to Go
		// lose their names or "namedness", e.g. cannot satisfy anything
		// but empty interfaces, and have no methods.

		// We switch on baseOf(t).
		panic("should not happen")
	case *PackageType:
		panic("should not happen")
	case *nativeType:
		return ct.Type
	default:
		panic(fmt.Sprintf("unexpected type %v with base %v", t, baseOf(t)))
	}
}

// rv must be addressable, or zero (invalid) (say if tv is referred to from a
// gno.PointerValue). In the latter case, an addressable one will be
// constructed and returned, otherwise returns rv.  if tv is undefined, rv must
// be valid.
func gno2GoValue(tv *TypedValue, rv reflect.Value) (ret reflect.Value) {
	if tv.IsUndefined() {
		if debug {
			if !rv.IsValid() {
				panic("unexpected undefined gno value")
			}
		}
		return rv
	}
	bt := baseOf(tv.T)
	if !rv.IsValid() {
		rt := gno2GoType(bt)
		rv = reflect.New(rt).Elem()
		ret = rv
	} else if rv.Kind() == reflect.Interface && rv.IsZero() {
		rt := gno2GoType(bt)
		rv1 := rv
		rv2 := reflect.New(rt).Elem()
		rv = rv2       // swaparoo
		defer func() { // TODO: improve?
			rv1.Set(rv2)
			ret = rv
		}()
	} else {
		ret = rv
	}
	switch ct := bt.(type) {
	case PrimitiveType:
		switch ct {
		case BoolType, UntypedBoolType:
			rv.SetBool(tv.GetBool())
		case StringType, UntypedStringType:
			rv.SetString(string(tv.GetString()))
		case IntType:
			rv.SetInt(int64(tv.GetInt()))
		case Int8Type:
			rv.SetInt(int64(tv.GetInt8()))
		case Int16Type:
			rv.SetInt(int64(tv.GetInt16()))
		case Int32Type, UntypedRuneType:
			rv.SetInt(int64(tv.GetInt32()))
		case Int64Type:
			rv.SetInt(int64(tv.GetInt64()))
		case UintType:
			rv.SetUint(uint64(tv.GetUint()))
		case Uint8Type:
			rv.SetUint(uint64(tv.GetUint8()))
		case Uint16Type:
			rv.SetUint(uint64(tv.GetUint16()))
		case Uint32Type:
			rv.SetUint(uint64(tv.GetUint32()))
		case Uint64Type:
			rv.SetUint(uint64(tv.GetUint64()))
		default:
			panic(fmt.Sprintf(
				"unexpected type %s",
				tv.T.String()))
		}
	case *PointerType:
		// This doesn't take into account pointer relativity, or even
		// identical pointers -- every non-nil gno pointer type results in a
		// new addressable value in go.
		if tv.V == nil {
			// do nothing
		} else {
			rv2 := gno2GoValue(tv.V.(PointerValue).TV, reflect.Value{})
			rv.Set(rv2.Addr())
		}
	case *ArrayType:
		if debug {
			if tv.V == nil {
				// all arguments and recursively fetched arrays
				// should have been initialized if not already so.
				panic("unexpected uninitialized array")
			}
		}
		// General case.
		av := tv.V.(*ArrayValue)
		if av.Data == nil {
			for i := 0; i < ct.Len; i++ {
				etv := &av.List[i]
				if etv.IsUndefined() {
					continue
				}
				gno2GoValue(etv, rv.Index(i))
			}
		} else {
			panic("not yet implemented")
			/*
				el := av.GetLength()
				ec := av.GetCapacity()
				data := make([]byte, el, ec)
				copy(data, av.Data)
				rv = reflect.ValueOf(data)
			*/
		}
	case *SliceType:
		st := gno2GoType(ct)
		// If uninitialized slice, return zero value.
		if tv.V == nil {
			return
		}
		// General case.
		sv := tv.V.(*SliceValue)
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
	case *StructType:
		// If uninitialized struct, return zero value.
		if tv.V == nil {
			return
		}
		// General case.
		sv := tv.V.(*StructValue)
		switch ct.PkgPath {
		case float32PkgPath:
			// Special case if Float32.
			if debug {
				if len(sv.Fields) != 1 {
					panic("should not happen")
				}
				if sv.Fields[0].T != Uint32Type {
					panic("invalid Float32Type, expected Uint32 field")
				}
			}
			u32 := sv.Fields[0].GetUint32()
			fl := math.Float32frombits(u32)
			rv.SetFloat(float64(fl))
		case float64PkgPath:
			// Special case if Float64.
			if debug {
				if len(sv.Fields) != 1 {
					panic("should not happen")
				}
				if sv.Fields[0].T != Uint64Type {
					panic("invalid Float64Type, expected Uint64 field")
				}
			}
			u64 := sv.Fields[0].GetUint64()
			fl := math.Float64frombits(u64)
			rv.SetFloat(fl)
		default:
			// General case.
			for i := range ct.Fields {
				ftv := &(sv.Fields[i])
				if ftv.IsUndefined() {
					continue
				}
				gno2GoValue(ftv, rv.Field(i))
			}
		}
	case *MapType:
		// If uninitialized map, return zero value.
		if tv.V == nil {
			return
		}
		// General case.
		mv := tv.V.(*MapValue)
		mt := gno2GoType(ct)
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
	case *nativeType:
		// If uninitialized native type, leave rv uninitialized.
		if tv.V == nil {
			return
		}
		// General case.
		rv.Set(tv.V.(*nativeValue).Value)
	case *DeclaredType:
		// See corresponding note on gno2GoType().
		panic("should not happen") // we switch on baseOf().
	case *FuncType:
		// TODO: if tv.V.(*nativeValue), just return.
		// TODO: otherwise, set rv to wrapper.
		panic("gno2Go not supported for gno functions yet")
	default:
		panic(fmt.Sprintf(
			"unexpected type %s",
			tv.T.String()))
	}
	return
}

//----------------------------------------
// PackageNode methods

func (pn *PackageNode) DefineGoNativeType(rt reflect.Type) {
	if debug {
		debug.Printf("*PackageNode.DefineGoNativeType(%s)\n", rt.String())
	}
	pkgp := rt.PkgPath()
	if pkgp == "" {
		// DefineGoNativeType can only work with defined exported types.
		// Unexported types should be composed, and primitive types
		// should just use Gno types.
		panic(fmt.Sprintf(
			"reflect.Type %s has no package path",
			rt.String()))
	}
	name := rt.Name()
	if name == "" {
		panic(fmt.Sprintf(
			"reflect.Type %s is not named",
			rt.String()))
	}
	if rt.PkgPath() == "" {
		panic(fmt.Sprintf(
			"reflect.Type %s is not defined/exported",
			rt.String()))
	}
	nt := &nativeType{Type: rt}
	pn.Define(Name(name), asValue(nt))
}

func (pn *PackageNode) DefineGoNativeValue(n Name, nv interface{}) {
	if debug {
		debug.Printf("*PackageNode.DefineGoNativeValue(%s)\n", reflect.ValueOf(nv).String())
	}
	rv := reflect.ValueOf(nv)
	// rv is not settable, so create something that is.
	rt := rv.Type()
	rv2 := reflect.New(rt).Elem()
	rv2.Set(rv)
	pn.Define(n, go2GnoValue(rv2))
}

func (pn *PackageNode) DefineGoNativeFunc(n Name, fn interface{}) {
	if debug {
		debug.Printf("*PackageNode.DefineGoNativeFunc(%s)\n", reflect.ValueOf(fn).String())
	}
	if reflect.TypeOf(fn).Kind() != reflect.Func {
		panic(fmt.Sprintf(
			"DefineGoNativeFunc expects a function, but got %s",
			reflect.TypeOf(fn).String()))
	}
	rv := reflect.ValueOf(fn)
	pn.Define(n, go2GnoValue(rv))
}

//----------------------------------------
// Machine methods

func (m *Machine) doOpStructLitGoNative() {
	// assess performance TODO
	x := m.PopExpr().(*CompositeLitExpr)
	el := len(x.Elts) // may be incomplete
	// peek struct type.
	xt := m.PeekValue(1 + el).V.(TypeValue).Type
	nt := xt.(*nativeType)
	rv := reflect.New(nt.Type).Elem()
	// whether composite lit had field names or not...
	if el == 0 {
		// zero struct with no fields set.
	} else if x.Elts[0].Key == nil {
		// field values are in order.
		ftvs := m.PopValues(el)
		for i := 0; i < el; i++ {
			rf := rv.Field(i)
			gno2GoValue(&ftvs[i], rf)
		}
	} else {
		// field values are by name and may be out of order.
		ftvs := m.PopValues(el)
		for i := 0; i < el; i++ {
			fnx := x.Elts[i].Key.(*NameExpr)
			rf := rv.FieldByName(string(fnx.Name))
			gno2GoValue(&ftvs[i], rf)
		}
	}
	// construct and push value.
	if debug {
		if m.PopValue().V.(TypeValue).Type != nt {
			panic("should not happen")
		}
	} else {
		m.PopValue()
	}
	nv := &nativeValue{
		Value: rv,
	}
	m.PushValue(TypedValue{
		T: nt,
		V: nv,
	})
}

// NOTE: Unlike doOpCall(), doOpCallGoNative() also handles
// conversions, similarly to doOpConvert().
func (m *Machine) doOpCallGoNative() {
	fr := m.LastFrame()
	fv := fr.GoFunc
	ft := fv.Value.Type()
	hasVarg := ft.IsVariadic()
	numParams := ft.NumIn()
	isVarg := fr.IsVarg
	// pop and convert params.
	ptvs := m.PopCopyValues(fr.NumArgs)
	prvs := make([]reflect.Value, 0, len(ptvs))
	for i := 0; i < fr.NumArgs; i++ {
		ptv := &ptvs[i]
		if ptv.IsUndefined() {
			var it reflect.Type
			if hasVarg && numParams-1 <= i && !isVarg {
				it = ft.In(numParams - 1)
				it = it.Elem()
			} else {
				it = ft.In(i)
			}
			erv := reflect.New(it).Elem()
			prvs = append(prvs, gno2GoValue(ptv, erv))
		} else {
			if hasVarg && isVarg && numParams-1 == i {
				panic("not yet supported")
			} else {
				prvs = append(prvs, gno2GoValue(ptv, reflect.Value{}))
			}
		}
	}
	// call and get results.
	rrvs := fv.Value.Call(prvs)
	// convert and push results.
	for _, rvs := range rrvs {
		rtv := go2GnoValue(rvs)
		m.PushValue(rtv)
	}
	// carry writes to params if needed.
	for i := 0; i < fr.NumArgs; i++ {
		ptv := &ptvs[i]
		prv := prvs[i]
		if !ptv.IsUndefined() {
			go2GnoValueUpdate(m.Realm, 0, ptv, prv)
		}
	}
	// cleanup
	m.NumResults = fv.Value.Type().NumOut()
	m.PopFrame()
}

//----------------------------------------
// misc

func toChanDir(dir reflect.ChanDir) ChanDir {
	switch dir {
	case reflect.RecvDir:
		return RECV
	case reflect.SendDir:
		return SEND
	case reflect.BothDir:
		return BOTH
	default:
		panic("should not happn")
	}
}
