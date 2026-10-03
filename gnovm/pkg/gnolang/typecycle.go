package gnolang

import (
	"fmt"
	"slices"

	"github.com/gnolang/gno/tm2/pkg/store"
)

// Validation of cycles among type declarations. It runs before the
// declarations are predefined, so no invalid type is ever constructed.
//
// A cycle is invalid when every edge on it is a direct containment (a struct
// field, an array element, an embedded interface or a bare name), which gives
// the type infinite size, or when every declaration on it is an alias, which
// leaves the aliases nothing to resolve to. Any other cycle is legal: it
// passes through an indirection and contains a defined type. See
// golang/go#25838.
//
// The walks here run outside the op loop and are billed to the preprocess
// gas meter per unit of work (chargeCPUGas), like embedWalk.

// typeDeclSite is a type declaration with the block node that holds it, for
// error locations.
type typeDeclSite struct {
	decl  *TypeDecl
	block BlockNode
}

type typeDep struct {
	name   Name
	direct bool
}

type typeDeclGraph struct {
	sites        map[Name]typeDeclSite
	deps         map[Name][]typeDep
	names        []Name // declaration order, for deterministic reports
	hasAliasEdge bool   // some edge joins two aliases
	nodes, edges int64  // work done, for gas
}

// newTypeDeclGraph indexes the declarations in sites and the references
// between them. A name that sites do not declare (another package, uverse,
// an enclosing scope) is a leaf. Like predefineDeclIndex, a redeclared name
// resolves to its last declaration.
func newTypeDeclGraph(sites []typeDeclSite) *typeDeclGraph {
	g := &typeDeclGraph{
		sites: make(map[Name]typeDeclSite, len(sites)),
		deps:  make(map[Name][]typeDep),
	}
	for _, s := range sites {
		name := s.decl.Name
		if name == blankIdentifier {
			continue
		}
		if _, ok := g.sites[name]; !ok {
			g.names = append(g.names, name)
		}
		g.sites[name] = s
	}
	for _, name := range g.names {
		from := g.sites[name].decl
		g.nodes += collectTypeDeps(from.Type, true, func(dep Name, direct bool) {
			to, ok := g.sites[dep]
			if !ok {
				return
			}
			g.deps[name] = append(g.deps[name], typeDep{dep, direct})
			g.hasAliasEdge = g.hasAliasEdge || (from.IsAlias && to.decl.IsAlias)
		})
	}
	return g
}

// collectTypeDeps calls add for every name that x refers to as a type and
// returns the number of nodes visited. direct is false once the reference
// sits behind an indirection.
func collectTypeDeps(x Expr, direct bool, add func(Name, bool)) (nodes int64) {
	nodes = 1
	switch x := x.(type) {
	case *NameExpr:
		add(x.Name, direct)
	case *StarExpr:
		nodes += collectTypeDeps(x.X, false, add)
	case *SliceTypeExpr:
		nodes += collectTypeDeps(x.Elt, false, add)
	case *ArrayTypeExpr:
		nodes += collectTypeDeps(x.Elt, direct, add)
	case *MapTypeExpr:
		nodes += collectTypeDeps(x.Key, false, add)
		nodes += collectTypeDeps(x.Value, false, add)
	case *ChanTypeExpr:
		nodes += collectTypeDeps(x.Value, false, add)
	case *FuncTypeExpr:
		for i := range x.Params {
			nodes += collectTypeDeps(x.Params[i].Type, false, add)
		}
		for i := range x.Results {
			nodes += collectTypeDeps(x.Results[i].Type, false, add)
		}
	case *InterfaceTypeExpr:
		// An embedded interface is a bare name and direct; a method
		// signature is a FuncTypeExpr and so indirect.
		for i := range x.Methods {
			nodes += collectTypeDeps(x.Methods[i].Type, direct, add)
		}
	case *StructTypeExpr:
		for i := range x.Fields {
			nodes += collectTypeDeps(x.Fields[i].Type, direct, add)
		}
	case *FieldTypeExpr:
		nodes += collectTypeDeps(x.Type, direct, add)
	case *SelectorExpr:
		// pkg.T names another package's type, unless pkg is a member of
		// this group: `type time time.Duration` with no import.
		if nx, ok := x.X.(*NameExpr); ok {
			add(nx.Name, direct)
		}
	}
	// A *constTypeExpr is already a type and cannot refer back into
	// this group.
	return nodes
}

