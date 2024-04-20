package gnolang

import (
	"errors"
	"fmt"
	"reflect"
)

// here are a range of rules predefined for preprocessor to check the compatibility between operands and operators
// e,g. for binary expr x + y, x, y can only be numeric or string, 1+2, "a" + "b"
// this is used in checkOperandWithOp().
var (
	binaryChecker = map[Word]func(t Type) bool{
		ADD:      isNumericOrString,
		SUB:      isNumeric,
		MUL:      isNumeric,
		QUO:      isNumeric,
		REM:      isIntNum,
		SHL:      isIntNum,
		SHR:      isIntNum,
		BAND:     isIntNum, // bit ops
		XOR:      isIntNum,
		BOR:      isIntNum,
		BAND_NOT: isIntNum,
		LAND:     isBoolean, // logic
		LOR:      isBoolean,
		LSS:      isOrdered, // compare
		LEQ:      isOrdered,
		GTR:      isOrdered,
		GEQ:      isOrdered,
	}
	// TODO: star, addressable
	unaryChecker = map[Word]func(t Type) bool{
		ADD: isNumeric,
		SUB: isNumeric,
		XOR: isIntNum,
		NOT: isBoolean,
	}
	IncDecStmtChecker = map[Word]func(t Type) bool{
		INC: isNumeric,
		DEC: isNumeric,
	}
	AssignStmtChecker = map[Word]func(t Type) bool{
		ADD_ASSIGN:      isNumericOrString,
		SUB_ASSIGN:      isNumeric,
		MUL_ASSIGN:      isNumeric,
		QUO_ASSIGN:      isNumeric,
		REM_ASSIGN:      isIntNum,
		SHL_ASSIGN:      isNumeric,
		SHR_ASSIGN:      isNumeric,
		BAND_ASSIGN:     isIntNum,
		XOR_ASSIGN:      isIntNum,
		BOR_ASSIGN:      isIntNum,
		BAND_NOT_ASSIGN: isIntNum,
	}
)

type category int

const (
	IsBoolean category = 1 << iota
	IsInteger
	IsFloat
	IsString
	IsBigInt
	IsBigDec

	IsNumeric = IsInteger | IsFloat | IsBigInt | IsBigDec
	IsOrdered = IsNumeric | IsString
)

func (pt PrimitiveType) category() category {
	switch pt.Kind() {
	case BoolKind:
		return IsBoolean
	case StringKind:
		return IsString
	case IntKind, Int8Kind, Int16Kind, Int32Kind, Int64Kind, UintKind, Uint8Kind, Uint16Kind, Uint32Kind, Uint64Kind:
		return IsInteger // UntypedRuneType is int32kind, DataByteType is uint8 kind
	case Float32Kind, Float64Kind:
		return IsFloat
	case BigintKind:
		return IsBigInt
	case BigdecKind:
		return IsBigDec
	default:
		panic(fmt.Sprintf("unexpected primitive type %v", pt))
	}
}

func isOrdered(t Type) bool {
	switch t := baseOf(t).(type) {
	case PrimitiveType:
		if t.category()&IsOrdered != 0 {
			return true
		}
		return false
	default:
		return false
	}
}

func isBoolean(t Type) bool {
	switch t := baseOf(t).(type) {
	case PrimitiveType:
		if t.category()&IsBoolean != 0 {
			return true
		}
		return false
	default:
		return false
	}
}

// rune can be numeric and string
func isNumeric(t Type) bool {
	switch t := baseOf(t).(type) {
	case PrimitiveType:
		if t.category()&IsNumeric != 0 {
			return true
		}
		return false
	default:
		return false
	}
}

func isIntNum(t Type) bool {
	switch t := baseOf(t).(type) {
	case PrimitiveType:
		if t.category()&IsInteger != 0 || t.category()&IsBigInt != 0 {
			return true
		}
		return false
	default:
		return false
	}
}

func isNumericOrString(t Type) bool {
	switch t := baseOf(t).(type) {
	case PrimitiveType:
		if t.category()&IsNumeric != 0 || t.category()&IsString != 0 {
			return true
		}
		return false
	default:
		return false
	}
}

// ===========================================================
func assertComparable(xt, dt Type) {
	switch baseOf(dt).(type) {
	case *SliceType, *FuncType, *MapType:
		if xt != nil {
			panic(fmt.Sprintf("%v can only be compared to nil", dt))
		}
	}
	assertComparable2(dt)
}

