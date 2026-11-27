package gnolang

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
)

// constShiftBound is the largest shift count go/types accepts for a constant
// shift (1023-1+52, so the smallest float64 stays expressible). A wider count
// is an invalid shift, not a typed overflow, and stays on the existing path.
const constShiftBound = 1023 - 1 + 52

// typedIntConstOverflow reports whether x is a typed integer constant
// operation whose mathematical result does not fit in the result type.
//
// The check runs before evalConst hands the expression to the machine.
// Fitting results still take that path: the machine's wrap equals the exact
// value whenever the value is representable, and preprocess gas stays on the
// meter the machine already charges. Overflow panics here, with gc's
// wording, instead of wrapping. Untyped integers, floats, and non-arithmetic
// operators are left alone (the second result is false).
func typedIntConstOverflow(x Expr) (msg string, overflow bool) {
	switch e := x.(type) {
	case *UnaryExpr:
		if e.Op != SUB {
			return "", false
		}
		cx, ok := e.X.(*ConstExpr)
		if !ok {
			return "", false
		}
		return unaryTypedIntOverflow(e, cx)
	case *BinaryExpr:
		lx, lok := e.Left.(*ConstExpr)
		rx, rok := e.Right.(*ConstExpr)
		if !lok || !rok {
			return "", false
		}
		return binaryTypedIntOverflow(e, lx, rx)
	default:
		return "", false
	}
}

func unaryTypedIntOverflow(e *UnaryExpr, cx *ConstExpr) (string, bool) {
	base, ok := typedIntBase(cx.T)
	if !ok {
		return "", false
	}
	v, ok := integerConstBig(&cx.TypedValue)
	if !ok {
		return "", false
	}
	res := new(big.Int).Neg(v)
	if intFits(res, base) {
		return "", false
	}
	return overflowMsg(e, res, cx.T), true
}

func binaryTypedIntOverflow(e *BinaryExpr, lx, rx *ConstExpr) (string, bool) {
	resultT := typedIntResultType(e.Op, lx.T, rx.T)
	base, ok := typedIntBase(resultT)
	if !ok {
		return "", false
	}
	lb, lok := integerConstBig(&lx.TypedValue)
	rb, rok := integerConstBig(&rx.TypedValue)
	if !lok || !rok {
		return "", false
	}
	res, ok := exactTypedInt(e.Op, lb, rb, base)
	if !ok {
		return "", false
	}
	if intFits(res, base) {
		return "", false
	}
	return overflowMsg(e, res, resultT), true
}

// typedIntResultType is the type of a constant integer operation.
// A shift's type is its left operand. Other arithmetic takes a typed
// operand over an untyped one.
func typedIntResultType(op Word, lt, rt Type) Type {
	if op == SHL || op == SHR {
		return lt
	}
	if _, ok := typedIntBase(lt); ok {
		return lt
	}
	return rt
}

// exactTypedInt folds op on arbitrary-precision integers.
// ok is false when op is not typed-integer arithmetic or the shift count
// is not a constant this checker owns (negative, or above constShiftBound).
func exactTypedInt(op Word, lb, rb *big.Int, base PrimitiveType) (res *big.Int, ok bool) {
	switch op {
	case ADD:
		return new(big.Int).Add(lb, rb), true
	case SUB:
		return new(big.Int).Sub(lb, rb), true
	case MUL:
		return new(big.Int).Mul(lb, rb), true
	case QUO:
		if rb.Sign() == 0 {
			return nil, false // division by zero is reported elsewhere
		}
		return new(big.Int).Quo(lb, rb), true
	case SHL, SHR:
		if rb.Sign() < 0 || !rb.IsUint64() {
			return nil, false
		}
		n := rb.Uint64()
		if n > constShiftBound {
			return nil, false
		}
		return exactShift(op, lb, n, integerBits(base)), true
	default:
		return nil, false
	}
}

func exactShift(op Word, v *big.Int, n uint64, bits int) *big.Int {
	if n == 0 || v.Sign() == 0 {
		return new(big.Int).Set(v)
	}
	if op == SHR {
		// Arithmetic right shift. At or past the width the result is 0, or
		// -1 when the value is negative, and both fit the type.
		if bits > 0 && n >= uint64(bits) {
			if v.Sign() < 0 {
				return big.NewInt(-1)
			}
			return big.NewInt(0)
		}
		return new(big.Int).Rsh(v, uint(n))
	}
	// Left shift by at least the width cannot fit a non-zero value.
	// The mathematical value is still what the diagnostic names.
	return new(big.Int).Lsh(v, uint(n))
}

