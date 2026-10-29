package gnolang

import (
	"fmt"
	"math"
	"math/big"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/gnolang/gno/tm2/pkg/store"
)

// Rules for import paths, constants and branch statements (Go spec:
// "Import declarations", "Constant expressions", "Break statements",
// "Continue statements").

// assertValidImportPath checks the spelling of an import path: not
// empty, valid UTF-8, graphic characters outside Go's excluded set, and
// not absolute.
func assertValidImportPath(path string) {
	if path == "" {
		fail(CodeImportPath, "invalid import path (empty string)")
	}
	if !utf8.ValidString(path) {
		failf(CodeImportPath, nil, "import path contains invalid UTF-8 character: %q", path)
	}
	for _, r := range path {
		if !unicode.IsGraphic(r) || unicode.IsSpace(r) ||
			strings.ContainsRune("!\"#$%&'()*,:;<=>?[\\]^`{|}", r) {
			failf(CodeImportPath, nil, "invalid import path (invalid character %#U)", r)
		}
	}
	if strings.HasPrefix(path, "/") {
		failf(CodeImportPath, nil, "import path cannot be absolute path: %q", path)
	}
}

// constWidth is the width go/types gives untyped integer constants.
const constWidth = 512

// maxConstShift is the largest count go/types accepts in a constant
// shift of a constant: enough to express the smallest float64.
const maxConstShift = 1023 - 1 + 52

// assertConstShiftCount rejects a constant shift of a constant by more
// than maxConstShift, before the fold, as go/types does. A negative
// count has its own error.
func assertConstShiftCount(rcx *ConstExpr) {
	if rcx.T == nil {
		return
	}
	over := false
	switch rcx.T.Kind() {
	case BigintKind:
		b := rcx.GetBigInt()
		over = b.Sign() >= 0 && (b.BitLen() > 63 || b.Int64() > maxConstShift)
	case Uint64Kind:
		over = rcx.GetUint64() > maxConstShift
	case IntKind, Int8Kind, Int16Kind, Int32Kind, Int64Kind, UintKind, Uint8Kind, Uint16Kind, Uint32Kind:
		tv := rcx.TypedValue // ConvertGetInt converts in place
		over = tv.ConvertGetInt() > maxConstShift
	}
	if !over {
		return
	}
	what := "untyped int constant"
	if !isUntyped(rcx.T) {
		what = fmt.Sprintf("constant %s of type %s", rcx.TypedValue.Sprint(nil), rcx.T)
	}
	expr := exprString(rcx)
	if nx, ok := rcx.Source.(*NameExpr); ok {
		expr = string(nx.Name)
	}
	failf(CodeShiftCount, nil, "invalid operation: invalid shift count %s (%s)", expr, what)
}

// assertConstWidth rejects a folded untyped integer constant wider than
// constWidth bits, naming the operation that produced it.
func assertConstWidth(x Expr, cv *TypedValue) {
	if cv.T != UntypedBigintType || cv.GetBigInt().BitLen() <= constWidth {
		return
	}
	op := ""
	switch x := x.(type) {
	case *BinaryExpr:
		op = constOpNames[x.Op]
		// A shifted float literal is converted first; go/types names no
		// operation.
		if lcx, ok := x.Left.(*ConstExpr); ok && (x.Op == SHL || x.Op == SHR) {
			if bl, ok := lcx.Source.(*BasicLitExpr); ok && bl.Kind == FLOAT {
				op = ""
			}
		}
	case *UnaryExpr:
		switch x.Op {
		case XOR:
			op = "bitwise complement"
		case SUB:
			op = "negation"
		}
	}
	if op == "" {
		fail(CodeConstOverflow, "constant overflow")
	}
	failf(CodeConstOverflow, nil, "constant %s overflow", op)
}

