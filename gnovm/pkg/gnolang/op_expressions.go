package gnolang

import (
	"fmt"
	"reflect"
)

// OpBinary1 defined in op_binary.go

// NOTE: keep in sync with doOpIndex2.
func (m *Machine) doOpIndex1() {
	if debug {
		_ = m.PopExpr().(*IndexExpr)
	} else {
		m.PopExpr()
	}
	iv := m.PopValue()   // index
	xv := m.PeekValue(1) // x
	switch ct := baseOf(xv.T).(type) {
	case *MapType:
		mv := xv.V.(*MapValue)
		vv, exists := mv.GetValueForKey(m.Store, iv)
		if exists {
			*xv = vv // reuse as result
		} else {
			vt := ct.Value
			*xv = TypedValue{ // reuse as result
				T: vt,
				V: defaultValue(m.Alloc, vt),
			}
		}
	default:
		res := xv.GetPointerAtIndex(m.Alloc, m.Store, iv)
		*xv = res.Deref() // reuse as result
	}
}

// NOTE: keep in sync with doOpIndex1.
func (m *Machine) doOpIndex2() {
	if debug {
		_ = m.PopExpr().(*IndexExpr)
	} else {
		m.PopExpr()
	}
	iv := m.PeekValue(1) // index
	xv := m.PeekValue(2) // x
	switch ct := baseOf(xv.T).(type) {
	case *MapType:
		vt := ct.Value
		if xv.V == nil { // uninitialized map
			*xv = TypedValue{ // reuse as result
				T: vt,
				V: defaultValue(m.Alloc, vt),
			}
			*iv = untypedBool(false) // reuse as result
		} else {
			mv := xv.V.(*MapValue)
			vv, exists := mv.GetValueForKey(m.Store, iv)
			if exists {
				*xv = vv                // reuse as result
				*iv = untypedBool(true) // reuse as result
			} else {
				*xv = TypedValue{ // reuse as result
					T: vt,
					V: defaultValue(m.Alloc, vt),
				}
				*iv = untypedBool(false) // reuse as result
			}
		}
	case *NativeType:
		// TODO: see doOpIndex1()
		panic("not yet implemented")
	default:
		panic("should not happen")
	}
}

func (m *Machine) doOpSelector() {
	sx := m.PopExpr().(*SelectorExpr)
	xv := m.PeekValue(1)
	res := xv.GetPointerTo(m.Alloc, m.Store, sx.Path).Deref()
	if debug {
		m.Printf("-v[S] %v\n", xv)
		m.Printf("+v[S] %v\n", res)
	}
	*xv = res // reuse as result
}

func (m *Machine) doOpSlice() {
	sx := m.PopExpr().(*SliceExpr)
	var low, high, max int = -1, -1, -1
	// max
	if sx.Max != nil {
		max = m.PopValue().ConvertGetInt()
	}
	// high
	if sx.High != nil {
		high = m.PopValue().ConvertGetInt()
	}
	// low
	if sx.Low != nil {
		low = m.PopValue().ConvertGetInt()
	} else {
		low = 0
	}
	// slice base x
	xv := m.PopValue()
	// if a is a pointer to an array, a[low : high : max] is
	// shorthand for (*a)[low : high : max]
	if xv.T.Kind() == PointerKind &&
		xv.T.Elem().Kind() == ArrayKind {
		// simply deref xv.
		*xv = xv.V.(PointerValue).Deref()
	}
	// fill default based on xv
	if sx.High == nil {
		high = xv.GetLength()
	}
	// all low:high:max cases
	if max == -1 {
		sv := xv.GetSlice(m.Alloc, low, high)
		m.PushValue(sv)
	} else {
		sv := xv.GetSlice2(m.Alloc, low, high, max)
		m.PushValue(sv)
	}
}

