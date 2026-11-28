package gnolang

import (
	"fmt"
	"go/token"
	"math"
	"math/big"
	"strings"

	"github.com/gnolang/gno/gnovm/pkg/gnolang/internal/goconst"
	"github.com/gnolang/gno/tm2/pkg/overflow"
	"github.com/gnolang/gno/tm2/pkg/store"
)

// Constant folding (Go spec, "Constant expressions") with go/constant,
// forked as internal/goconst so its representation choices do not follow
// the toolchain and so every operation reports its work to a Meter
// before running: exact on untyped constants, representable in the type
// for typed ones. Runtime arithmetic (op_binary.go, op_unary.go) is
// untouched. Gno's untyped kinds (bigint, rune, bigdec, string, bool) are
// carried beside the go/constant value, which has no rune kind.

// foldConst folds x, a binary, unary or shift operation on constant
// operands of primitive kinds, and reports false for any other shape so
// the caller evaluates it on a machine. alloc accounts the bytes of a
// string result; gm is charged once per call.
func foldConst(gm store.GasMeter, alloc *Allocator, x Expr) (TypedValue, bool) {
	return foldConstWith(func() goconst.Meter { return chargeConstFold(gm) }, alloc, x)
}

// foldConstWith is foldConst with the Meter made by meter once the
// operands are known to fold (tests pass one that records the work).
func foldConstWith(meter func() goconst.Meter, alloc *Allocator, x Expr) (TypedValue, bool) {
	switch x := x.(type) {
	case *BinaryExpr:
		l, lok := foldOperand(x.Left)
		r, rok := foldOperand(x.Right)
		if !lok || !rok {
			return TypedValue{}, false
		}
		return foldBinary(meter(), alloc, x, l, r), true
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
		return foldUnary(meter(), x, v), true
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

// chargeConstFold charges the flat cost of one fold, OpCPUSlopeConstFold,
// and returns the Meter that prices each unit of work the fold's
// operations report before running (goconst.Work): nil when gm is.
func chargeConstFold(gm store.GasMeter) goconst.Meter {
	chargeCPUGas(gm, OpCPUSlopeConstFold)
	return constMeterOf(gm)
}

// constMeterOf is the Meter charging gm for the work of a constant
// operation or conversion: nil when gm is.
func constMeterOf(gm store.GasMeter) goconst.Meter {
	if gm == nil {
		return nil
	}
	return constMeter{gm}
}

// constMeter prices the work goconst reports. Word-sized and 512-bit
// float operations are fixed cost; untyped integers are charged per
// kilobit as the machine did (quadratic for * / %, a shift by its count
// and width); rationals linearly and quadratically in their digits, as
// big.Rat cross-multiplies and normalizes by a gcd; a string comparison
// per byte (a concatenation is paid by its allocation).
type constMeter struct{ gm store.GasMeter }

func (m constMeter) Charge(w goconst.Work) {
	var gas int64
	switch w.Rep {
	case goconst.RepInt:
		lb := int64(w.Bits)
		switch {
		case w.Conv != goconst.ConvNone:
			gas = lb * OpCPUSlopeBigIntAdd / 1024
		case w.Unary && w.Op == token.XOR:
			gas = lb * OpCPUSlopeBigIntUxor / 1024
		case w.Unary:
			gas = lb * OpCPUSlopeBigIntUneg / 1024
		case w.Op == token.SHL:
			gas = int64(min(w.Shift, maxConstShift))*OpCPUSlopeBigIntShl/1024 + lb*OpCPUSlopeBigIntShl/1024
		case w.Op == token.SHR:
			gas = lb * OpCPUSlopeBigIntShr / 1024
		default:
			if slope, quadratic := bigIntQuadSlope(w.Op); quadratic {
				gas = overflow.Mulp(overflow.Mulp(lb/32, lb/32), slope) / 32
			} else {
				gas = lb * bigIntSlope(w.Op) / 1024
			}
		}
	case goconst.RepRat:
		d := int64(w.Bits) / 3 // decimal digits, as ratDigits
		q := overflow.Mulp(d/10, d/10)
		switch {
		case w.Conv != goconst.ConvNone, isComparison(toWord(w.Op)): // a cross-multiplication or a division, no gcd
			gas = OpCPUSlopeConstRatConv + d*OpCPUSlopeConstRatCmp/100 + overflow.Mulp(q, OpCPUSlopeConstRatCmpQ)/10
		case w.Unary: // -, +: a copy
			gas = OpCPUSlopeConstRatBase + d*OpCPUSlopeConstRatNeg/100
		default: // + - * /
			gas = OpCPUSlopeConstRatBase + d*OpCPUSlopeConstRat/100 + overflow.Mulp(q, OpCPUSlopeConstRatQ)/10
		}
	case goconst.RepFloat:
		gas = OpCPUSlopeConstFloat
		if w.Op == token.QUO {
			gas += OpCPUSlopeConstFloatQuo
		}
	case goconst.RepString:
		if isComparison(toWord(w.Op)) {
			gas = int64(w.Bits) * OpCPUSlopeBytesCmp
		}
	}
	chargeCPUGas(m.gm, gas)
}

func bigIntSlope(op token.Token) int64 {
	switch op {
	case token.ADD:
		return OpCPUSlopeBigIntAdd
	case token.SUB:
		return OpCPUSlopeBigIntSub
	case token.AND:
		return OpCPUSlopeBigIntBand
	case token.OR:
		return OpCPUSlopeBigIntBor
	case token.XOR:
		return OpCPUSlopeBigIntXor
	case token.AND_NOT:
		return OpCPUSlopeBigIntBandn
	case token.EQL, token.NEQ:
		return OpCPUSlopeBigIntEql
	}
	return OpCPUSlopeBigIntLss // <, <=, >, >=
}

// bigIntQuadSlope is the slope of an operation quadratic in its operands'
// width, and whether op is one.
func bigIntQuadSlope(op token.Token) (slope int64, quadratic bool) {
	switch op {
	case token.MUL:
		return OpCPUSlopeBigIntMulQ, true
	case token.QUO, token.QUO_ASSIGN:
		return OpCPUSlopeBigIntQuoQ, true
	case token.REM:
		return OpCPUSlopeBigIntRemQ, true
	}
	return 0, false
}

func foldBinary(m goconst.Meter, alloc *Allocator, x *BinaryExpr, l, r *TypedValue) TypedValue {
	lv, rv := constValueOf(m, l), constValueOf(m, r)
	switch x.Op {
	case EQL, NEQ, LSS, LEQ, GTR, GEQ:
		tv := *l // the machine folded in place; same bytes (see constTypedValue)
		tv.T = UntypedBoolType
		tv.V = nil
		tv.SetBool(goconst.Compare(m, lv, binaryToken(x.Op), rv))
		return tv
	case SHL, SHR:
		return foldShift(m, x, l, lv, rv)
	}
	tok := binaryToken(x.Op)
	if x.Op == QUO || x.Op == REM {
		if goconst.Sign(rv) == 0 {
			panic("invalid operation: division by zero")
		}
		if x.Op == QUO && isWhole(l.T) {
			tok = token.QUO_ASSIGN // integer division
		}
	}
	return constTypedValue(m, alloc, x, goconst.BinaryOp(m, lv, tok, rv), l)
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
func foldShift(m goconst.Meter, x *BinaryExpr, l *TypedValue, lv, rv goconst.Value) TypedValue {
	lv = goconst.ToInt(m, lv)
	if lv.Kind() != goconst.Int {
		panic(fmt.Sprintf("invalid operation: shifted operand %s (type %s) must be integer",
			constExprString(x.Left), l.T))
	}
	s, ok := goconst.Uint64Val(goconst.ToInt(m, rv))
	if !ok || s > maxConstShift {
		panic(fmt.Sprintf("invalid operation: invalid shift count %s", rv))
	}
	return constTypedValue(m, nil, x, goconst.Shift(m, lv, binaryToken(x.Op), uint(s)), l)
}

func foldUnary(m goconst.Meter, x *UnaryExpr, v *TypedValue) TypedValue {
	cv := constValueOf(m, v)
	var res goconst.Value
	switch x.Op {
	case ADD:
		res = cv
	case SUB:
		res = goconst.UnaryOp(m, token.SUB, cv, 0)
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
		res = goconst.UnaryOp(m, token.XOR, cv, prec)
	case NOT:
		res = goconst.UnaryOp(m, token.NOT, cv, 0)
	}
	return constTypedValue(m, nil, x, res, v)
}

// checkUntypedIntWidth rejects an untyped integer constant (bigint, or a
// rune past int32) wider than maxUntypedIntBits, with go/types' wording:
// "constant shift overflow", "constant overflow" for a literal.
func checkUntypedIntWidth(x Expr, tv *TypedValue) {
	// A BigintValue is an untyped bigint, or an untyped rune past int32.
	biv, ok := tv.V.(BigintValue)
	if !ok || biv.V.BitLen() <= maxUntypedIntBits {
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

// word2token is token2word (go2gno.go) inverted.
var word2token = func() map[Word]token.Token {
	m := make(map[Word]token.Token, len(token2word))
	for tok, w := range token2word {
		m[w] = tok
	}
	return m
}()

// binaryToken is the go/token operator of a binary operator Word.
func binaryToken(op Word) token.Token {
	if tok, ok := word2token[op]; ok && tok.IsOperator() {
		return tok
	}
	panic(fmt.Sprintf("not a constant operator: %s", op.TokenString()))
}

// constValueOf is the go/constant value of a constant of primitive kind.
// Typed floats are exact (a float32 is a float64 is a rational).
func constValueOf(m goconst.Meter, tv *TypedValue) goconst.Value {
	switch baseOf(tv.T).Kind() {
	case BoolKind:
		return goconst.MakeBool(tv.GetBool())
	case StringKind:
		return goconst.MakeString(tv.GetString())
	case IntKind:
		return goconst.MakeInt64(tv.GetInt())
	case Int8Kind:
		return goconst.MakeInt64(int64(tv.GetInt8()))
	case Int16Kind:
		return goconst.MakeInt64(int64(tv.GetInt16()))
	case Int32Kind:
		if bi, ok := tv.bigRune(); ok {
			return goconst.Make(m, bi)
		}
		return goconst.MakeInt64(int64(tv.GetInt32()))
	case Int64Kind:
		return goconst.MakeInt64(tv.GetInt64())
	case UintKind:
		return goconst.MakeUint64(tv.GetUint())
	case Uint8Kind:
		return goconst.MakeUint64(uint64(tv.GetUint8()))
	case Uint16Kind:
		return goconst.MakeUint64(uint64(tv.GetUint16()))
	case Uint32Kind:
		return goconst.MakeUint64(uint64(tv.GetUint32()))
	case Uint64Kind:
		return goconst.MakeUint64(tv.GetUint64())
	case Float32Kind:
		return goconst.MakeFloat64(float64(math.Float32frombits(tv.GetFloat32())))
	case Float64Kind:
		return goconst.MakeFloat64(math.Float64frombits(tv.GetFloat64()))
	case BigintKind:
		return goconst.Make(m, tv.GetBigInt())
	case BigdecKind:
		bd := tv.GetBigDec()
		switch {
		case bd.IsFloat():
			return goconst.Make(m, bd.F)
		case bd.V == nil:
			return goconst.Make(m, new(big.Rat))
		default:
			return goconst.Make(m, bd.V)
		}
	}
	panic(fmt.Sprintf("not a constant kind: %s", tv.T))
}

// constTypedValue is the Gno constant holding v, the folded result of x,
// of the type of the operand l. An untyped result is exact; a typed one
// must be representable in its type. The result starts as a copy of l and
// only the kind's bytes of N are written, as the machine's in-place fold
// did, so the persisted bytes of a constant are unchanged.
func constTypedValue(m goconst.Meter, alloc *Allocator, x Expr, v goconst.Value, l *TypedValue) TypedValue {
	if v.Kind() == goconst.Unknown {
		panic("constant result is not representable")
	}
	tv := *l
	t := l.T
	k := baseOf(t).Kind()
	switch k {
	case BoolKind:
		tv.SetBool(goconst.BoolVal(v))
	case StringKind:
		tv.V = alloc.NewString(goconst.StringVal(m, v))
	case BigintKind:
		tv.V = BigintValue{V: constBigInt(m, v)}
	case BigdecKind:
		tv.V = constBigdec(v)
	case Float32Kind:
		f, _ := goconst.Float32Val(m, v)
		if math.IsInf(float64(f), 0) {
			panicTypedConstOverflow(x, v, t)
		}
		tv.SetFloat32(math.Float32bits(posZero(f)))
	case Float64Kind:
		f, _ := goconst.Float64Val(m, v)
		if math.IsInf(f, 0) {
			panicTypedConstOverflow(x, v, t)
		}
		tv.SetFloat64(math.Float64bits(posZero(f)))
	default:
		bi := constBigInt(m, v)
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
func panicTypedConstOverflow(x Expr, v goconst.Value, t Type) {
	panic(fmt.Sprintf("%s (constant %s of type %s) overflows %s",
		constExprString(x), v, t, t))
}

// constBigInt is the integer value of v, which must be integral.
func constBigInt(m goconst.Meter, v goconst.Value) *big.Int {
	switch val := goconst.Val(goconst.ToInt(m, v)).(type) {
	case int64:
		return big.NewInt(val)
	case *big.Int:
		return new(big.Int).Set(val)
	}
	panic(fmt.Sprintf("constant %s is not an integer", v))
}

// constBigdec is v as an untyped bigdec: rational while both components
// fit ratOverflowBits, else the 512-bit float form, as go/goconst.
func constBigdec(v goconst.Value) BigdecValue {
	switch val := goconst.Val(v).(type) {
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
	if fitsInt32(bi) {
		tv.SetInt32(int32(bi.Int64()))
		return
	}
	tv.ClearNum()
	tv.V = BigintValue{V: bi}
}

func fitsInt32(bi *big.Int) bool {
	return bi.IsInt64() && bi.Int64() >= math.MinInt32 && bi.Int64() <= math.MaxInt32
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
	if r, ok := intKindBounds[k]; ok {
		return r[0], r[1]
	}
	panic(fmt.Sprintf("not an integer kind: %s", k))
}

var intKindBounds = map[Kind][2]*big.Int{
	Int8Kind:   {big.NewInt(math.MinInt8), big.NewInt(math.MaxInt8)},
	Int16Kind:  {big.NewInt(math.MinInt16), big.NewInt(math.MaxInt16)},
	Int32Kind:  {big.NewInt(math.MinInt32), big.NewInt(math.MaxInt32)},
	IntKind:    {big.NewInt(math.MinInt64), big.NewInt(math.MaxInt64)},
	Int64Kind:  {big.NewInt(math.MinInt64), big.NewInt(math.MaxInt64)},
	Uint8Kind:  {big.NewInt(0), big.NewInt(math.MaxUint8)},
	Uint16Kind: {big.NewInt(0), big.NewInt(math.MaxUint16)},
	Uint32Kind: {big.NewInt(0), big.NewInt(math.MaxUint32)},
	UintKind:   {big.NewInt(0), new(big.Int).SetUint64(math.MaxUint64)},
	Uint64Kind: {big.NewInt(0), new(big.Int).SetUint64(math.MaxUint64)},
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
