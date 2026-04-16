package bptree

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

func newTestTree(t *testing.T) *MutableTree {
	t.Helper()
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())
	return tree
}

func TestPersistence_SaveLoadVersion(t *testing.T) {
	tree := newTestTree(t)

	// Insert some keys and save
	for i := 0; i < 50; i++ {
		tree.Set(fmt.Appendf(nil, "k%03d", i), fmt.Appendf(nil, "v%03d", i))
	}
	hash1, v1, err := tree.SaveVersion()
	if err != nil {
		t.Fatalf("SaveVersion: %v", err)
	}
	if v1 != 1 {
		t.Fatalf("version = %d, want 1", v1)
	}
	if hash1 == nil {
		t.Fatalf("hash is nil")
	}

	// More mutations + save
	for i := 50; i < 80; i++ {
		tree.Set(fmt.Appendf(nil, "k%03d", i), fmt.Appendf(nil, "v%03d", i))
	}
	hash2, v2, err := tree.SaveVersion()
	if err != nil {
		t.Fatalf("SaveVersion v2: %v", err)
	}
	if v2 != 2 {
		t.Fatalf("version = %d, want 2", v2)
	}
	if bytes.Equal(hash1, hash2) {
		t.Fatalf("hash should differ between versions")
	}

	// Load version 1 in a new tree
	db2 := tree.ndb.db
	tree2 := NewMutableTreeWithDB(db2, 1000, NewNopLogger())
	loadedV, err := tree2.LoadVersion(1)
	if err != nil {
		t.Fatalf("LoadVersion(1): %v", err)
	}
	// LoadVersion returns the DB's latest version (matching IAVL), not the
	// requested version. The tree is loaded at version 1 but the return
	// value reflects the latest version in the DB.
	if loadedV < 1 {
		t.Fatalf("loaded version = %d, want >= 1", loadedV)
	}
	if tree2.Size() != 50 {
		t.Fatalf("loaded size = %d, want 50", tree2.Size())
	}

	// Verify keys
	for i := 0; i < 50; i++ {
		val, err := tree2.Get(fmt.Appendf(nil, "k%03d", i))
		if err != nil {
			t.Fatalf("Get k%03d: %v", i, err)
		}
		expected := fmt.Appendf(nil, "v%03d", i)
		if !bytes.Equal(val, expected) {
			t.Fatalf("k%03d: got %q, want %q", i, val, expected)
		}
	}

	// Key from v2 should not exist in v1
	val, _ := tree2.Get([]byte("k050"))
	if val != nil {
		t.Fatalf("k050 should not exist in v1")
	}
}

func TestPersistence_Load(t *testing.T) {
	tree := newTestTree(t)
	for i := 0; i < 30; i++ {
		tree.Set(fmt.Appendf(nil, "l%03d", i), []byte("v"))
	}
	tree.SaveVersion()
	tree.Set([]byte("l999"), []byte("extra"))
	tree.SaveVersion()

	// New tree from same DB, Load()
	tree2 := NewMutableTreeWithDB(tree.ndb.db, 1000, NewNopLogger())
	v, err := tree2.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if v != 2 {
		t.Fatalf("loaded version = %d, want 2", v)
	}
	if tree2.Size() != 31 {
		t.Fatalf("size = %d, want 31", tree2.Size())
	}
	has, _ := tree2.Has([]byte("l999"))
	if !has {
		t.Fatalf("l999 not found after Load")
	}
}

func TestPersistence_Rollback(t *testing.T) {
	tree := newTestTree(t)
	for i := 0; i < 20; i++ {
		tree.Set(fmt.Appendf(nil, "r%03d", i), []byte("v"))
	}
	tree.SaveVersion()

	// Mutate
	tree.Set([]byte("r999"), []byte("new"))
	tree.Remove([]byte("r000"))
	if tree.Size() != 20 {
		t.Fatalf("size after mutations = %d", tree.Size())
	}

	// Rollback
	tree.Rollback()
	if tree.Size() != 20 {
		t.Fatalf("size after rollback = %d, want 20", tree.Size())
	}
	has, _ := tree.Has([]byte("r000"))
	if !has {
		t.Fatalf("r000 should exist after rollback")
	}
	has, _ = tree.Has([]byte("r999"))
	if has {
		t.Fatalf("r999 should not exist after rollback")
	}
}

