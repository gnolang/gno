package gnolang

import (
	"fmt"
	"io"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	storetypes "github.com/gnolang/gno/tm2/pkg/store/types"
	"github.com/stretchr/testify/require"
)

// printOwnershipTree logs the ownership chain for the given object IDs,
// showing each object's type, hash (first 8 bytes), and owner.
// Run with `go test -v` to see the output.
func printOwnershipTree(t *testing.T, store Store, baseStore storetypes.Store, label string, oids []ObjectID) {
	t.Helper()
	t.Logf("=== Ownership tree: %s ===", label)
	for _, oid := range oids {
		obj := store.GetObject(oid)
		hash := loadObjectHashFromDB(baseStore, oid)
		ownerID := obj.GetOwnerID()
		typeName := fmt.Sprintf("%T", obj)
		if ownerID.IsZero() {
			t.Logf("  %s  %-24s  hash=%X  owner=(none)", oid, typeName, hash.Bytes()[:8])
		} else {
			t.Logf("  %s  %-24s  hash=%X  owner=%s", oid, typeName, hash.Bytes()[:8], ownerID)
		}
	}
}

// loadObjectHashFromDB reads an object directly from the baseStore DB,
// bypassing the in-memory cache. Returns the hash stored alongside the bytes.
func loadObjectHashFromDB(baseStore storetypes.Store, oid ObjectID) ValueHash {
	key := backendObjectKey(oid)
	hashbz := baseStore.Get(nil, []byte(key))
	if hashbz == nil {
		return ValueHash{}
	}
	hash := hashbz[:HashSize]
	return ValueHash{NewHashlet(hash)}
}

// loadObjectFromDB reads an object's raw amino bytes from the baseStore,
// bypassing the in-memory cache. Returns the deserialized object with
// RefValues intact (children are NOT hydrated — they remain as RefValue).
func loadObjectFromDB(baseStore storetypes.Store, oid ObjectID) Object {
	key := backendObjectKey(oid)
	hashbz := baseStore.Get(nil, []byte(key))
	if hashbz == nil {
		return nil
	}
	bz := hashbz[HashSize:]
	var oo Object
	amino.MustUnmarshal(bz, &oo)
	return oo
}

// loadObjectBytesFromDB returns the raw amino bytes for an object as stored
// in the backend (without the hash prefix).
func loadObjectBytesFromDB(baseStore storetypes.Store, oid ObjectID) []byte {
	key := backendObjectKey(oid)
	hashbz := baseStore.Get(nil, []byte(key))
	if hashbz == nil {
		return nil
	}
	return hashbz[HashSize:]
}