// assert value with dt is comparable
func assertComparable2(dt Type) {
	if debug {
		debug.Printf("---assertComparable2 dt: %v \n", dt)
	}
	switch cdt := baseOf(dt).(type) {
	case PrimitiveType:
	case *ArrayType:
		switch baseOf(cdt.Elem()).(type) {
		case PrimitiveType, *PointerType, *InterfaceType, *NativeType, *ArrayType, *StructType:
			assertComparable2(cdt.Elem())
		default:
			panic(fmt.Sprintf("%v is not comparable", dt))
		}
	case *StructType:
		for _, f := range cdt.Fields {
			switch cft := baseOf(f.Type).(type) {
			case PrimitiveType, *PointerType, *InterfaceType, *NativeType, *ArrayType, *StructType:
				assertComparable2(cft)
			default:
				panic(fmt.Sprintf("%v is not comparable", dt))
			}
		}
	case *PointerType: // &a == &b
	case *InterfaceType:
	case *SliceType, *FuncType, *MapType:
	case *NativeType:
		if !cdt.Type.Comparable() {
			panic(fmt.Sprintf("%v is not comparable", dt))
		}
	default:
		panic(fmt.Sprintf("%v is not comparable", dt))
	}
}

func maybeNil(t Type) bool {
	switch cxt := baseOf(t).(type) {
	case *SliceType, *FuncType, *MapType, *InterfaceType, *PointerType: //  we don't have unsafePointer
		return true
	case *NativeType:
		switch nk := cxt.Type.Kind(); nk {
		case reflect.Slice, reflect.Func, reflect.Map, reflect.Interface, reflect.Pointer:
			return true
		default:
			return false
		}
	default:
		return false
	}
}

func checkAssignableTo(xt, dt Type, autoNative bool) {
	err := tryCheckAssignableTo(xt, dt, autoNative)
	if err != nil {
		panic(err.Error())
	}
}