func TestPersistence_MultiVersion(t *testing.T) {
	tree := newTestTree(t)

	// Version 1: 10 keys
	for i := 0; i < 10; i++ {
		tree.Set(fmt.Appendf(nil, "mv%02d", i), []byte("v1"))
	}
	tree.SaveVersion()

	// Version 2: add 10 more
	for i := 10; i < 20; i++ {
		tree.Set(fmt.Appendf(nil, "mv%02d", i), []byte("v2"))
	}
	tree.SaveVersion()

	// Version 3: remove some from v1
	for i := 0; i < 5; i++ {
		tree.Remove(fmt.Appendf(nil, "mv%02d", i))
	}
	tree.SaveVersion()

	// GetImmutable for each version
	imm1, err := tree.GetImmutable(1)
	if err != nil {
		t.Fatalf("GetImmutable(1): %v", err)
	}
	if imm1.Size() != 10 {
		t.Fatalf("v1 size = %d, want 10", imm1.Size())
	}

	imm2, err := tree.GetImmutable(2)
	if err != nil {
		t.Fatalf("GetImmutable(2): %v", err)
	}
	if imm2.Size() != 20 {
		t.Fatalf("v2 size = %d, want 20", imm2.Size())
	}

	imm3, err := tree.GetImmutable(3)
	if err != nil {
		t.Fatalf("GetImmutable(3): %v", err)
	}
	if imm3.Size() != 15 {
		t.Fatalf("v3 size = %d, want 15", imm3.Size())
	}
}

func TestPersistence_HashStability(t *testing.T) {
	// Same operations on two separate trees should produce the same hash
	make := func() *MutableTree {
		return newTestTree(t)
	}

	t1 := make()
	t2 := make()
	for i := 0; i < 100; i++ {
		key := fmt.Appendf(nil, "hs%04d", i)
		val := fmt.Appendf(nil, "val%04d", i)
		t1.Set(key, val)
		t2.Set(key, val)
	}
	h1, _, _ := t1.SaveVersion()
	h2, _, _ := t2.SaveVersion()
	if !bytes.Equal(h1, h2) {
		t.Fatalf("same operations should produce same hash")
	}
}

func TestPersistence_EmptyVersion(t *testing.T) {
	tree := newTestTree(t)
	// Save empty tree
	hash, v, err := tree.SaveVersion()
	if err != nil {
		t.Fatalf("SaveVersion empty: %v", err)
	}
	if v != 1 {
		t.Fatalf("version = %d", v)
	}
	if hash == nil || len(hash) != 32 {
		t.Fatalf("empty tree hash should be SHA256(\"\"), got %x", hash)
	}

	// Load it back
	tree2 := NewMutableTreeWithDB(tree.ndb.db, 1000, NewNopLogger())
	lv, err := tree2.LoadVersion(1)
	if err != nil {
		t.Fatalf("LoadVersion(1): %v", err)
	}
	if lv != 1 || tree2.Size() != 0 || !tree2.IsEmpty() {
		t.Fatalf("loaded empty: v=%d size=%d empty=%v", lv, tree2.Size(), tree2.IsEmpty())
	}
}

func TestPersistence_VersionExists(t *testing.T) {
	tree := newTestTree(t)
	tree.Set([]byte("a"), []byte("b"))
	tree.SaveVersion()

	if !tree.VersionExists(1) {
		t.Fatalf("v1 should exist")
	}
	if tree.VersionExists(2) {
		t.Fatalf("v2 should not exist")
	}
}

func TestPersistence_AvailableVersions(t *testing.T) {
	tree := newTestTree(t)
	for i := 0; i < 3; i++ {
		tree.Set(fmt.Appendf(nil, "av%d", i), []byte("v"))
		tree.SaveVersion()
	}
	versions := tree.AvailableVersions()
	if len(versions) != 3 {
		t.Fatalf("available versions = %v", versions)
	}
	for i, v := range versions {
		if v != i+1 {
			t.Fatalf("version[%d] = %d, want %d", i, v, i+1)
		}
	}
}

func TestPersistence_WorkingVersion(t *testing.T) {
	tree := newTestTree(t)
	if tree.WorkingVersion() != 1 {
		t.Fatalf("initial working version = %d, want 1", tree.WorkingVersion())
	}
	tree.Set([]byte("a"), []byte("b"))
	tree.SaveVersion()
	if tree.WorkingVersion() != 2 {
		t.Fatalf("after v1 save, working version = %d, want 2", tree.WorkingVersion())
	}
}