// findCycle returns a cycle among the edges that keep accepts, or nil. The
// cycle is rotated so its earliest-declared name leads, and that name is
// repeated at the end, so the report is a property of the cycle alone.
func (g *typeDeclGraph) findCycle(keep func(from Name, dep typeDep) bool) []Name {
	const (
		white = iota
		gray
		black
	)
	color := make(map[Name]int, len(g.names))
	var path []Name
	var visit func(Name) []Name
	visit = func(n Name) []Name {
		color[n] = gray
		path = append(path, n)
		for _, dep := range g.deps[n] {
			g.edges++
			if !keep(n, dep) {
				continue
			}
			switch color[dep.name] {
			case gray:
				return slices.Clone(path[slices.Index(path, dep.name):])
			case white:
				if c := visit(dep.name); c != nil {
					return c
				}
			}
		}
		path = path[:len(path)-1]
		color[n] = black
		return nil
	}
	for _, n := range g.names {
		if color[n] != white {
			continue
		}
		cycle := visit(n)
		if cycle == nil {
			continue
		}
		first := 0
		for i := range cycle {
			if slices.Index(g.names, cycle[i]) < slices.Index(g.names, cycle[first]) {
				first = i
			}
		}
		cycle = append(cycle[first:], cycle[:first]...)
		return append(cycle, cycle[0])
	}
	return nil
}

// invalidCycle returns an invalid cycle among the declarations, or nil.
func (g *typeDeclGraph) invalidCycle() []Name {
	if len(g.deps) == 0 {
		return nil
	}
	// Every edge direct: the type has infinite size.
	if c := g.findCycle(func(_ Name, dep typeDep) bool { return dep.direct }); c != nil {
		return c
	}
	// Every declaration an alias: nothing to resolve to.
	if !g.hasAliasEdge {
		return nil
	}
	return g.findCycle(func(from Name, dep typeDep) bool {
		return g.sites[from].decl.IsAlias && g.sites[dep.name].decl.IsAlias
	})
}

// beginTypeDeclGroup runs before any declaration of a group is
// predefined: it rejects invalid cycles, then reserves every slot.
// endTypeDeclGroup closes the group.
func beginTypeDeclGroup(store Store, sites []typeDeclSite) {
	assertNoTypeDeclCycles(preprocessGasMeterOf(store), sites)
	reserveTypeDecls(store, sites)
}

// assertNoTypeDeclCycles panics, located at the first declaration of the
// cycle, if the type declarations in sites form an invalid cycle. The
// scan is billed to gm per node and edge.
func assertNoTypeDeclCycles(gm store.GasMeter, sites []typeDeclSite) {
	if len(sites) == 0 {
		return
	}
	g := newTypeDeclGraph(sites)
	cycle := g.invalidCycle()
	chargeCPUGas(gm, OpCPUSlopeTypeDeclSite*int64(len(g.names))+
		OpCPUSlopeTypeDepNode*g.nodes+OpCPUSlopeTypeDepEdge*g.edges)
	if cycle == nil {
		return
	}
	site := g.sites[cycle[0]]
	func() {
		defer doRecover([]BlockNode{site.block}, site.decl)
		panic(fmt.Sprintf("invalid recursive type: %s", Names(cycle).Join(" -> ")))
	}()
}

// appendTypeDeclSites appends the type declarations among decls that are
// not predefined yet, all held by block, to sites.
func appendTypeDeclSites(sites []typeDeclSite, block BlockNode, decls []Decl) []typeDeclSite {
	for _, d := range decls {
		if td, ok := d.(*TypeDecl); ok && td.GetAttribute(ATTR_PREDEFINED) != true {
			sites = append(sites, typeDeclSite{td, block})
		}
	}
	return sites
}

// endTypeDeclGroup runs the checks that need every type of the group
// settled: embed depth and map-key comparability. Seal meets a
// pointer-referenced member before its base is set, so it can judge neither.
func endTypeDeclGroup(store Store, sites []typeDeclSite) {
	gm := preprocessGasMeterOf(store)
	for _, s := range sites {
		if s.decl.Name != blankIdentifier {
			checkBuiltTypeDecl(store, gm, s)
		}
	}
}

