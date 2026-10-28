package gnolang

import (
	"fmt"
	"go/constant"
	"go/token"
	"math"
	"math/big"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/overflow"
	"github.com/gnolang/gno/tm2/pkg/store"
)

// Constant folding (Go spec, "Constant expressions") with go/constant:
// exact on untyped constants, representable in the type for typed ones.
// Runtime arithmetic (op_binary.go, op_unary.go) is untouched. Gno's
// untyped kinds (bigint, rune, bigdec, string, bool) are carried beside
// the go/constant value, which has no rune kind.

// foldConst folds x, a binary, unary or shift operation on constant
// operands of primitive kinds, and reports false for any other shape so
// the caller evaluates it on a machine. alloc accounts the bytes of a
// string result; gm is charged once per call.
func foldConst(gm store.GasMeter, alloc *Allocator, x Expr) (TypedValue, bool) {
	switch x := x.(type) {
	case *BinaryExpr:
		l, lok := foldOperand(x.Left)
		r, rok := foldOperand(x.Right)
		if !lok || !rok {
			return TypedValue{}, false
		}
		chargeConstFold(gm, x.Op, l, r)
		return foldBinary(alloc, x, l, r), true
	case *UnaryExpr:
		switch x.Op {
		case ADD, SUB, XOR, NOT:
		default:
			return TypedValue{}, false
		}
		v, ok := foldOperand(x.X)
		if !ok {
			return TypedValue{}, false
		}
		chargeConstFold(gm, x.Op, v, nil)
		return foldUnary(x, v), true
	}
	return TypedValue{}, false
}

// foldOperand returns the value of x when it is a constant of a primitive
// kind: bool, string, integer, float, bigint or bigdec.
func foldOperand(x Expr) (*TypedValue, bool) {
	cx, ok := x.(*ConstExpr)
	if !ok || cx.T == nil || isDataByte(cx.T) {
		return nil, false
	}
	if _, ok := baseOf(cx.T).(PrimitiveType); !ok {
		return nil, false
	}
	return &cx.TypedValue, true
}

// chargeConstFold charges one fold: OpCPUSlopeConstFold, plus size terms
// for the operands whose cost is not fixed: untyped integers per kilobit
// as the machine charged (quadratic for * / %), rationals by their digits
// (OpCPUSlopeConstRat*), and a string comparison per byte (a concatenation
// is paid by its allocation).
func chargeConstFold(gm store.GasMeter, op Word, l, r *TypedValue) {
	if gm == nil {
		return
	}
	gas := int64(OpCPUSlopeConstFold)
	switch baseOf(l.T).Kind() {
	case StringKind:
		if op != ADD {
			gas += int64(len(l.GetString())) * OpCPUSlopeBytesCmp
			gas += int64(len(r.GetString())) * OpCPUSlopeBytesCmp
		}
	case BigintKind, Int32Kind:
		lb := constIntBits(l)
		if r == nil {
			gas += lb * bigIntUnarySlope(op) / 1024
			break
		}
		switch op {
		case MUL, QUO, REM:
			rb := constIntBits(r)
			gas += overflow.Mulp(overflow.Mulp(lb/32, rb/32), bigIntQuadSlope(op)) / 32
		case SHL:
			s, _ := constant.Uint64Val(constant.ToInt(constValueOf(r)))
			gas += int64(min(s, maxConstShift)) * OpCPUSlopeBigIntShl / 1024
			gas += lb * OpCPUSlopeBigIntShl / 1024
		case SHR:
			gas += lb * OpCPUSlopeBigIntShr / 1024
		default:
			gas += max(lb, constIntBits(r)) * bigIntSlope(op) / 1024
		}
	case Float32Kind, Float64Kind:
		gas += OpCPUSlopeConstFloat
		if op == QUO {
			gas += OpCPUSlopeConstFloatQuo
		}
	case BigdecKind:
		// Only the rational form has size-dependent cost; the 512-bit float
		// form is a fixed-width operation.
		lb, rb := l.GetBigDec(), BigdecValue{}
		if r != nil {
			rb = r.GetBigDec()
		}
		if lb.IsFloat() || rb.IsFloat() {
			gas += OpCPUSlopeConstFloat
			if op == QUO {
				gas += OpCPUSlopeConstFloatQuo
			}
			break
		}
		d := bigdecDigits(lb)
		if r == nil {
			gas += d * OpCPUSlopeConstRatNeg / 100
			break
		}
		d = max(d, bigdecDigits(rb))
		switch op {
		case EQL, NEQ, LSS, LEQ, GTR, GEQ:
			gas += d * OpCPUSlopeConstRatCmp / 100
			gas += overflow.Mulp(overflow.Mulp(d/10, d/10), OpCPUSlopeConstRatCmpQ) / 10
		default: // + - * /: cross-multiplication and a gcd
			gas += d * OpCPUSlopeConstRat / 100
			gas += overflow.Mulp(overflow.Mulp(d/10, d/10), OpCPUSlopeConstRatQ) / 10
		}
	}
	chargeCPUGas(gm, gas)
}

