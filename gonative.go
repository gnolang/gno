package gno

import (
	"fmt"
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

// See go2GnoValue2(). Like go2GnoType() but also converts any
// top-level complex types (or pointers to them).  The result gets
// memoized in *nativeType.GnoType() for type inference in the
// preprocessor.
func go2GnoType2(rt reflect.Type) Type {
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
	case reflect.Array:
		return &ArrayType{
			Len: rt.Len(),
			Elt: go2GnoType(rt.Elem()),
			Vrd: false,
		}
	case reflect.Slice:
		return &SliceType{
			Elt: go2GnoType(rt.Elem()),
			Vrd: false,
		}
	case reflect.Chan:
		chdir := toChanDir(rt.ChanDir())
		return &ChanType{
			Dir: chdir,
			Elt: go2GnoType(rt.Elem()),
		}
	case reflect.Func:
		return go2GnoFuncType(rt)
	case reflect.Interface:
		panic("not yet immplemented")
	case reflect.Map:
		panic("not yet immplemented")
	case reflect.Ptr:
		return PointerType{
			// this is the only recursive call to go2GnoType2().
			Elt: go2GnoType2(rt.Elem()),
		}
	case reflect.Struct:
		nf := rt.NumField()
		fs := make([]FieldType, nf)
		mp := make([]int, nf)
		// NOTE: go-native struct fields don't stack/flatten like
		// nested gno struct fields, but embedded fields must be
		// referred to explicitly.
		for i := 0; i < nf; i++ {
			sf := rt.Field(i)
			fs[i] = FieldType{
				Name: Name(sf.Name),
				Type: go2GnoType(sf.Type),
			}
			mp[i] = i // see note
		}
		return &StructType{
			PkgPath: rt.PkgPath(),
			Fields:  fs,
			Mapping: mp,
		}
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
func go2GnoValue(rv reflect.Value) (tv TypedValue) {
	if rv.Type().PkgPath() != "" {
		rt := rv.Type()
		tv.T = &nativeType{Type: rt}
		tv.V = &nativeValue{Value: rv}
		return
	}
REFLECT_KIND_SWITCH:
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
	case reflect.Array:
		tv.V = &nativeValue{rv}
	case reflect.Slice:
		tv.V = &nativeValue{rv}
	case reflect.Chan:
		tv.V = &nativeValue{rv}
	case reflect.Func:
		tv.V = &nativeValue{rv}
	case reflect.Interface:
		if rv.IsNil() {
			tv.V = nil // nil-interface, "undefined".
		} else {
			rv = rv.Elem()
			goto REFLECT_KIND_SWITCH
		}
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
func go2GnoValueUpdate(lvl int, tv *TypedValue, rv reflect.Value) {
	// Special case if native type:
	if _, ok := tv.T.(*nativeType); ok {
		return // do nothing
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
			if rvl != tv.T.(*ArrayType).Len {
				panic("go-native update error: array length mismmatch")
			}
		}
		if av.Data == nil {
			at := baseOf(tv.T).(*ArrayType)
			et := at.Elt
			for i := 0; i < rvl; i++ {
				erv := rv.Index(i)
				etv := &av.List[i]
				if etv.T == nil && et.Kind() != InterfaceKind {
					etv.T = et
				}
				if etv.V == nil {
					etv.V = defaultValue(et)
				}
				go2GnoValueUpdate(lvl+1, etv, erv)
			}
		} else {
			for i := 0; i < rvl; i++ {
				erv := rv.Index(i)
				av.Data[i] = uint8(erv.Uint())
			}
		}
	case SliceKind:
		sv := tv.V.(*SliceValue)
		rvl := rv.Len()
		if debug {
			if rvl != sv.GetLength() {
				panic("go-native update error: slice length mismmatch")
			}
		}
		if sv.Base.Data == nil {
			st := baseOf(tv.T).(*SliceType)
			et := st.Elt
			for i := 0; i < rvl; i++ {
				erv := rv.Index(i)
				etv := &sv.Base.List[i]
				if etv.T == nil && et.Kind() != InterfaceKind {
					etv.T = et
				}
				if etv.V == nil {
					etv.V = defaultValue(et)
				}
				go2GnoValueUpdate(lvl+1, etv, erv)
			}
		} else {
			for i := 0; i < rvl; i++ {
				erv := rv.Index(i)
				sv.Base.Data[i] = uint8(erv.Uint())
			}
		}
	case PointerKind:
		pv := tv.V.(PointerValue)
		etv := pv.TypedValue
		erv := rv.Elem()
		go2GnoValueUpdate(lvl+1, etv, erv)
	case StructKind:
		st := baseOf(tv.T).(*StructType)
		sv := tv.V.(*StructValue)
		for orig, flat := range st.Mapping {
			ft := st.Fields[flat].Type
			ftv := &sv.Fields[flat]
			if ftv.T == nil && ft.Kind() != InterfaceKind {
				ftv.T = ft
			}
			if ftv.V == nil {
				ftv.V = defaultValue(ft)
			}
			frv := rv.Field(orig)
			go2GnoValueUpdate(lvl+1, ftv, frv)
		}
	case PackageKind:
		panic("not yet implemented")
	case InterfaceKind:
		panic("not yet implemented")
	case ChanKind:
		panic("not yet implemented")
	case FuncKind:
		panic("not yet implemented")
	case MapKind:
		panic("not yet implemented")
	case TypeKind:
		panic("not yet implemented")
	default:
		panic("should not happen: unexpected gno kind")
	}
	return
}

// This function is like go2GnoValue() but less lazy (but still not
// recursive/eager). It is for converting Go types to Gno types upon
// an explicit conversion (via ConvertTo).  Panics on
// unexported/private fields.
// Due to limitations of Go1.15, the namedness is dropped rather than
// converted.  This lets users convert go-native types to named or unnamed Gno
// types (sans private fields) via conversion.  The conversion is not
// recursive, and the extra conversion works on the top-level complex
// type/value, or a pointer to that type/value.  Some types that cannot
// be converted remain native.
func go2GnoValue2(rv reflect.Value) (tv TypedValue) {
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
		tv.T = PointerType{Elt: go2GnoType2(rv.Type().Elem())}
		val := go2GnoValue2(rv.Elem())
		tv.V = PointerValue{TypedValue: &val} // heap alloc
	case reflect.Struct:
		panic("not yet implemented")
	case reflect.UnsafePointer:
		panic("not yet implemented")
	default:
		panic("not yet implemented")
	}
	return
}

// converts native go function type to gno *FuncType,
// for the preprocessor to infer types of arguments.
func go2GnoFuncType(rt reflect.Type) *FuncType {
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
	return &FuncType{
		PkgPath: "", // dunno with native.
		Params:  ins,
		Results: outs,
	}
}

//----------------------------------------
// Gno to Go conversion

func gno2GoType(t Type) reflect.Type {
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
	case PointerType:
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
		panic("should not happen")
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
		panic("should not happen")
	}
}

