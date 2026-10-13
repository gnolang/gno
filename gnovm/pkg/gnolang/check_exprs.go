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

// assertComparableTo checks a == b. A slice, map or func compares only
// to the nil literal, also when the other side is an interface.
func assertComparableTo(a, b operand) {
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

// assertLenCapArg checks the operand of len or cap: a string (len only),
// array, pointer to array, slice, map (len only) or channel.
func assertLenCapArg(name Name, t Type) {
	if t == nil {
		return
	}
	switch bt := baseOf(t).(type) {
	case *ArrayType, *SliceType, *ChanType:
		return
	case *PointerType:
		if _, ok := baseOf(bt.Elt).(*ArrayType); ok {
			return
		}
	case *MapType:
		if name == "len" {
			return
		}
	}
	if t.Kind() == StringKind && name == "len" {
		return
	}
	failf(CodeLenCapArg, nil, "invalid argument: %v for built-in %s", t, name)
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

// assertConstIndex checks a constant index against a known length.
func assertConstIndex(index Expr, length int64) {
	if i, ok := constIndex(index); ok {
		if i < 0 {
			failf(CodeIndexBounds, nil, "invalid argument: index %d (constant of type int) must not be negative", i)
		}
		if length >= 0 && i >= length {
			failf(CodeIndexBounds, nil, "invalid argument: index %d out of bounds [0:%d]", i, length)
		}
	}
}

// assertConstSliceBounds checks constant slice indices: ordered, and
// within the length when it is known (an index may equal the length).
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
		if i < 0 {
			failf(CodeIndexBounds, nil, "invalid argument: index %d (constant of type int) must not be negative", i)
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

// assertUniqueLiteralIndices rejects a repeated constant index in an array
// or slice literal; elements without a key take the next index.
func assertUniqueLiteralIndices(gm store.GasMeter, n *CompositeLitExpr) {
	chargeCPUGas(gm, OpCPUSlopeCheckNode*int64(len(n.Elts)))
	seen := map[int64]bool{}
	next := int64(0)
	for _, elt := range n.Elts {
		idx := next
		if elt.Key != nil {
			i, ok := constIndex(elt.Key)
			if !ok {
				continue
			}
			idx = i
		}
		if seen[idx] {
			failf(CodeDuplicateIndex, nil, "duplicate index %d in array or slice literal", idx)
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
	if o.mode == opType {
		failf(CodeNotExpression, nil, "%s (type) is not an expression", exprString(o.expr))
	}
	failf(CodeNotUsed, nil, "%s (value of type %s) is not used", exprString(o.expr), typeString(o.typ))
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
func assertConstLiteralKey(key Expr) {
	if key == nil {
		return
	}
	if _, ok := key.(*ConstExpr); !ok {
		failf(CodeLiteralKeyConst, []Node{key}, "index %s must be integer constant", exprString(key))
	}
}

// isFuncDeclName reports whether nx names a declared function.
func isFuncDeclName(last BlockNode, nx *NameExpr) (is bool) {
	defer func() {
		if recover() != nil {
			is = false
		}
	}()
	if nx.Path.Type != VPBlock {
		return false
	}
	bn := last.GetBlockNodeForPath(nil, nx.Path)
	if bn == nil {
		return false
	}
	idx, ok := bn.GetLocalIndex(nx.Name)
	sources := bn.GetStaticBlock().NameSources
	return ok && int(idx) < len(sources) && sources[idx].Type == NSFuncDecl
}

// inFunction reports whether bn is inside a function declaration or
// literal.
func inFunction(bn BlockNode) bool {
	for ; bn != nil; bn = bn.GetParentNode(nil) {
		switch bn.(type) {
		case *FuncDecl, *FuncLitExpr:
			return true
		}
	}
	return false
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
