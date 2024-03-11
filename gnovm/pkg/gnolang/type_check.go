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
		REM:      isIntNum, // in compile stage good for bigdec, that can be converted to int, `checkAssignableTo`
		SHL:      isIntNum, // NOTE: 1.0 << 1 is legal in Go. consistent with op_binary for now.
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
	IsInvalid category = 0
	IsBoolean category = 1 << iota
	IsInteger
	IsUnsigned
	IsFloat
	IsString
	IsBigInt
	IsBigDec
	IsRune

	IsNumeric = IsInteger | IsUnsigned | IsFloat | IsBigInt | IsBigDec
	IsOrdered = IsNumeric | IsString
	//IsIntOrFloat = IsInteger | IsUnsigned | IsFloat | IsBigInt | IsBigDec
)

// category makes it more convenient than compare with types
func (pt PrimitiveType) predicate() category {
	switch pt {
	case InvalidType:
		return IsInvalid
	case UntypedBoolType:
		return IsBoolean
	case BoolType:
		return IsBoolean
	case UntypedStringType:
		return IsString
	case StringType:
		return IsString
	case IntType:
		return IsInteger
	case Int8Type:
		return IsInteger
	case Int16Type:
		return IsInteger
	case UntypedRuneType: // TODO: this is treat as DataByteType, GUESS, refer to op_inc_dec
		return IsRune
	case Int32Type:
		return IsInteger
	case Int64Type:
		return IsInteger
	case UintType:
		return IsUnsigned
	case Uint8Type:
		return IsUnsigned
	case DataByteType:
		return IsUnsigned // TODO: consider this
	case Uint16Type:
		return IsUnsigned
	case Uint32Type:
		return IsUnsigned
	case Uint64Type:
		return IsUnsigned
	case Float32Type:
		return IsFloat
	case Float64Type:
		return IsFloat
	case UntypedBigintType:
		return IsBigInt
	case BigintType:
		return IsBigInt
	case UntypedBigdecType:
		return IsBigDec
	case BigdecType:
		return IsBigDec
	default:
		panic(fmt.Sprintf("unexpected primitive type %d", pt))
	}
}

