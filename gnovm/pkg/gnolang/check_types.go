package gnolang

import (
	"fmt"
	"strings"
)

// Type rules the preprocessor enforces (Go spec: "Map types", "Struct
// types", "Interface types", "Array types", "Switch statements", "Type
// assertions", "Appending to and copying slices").

// typeComplete reports whether every declared type reachable from t
// through struct fields and array elements already has its base set.
// Comparability of an incomplete type is unknown, so its checks run
// again when the enclosing declaration finishes (see checkTypeShape).
func typeComplete(t Type) bool {
	seen := map[Type]bool{}
	var walk func(t Type) bool
	walk = func(t Type) bool {
		if seen[t] {
			return true
		}
		seen[t] = true
		switch ct := t.(type) {
		case *DeclaredType:
			return ct.Base != nil && walk(ct.Base)
		case *ArrayType:
			return walk(ct.Elt)
		case *StructType:
			for _, f := range ct.Fields {
				if !walk(f.Type) {
					return false
				}
			}
		}
		return true
	}
	return walk(t)
}

// mapKeyError reports why kt cannot key a map, or "" when it can or is
// still incomplete (then checkTypeExprShape decides later).
func mapKeyError(kt Type) string {
	if kt == nil || !typeComplete(kt) {
		return ""
	}
	if _, ok := baseOf(kt).(*ChanType); ok {
		return ""
	}
	if !isComparable(kt) {
		return fmt.Sprintf("invalid map key type %v", kt)
	}
	return ""
}

// embeddedFieldError reports why ft cannot be an embedded field: a
// pointer to a pointer or to an interface, or a named pointer type.
func embeddedFieldError(ft FieldType) string {
	if pt, ok := ft.Type.(*PointerType); ok {
		switch baseOf(pt.Elt).(type) {
		case *PointerType:
			return "embedded field type cannot be a pointer"
		case *InterfaceType:
			return "embedded field type cannot be a pointer to an interface"
		}
		return ""
	}
	if _, ok := baseOf(ft.Type).(*PointerType); ok {
		return "embedded field type cannot be a pointer"
	}
	return ""
}

// checkTypeExprShape re-runs the map key and embedded field rules over
// the type expression of a finished declaration, positioned at the
// spelling: checks skipped while the type was incomplete now decide.
func checkTypeExprShape(tx Expr) {
	Transcribe(tx, func(ns []Node, ftype TransField, index int, n Node, stage TransStage) (Node, TransCtrl) {
		if stage != TRANS_LEAVE {
			return n, TRANS_CONTINUE
		}
		t, _ := n.GetAttribute(ATTR_TYPE_VALUE).(Type)
		switch x := n.(type) {
		case *MapTypeExpr:
			if mt, ok := t.(*MapType); ok {
				if msg := mapKeyError(mt.Key); msg != "" {
					fail(CodeMapKey, msg, x.Key, x)
				}
			}
		case *StructTypeExpr:
			if st, ok := t.(*StructType); ok && len(st.Fields) == len(x.Fields) {
				for i, f := range st.Fields {
					if msg := embeddedFieldError(f); f.Embedded && msg != "" {
						fail(CodeEmbeddedField, msg, &x.Fields[i], x)
					}
				}
			}
		}
		return n, TRANS_CONTINUE
	})
}

// checkTypeShape re-runs the map key and embedded field rules over the
// package types reachable from dt once its base is set: a type finished
// earlier may hold a map keyed by dt, skipped while dt was incomplete.
func checkTypeShape(dt *DeclaredType) {
	seen := map[Type]bool{}
	var walk func(t Type)
	walk = func(t Type) {
		if t == nil || seen[t] {
			return
		}
		seen[t] = true
		switch ct := t.(type) {
		case *DeclaredType:
			if ct.PkgPath == dt.PkgPath {
				walk(ct.Base)
			}
		case *PointerType:
			walk(ct.Elt)
		case *SliceType:
			walk(ct.Elt)
		case *ArrayType:
			walk(ct.Elt)
		case *ChanType:
			walk(ct.Elt)
		case *MapType:
			if msg := mapKeyError(ct.Key); msg != "" {
				fail(CodeMapKey, msg)
			}
			walk(ct.Key)
			walk(ct.Value)
		case *FuncType:
			for _, p := range ct.Params {
				walk(p.Type)
			}
			for _, r := range ct.Results {
				walk(r.Type)
			}
		case *StructType:
			for _, f := range ct.Fields {
				if msg := embeddedFieldError(f); f.Embedded && msg != "" {
					fail(CodeEmbeddedField, msg)
				}
				walk(f.Type)
			}
		case *InterfaceType:
			for _, m := range ct.Methods {
				walk(m.Type)
			}
		}
	}
	walk(dt)
}

