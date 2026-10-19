package gnolang

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/store"
)

// Expression rules the preprocessor enforces (Go spec: "Comparison
// operators", "Index expressions", "Slice expressions", "Composite
// literals", "Length and capacity").

// isNilLiteral reports whether x is the nil constant as written, also
// after its conversion to the context type; a typed nil such as
// (func())(nil) is not the literal.
func isNilLiteral(x Expr) bool {
	cx, ok := x.(*ConstExpr)
	if !ok || cx.V != nil {
		return false
	}
	if cx.T == nil {
		return true
	}
	nx, ok := cx.Source.(*NameExpr)
	return ok && nx.Name == "nil"
}

// assertNotType rejects a type where a value is expected, named as
// written when it is a name, else by the type it denotes.
func assertNotType(o operand) {
	if !o.isType() {
		return
	}
	name := typeString(o.typ)
	switch o.expr.(type) {
	case *NameExpr, *SelectorExpr, *ConstExpr, *constTypeExpr:
		name = exprString(o.expr)
	}
	failf(CodeNotExpression, nil, "%s (type) is not an expression", name)
}

// assertComparableTo checks a == b. A slice, map or func compares only
// to the nil literal, also when the other side is an interface.
func assertComparableTo(a, b operand) {
	assertNotType(a)
	assertNotType(b)
	check := func(t Type, selfNil, otherNil bool) {
		if t == nil || selfNil || otherNil {
			return
		}
		switch baseOf(t).(type) {
		case *SliceType, *FuncType, *MapType:
			failf(CodeNilOnly, nil, "%v can only be compared to nil", t)
		}
		if !isComparable(t) {
			failf(CodeNotComparable, nil, "%v is not comparable", t)
		}
	}
	check(a.typ, a.mode == opNil, b.mode == opNil)
	check(b.typ, b.mode == opNil, a.mode == opNil)
}

// lenCapArgError checks the operand of len or cap: a string (len only),
// array, pointer to array, slice, map (len only) or channel. It returns
// the message rather than raising it: the call's type is int whatever the
// operand, so accumulate mode reports and goes on, as gc does.
func lenCapArgError(name Name, arg operand) string {
	assertNotType(arg)
	t := arg.typ
	if t == nil {
		return ""
	}
	switch bt := baseOf(t).(type) {
	case *ArrayType, *SliceType, *ChanType:
		return ""
	case *PointerType:
		if _, ok := baseOf(bt.Elt).(*ArrayType); ok {
			return ""
		}
	case *MapType:
		if name == "len" {
			return ""
		}
	}
	if t.Kind() == StringKind && name == "len" {
		return ""
	}
	return fmt.Sprintf("invalid argument: %v for built-in %s", t, name)
}

// assertIndexable checks the operand of an index expression: a string,
// array, pointer to array, slice or map. A pointer to anything else is not
// dereferenced (Go spec: "Index expressions").
func assertIndexable(x Expr, t Type) {
	if t != nil {
		switch bt := baseOf(t).(type) {
		case *ArrayType, *SliceType, *MapType:
			return
		case *PointerType:
			if _, ok := baseOf(bt.Elt).(*ArrayType); ok {
				return
			}
		default:
			if t.Kind() == StringKind {
				return
			}
		}
	}
	failf(CodeIndexOperand, nil, "invalid operation: cannot index %s (%s of type %v)", operandString(x), operandWord(x), t)
}

// operandWord is gc's noun for an operand in a message: a variable, else
// a value.
func operandWord(x Expr) string {
	if _, ok := x.(*NameExpr); ok {
		return "variable"
	}
	return "value"
}

// staticLengthOf returns the length a constant index can be checked
// against: an array's (through one pointer), a constant string's, else -1.
func staticLengthOf(t Type, x Expr) int64 {
	if t == nil {
		return -1
	}
	switch bt := baseOf(t).(type) {
	case *ArrayType:
		return int64(bt.Len)
	case *PointerType:
		if at, ok := baseOf(bt.Elt).(*ArrayType); ok {
			return int64(at.Len)
		}
	}
	if t.Kind() == StringKind {
		// Read the length off the value; GetString would allocate.
		if cx, ok := x.(*ConstExpr); ok {
			if sv, ok := cx.V.(StringValue); ok {
				return int64(len(sv))
			}
		}
	}
	return -1
}

