package bptree

// Comprehensive tests for bugs identified during code review.
// Each TestBugN_* group targets a specific issue.

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/rand"
	"runtime"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// =========================================================================
// Bug #1: Shallow Clone Shares Mutable Byte Slices
//
// Clone() does `c := *n` which copies fixed-size arrays by value but shares
// the individual []byte slices within them. Currently safe because the
// codebase only does slot reassignment, never in-place mutation. These tests
// verify the invariant holds and demonstrate what would break if violated.
// =========================================================================

func TestBug1_CloneKeySlicesAreShared(t *testing.T) {
	// Demonstrates that Clone() shares the underlying []byte key data.
	// This is the fundamental observation — not a bug today, but a latent risk.
	leaf := &LeafNode{miniTree: NewMiniMerkle()}
	leaf.keys[0] = []byte("hello")
	leaf.numKeys = 1
	leaf.valueHashes[0] = sha256.Sum256([]byte("world"))
	leaf.RebuildMiniMerkle()

	cloned := leaf.Clone()

	// The slice headers point to the same backing array
	if &leaf.keys[0][0] != &cloned.keys[0][0] {
		t.Fatal("expected Clone to share key backing arrays")
	}

	// In-place mutation of cloned key DOES corrupt the original
	cloned.keys[0][0] = 'X'
	if leaf.keys[0][0] != 'X' {
		t.Fatal("expected in-place mutation to affect original (shared backing)")
	}
	// Restore for cleanliness
	cloned.keys[0][0] = 'h'

	// But array slot reassignment is safe — this is what the codebase does
	cloned.keys[0] = []byte("different")
	if string(leaf.keys[0]) != "hello" {
		t.Fatal("slot reassignment should not affect original")
	}
}

func TestBug1_CloneInnerChildrenSlicesAreShared(t *testing.T) {
	inner := &InnerNode{
		nodeKey: &NodeKey{Version: 1, Nonce: 1},
		numKeys: 1,
		height:  1,
		size:    2,
	}
	inner.children[0] = (&NodeKey{Version: 1, Nonce: 10}).GetKey()
	inner.children[1] = (&NodeKey{Version: 1, Nonce: 11}).GetKey()

	cloned := inner.Clone()

	// children[i] are shared — in-place mutation crosses the boundary
	original0 := make([]byte, len(inner.children[0]))
	copy(original0, inner.children[0])
	cloned.children[0][0] = 0xFF
	if inner.children[0][0] != 0xFF {
		t.Fatal("expected in-place mutation of cloned children to affect original")
	}
	// Restore
	copy(inner.children[0], original0)

	// Slot reassignment is safe
	cloned.children[0] = []byte("replaced")
	if bytes.Equal(inner.children[0], []byte("replaced")) {
		t.Fatal("slot reassignment should not affect original")
	}
}

func TestBug1_COWSafetyUnderInsertAndRemove(t *testing.T) {
	// Verify that actual tree operations (insert, remove) never corrupt
	// a saved version through the shared-slice Clone() mechanism.
	tree := NewMutableTreeMem()

	// Build a tree and save V1
	for i := 0; i < 100; i++ {
		tree.Set(fmt.Appendf(nil, "bug1_%04d", i), fmt.Appendf(nil, "v1_%04d", i))
	}
	tree.SaveVersion()
	v1Hash := tree.Hash()

	// Collect all V1 key-value pairs
	v1KV := make(map[string]string)
	tree.Iterate(func(k, v []byte) bool {
		v1KV[string(k)] = string(v)
		return false
	})

	// Mutate heavily in working tree — insertions, updates, removals
	for i := 0; i < 50; i++ {
		tree.Remove(fmt.Appendf(nil, "bug1_%04d", i))
	}
	for i := 100; i < 200; i++ {
		tree.Set(fmt.Appendf(nil, "bug1_%04d", i), []byte("v2_new"))
	}
	for i := 50; i < 80; i++ {
		tree.Set(fmt.Appendf(nil, "bug1_%04d", i), []byte("v2_updated"))
	}

	// Rollback should perfectly restore V1
	tree.Rollback()

	if !bytes.Equal(tree.Hash(), v1Hash) {
		t.Fatal("hash changed after heavy mutation + rollback")
	}

	// Verify every key-value pair
	afterKV := make(map[string]string)
	tree.Iterate(func(k, v []byte) bool {
		afterKV[string(k)] = string(v)
		return false
	})
	if len(afterKV) != len(v1KV) {
		t.Fatalf("key count: got %d, want %d", len(afterKV), len(v1KV))
	}
	for k, v := range v1KV {
		if afterKV[k] != v {
			t.Fatalf("key %q: got %q, want %q", k, afterKV[k], v)
		}
	}
}