// constIntBits is the width of an untyped integer constant held as a
// BigintValue (a bigint, or a rune past int32); a typed or in-range
// operand folds in a word, paid by the flat charge.
func constIntBits(tv *TypedValue) int64 {
	if biv, ok := tv.V.(BigintValue); ok {
		return int64(biv.V.BitLen())
	}
	return 0
}

func bigIntSlope(op Word) int64 {
	switch op {
	case ADD:
		return OpCPUSlopeBigIntAdd
	case SUB:
		return OpCPUSlopeBigIntSub
	case BAND:
		return OpCPUSlopeBigIntBand
	case BOR:
		return OpCPUSlopeBigIntBor
	case XOR:
		return OpCPUSlopeBigIntXor
	case BAND_NOT:
		return OpCPUSlopeBigIntBandn
	case EQL, NEQ:
		return OpCPUSlopeBigIntEql
	}
	return OpCPUSlopeBigIntLss // <, <=, >, >=
}

func bigIntQuadSlope(op Word) int64 {
	switch op {
	case MUL:
		return OpCPUSlopeBigIntMulQ
	case QUO:
		return OpCPUSlopeBigIntQuoQ
	}
	return OpCPUSlopeBigIntRemQ
}

func bigIntUnarySlope(op Word) int64 {
	switch op {
	case SUB:
		return OpCPUSlopeBigIntUneg
	case XOR:
		return OpCPUSlopeBigIntUxor
	}
	return 0 // unary +
}

func foldBinary(alloc *Allocator, x *BinaryExpr, l, r *TypedValue) TypedValue {
	lv, rv := constValueOf(l), constValueOf(r)
	switch x.Op {
	case EQL, NEQ, LSS, LEQ, GTR, GEQ:
		tv := *l // the machine folded in place; same bytes (see constTypedValue)
		tv.T = UntypedBoolType
		tv.V = nil
		tv.SetBool(constant.Compare(lv, binaryToken(x.Op), rv))
		return tv
	case SHL, SHR:
		return foldShift(x, l, lv, rv)
	}
	tok := binaryToken(x.Op)
	switch x.Op {
	case QUO, REM:
		if constant.Sign(rv) == 0 {
			panic("invalid operation: division by zero")
		}
		if x.Op == QUO && isIntegerConst(l.T) {
			tok = token.QUO_ASSIGN // integer division
		}
	}
	return constTypedValue(alloc, x, constant.BinaryOp(lv, tok, rv), l)
}

// maxConstShift is the largest constant shift count, as go/types
// (shiftBound, 1023-1+52 so smallestFloat64 is expressible);
// maxUntypedIntBits the width of an untyped integer constant, as go/types
// (prec). Both bound every fold.
const (
	maxConstShift     = 1074
	maxUntypedIntBits = 512
)

// foldShift folds a constant shift: the count is a non-negative typed uint
// (assertShiftExprCompatible1), bounded by maxConstShift; an untyped
// float operand must be integral.
func foldShift(x *BinaryExpr, l *TypedValue, lv, rv constant.Value) TypedValue {
	lv = constant.ToInt(lv)
	if lv.Kind() != constant.Int {
		panic(fmt.Sprintf("invalid operation: shifted operand %s (type %s) must be integer",
			constExprString(x.Left), l.T))
	}
	s, ok := constant.Uint64Val(constant.ToInt(rv))
	if !ok || s > maxConstShift {
		panic(fmt.Sprintf("invalid operation: invalid shift count %s", rv))
	}
	return constTypedValue(nil, x, constant.Shift(lv, binaryToken(x.Op), uint(s)), l)
}