// typedIntBase reports the underlying fixed-width integer of a typed
// constant. Untyped integers (bigint, rune) are excluded: those fold
// exactly on the machine already, or are a different bug.
func typedIntBase(t Type) (PrimitiveType, bool) {
	if t == nil || isUntyped(t) {
		return 0, false
	}
	b, ok := baseOf(t).(PrimitiveType)
	if !ok {
		return 0, false
	}
	switch b {
	case IntType, Int8Type, Int16Type, Int32Type, Int64Type,
		UintType, Uint8Type, Uint16Type, Uint32Type, Uint64Type:
		return b, true
	default:
		return 0, false
	}
}

func integerBits(base PrimitiveType) int {
	switch base {
	case Int8Type, Uint8Type:
		return 8
	case Int16Type, Uint16Type:
		return 16
	case Int32Type, Uint32Type:
		return 32
	case Int64Type, Uint64Type:
		return 64
	case IntType, UintType:
		return strconv.IntSize
	default:
		return 0
	}
}

func integerConstBig(tv *TypedValue) (*big.Int, bool) {
	if tv == nil || tv.T == nil {
		return nil, false
	}
	switch baseOf(tv.T) {
	case IntType:
		return big.NewInt(tv.GetInt()), true
	case Int8Type:
		return big.NewInt(int64(tv.GetInt8())), true
	case Int16Type:
		return big.NewInt(int64(tv.GetInt16())), true
	case Int32Type, UntypedRuneType:
		return big.NewInt(int64(tv.GetInt32())), true
	case Int64Type:
		return big.NewInt(tv.GetInt64()), true
	case UintType:
		return new(big.Int).SetUint64(tv.GetUint()), true
	case Uint8Type:
		return new(big.Int).SetUint64(uint64(tv.GetUint8())), true
	case Uint16Type:
		return new(big.Int).SetUint64(uint64(tv.GetUint16())), true
	case Uint32Type:
		return new(big.Int).SetUint64(uint64(tv.GetUint32())), true
	case Uint64Type:
		return new(big.Int).SetUint64(tv.GetUint64()), true
	case UntypedBigintType:
		return new(big.Int).Set(tv.GetBigInt()), true
	default:
		return nil, false
	}
}

var (
	bigZero      = big.NewInt(0)
	bigMinInt8   = big.NewInt(math.MinInt8)
	bigMaxInt8   = big.NewInt(math.MaxInt8)
	bigMinInt16  = big.NewInt(math.MinInt16)
	bigMaxInt16  = big.NewInt(math.MaxInt16)
	bigMinInt32  = big.NewInt(math.MinInt32)
	bigMaxInt32  = big.NewInt(math.MaxInt32)
	bigMinInt64  = big.NewInt(math.MinInt64)
	bigMaxInt64  = big.NewInt(math.MaxInt64)
	bigMinInt    = big.NewInt(int64(math.MinInt))
	bigMaxInt    = big.NewInt(int64(math.MaxInt))
	bigMaxUint8  = big.NewInt(math.MaxUint8)
	bigMaxUint16 = big.NewInt(math.MaxUint16)
	bigMaxUint32 = big.NewInt(math.MaxUint32)
	bigMaxUint64 = new(big.Int).SetUint64(math.MaxUint64)
	bigMaxUint   = new(big.Int).SetUint64(uint64(math.MaxUint))
)

func intBounds(base PrimitiveType) (lo, hi *big.Int) {
	switch base {
	case Int8Type:
		return bigMinInt8, bigMaxInt8
	case Int16Type:
		return bigMinInt16, bigMaxInt16
	case Int32Type:
		return bigMinInt32, bigMaxInt32
	case Int64Type:
		return bigMinInt64, bigMaxInt64
	case IntType:
		return bigMinInt, bigMaxInt
	case Uint8Type:
		return bigZero, bigMaxUint8
	case Uint16Type:
		return bigZero, bigMaxUint16
	case Uint32Type:
		return bigZero, bigMaxUint32
	case Uint64Type:
		return bigZero, bigMaxUint64
	case UintType:
		return bigZero, bigMaxUint
	default:
		return nil, nil
	}
}

