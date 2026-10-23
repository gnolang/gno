package gnolang

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

// upgradeShape is what a versioned redeploy compares between two versions of
// a package: every global with its static type, every declared type with its
// underlying type and its methods in order, and the referenceable layout.
// Built from a PackageNode alone, so the chain and `gno upgrade check` share
// it.
type upgradeShape struct {
	names   map[Name]int     // every package-level name -> block slot
	vars    map[Name]slotSig // the globals among them
	types   map[Name]TypeID  // declared type -> TypeID of its underlying type
	methods map[Name][]slotSig
	// Importers compile package selectors to block slots and method
	// selectors to method indices, so a public realm may only append: the
	// layout is every referenceable name with its static type, in slot order.
	layout []slotSig
}

type slotSig struct {
	name  Name
	tid   TypeID
	index int
}

// upgradePlan is what a versioned redeploy carries over from the package live
// at the path: the prior shape, and the heap item persisting each global.
// Built while the prior package is still what the store serves; applied once
// the new one has replaced it.
type upgradePlan struct {
	shape  *upgradeShape
	vars   map[Name]ObjectID // the *HeapItemValue holding each global
	prior  *Block            // the prior package block, which still names the heap items
	public bool
}

// IsPkgInitFunc reports whether name is a package initializer: init, or
// migrate, its counterpart on a versioned redeploy. Both may be declared more
// than once, cannot be referenced, and run once at deploy.
func IsPkgInitFunc(name Name) bool {
	return name == "init" || name == "migrate"
}

// isUnreferenceableName reports a name no importer can compile against: a
// suffixed initializer (init.N, migrate.N) or a blank func (._N). They are
// reserved after every other name (reserveHiddenFuncs), so the layout an
// importer compiled against excludes them.
func isUnreferenceableName(n Name) bool {
	base, rest, dotted := strings.Cut(string(n), ".")
	return dotted && (IsPkgInitFunc(Name(base)) || (base == "" && strings.HasPrefix(rest, "_")))
}

// shapeOf reads a package's shape off its static block.
func shapeOf(pn *PackageNode) *upgradeShape {
	sb := pn.GetStaticBlock()
	s := &upgradeShape{
		names:   map[Name]int{},
		vars:    map[Name]slotSig{},
		types:   map[Name]TypeID{},
		methods: map[Name][]slotSig{},
	}
	heap := pn.GetHeapItems()
	for i, name := range pn.GetBlockNames() {
		sig := slotSig{name: name, tid: sb.Types[i].TypeID(), index: i}
		s.names[name] = i
		if heap[i] {
			s.vars[name] = sig
		}
		if !isUnreferenceableName(name) {
			s.layout = append(s.layout, sig)
		}
	}
	for _, dt := range ownDeclaredTypes(sb.Values, pn.PkgPath) {
		s.types[dt.Name] = dt.Base.TypeID()
		s.methods[dt.Name] = methodSigs(dt)
	}
	return s
}

func methodSigs(dt *DeclaredType) []slotSig {
	sigs := make([]slotSig, 0, len(dt.Methods))
	for i, m := range dt.Methods {
		sigs = append(sigs, slotSig{name: m.V.(*FuncValue).Name, tid: m.T.TypeID(), index: i})
	}
	return sigs
}

// ownDeclaredTypes returns the types among values declared by pkgPath.
// Aliases to uverse or other packages' types are left to their owner.
func ownDeclaredTypes(values []TypedValue, pkgPath string) []*DeclaredType {
	var dts []*DeclaredType
	for _, tv := range values {
		if tvv, ok := tv.V.(TypeValue); ok {
			if dt, ok := tvv.Type.(*DeclaredType); ok && dt.PkgPath == pkgPath {
				dts = append(dts, dt)
			}
		}
	}
	return dts
}

// planUpgrade reads the package live at pkgPath.
func planUpgrade(store Store, pkgPath string) *upgradePlan {
	pv := store.GetPackage(pkgPath, false)
	if pv == nil {
		panic(fmt.Sprintf("upgrade: no package at %s", pkgPath))
	}
	pn := store.GetPackageNode(pkgPath)
	pb := pv.GetBlock(store)
	plan := &upgradePlan{
		shape:  shapeOf(pn),
		vars:   map[Name]ObjectID{},
		prior:  pb,
		public: !pv.Private,
	}
	for name, sig := range plan.shape.vars {
		plan.vars[name] = pb.Values[sig.index].V.(ObjectIDer).GetObjectID()
	}
	return plan
}