func constIndex(x Expr) (int64, bool) {
	cx, ok := x.(*ConstExpr)
	if !ok || cx.T == nil || cx.IsUndefined() {
		return 0, false
	}
	switch cx.T.Kind() {
	case IntKind, Int8Kind, Int16Kind, Int32Kind, Int64Kind, UintKind, Uint8Kind, Uint16Kind, Uint32Kind, Uint64Kind, BigintKind:
		return cx.ConvertGetInt(), true
	}
	return 0, false
}

// assertConstIndex checks a constant index against a known length. A
// negative constant was already refused by checkOrConvertIndexKind.
func assertConstIndex(index Expr, length int64) {
	if i, ok := constIndex(index); ok && length >= 0 && i >= length {
		failf(CodeIndexBounds, nil, "invalid argument: index %d out of bounds [0:%d]", i, length)
	}
}

// assertConstSliceBounds checks constant slice indices: ordered, and
// within the length when it is known (an index may equal the length).
// Negative constants were already refused by checkOrConvertIndexKind.
func assertConstSliceBounds(n *SliceExpr, length int64) {
	prev := int64(-1)
	for _, ix := range []Expr{n.Low, n.High, n.Max} {
		if ix == nil {
			continue
		}
		i, ok := constIndex(ix)
		if !ok {
			continue
		}
		if length >= 0 && i > length {
			failf(CodeIndexBounds, nil, "invalid argument: index %d out of bounds [0:%d]", i, length+1)
		}
		if prev >= 0 && i < prev {
			failf(CodeSliceIndices, nil, "invalid slice indices: %d < %d", i, prev)
		}
		prev = i
	}
}

// assertLiteralIndices rejects a repeated constant index in an array or
// slice literal, and one at or past length when it is known (-1 for a
// slice); elements without a key take the next index. Charged per
// element; the error is positioned at the element.
func assertLiteralIndices(gm store.GasMeter, n *CompositeLitExpr, length int64) {
	chargeCPUGas(gm, OpCPUSlopeCheckNode*int64(len(n.Elts)))
	seen := map[int64]bool{}
	next := int64(0)
	for i := range n.Elts {
		elt := &n.Elts[i]
		idx := next
		if elt.Key != nil {
			i, ok := constIndex(elt.Key)
			if !ok {
				continue
			}
			idx = i
			if length >= 0 && idx >= length {
				failf(CodeIndexBounds, []Node{elt.Key, elt.Value}, "invalid argument: index %d out of bounds [0:%d]", idx, length)
			}
		} else if length >= 0 && idx >= length {
			failf(CodeIndexBounds, []Node{elt.Value}, "index %d is out of bounds (>= %d)", idx, length)
		}
		if seen[idx] {
			failf(CodeDuplicateIndex, []Node{elt.Key, elt.Value}, "duplicate index %d in array or slice literal", idx)
		}
		seen[idx] = true
		next = idx + 1
	}
}

// valueBuiltins are the built-in functions that may not be called in
// statement context (Go spec: "Expression statements").
var valueBuiltins = map[Name]bool{
	"append": true, "cap": true, "complex": true, "imag": true,
	"len": true, "make": true, "new": true, "real": true,
}

// assertExprStmtUsed rejects an expression statement that is not a call,
// a conversion, or a call to a value builtin. A type assertion statement
// is left alone: Gno runs it for its panic.
func assertExprStmtUsed(store Store, last BlockNode, n *ExprStmt) {
	call, ok := n.X.(*CallExpr)
	if !ok {
		if _, ok := n.X.(*TypeAssertExpr); ok {
			return
		}
		notUsed(operandOf(store, last, n.X))
	}
	fn := operandOf(store, last, call.Func)
	switch fn.mode {
	case opType:
		notUsed(operandOf(store, last, n.X))
	case opBuiltin:
		if valueBuiltins[fn.val.GetFunc().Name] {
			notUsed(operandOf(store, last, n.X))
		}
	}
}

func notUsed(o operand) {
	assertNotType(o)
	xs := exprString(o.expr)
	if _, ok := o.expr.(*FuncLitExpr); ok {
		xs = fmt.Sprintf("(%s literal)", typeString(o.typ)) // as go/types
	}
	failf(CodeNotUsed, nil, "%s (value of type %s) is not used", xs, typeString(o.typ))
}

// constStringLen folds len of a constant string into an int constant.
func constStringLen(name Name, arg Expr) *ConstExpr {
	cx, ok := arg.(*ConstExpr)
	if name != "len" || !ok || cx.T == nil || cx.T.Kind() != StringKind {
		return nil
	}
	tv := TypedValue{T: IntType}
	tv.SetInt(int64(len(cx.GetString())))
	return toConstExpr(cx, tv)
}