// findRefValueByOID searches the serialized form of a parent object for a
// RefValue that references the given child ObjectID. This lets us verify
// that the embedded hash matches the child's actual stored hash.
func findRefValueByOID(parent Object, childOID ObjectID) (RefValue, bool) {
	switch pv := parent.(type) {
	case *HeapItemValue:
		if ref, ok := pv.Value.V.(RefValue); ok && ref.ObjectID == childOID {
			return ref, true
		}
	case *Block:
		for _, tv := range pv.Values {
			if ref, ok := tv.V.(RefValue); ok && ref.ObjectID == childOID {
				return ref, true
			}
		}
	case *ArrayValue:
		for _, tv := range pv.List {
			if ref, ok := tv.V.(RefValue); ok && ref.ObjectID == childOID {
				return ref, true
			}
		}
	case *StructValue:
		for _, tv := range pv.Fields {
			if ref, ok := tv.V.(RefValue); ok && ref.ObjectID == childOID {
				return ref, true
			}
		}
	}
	return RefValue{}, false
}

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
	txSt1 := st.BeginTransaction(nil, nil, nil, nil)

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

	// --- Print and record hashes after init ---
	oids := []ObjectID{blockOID, heapItemOID, mapOID}
	printOwnershipTree(t, st, baseStore, "After init", oids)

	blockHashInit := loadObjectHashFromDB(baseStore, blockOID)
	heapItemHashInit := loadObjectHashFromDB(baseStore, heapItemOID)
	mapHashInit := loadObjectHashFromDB(baseStore, mapOID)

	require.False(t, blockHashInit.IsZero(), "block hash should be set after init")
	require.False(t, heapItemHashInit.IsZero(), "heapitem hash should be set after init")
	require.False(t, mapHashInit.IsZero(), "map hash should be set after init")

	// Verify OwnerID is persisted correctly in the DB.
	{
		mapFromDB := loadObjectFromDB(baseStore, mapOID)
		require.Equal(t, heapItemOID, mapFromDB.GetObjectInfo().OwnerID,
			"map's persisted OwnerID should point to heapitem")
	}

	// Verify GetOwnerID returns the correct persisted OwnerID,
	// even when the owner pointer hasn't been hydrated yet.
	{
		mapObj := st.GetObject(mapOID)
		require.Equal(t, heapItemOID, mapObj.GetOwnerID(),
			"GetOwnerID() should return the persisted OwnerID")
	}

	// Verify the RefValue.Hash chain is consistent after init.
	{
		hivFromDB := loadObjectFromDB(baseStore, heapItemOID)
		ref, found := findRefValueByOID(hivFromDB, mapOID)
		require.True(t, found, "heapitem should contain RefValue pointing to map")
		require.Equal(t, mapHashInit, ref.Hash,
			"heapitem's RefValue.Hash should match map's actual hash after init")
	}
	{
		blkFromDB := loadObjectFromDB(baseStore, blockOID)
		ref, found := findRefValueByOID(blkFromDB, heapItemOID)
		require.True(t, found, "block should contain RefValue pointing to heapitem")
		require.Equal(t, heapItemHashInit, ref.Hash,
			"block's RefValue.Hash should match heapitem's actual hash after init")
	}

	// --- Transaction 2: Run main() which modifies the map ---
	pv2 := st.GetPackage(pkgPath, false)
	m2 := NewMachineWithOptions(MachineOptions{
		PkgPath: pkgPath,
		Store:   st,
		Output:  io.Discard,
	})
	m2.SetActivePackage(pv2)
	m2.RunMain()

	// --- Verify ancestors were re-saved after main ---
	printOwnershipTree(t, st, baseStore, "After main()", oids)

	mapHashMain := loadObjectHashFromDB(baseStore, mapOID)
	heapItemHashMain := loadObjectHashFromDB(baseStore, heapItemOID)
	blockHashMain := loadObjectHashFromDB(baseStore, blockOID)

	// The map's hash MUST have changed (we modified m["a"] from 1 to 2).
	require.NotEqual(t, mapHashInit, mapHashMain,
		"map hash should change after modification in main()")

	// Ancestors must have been re-saved (hash changed), proving
	// markDirtyAncestors walked the full ownership chain.
	require.NotEqual(t, heapItemHashInit, heapItemHashMain,
		"heapitem hash should change — ancestor must be re-saved when child changes")
	require.NotEqual(t, blockHashInit, blockHashMain,
		"block hash should change — ancestor must be re-saved when descendant changes")

	// --- Verify RefValue.Hash chain is consistent after main ---
	// Each parent's stored bytes must embed the child's CURRENT hash.
	{
		hivFromDB := loadObjectFromDB(baseStore, heapItemOID)
		ref, found := findRefValueByOID(hivFromDB, mapOID)
		require.True(t, found, "heapitem should contain RefValue pointing to map")
		require.Equal(t, mapHashMain, ref.Hash,
			"heapitem's RefValue.Hash should match map's new hash after main")
	}
	{
		blkFromDB := loadObjectFromDB(baseStore, blockOID)
		ref, found := findRefValueByOID(blkFromDB, heapItemOID)
		require.True(t, found, "block should contain RefValue pointing to heapitem")
		require.Equal(t, heapItemHashMain, ref.Hash,
			"block's RefValue.Hash should match heapitem's new hash after main")
	}

	// --- Verify hash self-consistency ---
	// Each object's stored hash must equal HashBytes(its stored amino bytes).
	// This proves the hash is a faithful digest of the bytes, and the bytes
	// include the child's RefValue.Hash — so the parent hash transitively
	// depends on all descendant content.
	for _, oid := range oids {
		bz := loadObjectBytesFromDB(baseStore, oid)
		require.NotNil(t, bz, "object %s should exist in store", oid)
		storedHash := loadObjectHashFromDB(baseStore, oid)
		recomputed := ValueHash{HashBytes(bz)}
		require.Equal(t, storedHash, recomputed,
			"object %s: stored hash should equal HashBytes(stored bytes)", oid)
	}
}
