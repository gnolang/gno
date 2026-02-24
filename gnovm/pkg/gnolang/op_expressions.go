package gnolang

import (
	"fmt"
)

// OpBinary1 defined in op_binary.go

// NOTE: keep in sync with doOpIndex2.
func (m *Machine) doOpIndex1() {
	m.PopExpr()
	iv := m.PopValue()   // index
	xv := m.PeekValue(1) // x
	ro := m.IsReadonly(xv)
	switch ct := baseOf(xv.T).(type) {
	case *MapType:
		vt := ct.Value
		if xv.V == nil { // uninitialized map
			*xv = defaultTypedValue(m.Alloc, vt) // reuse as result
		} else {
			mv := xv.V.(*MapValue)
			vv, exists := mv.GetValueForKey(m.Store, iv)
			if exists {
				*xv = vv // reuse as result
			} else {
				*xv = defaultTypedValue(m.Alloc, vt) // reuse as result
			}
		}
	default:
		// NOTE: nilRealm is OK, not setting a map (w/ new key).
		res := xv.GetPointerAtIndex(nilRealm, m.Alloc, m.Store, iv)
		*xv = res.Deref() // reuse as result
	}
	xv.SetReadonly(ro)
}

// NOTE: keep in sync with doOpIndex1.
func (m *Machine) doOpIndex2() {
	m.PopExpr()
	iv := m.PeekValue(1) // index
	xv := m.PeekValue(2) // x
	ro := m.IsReadonly(xv)
	switch ct := baseOf(xv.T).(type) {
	case *MapType:
		vt := ct.Value
		if xv.V == nil { // uninitialized map
			*xv = defaultTypedValue(m.Alloc, vt) // reuse as result
			*iv = untypedBool(false)             // reuse as result
		} else {
			mv := xv.V.(*MapValue)
			vv, exists := mv.GetValueForKey(m.Store, iv)
			if exists {
				*xv = vv                // reuse as result
				*iv = untypedBool(true) // reuse as result
			} else {
				*xv = defaultTypedValue(m.Alloc, vt) // reuse as result
				*iv = untypedBool(false)             // reuse as result
			}
		}
	default:
		panic("should not happen")
	}
	xv.SetReadonly(ro)
}

func (m *Machine) doOpSelector() {
	sx := m.PopExpr().(*SelectorExpr)
	xv := m.PeekValue(1) // the base .X -- package, struct, etc.
	ro := m.IsReadonly(xv)
	res := xv.GetPointerToFromTV(m.Alloc, m.Store, sx.Path).Deref()
	if debug {
		m.Printf("-v[S] %v\n", xv)
		m.Printf("+v[S] %v\n", res)
	}
	*xv = res // reuse as result
	xv.SetReadonly(ro)
}

func (m *Machine) doOpSlice() {
	sx := m.PopExpr().(*SliceExpr)
	lowVal, highVal, maxVal := -1, -1, -1
	// max
	if sx.Max != nil {
		maxVal = int(m.PopValue().ConvertGetInt())
	}
	// high
	if sx.High != nil {
		highVal = int(m.PopValue().ConvertGetInt())
	}
	// low
	if sx.Low != nil {
		lowVal = int(m.PopValue().ConvertGetInt())
	} else {
		lowVal = 0
	}
	// slice base x
	xv := m.PopValue()
	ro := m.IsReadonly(xv)
	// if a is a pointer to an array, a[low : high : max] is
	// shorthand for (*a)[low : high : max]
	// XXX fix this in precompile instead.
	if xv.T.Kind() == PointerKind &&
		xv.T.Elem().Kind() == ArrayKind {
		if xv.V == nil {
			m.pushPanic(typedString("nil pointer dereference"))
			return
		}
		// simply deref xv.
		*xv = xv.V.(PointerValue).Deref()
		// check array also for ro.
		if !ro {
			ro = xv.IsReadonly()
		}
	}
	// fill default based on xv
	if sx.High == nil {
		highVal = xv.GetLength()
	}
	// all low:high:max cases
	if maxVal == -1 {
		sv := xv.GetSlice(m.Alloc, lowVal, highVal)
		m.PushValue(sv.WithReadonly(ro))
	} else {
		sv := xv.GetSlice2(m.Alloc, lowVal, highVal, maxVal)
		m.PushValue(sv.WithReadonly(ro))
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
		if xv.V == nil {
			m.pushPanic(typedString("nil pointer dereference"))
			return
		}

		pv := xv.V.(PointerValue)
		if pv.TV.T == DataByteType {
			tv := TypedValue{T: bt.Elt}
			dbv := pv.TV.V.(DataByteValue)
			tv.SetUint8(dbv.GetByte())
			m.PushValue(tv)
		} else {
			ro := m.IsReadonly(xv)
			pvtv := (*pv.TV).WithReadonly(ro)
			if xpt, ok := baseOf(xv.T).(*PointerType); ok {
				// e.g. type Foo; type Bar;
				// *((*Foo)(&Bar{})) should be Bar, not Foo.
				pvtv.T = xpt.Elem()
			}
			m.PushValue(pvtv)
		}
	case *TypeType:
		t := xv.GetType()
		pt := &PointerType{Elt: t}
		m.PushValue(asValue(pt))
	default:
		panic(fmt.Sprintf(
			"illegal star expression x type %s",
			xv.T.String()))
	}
}