// =========================================================================
// Bug #2: resolveValue Returns Hash as Value (Silent Wrong Data)
//
// When neither ndb nor memValues is set, resolveValue returns vh[:] — the
// 32-byte SHA256 hash itself — instead of an error. This means a caller
// gets garbage data silently.
// =========================================================================

func TestBug2_ResolveValueReturnsHashWhenNoResolver(t *testing.T) {
	// ImmutableTree with no valueResolver returns the hash bytes as "value"
	leaf := &LeafNode{miniTree: NewMiniMerkle()}
	leaf.keys[0] = []byte("key1")
	leaf.valueHashes[0] = sha256.Sum256([]byte("real_value"))
	leaf.numKeys = 1
	leaf.RebuildMiniMerkle()

	imm := NewImmutableTree(leaf, 1)
	// valueResolver is nil

	val, err := imm.Get([]byte("key1"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// BUG: val should be "real_value" or an error, but instead it's
	// the 32-byte SHA256 hash of "real_value"
	expectedHash := sha256.Sum256([]byte("real_value"))
	if !bytes.Equal(val, expectedHash[:]) {
		t.Fatal("expected resolveValue to return the hash bytes, but got something else")
	}
	if bytes.Equal(val, []byte("real_value")) {
		t.Fatal("value is correct — bug is fixed!")
	}

	// The returned "value" is always 32 bytes regardless of original value size
	if len(val) != 32 {
		t.Fatalf("expected 32 bytes (hash), got %d", len(val))
	}

	t.Logf("BUG CONFIRMED: Get returned %d-byte hash instead of actual value or error", len(val))
}

func TestBug2_MutableTreeResolveValueFallback(t *testing.T) {
	// MutableTree with ndb=nil and memValues=nil also returns the hash
	tree := &MutableTree{logger: NewNopLogger()}
	// Both ndb and memValues are nil

	leaf := &LeafNode{miniTree: NewMiniMerkle()}
	leaf.keys[0] = []byte("testkey")
	vh := sha256.Sum256([]byte("testvalue"))
	leaf.valueHashes[0] = vh
	leaf.numKeys = 1
	leaf.RebuildMiniMerkle()
	tree.root = leaf
	tree.size = 1

	val, err := tree.Get([]byte("testkey"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// BUG: returns the hash, not the value
	if bytes.Equal(val, []byte("testvalue")) {
		t.Fatal("value is correct — bug may be fixed")
	}
	if !bytes.Equal(val, vh[:]) {
		t.Fatalf("expected hash bytes as fallback, got %x", val)
	}
	t.Logf("BUG CONFIRMED: MutableTree.Get returned hash instead of value when resolvers are nil")
}

func TestBug2_GetValueByHashFallback(t *testing.T) {
	// GetValueByHash also falls back to returning hash bytes
	tree := &MutableTree{logger: NewNopLogger()}
	// Both ndb and memValues are nil

	vh := sha256.Sum256([]byte("some_value"))
	result, err := tree.GetValueByHash(vh)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Returns the hash itself instead of an error
	if !bytes.Equal(result, vh[:]) {
		t.Fatalf("expected hash fallback, got %x", result)
	}
	t.Logf("BUG CONFIRMED: GetValueByHash returned hash itself instead of error")
}

func TestBug2_ImmutableTreeIterateWithoutResolver(t *testing.T) {
	// Iterate without a valueResolver passes hash bytes as "values"
	leaf := &LeafNode{miniTree: NewMiniMerkle()}
	realValues := []string{"apple", "banana", "cherry"}
	for i, v := range realValues {
		leaf.keys[i] = []byte(fmt.Sprintf("k%d", i))
		leaf.valueHashes[i] = sha256.Sum256([]byte(v))
	}
	leaf.numKeys = 3
	leaf.RebuildMiniMerkle()

	imm := NewImmutableTree(leaf, 1)
	// No valueResolver set

	var gotValues [][]byte
	imm.Iterate(func(key, value []byte) bool {
		gotValues = append(gotValues, append([]byte(nil), value...))
		return false
	})

	// Every "value" should be 32 bytes (the hash) instead of the real value
	for i, gv := range gotValues {
		if len(gv) != 32 {
			t.Errorf("value[%d] length = %d, expected 32 (hash)", i, len(gv))
		}
		expectedHash := sha256.Sum256([]byte(realValues[i]))
		if !bytes.Equal(gv, expectedHash[:]) {
			t.Errorf("value[%d] is not the expected hash", i)
		}
	}
	t.Logf("BUG CONFIRMED: Iterate returned %d hash values instead of real values", len(gotValues))
}

// =========================================================================
// Bug #3: SaveValue Writes Directly to DB, Bypassing the Batch
//
// SaveValue uses ndb.db.Set() while SaveNode/SaveRoot use ndb.batch.Set().
// This breaks transactional atomicity: values persist even if SaveVersion
// fails or Rollback() is called.
// =========================================================================

func TestBug3_SaveValuePersistsBeforeCommit(t *testing.T) {
	// Values are written to DB immediately on Set(), before SaveVersion/Commit
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())

	value := []byte("persisted_early")
	vh := sha256.Sum256(value)

	tree.Set([]byte("key1"), value)

	// The value is already in the DB, even though SaveVersion hasn't been called
	valKey := make([]byte, 1+HashSize)
	valKey[0] = PrefixVal
	copy(valKey[1:], vh[:])

	data, err := db.Get(valKey)
	if err != nil {
		t.Fatalf("db.Get: %v", err)
	}
	if data == nil {
		t.Fatal("expected value to be in DB before SaveVersion — bug may be fixed")
	}
	if !bytes.Equal(data, value) {
		t.Fatalf("value in DB = %q, want %q", data, value)
	}
	t.Log("BUG CONFIRMED: value persists in DB before SaveVersion is called")
}

func TestBug3_RollbackLeavesOrphanedValues(t *testing.T) {
	// Set values, then Rollback — values remain in DB with no nodes referencing them
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())

	// Create an initial saved state
	tree.Set([]byte("base"), []byte("v0"))
	tree.SaveVersion()

	// Now set new values
	values := [][]byte{
		[]byte("orphan_value_1"),
		[]byte("orphan_value_2"),
		[]byte("orphan_value_3"),
	}
	for i, v := range values {
		tree.Set(fmt.Appendf(nil, "rollback_key_%d", i), v)
	}

	// Rollback — these values should be cleaned up
	tree.Rollback()

	// Check: values are STILL in the DB despite rollback
	orphanCount := 0
	for _, v := range values {
		vh := sha256.Sum256(v)
		valKey := make([]byte, 1+HashSize)
		valKey[0] = PrefixVal
		copy(valKey[1:], vh[:])

		data, _ := db.Get(valKey)
		if data != nil {
			orphanCount++
		}
	}

	if orphanCount == 0 {
		t.Fatal("all orphaned values were cleaned up — bug may be fixed")
	}
	t.Logf("BUG CONFIRMED: %d/%d values remain in DB after Rollback (orphaned forever)", orphanCount, len(values))
}

func TestBug3_FailedSaveVersionLeavesOrphanedValues(t *testing.T) {
	// Even if we never call SaveVersion, all Set values are in the DB
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())

	// Set many values without ever saving
	for i := 0; i < 100; i++ {
		tree.Set(
			fmt.Appendf(nil, "nosave_%04d", i),
			fmt.Appendf(nil, "value_%04d", i),
		)
	}

	// Count value entries in DB (prefix 'V')
	valueCount := countDBEntries(db, PrefixVal)

	// No SaveVersion was called, but values are in DB
	if valueCount == 0 {
		t.Fatal("no values found in DB — bug may be fixed (values now batched)")
	}
	t.Logf("BUG CONFIRMED: %d values written to DB without SaveVersion ever being called", valueCount)

	// Now count node entries — should be 0 since no SaveVersion
	nodeCount := countDBEntries(db, PrefixNode)
	rootCount := countDBEntries(db, PrefixRoot)
	if nodeCount != 0 || rootCount != 0 {
		t.Fatalf("expected 0 nodes and 0 roots before SaveVersion, got nodes=%d roots=%d",
			nodeCount, rootCount)
	}
	t.Logf("Nodes: %d, Roots: %d (correctly 0 — only values leak)", nodeCount, rootCount)
}

func TestBug3_ValueLeakGrowsUnbounded(t *testing.T) {
	// Demonstrate that repeated Set+Rollback cycles accumulate orphaned values
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())
	tree.Set([]byte("base"), []byte("v0"))
	tree.SaveVersion()

	for cycle := 0; cycle < 10; cycle++ {
		for i := 0; i < 50; i++ {
			tree.Set(
				fmt.Appendf(nil, "cycle_%d_%d", cycle, i),
				fmt.Appendf(nil, "val_%d_%d", cycle, i),
			)
		}
		tree.Rollback()
	}

	valueCount := countDBEntries(db, PrefixVal)
	// We saved 1 value ("v0") + 500 orphaned values (50 per cycle * 10 cycles)
	// Some may collide by hash, but most are unique
	if valueCount <= 10 {
		t.Fatalf("expected many orphaned values, got %d", valueCount)
	}
	t.Logf("BUG CONFIRMED: %d values in DB after 10 Set+Rollback cycles (only 1 should exist)",
		valueCount)
}

