package gnolang

import "fmt"

var (
	binaryPredicates = map[Word]func(t Type) bool{
		ADD:      isNumericOrString,
		SUB:      isNumeric,
		MUL:      isNumeric,
		QUO:      isNumeric,
		REM:      isIntNum,
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
	unaryPredicates = map[Word]func(t Type) bool{
		ADD: isNumeric,
		SUB: isNumeric,
		XOR: isIntNum,
		NOT: isBoolean,
	}
	IncDecStmtPredicates = map[Word]func(t Type) bool{ // NOTE: to be consistent with op_inc_dec.go, line3, no float support for now(while go does).
		INC: isNumeric,
		DEC: isNumeric,
	}
	AssignStmtPredicates = map[Word]func(t Type) bool{
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

	IsNumeric    = IsInteger | IsUnsigned | IsFloat | IsBigInt | IsBigDec
	IsOrdered    = IsNumeric | IsString
	IsIntOrFloat = IsInteger | IsUnsigned | IsFloat | IsBigInt | IsBigDec
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

func isIntOrFloat(t Type) bool {
	switch t := baseOf(t).(type) {
	case PrimitiveType:
		if t.predicate() != IsInvalid && t.predicate()&IsIntOrFloat != 0 || t.predicate()&IsRune != 0 {
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
