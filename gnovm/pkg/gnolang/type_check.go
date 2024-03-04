package gnolang

import (
	"fmt"
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
		REM:      isIntNum, // in compile stage good for bigdec, that can be converted to int, `checkAssignable`
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
	debug.Printf("---isOrdered, t is %v \n", t)
	switch t := baseOf(t).(type) {
	case PrimitiveType:
		if t.predicate() != IsInvalid && t.predicate()&IsOrdered != 0 || t.predicate()&IsRune != 0 {
			debug.Println("is Ordered!")
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
// TODO: consider, do we need complex?
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

//func isIntOrFloat(t Type) bool {
//	switch t := baseOf(t).(type) {
//	case PrimitiveType:
//		if t.predicate() != IsInvalid && t.predicate()&IsIntOrFloat != 0 || t.predicate()&IsRune != 0 {
//			return true
//		}
//		return false
//	default:
//		return false
//	}
//}

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

type Checker struct {
}

var AC *AssignabilityCache

// check both sides since no aware of which side is dest type
// TODO: add check assignable, 1.0 % uint64(1) is valid as is assignable
// that lt not compatible but rt is compatible would be good
// things like this would fail: 	println(1.0 % 1)

// AssertCompatible works as a pre-check prior to checkOrConvertType()
// It checks expressions to ensure the compatibility between operands and operators.
// e.g. "a" << 1, the left hand operand is not compatible with <<, it will fail the check.
// Overall,it efficiently filters out incompatible expressions, stopping before the next
// checkOrConvertType() operation to optimize performance.

// dt is a special case for binary expr that the dest type for the
// lhs determined by outer context
func (bx *BinaryExpr) AssertCompatible(store Store, last BlockNode, dt Type) {
	debug.Printf("---AssertCompatible, bx: %v \n", bx)

	debug.Printf("---AssertCompatible, bx.Left: %T \n", bx.Left)
	debug.Printf("---AssertCompatible, bx.Right: %T \n", bx.Right)
	// get left type and right type
	lt := evalStaticTypeOf(store, last, bx.Left)
	rt := evalStaticTypeOf(store, last, bx.Right)

	// we can't check compatible with native types
	// at current stage, so leave it to checkOrConvertType
	// to secondary call this assert logic again
	if _, ok := lt.(*NativeType); ok {
		debug.Println("---left native, return")
		return
	}
	if _, ok := rt.(*NativeType); ok {
		debug.Println("---right native, return")
		return
	}

	debug.Printf("AssertCompatible,lt: %v, rt: %v,op: %v \n", lt, rt, bx.Op)
	escapedOpStr := strings.Replace(wordTokenStrings[bx.Op], "%", "%%", 1)

	if isComparison(bx.Op) {
		switch bx.Op {
		case EQL, NEQ:
			assertComparable(lt, rt)
		case LSS, LEQ, GTR, GEQ:
			if pred, ok := binaryChecker[bx.Op]; ok {
				bx.assertCompatible2(lt, rt, pred, escapedOpStr, dt)
			} else {
				panic("should not happen")
			}
		default:
			panic("invalid comparison operator")
		}
	} else {
		if pred, ok := binaryChecker[bx.Op]; ok {
			bx.assertCompatible2(lt, rt, pred, escapedOpStr, dt)
		} else {
			panic("should not happen")
		}

		switch bx.Op {
		case ADD, SUB, MUL, QUO, REM, BAND, BOR, BAND_NOT, XOR, LAND, LOR:
			// if both typed
			if !isUntyped(lt) && !isUntyped(rt) {
				if lt != nil && rt != nil { // NOTE: is this necessary?
					if lt.TypeID() != rt.TypeID() {
						panic(fmt.Sprintf("invalid operation: mismatched types %v and %v \n", lt, rt))
					}
				}
			}
		default:
			// do nothing
		}
	}
}

// TODO: turn into method of bx
func (bx *BinaryExpr) assertCompatible2(lt, rt Type, pred func(t Type) bool, escapedOpStr string, dt Type) {
	debug.Println("---assertCompatible2, op: ", bx.Op)
	debug.Printf("---assertCompatible2, lt: %v, rt: %v \n", lt, rt)
	debug.Printf("---assertCompatible2, dt: %v \n", dt)
	AC = NewAssignabilityCache()
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
	cmp := cmpSpecificity(lt, rt)
	if !pred(lt) { // lt not compatible with op
		if !pred(rt) { // rt not compatible with op
			if lt != nil { // return error on left side that is checked first
				destKind = lt.Kind()
			}
			panic(fmt.Sprintf("operator %s not defined on: %v", escapedOpStr, destKind))
		} else {
			debug.Println("---2")
			debug.Println("---cmp: ", cmp)
			// left not compatible, right is compatible
			// cmp means the expected convert direction
			// if cmp < 0, means potential conversion
			// from left to right, so check assignable
			// if cmp > 0, means potential conversion to
			// left side that is not compatible, so stop
			// the check here, assertion fail.
			if cmp < 0 {
				checkAssignable(lt, rt, true)
				debug.Println("---assignable")
				// TODO: set attr
				AC.cache[ExprTypePair{X: bx.Left, T: rt}] = true
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
			checkAssignable(rt, lt, true)
		} else {
			if rt != nil { // return error on left side that is checked first
				destKind = rt.Kind()
			}
			panic(fmt.Sprintf("operator %s not defined on: %v", escapedOpStr, destKind))
		}
	} else {
		// both good
		debug.Println("---both good")
	}
}

func (ux *UnaryExpr) AssertCompatible(store Store, last BlockNode, dt Type) {
	debug.Printf("---AssertCompatible, ux: %v \n", ux)
	debug.Printf("---AssertCompatible, ux.X: %T \n", ux.X)

	var destKind interface{}

	// get left type and right type
	t := evalStaticTypeOf(store, last, ux.X)
	// we can't check compatible with native types
	// at current stage, so leave it to checkOrConvertType
	// to secondary call this assert logic again
	if _, ok := t.(*NativeType); ok {
		debug.Println("---left native, return")
		return
	}

	// check compatible
	if pred, ok := unaryChecker[ux.Op]; ok {
		if dt == nil {
			dt = t
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

func (idst *IncDecStmt) AssertCompatible(store Store, last BlockNode) {
	debug.Printf("---AssertCompatible, st: %v \n", idst)
	debug.Printf("---AssertCompatible, st.X: %T \n", idst.X)
	debug.Printf("---AssertCompatible, st.Op: %T \n", idst.Op)

	var destKind interface{}

	// get left type and right type
	t := evalStaticTypeOf(store, last, idst.X)

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
		if _, ok := lt.(*NativeType); ok {
			debug.Println("---left native, return")
			return
		}

		if _, ok := rt.(*NativeType); ok {
			debug.Println("---right native, return")
			return
		}

		debug.Printf("AssertCompatible,lt: %v, rt: %v,op: %v \n", lt, rt, as.Op)

		// check compatible
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
	}
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