func intFits(v *big.Int, base PrimitiveType) bool {
	lo, hi := intBounds(base)
	if lo == nil {
		return false
	}
	return v.Cmp(lo) >= 0 && v.Cmp(hi) <= 0
}

// overflowMsg matches gc (go/types operandString + "%s overflows %s").
// A defined type is "of int8 type MyInt"; a basic type is "of type int8".
// The final type is always the underlying type. byte and rune are universe
// aliases of uint8 and int32 here, so those two print as uint8 and int32
// where gc prints byte and rune.
func overflowMsg(expr Expr, val *big.Int, typ Type) string {
	return fmt.Sprintf("%s (constant %s of %s) overflows %s",
		constExprText(expr), val.String(), constTypePhrase(typ), baseOf(typ).String())
}

// constTypePhrase is the "of ..." fragment inside a constant overflow.
func constTypePhrase(t Type) string {
	base := baseOf(t).String()
	if dt, ok := t.(*DeclaredType); ok {
		return base + " type " + string(dt.Name)
	}
	return "type " + base
}

// constTypeWord is the constant's own type. A declared type prints as its
// name, which is what gc prints for a type in the same package. byte and
// rune are not declared types here: the universe binds them to uint8 and
// int32, so those print as uint8 and int32.
func constTypeWord(t Type) string {
	if dt, ok := t.(*DeclaredType); ok {
		return string(dt.Name)
	}
	return baseOf(t).String()
}

// constExprText prints a constant expression the way gc spells it in an
// overflow diagnostic: names without value paths, conversions as calls,
// and parentheses where precedence would otherwise lie.
func constExprText(x Expr) string {
	return constExprTextPrec(x, 0)
}

func constExprTextPrec(x Expr, minPrec int) string {
	for {
		cx, ok := x.(*ConstExpr)
		if !ok || cx.Source == nil {
			break
		}
		x = cx.Source
	}
	if x == nil {
		return ""
	}
	switch e := x.(type) {
	case *NameExpr:
		return string(e.Name)
	case *BasicLitExpr:
		return e.Value
	case *constTypeExpr:
		if e.Type != nil {
			return constTypeWord(e.Type)
		}
		return constExprTextPrec(e.Source, minPrec)
	case *CallExpr:
		args := make([]string, len(e.Args))
		for i, a := range e.Args {
			args[i] = constExprTextPrec(a, 0)
		}
		return constExprTextPrec(e.Func, 0) + "(" + strings.Join(args, ", ") + ")"
	case *SelectorExpr:
		return constExprTextPrec(e.X, 0) + "." + string(e.Sel)
	case *UnaryExpr:
		inner := constExprTextPrec(e.X, unaryPrec)
		if _, ok := unwrapConstSrc(e.X).(*UnaryExpr); ok {
			inner = "(" + constExprTextPrec(e.X, 0) + ")"
		}
		s := e.Op.TokenString() + inner
		if unaryPrec < minPrec {
			return "(" + s + ")"
		}
		return s
	case *BinaryExpr:
		p := binPrec(e.Op)
		s := constExprTextPrec(e.Left, p) + " " + e.Op.TokenString() + " " + constExprTextPrec(e.Right, p+1)
		if p < minPrec {
			return "(" + s + ")"
		}
		return s
	default:
		return e.String()
	}
}

func unwrapConstSrc(x Expr) Expr {
	for {
		cx, ok := x.(*ConstExpr)
		if !ok || cx.Source == nil {
			return x
		}
		x = cx.Source
	}
}

// unaryPrec is tighter than every binary operator.
const unaryPrec = 6

func binPrec(op Word) int {
	switch op {
	case MUL, QUO, REM, SHL, SHR, BAND, BAND_NOT:
		return 5
	case ADD, SUB, BOR, XOR:
		return 4
	case EQL, NEQ, LSS, LEQ, GTR, GEQ:
		return 3
	case LAND:
		return 2
	case LOR:
		return 1
	default:
		return 0
	}
}