func TestPersistence_InitialVersion(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger(), InitialVersionOption(10))
	tree.Set([]byte("x"), []byte("y"))
	_, v, err := tree.SaveVersion()
	if err != nil {
		t.Fatalf("SaveVersion: %v", err)
	}
	if v != 10 {
		t.Fatalf("version = %d, want 10", v)
	}
}

func TestPersistence_DeleteVersionsTo(t *testing.T) {
	tree := newTestTree(t)
	for i := 0; i < 5; i++ {
		tree.Set(fmt.Appendf(nil, "dv%d", i), []byte("v"))
		tree.SaveVersion()
	}

	err := tree.DeleteVersionsTo(3)
	if err != nil {
		t.Fatalf("DeleteVersionsTo(3): %v", err)
	}

	if tree.VersionExists(1) || tree.VersionExists(2) || tree.VersionExists(3) {
		t.Fatalf("versions 1-3 should be deleted")
	}
	if !tree.VersionExists(4) || !tree.VersionExists(5) {
		t.Fatalf("versions 4-5 should exist")
	}
}

func TestPersistence_LargeTree_SaveLoad(t *testing.T) {
	tree := newTestTree(t)
	n := 500
	for i := 0; i < n; i++ {
		tree.Set(fmt.Appendf(nil, "big%05d", i), fmt.Appendf(nil, "val%05d", i))
	}
	hash1, _, err := tree.SaveVersion()
	if err != nil {
		t.Fatalf("SaveVersion: %v", err)
	}

	// Load in new tree
	tree2 := NewMutableTreeWithDB(tree.ndb.db, 1000, NewNopLogger())
	tree2.Load()

	if tree2.Size() != int64(n) {
		t.Fatalf("loaded size = %d, want %d", tree2.Size(), n)
	}

	hash2 := tree2.WorkingHash()
	if !bytes.Equal(hash1, hash2) {
		t.Fatalf("loaded hash differs")
	}

	// Verify sorted order
	var keys []string
	tree2.Iterate(func(key, value []byte) bool {
		keys = append(keys, string(key))
		return false
	})
	if !sort.StringsAreSorted(keys) {
		t.Fatalf("keys not sorted after load")
	}

	// Spot check values
	val, _ := tree2.Get([]byte("big00042"))
	if !bytes.Equal(val, []byte("val00042")) {
		t.Fatalf("value mismatch: %q", val)
	}
}

func TestPersistence_SaveMutateSaveLoadBothVersions(t *testing.T) {
	tree := newTestTree(t)

	// V1: keys k000-k049
	for i := 0; i < 50; i++ {
		tree.Set(fmt.Appendf(nil, "k%03d", i), fmt.Appendf(nil, "v%03d", i))
	}
	tree.SaveVersion()

	// Mutate: update, delete, insert
	tree.Set([]byte("k010"), []byte("updated10"))
	tree.Remove([]byte("k020"))
	tree.Set([]byte("k099"), []byte("new99"))
	tree.SaveVersion()

	// Load v2 in fresh tree
	t2 := NewMutableTreeWithDB(tree.ndb.db, 1000, NewNopLogger())
	t2.LoadVersion(2)
	if t2.Size() != 50 { // 50 - 1 delete + 1 insert = 50
		t.Fatalf("v2 size = %d, want 50", t2.Size())
	}
	val, _ := t2.Get([]byte("k010"))
	if !bytes.Equal(val, []byte("updated10")) {
		t.Fatalf("v2 k010 = %q, want 'updated10'", val)
	}
	val, _ = t2.Get([]byte("k020"))
	if val != nil {
		t.Fatalf("v2 k020 should be deleted")
	}
	val, _ = t2.Get([]byte("k099"))
	if !bytes.Equal(val, []byte("new99")) {
		t.Fatalf("v2 k099 = %q, want 'new99'", val)
	}

	// Load v1 — should be unaffected
	t1 := NewMutableTreeWithDB(tree.ndb.db, 1000, NewNopLogger())
	t1.LoadVersion(1)
	if t1.Size() != 50 {
		t.Fatalf("v1 size = %d, want 50", t1.Size())
	}
	val, _ = t1.Get([]byte("k010"))
	if !bytes.Equal(val, fmt.Appendf(nil, "v%03d", 10)) {
		t.Fatalf("v1 k010 = %q, want original", val)
	}
	has, _ := t1.Has([]byte("k020"))
	if !has {
		t.Fatalf("v1 k020 should exist")
	}
	has, _ = t1.Has([]byte("k099"))
	if has {
		t.Fatalf("v1 k099 should not exist")
	}
}