// If the referred value is undefined, and the pointer
// elem kind is not an interface kind, the appropriate
// type is set (and value becomes a typed-nil value).
//
// NOTE: OpStar is ambiguous -- it either means to
// dereference a pointer value, or to refer to the referred
// value in lhs, or it means to get the pointer-of a
// type. The fact that the same symbol is used to refer to
// both dereferencing (values) as well as referencing
// (types) may be a confusing factor for those new to
// C-like syntax. (it was for me).  We simply switch on the
// type of *StarExpr.X.  Since pointers and typevals are
// distinctly different kinds, the type-checker should
// catch all potential ambiguities where the intent is to
// deref, but the result is a pointer-to type.
func (m *Machine) doOpStar() {
	xv := m.PopValue()
	switch bt := baseOf(xv.T).(type) {
	case *PointerType:
		pv := xv.V.(PointerValue)
		if pv.TV.T == DataByteType {
			tv := TypedValue{T: bt.Elt}
			dbv := pv.TV.V.(DataByteValue)
			tv.SetUint8(dbv.GetByte())
			m.PushValue(tv)
		} else {
			if pv.TV.IsUndefined() && bt.Elt.Kind() != InterfaceKind {
				refv := TypedValue{T: bt.Elt}
				m.PushValue(refv)
			} else {
				m.PushValue(*pv.TV)
			}
		}
	case *TypeType:
		t := xv.GetType()
		var pt Type
		if nt, ok := t.(*NativeType); ok {
			pt = &NativeType{Type: reflect.PointerTo(nt.Type)}
		} else {
			pt = &PointerType{Elt: t}
		}
		m.PushValue(asValue(pt))
	case *NativeType:
		panic("not yet implemented")
	default:
		panic(fmt.Sprintf(
			"illegal star expression x type %s",
			xv.T.String()))
	}
}

// XXX this is wrong, for var i interface{}; &i is *interface{}.
func (m *Machine) doOpRef() {
	rx := m.PopExpr().(*RefExpr)
	m.Alloc.AllocatePointer()
	xv := m.PopAsPointer(rx.X)
	if nv, ok := xv.TV.V.(*NativeValue); ok {
		// If a native pointer, ensure it is addressable.  This
		// way, PointerValue{*NativeValue{rv}} can be converted
		// to/from *NativeValue{rv.Addr()}.
		if !nv.Value.CanAddr() {
			rv := nv.Value
			rt := rv.Type()
			rv2 := reflect.New(rt).Elem()
			rv2.Set(rv)
			nv.Value = rv2
		}
	}
	m.PushValue(TypedValue{
		T: m.Alloc.NewType(&PointerType{Elt: xv.TV.T}),
		V: xv,
	})
}

// NOTE: keep in sync with doOpTypeAssert2.
func (m *Machine) doOpTypeAssert1() {
	m.PopExpr()
	// pop type
	t := m.PopValue().GetType()
	// peek x for re-use
	xv := m.PeekValue(1)
	xt := xv.T

	if t.Kind() == InterfaceKind { // is interface assert
		if it, ok := baseOf(t).(*InterfaceType); ok {
			// t is Gno interface.
			// assert that x implements type.
			impl := false
			impl = it.IsImplementedBy(xt)
			if !impl {
				// TODO: default panic type?
				ex := fmt.Sprintf(
					"%s doesn't implement %s",
					xt.String(),
					it.String())
				m.Panic(typedString(ex))
				return
			}
			// NOTE: consider ability to push an
			// interface-restricted form
			// *xv = *xv
		} else if nt, ok := baseOf(t).(*NativeType); ok {
			// t is Go interface.
			// assert that x implements type.
			impl := false
			if nxt, ok := xt.(*NativeType); ok {
				impl = nxt.Type.Implements(nt.Type)
			} else {
				impl = false
			}
			if !impl {
				// TODO: default panic type?
				ex := fmt.Sprintf(
					"%s doesn't implement %s",
					xt.String(),
					nt.String())
				m.Panic(typedString(ex))
				return
			}
			// keep xv as is.
			// *xv = *xv
		} else {
			panic("should not happen")
		}
	} else { // is concrete assert
		tid := t.TypeID()
		xtid := xt.TypeID()
		// assert that x is of type.
		same := tid == xtid
		if !same {
			// TODO: default panic type?
			ex := fmt.Sprintf(
				"%s is not of type %s",
				xt.String(),
				t.String())
			m.Panic(typedString(ex))
			return
		}
		// keep cxt as is.
		// *xv = *xv
	}
}

