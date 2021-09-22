package gno

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
	dst := xv
	res := xv.GetPointerAtIndex(m.Store, iv)
	*dst = res.Deref() // reuse as result
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
	vt := xv.T.(*MapType).Value
	if xv.V == nil { // uninitialized map
		*xv = TypedValue{ // reuse as result
			T: vt,
			V: defaultValue(vt),
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
				V: defaultValue(vt),
			}
			*iv = untypedBool(false) // reuse as result
		}
	}
}

func (m *Machine) doOpSelector() {
	sx := m.PopExpr().(*SelectorExpr)
	xv := m.PeekValue(1)
	res := xv.GetPointerTo(m.Store, sx.Path)
	*xv = res.Deref() // reuse as result
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
		sv := xv.GetSlice(low, high)
		m.PushValue(sv)
	} else {
		sv := xv.GetSlice2(low, high, max)
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
			tv := TypedValue{T: xv.T.(*PointerType).Elt}
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
		if nt, ok := t.(*nativeType); ok {
			pt = &nativeType{Type: reflect.PtrTo(nt.Type)}
		} else {
			pt = &PointerType{Elt: t}
		}
		m.PushValue(asValue(pt))
	case *nativeType:
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
	xv := m.PopAsPointer(rx.X)
	if nv, ok := xv.TV.V.(*nativeValue); ok {
		// If a native pointer, ensure it is addressable.  This
		// way, PointerValue{*nativeValue{rv}} can be converted
		// to/from *nativeValue{rv.Addr()}.
		if !nv.Value.CanAddr() {
			rv := nv.Value
			rt := rv.Type()
			rv2 := reflect.New(rt).Elem()
			rv2.Set(rv)
			nv.Value = rv2
		}
	}
	// XXX this is wrong, if rx.X is interface type,
	// XXX then the type should be &PointerType{Elt: staticTypeOf(xv)}
	m.PushValue(TypedValue{
		T: &PointerType{Elt: xv.TV.T},
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
			switch cxt := xt.(type) {
			case *InterfaceType:
				panic("should not happen")
				// impl = it.IsImplementedBy(cxt)
			case *DeclaredType, *PointerType:
				impl = it.IsImplementedBy(cxt)
			default:
				impl = it.IsEmptyInterface()
			}
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
		} else if nt, ok := baseOf(t).(*nativeType); ok {
			// t is Go interface.
			// assert that x implements type.
			impl := false
			if nxt, ok := xt.(*nativeType); ok {
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
			switch cxt := xt.(type) {
			case *InterfaceType:
				panic("should not happen")
				// impl = it.IsImplementedBy(cxt)
			case *DeclaredType:
				impl = it.IsImplementedBy(cxt)
			default:
				impl = it.IsEmptyInterface()
			}
			if impl {
				// *xv = *xv
				*tv = untypedBool(true)
			} else {
				// NOTE: consider ability to push an
				// interface-restricted form
				*xv = TypedValue{}
				*tv = untypedBool(false)
			}
		} else if nt, ok := baseOf(t).(*nativeType); ok {
			// t is Go interface.
			// assert that x implements type.
			impl := false
			if nxt, ok := xt.(*nativeType); ok {
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
				V: defaultValue(t),
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
	switch baseOf(t).(type) {
	case *ArrayType:
		m.PushOp(OpArrayLit)
		// evalaute field values
		for i := len(x.Elts) - 1; 0 <= i; i-- {
			m.PushExpr(x.Elts[i].Value)
			m.PushOp(OpEval)
		}
	case *SliceType:
		m.PushOp(OpSliceLit)
		// evalaute field values
		for i := len(x.Elts) - 1; 0 <= i; i-- {
			if x.Elts[i].Key != nil {
				panic("keys not yet supported in slice composite literals")
			}
			m.PushExpr(x.Elts[i].Value)
			m.PushOp(OpEval)
		}
	case *MapType:
		m.PushOp(OpMapLit)
		// evalaute map items
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
	case *nativeType:
		m.PushOp(OpStructLitGoNative)
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
	av := defaultArrayValue(bt)
	if 0 < ne {
		al := av.List
		vs := m.PopValues(ne)
		idx := 0
		for i, v := range vs {
			if kx := x.Elts[i].Key; kx != nil {
				// XXX why convert?
				k := kx.(*constExpr).ConvertGetInt()
				al[k] = v
				idx = k + 1
			} else {
				al[idx] = v
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
	// peek array type.
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
	sv := newSliceFromList(es)
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
	mv := &MapValue{}
	mv.MakeMap(0)
	if 0 < ne {
		kvs := m.PopValues(ne * 2)
		// TODO: future optimization
		// omitType := baseOf(mt).Elem().Kind() != InterfaceKind
		for i := 0; i < ne; i++ {
			ktv := &kvs[i*2]
			vtv := kvs[i*2+1]
			ptr := mv.GetPointerForKey(m.Store, ktv)
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
	sv := &StructValue{
		// will replace this with fs below.
		Fields: nil, // becomes fs.
	}
	fs := []TypedValue(nil)
	// NOTE includes embedded fields.
	if el == 0 {
		// zero struct with no fields set.
		// TODO: optimize and allow nil.
		fs = defaultStructFields(st)
	} else if x.Elts[0].Key == nil {
		// field values are in order.
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
						"Cannot initialize imported struct %s with nameless composite lit expression (has unexported fields)",
						st.String()))
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
		fs = defaultStructFields(st)
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
			fs[fnx.Path.Index] = ftv
		}
	}
	// construct and push value.
	m.PopValue() // baseOf() is st
	sv.Fields = fs
	m.PushValue(TypedValue{
		T: xt,
		V: sv,
	})
}

func (m *Machine) doOpFuncLit() {
	x := m.PopExpr().(*FuncLitExpr)
	ft := m.PopValue().V.(TypeValue).Type.(*FuncType)
	lb := m.LastBlock()
	m.PushValue(TypedValue{
		T: ft,
		V: &FuncValue{
			Type:       ft,
			IsMethod:   false,
			SourceLoc:  x.GetLocation(),
			Source:     x,
			Name:       "",
			Body:       x.Body,
			Closure:    lb,
			PkgPath:    m.Package.PkgPath,
			nativeBody: nil,
			pkg:        m.Package,
		},
	})
}

func (m *Machine) doOpConvert() {
	xv := m.PopValue()
	t := m.PopValue().GetType()
	ConvertTo(m.Store, xv, t)
	m.PushValue(*xv)
}
