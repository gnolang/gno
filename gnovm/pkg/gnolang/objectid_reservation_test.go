package gnolang

import (
	"io"
	"math"
	"sort"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	storetypes "github.com/gnolang/gno/tm2/pkg/store/types"
	"github.com/stretchr/testify/require"
)

func TestObjectID_AllocationStampsPkgIDNotNewTime(t *testing.T) {
	rlm := NewRealm("gno.land/r/demo/objid_reservation")

	alloc := NewAllocator(math.MaxInt64)
	alloc.currentRealmID = rlm.ID
	alloc.currentRealmPath = rlm.Path

	sv := alloc.NewStruct(nil, nil)

	oid := sv.GetObjectID()
	require.Equal(t, rlm.ID, oid.PkgID, "allocation must stamp the executing realm PkgID")
	require.Equal(t, uint64(0), oid.NewTime, "allocation must leave NewTime unassigned")
	require.False(t, oid.IsFinalized(), "freshly allocated object must not be finalized")
	require.False(t, sv.GetIsReal(), "freshly allocated object must not be real")
}

func TestObjectID_NewTimeAssignedOnlyAtFinalize(t *testing.T) {
	rlm := NewRealm("gno.land/r/demo/objid_reservation")
	alloc := NewAllocator(math.MaxInt64)
	alloc.currentRealmID = rlm.ID
	alloc.currentRealmPath = rlm.Path

	a := alloc.NewStruct(nil, nil)
	b := alloc.NewStruct(nil, nil)

	require.Equal(t, uint64(0), a.GetObjectID().NewTime)
	require.Equal(t, uint64(0), b.GetObjectID().NewTime)
	require.Equal(t, a.GetObjectID(), b.GetObjectID(),
		"two unfinalized objects share an indistinguishable ObjectID: the collision window #6026 describes")

	rlm.assignNewObjectID(nil, a)
	rlm.assignNewObjectID(nil, b)

	require.True(t, a.GetObjectID().IsFinalized())
	require.True(t, b.GetObjectID().IsFinalized())
	require.NotEqual(t, a.GetObjectID(), b.GetObjectID(),
		"finalize must give each object a distinct ObjectID")
	require.Equal(t, uint64(1), a.GetObjectID().NewTime)
	require.Equal(t, uint64(2), b.GetObjectID().NewTime)
}

func TestObjectID_RealmTimeIsMonotonicUniqueCounter(t *testing.T) {
	db := memdb.NewMemDB()
	baseStore := dbadapter.StoreConstructor(db, storetypes.StoreOptions{})
	iavlStore := dbadapter.StoreConstructor(memdb.NewMemDB(), storetypes.StoreOptions{})
	st := NewStore(NewAllocator(math.MaxInt64), baseStore, iavlStore)
	pkgPath := "gno.land/r/demo/objid_reservation"

	baseTx := baseStore.CacheWrap()
	iavlTx := iavlStore.CacheWrap()
	tx := st.BeginTransaction(baseTx, iavlTx, nil, nil)
	m := NewMachineWithOptions(MachineOptions{PkgPath: pkgPath, Store: tx, Output: io.Discard})
	m.RunMemPackage(&std.MemPackage{
		Type: MPUserProd,
		Name: "objid_reservation",
		Path: pkgPath,
		Files: []*std.MemFile{
			{Name: "gnomod.toml", Body: GenGnoModLatest(pkgPath)},
			{Name: "main.gno", Body: `package objid_reservation

type item struct {
	n int
}

var items []*item

func main() {
	items = append(items, &item{n: len(items)}, &item{n: len(items) + 1})
}
`},
		},
	}, true)
	tx.Write()
	baseTx.Write()
	iavlTx.Write()

	previousTime := st.BeginTransaction(nil, nil, nil, nil).GetPackageRealm(pkgPath).Time
	seen := map[ObjectID]bool{}
	const rounds = 3
	for range rounds {
		baseTx = baseStore.CacheWrap()
		iavlTx = iavlStore.CacheWrap()
		tx = st.BeginTransaction(baseTx, iavlTx, nil, nil)
		reloaded := tx.GetPackageRealm(pkgPath)
		require.Equal(t, previousTime, reloaded.Time, "Realm.Time must survive the transaction store round-trip")

		pv := tx.GetPackage(pkgPath, false)
		m = NewMachineWithOptions(MachineOptions{PkgPath: pkgPath, Store: tx, Output: io.Discard})
		m.SetActivePackage(pv)
		m.RunMain()
		tx.Write()
		baseTx.Write()
		iavlTx.Write()

		fresh := NewStore(NewAllocator(math.MaxInt64), baseStore, iavlStore)
		persisted := fresh.GetPackageRealm(pkgPath)
		require.Greater(t, persisted.Time, previousTime, "a transaction that finalizes objects must advance Realm.Time")

		ids := persistedStructObjectIDs(t, baseStore, persisted.ID)
		newIDs := make([]ObjectID, 0, 2)
		for _, oid := range ids {
			if seen[oid] {
				continue
			}
			require.Greater(t, oid.NewTime, previousTime, "new objects must use a counter after the prior transaction")
			seen[oid] = true
			newIDs = append(newIDs, oid)
		}
		require.Len(t, newIDs, 2, "each main call must persist two distinct item objects")
		require.Less(t, newIDs[0].NewTime, newIDs[1].NewTime, "finalized ObjectIDs must be strictly monotonic")
		previousTime = persisted.Time
	}
	require.Len(t, seen, 6)
}