func countDBEntries(db *memdb.MemDB, prefix byte) int {
	count := 0
	start := []byte{prefix}
	end := []byte{prefix + 1}
	itr, _ := db.Iterator(start, end)
	defer itr.Close()
	for ; itr.Valid(); itr.Next() {
		count++
	}
	return count
}

// =========================================================================
// Bug #4: No Bounds Checking on numKeys During Deserialization
//
// ReadNode deserializes numKeys from the wire without validating it's within
// [0, B-1] for inner nodes or [0, B] for leaf nodes. Malformed data could
// cause out-of-bounds array access.
// =========================================================================

func TestBug4_InnerNodeDeserializeOverflowNumKeys(t *testing.T) {
	// Craft a serialized inner node with numKeys > B-1
	var buf bytes.Buffer
	buf.WriteByte(TypeInner)                 // type
	writeTestUvarint(&buf, uint64(B+5))      // numKeys = 37 (> B-1 = 31)
	writeTestVarint(&buf, 100)               // size
	writeTestUvarint(&buf, 2)                // height

	// Write B+5 keys (more than the array can hold)
	for i := 0; i < B+5; i++ {
		key := fmt.Appendf(nil, "key%03d", i)
		writeTestUvarint(&buf, uint64(len(key)))
		buf.Write(key)
	}

	// Write B+6 children (numKeys+1)
	for i := 0; i < B+6; i++ {
		nk := NodeKey{Version: int64(i + 1), Nonce: uint32(i + 1)}
		buf.Write(nk.GetKey())
	}

	// Write B+6 child hashes
	for i := 0; i < B+6; i++ {
		h := sha256.Sum256([]byte(fmt.Sprintf("child%d", i)))
		buf.Write(h[:])
	}

	// Attempt to deserialize — should fail gracefully or panic
	nk := &NodeKey{Version: 1, Nonce: 1}
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("BUG CONFIRMED: deserialization panicked with overflow numKeys: %v", r)
				return
			}
		}()

		node, err := ReadNode(nk, buf.Bytes())
		if err != nil {
			t.Logf("Deserialization returned error (acceptable): %v", err)
			return
		}

		// If we got here, the node was created with invalid numKeys
		if inner, ok := node.(*InnerNode); ok {
			if inner.numKeys > int16(B-1) {
				t.Logf("BUG CONFIRMED: created InnerNode with numKeys=%d (max should be %d)",
					inner.numKeys, B-1)
				// Try to access out-of-bounds — this would panic in production
				defer func() {
					if r := recover(); r != nil {
						t.Logf("  → Accessing children panicked: %v", r)
					}
				}()
				_ = inner.NumChildren() // numKeys+1 = 38, but array is [32]
			}
		}
	}()
}

