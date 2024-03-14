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

// TODO: also compatible with 1.0, 2.0...
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

// main type-assertion functions.
// TODO: document what class of problems its for.
// One of them can be nil, and this lets uninitialized primitives
// and others serve as empty values.  See doOpAdd()
// usage: if debug { assertSameTypes() }
func assertSameTypes(lt, rt Type) {
	if lt == nil && rt == nil {
		// both are nil.
	} else if lt == nil || rt == nil {
		// one is nil.  see function comment.
	} else if lt.Kind() == rt.Kind() &&
		isUntyped(lt) || isUntyped(rt) {
		// one is untyped of same kind.
	} else if lt.Kind() == rt.Kind() &&
		isDataByte(lt) {
		// left is databyte of same kind,
		// specifically for assignments.
		// TODO: make another function
		// and remove this case?
	} else if lt.TypeID() == rt.TypeID() {
		// non-nil types are identical.
	} else {
		debug.Errorf(
			"incompatible operands in binary expression: %s and %s",
			lt.String(),
			rt.String(),
		)
	}
}

func isSameType(lt, rt Type) bool {
	debug.Printf("---isSameType, lt: %v, rt: %v \n", lt, rt)
	return lt == nil && rt == nil || // both are nil/undefined
		(lt != nil && rt != nil) && // both are defined
			(lt.TypeID() == rt.TypeID()) // and identical.
}

// runtime assert
// TODO: consider this!!!
func assertAssignable(lt, rt Type) {
	if debug {
		debug.Printf("assertAssignable, lt: %v, rt: %v, isLeftDataByte: %v, isRightDataByte: %v \n", lt, rt, isDataByte(lt), isDataByte(rt))
	}
	if isSameType(lt, rt) {
		println("1")
		// both are nil/undefined or same type.
	} else if lt == nil || rt == nil { // has support (interface{}) typed-nil, yet support for (native interface{}) typed-nil
		println("2")
		// LHS is undefined
	} else if lt.Kind() == rt.Kind() &&
		isUntyped(lt) || isUntyped(rt) {
		// one is untyped of same kind.
		println("3")
	} else if lt.Kind() == rt.Kind() &&
		isDataByte(lt) {
		// left is databyte of same kind,
		// specifically for assignments.
		// TODO: make another function
		// and remove this case?
		println("4")
	} else {
		//panic("---8")
		debug.Errorf(
			"incompatible operands in binary expression: %s and %s",
			lt.String(),
			rt.String(),
		)
	}
}

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
	case nil: // refer to 0a01, or that can be identified earlier in cmpSpecificity? to remove this check.
		assertMaybeNil("invalid operation, nil can not be compared to", xt)
	default:
		panic(fmt.Sprintf("%v is not comparable \n", dt))
	}
}

