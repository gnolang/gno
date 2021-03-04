package gno

import (
	"fmt"
	"reflect"
)

func (m *Machine) doOpFieldType() {
	x := m.PopExpr().(*FieldTypeExpr)
	t := m.PopValue().V.(TypeValue).Type
	n := x.Name
	tag := Tag("")
	if x.Tag != nil {
		tag = Tag(m.PopValue().GetString())
	}
	ft := FieldType{
		Name: n,
		Type: t,
		Tag:  tag,
	}
	m.PushValue(TypedValue{
		T: gTypeType,
		V: TypeValue{Type: ft},
	})
}

func (m *Machine) doOpArrayType() {
	x := m.PopExpr().(*ArrayTypeExpr)
	var t = &ArrayType{}
	if x.Len == nil { // variadic array lit
		t.Vrd = true
	} else {
		lv := m.PopValue()
		if debug {
			// This is how run-time untyped const
			// conversions would work, but we
			// expect the preprocessor to convert
			// these to *constExpr.
			// Numeric untyped types are always constant.
			/*
				// Convert if untyped.
				if isUntyped(lv.T) {
					ConvertUntypedTo(lv, IntType)
				}
			*/
			if isUntyped(lv.T) {
				panic("unexpected untyped const type for array type len during runtime")
			}
		}
		t.Len = lv.GetInt() // TODO lazy convert?
	}
	tv := m.PeekValue(1) // re-use
	t.Elt = tv.GetType()
	*tv = TypedValue{
		T: gTypeType,
		V: TypeValue{Type: t},
	}
}

func (m *Machine) doOpSliceType() {
	x := m.PopExpr().(*SliceTypeExpr)
	tv := m.PeekValue(1) // re-use as result.
	t := &SliceType{
		Elt: tv.GetType(),
		Vrd: x.Vrd,
	}
	*tv = TypedValue{
		T: gTypeType,
		V: TypeValue{Type: t},
	}
}

func (m *Machine) doOpFuncType() {
	x := m.PopExpr().(*FuncTypeExpr)
	// Allocate space for data.
	params := make([]FieldType, len(x.Params))
	results := make([]FieldType, len(x.Results))
	// Pop results.
	for i := len(x.Results) - 1; 0 <= i; i-- {
		results[i] = m.PopValue().V.(TypeValue).Type.(FieldType)
	}
	// Pop params.
	for i := len(x.Params) - 1; 0 <= i; i-- {
		params[i] = m.PopValue().V.(TypeValue).Type.(FieldType)
	}
	// Push func type.
	ft := &FuncType{
		PkgPath: m.Package.PkgPath,
		Params:  params,
		Results: results,
	}
	m.PushValue(TypedValue{
		T: gTypeType,
		V: TypeValue{Type: ft},
	})
}

func (m *Machine) doOpMapType() {
	vtv := m.PopValue()
	tv := m.PeekValue(1) // re-use as result.
	mt := &MapType{
		Key:   tv.GetType(),
		Value: vtv.GetType(),
	}
	*tv = TypedValue{
		T: gTypeType,
		V: TypeValue{Type: mt},
	}
}

func (m *Machine) doOpStructType() {
	x := m.PopExpr().(*StructTypeExpr)
	// pop fields
	ftvs := m.PopValues(len(x.Fields))
	// allocate (minimum) space for flat fields
	ffields := make([]FieldType, 0, len(x.Fields))
	mapping := make([]int, len(x.Fields))
	// populate ffields
	for i, ftv := range ftvs {
		mapping[i] = len(ffields)
		ft := ftv.V.(TypeValue).Type.(FieldType)
		fillEmbeddedName(&ft)
		ffields = append(ffields, ft)
		if ftv.T.Kind() == StructKind { // flatten
			st := baseOf(ft.Type).(*StructType)
			ffields = append(ffields, st.Fields...)
		}
	}
	// push struct type
	st := &StructType{
		PkgPath: m.Package.PkgPath,
		Fields:  ffields,
		Mapping: mapping,
	}
	m.PushValue(TypedValue{
		T: gTypeType,
		V: TypeValue{Type: st},
	})
}