func TestBug4_LeafNodeDeserializeOverflowNumKeys(t *testing.T) {
	// Craft a serialized leaf node with numKeys > B
	var buf bytes.Buffer
	buf.WriteByte(TypeLeaf)              // type
	writeTestUvarint(&buf, uint64(B+3))  // numKeys = 35 (> B = 32)

	// Write B+3 keys
	for i := 0; i < B+3; i++ {
		key := fmt.Appendf(nil, "lk%03d", i)
		writeTestUvarint(&buf, uint64(len(key)))
		buf.Write(key)
	}

	// Write B+3 value hashes
	for i := 0; i < B+3; i++ {
		h := sha256.Sum256([]byte(fmt.Sprintf("val%d", i)))
		buf.Write(h[:])
	}

	nk := &NodeKey{Version: 1, Nonce: 1}
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("BUG CONFIRMED: leaf deserialization panicked with overflow numKeys: %v", r)
				return
			}
		}()

		node, err := ReadNode(nk, buf.Bytes())
		if err != nil {
			t.Logf("Deserialization returned error (acceptable): %v", err)
			return
		}

		if leaf, ok := node.(*LeafNode); ok {
			if leaf.numKeys > int16(B) {
				t.Logf("BUG CONFIRMED: created LeafNode with numKeys=%d (max should be %d)",
					leaf.numKeys, B)
			}
		}
	}()
}

