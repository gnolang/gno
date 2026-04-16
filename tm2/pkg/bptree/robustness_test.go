package bptree

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"testing"

	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// -----------------------------------------------------------------------------
// Finding #18: InnerNode.Serialize asserts child-ref validity.
// -----------------------------------------------------------------------------

// TestInnerNodeSerialize_RejectsNilChildRef verifies that serializing an inner
// node with a nil (or short) child reference fails fast rather than silently
// desynchronising the stream. Before the fix, w.Write(nil) would emit 0 bytes
// and the reader would consume bytes from the next field as the NodeKey.
func TestInnerNodeSerialize_RejectsNilChildRef(t *testing.T) {
	inner := &InnerNode{
		nodeKey:  &NodeKey{Version: 1, Nonce: 1},
		numKeys:  1,
		height:   1,
		miniTree: NewMiniMerkle(),
	}
	inner.keys[0] = []byte("k")
	inner.childSizes[0] = 1
	inner.childSizes[1] = 1
	// Leave inner.children[0] and inner.children[1] as nil — simulates
	// an unsaved child slipping into SaveNode.
	var buf bytes.Buffer
	err := inner.Serialize(&buf)
	if err == nil {
		t.Fatalf("expected error from Serialize on nil child ref, got nil")
	}
	if !strings.Contains(err.Error(), "child[0]") {
		t.Fatalf("error should mention offending child index; got %q", err.Error())
	}
}

// TestInnerNodeSerialize_RejectsShortChildRef verifies the assertion fires on
// a non-NodeKeySize ref (length mismatch). The read path expects exactly
// NodeKeySize bytes; anything else desynchronises the stream.
func TestInnerNodeSerialize_RejectsShortChildRef(t *testing.T) {
	inner := &InnerNode{
		nodeKey:  &NodeKey{Version: 1, Nonce: 1},
		numKeys:  0,
		height:   1,
		miniTree: NewMiniMerkle(),
	}
	// numKeys=0 means NumChildren()=1, one child ref required.
	inner.childSizes[0] = 1
	inner.children[0] = []byte{0xAA, 0xBB} // 2 bytes, not NodeKeySize
	var buf bytes.Buffer
	err := inner.Serialize(&buf)
	if err == nil {
		t.Fatalf("expected error from Serialize on short child ref, got nil")
	}
	if !strings.Contains(err.Error(), "len 2") {
		t.Fatalf("error should report observed length; got %q", err.Error())
	}
}

// -----------------------------------------------------------------------------
// Finding #23: Bounds-check numKeys (uint64 → int16 cast) and valueKey length.
// -----------------------------------------------------------------------------

