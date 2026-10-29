package gnolang

// The operand model (after go/types): one answer to "what is this
// expression" for the rule files. An operand is built at the rule's call
// site from what the preprocessor already cached, so it allocates
// nothing and stores nothing on the tree.

// opMode says what kind of thing an expression denotes.
type opMode uint8

const (
	opInvalid  opMode = iota // dropped or unknown: rules say nothing
	opNoValue                // a call with no result
	opBuiltin                // a built-in function (uverse)
	opType                   // a type expression
	opConstant               // a constant, val holds it
	opNil                    // the untyped nil
	opVariable               // addressable
	opMapIndex               // m[k]: assignable, not addressable
	opValue                  // any other value
	opTuple                  // a call with several results
	opPackage                // an imported package name
)

var opModeNames = [...]string{
	"invalid", "no value", "built-in", "type", "constant", "nil",
	"variable", "map index expression", "value", "tuple", "package",
}

func (m opMode) String() string { return opModeNames[m] }

// operand is an expression as a rule sees it: its mode, its static type
// (nil for untyped nil and for no value), the expression and, for a
// constant, its value.
type operand struct {
	mode opMode
	typ  Type
	expr Expr
	val  *TypedValue
}

// assignable reports whether the operand can be the target of `=`.
func (o operand) assignable() bool {
	return o.mode == opVariable || o.mode == opMapIndex
}

// isType reports whether the operand denotes a type, not a value.
func (o operand) isType() bool { return o.mode == opType }

// operandOf classifies the preprocessed expression x in block last.
func operandOf(store Store, last BlockNode, x Expr) operand {
	o := operand{expr: x}
	switch cx := x.(type) {
	case nil:
		return o
	case *constTypeExpr:
		o.mode, o.typ = opType, cx.Type
		return o
	case *ConstExpr:
		// The static type may live on the node: a nil interface value
		// has T == nil in its TypedValue.
		t := evalStaticTypeOfRaw(store, last, cx)
		switch {
		case isNilLiteral(cx):
			// The nil literal, also once converted to its context type;
			// a typed nil such as (func())(nil) is a value.
			o.mode, o.typ = opNil, t
		case cx.V == nil && t != nil && mayBeNil(t):
			o.mode, o.typ = opValue, t
		case isTypeValue(cx.T):
			o.mode, o.typ = opType, cx.GetType()
		case isPackageValue(cx):
			o.mode, o.typ = opPackage, cx.T
		case isBuiltinFunc(cx):
			o.mode, o.typ, o.val = opBuiltin, cx.T, &cx.TypedValue
		default:
			o.mode, o.typ, o.val = opConstant, cx.T, &cx.TypedValue
		}
		return o
	}
	t := evalStaticTypeOfRaw(store, last, x)
	switch tt := t.(type) {
	case nil:
		o.mode = opNil
		return o
	case *tupleType:
		switch len(tt.Elts) {
		case 0:
			o.mode = opNoValue
		case 1:
			o.mode, o.typ = opValue, tt.Elts[0]
		default:
			o.mode, o.typ = opTuple, tt
		}
		return o
	case *TypeType:
		o.mode, o.typ = opType, evalStaticType(store, last, x)
		return o
	case *PackageType:
		o.mode, o.typ = opPackage, t
		return o
	}
	o.typ = t
	o.mode = valueMode(store, last, x, t)
	return o
}

// valueMode tells a variable (addressable) from a map index expression
// and from any other value (Go spec, "Address operators"). A name whose
// path does not resolve is invalid, so no rule speaks for it.
func valueMode(store Store, last BlockNode, x Expr, t Type) opMode {
	switch cx := x.(type) {
	case *NameExpr:
		src := nameSourceOf(last, cx)
		if src == nil {
			return opInvalid
		}
		switch src.Type {
		case NSFuncDecl, NSTypeDecl, NSImportDecl:
			return opValue
		}
		return opVariable
	case *StarExpr:
		return opVariable
	case *SelectorExpr:
		switch cx.Path.Type {
		case VPValMethod, VPPtrMethod, VPInterface,
			VPDerefValMethod, VPDerefPtrMethod, VPDerefInterface:
			return opValue // a method value
		}
		xt := evalStaticTypeOf(store, last, cx.X)
		switch baseOf(xt).(type) {
		case *PointerType:
			return opVariable
		case *PackageType:
			// pkg.Var is addressable; pkg.Func is not.
			if src := packageNameSourceOf(store, last, cx); src != nil && src.Type == NSFuncDecl {
				return opValue
			}
			return opVariable
		}
		if _, ok := baseOf(xt).(*StructType); ok && valueMode(store, last, cx.X, xt) == opVariable {
			return opVariable
		}
		return opValue
	case *IndexExpr:
		xt := evalStaticTypeOf(store, last, cx.X)
		switch baseOf(xt).(type) {
		case *SliceType:
			return opVariable
		case *MapType:
			return opMapIndex
		case *ArrayType:
			if valueMode(store, last, cx.X, xt) == opVariable {
				return opVariable
			}
		case *PointerType:
			return opVariable
		}
		return opValue
	case *CompositeLitExpr:
		return opValue
	}
	return opValue
}

// nameSourceOf returns how the name nx was declared, or nil when its
// path does not lead to a block in last's ancestry (uverse, a dropped
// declaration, a path deeper than the chain).
func nameSourceOf(last BlockNode, nx *NameExpr) *NameSource {
	if nx.Path.Type != VPBlock || nx.Path.Depth == 0 {
		return nil
	}
	// GetBlockNodeForPath walks Depth-1 parents (one more past a faux
	// child); check the chain is long enough, it does not.
	bn := last
	for i := 1; i < int(nx.Path.Depth); i++ {
		if fauxChildBlockNode(bn) {
			if bn = bn.GetParentNode(nil); bn == nil {
				return nil
			}
		}
		if bn = bn.GetParentNode(nil); bn == nil {
			return nil
		}
	}
	bn = last.GetBlockNodeForPath(nil, nx.Path)
	return nameSourceIn(bn, nx.Name)
}

// packageNameSourceOf returns how the member sx.Sel of the package sx.X
// names was declared, or nil when it cannot be resolved.
func packageNameSourceOf(store Store, last BlockNode, sx *SelectorExpr) *NameSource {
	// The package name folds to a constant holding a package ref; the
	// import's slot in the file block holds the package value.
	x := sx.X
	if cx, ok := x.(*ConstExpr); ok {
		x = cx.Source
	}
	nx, ok := x.(*NameExpr)
	if !ok {
		return nil
	}
	tv := last.GetSlot(store, nx.Name, true)
	if tv == nil {
		return nil
	}
	pv, ok := tv.V.(*PackageValue)
	if !ok {
		return nil
	}
	return nameSourceIn(pv.GetBlock(store).GetSource(store), sx.Sel)
}

// nameSourceIn returns the source of the name declared in bn, or nil.
func nameSourceIn(bn BlockNode, name Name) *NameSource {
	if bn == nil {
		return nil
	}
	idx, ok := bn.GetLocalIndex(name)
	sources := bn.GetStaticBlock().NameSources
	if !ok || int(idx) >= len(sources) {
		return nil
	}
	return &sources[idx]
}

func isTypeValue(t Type) bool {
	_, ok := t.(*TypeType)
	return ok
}

func isPackageValue(cx *ConstExpr) bool {
	_, ok := cx.T.(*PackageType)
	return ok
}

func isBuiltinFunc(cx *ConstExpr) bool {
	fv, ok := cx.V.(*FuncValue)
	return ok && fv.PkgPath == uversePkgPath
}
