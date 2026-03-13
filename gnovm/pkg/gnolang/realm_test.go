package gnolang

import (
	"io"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	storetypes "github.com/gnolang/gno/tm2/pkg/store/types"
	"github.com/stretchr/testify/require"
)

// loadObjectHashFromDB reads an object directly from the baseStore DB,
// bypassing the in-memory cache. Returns the hash stored alongside the bytes.
func loadObjectHashFromDB(baseStore storetypes.Store, oid ObjectID) ValueHash {
	key := backendObjectKey(oid)
	hashbz := baseStore.Get([]byte(key))
	if hashbz == nil {
		return ValueHash{}
	}
	hash := hashbz[:HashSize]
	return ValueHash{NewHashlet(hash)}
}

// TestMarkDirtyAncestors_HashConsistency proves that when a child object
// (like a map) is modified, its ancestors must also be re-saved so their
// hashes reflect the child's new hash via the RefValue{ObjectID, Hash} chain.
//
// Object layout after init:
//
//	:1 PackageValue
//	:2 Block (package block, owns :3 and :5)
//	:3 HeapItemValue (wraps the map var, owned by :2)
//	:4 MapValue (owned by :3)
//	:5 FuncValue (main, owned by :2)
//
// When main() runs m["a"] = 2, MapValue :4 is dirtied.
// markDirtyAncestors should walk :4 → :3 → :2, dirtying all of them.
// But GetOwnerID() returns zero for store-restored objects (owner pointer is nil),
// so the walk stops immediately and :3 and :2 are never re-saved.
func TestMarkDirtyAncestors_HashConsistency(t *testing.T) {
	// --- Setup ---
	db := memdb.NewMemDB()
	baseStore := dbadapter.StoreConstructor(db, storetypes.StoreOptions{})
	iavlStore := dbadapter.StoreConstructor(memdb.NewMemDB(), storetypes.StoreOptions{})
	st := NewStore(nil, baseStore, iavlStore)

	pkgPath := "gno.land/r/test_hash"
	pkgOID := ObjectIDFromPkgPath(pkgPath)
	blockOID := ObjectID{PkgID: pkgOID.PkgID, NewTime: 2}
	heapItemOID := ObjectID{PkgID: pkgOID.PkgID, NewTime: 3}
	mapOID := ObjectID{PkgID: pkgOID.PkgID, NewTime: 4}

	// --- Transaction 1: Initialize realm with a map ---
	// (following the filetest pattern: use transaction store, then commit)
	txSt1 := st.BeginTransaction(nil, nil, nil)

	m1 := NewMachineWithOptions(MachineOptions{
		PkgPath: pkgPath,
		Store:   txSt1,
		Output:  io.Discard,
	})

	mpkg := &std.MemPackage{
		Type: MPUserProd,
		Name: "test_hash",
		Path: pkgPath,
		Files: []*std.MemFile{
			{Name: "gnomod.toml", Body: GenGnoModLatest(pkgPath)},
			{Name: "test.gno", Body: `package test_hash

var m = map[string]int{"a": 1}

func main() {
	m["a"] = 2
}
`},
		},
	}

	m1.RunMemPackage(mpkg, true)
	txSt1.Write()

	// --- Record hashes after init (read directly from DB) ---
	blockHashInit := loadObjectHashFromDB(baseStore, blockOID)
	heapItemHashInit := loadObjectHashFromDB(baseStore, heapItemOID)
	mapHashInit := loadObjectHashFromDB(baseStore, mapOID)

	require.False(t, blockHashInit.IsZero(), "block hash should be set after init")
	require.False(t, heapItemHashInit.IsZero(), "heapitem hash should be set after init")
	require.False(t, mapHashInit.IsZero(), "map hash should be set after init")

	// Verify OwnerID is persisted via the DB bytes.
	{
		key := backendObjectKey(mapOID)
		hashbz := baseStore.Get([]byte(key))
		bz := hashbz[HashSize:]
		var oo Object
		amino.MustUnmarshal(bz, &oo)
		require.Equal(t, heapItemOID, oo.GetObjectInfo().OwnerID,
			"map's persisted OwnerID should point to heapitem")
	}

	// Verify GetOwnerID returns the correct persisted OwnerID,
	// even when the owner pointer hasn't been hydrated yet.
	{
		mapObj := st.GetObject(mapOID)
		require.Equal(t, heapItemOID, mapObj.GetOwnerID(),
			"GetOwnerID() should return the persisted OwnerID")
	}

	// --- Transaction 2: Run main() which modifies the map ---
	// (following filetest pattern: reload package from base store)
	pv2 := st.GetPackage(pkgPath, false)
	m2 := NewMachineWithOptions(MachineOptions{
		PkgPath: pkgPath,
		Store:   st,
		Output:  io.Discard,
	})
	m2.SetActivePackage(pv2)
	m2.RunMain()

	// --- Verify hashes after main (read directly from DB) ---
	mapHashMain := loadObjectHashFromDB(baseStore, mapOID)
	heapItemHashMain := loadObjectHashFromDB(baseStore, heapItemOID)
	blockHashMain := loadObjectHashFromDB(baseStore, blockOID)

	// The map's hash MUST have changed (we modified m["a"] from 1 to 2).
	require.NotEqual(t, mapHashInit, mapHashMain,
		"map hash should change after modification in main()")

	// The heapitem contains RefValue{Hash} for the map.
	// If the map hash changed, the heapitem must be re-saved with the new hash.
	if heapItemHashInit == heapItemHashMain {
		t.Errorf("MERKLE INCONSISTENCY at HeapItemValue (:3):\n"+
			"  heapitem hash unchanged despite child map hash changing.\n"+
			"  heapitem hash (init): %X\n"+
			"  heapitem hash (main): %X  (should differ!)\n"+
			"  map hash (init):      %X\n"+
			"  map hash (main):      %X",
			heapItemHashInit.Bytes(), heapItemHashMain.Bytes(),
			mapHashInit.Bytes(), mapHashMain.Bytes())
	}

	// The block contains RefValue{Hash} for the heapitem.
	// If the heapitem hash changed, the block must be re-saved too.
	if blockHashInit == blockHashMain {
		t.Errorf("MERKLE INCONSISTENCY at Block (:2):\n"+
			"  block hash unchanged despite descendant map hash changing.\n"+
			"  block hash (init):    %X\n"+
			"  block hash (main):    %X  (should differ!)\n"+
			"  map hash (init):      %X\n"+
			"  map hash (main):      %X",
			blockHashInit.Bytes(), blockHashMain.Bytes(),
			mapHashInit.Bytes(), mapHashMain.Bytes())
	}
}
