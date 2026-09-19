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
	hiv.SetPkgID(f.m.Realm.ID)
	hiv.SetOwner(f.owner)
	hiv.IncRefCount()

	return hiv, gno.TypedValue{V: gno.PointerValue{TV: &hiv.Value, Base: hiv}}
}

func (f *objectFixture) persist(oo gno.Object) {
	f.m.Realm.MarkNewReal(oo)
	f.m.Realm.FinalizeRealmTransaction(f.tx)
}

// The three states of an object's identity, in the order a realm meets them.
func TestObjectIDStampingWindow(t *testing.T) {
	f := newObjectFixture(t, "gno.land/r/demo/reflect")
	oo, tv := f.newStandaloneObject()

	// Created by the running call. The realm half of the ID is stamped by the
	// allocator, the clock half only at persistence, so there is no ID yet,
	// but it is an object, which is what tells "not yet" from "never".
	id, stamped, hasIdentity := X_objectID(f.m, tv)
	require.Equal(t, "", id)
	require.False(t, stamped)
	require.True(t, hasIdentity)

	// Reading is not issuing. A second read must agree, and must not have
	// advanced the realm clock to mint something.
	timeBefore := f.m.Realm.Time
	id, stamped, hasIdentity = X_objectID(f.m, tv)
	require.Equal(t, "", id)
	require.False(t, stamped)
	require.True(t, hasIdentity)
	require.Equal(t, timeBefore, f.m.Realm.Time, "reading an ID must not advance the realm clock")

	f.persist(oo)

	// Persisted: the ID exists, and it is the VM's own spelling of it, so an
	// ID in an event can be matched against a storage dump.
	id, stamped, hasIdentity = X_objectID(f.m, tv)
	require.True(t, stamped)
	require.True(t, hasIdentity)
	require.Equal(t, oo.GetObjectID().String(), id)
	require.NotEmpty(t, id)

	// And it does not move afterwards.
	timeBefore = f.m.Realm.Time
	id2, _, _ := X_objectID(f.m, tv)
	require.Equal(t, id, id2)
	require.Equal(t, timeBefore, f.m.Realm.Time)
}

// Two objects persisted by the same pass take distinct ticks of the realm
// clock. This is the property the whole package exists for: the realm chooses
// no part of it and cannot make two objects answer the same ID.
func TestObjectIDIsUniquePerObject(t *testing.T) {
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
		id, stamped, _ := X_objectID(f.m, tv)
		require.True(t, stamped)
		require.NotEmpty(t, id)
		require.NotContains(t, seen, id, "objects %d and %d share an ID", seen[id], i)
		seen[id] = i
	}
}

// Two realms mint their clocks independently, so the realm half has to take
// part in the ID or tick 2 of one realm would collide with tick 2 of another.
func TestObjectIDSeparatesRealms(t *testing.T) {
	ids := make(map[string]string, 2)
	for _, pkgPath := range []string{"gno.land/r/demo/reflect_a", "gno.land/r/demo/reflect_b"} {
		f := newObjectFixture(t, pkgPath)
		oo, tv := f.newStandaloneObject()
		f.persist(oo)

		id, stamped, _ := X_objectID(f.m, tv)
		require.True(t, stamped)
		require.NotContains(t, ids, id, "%s and %s share an ID", ids[id], pkgPath)
		ids[id] = pkgPath
	}
}

// Every value that has no identity of its own reports the same way: no ID, no
// claim to one, and no panic. A caller gets one uniform answer to check rather
// than a mix of empty strings and aborts, and the zero ID is never handed out
// as if it were an identity.
func TestObjectIDValuesWithoutIdentity(t *testing.T) {
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
			// &x.Field. Resolving it to the struct would make every field of
			// x answer x's identity as its own.
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
			id, stamped, hasIdentity := X_objectID(f.m, tt.tv)
			require.Equal(t, "", id)
			require.False(t, stamped)
			require.False(t, hasIdentity, "a value with no identity must not claim one is coming")
		})
	}
}

// A package value arrives as a RefValue carrying a PkgPath, and a bare heap
// item is an internal shape. TypedValue.GetFirstObject panics outright on both,
// so the native resolves pointer bases itself and reports these as no identity
// rather than aborting the transaction.
func TestObjectIDDoesNotPanicOnInternalShapes(t *testing.T) {
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
				id, stamped, hasIdentity := X_objectID(f.m, tt.tv)
				require.Equal(t, "", id)
				require.False(t, stamped)
				require.False(t, hasIdentity)
			})
		})
	}
}

// The docs promise an ObjectID is comparable and safe as a map key, which the
// Gno-side ObjectID inherits from the string this returns. Distinct objects
// must therefore land in distinct buckets, and the zero ID must not merge with
// anything: an object created and emitted before persistence reports ok=false,
// but a caller that stored the zero ID anyway would alias every such object.
func TestObjectIDIsUsableAsAMapKey(t *testing.T) {
	f := newObjectFixture(t, "gno.land/r/demo/reflect_mapkey")

	const n = 3
	byID := make(map[string]int, n)
	for i := range n {
		oo, tv := f.newStandaloneObject()
		f.persist(oo)

		id, stamped, _ := X_objectID(f.m, tv)
		require.True(t, stamped)
		byID[id] = i
	}
	require.Len(t, byID, n, "three objects must occupy three keys")

	// The zero ID is one key, shared by everything without an identity, which
	// is why ObjectIDOf reports ok rather than handing it out.
	noID, _, _ := X_objectID(f.m, gno.TypedValue{T: gno.StringType})
	require.NotContains(t, byID, noID)
}
