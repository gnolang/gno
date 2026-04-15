package bptree

import (
	"bytes"
	"fmt"
	"sort"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

func TestIterator_AscendingFull(t *testing.T) {
	tree := NewMutableTreeMem()
	n := 100
	for i := 0; i < n; i++ {
		tree.Set(fmt.Appendf(nil, "it%04d", i), []byte("v"))
	}

	itr, _ := tree.Iterator(nil, nil, true)
	defer itr.Close()

	var keys []string
	for itr.Valid() {
		keys = append(keys, string(itr.Key()))
		itr.Next()
	}
	if len(keys) != n {
		t.Fatalf("got %d keys, want %d", len(keys), n)
	}
	if !sort.StringsAreSorted(keys) {
		t.Fatalf("keys not sorted")
	}
}

func TestIterator_DescendingFull(t *testing.T) {
	tree := NewMutableTreeMem()
	n := 100
	for i := 0; i < n; i++ {
		tree.Set(fmt.Appendf(nil, "it%04d", i), []byte("v"))
	}

	itr, _ := tree.Iterator(nil, nil, false)
	defer itr.Close()

	var keys []string
	for itr.Valid() {
		keys = append(keys, string(itr.Key()))
		itr.Next()
	}
	if len(keys) != n {
		t.Fatalf("got %d keys, want %d", len(keys), n)
	}
	// Should be reverse sorted
	for i := 1; i < len(keys); i++ {
		if keys[i] >= keys[i-1] {
			t.Fatalf("keys not reverse sorted at %d: %s >= %s", i, keys[i], keys[i-1])
		}
	}
}

func TestIterator_Range(t *testing.T) {
	tree := NewMutableTreeMem()
	for i := 0; i < 100; i++ {
		tree.Set(fmt.Appendf(nil, "r%04d", i), []byte("v"))
	}

	// [r0020, r0030) — should get 10 keys
	start := []byte("r0020")
	end := []byte("r0030")
	itr, _ := tree.Iterator(start, end, true)
	defer itr.Close()

	var keys []string
	for itr.Valid() {
		keys = append(keys, string(itr.Key()))
		itr.Next()
	}
	if len(keys) != 10 {
		t.Fatalf("range got %d keys, want 10: %v", len(keys), keys)
	}
	if keys[0] != "r0020" || keys[9] != "r0029" {
		t.Fatalf("range bounds: first=%s last=%s", keys[0], keys[9])
	}
}

func TestIterator_RangeDescending(t *testing.T) {
	tree := NewMutableTreeMem()
	for i := 0; i < 100; i++ {
		tree.Set(fmt.Appendf(nil, "r%04d", i), []byte("v"))
	}

	start := []byte("r0020")
	end := []byte("r0030")
	itr, _ := tree.Iterator(start, end, false)
	defer itr.Close()

	var keys []string
	for itr.Valid() {
		keys = append(keys, string(itr.Key()))
		itr.Next()
	}
	if len(keys) != 10 {
		t.Fatalf("desc range got %d keys, want 10: %v", len(keys), keys)
	}
	if keys[0] != "r0029" || keys[9] != "r0020" {
		t.Fatalf("desc range: first=%s last=%s", keys[0], keys[9])
	}
}

func TestIterator_EmptyTree(t *testing.T) {
	tree := NewMutableTreeMem()
	itr, _ := tree.Iterator(nil, nil, true)
	defer itr.Close()
	if itr.Valid() {
		t.Fatalf("empty tree iterator should be invalid")
	}
}

func TestIterator_EmptyRange(t *testing.T) {
	tree := NewMutableTreeMem()
	for i := 0; i < 50; i++ {
		tree.Set(fmt.Appendf(nil, "e%04d", i), []byte("v"))
	}

	// Range that matches nothing
	itr, _ := tree.Iterator([]byte("zzz"), []byte("zzzz"), true)
	defer itr.Close()
	if itr.Valid() {
		t.Fatalf("empty range should produce invalid iterator")
	}
}

func TestIterator_SingleElement(t *testing.T) {
	tree := NewMutableTreeMem()
	tree.Set([]byte("only"), []byte("one"))

	itr, _ := tree.Iterator(nil, nil, true)
	defer itr.Close()
	if !itr.Valid() {
		t.Fatalf("should be valid")
	}
	if string(itr.Key()) != "only" {
		t.Fatalf("key = %s", itr.Key())
	}
	itr.Next()
	if itr.Valid() {
		t.Fatalf("should be invalid after single element")
	}
}

func TestIterator_StartOnly(t *testing.T) {
	tree := NewMutableTreeMem()
	for i := 0; i < 50; i++ {
		tree.Set(fmt.Appendf(nil, "s%04d", i), []byte("v"))
	}

	// Start from s0040, no end
	itr, _ := tree.Iterator([]byte("s0040"), nil, true)
	defer itr.Close()

	var keys []string
	for itr.Valid() {
		keys = append(keys, string(itr.Key()))
		itr.Next()
	}
	if len(keys) != 10 {
		t.Fatalf("start-only got %d keys, want 10", len(keys))
	}
}

func TestIterator_EndOnly(t *testing.T) {
	tree := NewMutableTreeMem()
	for i := 0; i < 50; i++ {
		tree.Set(fmt.Appendf(nil, "s%04d", i), []byte("v"))
	}

	// End at s0010 (exclusive)
	itr, _ := tree.Iterator(nil, []byte("s0010"), true)
	defer itr.Close()

	var keys []string
	for itr.Valid() {
		keys = append(keys, string(itr.Key()))
		itr.Next()
	}
	if len(keys) != 10 {
		t.Fatalf("end-only got %d keys, want 10", len(keys))
	}
}

func TestIterator_Domain(t *testing.T) {
	tree := NewMutableTreeMem()
	tree.Set([]byte("a"), []byte("1"))

	start := []byte("a")
	end := []byte("z")
	itr, _ := tree.Iterator(start, end, true)
	defer itr.Close()

	s, e := itr.Domain()
	if !bytes.Equal(s, start) || !bytes.Equal(e, end) {
		t.Fatalf("Domain: (%q, %q)", s, e)
	}
}

func TestIterator_CloseIdempotent(t *testing.T) {
	tree := NewMutableTreeMem()
	tree.Set([]byte("a"), []byte("1"))

	itr, _ := tree.Iterator(nil, nil, true)
	itr.Close()
	itr.Close() // should not panic
	if itr.Valid() {
		t.Fatalf("closed iterator should be invalid")
	}
}

func TestIterator_LargeTree(t *testing.T) {
	tree := NewMutableTreeMem()
	n := 1000
	for i := 0; i < n; i++ {
		tree.Set(fmt.Appendf(nil, "lt%06d", i), []byte("v"))
	}

	// Ascending full scan
	itr, _ := tree.Iterator(nil, nil, true)
	count := 0
	var prev string
	for itr.Valid() {
		k := string(itr.Key())
		if k <= prev && prev != "" {
			t.Fatalf("ascending order broken at %d: %s <= %s", count, k, prev)
		}
		prev = k
		count++
		itr.Next()
	}
	itr.Close()
	if count != n {
		t.Fatalf("ascending count = %d, want %d", count, n)
	}

	// Descending full scan
	itr, _ = tree.Iterator(nil, nil, false)
	count = 0
	prev = ""
	for itr.Valid() {
		k := string(itr.Key())
		if k >= prev && prev != "" {
			t.Fatalf("descending order broken at %d: %s >= %s", count, k, prev)
		}
		prev = k
		count++
		itr.Next()
	}
	itr.Close()
	if count != n {
		t.Fatalf("descending count = %d, want %d", count, n)
	}
}

func TestIterator_DBBacked_ReturnsValues(t *testing.T) {
	db := memdb.NewMemDB()
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())
	tree.Set([]byte("a"), []byte("alpha"))
	tree.Set([]byte("b"), []byte("beta"))
	tree.Set([]byte("c"), []byte("gamma"))
	tree.SaveVersion()

	// Reload
	tree2 := NewMutableTreeWithDB(db, 1000, NewNopLogger())
	tree2.Load()

	itr, _ := tree2.Iterator(nil, nil, true)
	defer itr.Close()

	expected := map[string]string{"a": "alpha", "b": "beta", "c": "gamma"}
	count := 0
	for itr.Valid() {
		k := string(itr.Key())
		v := string(itr.Value())
		if expected[k] != v {
			t.Fatalf("key %s: got %q, want %q", k, v, expected[k])
		}
		count++
		itr.Next()
	}
	if count != 3 {
		t.Fatalf("iterator count = %d, want 3", count)
	}
}

