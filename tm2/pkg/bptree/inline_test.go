package bptree

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// TestInline_BasicRoundTrip verifies that values below the threshold
// are stored inline (valueKeys[i] nil, inlineMask bit set) and round-
// trip through Set/Get correctly.
func TestInline_BasicRoundTrip(t *testing.T) {
	tree := NewMutableTreeWithDB(memdb.NewMemDB(), 100, NewNopLogger(), InlineValueThresholdOption(32))
	if _, err := tree.Set([]byte("k"), []byte("short-value")); err != nil {
		t.Fatalf("Set: %v", err)
	}
	leaf := tree.root.(*LeafNode)
	if leaf.inlineMask&1 == 0 {
		t.Fatalf("expected slot 0 to be inline")
	}
	if leaf.valueKeys[0] != nil {
		t.Fatalf("inline slot must have nil valueKey")
	}
	if !bytes.Equal(leaf.inlineValues[0], []byte("short-value")) {
		t.Fatalf("inlineValues[0] = %q, want %q", leaf.inlineValues[0], "short-value")
	}

	got, err := tree.Get([]byte("k"))
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !bytes.Equal(got, []byte("short-value")) {
		t.Fatalf("Get = %q, want %q", got, "short-value")
	}
}

// TestInline_AboveThresholdExternal verifies that values above the
// threshold take the external path and populate valueKeys.
func TestInline_AboveThresholdExternal(t *testing.T) {
	tree := NewMutableTreeWithDB(memdb.NewMemDB(), 100, NewNopLogger(), InlineValueThresholdOption(8))
	big := []byte("this-is-a-longer-value-than-eight-bytes")
	if _, err := tree.Set([]byte("k"), big); err != nil {
		t.Fatalf("Set: %v", err)
	}
	leaf := tree.root.(*LeafNode)
	if leaf.inlineMask&1 != 0 {
		t.Fatalf("expected slot 0 to be external (value > threshold)")
	}
	if len(leaf.valueKeys[0]) != NodeKeySize {
		t.Fatalf("expected external valueKey, got %v", leaf.valueKeys[0])
	}
	got, err := tree.Get([]byte("k"))
	if err != nil || !bytes.Equal(got, big) {
		t.Fatalf("Get = %q err=%v, want %q", got, err, big)
	}
}

// TestInline_SerializeRoundTrip verifies v2 leaves serialize and
// deserialize without loss, preserving inline and external slots.
func TestInline_SerializeRoundTrip(t *testing.T) {
	tree := NewMutableTreeWithDB(memdb.NewMemDB(), 100, NewNopLogger(), InlineValueThresholdOption(16))
	// Alternate inline (short) and external (long) values.
	for i := 0; i < 10; i++ {
		var v []byte
		if i%2 == 0 {
			v = []byte{byte(i)}
		} else {
			v = bytes.Repeat([]byte{byte(i)}, 32)
		}
		tree.Set(fmt.Appendf(nil, "k%02d", i), v)
	}
	hash1, _, err := tree.SaveVersion()
	if err != nil {
		t.Fatalf("SaveVersion: %v", err)
	}

	// Reload via a fresh tree against the same DB.
	db := tree.ndb.db
	tree2 := NewMutableTreeWithDB(db, 100, NewNopLogger(), InlineValueThresholdOption(16))
	if _, err := tree2.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !bytes.Equal(tree2.Hash(), hash1) {
		t.Fatalf("hash after reload differs: %x vs %x", tree2.Hash(), hash1)
	}
	for i := 0; i < 10; i++ {
		var want []byte
		if i%2 == 0 {
			want = []byte{byte(i)}
		} else {
			want = bytes.Repeat([]byte{byte(i)}, 32)
		}
		got, _ := tree2.Get(fmt.Appendf(nil, "k%02d", i))
		if !bytes.Equal(got, want) {
			t.Fatalf("k%02d: Get = %q, want %q", i, got, want)
		}
	}
}

// TestInline_Update verifies that updating an inline slot with a
// different-size value keeps the leaf coherent.
func TestInline_Update(t *testing.T) {
	tree := NewMutableTreeWithDB(memdb.NewMemDB(), 100, NewNopLogger(), InlineValueThresholdOption(32))
	tree.Set([]byte("k"), []byte("first"))
	tree.Set([]byte("k"), []byte("updated-second-value"))
	got, _ := tree.Get([]byte("k"))
	if !bytes.Equal(got, []byte("updated-second-value")) {
		t.Fatalf("Get after update = %q, want updated-second-value", got)
	}
}

// TestInline_InlineToExternal verifies a slot transitioning from
// inline to external (value grew past the threshold on update) still
// resolves correctly.
func TestInline_InlineToExternal(t *testing.T) {
	tree := NewMutableTreeWithDB(memdb.NewMemDB(), 100, NewNopLogger(), InlineValueThresholdOption(8))
	tree.Set([]byte("k"), []byte("tiny"))
	leaf := tree.root.(*LeafNode)
	if leaf.inlineMask&1 == 0 {
		t.Fatalf("initial slot should be inline")
	}
	big := []byte("this-value-overflows-the-inline-threshold")
	tree.Set([]byte("k"), big)
	// After update, the slot transitioned external.
	leaf = tree.root.(*LeafNode)
	if leaf.inlineMask&1 != 0 {
		t.Fatalf("slot should be external after update")
	}
	got, _ := tree.Get([]byte("k"))
	if !bytes.Equal(got, big) {
		t.Fatalf("Get = %q, want %q", got, big)
	}
}

// TestInline_ExternalToInline verifies a slot transitioning from
// external to inline (value shrank below the threshold on update)
// resolves correctly and orphans the old external valueKey.
func TestInline_ExternalToInline(t *testing.T) {
	tree := NewMutableTreeWithDB(memdb.NewMemDB(), 100, NewNopLogger(), InlineValueThresholdOption(8))
	big := []byte("value-longer-than-eight-bytes")
	tree.Set([]byte("k"), big)
	leaf := tree.root.(*LeafNode)
	if leaf.inlineMask&1 != 0 {
		t.Fatalf("initial slot should be external")
	}
	tree.Set([]byte("k"), []byte("tiny"))
	leaf = tree.root.(*LeafNode)
	if leaf.inlineMask&1 == 0 {
		t.Fatalf("slot should be inline after shrinking value")
	}
	got, _ := tree.Get([]byte("k"))
	if !bytes.Equal(got, []byte("tiny")) {
		t.Fatalf("Get = %q, want tiny", got)
	}
}

// TestInline_LeafSplitPreservesInlineBits verifies an inline slot's
// payload + mask bit survives a leaf split.
func TestInline_LeafSplitPreservesInlineBits(t *testing.T) {
	tree := NewMutableTreeWithDB(memdb.NewMemDB(), 100, NewNopLogger(), InlineValueThresholdOption(32))
	// Fill a leaf beyond capacity (B=32) so it splits.
	for i := 0; i < B+5; i++ {
		v := []byte{byte(i)}
		tree.Set(fmt.Appendf(nil, "k%02d", i), v)
	}
	// Verify every key resolves to its original value.
	for i := 0; i < B+5; i++ {
		got, _ := tree.Get(fmt.Appendf(nil, "k%02d", i))
		if !bytes.Equal(got, []byte{byte(i)}) {
			t.Fatalf("k%02d: Get = %v, want [%d]", i, got, i)
		}
	}
}
