package gnolang

import "fmt"

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
					failf(CodeAssignability, nil, "cannot use %s (untyped %s constant) as %v value in assignment",
						constString(cx), untypedKindName(cx.T), et)
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
			fail(CodeVariadicLast, "can only use ... with final parameter in list", &n.Params[i])
		}
	}
}

// assertStructLiteralCount: an unkeyed struct literal gives every field
// a value or none.
func assertStructLiteralCount(n *CompositeLitExpr, st *StructType, lt Type) {
	switch {
	case len(n.Elts) == 0:
	case len(n.Elts) < len(st.Fields):
		failf(CodeStructLiteralCount, nil, "too few values in struct literal of type %v", lt)
	case len(n.Elts) > len(st.Fields):
		failf(CodeStructLiteralCount, nil, "too many values in struct literal of type %v", lt)
	}
}

// assertNotImported: a package-level declaration may not take the name an
// import gives any file of the package (go/types: DuplicateDecl, reported
// at the declaration). Charged per file.
func assertNotImported(store Store, last BlockNode, nx *NameExpr) {
	if nx.Name == blankIdentifier {
		return
	}
	pn, ok := skipFile(last).(*PackageNode)
	if !ok || pn.FileSet == nil {
		return
	}
	chargeCPUGas(preprocessGasMeterOf(store), OpCPUSlopeCheckNode*int64(len(pn.FileSet.Files)))
	for _, fn := range pn.FileSet.Files {
		idx, exists := fn.GetLocalIndex(nx.Name)
		if !exists {
			continue
		}
		if src := fn.GetStaticBlock().NameSources[idx]; src.Type == NSImportDecl {
			path := src.Origin.(*ImportDecl).PkgPath
			pkg, ok := expectedPkgName(path)
			if !ok {
				pkg = path
			}
			failf(CodeImportName, []Node{nx}, "%s already declared through import of package %s (%q)", nx.Name, pkg, path)
		}
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
	failf(CodeStructField, []Node{key}, "unknown field %s in struct literal of type %s", key.Name, lt)
}

// assertSelectorNotBlank: a blank field or method is never found.
func assertSelectorNotBlank(n *SelectorExpr, xt Type) {
	if n.Sel == blankIdentifier {
		failf(CodeNoMethod, nil, "%s undefined (type %s has no field or method _)", exprString(n), xt)
	}
}

// receiverDeclaredType returns the declared type a method receiver type
// rt (or what it points to) names. It must be declared in pkgPath: a
// basic or foreign type cannot be given new methods, nothing else is a
// receiver type.
func receiverDeclaredType(rt Type, pkgPath string, at Node) *DeclaredType {
	bt := rt
	if pt, ok := rt.(*PointerType); ok {
		bt = pt.Elem()
	}
	switch t := bt.(type) {
	case *DeclaredType:
		if t.PkgPath == pkgPath {
			return t
		}
	case PrimitiveType:
	default:
		failf(CodeReceiverType, []Node{at}, "invalid receiver type %s", rt)
	}
	failf(CodeReceiverType, []Node{at}, "cannot define new methods on non-local type %s", bt)
	return nil
}

// assertNamesType: the name nx a type declaration gives as its type,
// resolved to the slot tv (nil: not yet defined) or to the package-level
// declaration dep, is a type and not a constant, variable or function.
func assertNamesType(nx *NameExpr, tv *TypedValue, dep Decl) {
	notType := false
	if tv != nil && tv.T != nil {
		_, isType := tv.T.(*TypeType)
		notType = !isType
	}
	switch dep.(type) {
	case *ValueDecl, *FuncDecl:
		notType = true
	}
	if notType {
		failf(CodeNotAType, []Node{nx}, "%s is not a type", nx.Name)
	}
}