// rv must be addressable, or zero (invalid) if tv is referred to from a
// gno.PointerValue.  If rv is zero, an addressable one will be constructed and
// returned, otherwise returns rv.
func gno2GoValue(tv *TypedValue, rv reflect.Value) reflect.Value {
	if tv.IsNilInterface() {
		rt := gno2GoType(tv.T)
		rv = reflect.New(rt).Elem()
		return rv
	} else if tv.IsUndefined() {
		if debug {
			if !rv.IsValid() {
				panic("unexpected undefined gno value")
			}
		}
		return rv
	}
	bt := baseOf(tv.T)
	var rt reflect.Type
	if !rv.IsValid() {
		rt = gno2GoType(bt)
		rv = reflect.New(rt).Elem()
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
	case PointerType:
		// This doesn't take into account pointer relativity, or even
		// identical pointers -- every non-nil gno pointer type results in a
		// new addressable value in go.
		rv2 := gno2GoValue(tv.V.(PointerValue).TypedValue, reflect.Value{})
		rv.Set(rv2.Addr())
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
			return rv
		}
		// General case.
		sv := tv.V.(*SliceValue)
		svo := sv.Offset
		svl := sv.Length
		svc := sv.Maxcap
		if sv.Base.Data == nil {
			rv.Set(reflect.MakeSlice(st, svl, svc))
			for i := 0; i < svl; i++ {
				etv := &(sv.Base.List[svo+i])
				if etv.IsUndefined() {
					continue
				}
				gno2GoValue(etv, rv.Index(i))
			}
		} else {
			data := make([]byte, svl, svc)
			copy(data[:svc], sv.Base.Data[svo:svo+svc])
			rv.Set(reflect.ValueOf(data))
		}
	case *StructType:
		// If uninitialized struct, return zero value.
		if tv.V == nil {
			return rv
		}
		// General case.
		sv := tv.V.(*StructValue)
		// Use st.Mapping to translate from Go to Gno field numbers.
		for orig, flat := range ct.Mapping {
			ftv := &(sv.Fields[flat])
			if ftv.IsUndefined() {
				continue
			}
			gno2GoValue(ftv, rv.Field(orig))
		}
	case *MapType:
		// If uninitialized map, return zero value.
		if tv.V == nil {
			return rv
		}
		// General case.
		mt := gno2GoType(ct)
		mv := tv.V.(*MapValue)
		rv.Set(reflect.MakeMapWithSize(mt, mv.List.Size))
		head := mv.List.Head
		for head != nil {
			ktv, vtv := &head.Key, &head.Value
			krv := gno2GoValue(ktv, reflect.Value{})
			vrv := gno2GoValue(vtv, reflect.Value{})
			rv.SetMapIndex(krv, vrv)
			head = head.Next
		}
	case *nativeType:
		// If uninitialized native type, leave rv uninitialized.
		if tv.V == nil {
			return rv
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
	return rv
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

func (m *Machine) doOpCallGoNative() {
	fr := m.LastFrame()
	fv := fr.GoFunc
	ft := fv.Value.Type()
	hasVarg := ft.IsVariadic()
	numParams := ft.NumIn()
	isVarg := fr.IsVarg
	// pop and convert params.
	ptvs := m.PopValues(fr.NumArgs)
	prvs := make([]reflect.Value, len(ptvs))
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
			prvs[i] = gno2GoValue(ptv, erv)
		} else {
			prvs[i] = gno2GoValue(ptv, reflect.Value{})
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
		go2GnoValueUpdate(0, ptv, prv)
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
