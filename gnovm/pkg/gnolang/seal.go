package gnolang

// Sealing: pre-filling the lazily-memoized caches on a graph that more than one
// goroutine will read.
//
// Types and block nodes both carry caches filled on first use rather than at
// construction: TypeID (and, for interfaces, the in-place sort of Methods that
// computing it performs), FuncType.bound, Declared/StructType.pkgID, the
// interface and struct effective counts, StructType.comparable,
// DeclaredType.methodIndex, StaticBlock.nameIndex. Every one is a check-then-set
// with no synchronisation, which is correct exactly as long as a single
// goroutine touches the object.
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
		if len(ct.Methods) > methodIndexThreshold {
			ct.buildMethodIndex()
		}
		s.sealType(ct.Base)
		for i := range ct.Methods {
			s.sealType(ct.Methods[i].T)
		}
	default:
		// PrimitiveType, PackageType, TypeType, etc.
		ct.TypeID()
	}
}

// sealBlockNode fills the caches on bn's static block and on every type it
// reaches, then follows the block's parent chain.
func (s *sealer) sealBlockNode(bn BlockNode) {
	if bn == nil || s.seenNodes[bn] {
		return
	}
	s.seenNodes[bn] = true

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
