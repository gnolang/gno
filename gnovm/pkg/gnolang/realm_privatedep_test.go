package gnolang

import "testing"

// newPrivateDepTestStore returns a store with two registered packages: a
// public one and a private one, identified by path.
func newPrivateDepTestStore(t *testing.T) (store Store, pubPath, privPath string) {
	t.Helper()
	pubPath, privPath = "gno.land/r/pub", "gno.land/r/priv"
	st := NewStore(nil, nil, nil)
	st.SetCachePackage(&PackageValue{PkgPath: pubPath, Private: false})
	st.SetCachePackage(&PackageValue{PkgPath: privPath, Private: true})
	return st, pubPath, privPath
}

func TestTypeHasPrivateDep_PublicStruct(t *testing.T) {
	t.Parallel()
	store, pubPath, _ := newPrivateDepTestStore(t)

	pub := &StructType{PkgPath: pubPath, Fields: []FieldType{{Name: "X", Type: IntType}}}

	if typeHasPrivateDep(store, pub) {
		t.Fatal("typeHasPrivateDep() = true, want false for a struct with no private dependency")
	}
}

func TestTypeHasPrivateDep_OwnPackagePrivate(t *testing.T) {
	t.Parallel()
	store, _, privPath := newPrivateDepTestStore(t)

	priv := &StructType{PkgPath: privPath, Fields: []FieldType{{Name: "X", Type: IntType}}}

	if !typeHasPrivateDep(store, priv) {
		t.Fatal("typeHasPrivateDep() = false, want true for a struct declared in a private package")
	}
}

func TestTypeHasPrivateDep_TransitiveViaField(t *testing.T) {
	t.Parallel()
	store, pubPath, privPath := newPrivateDepTestStore(t)

	priv := &StructType{PkgPath: privPath, Fields: []FieldType{{Name: "X", Type: IntType}}}
	container := &StructType{PkgPath: pubPath, Fields: []FieldType{{Name: "Priv", Type: priv}}}

	if !typeHasPrivateDep(store, container) {
		t.Fatal("typeHasPrivateDep() = false, want true: container has a field whose type is declared in a private package")
	}
}

// typeHasPrivateDep's whole point is to let assertTypeIsPublic skip its
// walk once a type is proven to have no private dependency anywhere.
// That's only sound if the answer is memoized, not just correct on a
// single call — mutating the store after the first call must not affect
// the second.
func TestTypeHasPrivateDep_CachesResult(t *testing.T) {
	t.Parallel()
	store, pubPath, _ := newPrivateDepTestStore(t)

	pub := &StructType{PkgPath: pubPath, Fields: []FieldType{{Name: "X", Type: IntType}}}

	if typeHasPrivateDep(store, pub) {
		t.Fatal("typeHasPrivateDep() = true, want false on first call")
	}

	pv := store.GetPackage(pubPath, false)
	pv.Private = true // simulate a stale/incorrect lookup; must not affect the cached answer

	if typeHasPrivateDep(store, pub) {
		t.Fatal("typeHasPrivateDep() after store mutation = true, want false (cached result must survive)")
	}
}

// A self-referential type (e.g. a linked-list node) must not infinite-
// loop, and — since resolving it required walking through a cycle — must
// not be permanently cached, unlike the acyclic cases above.
//
// Self-reference in real Gno always goes through a *DeclaredType (e.g.
// `type Node struct{ Next *Node }`): DeclaredType.TypeID() is a nominal
// PkgPath+Name hash, not a structural walk, which is exactly what keeps
// TypeID() itself from infinite-recursing on a cyclic type. A raw
// anonymous *StructType can't self-reference this way (there's no name
// to close the loop through), so the test builds the cycle the same way
// the preprocessor would.
func TestTypeHasPrivateDep_CyclicTypeNotCached(t *testing.T) {
	t.Parallel()
	store, pubPath, _ := newPrivateDepTestStore(t)

	nodeDT := &DeclaredType{PkgPath: pubPath, Name: "Node"}
	nodeDT.Base = &StructType{PkgPath: pubPath, Fields: []FieldType{
		{Name: "Next", Type: &PointerType{Elt: nodeDT}},
	}}

	if typeHasPrivateDep(store, nodeDT) {
		t.Fatal("typeHasPrivateDep() = true, want false for a self-referential type with no private dependency")
	}
	if nodeDT.privateDep != 0 {
		t.Fatalf("nodeDT.privateDep = %d, want 0 (uncached): a node only reachable through a cycle must not be memoized", nodeDT.privateDep)
	}
}