func TestBug4_InnerNodeDeserializeZeroNumKeys(t *testing.T) {
	// An inner node with 0 keys but 1 child — edge case that should
	// only exist at root during collapse, never in DB
	var buf bytes.Buffer
	buf.WriteByte(TypeInner)
	writeTestUvarint(&buf, 0) // numKeys = 0
	writeTestVarint(&buf, 5)  // size
	writeTestUvarint(&buf, 1) // height

	// 0 keys, 1 child (numKeys+1)
	nk := NodeKey{Version: 1, Nonce: 1}
	buf.Write(nk.GetKey())
	h := sha256.Sum256([]byte("onlychild"))
	buf.Write(h[:])

	nodeKey := &NodeKey{Version: 2, Nonce: 1}
	node, err := ReadNode(nodeKey, buf.Bytes())
	if err != nil {
		t.Fatalf("unexpected error for zero numKeys: %v", err)
	}
	inner := node.(*InnerNode)
	if inner.numKeys != 0 {
		t.Fatalf("numKeys = %d, want 0", inner.numKeys)
	}
	// This is technically valid but suspicious — no validation
	t.Logf("Zero numKeys inner node accepted without validation (NumChildren=%d)", inner.NumChildren())
}

func TestBug4_NegativeNumKeysViaOverflow(t *testing.T) {
	// numKeys is read as uvarint and cast to int16. A very large uvarint
	// value could overflow int16 and become negative.
	var buf bytes.Buffer
	buf.WriteByte(TypeLeaf)
	// Write a uvarint that, when cast to int16, becomes negative
	// int16 max = 32767. 32768 overflows to -32768
	writeTestUvarint(&buf, 32768)

	nk := &NodeKey{Version: 1, Nonce: 1}
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Panicked on negative numKeys overflow: %v", r)
				return
			}
		}()

		node, err := ReadNode(nk, buf.Bytes())
		if err != nil {
			// Error during read is acceptable — it will try to read 32768 keys
			// and fail (not enough data)
			t.Logf("Read error (acceptable, data exhaustion): %v", err)
			return
		}

		if leaf, ok := node.(*LeafNode); ok {
			if leaf.numKeys < 0 {
				t.Logf("BUG CONFIRMED: numKeys overflowed to %d (negative int16)", leaf.numKeys)
			}
		}
	}()
}