// NOTE: keep in sync with doOpTypeAssert1.
func (m *Machine) doOpTypeAssert2() {
	m.PopExpr()
	// peek type for re-use
	tv := m.PeekValue(1)
	t := tv.GetType()
	// peek x for re-use
	xv := m.PeekValue(2)
	xt := xv.T

	if t.Kind() == InterfaceKind { // is interface assert
		if it, ok := baseOf(t).(*InterfaceType); ok {
			// t is Gno interface.
			// assert that x implements type.
			impl := false
			impl = it.IsImplementedBy(xt)
			if impl {
				// *xv = *xv
				*tv = untypedBool(true)
			} else {
				// NOTE: consider ability to push an
				// interface-restricted form
				*xv = TypedValue{}
				*tv = untypedBool(false)
			}
		} else if nt, ok := baseOf(t).(*NativeType); ok {
			// t is Go interface.
			// assert that x implements type.
			impl := false
			if nxt, ok := xt.(*NativeType); ok {
				impl = nxt.Type.Implements(nt.Type)
			} else {
				impl = false
			}
			if impl {
				// *xv = *xv
				*tv = untypedBool(true)
			} else {
				*xv = TypedValue{}
				*tv = untypedBool(false)
			}
		} else {
			panic("should not happen")
		}
	} else { // is concrete assert
		tid := t.TypeID()
		xtid := xt.TypeID()
		// assert that x is of type.
		same := tid == xtid
		if same {
			// *xv = *xv
			*tv = untypedBool(true)
		} else {
			*xv = TypedValue{
				T: t,
				V: defaultValue(m.Alloc, t),
			}
			*tv = untypedBool(false)
		}
	}
}

func (m *Machine) doOpCompositeLit() {
	// composite lit expr
	x := m.PeekExpr(1).(*CompositeLitExpr)
	// composite type
	t := m.PeekValue(1).V.(TypeValue).Type
	// push elements
	switch bt := baseOf(t).(type) {
	case *ArrayType:
		m.PushOp(OpArrayLit)
		// evaluate item values
		for i := len(x.Elts) - 1; 0 <= i; i-- {
			m.PushExpr(x.Elts[i].Value)
			m.PushOp(OpEval)
		}
	case *SliceType:
		if len(x.Elts) > 0 && x.Elts[0].Key != nil {
			m.PushOp(OpSliceLit2)
			// evaluate item values
			for i := len(x.Elts) - 1; 0 <= i; i-- {
				if x.Elts[i].Key == nil {
					panic("slice composite literal cannot mix keyed and unkeyed elements")
				}
				m.PushExpr(x.Elts[i].Value)
				m.PushOp(OpEval)
				m.PushExpr(x.Elts[i].Key)
				m.PushOp(OpEval)
			}
		} else {
			m.PushOp(OpSliceLit)
			// evaluate item values
			for i := len(x.Elts) - 1; 0 <= i; i-- {
				if x.Elts[i].Key != nil {
					panic("slice composite literal cannot mix keyed and unkeyed elements")
				}
				m.PushExpr(x.Elts[i].Value)
				m.PushOp(OpEval)
			}
		}
	case *MapType:
		m.PushOp(OpMapLit)
		// evaluate map items
		for i := len(x.Elts) - 1; 0 <= i; i-- {
			// evaluate map value
			m.PushExpr(x.Elts[i].Value)
			m.PushOp(OpEval)
			// evaluate map key
			m.PushExpr(x.Elts[i].Key)
			m.PushOp(OpEval)
		}
	case *StructType:
		m.PushOp(OpStructLit)
		// evaluate field values
		for i := len(x.Elts) - 1; 0 <= i; i-- {
			m.PushExpr(x.Elts[i].Value)
			m.PushOp(OpEval)
		}
	case *NativeType:
		switch bt.Type.Kind() {
		case reflect.Array:
			m.PushOp(OpArrayLitGoNative)
			// evaluate item values
			for i := len(x.Elts) - 1; 0 <= i; i-- {
				m.PushExpr(x.Elts[i].Value)
				m.PushOp(OpEval)
			}
		case reflect.Slice:
			m.PushOp(OpSliceLitGoNative)
			// evaluate item values
			for i := len(x.Elts) - 1; 0 <= i; i-- {
				m.PushExpr(x.Elts[i].Value)
				m.PushOp(OpEval)
			}
		case reflect.Struct:
			m.PushOp(OpStructLitGoNative)
			// evaluate field values
			for i := len(x.Elts) - 1; 0 <= i; i-- {
				m.PushExpr(x.Elts[i].Value)
				m.PushOp(OpEval)
			}
		default:
			panic(fmt.Sprintf(
				"composite lit for native %v kind not yet supported",
				bt.Type.Kind()))
		}
	default:
		panic("not yet implemented")
	}
}