// Assert that xt can be assigned as dt (dest type).
// If autoNative is true, a broad range of xt can match against
// a target native dt type, if and only if dt is a native type.
func tryCheckAssignableTo(xt, dt Type, autoNative bool) error {
	if debug {
		debug.Printf("checkAssignableTo, xt: %v dt: %v \n", xt, dt)
	}
	// case0
	if xt == nil { // see test/files/types/0f18
		if !maybeNil(dt) {
			panic(fmt.Sprintf("invalid operation, nil can not be compared to %v", dt))
		}
		return nil
	} else if dt == nil { // _ = xxx, assign8.gno, 0f31. else cases?
		return nil
	}
	// case3
	if dt.Kind() == InterfaceKind { // note native interface
		if idt, ok := baseOf(dt).(*InterfaceType); ok {
			if idt.IsEmptyInterface() { // XXX, can this be merged with IsImplementedBy?
				// if dt is an empty Gno interface, any x ok.
				return nil // ok
			} else if idt.IsImplementedBy(xt) {
				// if dt implements idt, ok.
				return nil // ok
			} else {
				return fmt.Errorf(
					"%s does not implement %s",
					xt.String(),
					dt.String())
			}
		} else if ndt, ok := baseOf(dt).(*NativeType); ok {
			nidt := ndt.Type
			if nidt.NumMethod() == 0 {
				// if dt is an empty Go native interface, ditto.
				return nil // ok
			} else if nxt, ok := baseOf(xt).(*NativeType); ok {
				// if xt has native base, do the naive native.
				if nxt.Type.AssignableTo(nidt) {
					return nil // ok
				} else {
					return fmt.Errorf(
						"cannot use %s as %s",
						nxt.String(),
						nidt.String())
				}
			} else if pxt, ok := baseOf(xt).(*PointerType); ok {
				nxt, ok := pxt.Elt.(*NativeType)
				if !ok {
					return fmt.Errorf(
						"pointer to non-native type cannot satisfy non-empty native interface; %s doesn't implement %s",
						pxt.String(),
						nidt.String())
				}
				// if xt has native base, do the naive native.
				if reflect.PtrTo(nxt.Type).AssignableTo(nidt) {
					return nil // ok
				} else {
					return fmt.Errorf(
						"cannot use %s as %s",
						pxt.String(),
						nidt.String())
				}
			} else if xdt, ok := xt.(*DeclaredType); ok {
				if gno2GoTypeMatches(baseOf(xdt), ndt.Type) {
					return nil
				} // not check against native interface
			} else {
				return fmt.Errorf(
					"unexpected type pair: cannot use %s as %s",
					xt.String(),
					dt.String())
			}
		} else {
			return errors.New("should not happen")
		}
	}

	// case2
	// Special case if xt or dt is *PointerType to *NativeType,
	// convert to *NativeType of pointer kind.
	if pxt, ok := xt.(*PointerType); ok {
		// *gonative{x} is gonative{*x}
		//nolint:misspell
		if enxt, ok := pxt.Elt.(*NativeType); ok {
			xt = &NativeType{
				Type: reflect.PtrTo(enxt.Type),
			}
		}
	}
	if pdt, ok := dt.(*PointerType); ok {
		// *gonative{x} is gonative{*x}
		if endt, ok := pdt.Elt.(*NativeType); ok {
			dt = &NativeType{
				Type: reflect.PtrTo(endt.Type),
			}
		}
	}

	// Special case of xt or dt is *DeclaredType,
	// allow implicit conversion unless both are declared.
	// TODO simplify with .IsNamedType().
	if dxt, ok := xt.(*DeclaredType); ok {
		if ddt, ok := dt.(*DeclaredType); ok {
			// types must match exactly.
			if !dxt.sealed && !ddt.sealed &&
				dxt.PkgPath == ddt.PkgPath &&
				dxt.Name == ddt.Name { // not yet sealed
				return nil // ok
			} else if dxt.TypeID() == ddt.TypeID() {
				return nil // ok
			} else {
				return fmt.Errorf(
					"cannot use %s as %s without explicit conversion",
					dxt.String(),
					ddt.String())
			}
		} else {
			// special case if implicitly named primitive type.
			// TODO simplify with .IsNamedType().
			if _, ok := dt.(PrimitiveType); ok {
				return fmt.Errorf(
					"cannot use %s as %s without explicit conversion",
					dxt.String(),
					dt.String())
			} else {
				// carry on with baseOf(dxt)
				xt = dxt.Base // set as base to do the rest check
			}
		}
	} else if ddt, ok := dt.(*DeclaredType); ok {
		// special case if implicitly named primitive type.
		// TODO simplify with .IsNamedType().
		if _, ok := xt.(PrimitiveType); ok { // e.g. 1 == Int(1)
			if debug {
				debug.Printf("xt is primitiveType: %v, ddt: %v \n", xt, ddt)
			}
			// this is special when dt is the declared type of x
			if !isUntyped(xt) {
				return fmt.Errorf(
					"cannot use %s as %s without explicit conversion",
					xt.String(),
					ddt.String())
			} else { // xt untyped, carry on with check below
				dt = ddt.Base
			}
		} else {
			dt = ddt.Base
		}
	}

	// General cases.
	switch cdt := dt.(type) {
	case PrimitiveType: // case 1
		// if xt is untyped, ensure dt is compatible.
		switch xt {
		case UntypedBoolType:
			if dt.Kind() == BoolKind {
				return nil // ok
			} else {
				return fmt.Errorf(
					"cannot use untyped bool as %s",
					dt.Kind())
			}
		case UntypedStringType:
			if dt.Kind() == StringKind {
				return nil // ok
			} else {
				return fmt.Errorf(
					"cannot use untyped string as %s",
					dt.Kind())
			}
		// XXX, this is a loose check, we don't have the context
		// to check if it is an exact integer, e.g. 1.2 or 1.0(1.0 can be converted to int).
		// this ensure expr like (a % 1.0) pass check, while
		// expr like (a % 1.2) panic at ConvertUntypedTo, which is a delayed assertion when const evaluated.
		// assignable does not guarantee convertible.
		case UntypedBigdecType:
			switch dt.Kind() {
			case IntKind, Int8Kind, Int16Kind, Int32Kind,
				Int64Kind, UintKind, Uint8Kind, Uint16Kind,
				Uint32Kind, Uint64Kind, BigdecKind, Float32Kind, Float64Kind:
				return nil // ok
			default:
				panic(fmt.Sprintf(
					"cannot use untyped Bigdec as %s",
					dt.Kind()))
			}
		case UntypedBigintType:
			switch dt.Kind() {
			case IntKind, Int8Kind, Int16Kind, Int32Kind,
				Int64Kind, UintKind, Uint8Kind, Uint16Kind,
				Uint32Kind, Uint64Kind, BigintKind, BigdecKind, Float32Kind, Float64Kind: // see 0d0
				return nil // ok
			default:
				return fmt.Errorf(
					"cannot use untyped Bigint as %s",
					dt.Kind())
			}
		case UntypedRuneType:
			switch dt.Kind() {
			case IntKind, Int8Kind, Int16Kind, Int32Kind,
				Int64Kind, UintKind, Uint8Kind, Uint16Kind,
				Uint32Kind, Uint64Kind, BigintKind, BigdecKind, Float32Kind, Float64Kind:
				return nil // ok
			default:
				return fmt.Errorf(
					"cannot use untyped rune as %s",
					dt.Kind())
			}

		default:
			if isUntyped(xt) {
				panic("unexpected untyped type")
			}
			if xt.TypeID() == cdt.TypeID() {
				return nil // ok
			}
		}
	case *PointerType: // case 4 from here on
		if pt, ok := xt.(*PointerType); ok {
			checkAssignableTo(pt.Elt, cdt.Elt, false)
		}
		return nil
	case *ArrayType:
		if at, ok := xt.(*ArrayType); ok {
			if at.Len != cdt.Len {
				return fmt.Errorf(
					"cannot use %s as %s",
					at.String(),
					cdt.String())
			}
			checkAssignableTo(at.Elt, cdt.Elt, false)
			return nil
		}
	case *SliceType:
		if st, ok := xt.(*SliceType); ok {
			checkAssignableTo(st.Elt, cdt.Elt, false)
			return nil
		}
	case *MapType:
		if mt, ok := xt.(*MapType); ok {
			checkAssignableTo(mt.Key, cdt.Key, false)
			checkAssignableTo(mt.Value, cdt.Value, false)
			return nil
		}
	case *FuncType:
		if xt.TypeID() == cdt.TypeID() {
			return nil // ok
		}
	case *InterfaceType:
		return errors.New("should not happen")
	case *DeclaredType:
		panic("should not happen")
	case *StructType, *PackageType, *ChanType:
		if xt.TypeID() == cdt.TypeID() {
			return nil // ok
		}
	case *TypeType:
		if xt.TypeID() == cdt.TypeID() {
			return nil // ok
		}
	case *NativeType:
		if !autoNative {
			if debug {
				debug.Printf("native type, xt.TypeID: %v, cdt.TypeID: %v \n", xt.TypeID(), cdt.TypeID())
			}
			if xt.TypeID() == cdt.TypeID() {
				return nil // ok
			}
		} else {
			// autoNative, so check whether matches.
			// xt: any type but a *DeclaredType; could be native.
			// cdt: actual concrete native target type.
			// ie, if cdt can match against xt.
			if gno2GoTypeMatches(xt, cdt.Type) {
				return nil // ok
			}
		}
	default:
		return fmt.Errorf(
			"unexpected type %s",
			dt.String())
	}
	return fmt.Errorf(
		"cannot use %s as %s",
		xt.String(),
		dt.String())
}

