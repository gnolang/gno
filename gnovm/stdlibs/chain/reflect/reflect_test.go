package reflect

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	storetypes "github.com/gnolang/gno/tm2/pkg/store/types"
)

// objectFixture is a realm with a finalized owner, plus the pieces needed to
// create objects under it and persist them on demand.
type objectFixture struct {
	m     *gno.Machine
	tx    gno.TransactionStore
	alloc *gno.Allocator
	owner gno.Object
}

func newObjectFixture(t *testing.T, pkgPath string) *objectFixture {
	t.Helper()

	baseStore := dbadapter.StoreConstructor(memdb.NewMemDB(), storetypes.StoreOptions{})
	iavlStore := dbadapter.StoreConstructor(memdb.NewMemDB(), storetypes.StoreOptions{})
	store := gno.NewStore(gno.NewAllocator(math.MaxInt64), baseStore, iavlStore)
	rlm := gno.NewRealm(pkgPath)
	rlm.Time = 1
	store.SetPackageRealm(rlm)

	tx := store.BeginTransaction(baseStore.CacheWrap(), iavlStore.CacheWrap(), nil, nil)
	m := gno.NewMachineWithOptions(gno.MachineOptions{Store: tx})
	m.Realm = tx.GetPackageRealm(pkgPath)

	alloc := gno.NewAllocator(math.MaxInt64)
	owner := alloc.NewStruct(nil, nil)
	owner.SetPkgID(m.Realm.ID)
	owner.SetNewTime(1)

	return &objectFixture{m: m, tx: tx, alloc: alloc, owner: owner}
}

// newStandaloneObject allocates a heap item under the fixture's realm and
// returns the TypedValue a Gno pointer to it produces.
//
// This is the shape the native actually receives: `&T{...}` in Gno is a
// PointerValue whose Base is the heap item that gets persisted, and the heap
// item is the object that carries the ObjectID. Exercising a bare *StructValue
// instead would test a shape the VM never hands to this native.
func (f *objectFixture) newStandaloneObject() (gno.Object, gno.TypedValue) {
	hiv := f.alloc.NewHeapItem(nil, gno.TypedValue{})
	f.own(hiv)

	return hiv, gno.TypedValue{V: gno.PointerValue{TV: &hiv.Value, Base: hiv}}
}

func (f *objectFixture) own(oo gno.Object) {
	oo.SetPkgID(f.m.Realm.ID)
	oo.SetOwner(f.owner)
	oo.IncRefCount()
}

func (f *objectFixture) persist(oo gno.Object) {
	f.m.Realm.MarkNewReal(oo)
	f.m.Realm.FinalizeRealmTransaction(f.tx)
}

// The three states of an object's identity, in the order a realm meets them.
func TestObjectInfoStampingWindow(t *testing.T) {
	f := newObjectFixture(t, "gno.land/r/demo/reflect")
	oo, tv := f.newStandaloneObject()

	// Created by the running call. The realm half of the ID is stamped by the
	// allocator, the clock half only at persistence, so there is no identity
	// and no address to derive from half an ID. It is an object though, which
	// is what tells "not yet" from "never".
	id, addr, _, _, stamped, hasIdentity := X_objectInfo(f.m, tv)
	require.Equal(t, "", id)
	require.Equal(t, "", addr)
	require.False(t, stamped)
	require.True(t, hasIdentity)

	// Reading is not issuing. A second read must agree and must not have
	// advanced the realm clock to mint something.
	timeBefore := f.m.Realm.Time
	_, _, _, _, stamped, _ = X_objectInfo(f.m, tv)
	require.False(t, stamped)
	require.Equal(t, timeBefore, f.m.Realm.Time, "reading must not advance the realm clock")

	f.persist(oo)

	// Persisted: identity and address both exist.
	id, addr, pkgPath, _, stamped, hasIdentity := X_objectInfo(f.m, tv)
	require.True(t, stamped)
	require.True(t, hasIdentity)
	require.Equal(t, oo.GetObjectID().String(), id, "ID must be the VM's own spelling")
	require.Equal(t, "gno.land/r/demo/reflect", pkgPath, "PkgPath must name the creating realm")
	require.NotEmpty(t, addr)

	// The address is exactly what the VM derives from the ID, spelled out
	// independently of the native so a changed derivation reports as such.
	require.Equal(t, gno.DeriveObjectCryptoAddr(oo.GetObjectID()).String(), addr)
	require.Equal(t, "g1", addr[:2], "object addresses are ordinary g1 addresses")

	// And it does not move afterwards.
	timeBefore = f.m.Realm.Time
	_, addr2, _, _, _, _ := X_objectInfo(f.m, tv)
	require.Equal(t, addr, addr2)
	require.Equal(t, timeBefore, f.m.Realm.Time)
}