func assertMaybeNil(msg string, t Type) {
	if t == nil {
		panic(fmt.Sprintf("%s %s \n", msg, "nil"))
	}
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
// XXX. the name of checkAssignableTo should be considered.
// we have another func of assertAssignable for runtime check, that is a narrow version since we have all concrete types in runtime
// XXX: make it (t Type) CheckAssignableTo?
func checkAssignableTo(xt, dt Type, autoNative bool) (conversionNeeded bool) {
	if debug {
		debug.Printf("checkAssignableTo, xt: %v dt: %v \n", xt, dt)
	}
	// case0
	if xt == nil || dt == nil { // see 0f18, assign8.gno
		return
	}
	// case3
	if dt.Kind() == InterfaceKind { // note native interface
		debug.Println("---dt: ", dt)
		debug.Println("---baseOf dt: ", baseOf(dt))
		debug.Println("---dt.Kind: ", dt.Kind())
		debug.Printf("---xt: %v, baseOf(xt): %v \n", xt, baseOf(xt))
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
				//debug.Println("---matches!")
				//return
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
				xt = dxt.Base           // set as base to do the rest check
				conversionNeeded = true // conduct a type conversion from unnamed to named, it below checks pass
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
			conversionNeeded = true // conduct a type conversion from unnamed to named, it below checks pass
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

		case UntypedBigdecType: // can bigdec assign to bigint?
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
				Uint32Kind, Uint64Kind, BigintKind, BigdecKind, Float32Kind, Float64Kind:
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
			cdt := checkAssignableTo(pt.Elt, cdt.Elt, false)
			return cdt || conversionNeeded
		}
	case *ArrayType:
		if at, ok := xt.(*ArrayType); ok {
			cdt := checkAssignableTo(at.Elt, cdt.Elt, false)
			return cdt || conversionNeeded
		}
	case *SliceType:
		if st, ok := xt.(*SliceType); ok {
			cdt := checkAssignableTo(st.Elt, cdt.Elt, false)
			return cdt || conversionNeeded
		}
	case *MapType:
		if mt, ok := xt.(*MapType); ok {
			cn1 := checkAssignableTo(mt.Key, cdt.Key, false)
			cn2 := checkAssignableTo(mt.Value, cdt.Value, false)
			return cn1 || cn2 || conversionNeeded
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
			debug.Println("---going to check gno2go matches")
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

// type_check for shift expr is a `delayed check`, it has two stages:
// stage1. check on first encounter of a shift expr, if it asserts to be good, pass,
// and if it has const on both side, evalConst, e.g. float32(1 << 4) is legal.
// or if it fails assertion, e.g. uint64(1.0 << 1), tag it to be delayed, to be handled
// on stage2.
// stage2. when an untyped shift expr is used in somewhere, e.g. uint64(1.0 << 1),
// uint64 is used as the potential type of this expr, so uint64(1.0 << 1) is a
// valid representation.
// so dt would be either the type of lhs or the type from outer context.
// isFinal indicates whether it's first check or final check(this happens in checkOrConvertType)

// XXX, is this logic only for this special case?
func (bx *BinaryExpr) checkShiftExpr(store Store, last BlockNode, dt Type, isFinal bool) {
	debug.Printf("---checkShiftExpr: dt: %v, isFinal: %t \n", dt, isFinal)
	var destKind interface{}
	if dt != nil {
		destKind = dt.Kind()
	}
	if checker, ok := binaryChecker[bx.Op]; ok {
		if !checker(dt) {
			if !isFinal {
				// see 10a01, 10a02.
				if dt != nil && dt.Kind() == BigdecKind {
					if lcx, ok := bx.Left.(*ConstExpr); ok {
						if _, ok := bx.Right.(*ConstExpr); ok {
							convertConst(store, last, lcx, BigintType)
							return
						}
					}
				}
			}
			panic(fmt.Sprintf("operator %s not defined on: %v", wordTokenStrings[bx.Op], destKind))
		}
	} else {
		panic("should not happen")
	}
}

// check both sides since no aware of which side is dest type,
// that lt not compatible but rt is compatible would be good.

// AssertCompatible works as a pre-check prior to checkOrConvertType()
// It checks expressions to ensure the compatibility between operands and operators.
// e.g. "a" << 1, the left hand operand is not compatible with <<, it will fail the check.
// Overall,it efficiently filters out incompatible expressions, stopping before the next
// checkOrConvertType() operation to optimize performance.

// things like this would fail: 1.0 % 1, bigInt has a bigger specificity than bidDec.
func (bx *BinaryExpr) AssertCompatible(store Store, lt, rt Type) {
	debug.Printf("---AssertCompatible, bx: %v, lt: %v, rt: %v \n", bx, lt, rt)
	// we can't check compatible with native types at current stage,
	// so leave it to later operations(trans_leave on binaryExpr)
	// to be converted into gno(only for primitive types), and do
	// this check again. (AssertCompatible would be invoked again)
	// non-primitive types is a special case that is not handled
	// (not needed at all?), or it might be expanded to check for case
	// like a gno declared type implement a native interface?
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

	if isComparison(bx.Op) {
		switch bx.Op {
		case EQL, NEQ:
			cmp := cmpSpecificity(lt, rt) // check potential direction of type conversion
			if cmp <= 0 {
				assertComparable(lt, rt)        // only check if dest type is comparable
				checkAssignableTo(lt, rt, true) // check if src type is assignable to dest type
			} else {
				assertComparable(rt, lt)
				checkAssignableTo(rt, lt, true)
			}
		case LSS, LEQ, GTR, GEQ:
			if checker, ok := binaryChecker[bx.Op]; ok {
				bx.checkCompatibility(store, lt, rt, checker, escapedOpStr)
			} else {
				panic("should not happen")
			}
		default:
			panic("invalid comparison operator")
		}
	} else {
		if checker, ok := binaryChecker[bx.Op]; ok {
			bx.checkCompatibility(store, lt, rt, checker, escapedOpStr)
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

// XXX, these kind of type_check happens before checkOrConvertType which get type conversion all
// sorted out, kinda make the type check logic complex, for we have to deal with much undetermined
// types. the upside for this is it quick fails the incompatibility before  concrete type conversion
// happens in checkOrConvertType, but does this weigh out the complexity it brings in?
// TODO: maybe should simplify this to only check dt and left other work to checkOrConvertType
func (bx *BinaryExpr) checkCompatibility(store Store, lt, rt Type, checker func(t Type) bool, escapedOpStr string) {
	debug.Printf("---checkCompatibility, op: %v, lt: %v, rt: %v\n", bx.Op, lt, rt)
	// cache it to be reused in later stage in checkOrConvertType
	var destKind interface{}
	// XXX, this logic is based on an assumption that one side is not defined on op
	// while the other side is ok, and this side is assignable to the other side
	// (that would be converted to the other side eventually).
	// consider 1.2 % int64(1), and 1.0 % int64, is this the only scenario?
	cmp := cmpSpecificity(lt, rt)
	if !checker(lt) { // lt not compatible with op
		if !checker(rt) { // rt not compatible with op
			if lt != nil { // return error on left side that is checked first
				destKind = lt.Kind()
			}
			panic(fmt.Sprintf("operator %s not defined on: %v", escapedOpStr, destKind))
		} else {
			debug.Println("---cmp: ", cmp)
			// left not compatible, right is compatible
			// cmp means the expected convert direction
			// if cmp < 0, means potential conversion
			// from left to right, so check assignable
			// from left to right.
			// if cmp > 0, means potential conversion to
			// left side which is already asserted to be
			// not compatible, so assert fail.
			// no check for cmp == 0 since no possible they have
			// same specificity with one side is compatible
			// and the other is not while assert to be assignable.
			// e.g. "a" - 1.
			if cmp < 0 {
				checkAssignableTo(lt, rt, true)
				store.AddAssignableCheckResult(bx.Left, rt)
			} else {
				if lt != nil {
					destKind = lt.Kind()
				}
				panic(fmt.Sprintf("operator %s not defined on: %v", escapedOpStr, destKind))
			}
		}
	} else if !checker(rt) { // left is compatible, right is not
		if cmp > 0 { // right to left
			checkAssignableTo(rt, lt, true)
			store.AddAssignableCheckResult(bx.Right, lt)
		} else {
			if rt != nil {
				destKind = rt.Kind()
			}
			panic(fmt.Sprintf("operator %s not defined on: %v", escapedOpStr, destKind))
		}
	} else {
		// both good
		// only check both typed.
		//  should not check on both untyped case TypeID equality, e.g.
		// MaxRune         = '\U0010FFFF'
		// MaxRune + 1
		// that left type is untyped rune with default type of Int32Type
		// and right type is untyped bigInt can be converted to int32 too.
		if !isUntyped(lt) && !isUntyped(rt) { // in this stage, lt or rt maybe untyped, not converted yet
			// check when both typed
			if lt != nil && rt != nil { // XXX, this should already be excluded by previous checker check.
				// TODO: filter byte that has no typeID?
				if lt.TypeID() != rt.TypeID() {
					panic(fmt.Sprintf("invalid operation: mismatched types %v and %v \n", lt, rt))
				}
			}
		}
		// check even when both sides are compatible with op,
		if cmp <= 0 {
			checkAssignableTo(lt, rt, true)
			store.AddAssignableCheckResult(bx.Left, rt)

		} else {
			checkAssignableTo(rt, lt, true)
			store.AddAssignableCheckResult(bx.Right, lt)
		}
	}
}

func (ux *UnaryExpr) AssertCompatible(xt, dt Type) {
	debug.Printf("---AssertCompatible, ux: %v, xt: %v, dt: %v \n", ux, xt, dt)

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
	debug.Printf("---AssertCompatible, stmt: %v, t: %v, op: %v \n", idst, t, idst.Op)

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
	debug.Printf("---AssertCompatible, as: %v \n", as)

	escapedOpStr := strings.Replace(wordTokenStrings[as.Op], "%", "%%", 1)
	var destKind interface{}

	// XXX, assume lhs length is same with of rhs
	// TODO: how about call func case?
	// Call case: a, b = x(...)
	for i, x := range as.Lhs {
		lt := evalStaticTypeOf(store, last, x)
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

		debug.Printf("AssertCompatible,lt: %v, rt: %v,op: %v \n", lt, rt, as.Op)
		// check compatible
		if as.Op != ASSIGN {
			if checker, ok := AssignStmtChecker[as.Op]; ok {
				if !checker(lt) {
					if lt != nil {
						destKind = lt.Kind()
					}
					panic(fmt.Sprintf("operator %s not defined on: %v", escapedOpStr, destKind))
				}
				switch as.Op {
				case ADD_ASSIGN, SUB_ASSIGN, MUL_ASSIGN, QUO_ASSIGN, REM_ASSIGN, BAND_ASSIGN, BOR_ASSIGN, BAND_NOT_ASSIGN, XOR_ASSIGN:
					// if both typed
					if !isUntyped(lt) && !isUntyped(rt) { // in this stage, lt or rt maybe untyped, not converted yet
						if lt != nil && rt != nil {
							// TODO: filter byte that has no typeID?
							if lt.TypeID() != rt.TypeID() {
								panic(fmt.Sprintf("invalid operation: mismatched types %v and %v \n", lt, rt))
							}
						}
					}
					// TODO: checkAssignable
				default:
					// do nothing
				}
			} else {
				panic("should not happen")
			}
		} else if as.Op != SHR_ASSIGN && as.Op != SHL_ASSIGN {
			// TODO: test on simple assign
			// check assignable, after done check operand with op.
			// special case if rhs is(or embedded) untyped shift,
			// assignable is checked prior to this.
			checkAssignableTo(rt, lt, false) // TODO: should be for all
			store.AddAssignableCheckResult(as.Rhs[i], lt)
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
	debug.Printf("comSpecificity, t1: %v, t2: %v \n", t1, t2)

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