// ===========================================================
func (bx *BinaryExpr) checkShiftLhs(dt Type) {
	var destKind interface{}
	if dt != nil {
		destKind = dt.Kind()
	}
	if checker, ok := binaryChecker[bx.Op]; ok {
		if !checker(dt) {
			panic(fmt.Sprintf("operator %s not defined on: %v", bx.Op.TokenString(), destKind))
		}
	} else {
		panic(fmt.Sprintf("checker for %s does not exist", bx.Op))
	}
}

// AssertCompatible works as a pre-check prior to checkOrConvertType()
// It checks against expressions to ensure the compatibility between operands and operators.
// e.g. "a" << 1, the left hand operand is not compatible with <<, it will fail the check.
// Overall,it efficiently filters out incompatible expressions, stopping before the next
// checkOrConvertType() operation to optimize performance.
func (bx *BinaryExpr) AssertCompatible(lt, rt Type) {
	// native type will be converted to gno in latter logic,
	// this check logic will be conduct again
	if lnt, ok := lt.(*NativeType); ok {
		_, ok := go2GnoBaseType(lnt.Type).(PrimitiveType)
		if ok {
			return
		}
	}
	if rnt, ok := rt.(*NativeType); ok {
		_, ok := go2GnoBaseType(rnt.Type).(PrimitiveType)
		if ok {
			return
		}
	}

	OpStr := bx.Op.TokenString()

	xt, dt := lt, rt
	cmp := cmpSpecificity(lt, rt) // check potential direction of type conversion
	if cmp > 0 {
		xt, dt = dt, xt
	}

	if isComparison(bx.Op) {
		switch bx.Op {
		case EQL, NEQ:
			assertComparable(xt, dt)
		case LSS, LEQ, GTR, GEQ:
			if checker, ok := binaryChecker[bx.Op]; ok {
				bx.checkCompatibility(xt, dt, checker, OpStr)
			} else {
				panic(fmt.Sprintf("checker for %s does not exist", bx.Op))
			}
		default:
			panic("invalid comparison operator")
		}
	} else {
		if checker, ok := binaryChecker[bx.Op]; ok {
			bx.checkCompatibility(xt, dt, checker, OpStr)
		} else {
			panic(fmt.Sprintf("checker for %s does not exist", bx.Op))
		}

		switch bx.Op {
		case QUO, REM:
			// special case of zero divisor
			if isQuoOrRem(bx.Op) {
				if rcx, ok := bx.Right.(*ConstExpr); ok {
					if rcx.TypedValue.isZero() {
						panic("invalid operation: division by zero")
					}
				}
			}
		default:
			// do nothing
		}
	}
}