// assertConstLiteralKey: an array or slice literal index is a constant.
func assertConstLiteralKey(store Store, last BlockNode, key Expr) {
	if key == nil {
		return
	}
	o := operandOf(store, last, key)
	assertNotType(o)
	if o.mode != opConstant {
		failf(CodeLiteralKeyConst, []Node{key}, "index %s must be integer constant", exprString(key))
	}
}

// isFuncDeclName reports whether nx names a declared function. It walks
// the path as GetBlockNodeForPath does, stopping on a parent the walk
// has not linked yet instead of panicking.
func isFuncDeclName(last BlockNode, nx *NameExpr) bool {
	if nx.Path.Type != VPBlock || nx.Path.Depth == 0 {
		return false
	}
	bn := last
	for i := 1; i < int(nx.Path.Depth); i++ {
		if fauxChildBlockNode(bn) {
			bn = bn.GetParentNode(nil)
		}
		if bn == nil {
			return false
		}
		bn = bn.GetParentNode(nil)
		if bn == nil {
			return false
		}
	}
	switch fb := bn.(type) {
	case *IfCaseStmt, *SwitchClauseStmt:
		if pn := fb.GetParentNode(nil); pn != nil && nx.Path.Index < pn.GetNumNames() {
			bn = pn
		}
	}
	idx, ok := bn.GetLocalIndex(nx.Name)
	sources := bn.GetStaticBlock().NameSources
	return ok && int(idx) < len(sources) && sources[idx].Type == NSFuncDecl
}

// ===========================================================
// assertShiftExprCompatible1 check happens while trans_leave binary expression.
// it checks both lhs and rhs types of shift expression.
func (x *BinaryExpr) assertShiftExprCompatible1(store Store, last BlockNode, lt, rt Type) {
	if rt == nil {
		panic(fmt.Sprintf("cannot convert %v to type uint", x.Right))
	}

	lcx, lic := x.Left.(*ConstExpr)
	_, ric := x.Right.(*ConstExpr)
	// Step1, check RHS type.
	// Must be numeric.
	if !isNumeric(rt) {
		panic(fmt.Sprintf("cannot convert %v to type uint", x.Right))
	}
	// If not const, must be IntNum.
	if !isWhole(rt) && !ric {
		panic(fmt.Sprintf("invalid operation: invalid shift count: %v", x.Right))
	}

	if ric {
		rv := evalConst(store, last, x.Right)
		if rv.Sign() < 0 {
			panic(fmt.Sprintf("invalid operation: negative shift count: %v", x.Right))
		}

		if isWhole(rt) {
			// Good.
		} else if !IsExactBigDec(rv.V) { // e.g. 1.0 .
			panic(fmt.Sprintf("invalid operation: invalid shift count: %v", x.Right))
		}
	}

	// Step2, check lhs type.
	if checker, ok := binaryChecker[x.Op]; ok {
		if checker(lt) { // check pass
			return
		}

		// If lhs not IntNum, it must be const.
		if !lic {
			if isUntyped(lt) {
				lt = defaultTypeOf(lt)
			}
			panic(fmt.Sprintf("operator %s not defined on: %v", x.Op.TokenString(), kindString(lt)))
		}

		// LHS is const.
		// Special case for untyped const lhs.
		lv := evalConst(store, last, lcx)
		if !IsExactBigDec(lv.V) {
			panic(fmt.Sprintf("invalid operation: shifted operand %v (%v) must be integer", lv, lt))
		}

		// Both const. left is untypedBigDec & exact integer.
		if ric {
			// Representable as an integer. e.g. 1.0 << 1.
			// convert lhs to untypedBigint so it can be evaluated as const later.
			convertConst(store, last, x, lcx, UntypedBigintType)
		}
		return
	}
	panic(fmt.Sprintf("checker for %s does not exist", x.Op))
}