func (m *Machine) doOpArrayLit() {
	// assess performance TODO
	x := m.PopExpr().(*CompositeLitExpr)
	ne := len(x.Elts)
	// peek array type.
	at := m.PeekValue(1 + ne).V.(TypeValue).Type
	bt := baseOf(at).(*ArrayType)
	// construct array value.
	av := defaultArrayValue(m.Alloc, bt)
	if 0 < ne {
		al, ad := av.List, av.Data
		vs := m.PopValues(ne)
		set := make([]bool, bt.Len)
		idx := 0
		for i, v := range vs {
			if kx := x.Elts[i].Key; kx != nil {
				// XXX why convert?
				k := kx.(*ConstExpr).ConvertGetInt()
				if set[k] {
					// array index has already been assigned
					panic(fmt.Sprintf("duplicate index %d in array or slice literal", k))
				}
				set[k] = true
				if al == nil {
					ad[k] = v.GetUint8()
				} else {
					al[k] = v
				}
				idx = k + 1
			} else {
				if set[idx] {
					// array index has already been assigned
					panic(fmt.Sprintf("duplicate index %d in array or slice literal", idx))
				}
				set[idx] = true
				if al == nil {
					ad[idx] = v.GetUint8()
				} else {
					al[idx] = v
				}
				idx++
			}
		}
	}
	// pop array type.
	if debug {
		if m.PopValue().V.(TypeValue).Type != at {
			panic("should not happen")
		}
	} else {
		m.PopValue()
	}
	// push value
	m.PushValue(TypedValue{
		T: at,
		V: av,
	})
}

func (m *Machine) doOpSliceLit() {
	// assess performance TODO
	x := m.PopExpr().(*CompositeLitExpr)
	el := len(x.Elts)
	// peek slice type.
	st := m.PeekValue(1 + el).V.(TypeValue).Type
	// construct element buf slice.
	es := make([]TypedValue, el)
	for i := el - 1; 0 <= i; i-- {
		es[i] = *m.PopValue()
	}
	// construct and push value.
	if debug {
		if m.PopValue().V.(TypeValue).Type != st {
			panic("should not happen")
		}
	} else {
		m.PopValue()
	}
	sv := m.Alloc.NewSliceFromList(es)
	m.PushValue(TypedValue{
		T: st,
		V: sv,
	})
}

func (m *Machine) doOpSliceLit2() {
	// assess performance TODO
	x := m.PopExpr().(*CompositeLitExpr)
	el := len(x.Elts)
	tvs := m.PopValues(el * 2)
	// peek slice type.
	st := m.PeekValue(1).V.(TypeValue).Type
	// calculate maximum index.
	max := 0
	for i := 0; i < el; i++ {
		itv := tvs[i*2+0]
		idx := itv.ConvertGetInt()
		if idx > max {
			max = idx
		}
	}
	// construct element buf slice.
	es := make([]TypedValue, max+1)
	for i := 0; i < el; i++ {
		itv := tvs[i*2+0]
		vtv := tvs[i*2+1]
		idx := itv.ConvertGetInt()
		if es[idx].IsDefined() {
			// slice index has already been assigned
			panic(fmt.Sprintf("duplicate index %d in array or slice literal", idx))
		}
		es[idx] = vtv
	}
	// fill in empty values.
	ste := st.Elem()
	for i, etv := range es {
		if etv.IsUndefined() {
			es[i] = defaultTypedValue(m.Alloc, ste)
		}
	}
	// construct and push value.
	if debug {
		if m.PopValue().V.(TypeValue).Type != st {
			panic("should not happen")
		}
	} else {
		m.PopValue()
	}
	sv := m.Alloc.NewSliceFromList(es)
	m.PushValue(TypedValue{
		T: st,
		V: sv,
	})
}