func checkBuiltTypeDecl(store Store, gm store.GasMeter, s typeDeclSite) {
	tv := s.block.GetSlot(store, s.decl.Name, true)
	if tv == nil {
		return
	}
	// A bare name adds no embed or map of its own; the declaration it
	// names checks them.
	switch unconst(s.decl.Type).(type) {
	case *NameExpr, *SelectorExpr:
		return
	}
	defer doRecover([]BlockNode{s.block}, s.decl)
	t := tv.GetType()
	validateEmbedDepth(gm, t, string(s.decl.Name))
	var work int64
	key := uncomparableMapKey(baseOf(t), &work)
	chargeCPUGas(gm, OpCPUSlopeMapKeyNode*work)
	if key != nil {
		panic(fmt.Sprintf("invalid map key type %s", key.String()))
	}
}

// uncomparableMapKey returns the key type of the first map written inline
// in t whose key is not comparable, or nil, counting visited nodes in
// work. A declared type inside t is not entered: it is checked at its own
// declaration.
func uncomparableMapKey(t Type, work *int64) Type {
	*work++
	switch t := t.(type) {
	case *ArrayType:
		return uncomparableMapKey(t.Elt, work)
	case *SliceType:
		return uncomparableMapKey(t.Elt, work)
	case *PointerType:
		return uncomparableMapKey(t.Elt, work)
	case *StructType:
		return uncomparableMapKeyIn(work, t.Fields)
	case *FuncType:
		return uncomparableMapKeyIn(work, t.Params, t.Results)
	case *InterfaceType:
		return uncomparableMapKeyIn(work, t.Methods)
	case *MapType:
		if !isComparable(t.Key) {
			return t.Key
		}
		if key := uncomparableMapKey(t.Key, work); key != nil {
			return key
		}
		return uncomparableMapKey(t.Value, work)
	}
	return nil
}

func uncomparableMapKeyIn(work *int64, lists ...[]FieldType) Type {
	for _, fs := range lists {
		for i := range fs {
			if key := uncomparableMapKey(fs[i].Type, work); key != nil {
				return key
			}
		}
	}
	return nil
}

// reserveTypeDecls gives every declaration in sites its slot before any of
// them is predefined: a shell of the declared kind, or for an alias of a
// name in the group, the slot of the declaration the alias chain ends at.
// A legal cycle can then close on any member, whichever declaration is
// predefined first. An alias of a name outside the group is left to
// tryPredefine, which resolves it once that name is defined.
func reserveTypeDecls(store Store, sites []typeDeclSite) {
	byName := make(map[Name]typeDeclSite, len(sites))
	for _, s := range sites {
		byName[s.decl.Name] = s
	}
	var reserve func(s typeDeclSite) (t Type, ok bool)
	reserve = func(s typeDeclSite) (t Type, ok bool) {
		d, last := s.decl, s.block
		last2 := skipFile(last)
		if isLocallyDefined(last2, d.Name) {
			return last2.GetSlot(store, d.Name, true).GetType(), true
		}
		switch tx := d.Type.(type) {
		case *FuncTypeExpr:
			t = &FuncType{}
		case *ArrayTypeExpr:
			t = &ArrayType{}
		case *SliceTypeExpr:
			t = &SliceType{}
		case *InterfaceTypeExpr:
			t = &InterfaceType{}
		case *MapTypeExpr:
			t = &MapType{}
		case *StructTypeExpr:
			t = &StructType{}
		case *StarExpr:
			t = &PointerType{}
		case *NameExpr:
			if isBlankIdentifier(tx) || tx.Name == "nil" {
				return nil, false // tryPredefine reports it
			}
			if d.IsAlias {
				target, inGroup := byName[tx.Name]
				if !inGroup {
					return nil, false
				}
				// The alias-only cycle check has run, so this ends.
				if t, ok = reserve(target); !ok {
					return nil, false
				}
			}
		default:
			return nil, false
		}
		if !d.IsAlias {
			t = declareWith(packageOf(last).PkgPath, last, d.Name, t)
		}
		last2.Define2(true, d.Name, t, asValue(t), NameSource{&d.NameExpr, d, NSTypeDecl, -1})
		d.Path = last.GetPathForName(store, d.Name)
		return t, true
	}
	for _, s := range sites {
		if s.decl.Name != blankIdentifier {
			reserve(s)
		}
	}
}