func foldUnary(x *UnaryExpr, v *TypedValue) TypedValue {
	cv := constValueOf(v)
	var res constant.Value
	switch x.Op {
	case ADD:
		res = cv
	case SUB:
		res = constant.UnaryOp(token.SUB, cv, 0)
	case XOR:
		// An unsigned complement is taken at the type's width.
		var prec uint
		switch baseOf(v.T).Kind() {
		case Uint8Kind:
			prec = 8
		case Uint16Kind:
			prec = 16
		case Uint32Kind:
			prec = 32
		case UintKind, Uint64Kind:
			prec = 64
		}
		res = constant.UnaryOp(token.XOR, cv, prec)
	case NOT:
		res = constant.UnaryOp(token.NOT, cv, 0)
	}
	return constTypedValue(nil, x, res, v)
}

// checkUntypedIntWidth rejects an untyped integer constant (bigint, or a
// rune past int32) wider than maxUntypedIntBits, with go/types' wording:
// "constant shift overflow", "constant overflow" for a literal.
func checkUntypedIntWidth(x Expr, tv *TypedValue) {
	var bi *big.Int
	switch tv.T {
	case UntypedBigintType:
		bi = tv.GetBigInt()
	case UntypedRuneType:
		biv, ok := tv.V.(BigintValue)
		if !ok {
			return
		}
		bi = biv.V
	default:
		return
	}
	if bi.BitLen() <= maxUntypedIntBits {
		return
	}
	op := ""
	switch x := x.(type) {
	case *BinaryExpr:
		switch x.Op {
		case ADD:
			op = "addition "
		case SUB:
			op = "subtraction "
		case MUL:
			op = "multiplication "
		case XOR:
			op = "bitwise XOR "
		case SHL:
			op = "shift "
		}
	case *UnaryExpr:
		if x.Op == XOR {
			op = "bitwise complement "
		}
	}
	panic(fmt.Sprintf("constant %soverflow", op))
}

// isIntegerConst reports whether constants of type t are integers:
// typed integer kinds, untyped bigint and untyped rune.
func isIntegerConst(t Type) bool {
	switch baseOf(t).Kind() {
	case BigintKind:
		return true
	default:
		return isIntegerKind(baseOf(t).Kind())
	}
}

func binaryToken(op Word) token.Token {
	switch op {
	case ADD:
		return token.ADD
	case SUB:
		return token.SUB
	case MUL:
		return token.MUL
	case QUO:
		return token.QUO
	case REM:
		return token.REM
	case BAND:
		return token.AND
	case BOR:
		return token.OR
	case XOR:
		return token.XOR
	case BAND_NOT:
		return token.AND_NOT
	case LAND:
		return token.LAND
	case LOR:
		return token.LOR
	case SHL:
		return token.SHL
	case SHR:
		return token.SHR
	case EQL:
		return token.EQL
	case NEQ:
		return token.NEQ
	case LSS:
		return token.LSS
	case LEQ:
		return token.LEQ
	case GTR:
		return token.GTR
	case GEQ:
		return token.GEQ
	}
	panic(fmt.Sprintf("not a constant operator: %s", op.TokenString()))
}

// constValueOf is the go/constant value of a constant of primitive kind.
// Typed floats are exact (a float32 is a float64 is a rational).
func constValueOf(tv *TypedValue) constant.Value {
	switch baseOf(tv.T).Kind() {
	case BoolKind:
		return constant.MakeBool(tv.GetBool())
	case StringKind:
		return constant.MakeString(tv.GetString())
	case IntKind:
		return constant.MakeInt64(tv.GetInt())
	case Int8Kind:
		return constant.MakeInt64(int64(tv.GetInt8()))
	case Int16Kind:
		return constant.MakeInt64(int64(tv.GetInt16()))
	case Int32Kind:
		if tv.T == UntypedRuneType {
			return constant.Make(runeConstInt(tv))
		}
		return constant.MakeInt64(int64(tv.GetInt32()))
	case Int64Kind:
		return constant.MakeInt64(tv.GetInt64())
	case UintKind:
		return constant.MakeUint64(tv.GetUint())
	case Uint8Kind:
		return constant.MakeUint64(uint64(tv.GetUint8()))
	case Uint16Kind:
		return constant.MakeUint64(uint64(tv.GetUint16()))
	case Uint32Kind:
		return constant.MakeUint64(uint64(tv.GetUint32()))
	case Uint64Kind:
		return constant.MakeUint64(tv.GetUint64())
	case Float32Kind:
		return constant.MakeFloat64(float64(math.Float32frombits(tv.GetFloat32())))
	case Float64Kind:
		return constant.MakeFloat64(math.Float64frombits(tv.GetFloat64()))
	case BigintKind:
		return constant.Make(tv.GetBigInt())
	case BigdecKind:
		bd := tv.GetBigDec()
		switch {
		case bd.IsFloat():
			return constant.Make(bd.F)
		case bd.V == nil:
			return constant.Make(new(big.Rat))
		default:
			return constant.Make(bd.V)
		}
	}
	panic(fmt.Sprintf("not a constant kind: %s", tv.T))
}

