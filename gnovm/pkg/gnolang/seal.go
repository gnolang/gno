package gnolang

// Sealing: pre-filling the lazily-memoized caches on a graph that more than one
// goroutine will read.
//
// Types and block nodes both carry caches filled on first use rather than at
// construction: TypeID (and, for interfaces, the in-place sort of Methods that
// computing it performs), FuncType.bound, Declared/StructType.pkgID, the
// interface and struct effective counts, StructType.comparable,
// DeclaredType.methodIndex, PackageNode.pkgID, StaticBlock.nameIndex. Every
// one is a check-then-set with no synchronisation, which is correct exactly as
// long as a single goroutine touches the object.
//
// Two graphs break that assumption:
//
//   - the process-global uverse singletons, shared by every store in the
//     process (sealed at package init by sealUverseTypes);
//   - the block nodes published into a defaultStore's process-wide cacheNodes
//     map, shared by every transaction and query store forked from it. Before
//     the query connection stopped serialising, only one goroutine at a time
//     ever read them; now N concurrent queries do, and the first two to reach an
//     unfilled cache race on it.
//
// A block node reaches its types two ways, and both are walked. Named types
// hang off the static block (StaticBlock.Types and Block.Values); a composite
// type written in expression position is defined under no name at all and hangs
// off the expression instead, in ATTR_TYPE_VALUE, ATTR_TYPEOF_VALUE or a
// constTypeExpr. Those are the same shared objects, reached through the same
// shared node, so sealExprTypes walks them too.
//
// Sealing closes that by filling the caches once, on the single goroutine that
// publishes the graph, so every later reader finds them already populated and
// performs no write at all. It is pure cache warming: every value written here
// is the same value the lazy path would have computed, so nothing observable
// changes — no gas is charged, no allocator is touched, and no struct layout or
// amino encoding is affected.

// sealer walks a type and block-node graph, filling every lazily-memoized cache
// exactly once. The seen sets make it cycle-safe and keep repeated publication
// of an already-sealed graph cheap.
type sealer struct {
	seenTypes map[Type]bool
	seenNodes map[BlockNode]bool
}

func newSealer() *sealer {
	return &sealer{
		seenTypes: make(map[Type]bool),
		seenNodes: make(map[BlockNode]bool),
	}
}

// sealType fills the memoized caches on t and everything reachable from it.
func (s *sealer) sealType(t Type) {
	if t == nil || s.seenTypes[t] {
		return
	}
	s.seenTypes[t] = true
	switch ct := t.(type) {
	case *PointerType:
		ct.TypeID()
		s.sealType(ct.Elt)
	case *SliceType:
		ct.TypeID()
		s.sealType(ct.Elt)
	case *ArrayType:
		ct.TypeID()
		s.sealType(ct.Elt)
	case *ChanType:
		ct.TypeID()
		s.sealType(ct.Elt)
	case *MapType:
		ct.TypeID()
		s.sealType(ct.Key)
		s.sealType(ct.Value)
	case *FuncType:
		ct.TypeID()
		if len(ct.Params) > 0 {
			// Method lookups (findEmbeddedFieldType) return the bound type and
			// TypeID it at runtime (VerifyImplementedBy), so seal the bound's
			// own typeid too, not just create it.
			ct.BoundType().TypeID()
		}
		for i := range ct.Params {
			s.sealType(ct.Params[i].Type)
		}
		for i := range ct.Results {
			s.sealType(ct.Results[i].Type)
		}
	case *StructType:
		ct.TypeID()
		ct.GetPkgID()
		isComparable(ct) // fill the comparable tristate
		effectiveStructSurface(ct, map[Type]struct{}{})
		for i := range ct.Fields {
			s.sealType(ct.Fields[i].Type)
		}
	case *InterfaceType:
		if ct.Generic != "" {
			return // generic uverse type: no TypeID, never concurrently filled
		}
		ct.TypeID()
		effectiveInterfaceMethods(ct, map[Type]struct{}{})
		for i := range ct.Methods {
			s.sealType(ct.Methods[i].Type)
		}
	case *DeclaredType:
		ct.TypeID()
		ct.GetPkgID()
		// methodIndex builds only past methodIndexThreshold. Uverse singletons
		// never reach it, but a realm type readily does, and GetMethodIndex
		// would then build the map on whichever query touched it first.
		if ct.methodIndex == nil && len(ct.Methods) > methodIndexThreshold {
			ct.buildMethodIndex()
		}
		s.sealType(ct.Base)
		for i := range ct.Methods {
			s.sealType(ct.Methods[i].T)
		}
	case *tupleType:
		// Only reachable through an expression's ATTR_TYPEOF_VALUE, which is
		// why sealExprTypes has to exist for this case to fire at all.
		// tupleType.TypeID() recurses into Elts[i].TypeID(), but that fills
		// nothing else on the elements — their pkgID, comparable, bound and
		// effective counts still need the element walk below.
		ct.TypeID()
		for _, elt := range ct.Elts {
			s.sealType(elt)
		}
	default:
		// PrimitiveType, PackageType, TypeType, etc.
		ct.TypeID()
	}
}

