package gnolang

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/store"
)

// Assignability rules the preprocessor enforces (Go spec: "Type
// identity", "Assignment statements", "Address operators").

// tagsIdentical reports whether two types with the same TypeID also
// agree on struct tags, which TypeID leaves out but type identity
// includes. Declared types are identical by name alone.
func tagsIdentical(gm store.GasMeter, a, b Type) bool {
	chargeCPUGas(gm, OpCPUSlopeCheckNode)
	switch at := a.(type) {
	case *StructType:
		bt, ok := b.(*StructType)
		if !ok || len(at.Fields) != len(bt.Fields) {
			return false
		}
		for i := range at.Fields {
			if at.Fields[i].Tag != bt.Fields[i].Tag ||
				!tagsIdentical(gm, at.Fields[i].Type, bt.Fields[i].Type) {
				return false
			}
		}
	case *PointerType:
		if bt, ok := b.(*PointerType); ok {
			return tagsIdentical(gm, at.Elt, bt.Elt)
		}
	case *ArrayType:
		if bt, ok := b.(*ArrayType); ok {
			return tagsIdentical(gm, at.Elt, bt.Elt)
		}
	case *SliceType:
		if bt, ok := b.(*SliceType); ok {
			return tagsIdentical(gm, at.Elt, bt.Elt)
		}
	case *MapType:
		if bt, ok := b.(*MapType); ok {
			return tagsIdentical(gm, at.Key, bt.Key) && tagsIdentical(gm, at.Value, bt.Value)
		}
	case *FuncType:
		if bt, ok := b.(*FuncType); ok {
			return fieldTagsIdentical(gm, at.Params, bt.Params) && fieldTagsIdentical(gm, at.Results, bt.Results)
		}
	}
	return true
}

func fieldTagsIdentical(gm store.GasMeter, a, b []FieldType) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !tagsIdentical(gm, a[i].Type, b[i].Type) {
			return false
		}
	}
	return true
}

// inMapValue reports whether lx is a struct field or array element of a
// value held in a map, reached without a pointer: it is not addressable.
func inMapValue(store Store, last BlockNode, lx Expr) bool {
	chargeCPUGas(preprocessGasMeterOf(store), OpCPUSlopeCheckNode)
	var inner Expr
	switch x := lx.(type) {
	case *SelectorExpr:
		if _, ok := baseOf(evalStaticTypeOf(store, last, x.X)).(*StructType); !ok {
			return false
		}
		inner = x.X
	case *IndexExpr:
		if _, ok := baseOf(evalStaticTypeOf(store, last, x.X)).(*ArrayType); !ok {
			return false
		}
		inner = x.X
	default:
		return false
	}
	if ix, ok := inner.(*IndexExpr); ok {
		if _, ok := baseOf(evalStaticTypeOf(store, last, ix.X)).(*MapType); ok {
			return true
		}
	}
	return inMapValue(store, last, inner)
}

// mustAssignableTo panics if xt is not assignable to dt. store supplies the
// per-tx preprocess gas meter that bills the interface-satisfaction walk when
// dt is an interface (see checkAssignableTo).
func mustAssignableTo(store Store, n Node, xt, dt Type) {
	err := checkAssignableTo(store, n, xt, dt)
	if err != nil {
		if debug {
			debug.Printf("checkAssignableTo fail: %v\n", err)
		}
		panic(err.Error())
	}
}

