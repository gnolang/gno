package bptree

import (
	"bytes"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// TestValueKey_NonceStartsAtOne verifies that a fresh MutableTree does
// not allocate ValueKeys with nonce=0. Nonce=0 on version=0 would serialize
// to 12 zero bytes, which collides with the "missing" placeholder that
// LeafNode.Serialize writes for nil valueKeys. See Finding #6.
func TestValueKey_NonceStartsAtOne(t *testing.T) {
	tree := NewMutableTreeMem()
	if _, err := tree.Set([]byte("key"), []byte("val")); err != nil {
		t.Fatalf("Set: %v", err)
	}
	leaf, ok := tree.root.(*LeafNode)
	if !ok {
		t.Fatalf("expected root to be *LeafNode, got %T", tree.root)
	}
	vk := leaf.valueKeys[0]
	if len(vk) != NodeKeySize {
		t.Fatalf("valueKey length = %d, want %d", len(vk), NodeKeySize)
	}
	// The ValueKey must not be all-zero bytes (the "missing" sentinel).
	zero := make([]byte, NodeKeySize)
	if bytes.Equal(vk, zero) {
		t.Fatalf("valueKey is all-zero (collision with missing sentinel): %x", vk)
	}
	nk := GetNodeKey(vk)
	if nk.Nonce != 1 {
		t.Fatalf("first allocated valueKey nonce = %d, want 1", nk.Nonce)
	}
}

// TestValueKey_DBBackedNonceStartsAtOne is the DB-backed variant.
func TestValueKey_DBBackedNonceStartsAtOne(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 100, NewNopLogger(), InlineValueThresholdOption(-1))
	if _, err := tree.Set([]byte("key"), []byte("val")); err != nil {
		t.Fatalf("Set: %v", err)
	}
	leaf := tree.root.(*LeafNode)
	vk := leaf.valueKeys[0]
	zero := make([]byte, NodeKeySize)
	if bytes.Equal(vk, zero) {
		t.Fatalf("valueKey is all-zero: %x", vk)
	}
	nk := GetNodeKey(vk)
	if nk.Nonce != 1 {
		t.Fatalf("first allocated valueKey nonce = %d, want 1", nk.Nonce)
	}
}

// TestValueKey_NonceResetsToOne verifies that the nonce counter is reset
// to 1 (not 0) after SaveVersion. If it reset to 0, the first post-save
// Set on a version-0 import would produce an all-zero valueKey.
func TestValueKey_NonceResetsToOne(t *testing.T) {
	tree := NewMutableTreeMem()
	if _, err := tree.Set([]byte("a"), []byte("1")); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatalf("SaveVersion: %v", err)
	}
	if tree.nextValueNonce != 1 {
		t.Fatalf("nextValueNonce after SaveVersion = %d, want 1", tree.nextValueNonce)
	}
}

// TestValueKey_NonceResetsToOneAfterRollback verifies the same for
// Rollback.
func TestValueKey_NonceResetsToOneAfterRollback(t *testing.T) {
	tree := NewMutableTreeMem()
	if _, err := tree.Set([]byte("a"), []byte("1")); err != nil {
		t.Fatalf("Set: %v", err)
	}
	tree.Rollback()
	if tree.nextValueNonce != 1 {
		t.Fatalf("nextValueNonce after Rollback = %d, want 1", tree.nextValueNonce)
	}
}

// TestImporter_NonceStartsAtOne verifies that a fresh Importer allocates
// ValueKeys starting at nonce=1. Critical for Import(0) where version=0
// would otherwise collide with the missing sentinel.
func TestImporter_NonceStartsAtOne(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 100, NewNopLogger())

	imp, err := tree.Import(1)
	if err != nil {
		t.Fatalf("Import: %v", err)
	}
	if imp.nextNonce != 1 {
		t.Fatalf("Importer.nextNonce = %d, want 1", imp.nextNonce)
	}
}