var constOpNames = map[Word]string{
	ADD: "addition", SUB: "subtraction", MUL: "multiplication", QUO: "division",
	SHL: "shift", SHR: "shift", XOR: "bitwise XOR", BAND: "bitwise AND",
	BOR: "bitwise OR", BAND_NOT: "bitwise AND NOT",
}

// assertBranchTarget checks that the statement a labeled break or
// continue names can be its target: a loop, or for break also a switch
// or select.
func assertBranchTarget(n *BranchStmt, bn BlockNode) {
	switch bn.(type) {
	case *ForStmt, *RangeStmt:
		return
	case *SwitchClauseStmt, *SelectCaseStmt:
		if n.Op == BREAK {
			return
		}
	}
	failBranchLabel(n)
}

// failBranchLabel rejects the label of a break or continue whose target
// cannot be branched to.
func failBranchLabel(n *BranchStmt) {
	if n.Op == BREAK {
		failf(CodeBranchLabel, nil, "invalid break label %s", n.Label)
	}
	failf(CodeBranchLabel, nil, "invalid continue label %s", n.Label)
}

func assertValidConstExpr(store Store, last BlockNode, n *ValueDecl, expr Expr) {
	if n.Type != nil {
		nt := evalStaticType(store, last, n.Type)
		if _, ok := baseOf(nt).(PrimitiveType); !ok {
			panic(fmt.Sprintf("invalid constant type %s", nt.String()))
		}
	}

	nt := evalStaticTypeOf(store, last, expr)

	if nt == nil {
		panic(fmt.Sprintf("%s (variable of type nil) is not constant", expr))
	}

	if _, ok := baseOf(nt).(PrimitiveType); !ok {
		panic(fmt.Sprintf("%s (variable of type %s) is not constant", expr, nt))
	}

	assertValidConstValue(store, last, expr)
}