// Two objects persisted by the same pass take distinct ticks of the realm
// clock, so they take distinct addresses. This is the property the whole
// package exists for: the realm chooses no part of the address and cannot make
// two objects answer the same one.
func TestObjectAddressIsUniquePerObject(t *testing.T) {
	f := newObjectFixture(t, "gno.land/r/demo/reflect_unique")

	const n = 3
	tvs := make([]gno.TypedValue, 0, n)
	for range n {
		oo, tv := f.newStandaloneObject()
		f.m.Realm.MarkNewReal(oo)
		tvs = append(tvs, tv)
	}
	f.m.Realm.FinalizeRealmTransaction(f.tx)

	seen := make(map[string]int, n)
	for i, tv := range tvs {
		_, addr, _, _, stamped, _ := X_objectInfo(f.m, tv)
		require.True(t, stamped)
		require.NotEmpty(t, addr)
		require.NotContains(t, seen, addr, "objects %d and %d share an address", seen[addr], i)
		seen[addr] = i
	}
}

// Two realms mint their clocks independently, so the realm half has to take
// part in the address or tick 2 of one realm would collide with tick 2 of
// another. An object address must also never collide with the package address
// of the realm that owns it: the two are hashed from different preimages, and a
// realm being mistaken for one of its own objects would be a funds bug.
func TestObjectAddressSeparatesRealmsAndPackages(t *testing.T) {
	addrs := make(map[string]string, 4)
	for _, pkgPath := range []string{"gno.land/r/demo/reflect_a", "gno.land/r/demo/reflect_b"} {
		f := newObjectFixture(t, pkgPath)
		oo, tv := f.newStandaloneObject()
		f.persist(oo)

		_, addr, gotPath, _, stamped, _ := X_objectInfo(f.m, tv)
		require.True(t, stamped)
		require.Equal(t, pkgPath, gotPath)
		require.NotContains(t, addrs, addr, "%s and %s share an address", addrs[addr], pkgPath)
		addrs[addr] = pkgPath

		pkgAddr := gno.DerivePkgBech32Addr(pkgPath).String()
		require.NotContains(t, addrs, pkgAddr, "an object address collides with a package address")
		addrs[pkgAddr] = pkgPath + " (package)"
	}
}

// A func value is a reference, so a stored callback is addressable in its own
// right and every holder of it agrees on the address. That is what makes "this
// exact lambda" linkable, separately from whatever struct holds it.
//
// Only the unstamped half is checked here: persisting a synthetic FuncValue
// would need a real package in the store for assertObjectIsPublic to resolve,
// and fabricating one would test the fixture rather than the native. The
// stamped half is proven end to end by
// gnovm/tests/files/zrealm_reflect_proposal_fund.gno, where a realm's stored
// closure reports its own address, distinct from the struct holding it.
func TestFuncValueIsAnObject(t *testing.T) {
	f := newObjectFixture(t, "gno.land/r/demo/reflect_func")

	fv := &gno.FuncValue{}
	f.own(fv)

	_, addr, _, _, stamped, hasIdentity := X_objectInfo(f.m, gno.TypedValue{V: fv})
	require.True(t, hasIdentity, "a func value is an object")
	require.False(t, stamped, "and is unstamped until persisted, like any other")
	require.Equal(t, "", addr)
}