// TestReadInnerNode_RejectsOutOfRangeNumKeys verifies that a crafted
// InnerNode payload whose numKeys field is > B-1 is rejected BEFORE casting
// to int16. A value like 0xFFFF would wrap to -1 as int16 and slip past the
// negative check before the fix.
func TestReadInnerNode_RejectsOutOfRangeNumKeys(t *testing.T) {
	// Build a minimal payload: type(InnerNode) + numKeys(0xFFFF uvarint)
	var buf bytes.Buffer
	buf.WriteByte(TypeInner)
	var vb [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(vb[:], 0xFFFF)
	buf.Write(vb[:n])

	_, err := ReadNode(&NodeKey{Version: 1, Nonce: 1}, buf.Bytes())
	if err == nil {
		t.Fatalf("expected error from ReadNode on out-of-range numKeys")
	}
	if !strings.Contains(err.Error(), "out of range") {
		t.Fatalf("error should describe out-of-range; got %q", err.Error())
	}
}

// TestReadLeafNode_RejectsOutOfRangeNumKeys mirrors the inner case for
// LeafNode. The leaf bound is B (vs B-1 for inner).
func TestReadLeafNode_RejectsOutOfRangeNumKeys(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteByte(TypeLeaf)
	var vb [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(vb[:], 0xFFFF)
	buf.Write(vb[:n])

	_, err := ReadNode(&NodeKey{Version: 1, Nonce: 1}, buf.Bytes())
	if err == nil {
		t.Fatalf("expected error from ReadNode on out-of-range leaf numKeys")
	}
	if !strings.Contains(err.Error(), "out of range") {
		t.Fatalf("error should describe out-of-range; got %q", err.Error())
	}
}

// TestOrphanValueKey_ShortVKNoPanic verifies that orphanValueKey does not
// slice-bounds-panic if given a short/corrupt valueKey. Before the Finding
// #23 fix, `vk[:8]` on a 4-byte slice would panic.
func TestOrphanValueKey_ShortVKNoPanic(t *testing.T) {
	tree := NewMutableTreeMem()
	// Should simply log and return; must not panic.
	tree.orphanValueKey([]byte{0x01, 0x02, 0x03, 0x04})
}

// -----------------------------------------------------------------------------
// Finding #12: Unsupported APIs return ErrUnsupported instead of panicking.
// -----------------------------------------------------------------------------

// TestUnsupportedAPIs_ReturnErrUnsupported verifies both LoadVersionForOverwriting
// and DeleteVersionsFrom surface the typed sentinel so callers can probe for
// IAVL compatibility gaps without crashing the process.
func TestUnsupportedAPIs_ReturnErrUnsupported(t *testing.T) {
	tree := NewMutableTreeWithDB(memdb.NewMemDB(), 0, NewNopLogger())
	tree.Set([]byte("k"), []byte("v"))
	tree.SaveVersion()

	if err := tree.LoadVersionForOverwriting(1); !errors.Is(err, ErrUnsupported) {
		t.Fatalf("LoadVersionForOverwriting: got %v, want ErrUnsupported", err)
	}
	if err := tree.DeleteVersionsFrom(1); !errors.Is(err, ErrUnsupported) {
		t.Fatalf("DeleteVersionsFrom: got %v, want ErrUnsupported", err)
	}
}

// -----------------------------------------------------------------------------
// Finding #28: Set is atomic with value save.
// -----------------------------------------------------------------------------

// failingSaveValueDB is a DB wrapper that fails writes whose key begins with
// the value-prefix byte. Every other operation passes through to the inner
// memdb so the tree can load, read, and commit nodes normally.
type failingSaveValueDB struct {
	dbm.DB
	failValues bool
}

func newFailingSaveValueDB() *failingSaveValueDB {
	return &failingSaveValueDB{DB: memdb.NewMemDB()}
}

func (d *failingSaveValueDB) Set(key, value []byte) error {
	if d.failValues && len(key) > 0 && key[0] == PrefixVal {
		return fmt.Errorf("injected SaveValue failure")
	}
	return d.DB.Set(key, value)
}

func (d *failingSaveValueDB) SetSync(key, value []byte) error {
	if d.failValues && len(key) > 0 && key[0] == PrefixVal {
		return fmt.Errorf("injected SaveValue failure (sync)")
	}
	return d.DB.SetSync(key, value)
}

// TestSet_AtomicWithSaveValue_FirstInsert verifies that when SaveValue fails
// on the very first Set of a fresh tree, the tree remains empty (root == nil,
// size == 0). Before Finding #28, the leaf was stitched up first and then
// SaveValue was attempted, so the tree would reference a never-persisted vk.
func TestSet_AtomicWithSaveValue_FirstInsert(t *testing.T) {
	db := newFailingSaveValueDB()
	tree := NewMutableTreeWithDB(db, 100, NewNopLogger())

	db.failValues = true
	_, err := tree.Set([]byte("k"), []byte("v"))
	if err == nil {
		t.Fatalf("expected Set to return SaveValue error")
	}

	if !tree.IsEmpty() {
		t.Fatalf("tree should still be empty after failed Set; size=%d", tree.Size())
	}
	if tree.root != nil {
		t.Fatalf("tree.root should be nil after failed Set")
	}
}

// TestSet_AtomicWithSaveValue_Update verifies that when SaveValue fails on an
// update, the tree does not contain a leaf pointing at the new (never-written)
// vk. After the fix, the tree still has the OLD value for that key.
func TestSet_AtomicWithSaveValue_Update(t *testing.T) {
	db := newFailingSaveValueDB()
	tree := NewMutableTreeWithDB(db, 100, NewNopLogger())

	// Seed a successful Set + SaveVersion so we're updating, not inserting.
	if _, err := tree.Set([]byte("k"), []byte("v1")); err != nil {
		t.Fatalf("seed Set: %v", err)
	}
	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatalf("seed SaveVersion: %v", err)
	}

	// Now inject a failure and attempt an update.
	db.failValues = true
	_, err := tree.Set([]byte("k"), []byte("v2"))
	if err == nil {
		t.Fatalf("expected Set to return SaveValue error")
	}

	// The old value must still be readable via the current root — the tree
	// was NOT mutated to reference the new (never-saved) vk.
	db.failValues = false
	got, err := tree.Get([]byte("k"))
	if err != nil {
		t.Fatalf("Get after failed Set: %v", err)
	}
	if !bytes.Equal(got, []byte("v1")) {
		t.Fatalf("Get after failed update: got %q, want %q", got, "v1")
	}
}

// -----------------------------------------------------------------------------
// Findings #25 / #31: DeleteValueDirect errors are logged, not ignored.
// -----------------------------------------------------------------------------

// countingLogger captures the Error() calls made by the tree under test so
// we can assert Rollback logged a diagnostic rather than silently discarding
// DeleteValueDirect failures.
type countingLogger struct{ errors int }

func (l *countingLogger) Debug(_ string, _ ...any) {}
func (l *countingLogger) Info(_ string, _ ...any)  {}
func (l *countingLogger) Warn(_ string, _ ...any)  {}
func (l *countingLogger) Error(_ string, _ ...any) { l.errors++ }

// failingDeleteDB wraps a memdb and fails Delete() for keys whose first byte
// is PrefixVal. Used to exercise the error path in Rollback / orphanValueKey.
type failingDeleteDB struct {
	dbm.DB
	failDeletes bool
}

func newFailingDeleteDB() *failingDeleteDB {
	return &failingDeleteDB{DB: memdb.NewMemDB()}
}

func (d *failingDeleteDB) Delete(key []byte) error {
	if d.failDeletes && len(key) > 0 && key[0] == PrefixVal {
		return fmt.Errorf("injected Delete failure")
	}
	return d.DB.Delete(key)
}

func (d *failingDeleteDB) DeleteSync(key []byte) error {
	if d.failDeletes && len(key) > 0 && key[0] == PrefixVal {
		return fmt.Errorf("injected DeleteSync failure")
	}
	return d.DB.DeleteSync(key)
}

// TestRollback_LogsDeleteValueErrors verifies that Rollback logs a diagnostic
// when DeleteValueDirect fails — a space leak is still observable, and the
// tree's in-memory state is restored. Before the fix, Rollback silently
// dropped every error.
func TestRollback_LogsDeleteValueErrors(t *testing.T) {
	db := newFailingDeleteDB()
	logger := &countingLogger{}
	tree := NewMutableTreeWithDB(db, 100, logger)

	if _, err := tree.Set([]byte("k1"), []byte("v1")); err != nil {
		t.Fatalf("seed Set: %v", err)
	}

	db.failDeletes = true
	tree.Rollback()

	if logger.errors == 0 {
		t.Fatalf("Rollback should log Delete errors; got 0 Error calls")
	}
	// Tree state must still be restored to lastSaved (nil here — no SaveVersion).
	if tree.root != nil {
		t.Fatalf("Rollback did not restore root; got %v", tree.root)
	}
}
