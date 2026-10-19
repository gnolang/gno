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

// assertComparableTo checks a == b between the types at and bt, where
// aNil / bNil say whether the operand is the untyped nil literal. A slice,
// map or func compares only to nil, also when the other side is an
// interface.
func assertComparableTo(at, bt Type, aNil, bNil bool) {
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
	check(at, aNil, bNil)
	check(bt, bNil, aNil)
}

// lenCapArgError checks the operand of len or cap: a string (len only),
// array, pointer to array, slice, map (len only) or channel. It returns
// the message rather than raising it: the call's type is int whatever the
// operand, so accumulate mode reports and goes on, as gc does.
func lenCapArgError(name Name, t Type) string {
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
		notUsed(store, last, n.X)
	}
	switch fx := call.Func.(type) {
	case *constTypeExpr:
		notUsed(store, last, n.X)
	case *ConstExpr:
		if fv := fx.GetFunc(); fv != nil && fv.PkgPath == uversePkgPath && valueBuiltins[fv.Name] {
			notUsed(store, last, n.X)
		}
	}
}

func notUsed(store Store, last BlockNode, x Expr) {
	t := evalStaticTypeOf(store, last, x)
	if _, ok := t.(*TypeType); ok {
		failf(CodeNotExpression, nil, "%s (type) is not an expression", exprString(x))
	}
	xs := exprString(x)
	if _, ok := x.(*FuncLitExpr); ok {
		xs = fmt.Sprintf("(%s literal)", typeString(t)) // as go/types
	}
	failf(CodeNotUsed, nil, "%s (value of type %s) is not used", xs, typeString(t))
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