func (bx *BinaryExpr) checkCompatibility(xt, dt Type, checker func(t Type) bool, OpStr string) {
	if !checker(dt) {
		panic(fmt.Sprintf("operator %s not defined on: %v", OpStr, kindString(dt)))
	}

	// e.g. 1%1e9
	if !checker(xt) {
		err := tryCheckAssignableTo(xt, dt, false) // XXX, cache this?
		if err != nil {
			panic(fmt.Sprintf("operator %s not defined on: %v", OpStr, kindString(xt)))
		}
	}
}

func (ux *UnaryExpr) AssertCompatible(xt, dt Type) {
	if nt, ok := xt.(*NativeType); ok {
		if _, ok := go2GnoBaseType(nt.Type).(PrimitiveType); ok {
			return
		}
	}
	// check compatible
	if checker, ok := unaryChecker[ux.Op]; ok {
		if dt == nil {
			dt = xt
		}
		if !checker(dt) {
			panic(fmt.Sprintf("operator %s not defined on: %v", ux.Op.TokenString(), kindString(dt)))
		}
	} else {
		panic(fmt.Sprintf("checker for %s does not exist", ux.Op))
	}
}

func (idst *IncDecStmt) AssertCompatible(t Type) {
	if nt, ok := t.(*NativeType); ok {
		if _, ok := go2GnoBaseType(nt.Type).(PrimitiveType); ok {
			return
		}
	}
	// check compatible
	if checker, ok := IncDecStmtChecker[idst.Op]; ok {
		if !checker(t) {
			panic(fmt.Sprintf("operator %s not defined on: %v", idst.Op.TokenString(), kindString(t)))
		}
	} else {
		panic(fmt.Sprintf("checker for %s does not exist", idst.Op))
	}
}

func (rs *RangeStmt) AssertCompatible(store Store, last BlockNode) {
	if rs.Op == ASSIGN {
		if isBlankIdentifier(rs.Key) && isBlankIdentifier(rs.Value) {
			// both "_"
		} else {
			kt := evalStaticTypeOf(store, last, rs.Key)
			vt := evalStaticTypeOf(store, last, rs.Value)
			xt := evalStaticTypeOf(store, last, rs.X)
			switch cxt := xt.(type) {
			case *MapType:
				checkAssignableTo(cxt.Key, kt, false)
				checkAssignableTo(cxt.Value, vt, false)
			case *SliceType:
				if kt.Kind() != IntKind {
					panic(fmt.Sprintf("index type should be int, but got %v", kt))
				}
				checkAssignableTo(cxt.Elt, vt, false)
			case *ArrayType:
				if kt.Kind() != IntKind {
					panic(fmt.Sprintf("index type should be int, but got %v", kt))
				}
				checkAssignableTo(cxt.Elt, vt, false)
			case PrimitiveType:
				if cxt.Kind() == StringKind {
					if kt != nil && kt.Kind() != IntKind {
						panic(fmt.Sprintf("index type should be int, but got %v", kt))
					}
					if vt != nil && vt.Kind() != Int32Kind { // rune
						panic(fmt.Sprintf("value type should be int32, but got %v", kt))
					}
				}
			}
		}
	}
}