// XXX this is wrong, for var i interface{}; &i is *interface{}.
func (m *Machine) doOpRef() {
	rx := m.PopExpr().(*RefExpr)
	xv, ro := m.PopAsPointer2(rx.X)
	elt := xv.TV.T
	if elt == DataByteType {
		elt = xv.TV.V.(DataByteValue).ElemType
	}
	m.Alloc.AllocatePointer()
	m.PushValue(TypedValue{
		T: m.Alloc.NewType(&PointerType{Elt: elt}),
		V: xv,
	}.WithReadonly(ro))
}

// NOTE: keep in sync with doOpTypeAssert2.
func (m *Machine) doOpTypeAssert1() {
	m.PopExpr()
	// pop type
	t := m.PopValue().GetType() // type being asserted

	// peek x for re-use
	xv := m.PeekValue(1) // value result / value to assert
	xt := xv.T           // underlying value's type

	// xt may be nil, but we need to wait to return because the value of xt that is set
	// will depend on whether we are trying to assert to an interface or concrete type.
	// xt can be nil in the case where recover can't find a panic to recover from and
	// returns a bare TypedValue{}.

	if t.Kind() == InterfaceKind { // is interface assert
		if xt == nil || xv.IsNilInterface() {
			// TODO: default panic type?
			ex := fmt.Sprintf("interface conversion: interface is nil, not %s", t.String())
			m.pushPanic(typedString(ex))
			return
		}

		if it, ok := baseOf(t).(*InterfaceType); ok {
			// An interface type assertion on a value that doesn't have a concrete base
			// type should always fail.
			if _, ok := baseOf(xt).(*InterfaceType); ok {
				// TODO: default panic type?
				ex := fmt.Sprintf(
					"non-concrete %s doesn't implement %s",
					xt.String(),
					it.String())
				m.pushPanic(typedString(ex))
				return
			}

			// t is Gno interface.
			// assert that x implements type.
			err := it.VerifyImplementedBy(xt)
			if err != nil {
				// TODO: default panic type?
				ex := fmt.Sprintf(
					"%s doesn't implement %s (%s)",
					xt.String(),
					it.String(),
					err.Error())
				m.pushPanic(typedString(ex))
				return
			}
			// NOTE: consider ability to push an
			// interface-restricted form
			// *xv = *xv
		} else {
			panic("should not happen")
		}
	} else { // is concrete assert
		if xt == nil {
			ex := fmt.Sprintf("nil is not of type %s", t.String())
			m.pushPanic(typedString(ex))
			return
		}

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
			m.pushPanic(typedString(ex))
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
	tv := m.PeekValue(1) // boolean result
	t := tv.GetType()    // type being asserted

	// peek x for re-use
	xv := m.PeekValue(2) // value result / value to assert
	xt := xv.T           // underlying value's type

	// xt may be nil, but we need to wait to return because the value of xt that is set
	// will depend on whether we are trying to assert to an interface or concrete type.
	// xt can be nil in the case where recover can't find a panic to recover from and
	// returns a bare TypedValue{}.

	if t.Kind() == InterfaceKind { // is interface assert
		if xt == nil {
			*xv = TypedValue{}
			*tv = untypedBool(false)
			return
		}

		if it, ok := baseOf(t).(*InterfaceType); ok {
			// An interface type assertion on a value that doesn't have a concrete base
			// type should always fail.
			if _, ok := baseOf(xt).(*InterfaceType); ok {
				*xv = TypedValue{}
				*tv = untypedBool(false)
				return
			}

			// t is Gno interface.
			// assert that x implements type.
			impl := it.IsImplementedBy(xt)
			if impl {
				// *xv = *xv
				*tv = untypedBool(true)
			} else {
				// NOTE: consider ability to push an
				// interface-restricted form
				*xv = TypedValue{}
				*tv = untypedBool(false)
			}
		} else {
			panic("should not happen")
		}
	} else { // is concrete assert
		if xt == nil {
			*xv = defaultTypedValue(m.Alloc, t)
			*tv = untypedBool(false)
			return
		}

		tid := t.TypeID()
		xtid := xt.TypeID()
		// assert that x is of type.
		same := tid == xtid

		if same {
			// *xv = *xv
			*tv = untypedBool(true)
		} else {
			*xv = defaultTypedValue(m.Alloc, t)
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
	switch baseOf(t).(type) {
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
		var idx int64
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
					al[k] = v.Copy(m.Alloc)
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
					al[idx] = v.Copy(m.Alloc)
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
	baseArray := m.Alloc.NewListArray(el)
	es := baseArray.List
	m.PopCopyValues(es)
	// construct and push value.
	if debug {
		if m.PopValue().V.(TypeValue).Type != st {
			panic("should not happen")
		}
	} else {
		m.PopValue()
	}
	sv := m.Alloc.NewSlice(baseArray, 0, el, el)
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
	var maxVal int64
	for i := range el {
		itv := tvs[i*2+0]
		idx := itv.ConvertGetInt()
		if idx > maxVal {
			maxVal = idx
		}
	}
	// construct element buf slice.
	// alloc before the underlying array constructed
	baseArray := m.Alloc.NewListArray(int(maxVal + 1))
	es := baseArray.List

	for i := range el {
		itv := tvs[i*2+0]
		vtv := tvs[i*2+1]
		idx := itv.ConvertGetInt()
		if es[idx].IsDefined() {
			// slice index has already been assigned
			panic(fmt.Sprintf("duplicate index %d in array or slice literal", idx))
		}
		es[idx] = vtv.Copy(m.Alloc)
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
	sv := m.Alloc.NewSlice(baseArray, 0, int(maxVal+1), int(maxVal+1))
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
		for i := range ne {
			ktv := kvs[i*2].Copy(m.Alloc)
			vtv := kvs[i*2+1]
			ptr := mv.GetPointerForKey(m.Alloc, m.Store, ktv)
			*ptr.TV = vtv.Copy(m.Alloc)
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
		fs = make([]TypedValue, len(st.Fields))
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
				}
				// else, this is fine.
			}
		}
		m.PopCopyValues(fs)
	} else {
		// field values are by name and may be out of order.
		fs = defaultStructFields(m.Alloc, st)
		fsset := make([]bool, len(fs))
		ftvs := m.PopValues(el)
		for i := range el {
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
			fs[fnx.Path.Index] = ftv.Copy(m.Alloc)
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

	// First copy closure captured heap values
	// to *FuncValue. Later during doOpCall a block
	// will be created that copies these values for
	// every invocation of the function.
	captures := make([]TypedValue, 0, len(x.HeapCaptures))
	if m.Stage == StagePre {
		// TODO static block items aren't heap items.
		// continue
	} else {
		for _, nx := range x.HeapCaptures {
			ptr := lb.GetPointerToDirect(m.Store, nx.Path)
			// check that ptr.TV is a heap item value.
			// it must be in the form of:
			// {T:heapItemType{},V:HeapItemValue{...}}
			if _, ok := ptr.TV.T.(heapItemType); !ok {
				panic("should not happen, should be heapItemType: " + nx.String())
			}
			if _, ok := ptr.TV.V.(*HeapItemValue); !ok {
				panic("should not happen, should be heapItemValue: " + nx.String())
			}
			captures = append(captures, *ptr.TV)
		}
	}
	m.PushValue(TypedValue{
		T: ft,
		V: &FuncValue{
			Type:       ft,
			IsMethod:   false,
			IsClosure:  true,
			Source:     x,
			Name:       "",
			Parent:     nil,
			Captures:   captures,
			PkgPath:    m.Package.PkgPath,
			Crossing:   ft.IsCrossing(),
			body:       x.Body,
			nativeBody: nil,
		},
	})
}

func (m *Machine) doOpConvert() {
	xv := m.PopValue().Copy(m.Alloc)
	t := m.PopValue().GetType()

	// BEGIN conversion checks
	// These protect against inter-realm conversion exploits.

	// Case 1.
	// Do not allow conversion of value stored in eternal realm.
	// Otherwise anyone could convert an external object insecurely.
	if xv.T != nil && !xv.T.IsImmutable() && m.IsReadonly(&xv) {
		if xvdt, ok := xv.T.(*DeclaredType); ok &&
			xvdt.PkgPath == m.Realm.Path {
			// Except allow if xv.T is m.Realm.
			// XXX do we need/want this?
		} else {
			panic("illegal conversion of readonly or externally stored value")
		}
	}

	// Case 2.
	// Do not allow conversion to type of external realm.
	// Only code declared within the same realm my perform such
	// conversions, otherwise the realm could be tricked
	// into executing a subtle exploit of mutating some
	// value (say a pointer) stored in its own realm by
	// a hostile construction converted to look safe.
	if tdt, ok := t.(*DeclaredType); ok && !tdt.IsImmutable() && m.Realm != nil {
		if IsRealmPath(tdt.PkgPath) && tdt.PkgPath != m.Realm.Path {
			panic("illegal conversion to external realm type")
		}
	}
	// END conversion checks

	ConvertTo(m.Alloc, m.Store, &xv, t, false)
	m.PushValue(xv)
}
