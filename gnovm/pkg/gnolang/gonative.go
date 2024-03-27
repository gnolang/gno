package gnolang

import (
	"fmt"
	"reflect"
)

// NOTE
//
// GoNative, *NativeType, and *NativeValue are experimental and subject to
// change.
//
// Go 1.15 reflect has a bug in creating new types with methods -- namely, it
// cannot, and so you cannot create types through reflection that obey any
// interface but the empty interface.

// ----------------------------------------
// Go to Gno conversion

// See go2GnoValue(); this is lazy.
func go2GnoType(rt reflect.Type) Type {
	if rt.PkgPath() != "" {
		return &NativeType{Type: rt}
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
		return &NativeType{Type: rt}
	case reflect.Slice:
		return &NativeType{Type: rt}
	case reflect.Chan:
		return &NativeType{Type: rt}
	case reflect.Func:
		return &NativeType{Type: rt}
	case reflect.Interface:
		return &NativeType{Type: rt}
	case reflect.Map:
		return &NativeType{Type: rt}
	case reflect.Ptr:
		return &NativeType{Type: rt}
	case reflect.Struct:
		return &NativeType{Type: rt}
	case reflect.UnsafePointer:
		panic("not yet implemented")
	default:
		panic(fmt.Sprintf(
			"unexpected type %v", rt))
	}
}

// Implements Store.
func (ds *defaultStore) SetStrictGo2GnoMapping(strict bool) {
	ds.go2gnoStrict = strict
}