func (as *AssignStmt) AssertCompatible(store Store, last BlockNode) {
	Opstr := as.Op.TokenString()
	if as.Op == ASSIGN || as.Op == DEFINE {
		if len(as.Lhs) > len(as.Rhs) {
			if len(as.Rhs) != 1 {
				panic("should not happen")
			}
			switch cx := as.Rhs[0].(type) {
			case *CallExpr:
				// Call case: a, b = x(...)
				ift := evalStaticTypeOf(store, last, cx.Func)
				cft := getGnoFuncTypeOf(store, ift)
				if len(as.Lhs) != len(cft.Results) {
					panic(fmt.Sprintf(
						"assignment mismatch: "+
							"%d variables but %s returns %d values",
						len(as.Lhs), cx.Func.String(), len(cft.Results)))
				}
				if as.Op == ASSIGN {
					// check assignable
					for i, lx := range as.Lhs {
						if !isBlankIdentifier(lx) { // see composite3.gno
							lxt := evalStaticTypeOf(store, last, lx)
							checkAssignableTo(cft.Results[i].Type, lxt, false)
						}
					}
				}
			case *TypeAssertExpr:
				// Type-assert case: a, ok := x.(type)
				if len(as.Lhs) != 2 {
					panic("should not happen")
				}
				if as.Op == ASSIGN {
					// check assignable to first value
					if !isBlankIdentifier(as.Lhs[0]) { // see composite3.gno
						dt := evalStaticTypeOf(store, last, as.Lhs[0])
						ift := evalStaticTypeOf(store, last, cx)
						checkAssignableTo(ift, dt, false)
					}
					if !isBlankIdentifier(as.Lhs[1]) { // see composite3.gno
						dt := evalStaticTypeOf(store, last, as.Lhs[1])
						if dt != nil && dt.Kind() != BoolKind { // typed, not bool
							panic(fmt.Sprintf("want bool type got %v", dt))
						}
					}
				}
				cx.HasOK = true
			case *IndexExpr: // must be with map type when len(Lhs) > len(Rhs)
				if len(as.Lhs) != 2 {
					panic("should not happen")
				}
				if as.Op == ASSIGN {
					// check first value
					if !isBlankIdentifier(as.Lhs[0]) {
						lt := evalStaticTypeOf(store, last, as.Lhs[0])
						if nx, ok := cx.X.(*NameExpr); ok {
							rx := last.GetStaticBlock().GetBlock().GetPointerTo(store, nx.Path).Deref()
							if mt, ok := rx.T.(*MapType); ok {
								checkAssignableTo(mt.Value, lt, false)
							}
						} else if _, ok := cx.X.(*CompositeLitExpr); ok {
							cpt := evalStaticTypeOf(store, last, cx.X)
							if mt, ok := cpt.(*MapType); ok {
								checkAssignableTo(mt.Value, lt, false)
							} else {
								panic("should not happen")
							}
						}
					}
					if !isBlankIdentifier(as.Lhs[1]) {
						dt := evalStaticTypeOf(store, last, as.Lhs[1])
						if dt != nil && dt.Kind() != BoolKind { // typed, not bool
							panic(fmt.Sprintf("want bool type got %v", dt))
						}
					}
				}
				cx.HasOK = true
			default:
				panic("should not happen")
			}
		} else { // len(Lhs) == len(Rhs)
			if as.Op == ASSIGN {
				// check lhs
				for i, lx := range as.Lhs {
					rt := evalStaticTypeOf(store, last, as.Rhs[i])

					// check native cases
					if rnt, ok := rt.(*NativeType); ok {
						if _, ok := go2GnoBaseType(rnt.Type).(PrimitiveType); ok {
							return
						}
					}

					shouldPanic := true
					switch clx := lx.(type) {
					case *NameExpr, *StarExpr, *SelectorExpr:
						shouldPanic = false
					case *IndexExpr:
						xt := evalStaticTypeOf(store, last, clx.X)
						shouldPanic = xt != nil && xt.Kind() == StringKind
					default:
					}
					if shouldPanic {
						panic(fmt.Sprintf("cannot assign to %v", lx))
					}
				}
			} else { // define
				// NOTE: this is already checked while parsing file
				for i, lx := range as.Lhs {
					rt := evalStaticTypeOf(store, last, as.Rhs[i])
					if rnt, ok := rt.(*NativeType); ok {
						if _, ok := go2GnoBaseType(rnt.Type).(PrimitiveType); ok {
							return
						}
					}
					switch lx.(type) {
					case *NameExpr:
					default:
						panic(fmt.Sprintf("non-name %v on left side of :=", lx))
					}
				}
			}
		}
	} else { // else op other than assign and define
		// TODO: check length of expression
		for i, lx := range as.Lhs {
			lt := evalStaticTypeOf(store, last, lx)
			rt := evalStaticTypeOf(store, last, as.Rhs[i])

			if lnt, ok := lt.(*NativeType); ok {
				if _, ok := go2GnoBaseType(lnt.Type).(PrimitiveType); ok {
					return
				}
			}
			if rnt, ok := rt.(*NativeType); ok {
				if _, ok := go2GnoBaseType(rnt.Type).(PrimitiveType); ok {
					return
				}
			}

			if checker, ok := AssignStmtChecker[as.Op]; ok {
				if !checker(lt) {
					panic(fmt.Sprintf("operator %s not defined on: %v", Opstr, kindString(lt)))
				}
				switch as.Op {
				case ADD_ASSIGN, SUB_ASSIGN, MUL_ASSIGN, QUO_ASSIGN, REM_ASSIGN, BAND_ASSIGN, BOR_ASSIGN, BAND_NOT_ASSIGN, XOR_ASSIGN:
					// check when both typed
					if !isUntyped(lt) && !isUntyped(rt) { // in this stage, lt or rt maybe untyped, not converted yet
						if lt != nil && rt != nil {
							if lt.TypeID() != rt.TypeID() {
								panic(fmt.Sprintf("invalid operation: mismatched types %v and %v", lt, rt))
							}
						}
					}
				default:
					// do nothing
				}
			} else {
				panic(fmt.Sprintf("checker for %s does not exist", as.Op))
			}
		}
	}
}