func (m *Machine) doOpInterfaceType() {
	x := m.PopExpr().(*InterfaceTypeExpr)
	// allocate space
	methods := make([]FieldType, len(x.Methods))
	// pop methods
	for i := len(x.Methods) - 1; 0 <= i; i-- {
		ft := m.PopValue().V.(TypeValue).Type.(FieldType)
		fillEmbeddedName(&ft)
		methods[i] = ft
	}
	// push interface type
	it := &InterfaceType{
		PkgPath:   m.Package.PkgPath,
		Methods:   methods,
		IsUntyped: x.IsUntyped,
	}
	m.PushValue(TypedValue{
		T: gTypeType,
		V: TypeValue{Type: it},
	})
}

// Evaluate the type of a typed (i.e. not untyped) value.
// This function expects const expressions to have been
// already swapped for *constExpr in the preprocessor.  If not, panics.
func (m *Machine) doOpTypeOf() {
	x := m.PopExpr()
	switch x := x.(type) {
	case *NameExpr:
		// NOTE: duplicated from doOpEval
		if x.Path.Depth == 0 {
			// Name is in uverse (global).
			gv := Uverse().GetPointerTo(x.Path)
			m.PushValue(asValue(gv.T))
		} else {
			// Get value from scope.
			lb := m.LastBlock()
			tv := lb.GetPointerTo(x.Path)
			m.PushValue(asValue(tv.T))
		}
	case *BasicLitExpr:
		// Should already be swapped for *constExpr.
		// Also, this isn't needed.
		panic("*BasicLitExpr not supported with OpTypeOf")
	case *BinaryExpr:
		switch x.Op {
		case ADD, SUB, MUL, QUO, REM, BAND, BOR, XOR,
			SHL, SHR, BAND_NOT, LAND, LOR:
			m.PushExpr(x.Left)
			m.PushOp(OpTypeOf)
		case EQL, LSS, GTR, NEQ, LEQ, GEQ:
			m.PushValue(asValue(UntypedBoolType))
		}
	case *CallExpr:
		start := m.NumValues
		m.PushOp(OpHalt)
		m.PushExpr(x.Func)
		m.PushOp(OpTypeOf)
		m.Run() // XXX replace
		t := m.ReapValues(start)[0].GetType()
		switch bft := baseOf(t).(type) {
		case *FuncType:
			rs := bft.Results
			if len(rs) != 1 {
				panic(fmt.Sprintf(
					"cannot get type of function call with %d results",
					len(rs)))
			}
			m.PushValue(asValue(rs[0].Type))
		case *TypeType:
			start := m.NumValues
			m.PushOp(OpHalt)
			m.PushExpr(x.Func)
			m.PushOp(OpEval)
			m.Run() // XXX replace
			t := m.ReapValues(start)[0].GetType()
			m.PushValue(asValue(t))
		case *nativeType:
			numRes := bft.Type.NumOut()
			if numRes != 1 {
				panic(fmt.Sprintf(
					"cannot get type of (native) function call with %d results",
					numRes))
			}
			res0 := bft.Type.Out(0)
			m.PushValue(asValue(&nativeType{Type: res0}))
		default:
			panic(fmt.Sprintf(
				"unexpected call of expression type %s",
				t.String()))
		}
	case *IndexExpr:
		start := m.NumValues
		m.PushOp(OpHalt)
		m.PushExpr(x.X)
		m.PushOp(OpTypeOf)
		m.Run() // XXX replace
		xt := m.ReapValues(start)[0].GetType()
		m.PushValue(asValue(xt.Elem()))
	case *SelectorExpr:
		start := m.NumValues
		m.PushOp(OpHalt)
		m.PushExpr(x.X)
		m.PushOp(OpTypeOf)
		m.Run() // XXX replace
		xt := m.ReapValues(start)[0].GetType()
		path := x.Path
	TYPE_SWITCH:
		switch ct := xt.(type) {
		case *DeclaredType:
			if path.Depth <= 1 {
				ftv := ct.GetValueRefAt(path)
				ft := ftv.T.(*FuncType)
				t := ft.BoundType()
				m.PushValue(asValue(t))
			} else {
				xt = ct.Base
				goto TYPE_SWITCH
			}
		case PointerType:
			if dt, ok := ct.Elt.(*DeclaredType); ok {
				xt = dt
				goto TYPE_SWITCH
			} else {
				panic("should not happen")
			}
		case *StructType:
			for _, ft := range ct.Fields {
				if ft.Name == x.Sel {
					m.PushValue(asValue(ft.Type))
					return
				}
			}
			panic(fmt.Sprintf("struct type %v has no field %s",
				reflect.TypeOf(baseOf(xt)), x.Sel))
		case *TypeType:
			start := m.NumValues
			m.PushOp(OpHalt)
			m.PushExpr(x.X)
			m.PushOp(OpEval)
			m.Run() // XXX replace
			xv := m.ReapValues(start)[0]
			switch t := xv.GetType().(type) {
			case *DeclaredType:
				t2 := t.GetValueRefAt(path).T
				m.PushValue(asValue(t2))
				return
			case *nativeType:
				rt := t.Type
				mt, ok := rt.MethodByName(string(x.Sel))
				if !ok {
					if debug {
						panic(fmt.Sprintf(
							"native type %s has no method %s",
							rt.String(), x.Sel))
					}
					panic("unknown native method selector")
				}
				t2 := go2GnoType(mt.Type)
				m.PushValue(asValue(t2))
				return
			default:
				panic("unexpected selector base typeval.")
			}
		case *PackageType:
			start := m.NumValues
			m.PushOp(OpHalt)
			m.PushExpr(x.X)
			m.PushOp(OpEval)
			m.Run() // XXX replace
			xv := m.ReapValues(start)[0]
			pv := xv.V.(*PackageValue)
			t := pv.Source.GetStaticTypeOfAt(x.Path)
			m.PushValue(asValue(t))
			return
		case *nativeType:
			rt := ct.Type
			// switch on type and maybe match field.
			if rt.Kind() == reflect.Ptr {
				ert := rt.Elem()
				rft, ok := ert.FieldByName(string(x.Sel))
				if ok {
					ft := go2GnoType(rft.Type)
					m.PushValue(asValue(ft))
					return
				}
				// keep rt as is.
			} else if rt.Kind() == reflect.Interface {
				// keep rt as is.
			} else if rt.Kind() == reflect.Struct {
				rft, ok := rt.FieldByName(string(x.Sel))
				if ok {
					ft := go2GnoType(rft.Type)
					m.PushValue(asValue(ft))
					return
				}
				// make rt ptr.
				rt = reflect.PtrTo(rt)
			} else {
				// make rt ptr.
				rt = reflect.PtrTo(rt)
			}
			// match method.
			rmt, ok := rt.MethodByName(string(x.Sel))
			if ok {
				mt := go2GnoFuncType(rmt.Type)
				if rt.Kind() == reflect.Interface {
					m.PushValue(asValue(mt))
					return
				} else {
					bmt := mt.BoundType()
					m.PushValue(asValue(bmt))
					return
				}
			}
			panic(fmt.Sprintf(
				"native type %s has no method or field %s",
				ct.String(), x.Sel))
		default:
			panic(fmt.Sprintf("selector expression invalid for type %v",
				reflect.TypeOf(baseOf(xt))))
		}
	case *SliceExpr:
		start := m.NumValues
		m.PushOp(OpHalt)
		m.PushExpr(x.X)
		m.PushOp(OpTypeOf)
		m.Run() // XXX replace
		xt := m.ReapValues(start)[0].V.(TypeValue).Type
		m.PushValue(asValue(&SliceType{
			Elt: xt.Elem(),
		}))
	case *StarExpr:
		start := m.NumValues
		m.PushOp(OpHalt)
		m.PushExpr(x.X)
		m.PushOp(OpTypeOf)
		m.Run() // XXX replace
		xt := m.ReapValues(start)[0].GetType().(PointerType)
		m.PushValue(asValue(xt.Elt))
	case *RefExpr:
		start := m.NumValues
		m.PushOp(OpHalt)
		m.PushExpr(x.X)
		m.PushOp(OpTypeOf)
		m.Run() // XXX replace
		xt := m.ReapValues(start)[0].GetType()
		m.PushValue(asValue(PointerType{Elt: xt}))
	case *TypeAssertExpr:
		if x.HasOK {
			panic("type assert assignment used with return 2 values; has no type")
		} else {
			m.PushExpr(x.Type)
			m.PushOp(OpEval)
		}
	case *UnaryExpr:
		m.PushExpr(x.X)
		m.PushOp(OpTypeOf)
	case *CompositeLitExpr:
		m.PushExpr(x.Type)
		m.PushOp(OpEval)
	case *FuncLitExpr:
		m.PushExpr(&x.Type)
		m.PushOp(OpEval)
	case TypeExpr:
		m.PushValue(asValue(gTypeType))
	case *constExpr:
		m.PushValue(asValue(x.T))
	default:
		panic(fmt.Sprintf(
			"unexpected expression of type %v",
			reflect.TypeOf(x)))
	}
}
