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
	// allocate (minimum) space for fields
	fields := make([]FieldType, 0, len(x.Fields))
	// populate fields
	for _, ftv := range ftvs {
		ft := ftv.V.(TypeValue).Type.(FieldType)
		fillEmbeddedName(&ft)
		fields = append(fields, ft)
	}
	// push struct type
	st := &StructType{
		PkgPath: m.Package.PkgPath,
		Fields:  fields,
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
		PkgPath: m.Package.PkgPath,
		Methods: methods,
		Generic: x.Generic,
	}
	m.PushValue(TypedValue{
		T: gTypeType,
		V: TypeValue{Type: it},
	})
}

func (m *Machine) doOpChanType() {
	x := m.PopExpr().(*ChanTypeExpr)
	tv := m.PeekValue(1) // re-use as result.
	ct := &ChanType{
		Dir: x.Dir,
		Elt: tv.GetType(),
	}
	*tv = TypedValue{
		T: gTypeType,
		V: TypeValue{Type: ct},
	}
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
		t := getTypeOf(x)
		m.PushValue(asValue(t))
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
				switch path.Type {
				case VPTypeInterface:
					if debug {
						if ct.Base.Kind() != InterfaceKind {
							panic("should not happen")
						}
					}
					// If xt is a declared interface type, look the type
					// up from the interface.
					// NOTE: It wouldn't work to set depth > 1 because
					// in this case the runtime type is concrete, so the
					// method must be looked up by name anyways.
					ft := ct.Base.(*InterfaceType).
						GetMethodType(path.Name)
					m.PushValue(asValue(ft))
				case VPTypeMethod:
					if debug {
						if ct.Base.Kind() == InterfaceKind {
							panic("should not happen")
						}
					}
					ftv := ct.GetValueRefAt(path)
					ft := ftv.GetFunc().Type
					mt := ft.BoundType()
					m.PushValue(asValue(mt))
				default:
					panic("should not happen")
				}
			} else {
				path.Depth--
				xt = ct.Base
				goto TYPE_SWITCH
			}
		case *InterfaceType:
			if debug {
				if path.Depth != 1 {
					panic("should not happen")
				}
			}
			ft := ct.GetMethodType(path.Name)
			m.PushValue(asValue(ft))
		case *PointerType:
			if dt, ok := ct.Elt.(*DeclaredType); ok {
				if debug {
					if path.Type != VPTypeDeref &&
						path.Type != VPTypeMethod {
						panic("should not happen")
					}
				}
				xt = dt
				goto TYPE_SWITCH
			} else {
				panic("should not happen")
			}
		case *StructType:
			if debug {
				if path.Depth != 1 {
					panic("should not happen")
				}
			}
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
			case *PointerType:
				dt := t.Elt.(*DeclaredType)
				t2 := dt.GetValueRefAt(path).T
				m.PushValue(asValue(t2))
				return
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
				panic(fmt.Sprintf(
					"unexpected selector base typeval: %s of kind %s.",
					xv.GetType().String(),
					xv.GetType().Kind().String()))
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
				reflect.TypeOf(xt)))
		}
	case *SliceExpr:
		start := m.NumValues
		m.PushOp(OpHalt)
		m.PushExpr(x.X)
		m.PushOp(OpTypeOf)
		m.Run() // XXX replace
		xt := m.ReapValues(start)[0].V.(TypeValue).Type
		if pt, ok := xt.(*PointerType); ok {
			m.PushValue(asValue(&SliceType{
				Elt: pt.Elt.Elem(),
			}))
		} else {
			m.PushValue(asValue(&SliceType{
				Elt: xt.Elem(),
			}))
		}
	case *StarExpr:
		start := m.NumValues
		m.PushOp(OpHalt)
		m.PushExpr(x.X)
		m.PushOp(OpTypeOf)
		m.Run() // XXX replace
		xt := m.ReapValues(start)[0].GetType()
		if pt, ok := xt.(*PointerType); ok {
			m.PushValue(asValue(pt.Elt))
		} else if _, ok := xt.(*TypeType); ok {
			m.PushValue(asValue(gTypeType))
		} else {
			panic("should not happen")
		}
	case *RefExpr:
		start := m.NumValues
		m.PushOp(OpHalt)
		m.PushExpr(x.X)
		m.PushOp(OpTypeOf)
		m.Run() // XXX replace
		xt := m.ReapValues(start)[0].GetType()
		m.PushValue(asValue(&PointerType{Elt: xt}))
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
