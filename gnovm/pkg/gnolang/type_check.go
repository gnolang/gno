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

// check both sides since no aware of which side is dest type
// TODO: add check assignable, 1.0 % uint64(1) is valid as is assignable
// that lt not compatible but rt is compatible would be good
// things like this would fail: 	println(1.0 % 1)

func (bx *BinaryExpr) AssertCompatible(store Store, last BlockNode) {
	debug.Printf("---AssertCompatible, bx: %v \n", bx)

	debug.Printf("---AssertCompatible, bx.Left: %T \n", bx.Left)
	debug.Printf("---AssertCompatible, bx.Right: %T \n", bx.Right)

	// get left type and right type
	var lt, rt Type
	if lx, ok := (bx.Left).(*ConstExpr); ok {
		lt = lx.T
	} else if bx.Left != nil {
		lt = evalStaticTypeOf(store, last, bx.Left)
	}

	if rx, ok := (bx.Right).(*ConstExpr); ok {
		rt = rx.T
	} else if bx.Left != nil {
		rt = evalStaticTypeOf(store, last, bx.Right)
	}

	debug.Printf("AssertCompatible,lt: %v, rt: %v,op: %v \n", lt, rt, bx.Op)
	escapedOpStr := strings.Replace(wordTokenStrings[bx.Op], "%", "%%", 1)
	if isComparison(bx.Op) {
		switch bx.Op {
		case EQL, NEQ:
			assertComparable(lt, rt)
		case LSS, LEQ, GTR, GEQ:
			if pred, ok := binaryChecker[bx.Op]; ok {
				assertCompatible2(lt, rt, pred, escapedOpStr)
			} else {
				panic("should not happen")
			}
		default:
			panic("invalid comparison operator")
		}
	} else {
		if pred, ok := binaryChecker[bx.Op]; ok {
			assertCompatible2(lt, rt, pred, escapedOpStr)
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

func assertCompatible2(lt, rt Type, pred func(t Type) bool, escapedOpStr string) {
	var destKind interface{}
	cmp := cmpSpecificity(lt, rt)
	if !pred(lt) { // lt not compatible with op
		if !pred(rt) { // rt not compatible with op
			if lt != nil { // return error on left side that is checked first
				destKind = lt.Kind()
			}
			panic(fmt.Sprintf("operator %s not defined on: %v", escapedOpStr, destKind))
		} else {
			debug.Println("---2")
			// left not compatible, right is compatible
			// cmp means the expected convert direction
			// if cmp < 0, means potential conversion
			// from left to right, so check assignable
			// if cmp > 0, means potential conversion to
			// left side that is not compatible, so stop
			// the check here, assertion fail.
			if cmp < 0 {
				checkAssignable(lt, rt, false)
				debug.Println("---assignable")
				// TODO: set attr
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
			checkAssignable(rt, lt, false)
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