// misc
func kindString(xt Type) string {
	if xt != nil {
		return xt.Kind().String()
	}
	return "nil"
}

func isQuoOrRem(op Word) bool {
	switch op {
	case QUO, QUO_ASSIGN, REM, REM_ASSIGN:
		return true
	default:
		return false
	}
}

func isComparison(op Word) bool {
	switch op {
	case EQL, NEQ, LSS, LEQ, GTR, GEQ:
		return true
	default:
		return false
	}
}

func cmpSpecificity(t1, t2 Type) int {
	// check nil
	if t1 == nil { // see test file 0f46
		return -1 // also with both nil
	} else if t2 == nil {
		return 1
	}

	// check interface
	if it1, ok := baseOf(t1).(*InterfaceType); ok {
		if it1.IsEmptyInterface() {
			return 1 // left empty interface
		} else {
			if it2, ok := baseOf(t2).(*InterfaceType); ok {
				if it2.IsEmptyInterface() { // right empty interface
					return -1
				} else {
					return 0 // both non-empty interface
				}
			} else {
				return 1 // right not interface
			}
		}
	} else if _, ok := t2.(*InterfaceType); ok {
		return -1 // left not interface, right is interface
	}

	// primitive types
	t1s, t2s := 0, 0
	if t1p, ok := t1.(PrimitiveType); ok {
		t1s = t1p.Specificity()
	}
	if t2p, ok := t2.(PrimitiveType); ok {
		t2s = t2p.Specificity()
	}
	if t1s < t2s {
		// NOTE: higher specificity has lower value, so backwards.
		return 1
	} else if t1s == t2s {
		return 0
	} else {
		return -1
	}
}

func isBlankIdentifier(x Expr) bool {
	if nx, ok := x.(*NameExpr); ok {
		if nx.Path.Depth == 0 && nx.Path.Index == 0 && nx.Name == "_" {
			return true
		}
	}
	return false
}