// helpers for manual serialization
func writeTestUvarint(buf *bytes.Buffer, v uint64) {
	var b [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(b[:], v)
	buf.Write(b[:n])
}

func writeTestVarint(buf *bytes.Buffer, v int64) {
	var b [binary.MaxVarintLen64]byte
	n := binary.PutVarint(b[:], v)
	buf.Write(b[:n])
}

// =========================================================================
// Bug #5: Exporter Goroutine Leak
//
// Export() launches a goroutine that blocks on channel send. If the caller
// doesn't call Close() (or drain the channel), the goroutine is leaked.
// There's no context cancellation mechanism.
// =========================================================================

func TestBug5_ExporterGoroutineLeakWithoutClose(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())

	// Insert enough data that the exporter goroutine will block
	for i := 0; i < 200; i++ {
		tree.Set(fmt.Appendf(nil, "leak%04d", i), fmt.Appendf(nil, "val%04d", i))
	}
	tree.SaveVersion()

	imm, err := tree.GetImmutable(1)
	if err != nil {
		t.Fatal(err)
	}

	before := runtime.NumGoroutine()

	// Launch 5 exporters and abandon them without Close()
	leakedExporters := make([]*Exporter, 5)
	for i := 0; i < 5; i++ {
		exp, err := imm.Export(tree.ndb)
		if err != nil {
			t.Fatal(err)
		}
		// Read a few items to start the goroutine, then abandon
		for j := 0; j < 3; j++ {
			exp.Next()
		}
		leakedExporters[i] = exp
	}

	// Give goroutines time to block
	runtime.Gosched()
	time.Sleep(50 * time.Millisecond)

	after := runtime.NumGoroutine()
	leaked := after - before
	if leaked >= 5 {
		t.Logf("BUG CONFIRMED: %d goroutines leaked by abandoned exporters", leaked)
	} else if leaked > 0 {
		t.Logf("Partial leak: %d goroutines leaked (expected 5)", leaked)
	} else {
		t.Log("No goroutine leak detected — bug may be fixed")
	}

	// Clean up to avoid polluting other tests
	for _, exp := range leakedExporters {
		exp.Close()
	}

	// Verify cleanup works
	time.Sleep(50 * time.Millisecond)
	runtime.GC()
	final := runtime.NumGoroutine()
	if final > before+1 {
		t.Logf("After Close(): still %d extra goroutines", final-before)
	}
}