func (m *Machine) doOpMapLit() {
	x := m.PopExpr().(*CompositeLitExpr)
	ne := len(x.Elts)
	// peek map type.
	mt := m.PeekValue(1 + ne*2).V.(TypeValue).Type
	// bt := baseOf(at).(*MapType)
	// construct new map value.
	mv := m.Alloc.NewMap(0)
	if 0 < ne {
		kvs := m.PopValues(ne * 2)
		// TODO: future optimization
		// omitType := baseOf(mt).Elem().Kind() != InterfaceKind
		for i := 0; i < ne; i++ {
			ktv := &kvs[i*2]
			vtv := kvs[i*2+1]
			ptr := mv.GetPointerForKey(m.Alloc, m.Store, ktv)
			if ptr.TV.IsDefined() {
				// map key has already been assigned
				panic(fmt.Sprintf("duplicate key %s in map literal", ktv.V))
			}
			*ptr.TV = vtv
		}
	}
	// pop map type.
	if debug {
		if m.PopValue().GetType() != mt {
			panic("should not happen")
		}
	} else {
		m.PopValue()
	}
	// push value
	m.PushValue(TypedValue{
		T: mt,
		V: mv,
	})
}

func (m *Machine) doOpStructLit() {
	// assess performance TODO
	x := m.PopExpr().(*CompositeLitExpr)
	el := len(x.Elts) // may be incomplete
	// peek struct type.
	xt := m.PeekValue(1 + el).V.(TypeValue).Type
	st := baseOf(xt).(*StructType)
	nf := len(st.Fields)
	fs := []TypedValue(nil)
	// NOTE includes embedded fields.
	if el == 0 {
		// zero struct with no fields set.
		// TODO: optimize and allow nil.
		fs = defaultStructFields(m.Alloc, st)
	} else if x.Elts[0].Key == nil {
		// field values are in order.
		m.Alloc.AllocateStructFields(int64(len(st.Fields)))
		fs = make([]TypedValue, 0, len(st.Fields))
		if debug {
			if el == 0 {
				// this is fine.
			} else if el != nf {
				panic("Unnamed composite literals must have exact number of fields")
			} else {
				// If there are any unexported fields and the
				// package doesn't match, we cannot use this
				// method to initialize the struct.
				if FieldTypeList(st.Fields).HasUnexported() &&
					st.PkgPath != m.Package.PkgPath {
					panic(fmt.Sprintf(
						"Cannot initialize imported struct %s.%s with nameless composite lit expression (has unexported fields) from package %s",
						st.PkgPath, st.String(), m.Package.PkgPath))
				} else {
					// this is fine.
				}
			}
		}
		ftvs := m.PopValues(el)
		for _, ftv := range ftvs {
			if debug {
				if !ftv.IsUndefined() && ftv.T.Kind() == InterfaceKind {
					panic("should not happen")
				}
			}
			fs = append(fs, ftv)
		}
		if debug {
			if len(fs) != cap(fs) {
				panic("should not happen")
			}
		}
	} else {
		// field values are by name and may be out of order.
		fs = defaultStructFields(m.Alloc, st)
		fsset := make([]bool, len(fs))
		ftvs := m.PopValues(el)
		for i := 0; i < el; i++ {
			fnx := x.Elts[i].Key.(*NameExpr)
			ftv := ftvs[i]
			if debug {
				if fnx.Path.Depth != 0 {
					panic("unexpected struct composite lit key path generation value")
				}
				if !ftv.IsUndefined() && ftv.T.Kind() == InterfaceKind {
					panic("should not happen")
				}
			}
			if fsset[fnx.Path.Index] {
				// already set
				panic(fmt.Sprintf("duplicate field name %s in struct literal", fnx.Name))
			}
			fsset[fnx.Path.Index] = true
			fs[fnx.Path.Index] = ftv
		}
	}
	// construct and push value.
	m.PopValue() // baseOf() is st
	sv := m.Alloc.NewStruct(fs)
	m.PushValue(TypedValue{
		T: xt,
		V: sv,
	})
}

func (m *Machine) doOpFuncLit() {
	x := m.PopExpr().(*FuncLitExpr)
	ft := m.PopValue().V.(TypeValue).Type.(*FuncType)
	lb := m.LastBlock()
	m.Alloc.AllocateFunc()
	m.PushValue(TypedValue{
		T: ft,
		V: &FuncValue{
			Type:       ft,
			IsMethod:   false,
			Source:     x,
			Name:       "",
			Closure:    lb,
			PkgPath:    m.Package.PkgPath,
			body:       x.Body,
			nativeBody: nil,
		},
	})
}

func (m *Machine) doOpConvert() {
	xv := m.PopValue()
	t := m.PopValue().GetType()
	ConvertTo(m.Alloc, m.Store, xv, t)
	m.PushValue(*xv)
}
