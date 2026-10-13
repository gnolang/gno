package gnolang

import (
	"fmt"

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
		REM:      isWhole,
		SHL:      isWhole,
		SHR:      isWhole,
		BAND:     isWhole, // bit ops
		XOR:      isWhole,
		BOR:      isWhole,
		BAND_NOT: isWhole,
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
		XOR: isWhole,
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
		REM_ASSIGN:      isWhole,
		SHL_ASSIGN:      isWhole,
		SHR_ASSIGN:      isWhole,
		BAND_ASSIGN:     isWhole,
		XOR_ASSIGN:      isWhole,
		BOR_ASSIGN:      isWhole,
		BAND_NOT_ASSIGN: isWhole,
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
	IsWhole   = IsInteger | IsBigInt
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

func isNumericOrString(t Type) bool {
	switch t := baseOf(t).(type) {
	case PrimitiveType:
		return t.category()&IsNumeric != 0 || t.category()&IsString != 0
	default:
		return false
	}
}

func isInteger(t Type) bool {
	switch t := baseOf(t).(type) {
	case PrimitiveType:
		return t.category()&IsInteger != 0
	default:
		return false
	}
}

func isWhole(t Type) bool {
	switch t := baseOf(t).(type) {
	case PrimitiveType:
		return t.category()&IsWhole != 0
	default:
		return false
	}
}

func isIntegerKind(k Kind) bool {
	switch k {
	case IntKind, Int8Kind, Int16Kind, Int32Kind, Int64Kind,
		UintKind, Uint8Kind, Uint16Kind, Uint32Kind, Uint64Kind:
		return true
	default:
		return false
	}
}

// ===========================================================

func mayBeNil(t Type) bool {
	switch baseOf(t).(type) {
	case *SliceType, *FuncType, *MapType, *InterfaceType, *PointerType: //  we don't have unsafePointer
		return true
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

func kindString(xt Type) string {
	if xt != nil {
		return xt.Kind().String()
	}
	return "nil"
}

func isBlankIdentifier(x Expr) bool {
	if nx, ok := x.(*NameExpr); ok {
		return nx.Name == blankIdentifier
	}
	return false
}

// isNilIdentifier reports whether x is the bare `nil` keyword. By check time
// `nil` is always const-folded, so the original syntax is recovered from the
// ConstExpr's Source; a conversion like `T(nil)` also has a nil static type
// but a non-NameExpr Source, and reports false.
func isNilIdentifier(x Expr) bool {
	switch cx := x.(type) {
	case *ConstExpr:
		nx, ok := cx.Source.(*NameExpr)
		return ok && nx.Name == "nil"
	case *NameExpr:
		if cx.Name == "nil" {
			panic("should not happen: unfolded nil NameExpr in isNilIdentifier")
		}
		return false
	default:
		return false
	}
}

// isComparable returns true if the type can be compared with ==.
// This is used for map key validation and other comparability checks.
func isComparable(dt Type) bool {
	switch cdt := baseOf(dt).(type) {
	case PrimitiveType, *PointerType, *InterfaceType:
		return true
	case *ChanType:
		panic("channel type is not yet supported")
	case *ArrayType:
		return isComparable(cdt.Elt)
	case *StructType:
		// Memoized: a struct fans out into all its fields, so without a
		// cache an interface comparison whose dynamic type is a nested
		// struct re-walks the field graph exponentially on every compare.
		// Cycles always pass through a pointer/slice/map/interface leaf
		// (a direct value cycle is infinite-size and rejected at compile
		// time), so the recursion bottoms out before re-entering a struct
		// mid-computation; the write is unconditionally safe.
		if cdt.comparable != 0 {
			return cdt.comparable == 1
		}
		res := uint8(1)
		for _, f := range cdt.Fields {
			if !isComparable(f.Type) {
				res = 2
				break
			}
		}
		cdt.comparable = res
		return res == 1
	default:
		return false
	}
}
