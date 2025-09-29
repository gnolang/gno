package gnolang

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
		V: toTypeValue(ft),
	})
}

func (m *Machine) doOpArrayType() {
	x := m.PopExpr().(*ArrayTypeExpr)
	t := &ArrayType{}
	if x.Len == nil { // variadic array lit
		t.Vrd = true
	} else {
		lv := m.PopValue()
		if debug {
			// This is how run-time untyped const
			// conversions would work, but we
			// expect the preprocessor to convert
			// these to *ConstExpr.
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
		t.Len = int(lv.GetInt()) // TODO lazy convert?
	}
	tv := m.PeekValue(1) // re-use
	t.Elt = tv.GetType()
	*tv = TypedValue{
		T: gTypeType,
		V: toTypeValue(t),
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
		V: toTypeValue(t),
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
		Params:  params,
		Results: results,
	}
	m.PushValue(TypedValue{
		T: gTypeType,
		V: toTypeValue(ft),
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
		V: toTypeValue(mt),
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
		V: toTypeValue(st),
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
		PkgPath:    m.Package.PkgPath,
		FieldTypes: methods,
		Generic:    x.Generic,
	}
	m.PushValue(TypedValue{
		T: gTypeType,
		V: toTypeValue(it),
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
		V: toTypeValue(ct),
	}
}

// Evaluate the type of a typed (i.e. not untyped) value.
// This function expects const expressions to have been
// already swapped for *ConstExpr in the preprocessor.  If not, panics.
func (m *Machine) doOpStaticTypeOf() {
	x := m.PopExpr()
	switch x := x.(type) {
	case *NameExpr:
		// NOTE: duplicated from doOpEval
		if x.Path.Depth == 0 {
			// Name is in uverse (global).
			gv := Uverse().GetBlock(nil).GetPointerTo(nil, x.Path)
			m.PushValue(asValue(gv.TV.T))
		} else {
			// Get static type from source.
			lb := m.LastBlock()
			st := lb.GetSource(m.Store).GetStaticTypeOfAt(m.Store, x.Path)
			m.PushValue(asValue(st))
		}
	case *BasicLitExpr:
		// Should already be swapped for *ConstExpr.
		// Also, this isn't needed.
		panic("*BasicLitExpr not supported with OpStaticTypeOf")
	case *BinaryExpr:
		switch x.Op {
		case SHL, SHR:
			fallthrough
		case ADD, SUB, MUL, QUO, REM, BAND, BOR, XOR,
			BAND_NOT, LAND, LOR:
			m.PushExpr(x.Left)
			m.PushOp(OpStaticTypeOf)
		case EQL, LSS, GTR, NEQ, LEQ, GEQ:
			m.PushValue(asValue(UntypedBoolType))
		}
	case *CallExpr:
		t := getTypeOf(x)
		m.PushValue(asValue(t))
	case *IndexExpr:
		start := len(m.Values)
		m.PushOp(OpHalt)
		m.PushExpr(x.X)
		m.PushOp(OpStaticTypeOf)
		m.Run(StageRun)
		xt := m.ReapValues(start)[0].GetType()
		m.PushValue(asValue(xt.Elem()))
	case *SelectorExpr:
		start := len(m.Values)
		m.PushOp(OpHalt)
		m.PushExpr(x.X)
		m.PushOp(OpStaticTypeOf)
		m.Run(StageRun)
		xt := m.ReapValues(start)[0].GetType()

		// NOTE: this code segment similar to that in op_types.go
		var dxt Type
		path := x.Path // mutable
		switch path.Type {
		case VPField:
			switch path.Depth { // see tests/selector_test.go for cases.
			case 0:
				dxt = xt
			case 1:
				dxt = baseOf(xt)
				path.SetDepth(0)
			default:
				panic("should not happen")
			}
		case VPSubrefField:
			switch path.Depth {
			case 0:
				dxt = xt.Elem()
				path.SetDepth(0)
			case 1:
				dxt = xt.Elem()
				path.SetDepth(0)
			case 2:
				dxt = baseOf(xt.Elem())
				path.SetDepth(0)
			case 3:
				dxt = baseOf(xt.Elem())
				path.SetDepth(0)
			default:
				panic("should not happen")
			}
		case VPDerefField:
			switch path.Depth {
			case 0:
				dxt = xt.Elem()
				path.Type = VPField
				path.SetDepth(0)
			case 1:
				dxt = xt.Elem()
				path.Type = VPField
				path.SetDepth(0)
			case 2:
				dxt = baseOf(xt.Elem())
				path.Type = VPField
				path.SetDepth(0)
			case 3:
				dxt = baseOf(xt.Elem())
				path.Type = VPField
				path.SetDepth(0)
			default:
				panic("should not happen")
			}
		case VPDerefValMethod:
			dxt = xt.Elem()
			path.Type = VPValMethod
		case VPDerefPtrMethod:
			dxt = xt.Elem()
			path.Type = VPPtrMethod // XXX pseudo
		case VPDerefInterface:
			dxt = xt.Elem()
			path.Type = VPInterface
		default:
			dxt = xt
		}
		if debug {
			path.Validate()
		}

		switch path.Type {
		case VPBlock:
			switch dxt.(type) {
			case *PackageType:
				start := len(m.Values)
				m.PushOp(OpHalt)
				m.PushExpr(x.X)
				m.PushOp(OpEval)
				m.Run(StageRun)
				xv := m.ReapValues(start)[0]
				pv := xv.V.(*PackageValue)
				t := pv.GetBlock(m.Store).GetSource(m.Store).GetStaticTypeOfAt(m.Store, x.Path)
				m.PushValue(asValue(t))
				return
			default:
				panic("should not happen")
			}
		case VPField:
			switch cxt := dxt.(type) {
			case *StructType:
				for _, ft := range cxt.Fields {
					if ft.Name == x.Sel {
						m.PushValue(asValue(ft.Type))
						return
					}
				}
				panic(fmt.Sprintf("struct type %v has no field %s",
					reflect.TypeOf(baseOf(xt)), x.Sel))
			case *TypeType:
				start := len(m.Values)
				m.PushOp(OpHalt)
				m.PushExpr(x.X)
				m.PushOp(OpEval)
				m.Run(StageRun)
				xt := m.ReapValues(start)[0].GetType()
				switch cxt := xt.(type) {
				case *PointerType:
					dt := cxt.Elt.(*DeclaredType)
					t2 := dt.GetStaticValueAt(path).T
					m.PushValue(asValue(t2))
					return
				case *DeclaredType:
					t2 := cxt.GetStaticValueAt(path).T
					m.PushValue(asValue(t2))
					return
				default:
					panic(fmt.Sprintf(
						"unexpected selector base typeval: %s of kind %s",
						xt.String(),
						xt.Kind().String())) // XXX kind?
				}
			default:
				panic(fmt.Sprintf(
					"unexpected selector base type: %s (%s) of kind %s",
					dxt.String(),
					reflect.TypeOf(dxt),
					dxt.Kind().String()))
			}
		case VPSubrefField:
			switch cxt := dxt.(type) {
			case *StructType:
				for _, ft := range cxt.Fields {
					if ft.Name == x.Sel {
						m.PushValue(asValue(
							&PointerType{Elt: ft.Type},
						))
						return
					}
				}
				panic(fmt.Sprintf("struct type %v has no field %s",
					reflect.TypeOf(baseOf(xt)), x.Sel))
			default:
				panic(fmt.Sprintf(
					"unexpected (subref) selector base type: %s (%s) of kind %s",
					dxt.String(),
					reflect.TypeOf(dxt),
					dxt.Kind().String()))
			}
		case VPValMethod, VPPtrMethod:
			ftv := dxt.(*DeclaredType).GetStaticValueAt(path)
			ft := ftv.GetFunc().GetType(m.Store)
			mt := ft.BoundType()
			m.PushValue(asValue(mt))
		case VPInterface:
			_, _, _, ft, _ := findEmbeddedFieldType(dxt.GetPkgPath(), dxt, path.Name, nil)
			m.PushValue(asValue(ft))
		default:
			panic(fmt.Sprintf(
				"unknown value path type %v in selector %s (path %s)",
				path.Type,
				x.String(),
				path.String()))
		}

	case *SliceExpr:
		start := len(m.Values)
		m.PushOp(OpHalt)
		m.PushExpr(x.X)
		m.PushOp(OpStaticTypeOf)
		m.Run(StageRun)
		xt := m.ReapValues(start)[0].V.(TypeValue).Type
		if pt, ok := baseOf(xt).(*PointerType); ok {
			m.PushValue(asValue(&SliceType{
				Elt: pt.Elt.Elem(),
			}))
		} else if xt.Kind() == StringKind {
			m.PushValue(asValue(StringType))
		} else {
			m.PushValue(asValue(&SliceType{
				Elt: xt.Elem(),
			}))
		}
	case *StarExpr:
		start := len(m.Values)
		m.PushOp(OpHalt)
		m.PushExpr(x.X)
		m.PushOp(OpStaticTypeOf)
		m.Run(StageRun)
		xt := m.ReapValues(start)[0].GetType()
		if pt, ok := baseOf(xt).(*PointerType); ok {
			m.PushValue(asValue(pt.Elt))
		} else if _, ok := xt.(*TypeType); ok {
			m.PushValue(asValue(gTypeType))
		} else {
			panic("unexpected star expression")
		}
	case *RefExpr:
		start := len(m.Values)
		m.PushOp(OpHalt)
		m.PushExpr(x.X)
		m.PushOp(OpStaticTypeOf)
		m.Run(StageRun)
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
		m.PushOp(OpStaticTypeOf)
	case *CompositeLitExpr:
		m.PushExpr(x.Type)
		m.PushOp(OpEval)
	case *FuncLitExpr:
		m.PushExpr(&x.Type)
		m.PushOp(OpEval)
	case TypeExpr:
		m.PushValue(asValue(gTypeType))
	case *ConstExpr:
		m.PushValue(asValue(x.T))
	default:
		panic(fmt.Sprintf(
			"unexpected expression of type %v",
			reflect.TypeOf(x)))
	}
}