// assertShiftExprCompatible2 checks if untyped (non-const)
// shift expr is compatible with t finally, which is type info
// from context. e.g. var y int = 1.0 << x.
func (x *BinaryExpr) assertShiftExprCompatible2(t Type) {
	// check lhs type
	if checker, ok := binaryChecker[x.Op]; ok {
		if !checker(t) {
			panic(fmt.Sprintf("operator %s not defined on: %v", x.Op.TokenString(), kindString(t)))
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
// store supplies the preprocess gas meter for the EQL/NEQ operand check
// (`S{} == i` runs the same interface-satisfaction walk as an assignment).
func (x *BinaryExpr) AssertCompatible(store Store, lt, rt Type) {
	xt, dt, swapped := lt, rt, false
	if shouldSwapOnSpecificity(lt, rt) {
		xt, dt, swapped = dt, xt, true
	}

	if isComparison(x.Op) {
		switch x.Op {
		case EQL, NEQ:
			// Slice, func, and map can only be compared to nil.
			switch baseOf(dt).(type) {
			case *SliceType, *FuncType, *MapType:
				if xt != nil {
					panic(fmt.Sprintf("%v can only be compared to nil", dt))
				}
			default:
				if !isComparable(dt) {
					panic(fmt.Sprintf("%v is not comparable", dt))
				}
			}
			// An interface operand does not make a slice, map or func
			// on the other side comparable.
			if xt != nil {
				switch baseOf(xt).(type) {
				case *SliceType, *FuncType, *MapType:
					panic(fmt.Sprintf("%v can only be compared to nil", xt))
				}
			}
			err := checkAssignableTo(store, x, xt, dt)
			if err != nil {
				if debug {
					debug.Printf("checkAssignableTo fail: %v\n", err)
				}
				panic(fmt.Sprintf("invalid operation: (mismatched types %v and %v)", xt, dt))
			}
		case LSS, LEQ, GTR, GEQ:
			if checker, ok := binaryChecker[x.Op]; ok {
				x.checkCompatibility(store, x, xt, dt, checker, x.Op.TokenString(), swapped)
			} else {
				panic(fmt.Sprintf("checker for %s does not exist", x.Op))
			}
		default:
			panic("invalid comparison operator")
		}
	} else {
		if checker, ok := binaryChecker[x.Op]; ok {
			x.checkCompatibility(store, x, xt, dt, checker, x.Op.TokenString(), swapped)
		} else {
			panic(fmt.Sprintf("checker for %s does not exist", x.Op))
		}

		switch x.Op {
		case QUO, REM:
			// special case of zero divisor
			if isQuoOrRem(x.Op) {
				if rcx, ok := x.Right.(*ConstExpr); ok {
					if rcx.TypedValue.Sign() == 0 {
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
func (x *BinaryExpr) checkCompatibility(store Store, n Node, xt, dt Type, checker func(t Type) bool, OpStr string, swapped bool) {
	if !checker(dt) {
		panic(fmt.Sprintf("operator %s not defined on: %v", OpStr, kindString(dt)))
	}

	// display xt as "untyped nil" if nil as Go does.
	untypedNil := func(t Type) string {
		if t == nil {
			return "untyped nil"
		} else {
			return t.String()
		}
	}

	// if both typed
	if !isUntyped(xt) && !isUntyped(dt) {
		err := checkAssignableTo(store, n, xt, dt)
		if err != nil {
			if debug {
				debug.Printf("checkAssignableTo fail: %v\n", err)
			}
			if swapped {
				panic(fmt.Sprintf("invalid operation: %v (mismatched types %v and %v)", n, dt, untypedNil(xt)))
			} else {
				panic(fmt.Sprintf("invalid operation: %v (mismatched types %v and %v)", n, untypedNil(xt), dt))
			}
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

// shouldSwapOnSpecificity determines the potential direction for
// checkOrConvertType. it checks whether a swap is needed between two types
// based on their specificity. If t2 has a lower specificity than t1, it returns
// false, indicating no swap is needed. If t1 has a lower specificity than t2,
// it returns true, indicating a swap is needed.
func shouldSwapOnSpecificity(t1, t2 Type) bool {
	// check nil
	if t1 == nil {
		return false // also with both nil
	} else if t2 == nil {
		return true
	}

	// check interface
	if it1, ok := baseOf(t1).(*InterfaceType); ok {
		if it1.IsEmptyInterface() {
			return true // left empty interface
		} else {
			if it2, ok := baseOf(t2).(*InterfaceType); ok {
				if it2.IsEmptyInterface() {
					return false
				}
				// The more methods, the more specific.
				return len(it1.Methods) < len(it2.Methods)
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

// hasCall reports whether x holds a function call other than a conversion
// or a constant len/cap of an array: len and cap of an array are constant
// only then (Go spec: "Length and capacity"; Gno has no channel receive).
// Charged per node entered.
func hasCall(gm store.GasMeter, store Store, last BlockNode, x Expr) bool {
	found := false
	Transcribe(x, func(ns []Node, ftype TransField, index int, n Node, stage TransStage) (Node, TransCtrl) {
		if stage != TRANS_ENTER {
			return n, TRANS_CONTINUE
		}
		chargeCPUGas(gm, OpCPUSlopeCheckNode)
		switch n := n.(type) {
		case *CallExpr:
			if isConstType(n.Func) {
				return n, TRANS_CONTINUE // a conversion
			}
			if cx, ok := n.Func.(*ConstExpr); ok && len(n.Args) == 1 {
				if fv, ok := cx.V.(*FuncValue); ok && fv.PkgPath == uversePkgPath && (fv.Name == "len" || fv.Name == "cap") {
					at := evalStaticTypeOf(store, last, n.Args[0])
					if _, ok := unwrapPointerType(baseOf(at)).(*ArrayType); ok {
						return n, TRANS_CONTINUE // constant unless its operand is not
					}
				}
			}
			found = true
			return n, TRANS_EXIT
		}
		return n, TRANS_CONTINUE
	})
	return found
}

// eltFailure is a composite literal element that failed a check: the
// message and the element to report it at.
type eltFailure struct {
	code ErrCode
	msg  string
	at   Node
}

// checkElt runs check on one element of a composite literal and records
// a checking error it raises in fails, to be reported at the element; a
// VM panic or an error that is positioned already passes through.
func checkElt(fails *[]eltFailure, at Node, check func()) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		if isRuntimePanic(r) {
			panic(r)
		}
		switch e := r.(type) {
		case *PreprocessError:
			panic(r)
		case *spanError:
			*fails = append(*fails, eltFailure{e.code, e.msg, at})
		case error:
			*fails = append(*fails, eltFailure{CodeNone, e.Error(), at})
		default:
			*fails = append(*fails, eltFailure{CodeNone, fmt.Sprint(r), at})
		}
	}()
	check()
}

// assertAppendArgs checks append's arity: a typed slice, then values, with
// ... only on a single second argument (Go spec: "Appending to and
// copying slices"). Wording and positions follow go/types.
func assertAppendArgs(store Store, last BlockNode, n *CallExpr) {
	if len(n.Args) == 0 {
		fail(CodeAppendArgs, "invalid operation: not enough arguments for append() (expected 1, found 0)")
	}
	st := evalStaticTypeOf(store, last, n.Args[0])
	if st == nil {
		fail(CodeAppendArgs, "first argument to append must be a typed slice; have untyped nil")
	}
	bst, isSlice := baseOf(st).(*SliceType)
	if !n.Varg {
		if isSlice {
			assertAppendValues(store, last, n, bst.Elt)
		}
		return
	}
	if len(n.Args) == 2 {
		return
	}
	want := "([]T, ...T)"
	if isSlice {
		want = fmt.Sprintf("(%s, ...%s)", typeString(st), typeString(bst.Elt))
	}
	have := "("
	for i, arg := range n.Args {
		if i > 0 {
			have += ", "
		}
		have += argTypeSummary(evalStaticTypeOf(store, last, arg))
	}
	have += "...)"
	if len(n.Args) > 2 {
		failf(CodeAppendArgs, []Node{n.Args[2]}, "too many arguments in call to append\n\thave %s\n\twant %s", have, want)
	}
	failf(CodeAppendArgs, nil, "not enough arguments in call to append\n\thave %s\n\twant %s", have, want)
}

// argTypeSummary names an argument's type as go/types does in have/want
// lists: an untyped constant by its kind, a numeric one as "number".
func argTypeSummary(t Type) string {
	switch {
	case t == nil:
		return "nil"
	case isUntyped(t):
		if isNumeric(t) {
			return "number"
		}
		return untypedKindName(t)
	}
	return typeString(t)
}

// assertAppendValues checks each typed value argument against the slice's
// element type with gc's wording; untyped constants are converted by the
// caller and nil is left to the general call check.
func assertAppendValues(store Store, last BlockNode, n *CallExpr, elt Type) {
	for _, arg := range n.Args[1:] {
		at := evalStaticTypeOf(store, last, arg)
		if at == nil || isUntyped(at) {
			continue
		}
		if checkAssignableTo(store, n, at, elt) != nil {
			failf(CodeAppendArgs, []Node{arg}, "cannot use %s (%s of type %s) as %s value in argument to append", exprString(arg), operandKind(arg), typeString(at), typeString(elt))
		}
	}
}

// operandKind is go/types' word for an operand: a named variable or a value.
func operandKind(x Expr) string {
	if _, ok := x.(*NameExpr); ok {
		return "variable"
	}
	return "value"
}

// operandTypeString prints a type as go/types names an operand's: an
// untyped constant's by its kind.
func operandTypeString(t Type) string {
	if t == nil {
		return "nil"
	}
	if isUntyped(t) {
		return "untyped " + untypedKindName(t)
	}
	return typeString(t)
}

// constOperandString prints a constant as go/types does: "-1 (constant of
// type int)", or "i (constant -1 of type int)" when the value is not the
// spelling. An untyped constant is printed after the conversion gc
// applies to it; a non-integral one keeps its untyped form.
func constOperandString(cx *ConstExpr, t Type) string {
	expr := constString(cx)
	tn := operandTypeString(cx.T)
	var val string
	switch {
	case cx.T == UntypedBigdecType && !cx.GetBigDec().IsInt():
		return fmt.Sprintf("%s (untyped float constant)", expr)
	case cx.T == UntypedBigdecType:
		val, tn = bigdecToBigInt(cx.GetBigDec()).String(), typeString(t)
	case isUntyped(cx.T):
		val, tn = cx.TypedValue.Sprint(nil), typeString(t)
	default:
		val = cx.TypedValue.Sprint(nil)
	}
	if expr == val {
		return fmt.Sprintf("%s (constant of type %s)", expr, tn)
	}
	return fmt.Sprintf("%s (constant %s of type %s)", expr, val, tn)
}

func (x *UnaryExpr) AssertCompatible(t Type) {
	// check compatible
	if checker, ok := unaryChecker[x.Op]; ok {
		if !checker(t) {
			panic(fmt.Sprintf("operator %s not defined on: %v", x.Op.TokenString(), kindString(t)))
		}
	} else {
		panic(fmt.Sprintf("checker for %s does not exist", x.Op))
	}
}

// AssertCompatible checks that x++/x-- is well-formed: the operand's type must
// support the operator, and the operand must be a valid (addressable, settable)
// assignment target. The operator/type error is reported first, matching
// go/types.
func (x *IncDecStmt) AssertCompatible(store Store, last BlockNode) {
	t := evalStaticTypeOf(store, last, x.X)
	// check compatible
	if checker, ok := IncDecStmtChecker[x.Op]; ok {
		if !checker(t) {
			panic(fmt.Sprintf("operator %s not defined on: %v", x.Op.TokenString(), kindString(t)))
		}
	} else {
		panic(fmt.Sprintf("checker for %s does not exist", x.Op))
	}
	// like go/types, operator/type errors take precedence over
	// target assignability.
	assertValidAssignLhs(store, last, x.X)
}

// AssertCompatible checks that the range key and value targets are valid
// assignment targets and that their types are assignable from the range
// expression's key/element types. Each operand is handled independently
// (like AssignStmt.AssertCompatible): kt/vt are nil iff that operand is
// blank, and the type checks are skipped for a nil operand.
func (x *RangeStmt) AssertCompatible(store Store, last BlockNode) {
	if x.Op != ASSIGN {
		return
	}
	kt := evalAssignLhsType(store, last, x.Key)
	var vt Type
	if x.Value != nil {
		vt = evalAssignLhsType(store, last, x.Value)
	}
	if kt == nil && vt == nil {
		return
	}

	xt := baseOf(evalStaticTypeOf(store, last, x.X))
	// Ranging over a pointer to array iterates the array itself (int keys,
	// array-element values), so unwrap it to the array type. Mirrors the
	// PointerKind handling in preprocess.
	if pt, ok := xt.(*PointerType); ok {
		xt = baseOf(pt.Elem())
	}
	if kt != nil {
		switch cxt := xt.(type) {
		case *MapType:
			mustAssignableTo(store, x, cxt.Key, kt)
		case *SliceType, *ArrayType:
			mustAssignableTo(store, x, IntType, kt)
		case PrimitiveType:
			if cxt.Kind() == StringKind {
				mustAssignableTo(store, x, IntType, kt)
			}
		}
	}
	if vt != nil {
		switch cxt := xt.(type) {
		case *MapType:
			mustAssignableTo(store, x, cxt.Value, vt)
		case *SliceType, *ArrayType:
			mustAssignableTo(store, x, cxt.Elem(), vt)
		case PrimitiveType:
			if cxt.Kind() == StringKind {
				mustAssignableTo(store, x, Int32Type, vt) // rune
			}
		}
	}
}