func assertValidConstValue(store Store, last BlockNode, currExpr Expr) {
Main:
	switch currExpr := currExpr.(type) {
	case *ConstExpr:
	case *UnaryExpr:
		// *, & is filter out previously since they are not primitive
		assertValidConstValue(store, last, currExpr.X)
	case *TypeAssertExpr:
		panic(fmt.Sprintf("%s (comma, ok expression of type %s) is not constant", currExpr.String(), currExpr.Type))
	case *CallExpr:
		ift := evalStaticTypeOf(store, last, currExpr.Func)
		switch baseOf(ift).(type) {
		case *FuncType:
			tup := evalStaticTypeOfRaw(store, last, currExpr).(*tupleType)

			// check for built-in functions
			if cx, ok := currExpr.Func.(*ConstExpr); ok {
				if fv, ok := cx.V.(*FuncValue); ok {
					if fv.PkgPath == uversePkgPath {
						// TODO: should support min, max, real, imag
						switch fv.Name {
						case "len", "cap":
							at := evalStaticTypeOf(store, last, currExpr.Args[0])
							if _, ok := unwrapPointerType(baseOf(at)).(*ArrayType); ok {
								if hasCall(preprocessGasMeterOf(store), store, last, currExpr.Args[0]) {
									panic(fmt.Sprintf("%s (value of type int) is not constant", exprString(currExpr)))
								}
								break Main
							}
							assertValidConstValue(store, last, currExpr.Args[0])
							break Main
						}
					}
				}
			}

			switch {
			case len(tup.Elts) == 0:
				panic(fmt.Sprintf("%s (no value) used as value", currExpr.String()))
			case len(tup.Elts) == 1:
				panic(fmt.Sprintf("%s (value of type %s) is not constant", currExpr.String(), tup.Elts[0]))
			default:
				panic(fmt.Sprintf("multiple-value %s (value of type %s) in single-value context", currExpr.String(), tup.Elts))
			}
		case *TypeType:
			// string([]byte(nil)): the target must be a constant type.
			if ct := evalStaticType(store, last, currExpr.Func); !isConstantType(ct) {
				panic(fmt.Sprintf("%s (value of type %s) is not constant", exprString(currExpr), ct))
			}
			for _, arg := range currExpr.Args {
				assertValidConstValue(store, last, arg)
			}
		default:
			panic(fmt.Sprintf(
				"unexpected func type %v (%v)",
				ift, reflect.TypeOf(ift)))
		}
	case *BinaryExpr:
		assertValidConstValue(store, last, currExpr.Left)
		assertValidConstValue(store, last, currExpr.Right)
	case *SelectorExpr:
		xt := evalStaticTypeOf(store, last, currExpr.X)
		switch xt := xt.(type) {
		case *PackageType:
			var pv *PackageValue
			if cx, ok := currExpr.X.(*ConstExpr); ok {
				// NOTE: *Machine.TestMemPackage() needs this
				// to pass in an imported package as *ConstEzpr.
				pv = cx.V.(*PackageValue)
			} else {
				// otherwise, packages can only be referred to by
				// *NameExprs, and cannot be copied.
				pvc := evalConst(store, last, currExpr.X)
				pv_, ok := pvc.V.(*PackageValue)
				if !ok {
					panic(fmt.Sprintf(
						"missing package in selector expr %s",
						currExpr.String()))
				}
				pv = pv_
			}
			if pv.GetBlock(store).Source.GetIsConst(store, currExpr.Sel) {
				break Main
			}

			tt := pv.GetBlock(store).Source.GetStaticTypeOf(store, currExpr.Sel)
			panic(fmt.Sprintf("%s (variable of type %s) is not constant", currExpr.String(), tt))
		case *PointerType, *DeclaredType, *StructType, *InterfaceType, *TypeType:
			ty := evalStaticTypeOf(store, last, currExpr)
			if _, ok := ty.(*TypeType); ok {
				ty = evalStaticType(store, last, currExpr)
			}
			panic(fmt.Sprintf("%s (variable of type %s) is not constant", currExpr.String(), ty))
		default:
			fail(CodeMissingField, fmt.Sprintf(
				"%s.%s undefined (type %s has no field or method %s)",
				exprString(currExpr.X), currExpr.Sel, xt.String(), currExpr.Sel), currExpr)
		}
	default:
		ift := evalStaticTypeOf(store, last, currExpr)
		if _, ok := ift.(*TypeType); ok {
			ift = evalStaticType(store, last, currExpr)
		}
		panic(fmt.Sprintf("%s (variable of type %s) is not constant", currExpr.String(), ift))
	}
}

// assertConstFinite rejects a folded untyped float constant beyond
// big.Float's exponent range (an infinity, go/constant's unknown value):
// a literal is malformed, an operation's result is not representable.
func assertConstFinite(x Expr, cv *TypedValue) {
	if cv.T != UntypedBigdecType {
		return
	}
	bd, ok := cv.V.(BigdecValue)
	if !ok || !bd.IsFloat() || !bd.F.IsInf() {
		return
	}
	if bl, ok := x.(*BasicLitExpr); ok {
		failf(CodeConstOverflow, nil, "malformed constant: %s", bl.Value)
	}
	fail(CodeConstOverflow, "constant result is not representable")
}

