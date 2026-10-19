package gnolang

import (
	"fmt"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/store"
)

// Type rules the preprocessor enforces (Go spec: "Map types", "Struct
// types", "Interface types", "Array types", "Switch statements", "Type
// assertions", "Appending to and copying slices").

// typeComplete reports whether every declared type reachable from t
// through struct fields and array elements already has its base set.
// Comparability of an incomplete type is unknown, so its checks run
// again when the enclosing declaration finishes (see checkTypeShape).
func typeComplete(gm store.GasMeter, t Type) bool {
	seen := map[Type]bool{}
	var walk func(t Type) bool
	walk = func(t Type) bool {
		if seen[t] {
			return true
		}
		seen[t] = true
		chargeCPUGas(gm, OpCPUSlopeCheckNode)
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
func mapKeyError(gm store.GasMeter, kt Type) string {
	if kt == nil || !typeComplete(gm, kt) {
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
func checkTypeExprShape(gm store.GasMeter, tx Expr) {
	Transcribe(tx, func(ns []Node, ftype TransField, index int, n Node, stage TransStage) (Node, TransCtrl) {
		if stage != TRANS_LEAVE {
			return n, TRANS_CONTINUE
		}
		chargeCPUGas(gm, OpCPUSlopeCheckNode)
		t, _ := n.GetAttribute(ATTR_TYPE_VALUE).(Type)
		switch x := n.(type) {
		case *MapTypeExpr:
			if mt, ok := t.(*MapType); ok {
				if msg := mapKeyError(gm, mt.Key); msg != "" {
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
// An error is positioned at its spelling in the owning declaration.
func checkTypeShape(gm store.GasMeter, last BlockNode, dt *DeclaredType) {
	seen := map[Type]bool{}
	var walk func(t Type, owner *DeclaredType)
	walk = func(t Type, owner *DeclaredType) {
		if t == nil || seen[t] {
			return
		}
		seen[t] = true
		chargeCPUGas(gm, OpCPUSlopeCheckNode)
		switch ct := t.(type) {
		case *DeclaredType:
			if ct.PkgPath == dt.PkgPath {
				walk(ct.Base, ct)
			}
		case *PointerType:
			walk(ct.Elt, owner)
		case *SliceType:
			walk(ct.Elt, owner)
		case *ArrayType:
			walk(ct.Elt, owner)
		case *ChanType:
			walk(ct.Elt, owner)
		case *MapType:
			if msg := mapKeyError(gm, ct.Key); msg != "" {
				x, _ := typeExprIn(gm, last, owner, ct).(*MapTypeExpr)
				if x == nil {
					fail(CodeMapKey, msg)
				}
				fail(CodeMapKey, msg, x.Key, x)
			}
			walk(ct.Key, owner)
			walk(ct.Value, owner)
		case *FuncType:
			for _, p := range ct.Params {
				walk(p.Type, owner)
			}
			for _, r := range ct.Results {
				walk(r.Type, owner)
			}
		case *StructType:
			for i, f := range ct.Fields {
				if msg := embeddedFieldError(f); f.Embedded && msg != "" {
					x, _ := typeExprIn(gm, last, owner, ct).(*StructTypeExpr)
					if x == nil || len(x.Fields) != len(ct.Fields) {
						fail(CodeEmbeddedField, msg)
					}
					fail(CodeEmbeddedField, msg, &x.Fields[i], x)
				}
				walk(f.Type, owner)
			}
		case *InterfaceType:
			for _, m := range ct.Methods {
				walk(m.Type, owner)
			}
		}
	}
	walk(dt, dt)
}

// typeExprIn returns the type expression in owner's package-level
// declaration that evaluated to t, or nil (local or unknown owner).
func typeExprIn(gm store.GasMeter, last BlockNode, owner *DeclaredType, t Type) (found Expr) {
	if owner == nil {
		return nil
	}
	fset := packageOf(last).FileSet
	if fset == nil {
		return nil
	}
	_, slot, ok := fset.GetDeclForSafe(owner.Name)
	if !ok {
		return nil
	}
	td, ok := (*slot).(*TypeDecl)
	if !ok {
		return nil
	}
	tx := td.Type
	if ctx, ok := tx.(*constTypeExpr); ok {
		tx = ctx.Source
	}
	Transcribe(tx, func(ns []Node, ftype TransField, index int, n Node, stage TransStage) (Node, TransCtrl) {
		if stage != TRANS_ENTER {
			return n, TRANS_CONTINUE
		}
		chargeCPUGas(gm, OpCPUSlopeCheckNode)
		if x, ok := n.(Expr); ok && n.GetAttribute(ATTR_TYPE_VALUE) == t {
			found = x
			return n, TRANS_EXIT
		}
		return n, TRANS_CONTINUE
	})
	return found
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
	gm := preprocessGasMeterOf(store)
	seen := map[MapKey]bool{}
	for i := range clauses {
		for _, cx := range clauses[i].Cases {
			chargeCPUGas(gm, OpCPUSlopeCheckNode)
			c, ok := cx.(*ConstExpr)
			if !ok || c.T == nil || c.T.Kind() == BoolKind || isInterfaceConversion(c) {
				continue
			}
			if _, ok := baseOf(c.T).(PrimitiveType); !ok {
				continue
			}
			key, isNaN := c.ComputeMapKey(gm, store, false)
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
	case *SliceTypeExpr:
		if x.Vrd {
			return "..." + exprString(x.Elt)
		}
		return "[]" + exprString(x.Elt)
	case *ArrayTypeExpr:
		if x.Len == nil {
			return "[...]" + exprString(x.Elt)
		}
		return "[" + exprString(x.Len) + "]" + exprString(x.Elt)
	case *MapTypeExpr:
		return "map[" + exprString(x.Key) + "]" + exprString(x.Value)
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
	switch src := c.Source.(type) {
	case *CallExpr:
		// A conversion the preprocessor inserted has no span: print
		// the operand as written.
		if _, ok := src.Func.(*constTypeExpr); ok && len(src.Args) == 1 && src.GetSpan().IsZero() {
			return exprString(src.Args[0])
		}
		return exprString(src)
	case *BasicLitExpr, *BinaryExpr, *UnaryExpr, *NameExpr, *SelectorExpr, *ConstExpr:
		return exprString(src) // as written, also after folding or conversion
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
		// Print the value alone; a declared type is named by the wrapper.
		tv := c.TypedValue
		tv.T = baseOf(tv.T)
		v = tv.Sprint(nil)
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
// nil case types are always possible. gm meters the method-set walk.
func dynamicTypeError(gm store.GasMeter, it, ct Type) error {
	iface, ok := baseOf(it).(*InterfaceType)
	if !ok || ct == nil || iface.Generic != "" {
		return nil
	}
	if _, ok := baseOf(ct).(*InterfaceType); ok {
		return nil
	}
	return iface.checkImplementedBy(gm, ct)
}

// assertCopyArgs checks copy(dst, src): dst a slice, src a slice of the
// identical element type (tags included), or a string when dst is a
// byte slice.
func assertCopyArgs(gm store.GasMeter, dt, st Type) {
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
	if sok && !tagsIdentical(gm, ds.Elt, ss.Elt) {
		failf(CodeCopyArgs, nil, "arguments to copy have different element types %v and %v: struct tags differ", ds.Elt, ss.Elt)
	}
	if isString && ds.Elt != Uint8Type {
		failf(CodeCopyArgs, nil, "arguments to copy have different element types %v and byte", ds.Elt)
	}
}