// Regression test for a fixed-point hazard: with two mutually-referencing
// types A <-> B where A also has an unrelated private-dependency field,
// naively caching each node's result as soon as its own local DFS
// finishes is WRONG. Exploring A visits B first (via the mutual
// reference); B's only other field is the cyclic back-edge to A, so if B
// were cached right there, it would be cached as "no private dep" before
// A's remaining field (the private one) is ever examined. A's true
// answer is discovered only afterward, and B is transitively just as
// tainted (B holds a field of type A, and A has a private dependency) —
// so a naive walk-order-dependent cache would freeze B at the wrong
// answer. Checking A first must not corrupt B's answer when B is queried
// independently afterward.
func TestTypeHasPrivateDep_MutualCycleDoesNotPoisonPeerCache(t *testing.T) {
	t.Parallel()
	store, pubPath, privPath := newPrivateDepTestStore(t)

	priv := &StructType{PkgPath: privPath, Fields: []FieldType{{Name: "X", Type: IntType}}}

	aDT := &DeclaredType{PkgPath: pubPath, Name: "A"}
	bDT := &DeclaredType{PkgPath: pubPath, Name: "B"}
	// aDT's field order matters: the mutual reference to bDT comes before
	// the private field, so exploring aDT reaches bDT (and, through it,
	// loops back to aDT) before ever seeing aDT's own private field.
	aDT.Base = &StructType{PkgPath: pubPath, Fields: []FieldType{
		{Name: "B", Type: &PointerType{Elt: bDT}},
		{Name: "Priv", Type: priv},
	}}
	bDT.Base = &StructType{PkgPath: pubPath, Fields: []FieldType{
		{Name: "A", Type: &PointerType{Elt: aDT}},
	}}

	if !typeHasPrivateDep(store, aDT) {
		t.Fatal("typeHasPrivateDep(aDT) = false, want true: aDT has a private field")
	}
	if !typeHasPrivateDep(store, bDT) {
		t.Fatal("typeHasPrivateDep(bDT) = false, want true: bDT transitively references aDT's private field through the mutual cycle")
	}
}

// assertTypeIsPublic is the actual security-relevant entry point;
// typeHasPrivateDep is only a fast-path short-circuit in front of it. These
// tests exercise assertTypeIsPublic itself to confirm the short-circuit
// didn't change its observable panic/no-panic behavior.

func TestAssertTypeIsPublic_NoPrivateDep_DoesNotPanic(t *testing.T) {
	t.Parallel()
	store, pubPath, _ := newPrivateDepTestStore(t)
	rlm := NewRealm(pubPath)

	pub := &StructType{PkgPath: pubPath, Fields: []FieldType{{Name: "X", Type: IntType}}}

	rlm.assertTypeIsPublic(store, pub, map[TypeID]struct{}{})
}

func TestAssertTypeIsPublic_PrivateDepFromOtherRealm_Panics(t *testing.T) {
	t.Parallel()
	store, pubPath, privPath := newPrivateDepTestStore(t)
	rlm := NewRealm(pubPath)

	priv := &StructType{PkgPath: privPath, Fields: []FieldType{{Name: "X", Type: IntType}}}
	container := &StructType{PkgPath: pubPath, Fields: []FieldType{{Name: "Priv", Type: priv}}}

	defer func() {
		if recover() == nil {
			t.Fatal("assertTypeIsPublic did not panic for a type with a private dependency from another realm")
		}
	}()
	rlm.assertTypeIsPublic(store, container, map[TypeID]struct{}{})
}

// A realm may always use its own types, even if it's private — the
// exemption is realm-dependent (pkgPath == rlm.Path), which is exactly
// why typeHasPrivateDep can't fold it into its permanent, realm-
// independent cache and must instead defer to the real walk whenever it
// finds any private package anywhere.
func TestAssertTypeIsPublic_OwnPrivateType_DoesNotPanic(t *testing.T) {
	t.Parallel()
	store, _, privPath := newPrivateDepTestStore(t)
	rlm := NewRealm(privPath)

	ownType := &StructType{PkgPath: privPath, Fields: []FieldType{{Name: "X", Type: IntType}}}

	rlm.assertTypeIsPublic(store, ownType, map[TypeID]struct{}{})
}