func TestBug5_ExporterVersionReaderLeakWithoutClose(t *testing.T) {
	// Even if the goroutine eventually exits (channel has buffer),
	// the version reader count is only decremented in Close().
	// Without Close(), the version is protected from pruning forever.
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())

	for i := 0; i < 30; i++ {
		tree.Set(fmt.Appendf(nil, "vr%03d", i), []byte("val"))
	}
	tree.SaveVersion()
	tree.Set([]byte("extra"), []byte("v2"))
	tree.SaveVersion()

	imm, _ := tree.GetImmutable(1)
	exp, _ := imm.Export(tree.ndb)

	// Read a few items but don't Close
	for i := 0; i < 5; i++ {
		exp.Next()
	}

	// Try to prune v1 — should fail because version reader is still active
	err := tree.DeleteVersionsTo(1)
	if err == nil {
		t.Fatal("expected pruning to fail with active version reader, but it succeeded")
	}
	t.Logf("BUG CONFIRMED: abandoned exporter blocks pruning: %v", err)

	// Close to clean up
	exp.Close()

	// Now pruning should work
	err = tree.DeleteVersionsTo(1)
	if err != nil {
		t.Fatalf("prune after Close failed: %v", err)
	}
}

// =========================================================================
// Bug #6: Pruning Deletes Shared Nodes After B+ Tree Splits
//
// walkAndPrune only checks the corresponding newNode's immediate children
// to determine if old children are shared. After an inner node split,
// children that moved to a sibling in the new version are incorrectly
// treated as orphaned and deleted.
//
// These tests trigger the bug with deterministic seeds. The key insight:
// random keys cause unpredictable splits, and splits move children between
// parents across versions. Single-version pruning is sufficient.
// =========================================================================

func TestBug6_SingleVersionPruneCorruptsTree(t *testing.T) {
	// This demonstrates the bug with per-block pruning — the exact
	// pattern a Cosmos chain would use. No bulk pruning needed.
	// Uses 500 mutations per block (matching the benchmark conditions)
	// and 32-byte random keys to maximize inner node splits.
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 10000, NewNopLogger())

	rng := rand.New(rand.NewSource(42))
	keepVersions := int64(20)
	panicCount := 0
	pruneErrCount := 0
	totalBlocks := 400

	for block := 0; block < totalBlocks; block++ {
		// 500 random 32-byte key mutations per block
		for i := 0; i < 500; i++ {
			key := make([]byte, 32)
			rng.Read(key)
			val := make([]byte, 100)
			rng.Read(val)
			tree.Set(key, val)
		}

		_, version, err := tree.SaveVersion()
		if err != nil {
			t.Fatalf("SaveVersion at block %d: %v", block, err)
		}

		// Prune exactly one version — the standard Cosmos pattern
		if version > keepVersions+1 {
			func() {
				defer func() {
					if r := recover(); r != nil {
						panicCount++
						if panicCount <= 3 {
							t.Logf("PANIC #%d at version %d: %v", panicCount, version, r)
						}
					}
				}()
				pruneTarget := version - keepVersions - 1
				if err := tree.DeleteVersionsTo(pruneTarget); err != nil {
					pruneErrCount++
					if pruneErrCount <= 3 {
						t.Logf("PRUNE ERROR #%d at version %d: %v",
							pruneErrCount, version, err)
					}
				}
			}()
		}
	}

	t.Logf("RESULT: %d blocks (500 ops each), %d panics, %d prune errors",
		totalBlocks, panicCount, pruneErrCount)
	if panicCount > 0 || pruneErrCount > 0 {
		t.Logf("BUG CONFIRMED: %d panics + %d errors during block-by-block pruning",
			panicCount, pruneErrCount)
	}
}