func TestPersistence_SaveVersionNoMutations(t *testing.T) {
	tree := newTestTree(t)
	for i := 0; i < 20; i++ {
		tree.Set(fmt.Appendf(nil, "nm%03d", i), []byte("v"))
	}
	hash1, _, _ := tree.SaveVersion()

	// Save v2 with no mutations
	hash2, v2, err := tree.SaveVersion()
	if err != nil {
		t.Fatalf("SaveVersion no-mutation: %v", err)
	}
	if v2 != 2 {
		t.Fatalf("version = %d, want 2", v2)
	}
	if !bytes.Equal(hash1, hash2) {
		t.Fatalf("no-mutation save should produce same hash")
	}

	// Both versions should be loadable
	t1 := NewMutableTreeWithDB(tree.ndb.db, 1000, NewNopLogger())
	t1.LoadVersion(1)
	if t1.Size() != 20 {
		t.Fatalf("v1 size = %d", t1.Size())
	}

	t2 := NewMutableTreeWithDB(tree.ndb.db, 1000, NewNopLogger())
	t2.LoadVersion(2)
	if t2.Size() != 20 {
		t.Fatalf("v2 size = %d", t2.Size())
	}
}

func TestPersistence_GetReturnsActualValues(t *testing.T) {
	tree := newTestTree(t)
	tree.Set([]byte("hello"), []byte("world"))
	tree.SaveVersion()

	// Reload
	tree2 := NewMutableTreeWithDB(tree.ndb.db, 1000, NewNopLogger())
	tree2.Load()

	val, err := tree2.Get([]byte("hello"))
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !bytes.Equal(val, []byte("world")) {
		t.Fatalf("Get returned %q (len=%d), want 'world'", val, len(val))
	}
}

func TestPersistence_ExportImport(t *testing.T) {
	// Build a tree and export it
	tree1 := newTestTree(t)
	for i := 0; i < 100; i++ {
		tree1.Set(fmt.Appendf(nil, "ei%04d", i), fmt.Appendf(nil, "val%04d", i))
	}
	tree1.SaveVersion()

	imm, _ := tree1.GetImmutable(1)
	exporter, err := imm.Export(tree1.ndb)
	if err != nil {
		t.Fatalf("Export: %v", err)
	}

	// Import into a new tree
	tree2 := newTestTree(t)
	importer, err := tree2.Import(1)
	if err != nil {
		t.Fatalf("Import: %v", err)
	}

	for {
		node, err := exporter.Next()
		if err == ErrExportDone {
			break
		}
		if err != nil {
			t.Fatalf("Next: %v", err)
		}
		if err := importer.Add(node); err != nil {
			t.Fatalf("Add: %v", err)
		}
	}
	exporter.Close()

	if err := importer.Commit(); err != nil {
		t.Fatalf("Commit: %v", err)
	}
	importer.Close()

	// Verify
	if tree2.Size() != 100 {
		t.Fatalf("imported size = %d, want 100", tree2.Size())
	}
	for i := 0; i < 100; i++ {
		val, _ := tree2.Get(fmt.Appendf(nil, "ei%04d", i))
		expected := fmt.Appendf(nil, "val%04d", i)
		if !bytes.Equal(val, expected) {
			t.Fatalf("ei%04d: got %q, want %q", i, val, expected)
		}
	}
}

// TestPersistence_LoadVersionForOverwriting_Unsupported verifies the API
// returns ErrUnsupported (Finding #12) rather than panicking.
func TestPersistence_LoadVersionForOverwriting_Unsupported(t *testing.T) {
	tree := newTestTree(t)
	tree.Set([]byte("k"), []byte("v"))
	tree.SaveVersion()
	err := tree.LoadVersionForOverwriting(1)
	if err == nil {
		t.Fatal("expected error from LoadVersionForOverwriting")
	}
	if !errors.Is(err, ErrUnsupported) {
		t.Fatalf("LoadVersionForOverwriting: got %v, want ErrUnsupported", err)
	}
}