// constTypedValue is the Gno constant holding v, the folded result of x,
// of the type of the operand l. An untyped result is exact; a typed one
// must be representable in its type. The result starts as a copy of l and
// only the kind's bytes of N are written, as the machine's in-place fold
// did, so the persisted bytes of a constant are unchanged.
func constTypedValue(alloc *Allocator, x Expr, v constant.Value, l *TypedValue) TypedValue {
	if v.Kind() == constant.Unknown {
		panic("constant result is not representable")
	}
	tv := *l
	t := l.T
	k := baseOf(t).Kind()
	switch k {
	case BoolKind:
		tv.SetBool(constant.BoolVal(v))
	case StringKind:
		tv.V = alloc.NewString(constant.StringVal(v))
	case BigintKind:
		tv.V = BigintValue{V: constBigInt(v)}
	case BigdecKind:
		tv.V = constBigdec(v)
	case Float32Kind:
		f, _ := constant.Float32Val(v)
		if math.IsInf(float64(f), 0) {
			panicTypedConstOverflow(x, v, t)
		}
		tv.SetFloat32(math.Float32bits(posZero(f)))
	case Float64Kind:
		f, _ := constant.Float64Val(v)
		if math.IsInf(f, 0) {
			panicTypedConstOverflow(x, v, t)
		}
		tv.SetFloat64(math.Float64bits(posZero(f)))
	default:
		bi := constBigInt(v)
		if t == UntypedRuneType {
			setRuneConst(&tv, bi)
			return tv
		}
		lo, hi := intKindRange(k)
		if bi.Cmp(lo) < 0 || bi.Cmp(hi) > 0 {
			panicTypedConstOverflow(x, v, t)
		}
		setIntKind(&tv, k, bi)
	}
	return tv
}

// panicTypedConstOverflow is gc's rejection of a typed constant operation
// whose result is not representable in its type.
func panicTypedConstOverflow(x Expr, v constant.Value, t Type) {
	panic(fmt.Sprintf("%s (constant %s of type %s) overflows %s",
		constExprString(x), v, t, t))
}

// constBigInt is the integer value of v, which must be integral.
func constBigInt(v constant.Value) *big.Int {
	switch val := constant.Val(constant.ToInt(v)).(type) {
	case int64:
		return big.NewInt(val)
	case *big.Int:
		return new(big.Int).Set(val)
	}
	panic(fmt.Sprintf("constant %s is not an integer", v))
}

// constBigdec is v as an untyped bigdec: rational while both components
// fit ratOverflowBits, else the 512-bit float form, as go/constant.
func constBigdec(v constant.Value) BigdecValue {
	switch val := constant.Val(v).(type) {
	case int64:
		return BigdecValue{V: new(big.Rat).SetInt64(val)}
	case *big.Int:
		return BigdecValue{V: new(big.Rat).SetInt(val)}
	case *big.Rat:
		return wrapRatOrPromote(new(big.Rat).Set(val))
	case *big.Float:
		return BigdecValue{F: new(big.Float).SetPrec(BigdecFloatPrec).Set(val)}
	}
	panic(fmt.Sprintf("constant %s is not a number", v))
}

// setRuneConst stores an untyped rune constant: int32 in N when it fits,
// the exact BigintValue otherwise.
func setRuneConst(tv *TypedValue, bi *big.Int) {
	tv.V = nil
	if bi.IsInt64() && bi.Int64() >= math.MinInt32 && bi.Int64() <= math.MaxInt32 {
		tv.SetInt32(int32(bi.Int64()))
		return
	}
	tv.ClearNum()
	tv.V = BigintValue{V: bi}
}