func TestBug6_PruneCorruptsNewerVersions(t *testing.T) {
	// Verify that after pruning corruption, the LATEST version is
	// also corrupted (not just the pruned one).
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 5000, NewNopLogger())

	rng := rand.New(rand.NewSource(42))

	// Build state
	for i := 0; i < 3000; i++ {
		key := make([]byte, 16)
		rng.Read(key)
		tree.Set(key, []byte("init"))
	}
	tree.SaveVersion()

	// Create many versions
	for v := 0; v < 300; v++ {
		for i := 0; i < 50; i++ {
			key := make([]byte, 16)
			rng.Read(key)
			tree.Set(key, fmt.Appendf(nil, "v%d", v+2))
		}
		tree.SaveVersion()
	}

	latestVersion := tree.Version()
	latestHash := tree.Hash()
	latestSize := tree.Size()
	t.Logf("Before prune: version=%d size=%d height=%d", latestVersion, latestSize, tree.Height())

	// Prune — may corrupt newer versions
	func() {
		defer func() {
			recover() // swallow panic if any
		}()
		tree.DeleteVersionsTo(latestVersion - 20)
	}()

	// Try to reload the latest version from a cold cache
	tree2 := NewMutableTreeWithDB(db, 0, NewNopLogger()) // cache=0 forces DB reads
	_, err := tree2.LoadVersion(latestVersion)
	if err != nil {
		t.Logf("BUG CONFIRMED: latest version %d is corrupted after pruning: %v",
			latestVersion, err)
		return
	}

	// Even if load succeeded, verify integrity by iterating all keys
	iterCount := int64(0)
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("BUG CONFIRMED: iteration of latest version panics: %v", r)
			}
		}()
		tree2.Iterate(func(key, value []byte) bool {
			iterCount++
			return false
		})
	}()

	// Check if data was lost
	if iterCount != latestSize {
		t.Logf("BUG CONFIRMED: latest version lost data: iterated %d, expected %d",
			iterCount, latestSize)
	}

	hash2 := tree2.Hash()
	if !bytes.Equal(hash2, latestHash) {
		t.Logf("BUG CONFIRMED: latest version hash changed after pruning")
	}

	if iterCount == latestSize && bytes.Equal(hash2, latestHash) {
		t.Log("Latest version appears intact (bug may not manifest at this scale/seed)")
	}
}

func TestBug6_PruneBricksNodeOnRestart(t *testing.T) {
	// Simulate the most dangerous production scenario:
	// 1. Run blocks with pruning
	// 2. Restart (cold cache)
	// 3. Node can't load latest version
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 10000, NewNopLogger())

	rng := rand.New(rand.NewSource(42))

	// Phase 1: run many blocks with pruning (simulates a running node)
	var lastGoodVersion int64
	for block := 0; block < 400; block++ {
		for i := 0; i < 100; i++ {
			key := make([]byte, 16)
			rng.Read(key)
			val := make([]byte, 40)
			rng.Read(val)
			tree.Set(key, val)
		}
		_, version, err := tree.SaveVersion()
		if err != nil {
			t.Fatalf("SaveVersion: %v", err)
		}
		lastGoodVersion = version

		if version > 20 {
			func() {
				defer func() { recover() }()
				tree.DeleteVersionsTo(version - 20)
			}()
		}
	}

	t.Logf("Phase 1 complete: version=%d size=%d height=%d",
		lastGoodVersion, tree.Size(), tree.Height())

	// Phase 2: simulate restart — create fresh tree with NO cache
	tree2 := NewMutableTreeWithDB(db, 0, NewNopLogger())

	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("BUG CONFIRMED: node bricked on restart — Load panicked: %v", r)
				return
			}
		}()

		_, err := tree2.Load()
		if err != nil {
			t.Logf("BUG CONFIRMED: node bricked on restart — Load failed: %v", err)
			return
		}

		// Try to use the tree
		key := make([]byte, 16)
		rng.Read(key)
		_, err = tree2.Get(key)
		if err != nil {
			t.Logf("BUG CONFIRMED: Get failed after restart: %v", err)
			return
		}

		// Try iteration
		count := int64(0)
		tree2.Iterate(func(k, v []byte) bool {
			count++
			return false
		})
		if count != tree2.Size() {
			t.Logf("BUG CONFIRMED: iteration count %d != size %d after restart",
				count, tree2.Size())
		} else {
			t.Logf("Restart succeeded — bug may not manifest at this scale (version=%d, size=%d)",
				tree2.Version(), tree2.Size())
		}
	}()
}