// checkValDefineMismatch checks for mismatch between the number of variables and values in a ValueDecl or AssignStmt.
func checkValDefineMismatch(n Node) {
	var (
		valueDecl *ValueDecl
		assign    *AssignStmt
		values    []Expr
		numNames  int
		numValues int
	)

	switch x := n.(type) {
	case *ValueDecl:
		valueDecl = x
		numNames = len(valueDecl.NameExprs)
		numValues = len(valueDecl.Values)
		values = valueDecl.Values
	case *AssignStmt:
		if x.Op != DEFINE {
			return
		}

		assign = x
		numNames = len(assign.Lhs)
		numValues = len(assign.Rhs)
		values = assign.Rhs
	default:
		panic(fmt.Sprintf("unexpected node type %T", n))
	}

	if numValues == 0 || numValues == numNames {
		return
	}

	// Special case for single value.
	// If the value is a call expression, type assertion, or index expression,
	// it can be assigned to multiple variables.
	if numValues == 1 {
		switch values[0].(type) {
		case *CallExpr:
			return
		case *TypeAssertExpr:
			if numNames != 2 {
				panic(fmt.Sprintf("assignment mismatch: %d variable(s) but %d value(s)", numNames, numValues))
			}
			return
		case *IndexExpr:
			if numNames != 2 {
				panic(fmt.Sprintf("assignment mismatch: %d variable(s) but %d value(s)", numNames, numValues))
			}
			return
		}
	}

	if valueDecl != nil {
		if numNames > numValues {
			failf(CodeMissingConstInit, nil, "missing init expr for %s", valueDecl.NameExprs[numValues].Name)
		}
		failf(CodeMissingConstInit, []Node{values[numNames]}, "extra init expr %s", values[numNames].String())
	}

	panic(fmt.Sprintf("assignment mismatch: %d variable(s) but %d value(s)", numNames, numValues))
}