// sealExprTypes seals the types a node holds in its expressions rather than in
// its static block.
//
// A composite type written in expression position — sink(struct{a int}{}),
// v.(struct{a int}), map[struct{k int}]bool{} — is defined under no name, so it
// appears in no StaticBlock.Types and no Block.Values, which is everything
// sealBlockNode looks at. Preprocessing parks it on the expression instead, in
// ATTR_TYPE_VALUE, ATTR_TYPEOF_VALUE or a constTypeExpr, and the expression
// belongs to a block node the defaultStore shares with every store forked from
// it.
//
// SaveBlockNodes calls this from the Transcribe it already performs, rather
// than the store calling it at publication time like the rest of sealing.
// Riding that walk is what makes it free: a second walk at publication measured
// +10% on genesis per file and +12% per node, against +0.5% here, and the extra
// was the traversal itself over a tree the preprocessor had just traversed.
//
// Publication is also the wrong place on its own terms. Transcribe cannot
// re-walk a tree once the preprocessor is done with it — a composite literal
// whose element value ends up nil is visited unguarded and panics, which
// loadStdlibs reaches on the amino-decoded nodes in a transaction's dirty set.
// Those nodes have nothing to seal anyway: Attributes.data is unexported, so
// amino cannot persist it, and the types this walk is for exist only on nodes
// preprocessed in-process. The SaveBlockNodes walk sees exactly that set, on
// the goroutine that built it, before the store has it.
func (s *sealer) sealExprTypes(n Node) {
	// Transcribe visits a few optional children unguarded, so a nil can in
	// principle arrive here. It never does on the walk this runs on — 162,498
	// nodes over a full genesis, none nil — and Transcribe's own switch panics
	// on one two statements later regardless. The check is only so that if it
	// ever happens the failure reads as Transcribe's "node missing for X"
	// rather than a nil dereference in here.
	if n == nil {
		return
	}
	switch x := n.(type) {
	case *constTypeExpr:
		s.sealType(x.Type)
	case *ConstExpr:
		s.sealTypedValue(&x.TypedValue)
	}
	if t, ok := n.GetAttribute(ATTR_TYPE_VALUE).(Type); ok {
		s.sealType(t)
	}
	// Read raw. cachedStaticTypeOf unwraps a 1-element tupleType to its
	// Elts[0], which would hand sealType the element and leave the tuple
	// wrapper — the object the *tupleType case exists for — unsealed.
	if t, ok := n.GetAttribute(ATTR_TYPEOF_VALUE).(Type); ok {
		s.sealType(t)
	}
}

// sealBlockNode fills the caches on bn's static block and on every type it
// reaches, then follows the block's parent chain.
func (s *sealer) sealBlockNode(bn BlockNode) {
	if bn == nil || s.seenNodes[bn] {
		return
	}
	s.seenNodes[bn] = true

	// PackageNode.pkgID is the same check-then-set memo as Declared/StructType's,
	// on the node the defaultStore shares with every store forked from it.
	// Preprocessing reaches it through packageOf(last).GetPkgID() — in
	// evalStaticTypeOfRaw and evalStaticTypeMachine — which every vm/qeval runs
	// against the loaded package's block, so more than one query can be the
	// first to fill it.
	if pn, ok := bn.(*PackageNode); ok {
		pn.GetPkgID()
	}

	sb := bn.GetStaticBlock()
	if sb == nil {
		return
	}

	// nameIndex builds only past nameIndexThreshold, on the first GetLocalIndex
	// against a wide block — which for a shared node is whichever goroutine got
	// there first. A realm package block crosses 32 names easily.
	if sb.nameIndex == nil && sb.NumNames > nameIndexThreshold {
		sb.buildNameIndex()
	}

	for _, t := range sb.Types {
		s.sealType(t)
	}
	for i := range sb.Block.Values {
		s.sealTypedValue(&sb.Block.Values[i])
	}

	s.sealBlockNode(sb.Parent)
}

// sealTypedValue seals the static type of tv and, for the value kinds that
// carry a type of their own, that type too.
func (s *sealer) sealTypedValue(tv *TypedValue) {
	s.sealType(tv.T)
	switch v := tv.V.(type) {
	case TypeValue:
		s.sealType(v.Type)
	case *FuncValue:
		// GetType(nil) returns the memoized *FuncType without needing a store.
		s.sealType(v.GetType(nil))
	}
}