// Implements Store.
// See go2GnoValue2(). Like go2GnoType() but also converts any
// top-level complex types (or pointers to them).  The result gets
// memoized in *NativeType.GnoType() for type inference in the
// preprocessor, as well as in the cacheNativeTypes lookup map to
// support recursive translations.
// The namedness of the native type gets converted to an
// appropriate gno *DeclaredType with native methods
// converted via go2GnoFuncType().
func (ds *defaultStore) Go2GnoType(rt reflect.Type) (t Type) {
	if gnot, ok := ds.cacheNativeTypes[rt]; ok {
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
			// mappings have been removed, so for any non-builtin type in strict mode,
			// this will panic.
			if ds.go2gnoStrict {
				// mapping failed and strict: error.
				gokey := pkgPath + "." + rt.Name()
				panic(fmt.Sprintf("native type does not exist for %s", gokey))
			}

			// generate a new gno type for testing.
			mtvs := []TypedValue(nil)
			if t.Kind() == InterfaceKind {
				// methods already set on t.Methods.
				// *DT.Methods not used in Go for interfaces.
			} else {
				prt := rt
				if rt.Kind() != reflect.Ptr {
					// NOTE: go reflect requires ptr kind
					// for methods with ptr receivers,
					// whereas gno methods are all
					// declared on the *DeclaredType.
					prt = reflect.PtrTo(rt)
				}
				nm := prt.NumMethod()
				mtvs = make([]TypedValue, nm)
				for i := 0; i < nm; i++ {
					mthd := prt.Method(i)
					ft := ds.go2GnoFuncType(mthd.Type)
					fv := &FuncValue{
						Type:       ft,
						IsMethod:   true,
						Source:     nil,
						Name:       Name(mthd.Name),
						Closure:    nil,
						PkgPath:    pkgPath,
						body:       nil, // XXX
						nativeBody: nil,
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
			if gnot, ok := ds.cacheNativeTypes[rt]; ok {
				if gnot.TypeID() != baseOf(t).TypeID() {
					panic("should not happen")
				}
			}
		}
		ds.cacheNativeTypes[rt] = t // may overwrite
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
		ds.cacheNativeTypes[rt] = at
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
		ds.cacheNativeTypes[rt] = ct
		// define gno type
		chdir := toChanDir(rt.ChanDir())
		ct.Dir = chdir
		ct.Elt = go2GnoType(rt.Elem())
		return ct
	case reflect.Func:
		return ds.go2GnoFuncType(rt)
	case reflect.Interface:
		// predefine gno type
		it := &InterfaceType{}
		ds.cacheNativeTypes[rt] = it
		// define gno type
		nm := rt.NumMethod()
		fs := make([]FieldType, nm)
		for i := 0; i < nm; i++ {
			mthd := rt.Method(i)
			fs[i] = FieldType{
				Name: Name(mthd.Name),
				Type: ds.Go2GnoType(mthd.Type), // recursive
			}
		}
		it.PkgPath = rt.PkgPath()
		it.Methods = fs
		return it
	case reflect.Map:
		// predefine gno type
		mt := &MapType{}
		ds.cacheNativeTypes[rt] = mt
		// define gno type
		mt.Key = go2GnoType(rt.Key())
		mt.Value = go2GnoType(rt.Elem())
		return mt
	case reflect.Ptr:
		return &PointerType{
			Elt: ds.Go2GnoType(rt.Elem()), // recursive
		}
	case reflect.Struct:
		// predefine gno type
		st := &StructType{}
		ds.cacheNativeTypes[rt] = st
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

// Converts native go function type to gno *FuncType,
// for the preprocessor to infer types of arguments.
// The argument and return types are shallowly converted
// to gno types, (to preserve the original native types).
func (ds *defaultStore) go2GnoFuncType(rt reflect.Type) *FuncType {
	// predefine func type
	ft := &FuncType{}
	ds.cacheNativeTypes[rt] = ft
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

// NOTE: used by vm module.  Recursively converts.
func Go2GnoValue(alloc *Allocator, store Store, rv reflect.Value) (tv TypedValue) {
	return go2GnoValue2(alloc, store, rv, true)
}

// NOTE: used by vm module. Shallow, preserves native namedness.
func Go2GnoNativeValue(alloc *Allocator, rv reflect.Value) (tv TypedValue) {
	return go2GnoValue(alloc, rv)
}

// NOTE: used by imports_test.go TestSetOrigCaller.
func Gno2GoValue(tv *TypedValue, rv reflect.Value) (ret reflect.Value) {
	return gno2GoValue(tv, rv)
}

// Default run-time representation of go-native values.  It is
// "lazy" in the sense that unnamed complex types like arrays and
// slices aren't translated to Gno canonical types except as
// *NativeType/*NativeValues, primarily for speed.  To force
// translation to Gno canonical types for unnamed complex types,
// call go2GnoValue2(), which is used by the implementation of
// ConvertTo().
// Unlike go2GnoValue2(), rv may be invalid.
func go2GnoValue(alloc *Allocator, rv reflect.Value) (tv TypedValue) {
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
		tv.T = alloc.NewType(&NativeType{Type: rt})
		tv.V = alloc.NewNative(rv)
		return
	}
	tv.T = alloc.NewType(go2GnoType(rv.Type()))
	switch rk := rv.Kind(); rk {
	case reflect.Bool:
		tv.SetBool(rv.Bool())
	case reflect.String:
		tv.V = alloc.NewString(rv.String())
	case reflect.Int:
		tv.SetInt(int(rv.Int()))
	case reflect.Int8:
		tv.SetInt8(int8(rv.Int()))
	case reflect.Int16:
		tv.SetInt16(int16(rv.Int()))
	case reflect.Int32:
		tv.SetInt32(int32(rv.Int()))
	case reflect.Int64:
		tv.SetInt64(rv.Int())
	case reflect.Uint:
		tv.SetUint(uint(rv.Uint()))
	case reflect.Uint8:
		tv.SetUint8(uint8(rv.Uint()))
	case reflect.Uint16:
		tv.SetUint16(uint16(rv.Uint()))
	case reflect.Uint32:
		tv.SetUint32(uint32(rv.Uint()))
	case reflect.Uint64:
		tv.SetUint64(rv.Uint())
	case reflect.Float32:
		tv.SetFloat32(float32(rv.Float()))
	case reflect.Float64:
		tv.SetFloat64(rv.Float())
	case reflect.Array:
		tv.V = alloc.NewNative(rv)
	case reflect.Slice:
		tv.V = alloc.NewNative(rv)
	case reflect.Chan:
		tv.V = alloc.NewNative(rv)
	case reflect.Func:
		tv.V = alloc.NewNative(rv)
	case reflect.Interface:
		panic("should not happen")
	case reflect.Map:
		tv.V = alloc.NewNative(rv)
	case reflect.Ptr:
		tv.V = alloc.NewNative(rv)
	case reflect.Struct:
		tv.V = alloc.NewNative(rv)
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
// become initialized.  NOTE: Due to limitations of Go 1.15
// reflection, any child Gno declared types cannot change
// types, and pointer values cannot change.
func go2GnoValueUpdate(alloc *Allocator, rlm *Realm, lvl int, tv *TypedValue, rv reflect.Value) {
	// Special case if nil:
	if tv.IsUndefined() {
		return // do nothing
	}
	// Special case if native type:
	if _, ok := tv.T.(*NativeType); ok {
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
			tv.V = alloc.NewString(rv.String())
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
			tv.SetInt64(rv.Int())
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
			tv.SetUint64(rv.Uint())
		}
	case Float32Kind:
		if lvl != 0 {
			tv.SetFloat32(float32(rv.Float()))
		}
	case Float64Kind:
		if lvl != 0 {
			tv.SetFloat64(rv.Float())
		}
	case BigintKind:
		panic("not yet implemented")
	case BigdecKind:
		panic("not yet implemented")
	case ArrayKind:
		av := tv.V.(*ArrayValue)
		rvl := rv.Len()
		if debug {
			if rvl != baseOf(tv.T).(*ArrayType).Len {
				panic("go-native update error: array length mismatch")
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
					etv.V = defaultValue(alloc, et)
				}
				go2GnoValueUpdate(alloc, rlm, lvl+1, etv, erv)
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
				panic("go-native update error: slice length mismatch")
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
					etv.V = defaultValue(alloc, et)
				}
				go2GnoValueUpdate(alloc, rlm, lvl+1, etv, erv)
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
		go2GnoValueUpdate(alloc, rlm, lvl+1, etv, erv)
	case StructKind:
		st := baseOf(tv.T).(*StructType)
		sv := tv.V.(*StructValue)
		for i := range st.Fields {
			ft := st.Fields[i].Type
			ftv := &sv.Fields[i]
			// XXX use Assign and Realm?
			if ftv.T == nil && ft.Kind() != InterfaceKind {
				ftv.T = ft
			}
			if ftv.V == nil {
				ftv.V = defaultValue(alloc, ft)
			}
			frv := rv.Field(i)
			go2GnoValueUpdate(alloc, rlm, lvl+1, ftv, frv)
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
				go2GnoValueUpdate(alloc, rlm, lvl+1, vtv, vrv)
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
			ktv := go2GnoValue(alloc, k)
			vtv := go2GnoValue(alloc, v)
			ptr := mv.GetPointerForKey(alloc, nil, &ktv)
			if debug {
				if !ptr.TV.IsUndefined() {
					panic("should not happen")
				}
			}
			ptr.Assign2(alloc, nil, rlm, vtv, false) // document false
		}
	case TypeKind:
		panic("not yet implemented")
	default:
		panic("should not happen: unexpected gno kind")
	}
	return
}

// used for direct comparison to error types
var tError = reflect.TypeOf(new(error)).Elem()

// If recursive is false, this function is like go2GnoValue() but less lazy
// (but still not recursive/eager). When recursive is false, it is for
// converting Go types to Gno types upon an explicit conversion (via
// ConvertTo).  Panics on unexported/private fields. Some types that cannot be
// converted remain native. Unlike go2GnoValue(), rv must be valid.
func go2GnoValue2(alloc *Allocator, store Store, rv reflect.Value, recursive bool) (tv TypedValue) {
	if debug {
		if !rv.IsValid() {
			panic("go2GnoValue2() requires valid rv")
		}
	}
	tv.T = store.Go2GnoType(rv.Type())
	switch rk := rv.Kind(); rk {
	case reflect.Bool:
		tv.SetBool(rv.Bool())
	case reflect.String:
		tv.V = alloc.NewString(rv.String())
	case reflect.Int:
		tv.SetInt(int(rv.Int()))
	case reflect.Int8:
		tv.SetInt8(int8(rv.Int()))
	case reflect.Int16:
		tv.SetInt16(int16(rv.Int()))
	case reflect.Int32:
		tv.SetInt32(int32(rv.Int()))
	case reflect.Int64:
		tv.SetInt64(rv.Int())
	case reflect.Uint:
		tv.SetUint(uint(rv.Uint()))
	case reflect.Uint8:
		tv.SetUint8(uint8(rv.Uint()))
	case reflect.Uint16:
		tv.SetUint16(uint16(rv.Uint()))
	case reflect.Uint32:
		tv.SetUint32(uint32(rv.Uint()))
	case reflect.Uint64:
		tv.SetUint64(rv.Uint())
	case reflect.Float32:
		tv.SetFloat32(float32(rv.Float()))
	case reflect.Float64:
		tv.SetFloat64(rv.Float())
	case reflect.Array:
		rvl := rv.Len()
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			av := alloc.NewDataArray(rvl)
			data := av.Data
			reflect.Copy(reflect.ValueOf(data), rv)
			tv.V = av
		} else {
			av := alloc.NewListArray(rvl)
			list := av.List
			for i := 0; i < rvl; i++ {
				if recursive {
					list[i] = go2GnoValue2(alloc, store, rv.Index(i), true)
				} else {
					list[i] = go2GnoValue(alloc, rv.Index(i))
				}
			}
			tv.V = av
		}
	case reflect.Slice:
		rvl := rv.Len()
		rvc := rv.Cap()
		list := make([]TypedValue, rvl, rvc)
		for i := 0; i < rvl; i++ {
			if recursive {
				list[i] = go2GnoValue2(alloc, store, rv.Index(i), true)
			} else {
				list[i] = go2GnoValue(alloc, rv.Index(i))
			}
		}
		tv.V = alloc.NewSliceFromList(list)
	case reflect.Chan:
		panic("not yet implemented")
	case reflect.Func:
		// NOTE: the type may be a full gno type, either a
		// *FuncType or *DeclaredType.  The value may still be a
		// *NativeValue though, and the function can be called
		// regardless.
		tv.V = alloc.NewNative(rv)
	case reflect.Interface:
		// special case for errors, which are very often used especially in
		// native bindings
		if rv.Type() == tError {
			tv.T = gErrorType
			if !rv.IsNil() {
				tv.V = alloc.NewNative(rv.Elem())
			}
			return
		}
		panic("not yet implemented")
	case reflect.Map:
		panic("not yet implemented")
	case reflect.Ptr:
		val := go2GnoValue2(alloc, store, rv.Elem(), recursive)
		tv.V = PointerValue{TV: &val} // heap alloc
	case reflect.Struct:
		nf := rv.NumField()
		fs := alloc.NewStructFields(nf)
		for i := 0; i < nf; i++ {
			frv := rv.Field(i)
			if recursive {
				fs[i] = go2GnoValue2(alloc, store, frv, true)
			} else {
				fs[i] = go2GnoValue(alloc, frv)
			}
		}
		tv.V = alloc.NewStruct(fs)
	case reflect.UnsafePointer:
		panic("not yet implemented")
	default:
		panic("not yet implemented")
	}
	return
}

// ----------------------------------------
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
		case Float32Type:
			return reflect.TypeOf(float32(0))
		case Float64Type:
			return reflect.TypeOf(float64(0))
		case BigintType, UntypedBigintType:
			panic("not yet implemented")
		case BigdecType, UntypedBigdecType:
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
	case *NativeType:
		return ct.Type
	default:
		panic(fmt.Sprintf("unexpected type %v with base %v", t, baseOf(t)))
	}
}

// If gno2GoTypeMatches(t, rt) is true, a t value can
// be converted to an rt native value using gno2GoValue(v, rv).
// This is called when autoNative is true in checkType().
// This is used for all native function calls, and also
// for testing whether a native value implements a gno interface.
func gno2GoTypeMatches(t Type, rt reflect.Type) (result bool) {
	if rt == nil {
		panic("should not happen")
	}
	// special case if t == Float32Type or Float64Type
	if t == Float32Type {
		return rt.Kind() == reflect.Float32
	} else if t == Float64Type {
		return rt.Kind() == reflect.Float64
	}
	switch ct := baseOf(t).(type) {
	case PrimitiveType:
		switch ct {
		case BoolType, UntypedBoolType:
			return rt.Kind() == reflect.Bool
		case StringType, UntypedStringType:
			return rt.Kind() == reflect.String
		case IntType:
			return rt.Kind() == reflect.Int
		case Int8Type:
			return rt.Kind() == reflect.Int8
		case Int16Type:
			return rt.Kind() == reflect.Int16
		case Int32Type, UntypedRuneType:
			return rt.Kind() == reflect.Int32
		case Int64Type:
			return rt.Kind() == reflect.Int64
		case UintType:
			return rt.Kind() == reflect.Uint
		case Uint8Type:
			return rt.Kind() == reflect.Uint8
		case Uint16Type:
			return rt.Kind() == reflect.Uint16
		case Uint32Type:
			return rt.Kind() == reflect.Uint32
		case Uint64Type:
			return rt.Kind() == reflect.Uint64
		case Float32Type:
			return rt.Kind() == reflect.Float32
		case Float64Type:
			return rt.Kind() == reflect.Float64
		case BigintType, UntypedBigintType:
			panic("not yet implemented")
		case BigdecType, UntypedBigdecType:
			panic("not yet implemented")
		default:
			panic("should not happen")
		}
	case *PointerType:
		if rt.Kind() != reflect.Ptr {
			return false
		}
		return gno2GoTypeMatches(ct.Elt, rt.Elem())
	case *ArrayType:
		if rt.Kind() != reflect.Array {
			return false
		}
		if ct.Len != rt.Len() {
			return false
		}
		return gno2GoTypeMatches(ct.Elt, rt.Elem())
	case *SliceType:
		if rt.Kind() != reflect.Slice {
			return false
		}
		return gno2GoTypeMatches(ct.Elt, rt.Elem())
	case *StructType:
		// TODO maybe consider automatically skipping private native fields?
		for i, field := range ct.Fields {
			rft := rt.Field(i).Type
			if !gno2GoTypeMatches(field.Type, rft) {
				return false
			}
		}
		return true
	case *MapType:
		if !gno2GoTypeMatches(ct.Key, rt.Key()) {
			return false
		}
		if !gno2GoTypeMatches(ct.Value, rt.Elem()) {
			return false
		}
		return true
	case *FuncType:
		// TODO: there is a recursion issue when a native func
		// takes a func type as an argument.  as implemented,
		// we match too broadly.
		//
		// args must auto-match.
		for i, pt := range ct.Params {
			if !gno2GoTypeMatches(pt.Type, rt.In(i)) {
				return false
			}
		}
		// go2GnoType(result) must match directly.
		for i, rct := range ct.Results {
			rrt := rt.Out(i)
			gnorrt := go2GnoType(rrt)
			if rct.Type.Kind() == InterfaceKind {
				if !IsImplementedBy(rct.Type, gnorrt) {
					return false
				}
			} else if rct.Type.TypeID() != gnorrt.TypeID() {
				return false
			}
		}
		// variadicity must be the same
		if ct.HasVarg() != rt.IsVariadic() {
			return false
		}
		return true
	case *InterfaceType:
		if rt.Kind() != reflect.Interface {
			return false
		}
		if ct.IsEmptyInterface() {
			return rt.NumMethod() == 0
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
	case *NativeType:
		return ct.Type.AssignableTo(rt)
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
	var rt reflect.Type
	bt := baseOf(tv.T)
	if !rv.IsValid() {
		rt = gno2GoType(bt)
		rv = reflect.New(rt).Elem()
		ret = rv
	} else if rv.Kind() == reflect.Interface {
		if debug {
			if !rv.IsZero() {
				panic("should not happen")
			}
		}
		rt = gno2GoType(bt)
		rv1 := rv
		rv2 := reflect.New(rt).Elem()
		rv = rv2       // swaparoo
		defer func() { // TODO: improve?
			rv1.Set(rv2)
			ret = rv
		}()
	} else {
		ret = rv
		rt = rv.Type()
	}
	switch ct := bt.(type) {
	case PrimitiveType:
		switch ct {
		case BoolType, UntypedBoolType:
			rv.SetBool(tv.GetBool())
		case StringType, UntypedStringType:
			rv.SetString(tv.GetString())
		case IntType:
			rv.SetInt(int64(tv.GetInt()))
		case Int8Type:
			rv.SetInt(int64(tv.GetInt8()))
		case Int16Type:
			rv.SetInt(int64(tv.GetInt16()))
		case Int32Type, UntypedRuneType:
			rv.SetInt(int64(tv.GetInt32()))
		case Int64Type:
			rv.SetInt(tv.GetInt64())
		case UintType:
			rv.SetUint(uint64(tv.GetUint()))
		case Uint8Type:
			rv.SetUint(uint64(tv.GetUint8()))
		case Uint16Type:
			rv.SetUint(uint64(tv.GetUint16()))
		case Uint32Type:
			rv.SetUint(uint64(tv.GetUint32()))
		case Uint64Type:
			rv.SetUint(tv.GetUint64())
		case Float32Type:
			rv.SetFloat(float64(tv.GetFloat32()))
		case Float64Type:
			rv.SetFloat(tv.GetFloat64())
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
			rve := reflect.New(rv.Type().Elem()).Elem()
			rv2 := gno2GoValue(tv.V.(PointerValue).TV, rve)
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
			for i := 0; i < ct.Len; i++ {
				val := av.Data[i]
				erv := rv.Index(i)
				erv.SetUint(uint64(val))
			}
		}
	case *SliceType:
		st := rt
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
		for i := range ct.Fields {
			ftv := &(sv.Fields[i])
			if ftv.IsUndefined() {
				continue
			}
			gno2GoValue(ftv, rv.Field(i))
		}
	case *MapType:
		// If uninitialized map, return zero value.
		if tv.V == nil {
			return
		}
		// General case.
		mv := tv.V.(*MapValue)
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
	case *NativeType:
		// If uninitialized native type, leave rv uninitialized.
		if tv.V == nil {
			return
		}
		// General case.
		rv.Set(tv.V.(*NativeValue).Value)
	case *DeclaredType:
		// See corresponding note on gno2GoType().
		panic("should not happen") // we switch on baseOf().
	case *FuncType:
		// TODO: if tv.V.(*NativeValue), just return.
		// TODO: otherwise, set rv to wrapper.
		panic("gno2Go not supported for gno functions yet")
	default:
		panic(fmt.Sprintf(
			"unexpected type %s",
			tv.T.String()))
	}
	return
}

// ----------------------------------------
// PackageNode methods

func (x *PackageNode) DefineGoNativeType(rt reflect.Type) {
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
	nt := &NativeType{Type: rt}
	x.Define(Name(name), asValue(nt))
}

func (x *PackageNode) DefineGoNativeValue(n Name, nv interface{}) {
	if debug {
		debug.Printf("*PackageNode.DefineGoNativeValue(%s)\n", reflect.ValueOf(nv).String())
	}
	rv := reflect.ValueOf(nv)
	// rv is not settable, so create something that is.
	rt := rv.Type()
	rv2 := reflect.New(rt).Elem()
	rv2.Set(rv)
	x.Define(n, go2GnoValue(nilAllocator, rv2))
}

// ----------------------------------------
// Machine methods

func (m *Machine) doOpArrayLitGoNative() {
	// assess performance TODO
	x := m.PopExpr().(*CompositeLitExpr)
	el := len(x.Elts) // may be incomplete
	// peek array type.
	xt := m.PeekValue(1 + el).V.(TypeValue).Type
	nt := xt.(*NativeType)
	rv := reflect.New(nt.Type).Elem()
	// construct array value.
	if 0 < el {
		itvs := m.PopValues(el)
		for i := 0; i < el; i++ {
			if kx := x.Elts[i].Key; kx != nil {
				// XXX why convert? (also see doOpArrayLit())
				k := kx.(*ConstExpr).ConvertGetInt()
				rf := rv.Index(k)
				gno2GoValue(&itvs[i], rf)
			} else {
				rf := rv.Index(i)
				gno2GoValue(&itvs[i], rf)
			}
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
	nv := m.Alloc.NewNative(rv)
	m.PushValue(TypedValue{
		T: nt,
		V: nv,
	})
}

func (m *Machine) doOpSliceLitGoNative() {
	// assess performance TODO
	x := m.PopExpr().(*CompositeLitExpr)
	el := len(x.Elts) // may be incomplete
	// peek slice type.
	xt := m.PeekValue(1 + el).V.(TypeValue).Type
	nt := xt.(*NativeType)
	at := reflect.ArrayOf(el, nt.Type.Elem())
	rv := reflect.New(at).Elem()
	// construct array value.
	if 0 < el {
		itvs := m.PopValues(el)
		for i := 0; i < el; i++ {
			if kx := x.Elts[i].Key; kx != nil {
				// XXX why convert? (also see doOpArrayLit())
				k := kx.(*ConstExpr).ConvertGetInt()
				rf := rv.Index(k)
				gno2GoValue(&itvs[i], rf)
			} else {
				rf := rv.Index(i)
				gno2GoValue(&itvs[i], rf)
			}
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
	nv := m.Alloc.NewNative(rv.Slice(0, el))
	m.PushValue(TypedValue{
		T: nt,
		V: nv,
	})
}

func (m *Machine) doOpStructLitGoNative() {
	// assess performance TODO
	x := m.PopExpr().(*CompositeLitExpr)
	el := len(x.Elts) // may be incomplete
	// peek struct type.
	xt := m.PeekValue(1 + el).V.(TypeValue).Type
	nt := xt.(*NativeType)
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
	m.PushValue(TypedValue{
		T: nt,
		V: m.Alloc.NewNative(rv),
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
	m.PopValue() // func value
	prvs := make([]reflect.Value, 0, len(ptvs))
	for i := 0; i < fr.NumArgs; i++ {
		ptv := &ptvs[i]
		var it reflect.Type
		if hasVarg && numParams-1 <= i && !isVarg {
			it = ft.In(numParams - 1)
			it = it.Elem()
		} else {
			it = ft.In(i)
		}
		erv := reflect.New(it).Elem()
		prvs = append(prvs, gno2GoValue(ptv, erv))
	}
	// call and get results.
	rrvs := []reflect.Value(nil)
	if isVarg {
		rrvs = fv.Value.CallSlice(prvs)
	} else {
		rrvs = fv.Value.Call(prvs)
	}
	// convert and push results.
	for _, rvs := range rrvs {
		// TODO instead of this shallow conversion,
		// look at expected Gno type and convert appropriately.
		rtv := go2GnoValue(m.Alloc, rvs)
		m.PushValue(rtv)
	}
	// carry writes to params if needed.
	for i := 0; i < fr.NumArgs; i++ {
		ptv := &ptvs[i]
		prv := prvs[i]
		if !ptv.IsUndefined() {
			go2GnoValueUpdate(m.Alloc, m.Realm, 0, ptv, prv)
		}
	}
	// cleanup
	m.NumResults = fv.Value.Type().NumOut()
	m.PopFrame()
}

// ----------------------------------------
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