// assertArrayLength rejects a negative constant array length.
func assertArrayLength(n int64) {
	if n < 0 {
		failf(CodeArrayLength, nil, "invalid array length %d", n)
	}
}

// dupCase is a duplicate expression-switch case: a message and where to
// report it.
type dupCase struct {
	msg string
	at  []Node
}

// duplicateSwitchCases finds constant cases of the same type and value
// in an expression switch. Booleans are exempt, as in Go.
func duplicateSwitchCases(store Store, clauses []SwitchClauseStmt) (dups []dupCase) {
	seen := map[MapKey]bool{}
	for i := range clauses {
		for _, cx := range clauses[i].Cases {
			c, ok := cx.(*ConstExpr)
			if !ok || c.T == nil || c.T.Kind() == BoolKind || isInterfaceConversion(c) {
				continue
			}
			if _, ok := baseOf(c.T).(PrimitiveType); !ok {
				continue
			}
			key, isNaN := c.ComputeMapKey(nil, store, false)
			if isNaN {
				continue
			}
			if seen[key] {
				dups = append(dups, dupCase{
					msg: fmt.Sprintf("duplicate case %s in expression switch", constString(c)),
					at:  []Node{cx, &clauses[i]},
				})
			}
			seen[key] = true
		}
	}
	return dups
}

// isInterfaceConversion reports whether c was written as a conversion to
// an interface type, which Go does not treat as a constant.
func isInterfaceConversion(c *ConstExpr) bool {
	call, ok := c.Source.(*CallExpr)
	if !ok {
		return false
	}
	var t Type
	if ct, ok := call.Func.(*constTypeExpr); ok {
		t = ct.Type
	} else {
		t, _ = call.Func.GetAttribute(ATTR_TYPE_VALUE).(Type)
	}
	return t != nil && t.Kind() == InterfaceKind
}

// exprString prints an expression for a message: names bare, constants
// as written, selectors and indexes composed from their parts.
func exprString(x Expr) string {
	switch x := x.(type) {
	case *NameExpr:
		return string(x.Name)
	case *BasicLitExpr:
		return x.Value
	case *ConstExpr:
		return constString(x)
	case *constTypeExpr:
		return x.Type.String()
	case *BinaryExpr:
		return exprString(x.Left) + " " + x.Op.TokenString() + " " + exprString(x.Right)
	case *UnaryExpr:
		return x.Op.TokenString() + exprString(x.X)
	case *SelectorExpr:
		return exprString(x.X) + "." + string(x.Sel)
	case *IndexExpr:
		return exprString(x.X) + "[" + exprString(x.Index) + "]"
	case *StarExpr:
		return "*" + exprString(x.X)
	case *CallExpr:
		args := make([]string, len(x.Args))
		for i, a := range x.Args {
			args[i] = exprString(a)
		}
		return exprString(x.Func) + "(" + strings.Join(args, ", ") + ")"
	}
	return x.String()
}

// typeString names a type for a message; nil is the untyped nil.
func typeString(t Type) string {
	if t == nil {
		return "untyped nil"
	}
	return t.String()
}

// constString prints a folded constant Go style: the value for an
// untyped constant, T(value) for a typed one, strings quoted.
func constString(c *ConstExpr) string {
	switch c.Source.(type) {
	case *BasicLitExpr, *CallExpr, *BinaryExpr, *UnaryExpr, *NameExpr, *SelectorExpr:
		return exprString(c.Source) // as written, also after folding
	}
	var v string
	switch {
	case c.T == nil:
		return "nil"
	case c.T.Kind() == StringKind:
		v = fmt.Sprintf("%q", c.GetString())
	case c.T.Kind() == BoolKind:
		v = fmt.Sprintf("%t", c.GetBool())
	case isUntyped(c.T):
		return c.TypedValue.Sprint(nil)
	default:
		v = c.TypedValue.Sprint(nil)
	}
	if isUntyped(c.T) {
		return v
	}
	return fmt.Sprintf("%v(%s)", c.T, v)
}

// assertSingleDefault rejects a second default clause in a switch.
func assertSingleDefault(clauses []SwitchClauseStmt) {
	seen := false
	for i := range clauses {
		if len(clauses[i].Cases) != 0 {
			continue
		}
		if seen {
			fail(CodeMultipleDefaults, "multiple defaults in switch", &clauses[i])
		}
		seen = true
	}
}