func TestIterator_ImmutableTree(t *testing.T) {
	tree := NewMutableTreeMem()
	for i := 0; i < 50; i++ {
		tree.Set(fmt.Appendf(nil, "im%04d", i), []byte("v"))
	}
	imm := tree.Snapshot(1)

	itr, _ := imm.Iterator(nil, nil, true)
	defer itr.Close()

	count := 0
	for itr.Valid() {
		count++
		itr.Next()
	}
	if count != 50 {
		t.Fatalf("immutable iterator count = %d, want 50", count)
	}
}

func TestIterator_IterateRange(t *testing.T) {
	tree := NewMutableTreeMem()
	for i := 0; i < 100; i++ {
		tree.Set(fmt.Appendf(nil, "ir%04d", i), []byte("v"))
	}

	var keys []string
	tree.IterateRange([]byte("ir0010"), []byte("ir0020"), true, func(key, value []byte) bool {
		keys = append(keys, string(key))
		return false
	})
	if len(keys) != 10 {
		t.Fatalf("IterateRange got %d keys, want 10", len(keys))
	}
}

func TestIterator_IterateRange_Descending(t *testing.T) {
	tree := NewMutableTreeMem()
	for i := 0; i < 100; i++ {
		tree.Set(fmt.Appendf(nil, "ir%04d", i), []byte("v"))
	}

	var keys []string
	tree.IterateRange([]byte("ir0010"), []byte("ir0020"), false, func(key, value []byte) bool {
		keys = append(keys, string(key))
		return false
	})
	if len(keys) != 10 {
		t.Fatalf("IterateRange desc got %d keys, want 10", len(keys))
	}
	if keys[0] != "ir0019" {
		t.Fatalf("first key = %s, want ir0019", keys[0])
	}
}