// Assert that xt can be assigned as dt (dest type).
// store supplies the per-tx preprocess gas meter (preprocessGasMeterOf) that
// bills the interface-satisfaction walk when dt is an interface; a nil store
// (tests) leaves it unmetered.
func checkAssignableTo(store Store, n Node, xt, dt Type) (err error) {
	if debug {
		debug.Printf("checkAssignableTo, xt: %v dt: %v \n", xt, dt)
	}
	// dt (the destination type) must be non-nil. A nil dt is ambiguous: it
	// is the static type both of a blank/absent target (`_ = x`, blank range
	// operands), which should be skipped, and of an untyped-nil lvalue
	// (`for k, nil = range m`), which must be rejected. Callers disambiguate
	// upstream — skipping blank operands and guarding "no destination type"
	// paths with `if t != nil` — so a nil reaching here is a caller bug,
	// caught loudly instead of silently dropping the check.
	if dt == nil {
		panic("should not happen: nil dt in checkAssignableTo (blank targets must be skipped by the caller)")
	}
	// case0
	if xt == nil { // untyped nil, see test/files/types/eql_0f18
		if dt.Kind() == InterfaceKind {
			return nil
		}
		if !mayBeNil(dt) {
			switch n := n.(type) {
			case *ValueDecl:
				return errors.New("cannot use nil as %v value in variable declaration", dt)
			case *AssignStmt:
				return errors.New("cannot use nil as %v value in assignment", dt)
			case *CompositeLitExpr:
				return errors.New("cannot use nil as %v value in array, slice literal or map literal", dt)
			case *CallExpr:
				return errors.New("cannot use nil as %v value in argument to %v", dt, n.Func)
			case *BinaryExpr:
				return errors.New("invalid operation: %v (mismatched types %v and untyped nil)", n, dt)
			default:
				return errors.New("cannot use nil as %v value", dt)
			}
		}
		return nil
	}
	// case3
	if dt.Kind() == InterfaceKind { // note native interface
		if idt, ok := baseOf(dt).(*InterfaceType); ok {
			if idt.IsEmptyInterface() { // XXX, can this be merged with isImplementedBy?
				// if dt is an empty Gno interface, any x ok.
				return nil // ok
			} else if err := idt.checkImplementedByStatic(preprocessGasMeterOf(store), xt); err == nil {
				// if dt implements idt, ok.
				return nil // ok
			} else {
				return errors.New(
					"%s does not implement %s (%s)",
					xt.String(),
					dt.String(),
					err.Error())
			}
		} else {
			panic("should not happen")
		}
	}

	// Special case of xt or dt is *DeclaredType,
	// allow implicit conversion unless both are declared.
	// TODO simplify with .IsNamedType().
	if dxt, ok := xt.(*DeclaredType); ok {
		if ddt, ok := dt.(*DeclaredType); ok {
			// types must match exactly.
			if !dxt.sealed && !ddt.sealed &&
				dxt.PkgPath == ddt.PkgPath &&
				dxt.Name == ddt.Name { // not yet sealed
				return nil // ok
			} else if dxt.TypeID() == ddt.TypeID() {
				return nil // ok
			} else {
				return errors.New(
					"cannot use %s as %s without explicit conversion",
					dxt.String(),
					ddt.String())
			}
		} else {
			// special case if implicitly named primitive type.
			// TODO simplify with .IsNamedType().
			if _, ok := dt.(PrimitiveType); ok {
				return errors.New(
					"cannot use %s as %s without explicit conversion",
					dxt.String(),
					dt.String())
			} else {
				// carry on with baseOf(dxt)
				xt = dxt.Base // set as base to do the rest check
			}
		}
	} else if ddt, ok := dt.(*DeclaredType); ok {
		// special case if implicitly named primitive type.
		// TODO simplify with .IsNamedType().
		if _, ok := xt.(PrimitiveType); ok { // e.g. 1 == Int(1)
			if debug {
				debug.Printf("xt is primitiveType: %v, ddt: %v \n", xt, ddt)
			}
			// this is special when dt is the declared type of x
			if !isUntyped(xt) {
				return errors.New(
					"cannot use %s as %s without explicit conversion",
					xt.String(),
					ddt.String())
			} else { // xt untyped, carry on with check below
				dt = ddt.Base
			}
		} else {
			dt = ddt.Base
		}
	}

	// General cases.
	switch cdt := dt.(type) {
	case PrimitiveType: // case 1
		// if xt is untyped, ensure dt is compatible.
		switch xt {
		case UntypedBoolType:
			if dt.Kind() == BoolKind {
				return nil // ok
			} else {
				return errors.New(
					"cannot use untyped bool as %s",
					dt.Kind())
			}
		case UntypedStringType:
			if dt.Kind() == StringKind {
				return nil // ok
			} else {
				return errors.New(
					"cannot use untyped string as %s",
					dt.Kind())
			}
		// XXX, this is a loose check, we don't have the context
		// to check if it is an exact integer, e.g. 1.2 or 1.0(1.0 can be converted to int).
		// this ensure expr like (a % 1.0) pass check, while
		// expr like (a % 1.2) panic at ConvertUntypedTo, which is a delayed assertion after const evaluated.
		// assignable does not guarantee convertible.
		case UntypedBigdecType:
			switch dt.Kind() {
			case IntKind, Int8Kind, Int16Kind, Int32Kind,
				Int64Kind, UintKind, Uint8Kind, Uint16Kind,
				Uint32Kind, Uint64Kind, BigdecKind, Float32Kind, Float64Kind:
				return nil // ok
			default:
				return errors.New(
					"cannot use untyped Bigdec as %s",
					dt.Kind())
			}
		case UntypedBigintType:
			switch dt.Kind() {
			case IntKind, Int8Kind, Int16Kind, Int32Kind,
				Int64Kind, UintKind, Uint8Kind, Uint16Kind,
				Uint32Kind, Uint64Kind, BigintKind, BigdecKind, Float32Kind, Float64Kind: // see 0d0
				return nil // ok
			default:
				return errors.New(
					"cannot use untyped Bigint as %s",
					dt.Kind())
			}
		case UntypedRuneType:
			switch dt.Kind() {
			case IntKind, Int8Kind, Int16Kind, Int32Kind,
				Int64Kind, UintKind, Uint8Kind, Uint16Kind,
				Uint32Kind, Uint64Kind, BigintKind, BigdecKind, Float32Kind, Float64Kind:
				return nil // ok
			default:
				return errors.New(
					"cannot use untyped rune as %s",
					dt.Kind())
			}

		default:
			if isUntyped(xt) {
				return errors.New("unexpected untyped type")
			}
			if xt.TypeID() == cdt.TypeID() {
				return nil // ok
			}
		}
	case *PointerType: // case 4 from here on
		// Pointer types are assignable only with identical elements.
		if pt, ok := xt.(*PointerType); ok {
			if pt.Elt.TypeID() == cdt.Elt.TypeID() && tagsIdentical(preprocessGasMeterOf(store), pt.Elt, cdt.Elt) {
				return nil
			}
			return errors.New("cannot use %s as %s", xt.String(), dt.String())
		}
	case *ArrayType:
		if at, ok := xt.(*ArrayType); ok {
			if at.Len != cdt.Len {
				return errors.New(
					"cannot use %s as %s",
					at.String(),
					cdt.String())
			}
			err := checkSame(at.Elt, cdt.Elt, "")
			if err != nil {
				return errors.New(
					"cannot use %s as %s",
					at.String(),
					cdt.String())
			}
			return nil
		}
	case *SliceType:
		if st, ok := xt.(*SliceType); ok {
			if cdt.Vrd {
				return checkAssignableTo(store, n, st.Elt, cdt.Elt)
			} else {
				err := checkSame(st.Elt, cdt.Elt, "")
				if err != nil {
					return errors.New(
						"cannot use %s as %s",
						st.String(),
						cdt.String())
				}
				return nil
			}
		}
	case *MapType:
		if mt, ok := xt.(*MapType); ok {
			err := checkSame(mt.Key, cdt.Key, "")
			if err != nil {
				return errors.New(
					"cannot use %s as %s",
					mt.String(),
					cdt.String()).Stacktrace()
			}
			return nil
		}
	case *InterfaceType:
		panic("should not happen")
	case *DeclaredType:
		panic("should not happen")
	case *FuncType, *StructType, *PackageType, *TypeType:
		if xt.TypeID() == cdt.TypeID() {
			if f := embeddedFieldMismatch(xt, cdt); f != "" {
				return errors.New(
					"cannot use %s as %s (field %s is embedded in only one)",
					xt.String(), dt.String(), f)
			}
			if tagsIdentical(preprocessGasMeterOf(store), xt, cdt) {
				return nil // ok
			}
			return errors.New("cannot use %s as %s: struct tags differ", xt.String(), dt.String())
		}
	default:
		return errors.New(
			"unexpected type %s",
			dt.String())
	}
	return errors.New(
		"cannot use %s as %s",
		xt.String(),
		dt.String()).Stacktrace()
}