// setIntKind stores bi, known to be in range, in a typed integer of kind k.
func setIntKind(tv *TypedValue, k Kind, bi *big.Int) {
	switch k {
	case IntKind:
		tv.SetInt(bi.Int64())
	case Int8Kind:
		tv.SetInt8(int8(bi.Int64()))
	case Int16Kind:
		tv.SetInt16(int16(bi.Int64()))
	case Int32Kind:
		tv.SetInt32(int32(bi.Int64()))
	case Int64Kind:
		tv.SetInt64(bi.Int64())
	case UintKind:
		tv.SetUint(bi.Uint64())
	case Uint8Kind:
		tv.SetUint8(uint8(bi.Uint64()))
	case Uint16Kind:
		tv.SetUint16(uint16(bi.Uint64()))
	case Uint32Kind:
		tv.SetUint32(uint32(bi.Uint64()))
	case Uint64Kind:
		tv.SetUint64(bi.Uint64())
	default:
		panic(fmt.Sprintf("not an integer kind: %s", k))
	}
}

// intKindRange is the range of values an integer kind represents.
func intKindRange(k Kind) (lo, hi *big.Int) {
	switch k {
	case Int8Kind:
		return big.NewInt(math.MinInt8), big.NewInt(math.MaxInt8)
	case Int16Kind:
		return big.NewInt(math.MinInt16), big.NewInt(math.MaxInt16)
	case Int32Kind:
		return big.NewInt(math.MinInt32), big.NewInt(math.MaxInt32)
	case IntKind, Int64Kind:
		return big.NewInt(math.MinInt64), big.NewInt(math.MaxInt64)
	case Uint8Kind:
		return big.NewInt(0), big.NewInt(math.MaxUint8)
	case Uint16Kind:
		return big.NewInt(0), big.NewInt(math.MaxUint16)
	case Uint32Kind:
		return big.NewInt(0), big.NewInt(math.MaxUint32)
	case UintKind, Uint64Kind:
		return big.NewInt(0), new(big.Int).SetUint64(math.MaxUint64)
	}
	panic(fmt.Sprintf("not an integer kind: %s", k))
}

// constExprString prints a constant operation as written (as gc does),
// through the folded constants preprocessing put in place of its
// subexpressions.
func constExprString(x Expr) string {
	switch x := x.(type) {
	case *ConstExpr:
		switch src := x.Source.(type) {
		case *BinaryExpr, *UnaryExpr, *NameExpr, *BasicLitExpr, *SelectorExpr, *CallExpr, *ConstExpr:
			return constExprString(src)
		}
		return x.String()
	case *BinaryExpr:
		return constOperandString(x.Left, x.Op, false) + " " +
			x.Op.TokenString() + " " + constOperandString(x.Right, x.Op, true)
	case *UnaryExpr:
		return x.Op.TokenString() + constOperandString(x.X, x.Op, false)
	case *NameExpr:
		return string(x.Name)
	case *BasicLitExpr:
		return x.Value
	case *SelectorExpr:
		return constExprString(x.X) + "." + string(x.Sel)
	case *constTypeExpr:
		return x.Type.String()
	case *CallExpr:
		args := make([]string, len(x.Args))
		for i, a := range x.Args {
			args[i] = constExprString(a)
		}
		return constExprString(x.Func) + "(" + strings.Join(args, ", ") + ")"
	}
	return x.String()
}

// constOperandString is constExprString, parenthesized when x binds
// looser than the operation it is an operand of (as go/printer does).
func constOperandString(x Expr, op Word, right bool) string {
	if cx, ok := x.(*ConstExpr); ok {
		x = cx.Source
		// A shift count was wrapped in uint() by the preprocessor.
		if call, ok := x.(*CallExpr); ok && call.GetAttribute(ATTR_SHIFT_RHS) != nil {
			x = call.Args[0]
		}
	}
	if bx, ok := x.(*BinaryExpr); ok {
		if p, q := binaryPrecedence(bx.Op), binaryPrecedence(op); p < q || (p == q && right) {
			return "(" + constExprString(bx) + ")"
		}
	}
	return constExprString(x)
}

func binaryPrecedence(op Word) int {
	switch op {
	case LOR:
		return 1
	case LAND:
		return 2
	case EQL, NEQ, LSS, LEQ, GTR, GEQ:
		return 3
	case ADD, SUB, BOR, XOR:
		return 4
	case MUL, QUO, REM, SHL, SHR, BAND, BAND_NOT:
		return 5
	}
	return 6
}