// assertConstRepresentable rejects the constant operation x, about to be
// folded, when its operands are typed integer constants and the exact
// result is not representable in their type (Go spec, Constant
// expressions). The machine folds with runtime arithmetic, which wraps;
// a constant must not. Shifts are checked by the fold, the other
// operations cannot leave the type, and a zero divisor keeps its error.
func assertConstRepresentable(gm store.GasMeter, x Expr) {
	var t Type
	var exact *big.Int
	switch x := x.(type) {
	case *BinaryExpr:
		l, lok := typedIntConst(x.Left)
		r, rok := typedIntConst(x.Right)
		if !lok || !rok {
			return
		}
		t = x.Left.(*ConstExpr).T
		switch x.Op {
		case ADD:
			exact = new(big.Int).Add(l, r)
		case SUB:
			exact = new(big.Int).Sub(l, r)
		case MUL:
			exact = new(big.Int).Mul(l, r)
		case QUO:
			if r.Sign() == 0 {
				return
			}
			exact = new(big.Int).Quo(l, r)
		default:
			return
		}
	case *UnaryExpr:
		v, ok := typedIntConst(x.X)
		if !ok || x.Op != SUB {
			return
		}
		t = x.X.(*ConstExpr).T
		exact = new(big.Int).Neg(v)
	default:
		return
	}
	// One big operation on operands of at most 64 bits: bounded cost.
	chargeCPUGas(gm, OpCPUSlopeCheckNode)
	lo, hi := intKindRange(baseOf(t).Kind())
	if exact.Cmp(lo) >= 0 && exact.Cmp(hi) <= 0 {
		return
	}
	failf(CodeConstOverflow, []Node{x}, "%s (constant %s of type %s) overflows %s",
		constExprString(x), exact, t, t)
}

// typedIntConst returns the value of x when it is a typed constant of an
// integer kind.
func typedIntConst(x Expr) (*big.Int, bool) {
	cx, ok := x.(*ConstExpr)
	if !ok || cx.T == nil || isUntyped(cx.T) {
		return nil, false
	}
	tv := &cx.TypedValue
	switch baseOf(cx.T).Kind() {
	case IntKind:
		return big.NewInt(tv.GetInt()), true
	case Int8Kind:
		return big.NewInt(int64(tv.GetInt8())), true
	case Int16Kind:
		return big.NewInt(int64(tv.GetInt16())), true
	case Int32Kind:
		return big.NewInt(int64(tv.GetInt32())), true
	case Int64Kind:
		return big.NewInt(tv.GetInt64()), true
	case UintKind:
		return new(big.Int).SetUint64(tv.GetUint()), true
	case Uint8Kind:
		return new(big.Int).SetUint64(uint64(tv.GetUint8())), true
	case Uint16Kind:
		return new(big.Int).SetUint64(uint64(tv.GetUint16())), true
	case Uint32Kind:
		return new(big.Int).SetUint64(uint64(tv.GetUint32())), true
	case Uint64Kind:
		return new(big.Int).SetUint64(tv.GetUint64()), true
	}
	return nil, false
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

// constExprString prints a constant operation as written, through the
// folded constants preprocessing put in place of its subexpressions.
func constExprString(x Expr) string {
	switch x := x.(type) {
	case *ConstExpr:
		switch x.Source.(type) {
		case *BinaryExpr, *UnaryExpr:
			return constExprString(x.Source)
		}
	case *BinaryExpr:
		return typedConstOperandString(x.Left, x.Op, false) + " " +
			x.Op.TokenString() + " " + typedConstOperandString(x.Right, x.Op, true)
	case *UnaryExpr:
		return x.Op.TokenString() + typedConstOperandString(x.X, x.Op, false)
	}
	return exprString(x)
}

// typedConstOperandString is constExprString, parenthesized when x binds
// looser than the operation it is an operand of (as go/printer does).
func typedConstOperandString(x Expr, op Word, right bool) string {
	if cx, ok := x.(*ConstExpr); ok {
		x = cx.Source
	}
	if bx, ok := x.(*BinaryExpr); ok {
		if p, q := binaryPrecedence(bx.Op), binaryPrecedence(op); p < q || (p == q && right) {
			return "(" + constExprString(bx) + ")"
		}
	}
	return constExprString(x)
}

// binaryPrecedence is the Go precedence of a binary operator; a unary
// operator binds tightest.
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

// isConstantType reports whether t can be the type of a constant: a
// boolean, numeric or string type (Go spec, "Constants").
func isConstantType(t Type) bool {
	_, ok := baseOf(t).(PrimitiveType)
	return ok
}