func (x *AssignStmt) AssertCompatible(store Store, last BlockNode) {
	if x.Op == ASSIGN || x.Op == DEFINE {
		assertValidAssignRhs(store, last, x)
		if len(x.Lhs) > len(x.Rhs) {
			if len(x.Rhs) != 1 {
				panic(fmt.Sprintf("assignment mismatch: %d variables but %d values", len(x.Lhs), len(x.Rhs)))
			}
			switch cx := x.Rhs[0].(type) {
			case *CallExpr:
				// Call case: a, b = x(...)
				ift := evalStaticTypeOf(store, last, cx.Func)
				cft := getGnoFuncTypeOf(store, ift)
				if len(x.Lhs) != len(cft.Results) {
					panic(fmt.Sprintf(
						"assignment mismatch: "+
							"%d variables but %s returns %d values",
						len(x.Lhs), cx.Func.String(), len(cft.Results)))
				}
				if x.Op == ASSIGN {
					// check assignable
					for i, lx := range x.Lhs {
						if lxt := evalAssignLhsType(store, last, lx); lxt != nil {
							mustAssignableTo(store, x, cft.Results[i].Type, lxt)
						}
					}
				}
			case *TypeAssertExpr:
				// Type-assert case: a, ok := x.(type)
				if len(x.Lhs) != 2 {
					panic("should not happen")
				}
				if x.Op == ASSIGN {
					// check first value
					if dt := evalAssignLhsType(store, last, x.Lhs[0]); dt != nil { // see composite3.gno
						ift := evalStaticTypeOf(store, last, cx)
						mustAssignableTo(store, x, ift, dt)
					}
					// check second value
					if dt := evalAssignLhsType(store, last, x.Lhs[1]); dt != nil { // see composite3.gno
						if dt.Kind() != BoolKind { // typed, not bool
							panic(fmt.Sprintf("want bool type got %v", dt))
						}
					}
				}
				cx.HasOK = true
			case *IndexExpr: // must be with map type when len(Lhs) > len(Rhs)
				if len(x.Lhs) != 2 {
					panic("should not happen")
				}
				if x.Op == ASSIGN {
					if lt := evalAssignLhsType(store, last, x.Lhs[0]); lt != nil {
						if _, ok := cx.X.(*NameExpr); ok {
							rt := baseOf(evalStaticTypeOf(store, last, cx.X))
							if mt, ok := rt.(*MapType); ok {
								mustAssignableTo(store, x, mt.Value, lt)
							}
						} else if _, ok := cx.X.(*CompositeLitExpr); ok {
							cpt := baseOf(evalStaticTypeOf(store, last, cx.X))
							if mt, ok := cpt.(*MapType); ok {
								mustAssignableTo(store, x, mt.Value, lt)
							} else {
								panic("should not happen")
							}
						}
					}

					if dt := evalAssignLhsType(store, last, x.Lhs[1]); dt != nil {
						if dt.Kind() != BoolKind { // typed, not bool
							panic(fmt.Sprintf("want bool type got %v", dt))
						}
					}
				}
				cx.HasOK = true
			default:
				panic(fmt.Sprintf("RHS should not be %v when len(Lhs) > len(Rhs)", cx))
			}
		} else { // len(Lhs) == len(Rhs)
			if x.Op == ASSIGN {
				// assert valid left value
				//
				// Only the left values are validated here. The rt-to-lt
				// assignability of each pair is checked by the
				// checkOrConvertType that preprocess runs over every Rhs[i]
				// right after this call (the ASSIGN/len(Lhs)==len(Rhs)
				// default branch), against the same lt — and this is the
				// walk-bearing check: for an interface lt it runs the whole
				// embedding-graph satisfaction walk, so doing it in both
				// places billed it (and ran it) twice per assignment.
				for _, lx := range x.Lhs {
					evalAssignLhsType(store, last, lx)
				}
			}
		}
	} else { // Ops other than assign and define
		// If this is an assignment operation, ensure there's only 1
		// expr on lhs/rhs.
		if len(x.Lhs) != 1 || len(x.Rhs) != 1 {
			panic("assignment operator " + x.Op.TokenString() +
				" requires only one expression on lhs and rhs")
		}
		// `_ += x` reads the target (it desugars to `_ = _ + x`), unlike
		// plain assignment where blank is a valid (discarded) target.
		if isBlankIdentifier(x.Lhs[0]) {
			panic("cannot use _ as value or type")
		}
		assertValidAssignRhs(store, last, x)
		lt := evalStaticTypeOf(store, last, x.Lhs[0])
		rt := evalStaticTypeOf(store, last, x.Rhs[0])

		if checker, ok := AssignStmtChecker[x.Op]; ok {
			if !checker(lt) {
				panic(fmt.Sprintf("operator %s not defined on: %v", x.Op.TokenString(), kindString(lt)))
			}
			switch x.Op {
			case ADD_ASSIGN, SUB_ASSIGN, MUL_ASSIGN, QUO_ASSIGN, REM_ASSIGN, BAND_ASSIGN, BOR_ASSIGN, BAND_NOT_ASSIGN, XOR_ASSIGN:
				// check when both typed
				if !isUntyped(lt) && !isUntyped(rt) { // in this stage, lt or rt maybe untyped, not converted yet
					if lt != nil && rt != nil {
						if lt.TypeID() != rt.TypeID() {
							panic(fmt.Sprintf("invalid operation: mismatched types %v and %v", lt, rt))
						}
					}
				}
			case SHL_ASSIGN, SHR_ASSIGN:
				if !isWhole(rt) {
					panic(fmt.Sprintf("invalid operation: invalid shift count: %v", x.Rhs[0]))
				}
				_, ric := x.Rhs[0].(*ConstExpr)
				// check negative
				if ric {
					rv := evalConst(store, last, x.Rhs[0])
					if rv.TypedValue.Sign() < 0 {
						panic(fmt.Sprintf("invalid operation: negative shift count: %v", &rv.TypedValue))
					}
				}
			default:
				// do nothing
			}
		} else {
			panic(fmt.Sprintf("checker for %s does not exist", x.Op))
		}
		// like go/types, operator/type errors take precedence over
		// target assignability.
		assertValidAssignLhs(store, last, x.Lhs[0])
	}
}

