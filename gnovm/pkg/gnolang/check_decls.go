package gnolang

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/store"
)

// Declaration and literal rules the preprocessor enforces (Go spec:
// "Short variable declarations", "Function types", "Composite literals",
// "For statements", "Type switches").

// checkRedefinedTargets: a name on the left of := that this block already
// declares is assigned, not redeclared, so the value must be assignable
// to its type; an untyped constant converts to it.
func checkRedefinedTargets(store Store, bn BlockNode, n Node, nameExprs []*NameExpr, valueExprs []Expr) {
	node := skipFile(bn)
	var results []Type
	if len(valueExprs) == 1 && len(nameExprs) > 1 {
		if tt, ok := evalStaticTypeOfRaw(store, bn, valueExprs[0]).(*tupleType); ok {
			results = tt.Elts
		}
	}
	for i, nx := range nameExprs {
		if nx.Name == blankIdentifier {
			continue
		}
		src := definedBy(node, nx.Name)
		if src == nil || src.Origin == n {
			continue // new, or reserved for this statement
		}
		et := node.GetStaticTypeOf(store, nx.Name)
		if et == nil {
			continue
		}
		switch {
		case results != nil:
			if i < len(results) {
				mustAssignableTo(store, n, results[i], et)
			}
		case i < len(valueExprs):
			if cx, ok := valueExprs[i].(*ConstExpr); ok && wasUntyped(cx) {
				// Already given its default type; convert as Go would.
				convertConst(store, bn, n, cx, et)
				continue
			}
			checkOrConvertType(store, bn, n, &valueExprs[i], et)
		}
	}
}

// wasUntyped reports whether a folded constant was written untyped.
func wasUntyped(cx *ConstExpr) bool {
	if cx.T == nil || isUntyped(cx.T) {
		return true
	}
	_, lit := cx.Source.(*BasicLitExpr)
	return lit
}

// assertFuncTypeNames rejects a variadic parameter that is not the last
// one. (The parser already rejects a name used twice.)
func assertFuncTypeNames(n *FuncTypeExpr) {
	for i := 0; i+1 < len(n.Params); i++ {
		if st, ok := n.Params[i].Type.(*SliceTypeExpr); ok && st.Vrd {
			panicAt("can only use ... with final parameter in list", &n.Params[i])
		}
	}
}

// assertStructLiteralCount: an unkeyed struct literal gives every field
// a value or none.
func assertStructLiteralCount(n *CompositeLitExpr, st *StructType, lt Type) {
	switch {
	case len(n.Elts) == 0:
	case len(n.Elts) < len(st.Fields):
		panic(fmt.Sprintf("too few values in struct literal of type %v", lt))
	case len(n.Elts) > len(st.Fields):
		panic(fmt.Sprintf("too many values in struct literal of type %v", lt))
	}
}

// assertLiteralIndicesFit rejects an array literal element whose index,
// written or implied, is not below the array length.
func assertLiteralIndicesFit(gm store.GasMeter, n *CompositeLitExpr, length int) {
	chargeCPUGas(gm, OpCPUSlopeCheckNode*int64(len(n.Elts)))
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
		if idx >= int64(length) {
			panic(fmt.Sprintf("index %d out of bounds [0:%d]", idx, length))
		}
		next = idx + 1
	}
}
