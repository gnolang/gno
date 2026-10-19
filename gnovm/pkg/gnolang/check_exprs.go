package gnolang

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/store"
)

// Expression rules the preprocessor enforces (Go spec: "Comparison
// operators", "Index expressions", "Slice expressions", "Composite
// literals", "Length and capacity").

// isNilLiteral reports whether x is the untyped nil constant.
func isNilLiteral(x Expr) bool {
	cx, ok := x.(*ConstExpr)
	return ok && cx.T == nil && cx.V == nil
}

// assertComparableTo checks a == b between the types at and bt, where
// aNil / bNil say whether the operand is the untyped nil literal. A slice,
// map or func compares only to nil, also when the other side is an
// interface.
func assertComparableTo(at, bt Type, aNil, bNil bool) {
	check := func(t Type, otherNil bool) {
		if t == nil || otherNil {
			return
		}
		switch baseOf(t).(type) {
		case *SliceType, *FuncType, *MapType:
			panic(fmt.Sprintf("%v can only be compared to nil", t))
		}
		if !isComparable(t) {
			panic(fmt.Sprintf("%v is not comparable", t))
		}
	}
	check(at, bNil)
	check(bt, aNil)
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
	panic(fmt.Sprintf("invalid argument: %v for built-in %s", t, name))
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
		panic(fmt.Sprintf("invalid argument: index %d out of bounds [0:%d]", i, length))
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
			panic(fmt.Sprintf("invalid argument: index %d out of bounds [0:%d]", i, length+1))
		}
		if prev >= 0 && i < prev {
			panic(fmt.Sprintf("invalid slice indices: %d < %d", i, prev))
		}
		prev = i
	}
}

// assertLiteralIndices rejects a repeated constant index in an array or
// slice literal, and one at or past length when it is known (-1 for a
// slice); elements without a key take the next index. Charged per element.
func assertLiteralIndices(gm store.GasMeter, n *CompositeLitExpr, length int64) {
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
			if length >= 0 && idx >= length {
				panic(fmt.Sprintf("invalid argument: index %d out of bounds [0:%d]", idx, length))
			}
		} else if length >= 0 && idx >= length {
			panic(fmt.Sprintf("index %d is out of bounds (>= %d)", idx, length))
		}
		if seen[idx] {
			panic(fmt.Sprintf("duplicate index %d in array or slice literal", idx))
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
		panic(fmt.Sprintf("%s (type) is not an expression", exprString(x)))
	}
	xs := exprString(x)
	if _, ok := x.(*FuncLitExpr); ok {
		xs = fmt.Sprintf("(%s literal)", typeString(t)) // as go/types
	}
	panic(fmt.Sprintf("%s (value of type %s) is not used", xs, typeString(t)))
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
		panic(fmt.Sprintf("index %s must be integer constant", exprString(key)))
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

// assertSwitchTag checks an expression switch's tag: comparable, or a
// slice, map or func, which the clauses may compare to nil only.
func assertSwitchTag(t Type) {
	if t == nil {
		return
	}
	switch baseOf(t).(type) {
	case *SliceType, *FuncType, *MapType:
		return
	}
	if !isComparable(t) {
		panic(fmt.Sprintf("cannot switch on %v (%v is not comparable)", t, t))
	}
}