// Every value with no identity of its own reports the same way: nothing, no
// claim to anything, and no panic. A caller gets one uniform answer to check
// rather than a mix of empty strings and aborts, and no address is ever
// invented for a value that has none.
func TestObjectInfoValuesWithoutIdentity(t *testing.T) {
	f := newObjectFixture(t, "gno.land/r/demo/reflect_nonobjects")
	container, _ := f.newStandaloneObject()
	f.persist(container)

	structValue := f.alloc.NewStruct(nil, []gno.TypedValue{{}})
	structValue.SetPkgID(f.m.Realm.ID)
	arrayValue := f.alloc.NewListArray(nil, 1)
	arrayValue.SetPkgID(f.m.Realm.ID)

	tests := []struct {
		name string
		tv   gno.TypedValue
	}{
		{name: "nil interface"},
		{
			name: "scalar",
			tv:   gno.TypedValue{T: gno.StringType},
		},
		{
			name: "nil pointer",
			tv:   gno.TypedValue{V: gno.PointerValue{}},
		},
		{
			// &x.Field. Resolving it to the struct would make every field of x
			// answer x's address as its own.
			name: "pointer into a struct field",
			tv:   gno.TypedValue{V: gno.PointerValue{TV: &structValue.Fields[0], Base: structValue}},
		},
		{
			// &arr[i]. Same aliasing, one container along.
			name: "pointer into an array element",
			tv:   gno.TypedValue{V: gno.PointerValue{TV: &arrayValue.List[0], Base: arrayValue}},
		},
		{
			// A slice resolves to its backing array, shared with every other
			// view of it.
			name: "slice",
			tv:   gno.TypedValue{V: &gno.SliceValue{Base: arrayValue, Length: 1, Maxcap: 1}},
		},
		{
			// A struct arriving by value: whether this is the persisted object
			// or a copy of it depends on how it got here, so it is not claimed.
			name: "struct by value",
			tv:   gno.TypedValue{V: structValue},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, addr, pkgPath, typ, stamped, hasIdentity := X_objectInfo(f.m, tt.tv)
			require.Equal(t, "", id)
			require.Equal(t, "", addr, "a value with no identity must never get an address")
			require.Equal(t, "", pkgPath)
			require.Equal(t, "", typ)
			require.False(t, stamped)
			require.False(t, hasIdentity, "a value with no identity must not claim one is coming")
		})
	}
}

// A package value arrives as a RefValue carrying a PkgPath, and a bare heap
// item is an internal shape. TypedValue.GetFirstObject panics outright on both,
// so the native resolves pointer bases itself and reports these as no identity
// rather than aborting the transaction.
func TestObjectInfoDoesNotPanicOnInternalShapes(t *testing.T) {
	f := newObjectFixture(t, "gno.land/r/demo/reflect_internal")
	hiv := f.alloc.NewHeapItem(nil, gno.TypedValue{})

	tests := []struct {
		name string
		tv   gno.TypedValue
	}{
		{
			name: "package ref value",
			tv:   gno.TypedValue{V: gno.RefValue{PkgPath: "gno.land/r/demo/reflect_internal"}},
		},
		{
			name: "bare heap item",
			tv:   gno.TypedValue{V: hiv},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotPanics(t, func() {
				_, addr, _, _, stamped, hasIdentity := X_objectInfo(f.m, tt.tv)
				require.Equal(t, "", addr)
				require.False(t, stamped)
				require.False(t, hasIdentity)
			})
		})
	}
}

// The docs promise the address is deterministic: a pure function of the ID, so
// a second derivation of the same ID is identical and two different IDs never
// agree. An incomplete ID names no object and has nothing to derive from.
func TestDeriveObjectCryptoAddrIsDeterministic(t *testing.T) {
	t.Parallel()

	pkgID := gno.PkgIDFromPkgPath("gno.land/r/demo/reflect_derive")
	other := gno.PkgIDFromPkgPath("gno.land/r/demo/reflect_derive_other")

	a := gno.DeriveObjectCryptoAddr(gno.ObjectID{PkgID: pkgID, NewTime: 7})
	require.Equal(t, a, gno.DeriveObjectCryptoAddr(gno.ObjectID{PkgID: pkgID, NewTime: 7}))
	require.NotEqual(t, a, gno.DeriveObjectCryptoAddr(gno.ObjectID{PkgID: pkgID, NewTime: 8}))
	require.NotEqual(t, a, gno.DeriveObjectCryptoAddr(gno.ObjectID{PkgID: other, NewTime: 7}))

	for _, oid := range []gno.ObjectID{
		{},
		{NewTime: 7},
		{PkgID: pkgID},
	} {
		require.Panics(t, func() { gno.DeriveObjectCryptoAddr(oid) })
	}
}