// CheckUpgrade reports what a versioned redeploy of next over prior would
// refuse, with the chain's message; nil when the carry would take it. public
// applies the append-only layout rule of a realm others may import.
func CheckUpgrade(prior, next *PackageNode, public bool) error {
	ps := shapeOf(prior)
	if err := compareUpgrade(ps, shapeOf(next), public); err != nil {
		return err
	}
	for _, fn := range next.FileSet.Files {
		for _, d := range fn.Decls {
			if _, err := carriedDecl(d, ps.vars); err != nil {
				return err
			}
		}
	}
	return nil
}

// compareUpgrade refuses any change from prior to next the carry cannot take.
func compareUpgrade(prior, next *upgradeShape, public bool) error {
	if public {
		if err := assertPrefix(prior.layout, next.layout, "declaration"); err != nil {
			return err
		}
		for _, name := range slices.Sorted(maps.Keys(next.types)) {
			if err := assertPrefix(prior.methods[name], next.methods[name], "method"); err != nil {
				return err
			}
		}
	}
	for _, name := range slices.Sorted(maps.Keys(next.types)) {
		if old, ok := prior.types[name]; ok && old != next.types[name] {
			return fmt.Errorf("upgrade: type %s changed its underlying type; "+
				"a versioned redeploy may change methods, not a type's shape", name)
		}
	}
	for _, name := range slices.Sorted(maps.Keys(prior.types)) {
		if _, ok := next.types[name]; !ok {
			return fmt.Errorf("upgrade: type %s was removed; persisted values may still have it", name)
		}
	}
	for _, name := range slices.Sorted(maps.Keys(prior.vars)) {
		nv, isVar := next.vars[name]
		if _, ok := next.names[name]; ok && !isVar {
			return fmt.Errorf("upgrade: %s is no longer a variable", name)
		}
		if isVar && nv.tid != prior.vars[name].tid {
			return fmt.Errorf("upgrade: variable %s changed type from %s to %s; "+
				"declare the new one under another name and migrate", name, prior.vars[name].tid, nv.tid)
		}
		// Absent: dropped, which a private realm may do; assertPrefix refused
		// it for a public one above.
	}
	return nil
}

// assertPrefix refuses a redeploy whose have does not start with prior: what
// names the layout in the message.
func assertPrefix(prior, have []slotSig, what string) error {
	for i, ps := range prior {
		if i >= len(have) || have[i].name != ps.name {
			return fmt.Errorf("upgrade: %s %s moved or was removed; importers compiled "+
				"against the %s order, so a public realm may only append", what, ps.name, what)
		}
		if have[i].tid != ps.tid {
			return fmt.Errorf("upgrade: %s %s changed type from %s to %s; importers compiled against it",
				what, ps.name, ps.tid, have[i].tid)
		}
	}
	return nil
}

// applyUpgradePlan points the new package block at the heap items the prior
// version persisted, refusing any change the plan cannot carry. It runs after
// PrepareNewValues and before any declaration executes.
//
// The prior blocks are not deleted: their reference counts include every
// function and file block parented under them, which is the cleanup #4949
// is about. The prior block's slots for the carried heap items are cleared,
// so that cleanup cannot later count the live items as its own.
func (m *Machine) applyUpgradePlan(pn *PackageNode, pv *PackageValue, plan *upgradePlan) {
	store := m.Store
	pb := pv.GetBlock(store)
	next := shapeOf(pn)
	if err := compareUpgrade(plan.shape, next, plan.public); err != nil {
		panic(err.Error())
	}
	// Types first: the heap items loaded below resolve their TypeIDs through
	// the store and must see this version's definitions, methods included.
	m.saveDeclaredTypes(pv, true)

	rlm := pv.GetRealm()
	for _, name := range slices.Sorted(maps.Keys(plan.vars)) {
		hiv := store.GetObject(plan.vars[name]).(*HeapItemValue)
		// Detach from the prior block; finalize re-adopts a carried item
		// under the new one, and releases a dropped one with what it owns.
		hiv.DecRefCount()
		plan.prior.Values[plan.shape.vars[name].index] = TypedValue{}
		if nv, ok := next.vars[name]; ok {
			hiv.SetOwner(nil)
			pb.Values[nv.index] = TypedValue{T: heapItemType{}, V: hiv}
		} else {
			rlm.MarkNewDeleted(hiv)
		}
	}
	rlm.MarkDirty(plan.prior)
}

// carriedDecl reports whether decl declares carried globals, refusing a
// declaration that mixes carried and new names.
func carriedDecl(decl Decl, vars map[Name]slotSig) (bool, error) {
	vd, ok := decl.(*ValueDecl)
	if !ok {
		return false, nil
	}
	n := 0
	for _, nx := range vd.NameExprs {
		if _, ok := vars[nx.Name]; ok {
			n++
		}
	}
	switch n {
	case 0:
		return false, nil
	case len(vd.NameExprs):
		return true, nil
	default:
		return false, fmt.Errorf("upgrade: declaration %v mixes carried and new globals", vd.GetDeclNames())
	}
}