func TestObjectID_ForeignObjectUsesOwningRealmCounter(t *testing.T) {
	baseStore := dbadapter.StoreConstructor(memdb.NewMemDB(), storetypes.StoreOptions{})
	st := NewStore(NewAllocator(math.MaxInt64), baseStore, baseStore)
	finalizingRealm := NewRealm("gno.land/r/demo/finalizer")
	ownerRealm := NewRealm("gno.land/r/demo/owner")
	ownerRealm.Time = 7
	ownerNode := NewPackageNode("owner", ownerRealm.Path, &FileSet{})
	st.SetCachePackage(ownerNode.NewPackage(NewAllocator(math.MaxInt64)))
	st.SetPackageRealm(ownerRealm)

	alloc := NewAllocator(math.MaxInt64)
	alloc.currentRealmID = ownerRealm.ID
	alloc.currentRealmPath = ownerRealm.Path
	object := alloc.NewStruct(nil, nil)

	oid := finalizingRealm.assignNewObjectID(st, object)
	require.Equal(t, ObjectID{PkgID: ownerRealm.ID, NewTime: 8}, oid)
	require.Zero(t, finalizingRealm.Time, "the finalizing realm must not issue a foreign object's NewTime")
	require.Equal(t, uint64(8), ownerRealm.Time)

	finalizingRealm.FinalizeRealmTransaction(st)
	fresh := NewStore(NewAllocator(math.MaxInt64), baseStore, baseStore)
	require.Equal(t, uint64(8), fresh.GetPackageRealm(ownerRealm.Path).Time,
		"the owning realm's advanced counter must persist")
}

func TestObjectID_AdoptedObjectsUsePersistingRealmProvenance(t *testing.T) {
	rlm := NewRealm("gno.land/r/demo/adopter")
	baseStore := dbadapter.StoreConstructor(memdb.NewMemDB(), storetypes.StoreOptions{})
	st := NewStore(NewAllocator(math.MaxInt64), baseStore, baseStore)

	ephemeralPath := "gno.land/e/g1user123/run"
	ephemeralNode := NewPackageNode("run", ephemeralPath, &FileSet{})
	st.SetCachePackage(ephemeralNode.NewPackage(NewAllocator(math.MaxInt64)))
	stdlibNode := NewPackageNode("bytes", "bytes", &FileSet{})
	st.SetCachePackage(stdlibNode.NewPackage(NewAllocator(math.MaxInt64)))

	for _, sourcePath := range []string{"bytes", ephemeralPath} {
		alloc := NewAllocator(math.MaxInt64)
		alloc.currentRealmID = PkgIDFromPkgPath(sourcePath)
		alloc.currentRealmPath = sourcePath
		object := alloc.NewStruct(nil, nil)

		oid := rlm.assignNewObjectID(st, object)
		require.Equal(t, rlm.ID, oid.PkgID, "%s object must be adopted by the persisting realm", sourcePath)
		require.Equal(t, rlm.Time, oid.NewTime, "%s object must use the persisting realm counter", sourcePath)
	}
	require.Equal(t, uint64(2), rlm.Time)
}

func persistedStructObjectIDs(t *testing.T, baseStore storetypes.Store, pkgID PkgID) []ObjectID {
	t.Helper()
	iter := baseStore.Iterator(nil, nil, nil)
	defer iter.Close()

	ids := make([]ObjectID, 0)
	for ; iter.Valid(); iter.Next() {
		key := string(iter.Key())
		if len(key) < 6 || key[:4] != "oid:" || key[len(key)-6:] == "#realm" {
			continue
		}
		var oid ObjectID
		require.NoError(t, oid.UnmarshalAmino(key[4:]))
		if oid.PkgID == pkgID {
			if _, ok := loadObjectFromDB(baseStore, oid).(*StructValue); ok {
				ids = append(ids, oid)
			}
		}
	}
	require.NoError(t, iter.Error())
	sort.Slice(ids, func(i, j int) bool { return ids[i].NewTime < ids[j].NewTime })
	return ids
}

func TestObjectID_SharedRealmTimeModelNeverCollides(t *testing.T) {
	rlm := NewRealm("gno.land/r/demo/objid_reservation")
	alloc := NewAllocator(math.MaxInt64)
	alloc.currentRealmID = rlm.ID
	alloc.currentRealmPath = rlm.Path

	used := map[uint64]bool{}

	reserveID := func() uint64 {
		if rlm.Time == math.MaxUint64 {
			t.Fatal("realm time overflow")
		}
		rlm.Time++
		return rlm.Time
	}

	finalizeOne := func() uint64 {
		sv := alloc.NewStruct(nil, nil)
		rlm.assignNewObjectID(nil, sv)
		return sv.GetObjectID().NewTime
	}

	record := func(v uint64) {
		require.False(t, used[v], "value %d handed out twice across reservations and finalizations", v)
		used[v] = true
	}

	record(finalizeOne())
	record(reserveID())
	record(reserveID())
	record(finalizeOne())
	record(reserveID())
	record(finalizeOne())

	require.Len(t, used, 6)

	next := finalizeOne()
	for v := range used {
		require.NotEqual(t, v, next, "a later object must not reuse a reserved value")
	}
}

func TestObjectID_RealmTimeOverflowGuard(t *testing.T) {
	t.Skip("assignNewObjectID has no overflow guard; enable when the production guard is added")

	rlm := NewRealm("gno.land/r/demo/objid_reservation")
	rlm.Time = math.MaxUint64
	alloc := NewAllocator(math.MaxInt64)
	alloc.currentRealmID = rlm.ID
	alloc.currentRealmPath = rlm.Path
	object := alloc.NewStruct(nil, nil)

	require.Panics(t, func() { rlm.assignNewObjectID(nil, object) })
}