func isOrdered(t Type) bool {
	switch t := baseOf(t).(type) {
	case PrimitiveType:
		if t.predicate() != IsInvalid && t.predicate()&IsOrdered != 0 || t.predicate()&IsRune != 0 {
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
		if t.predicate() != IsInvalid && t.predicate()&IsBoolean != 0 {
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
		if t.predicate() != IsInvalid && t.predicate()&IsNumeric != 0 || t.predicate()&IsRune != 0 {
			return true
		}
		return false
	default:
		return false
	}
}

// signed or unsigned int
func isIntNum(t Type) bool {
	switch t := baseOf(t).(type) {
	case PrimitiveType:
		if t.predicate() != IsInvalid && t.predicate()&IsInteger != 0 || t.predicate()&IsUnsigned != 0 || t.predicate()&IsBigInt != 0 || t.predicate()&IsRune != 0 {
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
		if t.predicate() != IsInvalid && t.predicate()&IsNumeric != 0 || t.predicate()&IsString != 0 || t.predicate()&IsRune != 0 {
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
		// panic("panic assertSameTypes")
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
	} else if lt.Kind() == InterfaceKind &&
		IsImplementedBy(lt, rt) { // TODO: consider this
		// rt implements lt (and lt is nil interface).
	} else if rt.Kind() == InterfaceKind &&
		IsImplementedBy(rt, lt) { // TODO: consider this
		println("5")
		// lt implements rt (and rt is nil interface).
	} else if nrt, ok := rt.(*NativeType); ok { // see 0f2, for native types
		if gno2GoTypeMatches(lt, nrt.Type) {
			println("6.1")
		} else {
			println("6.2")
		}
	} else if nlt, ok := lt.(*NativeType); ok {
		if gno2GoTypeMatches(rt, nlt.Type) {
			println("7.1")
		} else {
			println("7.2")
		}
	} else {
		panic("---8")
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
// TODO: move this to type_checker? as a method?
// TODO: check mismatch on this, not postpone to checkOrConvertType
func assertComparable(xt, dt Type) {
	if debug {
		debug.Printf("--- assertComparable---, xt: %v, dt: %v \n", xt, dt)
	}
	switch cdt := baseOf(dt).(type) {
	case PrimitiveType:
		// both typed primitive types
		if _, ok := baseOf(xt).(PrimitiveType); ok {
			//var isMixed bool
			//if _, ok := xt.(*NativeType); ok {
			//	if _, ok := dt.(*NativeType); ok {
			//		isMixed = false
			//	} else {
			//		isMixed = true
			//	}
			//} else {
			//	if _, ok := dt.(*NativeType); ok {
			//		isMixed = true
			//	}
			//}
			//if !isMixed { // both not native or both native we can check typeID ==
			if !isUntyped(xt) && !isUntyped(dt) { // in this stage, lt or rt maybe untyped, not converted yet
				debug.Println("---both typed")
				if xt != nil && dt != nil {
					// TODO: filter byte that has no typeID?
					if xt.TypeID() != dt.TypeID() {
						panic(fmt.Sprintf("invalid operation: mismatched types %v and %v \n", xt, dt))
					}
				}
			}
			//}
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

// AssignabilityCache caches the results of assignability checks.
type AssignabilityCache struct {
	storage map[string]map[string]bool
}

func NewAssignabilityCache() *AssignabilityCache {
	return &AssignabilityCache{
		storage: make(map[string]map[string]bool),
	}
}

func (ac *AssignabilityCache) Add(key Expr, value Type) {
	if ac.storage[key.String()] == nil {
		ac.storage[key.String()] = make(map[string]bool)
	}
	ac.storage[key.String()][value.String()] = true
}

func (ac *AssignabilityCache) Exists(key Expr, value Type) bool {
	if valSet, ok := ac.storage[key.String()]; ok {
		// Check if the value exists in the set
		return valSet[value.String()]
	}
	return false
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
// TODO: make it (t Type) CheckAssignableTo
// TODO: make cmp specificity here
func checkAssignableTo(xt, dt Type, autoNative bool) (conversionNeeded bool) {
	if debug {
		debug.Printf("checkAssignableTo, xt: %v dt: %v \n", xt, dt)
	}
	// case0
	if xt == nil { // refer to 0f18_filetest
		// XXX, this seems duplicated with assertComparable
		//assertMaybeNil("invalid operation, nil can not be compared to", dt)
		return
	}
	if dt == nil { // refer to assign8.gno
		return
	}
	// case3
	if dt.Kind() == InterfaceKind {
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
				gno2GoTypeMatches(baseOf(xdt), ndt.Type)
				debug.Println("---matches!")
				return
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

// TODO: expand this
//type Checker struct {
//}

// check both sides since no aware of which side is dest type
// that lt not compatible but rt is compatible would be good
// things like this would fail: 1.0 % 1, bigInt has a bigger specificity than bidDec.

// AssertCompatible works as a pre-check prior to checkOrConvertType()
// It checks expressions to ensure the compatibility between operands and operators.
// e.g. "a" << 1, the left hand operand is not compatible with <<, it will fail the check.
// Overall,it efficiently filters out incompatible expressions, stopping before the next
// checkOrConvertType() operation to optimize performance.

// dt is a special case for binary expr that the dest type for the
// lhs determined by outer context
func (bx *BinaryExpr) AssertCompatible(lt, rt, dt Type) {
	debug.Printf("---AssertCompatible, bx: %v \n", bx)
	debug.Printf("---AssertCompatible, bx.Left: %T \n", bx.Left)
	debug.Printf("---AssertCompatible, bx.Right: %T \n", bx.Right)

	// get left type and right type
	//lt := evalStaticTypeOf(store, last, bx.Left)
	//rt := evalStaticTypeOf(store, last, bx.Right)

	// we can't check compatible with native types
	// at current stage, so leave it to checkOrConvertType
	// to secondary call this assert logic again

	//var leftNative, rightNative bool
	if lnt, ok := lt.(*NativeType); ok {
		debug.Println("---lt native")
		//leftNative = true
		_, ok := go2GnoBaseType(lnt.Type).(PrimitiveType)
		if ok {
			debug.Println("---lt native primitive, return")
			return
		}
	}
	if rnt, ok := rt.(*NativeType); ok {
		//rightNative = true
		debug.Println("---rt native, gnoType", rnt.gnoType)
		_, ok := go2GnoBaseType(rnt.Type).(PrimitiveType) // TODO: check kind instead
		if ok {
			debug.Println("---right native primitive, return")
			return
		}
	}

	debug.Printf("AssertCompatible,lt: %v, rt: %v,op: %v \n", lt, rt, bx.Op)
	escapedOpStr := strings.Replace(wordTokenStrings[bx.Op], "%", "%%", 1)

	if isComparison(bx.Op) {
		switch bx.Op {
		case EQL, NEQ:
			assertComparable(lt, rt)
			// XXX, should not check compatible for equality compare?
			// the native type here are both non-primitive
			// not able to check assignable, e.g. MyError == err, that if left satisfied right native interface.
			//if !leftNative && !rightNative {
			cmp := cmpSpecificity(lt, rt)
			if cmp <= 0 {
				checkAssignableTo(lt, rt, true)
			} else {
				checkAssignableTo(rt, lt, true)
			}
			//}
		case LSS, LEQ, GTR, GEQ:
			if pred, ok := binaryChecker[bx.Op]; ok {
				bx.checkCompatibility(lt, rt, pred, escapedOpStr, dt)
			} else {
				panic("should not happen")
			}
		default:
			panic("invalid comparison operator")
		}
	} else {
		// TODO: if not assign, check implicit compatible
		if pred, ok := binaryChecker[bx.Op]; ok {
			bx.checkCompatibility(lt, rt, pred, escapedOpStr, dt)
		} else {
			panic("should not happen")
		}

		switch bx.Op {
		case ADD, SUB, MUL, QUO, REM, BAND, BOR, BAND_NOT, XOR, LAND, LOR:
			// if both typed
			//if !isMixed { // both native or both not native that we can check
			if !isUntyped(lt) && !isUntyped(rt) {
				if lt != nil && rt != nil { // NOTE: is this necessary?
					if lt.TypeID() != rt.TypeID() {
						panic(fmt.Sprintf("invalid operation: mismatched types %v and %v \n", lt, rt))
					}
				}
			}
			//}
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

func (bx *BinaryExpr) checkCompatibility(lt, rt Type, pred func(t Type) bool, escapedOpStr string, dt Type) {
	debug.Println("---checkCompatibility, op: ", bx.Op)
	debug.Printf("---checkCompatibility, lt: %v, rt: %v \n", lt, rt)
	debug.Printf("---checkCompatibility, dt: %v \n", dt)
	AssignableCheckCache = NewAssignabilityCache()
	var destKind interface{}

	// shl/shr
	if bx.Op == SHL || bx.Op == SHR {
		if dt == nil {
			dt = lt
		}
		if dt != nil {
			destKind = dt.Kind()
		}
		if !pred(dt) {
			panic(fmt.Sprintf("operator %s not defined on: %v", escapedOpStr, destKind))
		}
		return
	}

	// other cases
	// TODO: XXX, this logic based on an assumption that one side is not defined on op
	// while the other side is ok, and this side is assignable to the other side, pass
	// is this general??? or the only case?
	// consider 1.2 % int64(1), and 1.0 % int64
	cmp := cmpSpecificity(lt, rt)
	if !pred(lt) { // lt not compatible with op
		if !pred(rt) { // rt not compatible with op
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
			// if cmp > 0, means potential conversion to
			// left side that is not compatible, so stop
			// the check here, assertion fail.
			if cmp < 0 {
				checkAssignableTo(lt, rt, true)
				debug.Println("---assignable")
				// cache, XXX, is this needed?
				AssignableCheckCache.Add(bx.Left, rt)
			} else {
				if lt != nil { // return error on left side that is checked first
					destKind = lt.Kind()
				}
				panic(fmt.Sprintf("operator %s not defined on: %v", escapedOpStr, destKind))
			}
		}
	} else if !pred(rt) { // if left is compatible, right is not
		// we don't need to check cmp == 0, for instance like 1 - "a",
		// xxx, the fact of one of them is compatible while the other is not
		// when they share same specificity implies not assignable?
		if cmp > 0 { // right to left
			checkAssignableTo(rt, lt, true)
			AssignableCheckCache.Add(bx.Right, lt)
		} else {
			if rt != nil { // return error on left side that is checked first
				destKind = rt.Kind()
			}
			panic(fmt.Sprintf("operator %s not defined on: %v", escapedOpStr, destKind))
		}
	} else {
		// both good
		// TODO: both untyped or both typed
		if !isUntyped(lt) && !isUntyped(rt) { // in this stage, lt or rt maybe untyped, not converted yet
			// check when both typed
			if lt != nil && rt != nil { // XXX, this should already be excluded by previous pred check.
				// TODO: filter byte that has no typeID?
				if lt.TypeID() != rt.TypeID() {
					panic(fmt.Sprintf("invalid operation: mismatched types %v and %v \n", lt, rt))
				}
			}
		}
		// check even when both side compatible with op
		// it's ok since checkCompatibility only for lss, add, etc
		// that only with isOrdered
		if cmp <= 0 {
			checkAssignableTo(lt, rt, true)
		} else {
			checkAssignableTo(rt, lt, true)
		}
		// TODO: cache
	}
}

func (ux *UnaryExpr) AssertCompatible(uxt, dt Type) {
	debug.Printf("---AssertCompatible, ux: %v \n", ux)
	debug.Printf("---AssertCompatible, ux.X: %T \n", ux.X)

	var destKind interface{}

	// get left type and right type
	//t := evalStaticTypeOf(store, last, ux.X)
	// we can't check compatible with native types
	// at current stage, so leave it to checkOrConvertType
	// to secondary call this assert logic again
	if _, ok := uxt.(*NativeType); ok {
		debug.Println("---left native, return")
		return
	}

	// check compatible
	if pred, ok := unaryChecker[ux.Op]; ok {
		if dt == nil {
			dt = uxt
		}
		if !pred(dt) {
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
	debug.Printf("---AssertCompatible, st: %v \n", idst)
	debug.Printf("---AssertCompatible, st.X: %T \n", idst.X)
	debug.Printf("---AssertCompatible, st.Op: %T \n", idst.Op)

	var destKind interface{}

	// get left type and right type
	//t := evalStaticTypeOf(store, last, idst.X)

	// we can't check compatible with native types
	// at current stage, so leave it to checkOrConvertType
	// to secondary call this assert logic again
	if _, ok := t.(*NativeType); ok {
		debug.Println("---left native, return")
		return
	}

	// check compatible
	if pred, ok := IncDecStmtChecker[idst.Op]; ok {
		if !pred(t) {
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
	debug.Printf("---AssertCompatible, as len lhs: %T \n", len(as.Lhs))
	//debug.Printf("---AssertCompatible, as.X: %T \n", len(as.Lhs))
	debug.Printf("---AssertCompatible, as.Op: %v \n", as.Op)

	escapedOpStr := strings.Replace(wordTokenStrings[as.Op], "%", "%%", 1)

	var destKind interface{}

	// XXX, assume lhs length is same with of rhs
	// TODO: check call case?
	for i, x := range as.Lhs {
		lt := evalStaticTypeOf(store, last, x)
		rt := evalStaticTypeOf(store, last, as.Rhs[i])

		// we can't check compatible with native types
		// at current stage, so leave it to checkOrConvertType
		// to secondary call this assert logic again
		if lnt, ok := lt.(*NativeType); ok {
			if _, ok := go2GnoBaseType(lnt.Type).(PrimitiveType); ok {
				debug.Println("---left native, return")
				return
			}
		}

		if rnt, ok := rt.(*NativeType); ok {
			if _, ok := go2GnoBaseType(rnt.Type).(PrimitiveType); ok {
				debug.Println("---right native, return")
				return
			}
		}

		debug.Printf("AssertCompatible,lt: %v, rt: %v,op: %v \n", lt, rt, as.Op)

		// check compatible
		if as.Op != ASSIGN {
			if pred, ok := AssignStmtChecker[as.Op]; ok {
				if !pred(lt) {
					if lt != nil {
						destKind = lt.Kind()
					}
					panic(fmt.Sprintf("operator %s not defined on: %v", escapedOpStr, destKind))
				}
				switch as.Op {
				case ADD_ASSIGN, SUB_ASSIGN, MUL_ASSIGN, QUO_ASSIGN, REM_ASSIGN, BAND_ASSIGN, BOR_ASSIGN, BAND_NOT_ASSIGN, XOR_ASSIGN:
					// if both typed
					if !isUntyped(lt) && !isUntyped(rt) { // in this stage, lt or rt maybe untyped, not converted yet
						debug.Println("---both typed")
						if lt != nil && rt != nil {
							// TODO: filter byte that has no typeID?
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
		} else {
			// check assignable, if done check operand with op
			// XXX, always right -> left? see preprocess line1746
			// special case if rhs is(or contained) untyped shift,
			// type of RHS is determined by LHS, no simply check
			// assignable for this case.
			checkAssignableTo(rt, lt, false)
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

	if nt1, ok := t1.(*NativeType); ok {
		if nt1.Type.Kind() == reflect.Interface { // is interface
			if nt1.Type.NumMethod() == 0 { // empty interface
				return 1
			} else if t2 != nil && t2.Kind() != InterfaceKind {
				return 1
			}
		} else {
			t1 = go2GnoType(nt1.Type) // if not interface, convert to gno type
		}
	} else if nt2, ok := t2.(*NativeType); ok {
		if nt2.Type.Kind() == reflect.Interface { // is interface
			if nt2.Type.NumMethod() == 0 {
				return -1
			} else if t1 != nil && t1.Kind() != InterfaceKind {
				return -1
			}
		} else {
			t2 = go2GnoType(nt2.Type)
		}
	}

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
