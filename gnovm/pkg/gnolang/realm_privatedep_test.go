package gnolang

import (
	"io"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	storetypes "github.com/gnolang/gno/tm2/pkg/store/types"
)

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
// single call — mutating package privacy after the first call must not
// change the cached verdict.
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

// The store-level, TypeID-keyed cache must survive a transaction boundary
// — this is the whole point of the redesign. Each transaction reloads
// types as fresh objects, so a verdict computed under one transaction
// must be reachable by a later transaction querying the same TypeID.
//
// Regression test for the original design, where the memo lived on the
// per-tx Type object and never survived the commit it was supposed to
// span.
func TestTypeHasPrivateDep_CacheSurvivesTransaction(t *testing.T) {
	t.Parallel()
	db := memdb.NewMemDB()
	tm2Store := dbadapter.StoreConstructor(db, storetypes.StoreOptions{})
	st := NewStore(nil, tm2Store, tm2Store)

	w1 := tm2Store.CacheWrap()
	tx1 := st.BeginTransaction(w1, w1, nil, nil)
	m := NewMachineWithOptions(MachineOptions{PkgPath: "gno.vm/t/hello", Store: tx1, Output: io.Discard})
	m.RunMemPackage(&std.MemPackage{
		Type: MPUserProd, Name: "hello", Path: "gno.vm/t/hello",
		Files: []*std.MemFile{{Name: "hello.gno", Body: "package hello\n\ntype Coin struct{ Denom string; Amount int }\n"}},
	}, true)
	tx1.Write()
	w1.Write()

	first := st.GetType("gno.vm/t/hello.Coin").(*DeclaredType)
	tid := first.TypeID()
	typeHasPrivateDep(st, first)

	if _, ok := storeTypePrivacyCache(st).get(tid); !ok {
		t.Fatal("verdict not memoized after the first walk")
	}

	// A later transaction gets a fresh Type object for the same TypeID.
	// The shared, TypeID-keyed cache must already hold the verdict.
	w2 := tm2Store.CacheWrap()
	tx2 := st.BeginTransaction(w2, w2, nil, nil)
	second := tx2.GetType("gno.vm/t/hello.Coin").(*DeclaredType)
	if first == second {
		t.Fatal("precondition failed: expected a distinct Type object in the second transaction")
	}
	if _, ok := storeTypePrivacyCache(tx2).get(second.TypeID()); !ok {
		t.Fatal("verdict does not survive the transaction boundary: cache miss in the next transaction")
	}
}

// A DeclaredType with a method is self-referential (the method's receiver
// is the type itself), so it forms a cycle. The redesign caches the
// queried root regardless of cycles, so a method-bearing type must
// memoize — the original per-node scheme discarded the whole walk on any
// cycle and never cached these (the common case, since most user types
// carry methods).
func TestTypeHasPrivateDep_MethodBearingTypeIsCached(t *testing.T) {
	t.Parallel()
	db := memdb.NewMemDB()
	tm2Store := dbadapter.StoreConstructor(db, storetypes.StoreOptions{})
	st := NewStore(nil, tm2Store, tm2Store)
	w := tm2Store.CacheWrap()
	tx := st.BeginTransaction(w, w, nil, nil)
	m := NewMachineWithOptions(MachineOptions{PkgPath: "gno.vm/t/hello", Store: tx, Output: io.Discard})
	m.RunMemPackage(&std.MemPackage{
		Type: MPUserProd, Name: "hello", Path: "gno.vm/t/hello",
		Files: []*std.MemFile{{Name: "hello.gno", Body: "package hello\n\ntype Coin struct{ Amount int }\n\nfunc (c Coin) Get() int { return c.Amount }\n"}},
	}, true)

	dt := tx.GetType("gno.vm/t/hello.Coin").(*DeclaredType)
	if len(dt.Methods) == 0 {
		t.Fatal("precondition failed: expected Coin to carry a method")
	}
	typeHasPrivateDep(tx, dt)

	if _, ok := storeTypePrivacyCache(tx).get(dt.TypeID()); !ok {
		t.Fatal("a method-bearing type was not memoized (the self-cycle via the method receiver must not block caching)")
	}
}

// A self-referential type (e.g. a linked-list node) must not infinite-
// loop, must resolve to the correct verdict, AND — unlike the previous
// design — must be cached at the queried root. Self-reference in real Gno
// always goes through a *DeclaredType (DeclaredType.TypeID() is a nominal
// PkgPath+Name hash, not a structural walk, which keeps TypeID() itself
// from infinite-recursing); a raw anonymous *StructType can't close the
// loop, so the test builds the cycle the way the preprocessor would.
func TestTypeHasPrivateDep_SelfReferentialIsCached(t *testing.T) {
	t.Parallel()
	store, pubPath, _ := newPrivateDepTestStore(t)

	nodeDT := &DeclaredType{PkgPath: pubPath, Name: "Node"}
	nodeDT.Base = &StructType{PkgPath: pubPath, Fields: []FieldType{
		{Name: "Next", Type: &PointerType{Elt: nodeDT}},
	}}

	if typeHasPrivateDep(store, nodeDT) {
		t.Fatal("typeHasPrivateDep() = true, want false for a self-referential type with no private dependency")
	}
	if _, ok := storeTypePrivacyCache(store).get(nodeDT.TypeID()); !ok {
		t.Fatal("a self-referential type's root verdict was not memoized")
	}
}

// Regression test for a fixed-point hazard: with two mutually-referencing
// types A <-> B where A also has an unrelated private-dependency field,
// caching an intermediate node's provisional result mid-cycle would be
// wrong. The redesign sidesteps this entirely by caching only the queried
// root (whose complete closure is always explored before it returns), so
// each independent query must still return the correct verdict.
func TestTypeHasPrivateDep_MutualCycleResolvesCorrectly(t *testing.T) {
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
