package gnolang

import (
	"fmt"
	"reflect"
	"strings"
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
	IncDecStmtChecker = map[Word]func(t Type) bool{ // NOTE: to be consistent with op_inc_dec.go, line3, no float support for now(while go does).
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
// assertComparable is used in preprocess.
// assert value with dt is comparable
// special case when both typed, check if type identical
func assertComparable(xt, dt Type) {
	if debug {
		debug.Printf("--- assertComparable---, xt: %v, dt: %v \n", xt, dt)
	}
	switch cdt := baseOf(dt).(type) {
	case PrimitiveType:
		// both typed primitive types
		if _, ok := baseOf(xt).(PrimitiveType); ok {
			if !isUntyped(xt) && !isUntyped(dt) { // in this stage, lt or rt maybe untyped, not converted yet
				if xt != nil && dt != nil {
					if xt.TypeID() != dt.TypeID() {
						panic(fmt.Sprintf("invalid operation: mismatched types %v and %v \n", xt, dt))
					}
				}
			}
		}
	case *ArrayType: // NOTE: no recursive allowed
		switch baseOf(cdt.Elem()).(type) {
		case PrimitiveType, *PointerType, *InterfaceType, *NativeType: // NOTE: nativeType?
			switch cxt := baseOf(xt).(type) {
			case *ArrayType:
				if cxt.Len != cdt.Len { // check length
					panic(fmt.Sprintf("%v and %v cannot be compared \n", cxt, cdt))
				}
			default:
				panic(fmt.Sprintf("%v and %v cannot be compared \n", cxt, cdt))
			}
		default:
			panic(fmt.Sprintf("%v and %v cannot be compared \n", xt, cdt))
		}
	case *StructType:
		for _, f := range cdt.Fields {
			switch baseOf(f.Type).(type) {
			case PrimitiveType, *PointerType, *InterfaceType, *NativeType:
			default:
				panic(fmt.Sprintf("%v and %v cannot be compared \n", xt, cdt))
			}
		}
	case *PointerType: // &a == &b
	case *InterfaceType: // var a, b interface{}
	case *SliceType, *FuncType, *MapType:
		if xt != nil {
			panic(fmt.Sprintf("%v can only be compared to nil \n", dt))
		}
	case *NativeType:
		if !cdt.Type.Comparable() {
			panic(fmt.Sprintf("%v is not comparable \n", dt))
		}
	case nil: // see 0a01, or that can be identified earlier in cmpSpecificity? to remove this check.
		if xt == nil {
			panic(fmt.Sprintf("invalid operation, nil can not be compared to %s \n", "nil"))
		}
		assertMaybeNil("invalid operation, nil can not be compared to", xt)
	default:
		panic(fmt.Sprintf("%v is not comparable \n", dt))
	}
}

func assertMaybeNil(msg string, t Type) {
	switch cxt := baseOf(t).(type) {
	case *SliceType, *FuncType, *MapType, *InterfaceType, *PointerType: //  we don't have unsafePointer
	case *NativeType:
		switch nk := cxt.Type.Kind(); nk {
		case reflect.Slice, reflect.Func, reflect.Map, reflect.Interface, reflect.Pointer:
		default:
			panic(fmt.Sprintf("%s %s \n", msg, nk))
		}
	default:
		panic(fmt.Sprintf("%s %s \n", msg, t))
	}
}

// Assert that xt can be assigned as dt (dest type).
// If autoNative is true, a broad range of xt can match against
// a target native dt type, if and only if dt is a native type.

// check if xt can be assigned to dt. conversionNeeded indicates further conversion needed especially from unnamed -> named
// case 0. nil check
// case 1. untyped const to typed const with same kind
// case 2. unnamed to named
// case 3. dt is interface, xt satisfied dt
// case 4. general cases for primitives and composite.
func checkAssignableTo(xt, dt Type, autoNative bool) {
	if debug {
		debug.Printf("checkAssignableTo, xt: %v dt: %v \n", xt, dt)
	}
	// case0
	if xt == nil || dt == nil { // see 0f18, assign8.gno
		return
	}
	// case3
	if dt.Kind() == InterfaceKind { // note native interface
		if idt, ok := baseOf(dt).(*InterfaceType); ok {
			if idt.IsEmptyInterface() { // XXX, can this be merged with IsImplementedBy?
				// if dt is an empty Gno interface, any x ok.
				return // ok
			} else if idt.IsImplementedBy(xt) {
				// if dt implements idt, ok.
				return // ok
			} else {
				panic(fmt.Sprintf(
					"%s does not implement %s",
					xt.String(),
					dt.String()))
			}
		} else if ndt, ok := baseOf(dt).(*NativeType); ok {
			debug.Printf("---ndt: %v \n", ndt)
			nidt := ndt.Type
			if nidt.NumMethod() == 0 {
				// if dt is an empty Go native interface, ditto.
				return // ok
			} else if nxt, ok := baseOf(xt).(*NativeType); ok {
				// if xt has native base, do the naive native.
				if nxt.Type.AssignableTo(nidt) {
					return // ok
				} else {
					panic(fmt.Sprintf(
						"cannot use %s as %s",
						nxt.String(),
						nidt.String()))
				}
			} else if pxt, ok := baseOf(xt).(*PointerType); ok {
				nxt, ok := pxt.Elt.(*NativeType)
				if !ok {
					panic(fmt.Sprintf(
						"pointer to non-native type cannot satisfy non-empty native interface; %s doesn't implement %s",
						pxt.String(),
						nidt.String()))
				}
				// if xt has native base, do the naive native.
				if reflect.PtrTo(nxt.Type).AssignableTo(nidt) {
					return // ok
				} else {
					panic(fmt.Sprintf(
						"cannot use %s as %s",
						pxt.String(),
						nidt.String()))
				}
			} else if xdt, ok := xt.(*DeclaredType); ok {
				if gno2GoTypeMatches(baseOf(xdt), ndt.Type) {
					return
				} // not check against native interface
			} else {
				panic(fmt.Sprintf(
					"unexpected type pair: cannot use %s as %s",
					xt.String(),
					dt.String()))
			}
		} else {
			panic("should not happen")
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
				return // ok
			} else if dxt.TypeID() == ddt.TypeID() {
				return // ok
			} else {
				panic(fmt.Sprintf(
					"cannot use %s as %s without explicit conversion",
					dxt.String(),
					ddt.String()))
			}
		} else {
			// special case if implicitly named primitive type.
			// TODO simplify with .IsNamedType().
			if _, ok := dt.(PrimitiveType); ok {
				panic(fmt.Sprintf(
					"cannot use %s as %s without explicit conversion",
					dxt.String(),
					dt.String()))
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
				panic(fmt.Sprintf(
					"cannot use %s as %s without explicit conversion",
					xt.String(),
					ddt.String()))
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
				return // ok
			} else {
				panic(fmt.Sprintf(
					"cannot use untyped bool as %s",
					dt.Kind()))
			}
		case UntypedStringType:
			if dt.Kind() == StringKind {
				return // ok
			} else {
				panic(fmt.Sprintf(
					"cannot use untyped string as %s",
					dt.Kind()))
			}
		// XXX, this is a loose check, we don't have the context
		// to check if it is an exact integer, e.g. 1.2 or 1.0(1.0 can be converted to int).
		// this ensure expr like (a % 1.0) pass check, while
		// expr like (a % 1.2) panic at ConvertUntypedTo, which is a delayed assertion when const evaluated.
		// assignable does not guarantee convertable.
		case UntypedBigdecType:
			switch dt.Kind() {
			case IntKind, Int8Kind, Int16Kind, Int32Kind,
				Int64Kind, UintKind, Uint8Kind, Uint16Kind,
				Uint32Kind, Uint64Kind, BigdecKind, Float32Kind, Float64Kind:
				return // ok
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
				return // ok
			default:
				panic(fmt.Sprintf(
					"cannot use untyped Bigint as %s",
					dt.Kind()))
			}
		case UntypedRuneType:
			switch dt.Kind() {
			case IntKind, Int8Kind, Int16Kind, Int32Kind,
				Int64Kind, UintKind, Uint8Kind, Uint16Kind,
				Uint32Kind, Uint64Kind, BigintKind, BigdecKind, Float32Kind, Float64Kind:
				return // ok
			default:
				panic(fmt.Sprintf(
					"cannot use untyped rune as %s",
					dt.Kind()))
			}

		default:
			if isUntyped(xt) {
				panic("unexpected untyped type")
			}
			if xt.TypeID() == cdt.TypeID() {
				return // ok
			}
		}
	case *PointerType: // case 4 from here on
		if pt, ok := xt.(*PointerType); ok {
			checkAssignableTo(pt.Elt, cdt.Elt, false)
		}
		return
	case *ArrayType:
		if at, ok := xt.(*ArrayType); ok {
			checkAssignableTo(at.Elt, cdt.Elt, false)
			return
		}
	case *SliceType:
		if st, ok := xt.(*SliceType); ok {
			checkAssignableTo(st.Elt, cdt.Elt, false)
			return
		}
	case *MapType:
		if mt, ok := xt.(*MapType); ok {
			checkAssignableTo(mt.Key, cdt.Key, false)
			checkAssignableTo(mt.Value, cdt.Value, false)
			return
		}
	case *FuncType:
		if xt.TypeID() == cdt.TypeID() {
			return // ok
		}
	case *InterfaceType:
		panic("should not happen")
	case *DeclaredType:
		// do nothing, untyped to declared type
		return
		// panic("should not happen")
	case *StructType, *PackageType, *ChanType:
		if xt.TypeID() == cdt.TypeID() {
			return // ok
		}
	case *TypeType:
		if xt.TypeID() == cdt.TypeID() {
			return // ok
		}
	case *NativeType:
		if !autoNative {
			if debug {
				debug.Printf("native type, xt.TypeID: %v, cdt.TypeID: %v \n", xt.TypeID(), cdt.TypeID())
			}
			if xt.TypeID() == cdt.TypeID() {
				return // ok
			}
		} else {
			// autoNative, so check whether matches.
			// xt: any type but a *DeclaredType; could be native.
			// cdt: actual concrete native target type.
			// ie, if cdt can match against xt.
			if gno2GoTypeMatches(xt, cdt.Type) {
				return // ok
			}
		}
	default:
		panic(fmt.Sprintf(
			"unexpected type %s",
			dt.String()))
	}
	panic(fmt.Sprintf(
		"cannot use %s as %s",
		xt.String(),
		dt.String()))
}

// ===========================================================
func (bx *BinaryExpr) checkShiftExpr(dt Type) {
	var destKind interface{}
	if dt != nil {
		destKind = dt.Kind()
	}
	if checker, ok := binaryChecker[bx.Op]; ok {
		if !checker(dt) {
			panic(fmt.Sprintf("operator %s not defined on: %v", wordTokenStrings[bx.Op], destKind))
		}
	} else {
		panic("should not happen")
	}
}

// AssertCompatible works as a pre-check prior to checkOrConvertType()
// It checks against expressions to ensure the compatibility between operands and operators.
// e.g. "a" << 1, the left hand operand is not compatible with <<, it will fail the check.
// Overall,it efficiently filters out incompatible expressions, stopping before the next
// checkOrConvertType() operation to optimize performance.
func (bx *BinaryExpr) AssertCompatible(lt, rt Type) {
	// we can't check compatible with native types at current stage,
	// so leave it to later operations(trans_leave on binaryExpr)
	// to be converted into gno(only for primitive types), and do
	// this check again. (AssertCompatible would be invoked again)
	// non-primitive types is a special case that is not handled.
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

	escapedOpStr := strings.Replace(wordTokenStrings[bx.Op], "%", "%%", 1)

	var xt, dt Type
	cmp := cmpSpecificity(lt, rt) // check potential direction of type conversion
	if cmp <= 0 {
		xt = lt
		dt = rt
	} else {
		xt = rt
		dt = lt
	}

	if isComparison(bx.Op) {
		switch bx.Op {
		case EQL, NEQ:
			assertComparable(xt, dt) // only check if dest type is comparable
		case LSS, LEQ, GTR, GEQ:
			if checker, ok := binaryChecker[bx.Op]; ok {
				bx.checkCompatibility(xt, dt, checker, escapedOpStr)
			} else {
				panic("should not happen")
			}
		default:
			panic("invalid comparison operator")
		}
	} else {
		if checker, ok := binaryChecker[bx.Op]; ok {
			bx.checkCompatibility(xt, dt, checker, escapedOpStr)
		} else {
			panic("should not happen")
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

func (bx *BinaryExpr) checkCompatibility(xt, dt Type, checker func(t Type) bool, escapedOpStr string) {
	var destKind interface{}

	if !checker(dt) {
		if dt != nil {
			destKind = dt.Kind()
		}
		panic(fmt.Sprintf("operator %s not defined on: %v", escapedOpStr, destKind))
	}

	defer func() { // rewrite err msg
		if r := recover(); r != nil {
			if xt != nil {
				destKind = xt.Kind()
			}
			panic(fmt.Sprintf("operator %s not defined on: %v", escapedOpStr, destKind))
		}
	}()

	// e.g. 1%1e9
	if !checker(xt) {
		checkAssignableTo(xt, dt, false) // XXX, cache this?
	}
}

func (ux *UnaryExpr) AssertCompatible(xt, dt Type) {
	var destKind interface{}

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
			if dt != nil {
				destKind = dt.Kind()
			}
			panic(fmt.Sprintf("operator %s not defined on: %v", wordTokenStrings[ux.Op], destKind))
		}
	} else {
		panic("should not happen")
	}
}

func (idst *IncDecStmt) AssertCompatible(t Type) {
	var destKind interface{}

	if nt, ok := t.(*NativeType); ok {
		if _, ok := go2GnoBaseType(nt.Type).(PrimitiveType); ok {
			return
		}
	}
	// check compatible
	if checker, ok := IncDecStmtChecker[idst.Op]; ok {
		if !checker(t) {
			if t != nil {
				destKind = t.Kind()
			}
			panic(fmt.Sprintf("operator %s not defined on: %v", wordTokenStrings[idst.Op], destKind))
		}
	} else {
		panic("should not happen")
	}
}

func (as *AssignStmt) AssertCompatible(store Store, last BlockNode) {
	escapedOpStr := strings.Replace(wordTokenStrings[as.Op], "%", "%%", 1)
	var destKind interface{}

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
				// check assignable
				for i, lx := range as.Lhs {
					lxt := evalStaticTypeOf(store, last, lx)
					checkAssignableTo(cft.Results[i].Type, lxt, false) // TODO: autoNative?
				}
			case *TypeAssertExpr:
				// Type-assert case: a, ok := x.(type)
				if len(as.Lhs) != 2 {
					panic("should not happen")
				}
				debug.Println("---type assertion, assert type, reflect type of at: ", cx.Type, reflect.TypeOf(cx.Type))
				if ctex, ok := cx.Type.(*constTypeExpr); ok {
					// check assignable
					dt := evalStaticTypeOf(store, last, as.Lhs[0])
					checkAssignableTo(ctex.Type, dt, false)
				} else if _, ok := cx.Type.(*InterfaceTypeExpr); ok {
					dt := evalStaticTypeOf(store, last, as.Lhs[0])
					if isBlankIdentifier(as.Lhs[0]) { // see composite3.gno
						debug.Println("---blank")
					} else if dt != nil && dt.Kind() == InterfaceKind {
						// do nothing
					} else {
						panic(fmt.Sprintf("cannot assign to %v \n", as.Lhs[0]))
					}
				}
				cx.HasOK = true
			case *IndexExpr:
				if len(as.Lhs) != 2 {
					panic("should not happen")
				}
				lt := evalStaticTypeOf(store, last, as.Lhs[0])
				debug.Println("---cx.X, type of: ", cx.X, reflect.TypeOf(cx.X))
				if nx, ok := cx.X.(*NameExpr); ok {
					rx := last.GetStaticBlock().GetBlock().GetPointerTo(store, nx.Path).Deref()
					debug.Println("---rx, type of rx: ", rx, reflect.TypeOf(rx))
					debug.Println("---rx.T: ", rx.T)
					debug.Println("---rx.V: ", rx.V)

					if mt, ok := rx.T.(*MapType); ok {
						debug.Println("---mt: ", mt)
						debug.Println("---type of value of map type: ", mt.Value)
						checkAssignableTo(mt.Value, lt, false)
					}
				} else if _, ok := cx.X.(*CompositeLitExpr); ok {
					cpt := evalStaticTypeOf(store, last, cx.X)
					debug.Println("---cpt: ", cpt)
					if mt, ok := cpt.(*MapType); ok {
						debug.Println("---mt: ", mt)
						checkAssignableTo(mt.Value, lt, false)
					}
				}
				cx.HasOK = true
			default:
				panic("should not happen")
			}
		} else {
			if as.Op == ASSIGN {
				for i, lx := range as.Lhs {
					rt := evalStaticTypeOf(store, last, as.Rhs[i])

					debug.Println("lx, type of lx: ", lx, reflect.TypeOf(lx))

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
						debug.Println("---ix.X, type of ix.X: ", clx.X, reflect.TypeOf(clx.X))
						xt := evalStaticTypeOf(store, last, clx.X)
						shouldPanic = xt != nil && xt.Kind() == StringKind
					default:
					}
					if shouldPanic {
						panic(fmt.Sprintf("cannot assign to %v \n", lx))
					}
				}
			} else {
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
						panic(fmt.Sprintf("non-name %v on left side of := \n", lx))
					}
				}
			}
		}
	} else {
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
					if lt != nil {
						destKind = lt.Kind()
					}
					panic(fmt.Sprintf("operator %s not defined on: %v", escapedOpStr, destKind))
				}
				switch as.Op {
				case ADD_ASSIGN, SUB_ASSIGN, MUL_ASSIGN, QUO_ASSIGN, REM_ASSIGN, BAND_ASSIGN, BOR_ASSIGN, BAND_NOT_ASSIGN, XOR_ASSIGN:
					// check when both typed
					if !isUntyped(lt) && !isUntyped(rt) { // in this stage, lt or rt maybe untyped, not converted yet
						if lt != nil && rt != nil {
							if lt.TypeID() != rt.TypeID() {
								panic(fmt.Sprintf("invalid operation: mismatched types %v and %v \n", lt, rt))
							}
						}
					}
				default:
					// do nothing
				}
			} else {
				panic("should not happen")
			}
		}
	}
}

// misc
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