// misc
func assertValidAssignLhs(store Store, last BlockNode, lx Expr) {
	if nx, ok := lx.(*NameExpr); ok {
		switch {
		case nx.Name == blankIdentifier:
			return
		case nx.Path.Type == VPUverse:
			panic(fmt.Sprintf("cannot assign to uverse %v", nx.Name))
		case last.GetIsConst(store, nx.Name):
			panic(fmt.Sprintf("cannot assign to const %v", nx.Name))
		}
	}
	// The target is a variable or a map index expression.
	o := operandOf(store, last, lx)
	assertNotType(o)
	if o.assignable() {
		return
	}
	if sx, ok := lx.(*SelectorExpr); ok && inMapValue(store, last, sx) {
		failf(CodeMapValueAssign, nil, "cannot assign to struct field %s in map", exprString(lx))
	}
	switch lx.(type) {
	case *SelectorExpr, *IndexExpr, *NameExpr:
		failf(CodeNotAssignable, nil, "cannot assign to %s (neither addressable nor a map index expression)", exprString(lx))
	}
	panic(fmt.Sprintf("cannot assign to %s", exprString(lx)))
}

// evalAssignLhsType asserts that lx is a valid assignment target and returns
// its static type, or nil iff lx is blank: blank is a valid target but has no
// static type, so callers skip type checks on a nil result.
func evalAssignLhsType(store Store, last BlockNode, lx Expr) Type {
	assertValidAssignLhs(store, last, lx)
	if isBlankIdentifier(lx) {
		return nil
	}
	return evalStaticTypeOf(store, last, lx)
}

