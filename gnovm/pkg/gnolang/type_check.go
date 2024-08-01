package gnolang

import (
	"fmt"
	"reflect"

	"github.com/gnolang/gno/tm2/pkg/errors"
)

// here are a range of rules predefined for preprocessor to check the compatibility between operands and operators
// e,g. for binary expr x + y, x, y can only be numeric or string, 1+2, "a" + "b"
// this is used in assertCompatible()s.
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
		return t.category()&IsOrdered != 0
	default:
		return false
	}
}

func isBoolean(t Type) bool {
	switch t := baseOf(t).(type) {
	case PrimitiveType:
		return t.category()&IsBoolean != 0
	default:
		return false
	}
}

// rune can be numeric and string
func isNumeric(t Type) bool {
	switch t := baseOf(t).(type) {
	case PrimitiveType:
		return t.category()&IsNumeric != 0
	default:
		return false
	}
}

func isIntNum(t Type) bool {
	switch t := baseOf(t).(type) {
	case PrimitiveType:
		return t.category()&IsInteger != 0 || t.category()&IsBigInt != 0
	default:
		return false
	}
}

func isNumericOrString(t Type) bool {
	switch t := baseOf(t).(type) {
	case PrimitiveType:
		return t.category()&IsNumeric != 0 || t.category()&IsString != 0
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
		debug.Printf("assertComparable2 dt: %v \n", dt)
	}
	switch cdt := baseOf(dt).(type) {
	case PrimitiveType:
	case *ArrayType:
		switch baseOf(cdt.Elem()).(type) {
		case PrimitiveType, *PointerType, *InterfaceType, *NativeType, *ArrayType, *StructType, *ChanType:
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
	case *SliceType, *FuncType, *MapType, *InterfaceType, *PointerType, *ChanType: //  we don't have unsafePointer
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

func checkSame(at, bt Type, msg string) error {
	if debug {
		debug.Printf("checkSame, at: %v bt: %v \n", at, bt)
	}
	if at.TypeID() != bt.TypeID() {
		return errors.New("incompatible types %v and %v %s",
			at.TypeID(), bt.TypeID(), msg)
	}
	return nil
}

func assertAssignableTo(xt, dt Type, autoNative bool) {
	err := checkAssignableTo(xt, dt, autoNative)
	if err != nil {
		panic(err.Error())
	}
}

// TODO: check operands unary & binary expr are const
func checkValConstType(d *ValueDecl) {
	for _, vx := range d.Values {
		switch vx.(type) {
		case *BasicLitExpr, *BinaryExpr, *UnaryExpr:
			// Valid constant expression
			break
		default:
			panic("const type should be a basic type")
		}
	}
}

// checkValDefineMismatch checks for mismatch between the number of variables and values in a ValueDecl or AssignStmt.
func checkValDefineMismatch(n Node) {
	var (
		valueDecl *ValueDecl
		assign    *AssignStmt
		values    []Expr
		numNames  int
		numValues int
	)

	switch x := n.(type) {
	case *ValueDecl:
		valueDecl = x
		numNames = len(valueDecl.NameExprs)
		numValues = len(valueDecl.Values)
		values = valueDecl.Values
	case *AssignStmt:
		if x.Op != DEFINE {
			return
		}

		assign = x
		numNames = len(assign.Lhs)
		numValues = len(assign.Rhs)
		values = assign.Rhs
	default:
		panic(fmt.Sprintf("unexpected node type %T", n))
	}

	if numValues == 0 || numValues == numNames {
		return
	}

	// Special case for single value.
	// If the value is a call expression, type assertion, or index expression,
	// it can be assigned to multiple variables.
	if numValues == 1 {
		switch values[0].(type) {
		case *CallExpr:
			return
		case *TypeAssertExpr:
			if numNames != 2 {
				panic(fmt.Sprintf("assignment mismatch: %d variable(s) but %d value(s)", numNames, numValues))
			}
			return
		case *IndexExpr:
			if numNames != 2 {
				panic(fmt.Sprintf("assignment mismatch: %d variable(s) but %d value(s)", numNames, numValues))
			}
			return
		}
	}

	if valueDecl != nil {
		if numNames > numValues {
			panic(fmt.Sprintf("missing init expr for %s", valueDecl.NameExprs[numValues].String()))
		}

		panic(fmt.Sprintf("extra init expr %s", values[numNames].String()))
	}

	panic(fmt.Sprintf("assignment mismatch: %d variable(s) but %d value(s)", numNames, numValues))
}

// Assert that xt can be assigned as dt (dest type).
// If autoNative is true, a broad range of xt can match against
// a target native dt type, if and only if dt is a native type.
func checkAssignableTo(xt, dt Type, autoNative bool) error {
	if debug {
		debug.Printf("checkAssignableTo, xt: %v dt: %v \n", xt, dt)
	}
	// case0
	if xt == nil { // see test/files/types/eql_0f18
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
			} else if err := idt.VerifyImplementedBy(xt); err == nil {
				// if dt implements idt, ok.
				return nil // ok
			} else {
				return errors.New(
					"%s does not implement %s (%s)",
					xt.String(),
					dt.String(),
					err.Error())
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
					return errors.New(
						"cannot use %s as %s",
						nxt.String(),
						nidt.String())
				}
			} else if pxt, ok := baseOf(xt).(*PointerType); ok {
				nxt, ok := pxt.Elt.(*NativeType)
				if !ok {
					return errors.New(
						"pointer to non-native type cannot satisfy non-empty native interface; %s doesn't implement %s",
						pxt.String(),
						nidt.String())
				}
				// if xt has native base, do the naive native.
				if reflect.PtrTo(nxt.Type).AssignableTo(nidt) {
					return nil // ok
				} else {
					return errors.New(
						"cannot use %s as %s",
						pxt.String(),
						nidt.String())
				}
			} else if xdt, ok := xt.(*DeclaredType); ok {
				if gno2GoTypeMatches(baseOf(xdt), ndt.Type) {
					return nil
				} // not check against native interface
			} else {
				return errors.New(
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
				return errors.New(
					"cannot use %s as %s without explicit conversion",
					dxt.String(),
					ddt.String())
			}
		} else {
			// special case if implicitly named primitive type.
			// TODO simplify with .IsNamedType().
			if _, ok := dt.(PrimitiveType); ok {
				return errors.New(
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
				return errors.New(
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
				return errors.New(
					"cannot use untyped bool as %s",
					dt.Kind())
			}
		case UntypedStringType:
			if dt.Kind() == StringKind {
				return nil // ok
			} else {
				return errors.New(
					"cannot use untyped string as %s",
					dt.Kind())
			}
		// XXX, this is a loose check, we don't have the context
		// to check if it is an exact integer, e.g. 1.2 or 1.0(1.0 can be converted to int).
		// this ensure expr like (a % 1.0) pass check, while
		// expr like (a % 1.2) panic at ConvertUntypedTo, which is a delayed assertion after const evaluated.
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
				return errors.New(
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
				return errors.New(
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
			return checkAssignableTo(pt.Elt, cdt.Elt, false)
		}
	case *ArrayType:
		if at, ok := xt.(*ArrayType); ok {
			if at.Len != cdt.Len {
				return errors.New(
					"cannot use %s as %s",
					at.String(),
					cdt.String())
			}
			err := checkSame(at.Elt, cdt.Elt, "")
			if err != nil {
				return errors.New(
					"cannot use %s as %s",
					at.String(),
					cdt.String())
			}
			return nil
		}
	case *SliceType:
		if st, ok := xt.(*SliceType); ok {
			if cdt.Vrd {
				return checkAssignableTo(st.Elt, cdt.Elt, false)
			} else {
				err := checkSame(st.Elt, cdt.Elt, "")
				if err != nil {
					return errors.New(
						"cannot use %s as %s",
						st.String(),
						cdt.String())
				}
				return nil
			}
		}
	case *MapType:
		if mt, ok := xt.(*MapType); ok {
			err := checkSame(mt.Key, cdt.Key, "")
			if err != nil {
				return errors.New(
					"cannot use %s as %s",
					mt.String(),
					cdt.String()).Stacktrace()
			}
			return nil
		}
	case *InterfaceType:
		return errors.New("should not happen")
	case *DeclaredType:
		panic("should not happen")
	case *FuncType, *StructType, *PackageType, *ChanType, *TypeType:
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
		return errors.New(
			"unexpected type %s",
			dt.String())
	}
	return errors.New(
		"cannot use %s as %s",
		xt.String(),
		dt.String()).Stacktrace()
}

// ===========================================================
func (x *BinaryExpr) checkShiftLhs(dt Type) {
	if checker, ok := binaryChecker[x.Op]; ok {
		if !checker(dt) {
			panic(fmt.Sprintf("operator %s not defined on: %v", x.Op.TokenString(), kindString(dt)))
		}
	} else {
		panic(fmt.Sprintf("checker for %s does not exist", x.Op))
	}
}

// AssertCompatible works as a pre-check prior to checkOrConvertType.
// It checks against expressions to ensure the compatibility between operands and operators.
// e.g. "a" << 1, the left hand operand is not compatible with <<, it will fail the check.
// Overall,it efficiently filters out incompatible expressions, stopping before the next
// checkOrConvertType() operation to optimize performance.
func (x *BinaryExpr) AssertCompatible(lt, rt Type) {
	// native type will be converted to gno in latter logic,
	// this check logic will be conduct again from trans_leave *BinaryExpr.
	lnt, lin := lt.(*NativeType)
	rnt, rin := rt.(*NativeType)
	if lin && rin {
		if lt.TypeID() != rt.TypeID() {
			panic(fmt.Sprintf(
				"incompatible operands in binary expression: %s %s %s",
				lt.TypeID(), x.Op, rt.TypeID()))
		}
	}
	if lin {
		if _, ok := go2GnoBaseType(lnt.Type).(PrimitiveType); ok {
			return
		}
	}
	if rin {
		if _, ok := go2GnoBaseType(rnt.Type).(PrimitiveType); ok {
			return
		}
	}

	xt, dt := lt, rt
	if shouldSwapOnSpecificity(lt, rt) {
		xt, dt = dt, xt
	}

	if isComparison(x.Op) {
		switch x.Op {
		case EQL, NEQ:
			assertComparable(xt, dt)
			if !isUntyped(xt) && !isUntyped(dt) {
				assertAssignableTo(xt, dt, false)
			}
		case LSS, LEQ, GTR, GEQ:
			if checker, ok := binaryChecker[x.Op]; ok {
				x.checkCompatibility(xt, dt, checker, x.Op.TokenString())
			} else {
				panic(fmt.Sprintf("checker for %s does not exist", x.Op))
			}
		default:
			panic("invalid comparison operator")
		}
	} else {
		if checker, ok := binaryChecker[x.Op]; ok {
			x.checkCompatibility(xt, dt, checker, x.Op.TokenString())
		} else {
			panic(fmt.Sprintf("checker for %s does not exist", x.Op))
		}

		switch x.Op {
		case QUO, REM:
			// special case of zero divisor
			if isQuoOrRem(x.Op) {
				if rcx, ok := x.Right.(*ConstExpr); ok {
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

// Check compatibility of the destination type (dt) with the operator.
// If both source type (xt) and destination type (dt) are typed:
// Verify that xt is assignable to dt.
// If xt is untyped:
// The function checkOrConvertType will be invoked after this check.
// NOTE: dt is established based on a specificity check between xt and dt,
// confirming dt as the appropriate destination type for this context.
func (x *BinaryExpr) checkCompatibility(xt, dt Type, checker func(t Type) bool, OpStr string) {
	if !checker(dt) {
		panic(fmt.Sprintf("operator %s not defined on: %v", OpStr, kindString(dt)))
	}

	// if both typed
	if !isUntyped(xt) && !isUntyped(dt) {
		err := checkAssignableTo(xt, dt, false)
		if err != nil {
			panic(fmt.Sprintf("invalid operation: mismatched types %v and %v", xt, dt))
		}
	}
}

func (x *UnaryExpr) AssertCompatible(t Type) {
	if nt, ok := t.(*NativeType); ok {
		if _, ok := go2GnoBaseType(nt.Type).(PrimitiveType); ok {
			return
		}
	}
	// check compatible
	if checker, ok := unaryChecker[x.Op]; ok {
		if !checker(t) {
			panic(fmt.Sprintf("operator %s not defined on: %v", x.Op.TokenString(), kindString(t)))
		}
	} else {
		panic(fmt.Sprintf("checker for %s does not exist", x.Op))
	}
}

func (x *IncDecStmt) AssertCompatible(t Type) {
	if nt, ok := t.(*NativeType); ok {
		if _, ok := go2GnoBaseType(nt.Type).(PrimitiveType); ok {
			return
		}
	}
	// check compatible
	if checker, ok := IncDecStmtChecker[x.Op]; ok {
		if !checker(t) {
			panic(fmt.Sprintf("operator %s not defined on: %v", x.Op.TokenString(), kindString(t)))
		}
	} else {
		panic(fmt.Sprintf("checker for %s does not exist", x.Op))
	}
}

func assertIndexTypeIsInt(kt Type) {
	if kt.Kind() != IntKind {
		panic(fmt.Sprintf("index type should be int, but got %v", kt))
	}
}

func (x *RangeStmt) AssertCompatible(store Store, last BlockNode) {
	if x.Op != ASSIGN {
		return
	}
	if isBlankIdentifier(x.Key) && isBlankIdentifier(x.Value) {
		// both "_"
		return
	}
	assertValidAssignLhs(store, last, x.Key)
	// if is valid left value

	kt := evalStaticTypeOf(store, last, x.Key)
	var vt Type
	if x.Value != nil {
		vt = evalStaticTypeOf(store, last, x.Value)
	}

	xt := evalStaticTypeOf(store, last, x.X)
	switch cxt := xt.(type) {
	case *MapType:
		assertAssignableTo(cxt.Key, kt, false)
		if vt != nil {
			assertAssignableTo(cxt.Value, vt, false)
		}
	case *SliceType:
		assertIndexTypeIsInt(kt)
		if vt != nil {
			assertAssignableTo(cxt.Elt, vt, false)
		}
	case *ArrayType:
		assertIndexTypeIsInt(kt)
		if vt != nil {
			assertAssignableTo(cxt.Elt, vt, false)
		}
	case PrimitiveType:
		if cxt.Kind() == StringKind {
			if kt != nil && kt.Kind() != IntKind {
				panic(fmt.Sprintf("index type should be int, but got %v", kt))
			}
			if vt != nil {
				if vt.Kind() != Int32Kind { // rune
					panic(fmt.Sprintf("value type should be int32, but got %v", kt))
				}
			}
		}
	}
}

func (x *AssignStmt) AssertCompatible(store Store, last BlockNode) {
	if x.Op == ASSIGN || x.Op == DEFINE {
		if len(x.Lhs) > len(x.Rhs) {
			if len(x.Rhs) != 1 {
				panic(fmt.Sprintf("assignment mismatch: %d variables but %d values", len(x.Lhs), len(x.Rhs)))
			}
			switch cx := x.Rhs[0].(type) {
			case *CallExpr:
				// Call case: a, b = x(...)
				ift := evalStaticTypeOf(store, last, cx.Func)
				cft := getGnoFuncTypeOf(store, ift)
				if len(x.Lhs) != len(cft.Results) {
					panic(fmt.Sprintf(
						"assignment mismatch: "+
							"%d variables but %s returns %d values",
						len(x.Lhs), cx.Func.String(), len(cft.Results)))
				}
				if x.Op == ASSIGN {
					// check assignable
					for i, lx := range x.Lhs {
						if !isBlankIdentifier(lx) {
							assertValidAssignLhs(store, last, lx)
							lxt := evalStaticTypeOf(store, last, lx)
							assertAssignableTo(cft.Results[i].Type, lxt, false)
						}
					}
				}
			case *TypeAssertExpr:
				// Type-assert case: a, ok := x.(type)
				if len(x.Lhs) != 2 {
					panic("should not happen")
				}
				if x.Op == ASSIGN {
					// check assignable to first value
					if !isBlankIdentifier(x.Lhs[0]) { // see composite3.gno
						assertValidAssignLhs(store, last, x.Lhs[0])
						dt := evalStaticTypeOf(store, last, x.Lhs[0])
						ift := evalStaticTypeOf(store, last, cx)
						assertAssignableTo(ift, dt, false)
					}
					if !isBlankIdentifier(x.Lhs[1]) { // see composite3.gno
						assertValidAssignLhs(store, last, x.Lhs[1])
						dt := evalStaticTypeOf(store, last, x.Lhs[1])
						if dt.Kind() != BoolKind { // typed, not bool
							panic(fmt.Sprintf("want bool type got %v", dt))
						}
					}
				}
				cx.HasOK = true
			case *IndexExpr: // must be with map type when len(Lhs) > len(Rhs)
				if len(x.Lhs) != 2 {
					panic("should not happen")
				}
				if x.Op == ASSIGN {
					if !isBlankIdentifier(x.Lhs[0]) {
						assertValidAssignLhs(store, last, x.Lhs[0])
						lt := evalStaticTypeOf(store, last, x.Lhs[0])
						if _, ok := cx.X.(*NameExpr); ok {
							rt := evalStaticTypeOf(store, last, cx.X)
							if mt, ok := rt.(*MapType); ok {
								assertAssignableTo(mt.Value, lt, false)
							}
						} else if _, ok := cx.X.(*CompositeLitExpr); ok {
							cpt := evalStaticTypeOf(store, last, cx.X)
							if mt, ok := cpt.(*MapType); ok {
								assertAssignableTo(mt.Value, lt, false)
							} else {
								panic("should not happen")
							}
						}
					}
					if !isBlankIdentifier(x.Lhs[1]) {
						assertValidAssignLhs(store, last, x.Lhs[1])
						dt := evalStaticTypeOf(store, last, x.Lhs[1])
						if dt != nil && dt.Kind() != BoolKind { // typed, not bool
							panic(fmt.Sprintf("want bool type got %v", dt))
						}
					}
				}
				cx.HasOK = true
			default:
				panic(fmt.Sprintf("RHS should not be %v when len(Lhs) > len(Rhs)", cx))
			}
		} else { // len(Lhs) == len(Rhs)
			if x.Op == ASSIGN {
				// assert valid left value
				for _, lx := range x.Lhs {
					assertValidAssignLhs(store, last, lx)
				}
			}
		}
	} else { // Ops other than assign and define
		// If this is an assignment operation, ensure there's only 1
		// expr on lhs/rhs.
		if len(x.Lhs) != 1 || len(x.Rhs) != 1 {
			panic("assignment operator " + x.Op.TokenString() +
				" requires only one expression on lhs and rhs")
		}
		for i, lx := range x.Lhs {
			lt := evalStaticTypeOf(store, last, lx)
			rt := evalStaticTypeOf(store, last, x.Rhs[i])

			if checker, ok := AssignStmtChecker[x.Op]; ok {
				if !checker(lt) {
					panic(fmt.Sprintf("operator %s not defined on: %v", x.Op.TokenString(), kindString(lt)))
				}
				switch x.Op {
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
				panic(fmt.Sprintf("checker for %s does not exist", x.Op))
			}
		}
	}
}

// misc
func assertValidAssignLhs(store Store, last BlockNode, lx Expr) {
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

// shouldSwapOnSpecificity determines the potential direction for
// checkOrConvertType. it checks whether a swap is needed between two types
// based on their specificity. If t2 has a lower specificity than t1, it returns
// false, indicating no swap is needed. If t1 has a lower specificity than t2,
// it returns true, indicating a swap is needed.
func shouldSwapOnSpecificity(t1, t2 Type) bool {
	// check nil
	if t1 == nil { // see test file 0f46
		return false // also with both nil
	} else if t2 == nil {
		return true
	}

	// check interface
	if it1, ok := baseOf(t1).(*InterfaceType); ok {
		if it1.IsEmptyInterface() {
			return true // left empty interface
		} else {
			if _, ok := baseOf(t2).(*InterfaceType); ok {
				return false
			} else {
				return true // right not interface
			}
		}
	} else if _, ok := t2.(*InterfaceType); ok {
		return false // left not interface, right is interface
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
		return true
	} else {
		return false
	}
}

func isBlankIdentifier(x Expr) bool {
	if nx, ok := x.(*NameExpr); ok {
		return nx.Name == "_"
	}
	return false
}
