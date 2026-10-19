package gnolang

import (
	"fmt"
)

// Declaration and literal rules the preprocessor enforces (Go spec:
// "Short variable declarations", "Function types", "Composite literals",
// "For statements", "Type switches").

// checkRedefinedTargets: a name on the left of := that this block already
// declares is assigned, not redeclared, so the value must be assignable
// to its type; an untyped constant converts to it. Runs before the
// values take their default types, so untyped constants are still seen.
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
			if cx, ok := valueExprs[i].(*ConstExpr); ok && isUntyped(cx.T) {
				if err := checkAssignableTo(store, n, cx.T, et); err != nil {
					panic(fmt.Sprintf("cannot use %s (untyped %s constant) as %v value in assignment",
						constString(cx), untypedKindName(cx.T), et))
				}
				convertConst(store, bn, n, cx, et) // representability
				continue
			}
			checkOrConvertType(store, bn, n, &valueExprs[i], et)
		}
	}
}

// untypedKindName is the Go name of an untyped constant kind.
func untypedKindName(t Type) string {
	switch t {
	case UntypedBigintType:
		return "int"
	case UntypedBigdecType:
		return "float"
	case UntypedRuneType:
		return "rune"
	case UntypedStringType:
		return "string"
	case UntypedBoolType:
		return "bool"
	}
	return t.String()
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

// invalidStructKey is the error for a keyed struct literal element whose
// key is not a field name, "" when it is one. Checked before the key is
// resolved as an expression, so every bad key of a literal is reported.
func invalidStructKey(key Expr) string {
	if _, ok := key.(*NameExpr); ok {
		return ""
	}
	return fmt.Sprintf("invalid field name %s in struct literal", exprString(key))
}

// assertStructField: the key of a keyed struct literal names a field of
// st (the literal's type lt as written).
func assertStructField(st *StructType, lt Type, key *NameExpr) {
	for i := range st.Fields {
		if st.Fields[i].Name == key.Name {
			return
		}
	}
	panicAt(fmt.Sprintf("unknown field %s in struct literal of type %s", key.Name, lt), key)
}