func assertValidAssignRhs(store Store, last BlockNode, n Node) {
	var exps []Expr
	switch x := n.(type) {
	case *ValueDecl:
		exps = x.Values
	case *AssignStmt:
		exps = x.Rhs
	default:
		panic(fmt.Sprintf("unexpected node type %T", n))
	}

	for i, exp := range exps {
		tt := evalStaticTypeOfRaw(store, last, exp)
		if tt == nil {
			switch x := n.(type) {
			case *ValueDecl:
				if x.Type != nil {
					continue
				}
				panic("use of untyped nil in variable declaration")
			case *AssignStmt:
				if x.Op != DEFINE {
					// Assigning nil to a typed target is judged later by
					// the assignability check against that target's type.
					// A blank target has no type and no later check, so
					// `_ = nil` must be rejected here.
					if i < len(x.Lhs) && isBlankIdentifier(x.Lhs[i]) && isNilIdentifier(exp) {
						panic("use of untyped nil in assignment")
					}
					continue
				}
				panic("use of untyped nil in assignment")
			}
		}
		if _, ok := tt.(*TypeType); ok {
			tt = evalStaticType(store, last, exp)
			panic(fmt.Sprintf("%s (type) is not an expression", tt.String()))
		}

		// Ensures that function used in ValueDecl or AssignStmt must return at least 1 value.
		if cx, ok := exp.(*CallExpr); ok {
			tType, ok := tt.(*tupleType)
			if ok && len(tType.Elts) == 0 {
				panic(fmt.Sprintf("%s (no value) used as value", cx.Func.String()))
			}
		}
	}
}

// embeddedFieldMismatch names the first field embedded in exactly one of
// two unnamed struct types with the same TypeID, which does not record
// embedding (Go spec, "Type identity": fields must agree on it), or "".
func embeddedFieldMismatch(at, bt Type) Name {
	ast, ok := at.(*StructType)
	if !ok {
		return ""
	}
	bst, ok := bt.(*StructType)
	if !ok || len(ast.Fields) != len(bst.Fields) {
		return ""
	}
	for i := range ast.Fields {
		if ast.Fields[i].Embedded != bst.Fields[i].Embedded {
			return ast.Fields[i].Name
		}
	}
	return ""
}