// TestPersistence_DeleteVersionsFrom_Unsupported verifies the API returns
// ErrUnsupported (Finding #12) rather than panicking.
func TestPersistence_DeleteVersionsFrom_Unsupported(t *testing.T) {
	tree := newTestTree(t)
	tree.Set([]byte("k"), []byte("v"))
	tree.SaveVersion()
	tree.SaveVersion()
	err := tree.DeleteVersionsFrom(1)
	if err == nil {
		t.Fatal("expected error from DeleteVersionsFrom")
	}
	if !errors.Is(err, ErrUnsupported) {
		t.Fatalf("DeleteVersionsFrom: got %v, want ErrUnsupported", err)
	}
}

func TestPersistence_VersionReaders_BlockPruning(t *testing.T) {
	tree := newTestTree(t)
	for i := 0; i < 30; i++ {
		tree.Set(fmt.Appendf(nil, "vr%03d", i), []byte("v"))
	}
	tree.SaveVersion()
	tree.Set([]byte("vr_extra"), []byte("v"))
	tree.SaveVersion()

	// Create exporter on version 1 (increments version reader)
	imm, _ := tree.GetImmutable(1)
	exporter, err := imm.Export(tree.ndb)
	if err != nil {
		t.Fatalf("Export: %v", err)
	}

	// Try to delete version 1 — should fail with active reader
	err = tree.DeleteVersionsTo(1)
	if err == nil {
		t.Fatalf("DeleteVersionsTo should fail with active reader")
	}

	// Close exporter AND immutable snapshot — decrements both readers
	exporter.Close()
	imm.Close()

	// Now deletion should succeed
	err = tree.DeleteVersionsTo(1)
	if err != nil {
		t.Fatalf("DeleteVersionsTo after close: %v", err)
	}
	if tree.VersionExists(1) {
		t.Fatalf("version 1 should be deleted")
	}
}

func TestPersistence_Close(t *testing.T) {
	tree := newTestTree(t)
	tree.Set([]byte("a"), []byte("b"))
	tree.SaveVersion()

	err := tree.Close()
	if err != nil {
		t.Fatalf("Close: %v", err)
	}

	// After close, the tree's ndb batch is nil, but the DB is still usable
	// (nodeDB.Close does not close the underlying DB)
}

func TestPersistence_IterateReturnsValues(t *testing.T) {
	tree := newTestTree(t)
	tree.Set([]byte("ik1"), []byte("value_one"))
	tree.Set([]byte("ik2"), []byte("value_two"))
	tree.SaveVersion()

	// Reload from DB
	tree2 := NewMutableTreeWithDB(tree.ndb.db, 1000, NewNopLogger())
	tree2.Load()

	var kvs []string
	tree2.Iterate(func(key, value []byte) bool {
		kvs = append(kvs, string(key)+"="+string(value))
		return false
	})

	if len(kvs) != 2 {
		t.Fatalf("iterate count = %d, want 2", len(kvs))
	}
	if kvs[0] != "ik1=value_one" {
		t.Fatalf("first kv = %q, want 'ik1=value_one'", kvs[0])
	}
	if kvs[1] != "ik2=value_two" {
		t.Fatalf("second kv = %q, want 'ik2=value_two'", kvs[1])
	}
}

func TestPersistence_ExportImportHashMatch(t *testing.T) {
	// The import rebuilds the tree via Set() calls, which may produce a
	// different tree structure (and thus different hash) than the original.
	// This test documents that behavior.
	tree1 := newTestTree(t)
	for i := 0; i < 100; i++ {
		tree1.Set(fmt.Appendf(nil, "hm%04d", i), fmt.Appendf(nil, "val%04d", i))
	}
	hash1, _, _ := tree1.SaveVersion()

	imm, _ := tree1.GetImmutable(1)
	exporter, _ := imm.Export(tree1.ndb)

	tree2 := newTestTree(t)
	importer, _ := tree2.Import(1)
	for {
		node, err := exporter.Next()
		if err == ErrExportDone {
			break
		}
		importer.Add(node)
	}
	exporter.Close()
	importer.Commit()

	hash2 := tree2.WorkingHash()

	// Both trees have the same keys and values
	if tree2.Size() != 100 {
		t.Fatalf("imported size = %d", tree2.Size())
	}

	// Structural export/import must preserve the exact hash.
	if !bytes.Equal(hash1, hash2) {
		t.Fatalf("export/import hash mismatch: %x != %x", hash1, hash2)
	}
}