// TestLoadVersion_OldVersionDoesNotCorruptValues is a regression test for
// the bug where LoadVersion(old) + Set would allocate ValueKeys colliding
// with persisted values under the working-version namespace and — because
// SaveValue writes directly to disk — silently overwrite the live values.
// See Finding #1.1.
//
// Scenario:
//   - v1: Set keyA
//   - v2: Set keyB → persists value under (2, 1)
//   - v3: Set keyC
//   - LoadVersion(1); Set attacker → must NOT overwrite (2, 1)
//   - v2's keyB must still resolve to its original value.
func TestLoadVersion_OldVersionDoesNotCorruptValues(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 100, NewNopLogger())

	mustSetSave := func(key, value string) {
		if _, err := tree.Set([]byte(key), []byte(value)); err != nil {
			t.Fatalf("Set %q: %v", key, err)
		}
		if _, _, err := tree.SaveVersion(); err != nil {
			t.Fatalf("SaveVersion after %q: %v", key, err)
		}
	}
	mustSetSave("keyA", "v1_valA")
	mustSetSave("keyB", "v2_valB")
	mustSetSave("keyC", "v3_valC")

	// Snapshot version 2's view of keyB before the attack.
	imm2, err := tree.GetImmutable(2)
	if err != nil {
		t.Fatalf("GetImmutable(2): %v", err)
	}
	pre, _ := imm2.Get([]byte("keyB"))
	imm2.Close()
	if !bytes.Equal(pre, []byte("v2_valB")) {
		t.Fatalf("pre-check: v2.keyB = %q, want v2_valB", pre)
	}

	// Attack: load an older version, Set a new key. A naive allocator
	// reuses nonces starting at 1 in the (v1+1)=v2 namespace, which
	// collides with v2's existing (2,1) slot for keyB.
	if _, err := tree.LoadVersion(1); err != nil {
		t.Fatalf("LoadVersion(1): %v", err)
	}
	if _, err := tree.Set([]byte("attacker"), []byte("ATTACKER_DATA")); err != nil {
		t.Fatalf("Set attacker: %v", err)
	}

	// v2's keyB must still resolve to its original value.
	imm2b, err := tree.GetImmutable(2)
	if err != nil {
		t.Fatalf("GetImmutable(2) post-attack: %v", err)
	}
	post, _ := imm2b.Get([]byte("keyB"))
	imm2b.Close()
	if !bytes.Equal(post, []byte("v2_valB")) {
		t.Fatalf("post-attack: v2.keyB = %q, want v2_valB (disk corruption!)", post)
	}
}

// TestValueKey_LegacyNonceZeroStoreReadable verifies that a store
// previously written with nonce=0 (before Finding #6 was fixed) remains
// readable. We simulate this by manually writing a value under the
// legacy all-zero valueKey and confirming Get retrieves it.
//
// This establishes the migration path: the fix changes the allocator,
// but the reader makes no distinction between "missing sentinel" and
// "valid nonce=0 key" — it simply does a DB lookup. Therefore, existing
// stores with nonce=0 valueKeys continue to resolve correctly.
func TestValueKey_LegacyNonceZeroStoreReadable(t *testing.T) {
	db := memdb.NewMemDB()
	// Force external storage so we exercise the ValueKey indirection
	// under test. With inline values enabled (the default), a 3-byte
	// "v" would take the in-leaf path and bypass the ValueKey entirely.
	tree := NewMutableTreeWithDB(db, 100, NewNopLogger(), InlineValueThresholdOption(-1))

	if _, err := tree.Set([]byte("k"), []byte("original")); err != nil {
		t.Fatalf("Set: %v", err)
	}
	leaf := tree.root.(*LeafNode)

	// Overwrite with a legacy-format valueKey and stash the value
	// under the legacy key directly. Clear any inline state and the
	// fast-node cache entry so the reader actually goes through the
	// ValueKey resolver path under test.
	leaf.inlineMask = 0
	leaf.inlineValues[0] = nil
	if tree.fastNodes != nil {
		tree.fastNodes.Purge()
	}
	legacyVK := make([]byte, NodeKeySize) // all zeros
	leaf.valueKeys[0] = legacyVK
	if err := tree.ndb.SaveValue([]byte("legacy-value"), legacyVK); err != nil {
		t.Fatalf("SaveValue: %v", err)
	}
	if err := tree.ndb.Commit(); err != nil {
		t.Fatalf("Commit: %v", err)
	}

	// Read back — must resolve to the stored value, not panic or
	// treat the all-zero key as missing.
	got, err := tree.Get([]byte("k"))
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !bytes.Equal(got, []byte("legacy-value")) {
		t.Fatalf("Get = %q, want %q", got, "legacy-value")
	}
}