func TestIterator_StartEqualsEnd(t *testing.T) {
	tree := NewMutableTreeMem()
	for i := 0; i < 50; i++ {
		tree.Set(fmt.Appendf(nil, "se%04d", i), []byte("v"))
	}
	itr, _ := tree.Iterator([]byte("se0025"), []byte("se0025"), true)
	defer itr.Close()
	if itr.Valid() {
		t.Fatalf("start==end should produce 0 results")
	}
}

func TestIterator_StartGreaterThanEnd(t *testing.T) {
	tree := NewMutableTreeMem()
	for i := 0; i < 50; i++ {
		tree.Set(fmt.Appendf(nil, "sg%04d", i), []byte("v"))
	}
	itr, _ := tree.Iterator([]byte("sg0030"), []byte("sg0010"), true)
	defer itr.Close()
	if itr.Valid() {
		t.Fatalf("start>end should produce 0 results")
	}
}

func TestIterator_LeafBoundaryCrossing(t *testing.T) {
	tree := NewMutableTreeMem()
	n := B + 1 // 33 — forces exactly 2 leaves
	for i := 0; i < n; i++ {
		tree.Set(fmt.Appendf(nil, "lb%04d", i), []byte("v"))
	}
	if tree.Height() < 1 {
		t.Fatalf("need at least 2 leaves (height >= 1)")
	}

	// Ascending — verify no gaps at boundary
	itr, _ := tree.Iterator(nil, nil, true)
	defer itr.Close()
	var keys []string
	for itr.Valid() {
		keys = append(keys, string(itr.Key()))
		itr.Next()
	}
	if len(keys) != n {
		t.Fatalf("got %d keys, want %d", len(keys), n)
	}
	for i := 1; i < len(keys); i++ {
		if keys[i] <= keys[i-1] {
			t.Fatalf("gap or duplicate at boundary: %s, %s", keys[i-1], keys[i])
		}
	}
}

func TestIterator_DescendingRangeCrossLeafBoundary(t *testing.T) {
	tree := NewMutableTreeMem()
	n := B + 1
	for i := 0; i < n; i++ {
		tree.Set(fmt.Appendf(nil, "dc%04d", i), []byte("v"))
	}

	// Range that spans the leaf boundary
	itr, _ := tree.Iterator([]byte("dc0010"), []byte("dc0025"), false)
	defer itr.Close()
	var keys []string
	for itr.Valid() {
		keys = append(keys, string(itr.Key()))
		itr.Next()
	}
	if len(keys) != 15 {
		t.Fatalf("desc cross-boundary got %d keys, want 15: %v", len(keys), keys)
	}
	// Verify reverse order
	for i := 1; i < len(keys); i++ {
		if keys[i] >= keys[i-1] {
			t.Fatalf("not descending at %d: %s >= %s", i, keys[i], keys[i-1])
		}
	}
}

func TestIterator_NextAfterInvalid(t *testing.T) {
	tree := NewMutableTreeMem()
	tree.Set([]byte("x"), []byte("y"))

	itr, _ := tree.Iterator(nil, nil, true)
	itr.Next() // exhaust the single element
	if itr.Valid() {
		t.Fatalf("should be invalid after exhaustion")
	}
	// Calling Next on invalid should be a no-op, not panic
	itr.Next()
	itr.Next()
	if itr.Valid() {
		t.Fatalf("should still be invalid")
	}
	itr.Close()
}

func TestIterator_IterateRange_StopEarly(t *testing.T) {
	tree := NewMutableTreeMem()
	for i := 0; i < 100; i++ {
		tree.Set(fmt.Appendf(nil, "se%04d", i), []byte("v"))
	}

	count := 0
	stopped := tree.IterateRange(nil, nil, true, func(key, value []byte) bool {
		count++
		return count >= 5 // stop after 5
	})
	if !stopped {
		t.Fatalf("should have stopped early")
	}
	if count != 5 {
		t.Fatalf("count = %d, want 5", count)
	}
}