// dynamicTypeError reports why a value of interface type it can never
// have the concrete dynamic type ct, or nil when it can. Interface and
// nil case types are always possible.
func dynamicTypeError(it, ct Type) error {
	iface, ok := baseOf(it).(*InterfaceType)
	if !ok || ct == nil || iface.Generic != "" {
		return nil
	}
	if _, ok := baseOf(ct).(*InterfaceType); ok {
		return nil
	}
	return iface.checkImplementedBy(nil, ct)
}

// assertCopyArgs checks copy(dst, src): dst a slice, src a slice of the
// same element type, or a string when dst is a byte slice.
func assertCopyArgs(dt, st Type) {
	ds, dok := baseOf(dt).(*SliceType)
	ss, sok := baseOf(st).(*SliceType)
	isString := st != nil && st.Kind() == StringKind
	if !dok || !sok && !isString {
		failf(CodeCopyArgs, nil, "invalid argument: copy expects slice arguments; found %s and %s",
			typeString(dt), typeString(st))
	}
	if sok && ds.Elt.TypeID() != ss.Elt.TypeID() {
		failf(CodeCopyArgs, nil, "arguments to copy have different element types %v and %v", ds.Elt, ss.Elt)
	}
	if isString && ds.Elt != Uint8Type {
		failf(CodeCopyArgs, nil, "arguments to copy have different element types %v and byte", ds.Elt)
	}
}

// assertSwitchTag rejects a switch on an untyped nil or on a value that
// == cannot compare; a slice, map or func tag is allowed, its cases must
// be nil.
func assertSwitchTag(x Expr, tt Type) {
	if tt == nil {
		fail(CodeSwitchTag, "use of untyped nil in switch expression", x)
	}
	switch baseOf(tt).(type) {
	case *SliceType, *MapType, *FuncType:
		return
	}
	if !isComparable(tt) {
		failf(CodeSwitchTag, []Node{x}, "cannot switch on %s (%v is not comparable)", exprString(x), tt)
	}
}

// assertMethodExprReceiver: the method expression T.m needs *T when m
// has a pointer receiver (Go spec, "Method expressions").
func assertMethodExprReceiver(xt Type, dt *DeclaredType, sel Name, recv Type, isPtr bool) {
	if _, ptrRecv := recv.(*PointerType); ptrRecv && !isPtr {
		failf(CodeMethodExprReceiver, nil,
			"invalid method expression %s.%s (needs pointer receiver (*%s).%s)", dt.Name, sel, dt.Name, sel)
	}
}

// promotedMethodExpr lowers the method expression n (X.Sel, X a type)
// whose method is promoted through an embedded field to
// func(r X, args...) results { return r.Sel(args...) }.
func promotedMethodExpr(store Store, last BlockNode, pkgPath string, n *SelectorExpr, xt Type, isPtr bool) *FuncLitExpr {
	trail, hasPtr, rcvr, ft, status := findEmbeddedFieldType(preprocessGasMeterOf(store), pkgPath, xt, n.Sel)
	if status != embedLookupFound || len(trail) == 0 || trail[len(trail)-1].Type == VPField || trail[len(trail)-1].Type == VPDerefField {
		failf(CodeNoMethod, nil, "%s.%s undefined (type %v has no method %s)", exprString(n.X), n.Sel, xt, n.Sel)
	}
	if _, ptrRecv := rcvr.(*PointerType); ptrRecv && !isPtr && !hasPtr {
		failf(CodeMethodExprReceiver, nil,
			"invalid method expression %s.%s (needs pointer receiver (*%s).%s)", exprString(n.X), n.Sel, exprString(n.X), n.Sel)
	}
	mt := ft.(*FuncType)
	recv := Name(".recv")
	params := FieldTypeExprs{{NameExpr: *Nx(recv), Type: toConstTypeExpr(last, n.X, xt)}}
	args := make([]any, len(mt.Params))
	for i, p := range mt.Params {
		name := Name(fmt.Sprintf(".arg%d", i))
		params = append(params, FieldTypeExpr{NameExpr: *Nx(name), Type: toConstTypeExpr(last, Nx(name), p.Type)})
		args[i] = Nx(name)
	}
	results := make(FieldTypeExprs, len(mt.Results))
	for i, r := range mt.Results {
		results[i] = FieldTypeExpr{Type: toConstTypeExpr(last, Nx(".res"), r.Type)}
	}
	call := Call(Sel(Nx(recv), n.Sel), args...)
	call.Varg = mt.HasVarg()
	var body Body
	if len(results) > 0 {
		body = Body{Return(call)}
	} else {
		body = Body{&ExprStmt{X: call}}
	}
	lit := &FuncLitExpr{Type: *FuncT(params, results), Body: body}
	lit.SetSpan(n.GetSpan())
	loc := last.GetLocation()
	loc.Span = n.GetSpan()
	lit.SetLocation(loc)
	return lit
}
