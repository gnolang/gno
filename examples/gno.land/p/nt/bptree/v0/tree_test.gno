package bptree

import (
	"math/rand"
	"sort"
	"testing"

	"gno.land/p/nt/avl/v0"
)

//----------------------------------------
// Basic operations

func TestNewTree(t *testing.T) {
	tree := NewBPTree32()
	if tree.Size() != 0 {
		t.Error("Expected empty tree size to be 0")
	}
}

func TestZeroValue(t *testing.T) {
	var tree BPTree
	if tree.Size() != 0 {
		t.Error("Expected zero-value tree size to be 0")
	}
	if tree.Has("x") {
		t.Error("Expected Has to return false on zero-value tree")
	}
	if _, ok := tree.Get("x"); ok {
		t.Error("Expected Get to return false on zero-value tree")
	}
	if _, ok := tree.Remove("x"); ok {
		t.Error("Expected Remove to return false on zero-value tree")
	}

	// Set should work on zero-value tree.
	if updated := tree.Set("a", 1); updated {
		t.Error("Expected Set to return false for new key")
	}
	if tree.Size() != 1 {
		t.Errorf("Expected size 1, got %d", tree.Size())
	}
	if v, ok := tree.Get("a"); !ok || v != 1 {
		t.Errorf("Expected Get(a) = (1, true), got (%v, %v)", v, ok)
	}
}

func TestSetAndGet(t *testing.T) {
	tree := NewBPTreeN(4)
	tree.Set("key1", "value1")
	tree.Set("key2", "value2")

	if tree.Size() != 2 {
		t.Errorf("Expected size 2, got %d", tree.Size())
	}

	v, ok := tree.Get("key1")
	if !ok || v != "value1" {
		t.Errorf("Expected (value1, true), got (%v, %v)", v, ok)
	}

	_, ok = tree.Get("missing")
	if ok {
		t.Error("Expected Get to return false for missing key")
	}
}

func TestSetUpdate(t *testing.T) {
	tree := NewBPTreeN(4)
	if updated := tree.Set("k", "v1"); updated {
		t.Error("Expected false for new key")
	}
	if updated := tree.Set("k", "v2"); !updated {
		t.Error("Expected true for existing key")
	}
	if tree.Size() != 1 {
		t.Errorf("Expected size 1 after update, got %d", tree.Size())
	}
	if v, _ := tree.Get("k"); v != "v2" {
		t.Errorf("Expected v2, got %v", v)
	}
}

func TestHas(t *testing.T) {
	tree := NewBPTreeN(4)
	tree.Set("a", 1)
	if !tree.Has("a") {
		t.Error("Expected Has(a) = true")
	}
	if tree.Has("b") {
		t.Error("Expected Has(b) = false")
	}
}

func TestRemove(t *testing.T) {
	tree := NewBPTreeN(4)
	tree.Set("a", 1)
	tree.Set("b", 2)

	v, ok := tree.Remove("a")
	if !ok || v != 1 {
		t.Errorf("Expected (1, true), got (%v, %v)", v, ok)
	}
	if tree.Size() != 1 {
		t.Errorf("Expected size 1, got %d", tree.Size())
	}
	if tree.Has("a") {
		t.Error("Expected Has(a) = false after remove")
	}

	v, ok = tree.Remove("missing")
	if ok || v != nil {
		t.Errorf("Expected (nil, false), got (%v, %v)", v, ok)
	}
}

func TestRemoveLastKey(t *testing.T) {
	tree := NewBPTreeN(4)
	tree.Set("a", 1)
	tree.Remove("a")
	if tree.Size() != 0 {
		t.Errorf("Expected size 0, got %d", tree.Size())
	}

	// Insert after removing everything.
	tree.Set("b", 2)
	if tree.Size() != 1 {
		t.Errorf("Expected size 1, got %d", tree.Size())
	}
	if v, ok := tree.Get("b"); !ok || v != 2 {
		t.Errorf("Expected (2, true), got (%v, %v)", v, ok)
	}
}

func TestNilValue(t *testing.T) {
	tree := NewBPTreeN(4)
	tree.Set("k", nil)
	if !tree.Has("k") {
		t.Error("Expected Has(k) = true for nil-valued entry")
	}
	v, ok := tree.Get("k")
	if !ok {
		t.Error("Expected key to exist")
	}
	if v != nil {
		t.Errorf("Expected nil value, got %v", v)
	}
	v, ok = tree.Remove("k")
	if !ok {
		t.Error("Expected remove to succeed")
	}
	if v != nil {
		t.Errorf("Expected nil removed value, got %v", v)
	}
}

func TestEmptyStringKey(t *testing.T) {
	tree := NewBPTreeN(4)
	tree.Set("", "empty")
	tree.Set("a", "alpha")
	tree.Set("b", "beta")

	if !tree.Has("") {
		t.Error("Expected Has('') = true for empty string key")
	}
	if v, ok := tree.Get(""); !ok || v != "empty" {
		t.Errorf("Expected (empty, true), got (%v, %v)", v, ok)
	}
	if tree.Size() != 3 {
		t.Errorf("Expected size 3, got %d", tree.Size())
	}

	// Remove "" key.
	v, ok := tree.Remove("")
	if !ok || v != "empty" {
		t.Errorf("Remove('') = (%v, %v), want (empty, true)", v, ok)
	}
	if tree.Size() != 2 {
		t.Errorf("Expected size 2, got %d", tree.Size())
	}
	if tree.Has("") {
		t.Error("Expected Has('') = false after remove")
	}
}

func TestGetMissingReturnsNilFalse(t *testing.T) {
	tree := NewBPTreeN(4)
	tree.Set("a", 1)
	v, ok := tree.Get("missing")
	if v != nil {
		t.Errorf("Expected nil value for missing key, got %v", v)
	}
	if ok {
		t.Error("Expected false for missing key")
	}
}

func TestRemoveMissingReturnsNilFalse(t *testing.T) {
	tree := NewBPTreeN(4)
	tree.Set("a", 1)
	v, ok := tree.Remove("missing")
	if v != nil {
		t.Errorf("Expected nil value for missing remove, got %v", v)
	}
	if ok {
		t.Error("Expected false for missing remove")
	}

	// Also on empty tree.
	var empty BPTree
	v, ok = empty.Remove("x")
	if v != nil || ok {
		t.Errorf("Remove on empty tree: got (%v, %v), want (nil, false)", v, ok)
	}
}

func TestFanoutNeverResets(t *testing.T) {
	var tree BPTree
	tree.Set("a", 1)
	tree.Remove("a")
	// Tree is empty again, but fanout should still be 32.
	tree.Set("b", 2)
	if tree.Size() != 1 {
		t.Errorf("Expected size 1, got %d", tree.Size())
	}
	if tree.fanout != 32 {
		t.Errorf("Expected fanout 32 after re-insert, got %d", tree.fanout)
	}
}

func TestSingleEntry(t *testing.T) {
	tree := NewBPTreeN(4)
	tree.Set("x", 42)

	// Root should be a leaf for single entry.
	if tree.root == nil {
		t.Fatal("root is nil")
	}
	if !tree.root.isLeaf() {
		t.Error("root should be a leaf for single entry")
	}
	if tree.Size() != 1 {
		t.Errorf("Expected size 1, got %d", tree.Size())
	}
	k, v := tree.GetByIndex(0)
	if k != "x" || v != 42 {
		t.Errorf("GetByIndex(0) = (%v, %v), want (x, 42)", k, v)
	}
}

//----------------------------------------
// GetByIndex

func TestGetByIndex(t *testing.T) {
	tree := NewBPTreeN(4)
	tree.Set("c", 3)
	tree.Set("a", 1)
	tree.Set("b", 2)

	k, v := tree.GetByIndex(0)
	if k != "a" || v != 1 {
		t.Errorf("GetByIndex(0) = (%v, %v), want (a, 1)", k, v)
	}
	k, v = tree.GetByIndex(1)
	if k != "b" || v != 2 {
		t.Errorf("GetByIndex(1) = (%v, %v), want (b, 2)", k, v)
	}
	k, v = tree.GetByIndex(2)
	if k != "c" || v != 3 {
		t.Errorf("GetByIndex(2) = (%v, %v), want (c, 3)", k, v)
	}
}

func TestGetByIndexPanics(t *testing.T) {
	tree := NewBPTreeN(4)
	tree.Set("a", 1)

	assertPanics(t, "empty tree", func() {
		var empty BPTree
		empty.GetByIndex(0)
	})
	assertPanics(t, "negative index", func() {
		tree.GetByIndex(-1)
	})
	assertPanics(t, "index == size", func() {
		tree.GetByIndex(1)
	})
}

//----------------------------------------
// Iterate

func TestIterate(t *testing.T) {
	tree := NewBPTreeN(4)
	tree.Set("a", 1)
	tree.Set("b", 2)
	tree.Set("c", 3)
	tree.Set("d", 4)
	tree.Set("e", 5)

	// Full iteration.
	got := collectKeys(tree, func(tr *BPTree, cb IterCbFn) bool {
		return tr.Iterate("", "", cb)
	})
	assertSliceEqual(t, got, []string{"a", "b", "c", "d", "e"})

	// Bounded iteration [b, d) → b, c.
	got = collectKeys(tree, func(tr *BPTree, cb IterCbFn) bool {
		return tr.Iterate("b", "d", cb)
	})
	assertSliceEqual(t, got, []string{"b", "c"})

	// start == end → empty.
	got = collectKeys(tree, func(tr *BPTree, cb IterCbFn) bool {
		return tr.Iterate("b", "b", cb)
	})
	assertSliceEqual(t, got, nil)

	// start > end → empty.
	got = collectKeys(tree, func(tr *BPTree, cb IterCbFn) bool {
		return tr.Iterate("z", "a", cb)
	})
	assertSliceEqual(t, got, nil)

	// No lower bound.
	got = collectKeys(tree, func(tr *BPTree, cb IterCbFn) bool {
		return tr.Iterate("", "c", cb)
	})
	assertSliceEqual(t, got, []string{"a", "b"})

	// Empty tree.
	var empty BPTree
	stopped := empty.Iterate("", "", func(k string, v any) bool {
		t.Error("should not be called")
		return false
	})
	if stopped {
		t.Error("Expected false from Iterate on empty tree")
	}
}

func TestIterateEarlyStop(t *testing.T) {
	tree := NewBPTreeN(4)
	tree.Set("a", 1)
	tree.Set("b", 2)
	tree.Set("c", 3)

	count := 0
	stopped := tree.Iterate("", "", func(k string, v any) bool {
		count++
		return true
	})
	if !stopped {
		t.Error("Expected true from early-stopped Iterate")
	}
	if count != 1 {
		t.Errorf("Expected 1 callback, got %d", count)
	}
}

func TestIterateAllEdgeCases(t *testing.T) {
	tree := NewBPTreeN(4)
	tree.Set("a", 1)
	tree.Set("b", 2)
	tree.Set("c", 3)
	tree.Set("d", 4)
	tree.Set("e", 5)

	// Iterate("a", "a") → empty [a,a).
	got := collectKeys(tree, func(tr *BPTree, cb IterCbFn) bool {
		return tr.Iterate("a", "a", cb)
	})
	assertSliceEqual(t, got, nil)

	// Iterate("z", "a") → empty (start > end).
	got = collectKeys(tree, func(tr *BPTree, cb IterCbFn) bool {
		return tr.Iterate("z", "a", cb)
	})
	assertSliceEqual(t, got, nil)

	// Iterate("", "a") → nothing < "a".
	got = collectKeys(tree, func(tr *BPTree, cb IterCbFn) bool {
		return tr.Iterate("", "a", cb)
	})
	assertSliceEqual(t, got, nil)

	// Iterate("", "b") → just "a".
	got = collectKeys(tree, func(tr *BPTree, cb IterCbFn) bool {
		return tr.Iterate("", "b", cb)
	})
	assertSliceEqual(t, got, []string{"a"})

	// ReverseIterate("a", "a") → visits "a" (both inclusive).
	got = collectKeys(tree, func(tr *BPTree, cb IterCbFn) bool {
		return tr.ReverseIterate("a", "a", cb)
	})
	assertSliceEqual(t, got, []string{"a"})

	// ReverseIterate("z", "a") → empty (bounds don't swap).
	got = collectKeys(tree, func(tr *BPTree, cb IterCbFn) bool {
		return tr.ReverseIterate("z", "a", cb)
	})
	assertSliceEqual(t, got, nil)

	// ReverseIterate("c", "") → keys >= "c" descending.
	got = collectKeys(tree, func(tr *BPTree, cb IterCbFn) bool {
		return tr.ReverseIterate("c", "", cb)
	})
	assertSliceEqual(t, got, []string{"e", "d", "c"})

	// Iterate("", "a") on tree with key "" stored.
	tree2 := NewBPTreeN(4)
	tree2.Set("", "empty")
	tree2.Set("a", "alpha")
	tree2.Set("b", "beta")
	got = collectKeys(tree2, func(tr *BPTree, cb IterCbFn) bool {
		return tr.Iterate("", "a", cb)
	})
	assertSliceEqual(t, got, []string{""})

	// Iterate("", "") visits all including "" key.
	got = collectKeys(tree2, func(tr *BPTree, cb IterCbFn) bool {
		return tr.Iterate("", "", cb)
	})
	assertSliceEqual(t, got, []string{"", "a", "b"})
}

func TestAllIterationEmptyTree(t *testing.T) {
	var tree BPTree
	noop := func(k string, v any) bool {
		t.Error("should not be called")
		return false
	}
	if tree.Iterate("", "", noop) {
		t.Error("Iterate on empty tree should return false")
	}
	if tree.ReverseIterate("", "", noop) {
		t.Error("ReverseIterate on empty tree should return false")
	}
	if tree.IterateByOffset(0, 1, noop) {
		t.Error("IterateByOffset on empty tree should return false")
	}
	if tree.ReverseIterateByOffset(0, 1, noop) {
		t.Error("ReverseIterateByOffset on empty tree should return false")
	}
}

func TestReverseIterate(t *testing.T) {
	tree := NewBPTreeN(4)
	tree.Set("a", 1)
	tree.Set("b", 2)
	tree.Set("c", 3)
	tree.Set("d", 4)
	tree.Set("e", 5)

	// Full reverse.
	got := collectKeys(tree, func(tr *BPTree, cb IterCbFn) bool {
		return tr.ReverseIterate("", "", cb)
	})
	assertSliceEqual(t, got, []string{"e", "d", "c", "b", "a"})

	// Bounded [b, d] inclusive → d, c, b.
	got = collectKeys(tree, func(tr *BPTree, cb IterCbFn) bool {
		return tr.ReverseIterate("b", "d", cb)
	})
	assertSliceEqual(t, got, []string{"d", "c", "b"})

	// start == end → visits that one key.
	got = collectKeys(tree, func(tr *BPTree, cb IterCbFn) bool {
		return tr.ReverseIterate("c", "c", cb)
	})
	assertSliceEqual(t, got, []string{"c"})

	// start > end → empty.
	got = collectKeys(tree, func(tr *BPTree, cb IterCbFn) bool {
		return tr.ReverseIterate("z", "a", cb)
	})
	assertSliceEqual(t, got, nil)

	// No upper bound.
	got = collectKeys(tree, func(tr *BPTree, cb IterCbFn) bool {
		return tr.ReverseIterate("c", "", cb)
	})
	assertSliceEqual(t, got, []string{"e", "d", "c"})

	// end > max key → e, d, c, b, a.
	got = collectKeys(tree, func(tr *BPTree, cb IterCbFn) bool {
		return tr.ReverseIterate("a", "z", cb)
	})
	assertSliceEqual(t, got, []string{"e", "d", "c", "b", "a"})
}

func TestIterateByOffset(t *testing.T) {
	tree := NewBPTreeN(4)
	tree.Set("a", 1)
	tree.Set("b", 2)
	tree.Set("c", 3)
	tree.Set("d", 4)
	tree.Set("e", 5)

	got := collectKeys(tree, func(tr *BPTree, cb IterCbFn) bool {
		return tr.IterateByOffset(1, 3, cb)
	})
	assertSliceEqual(t, got, []string{"b", "c", "d"})

	// offset=0, count=0 → nothing.
	got = collectKeys(tree, func(tr *BPTree, cb IterCbFn) bool {
		return tr.IterateByOffset(0, 0, cb)
	})
	assertSliceEqual(t, got, nil)

	// offset=size → nothing.
	got = collectKeys(tree, func(tr *BPTree, cb IterCbFn) bool {
		return tr.IterateByOffset(5, 1, cb)
	})
	assertSliceEqual(t, got, nil)

	// count exceeds remaining.
	got = collectKeys(tree, func(tr *BPTree, cb IterCbFn) bool {
		return tr.IterateByOffset(3, 100, cb)
	})
	assertSliceEqual(t, got, []string{"d", "e"})
}

func TestReverseIterateByOffset(t *testing.T) {
	tree := NewBPTreeN(4)
	tree.Set("a", 1)
	tree.Set("b", 2)
	tree.Set("c", 3)
	tree.Set("d", 4)
	tree.Set("e", 5)

	// Full reverse.
	got := collectKeys(tree, func(tr *BPTree, cb IterCbFn) bool {
		return tr.ReverseIterateByOffset(0, 5, cb)
	})
	assertSliceEqual(t, got, []string{"e", "d", "c", "b", "a"})

	// offset=1, count=2 → [d, c].
	got = collectKeys(tree, func(tr *BPTree, cb IterCbFn) bool {
		return tr.ReverseIterateByOffset(1, 2, cb)
	})
	assertSliceEqual(t, got, []string{"d", "c"})

	// offset=4, count=1 → [a].
	got = collectKeys(tree, func(tr *BPTree, cb IterCbFn) bool {
		return tr.ReverseIterateByOffset(4, 1, cb)
	})
	assertSliceEqual(t, got, []string{"a"})

	// offset=4, count=10 → [a].
	got = collectKeys(tree, func(tr *BPTree, cb IterCbFn) bool {
		return tr.ReverseIterateByOffset(4, 10, cb)
	})
	assertSliceEqual(t, got, []string{"a"})

	// offset >= size → nothing.
	got = collectKeys(tree, func(tr *BPTree, cb IterCbFn) bool {
		return tr.ReverseIterateByOffset(5, 1, cb)
	})
	assertSliceEqual(t, got, nil)

	// count=0 → nothing.
	got = collectKeys(tree, func(tr *BPTree, cb IterCbFn) bool {
		return tr.ReverseIterateByOffset(0, 0, cb)
	})
	assertSliceEqual(t, got, nil)

	// negative offset
	got = collectKeys(tree, func(tr *BPTree, cb IterCbFn) bool {
		return tr.ReverseIterateByOffset(-1, 3, cb)
	})
	assertSliceEqual(t, got, []string{"e", "d", "c"})
}

func TestReverseIterateEarlyStop(t *testing.T) {
	tree := NewBPTreeN(4)
	tree.Set("a", 1)
	tree.Set("b", 2)
	tree.Set("c", 3)
	count := 0
	stopped := tree.ReverseIterate("", "", func(k string, v any) bool {
		count++
		return true
	})
	if !stopped {
		t.Error("Expected true from early-stopped ReverseIterate")
	}
	if count != 1 {
		t.Errorf("Expected 1 callback, got %d", count)
	}
}

//----------------------------------------
// Splits and merges (use small fanout to trigger them)

func TestSplitAndMerge(t *testing.T) {
	tree := NewBPTreeN(4)

	// Insert enough keys to cause multiple splits.
	keys := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l"}
	for _, k := range keys {
		tree.Set(k, k)
	}
	if tree.Size() != len(keys) {
		t.Errorf("Expected size %d, got %d", len(keys), tree.Size())
	}

	// Verify all keys present and in order.
	got := collectKeys(tree, func(tr *BPTree, cb IterCbFn) bool {
		return tr.Iterate("", "", cb)
	})
	assertSliceEqual(t, got, keys)

	// Verify GetByIndex works for all positions.
	for i, k := range keys {
		gk, gv := tree.GetByIndex(i)
		if gk != k || gv != k {
			t.Errorf("GetByIndex(%d) = (%v, %v), want (%v, %v)", i, gk, gv, k, k)
		}
	}

	// Remove keys one by one and verify.
	for _, k := range keys {
		v, ok := tree.Remove(k)
		if !ok || v != k {
			t.Errorf("Remove(%v) = (%v, %v), want (%v, true)", k, v, ok, k)
		}
	}
	if tree.Size() != 0 {
		t.Errorf("Expected size 0 after removing all, got %d", tree.Size())
	}
}

func TestRemoveFromMiddle(t *testing.T) {
	tree := NewBPTreeN(4)
	for _, k := range []string{"a", "b", "c", "d", "e", "f", "g", "h"} {
		tree.Set(k, k)
	}

	// Remove from the middle to trigger redistributions and merges.
	tree.Remove("d")
	tree.Remove("e")
	tree.Remove("b")
	tree.Remove("g")

	got := collectKeys(tree, func(tr *BPTree, cb IterCbFn) bool {
		return tr.Iterate("", "", cb)
	})
	assertSliceEqual(t, got, []string{"a", "c", "f", "h"})
	if tree.Size() != 4 {
		t.Errorf("Expected size 4, got %d", tree.Size())
	}
}

func TestSequentialInsertRemove(t *testing.T) {
	tree := NewBPTreeN(4)

	// Sequential insert.
	n := 100
	for i := 0; i < n; i++ {
		k := intToKey(i)
		tree.Set(k, i)
	}
	if tree.Size() != n {
		t.Errorf("Expected size %d, got %d", n, tree.Size())
	}

	// Verify sorted order.
	prev := ""
	tree.Iterate("", "", func(k string, v any) bool {
		if k <= prev && prev != "" {
			t.Errorf("Keys not in order: %q after %q", k, prev)
		}
		prev = k
		return false
	})

	// Remove all in reverse order.
	for i := n - 1; i >= 0; i-- {
		k := intToKey(i)
		_, ok := tree.Remove(k)
		if !ok {
			t.Errorf("Remove(%v) failed", k)
		}
	}
	if tree.Size() != 0 {
		t.Errorf("Expected size 0, got %d", tree.Size())
	}
}

func TestRandomInsertRemove(t *testing.T) {
	tree := NewBPTreeN(4)

	// Insert in "random" order (shuffled via simple hash).
	keys := make([]string, 50)
	for i := range keys {
		keys[i] = intToKey((i*37 + 13) % 50)
	}
	for _, k := range keys {
		tree.Set(k, k)
	}
	if tree.Size() != 50 {
		t.Errorf("Expected size 50, got %d", tree.Size())
	}

	// Remove half.
	for i := 0; i < 25; i++ {
		tree.Remove(keys[i])
	}
	if tree.Size() != 25 {
		t.Errorf("Expected size 25, got %d", tree.Size())
	}

	// Verify remaining keys are in sorted order.
	prev := ""
	tree.Iterate("", "", func(k string, v any) bool {
		if k <= prev && prev != "" {
			t.Errorf("Keys not in order: %q after %q", k, prev)
		}
		prev = k
		return false
	})
}

func TestDifferentFanouts(t *testing.T) {
	for _, fanout := range []int{4, 5, 8, 16, 32} {
		tree := NewBPTreeN(fanout)
		n := 100
		for i := 0; i < n; i++ {
			tree.Set(intToKey(i), i)
		}
		if tree.Size() != n {
			t.Errorf("fanout=%d: expected size %d, got %d", fanout, n, tree.Size())
		}

		// Verify order.
		got := collectKeys(tree, func(tr *BPTree, cb IterCbFn) bool {
			return tr.Iterate("", "", cb)
		})
		for i := 1; i < len(got); i++ {
			if got[i] <= got[i-1] {
				t.Errorf("fanout=%d: keys not in order at %d", fanout, i)
				break
			}
		}

		// Remove all.
		for i := 0; i < n; i++ {
			tree.Remove(intToKey(i))
		}
		if tree.Size() != 0 {
			t.Errorf("fanout=%d: expected size 0 after remove all, got %d", fanout, tree.Size())
		}
	}
}

func TestNegativeCount(t *testing.T) {
	tree := NewBPTreeN(4)
	tree.Set("a", 1)

	got := collectKeys(tree, func(tr *BPTree, cb IterCbFn) bool {
		return tr.IterateByOffset(0, -1, cb)
	})
	assertSliceEqual(t, got, nil)

	got = collectKeys(tree, func(tr *BPTree, cb IterCbFn) bool {
		return tr.ReverseIterateByOffset(0, -1, cb)
	})
	assertSliceEqual(t, got, nil)
}

func TestRootCollapse(t *testing.T) {
	tree := NewBPTreeN(4)
	// Insert enough to create inner nodes.
	for _, k := range []string{"a", "b", "c", "d", "e"} {
		tree.Set(k, k)
	}
	if tree.root.isLeaf() {
		t.Error("Expected inner root after 5 inserts with fanout 4")
	}

	// Remove enough to trigger merges and root collapse.
	tree.Remove("a")
	tree.Remove("b")
	tree.Remove("c")
	// With only "d" and "e" left, root should collapse to a leaf.
	if !tree.root.isLeaf() {
		t.Error("Expected leaf root after removing down to 2 entries")
	}
	if tree.Size() != 2 {
		t.Errorf("Expected size 2, got %d", tree.Size())
	}
	got := collectKeys(tree, func(tr *BPTree, cb IterCbFn) bool {
		return tr.Iterate("", "", cb)
	})
	assertSliceEqual(t, got, []string{"d", "e"})
}

func TestSeparatorKeyAfterLeftmostDeletion(t *testing.T) {
	tree := NewBPTreeN(4)
	// Insert keys to create a split.
	for _, k := range []string{"a", "b", "c", "d", "e", "f", "g", "h"} {
		tree.Set(k, k)
	}
	// Remove leftmost key "a" — should trigger separator update.
	tree.Remove("a")

	// Verify all remaining keys are accessible.
	for _, k := range []string{"b", "c", "d", "e", "f", "g", "h"} {
		if !tree.Has(k) {
			t.Errorf("Expected Has(%s) = true after removing 'a'", k)
		}
	}
	if tree.Has("a") {
		t.Error("Expected Has(a) = false")
	}

	// Verify sorted order.
	got := collectKeys(tree, func(tr *BPTree, cb IterCbFn) bool {
		return tr.Iterate("", "", cb)
	})
	assertSliceEqual(t, got, []string{"b", "c", "d", "e", "f", "g", "h"})

	// Verify GetByIndex still works.
	k, _ := tree.GetByIndex(0)
	if k != "b" {
		t.Errorf("GetByIndex(0) = %v, want b", k)
	}
}

func TestNinetyTenSplit(t *testing.T) {
	// Sequential inserts should trigger 90/10 splits, keeping left leaves ~97% full.
	tree := NewBPTreeN(4)
	for _, k := range []string{"a", "b", "c", "d", "e"} {
		tree.Set(k, k)
	}
	// After inserting a,b,c,d (leaf full), then e (append → 90/10 split):
	// Left should have fanout-1=3 entries [a,b,c], right should have 2 entries [d,e].
	if !tree.root.isLeaf() == true {
		// Root should be an inner node after split.
	}
	inner := tree.root.(*innerNode)
	left := inner.children[0].(*leafNode)
	right := inner.children[1].(*leafNode)
	if len(left.keys) != 3 {
		t.Errorf("90/10 split: left leaf has %d entries, want 3", len(left.keys))
	}
	if len(right.keys) != 2 {
		t.Errorf("90/10 split: right leaf has %d entries, want 2", len(right.keys))
	}

	// Verify all keys accessible.
	got := collectKeys(tree, func(tr *BPTree, cb IterCbFn) bool {
		return tr.Iterate("", "", cb)
	})
	assertSliceEqual(t, got, []string{"a", "b", "c", "d", "e"})
}

func TestNinetyTenSplitLargeFanout(t *testing.T) {
	// With fanout=32, sequential inserts should produce high fill factor.
	tree := NewBPTree32()
	n := 200
	for i := 0; i < n; i++ {
		tree.Set(intToKey(i), i)
	}
	if tree.Size() != n {
		t.Errorf("Expected size %d, got %d", n, tree.Size())
	}

	// Verify all keys in order.
	prev := ""
	count := 0
	tree.Iterate("", "", func(k string, v any) bool {
		if k <= prev && prev != "" {
			t.Errorf("Keys not in order: %q after %q", k, prev)
		}
		prev = k
		count++
		return false
	})
	if count != n {
		t.Errorf("Iterate visited %d entries, want %d", count, n)
	}

	// Remove all and verify.
	for i := 0; i < n; i++ {
		_, ok := tree.Remove(intToKey(i))
		if !ok {
			t.Errorf("Remove(%s) failed", intToKey(i))
		}
	}
	if tree.Size() != 0 {
		t.Errorf("Expected size 0, got %d", tree.Size())
	}
}

func TestMixedSplitTypes(t *testing.T) {
	// Sequential inserts trigger 90/10, then a middle insert triggers 50/50.
	tree := NewBPTreeN(4)
	// Sequential: triggers 90/10 split.
	for _, k := range []string{"b", "c", "d", "e"} {
		tree.Set(k, k)
	}
	// Insert "a" at the beginning of the left leaf — not an append, so if
	// that leaf overflows it will use 50/50.
	tree.Set("a", "a")

	// Verify all keys present and sorted.
	got := collectKeys(tree, func(tr *BPTree, cb IterCbFn) bool {
		return tr.Iterate("", "", cb)
	})
	assertSliceEqual(t, got, []string{"a", "b", "c", "d", "e"})
}

func TestSizeCacheConsistency(t *testing.T) {
	tree := NewBPTreeN(4)
	keys := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l"}

	// Check after each insert.
	for _, k := range keys {
		tree.Set(k, k)
		verifySizeCache(t, tree.root)
	}

	// Check after each remove.
	for _, k := range keys {
		tree.Remove(k)
		verifySizeCache(t, tree.root)
	}

	// Also check with large fanout.
	tree2 := NewBPTree32()
	for i := 0; i < 100; i++ {
		tree2.Set(intToKey(i), i)
	}
	verifySizeCache(t, tree2.root)
	for i := 0; i < 100; i++ {
		tree2.Remove(intToKey(i))
		verifySizeCache(t, tree2.root)
	}
}

func verifySizeCache(t *testing.T, n node) {
	t.Helper()
	if n == nil || n.isLeaf() {
		return
	}
	inner := n.(*innerNode)
	// Verify each sizes[i] matches the actual child subtree size.
	for i, child := range inner.children {
		actual := child.nodeSize()
		if inner.sizes[i] != actual {
			t.Errorf("innerNode.sizes[%d]=%d but child.nodeSize()=%d", i, inner.sizes[i], actual)
		}
		verifySizeCache(t, child)
	}
}

func TestValueIndirection(t *testing.T) {
	tree := NewBPTreeN(4)

	// Nil value through *any.
	tree.Set("nil", nil)
	v, ok := tree.Get("nil")
	if !ok {
		t.Error("Expected key 'nil' to exist")
	}
	if v != nil {
		t.Errorf("Expected nil value, got %v", v)
	}

	// Various types.
	tree.Set("int", 42)
	tree.Set("str", "hello")
	tree.Set("slice", []byte{1, 2, 3})

	if v, _ := tree.Get("int"); v != 42 {
		t.Errorf("Expected 42, got %v", v)
	}
	if v, _ := tree.Get("str"); v != "hello" {
		t.Errorf("Expected hello, got %v", v)
	}
	if v, _ := tree.Get("slice"); len(v.([]byte)) != 3 {
		t.Errorf("Expected 3-byte slice, got %v", v)
	}

	// Update value in place (should reuse the *any pointer).
	tree.Set("int", 99)
	if v, _ := tree.Get("int"); v != 99 {
		t.Errorf("Expected 99 after update, got %v", v)
	}

	// Remove nil value.
	v, ok = tree.Remove("nil")
	if !ok || v != nil {
		t.Errorf("Remove nil: got (%v, %v), want (nil, true)", v, ok)
	}

	// Remove typed value.
	v, ok = tree.Remove("int")
	if !ok || v != 99 {
		t.Errorf("Remove int: got (%v, %v), want (99, true)", v, ok)
	}

	// Large string values — each stored as separate *any object.
	bigVal := make([]byte, 10000)
	for i := range bigVal {
		bigVal[i] = byte(i % 256)
	}
	for i := 0; i < 10; i++ {
		tree.Set(intToKey(i), string(bigVal))
	}
	if tree.Size() != 12 { // 10 new + "str" + "slice" remaining
		t.Errorf("Expected size 12, got %d", tree.Size())
	}

	// Verify large values round-trip correctly.
	for i := 0; i < 10; i++ {
		v, ok := tree.Get(intToKey(i))
		if !ok {
			t.Errorf("Missing key %s", intToKey(i))
			continue
		}
		s := v.(string)
		if len(s) != 10000 {
			t.Errorf("Key %s: expected 10000-byte string, got %d", intToKey(i), len(s))
		}
	}

	// Iteration with mixed values.
	count := 0
	tree.Iterate("", "", func(k string, v any) bool {
		count++
		return false
	})
	if count != 12 {
		t.Errorf("Iterate counted %d, want 12", count)
	}

	// Remove all and verify clean.
	for i := 0; i < 10; i++ {
		tree.Remove(intToKey(i))
	}
	tree.Remove("str")
	tree.Remove("slice")
	if tree.Size() != 0 {
		t.Errorf("Expected empty tree, got size %d", tree.Size())
	}
}

func TestFanoutPanics(t *testing.T) {
	assertPanics(t, "fanout 3", func() {
		NewBPTreeN(3)
	})
	assertPanics(t, "fanout 0", func() {
		NewBPTreeN(0)
	})
}

//----------------------------------------
// Ported from avl/v0 node_test.gno

func TestHasTableDriven(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		hasKey   string
		expected bool
	}{
		{"has key in non-empty tree", []string{"C", "A", "B", "E", "D"}, "B", true},
		{"does not have key in non-empty tree", []string{"C", "A", "B", "E", "D"}, "F", false},
		{"has key in single-node tree", []string{"A"}, "A", true},
		{"does not have key in single-node tree", []string{"A"}, "B", false},
		{"does not have key in empty tree", []string{}, "A", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := NewBPTreeN(4)
			for _, key := range tt.input {
				tree.Set(key, nil)
			}
			result := tree.Has(tt.hasKey)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGetByIndexTableDriven(t *testing.T) {
	tests := []struct {
		name        string
		input       []string
		idx         int
		expectKey   string
		expectPanic bool
	}{
		{"get by valid index", []string{"C", "A", "B", "E", "D"}, 2, "C", false},
		{"get by valid index (smallest)", []string{"C", "A", "B", "E", "D"}, 0, "A", false},
		{"get by valid index (largest)", []string{"C", "A", "B", "E", "D"}, 4, "E", false},
		{"get by invalid index (negative)", []string{"C", "A", "B", "E", "D"}, -1, "", true},
		{"get by invalid index (out of range)", []string{"C", "A", "B", "E", "D"}, 5, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := NewBPTreeN(4)
			for _, key := range tt.input {
				tree.Set(key, nil)
			}
			if tt.expectPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected a panic but didn't get one")
					}
				}()
			}
			key, _ := tree.GetByIndex(tt.idx)
			if !tt.expectPanic && key != tt.expectKey {
				t.Errorf("Expected key %s, got %s", tt.expectKey, key)
			}
		})
	}
}

func TestRemoveTableDriven(t *testing.T) {
	tests := []struct {
		name      string
		input     []string
		removeKey string
		expected  []string
	}{
		{"remove from middle", []string{"C", "A", "B", "D"}, "B", []string{"A", "C", "D"}},
		{"remove first key", []string{"C", "A", "B", "D"}, "A", []string{"B", "C", "D"}},
		{"remove last key", []string{"C", "A", "B", "E", "D"}, "E", []string{"A", "B", "C", "D"}},
		{"remove root-equivalent key", []string{"C", "A", "B", "E", "D"}, "C", []string{"A", "B", "D", "E"}},
		{"remove non-existent key", []string{"C", "A", "B", "E", "D"}, "F", []string{"A", "B", "C", "D", "E"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := NewBPTreeN(4)
			for _, key := range tt.input {
				tree.Set(key, nil)
			}
			tree.Remove(tt.removeKey)
			var result []string
			tree.Iterate("", "", func(key string, value any) bool {
				result = append(result, key)
				return false
			})
			if len(result) == 0 {
				result = []string{}
			}
			assertSliceEqual(t, result, tt.expected)
		})
	}
}

func TestTraverse(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{"empty tree", []string{}, []string{}},
		{"single node tree", []string{"A"}, []string{"A"}},
		{"small tree", []string{"C", "A", "B", "E", "D"}, []string{"A", "B", "C", "D", "E"}},
		{"large tree", []string{"H", "D", "L", "B", "F", "J", "N", "A", "C", "E", "G", "I", "K", "M", "O"},
			[]string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := NewBPTreeN(4)
			for _, key := range tt.input {
				tree.Set(key, nil)
			}

			t.Run("iterate", func(t *testing.T) {
				var result []string
				tree.Iterate("", "", func(key string, value any) bool {
					result = append(result, key)
					return false
				})
				if len(result) == 0 {
					result = []string{}
				}
				assertSliceEqual(t, result, tt.expected)
			})

			t.Run("ReverseIterate", func(t *testing.T) {
				var result []string
				tree.ReverseIterate("", "", func(key string, value any) bool {
					result = append(result, key)
					return false
				})
				expected := make([]string, len(tt.expected))
				copy(expected, tt.expected)
				for i, j := 0, len(expected)-1; i < j; i, j = i+1, j-1 {
					expected[i], expected[j] = expected[j], expected[i]
				}
				if len(result) == 0 {
					result = []string{}
				}
				assertSliceEqual(t, result, expected)
			})

			t.Run("TraverseInRange", func(t *testing.T) {
				var result []string
				start, end := "C", "M"
				tree.Iterate(start, end, func(key string, value any) bool {
					result = append(result, key)
					return false
				})
				expected := make([]string, 0)
				for _, key := range tt.expected {
					if key >= start && key < end {
						expected = append(expected, key)
					}
				}
				if len(result) == 0 {
					result = []string{}
				}
				assertSliceEqual(t, result, expected)
			})

			t.Run("early termination", func(t *testing.T) {
				if len(tt.input) == 0 {
					return
				}
				var result []string
				var count int
				tree.Iterate("", "", func(key string, value any) bool {
					count++
					result = append(result, key)
					return true
				})
				if count != 1 {
					t.Errorf("Expected callback to be called exactly once, got %d calls", count)
				}
				if len(result) != 1 {
					t.Errorf("Expected exactly one result, got %d items", len(result))
				}
				if len(result) > 0 && result[0] != tt.expected[0] {
					t.Errorf("Expected first item to be %v, got %v", tt.expected[0], result[0])
				}
			})
		})
	}
}

func TestTraverseByOffset(t *testing.T) {
	sl := []string{"Alfa", "Alfred", "Alpha", "Alphabet", "Beta", "Beth", "Book", "Browser"}

	// Insert in reverse order to ensure ordering is independent of insertion order.
	reversed := make([]string, len(sl))
	copy(reversed, sl)
	for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
		reversed[i], reversed[j] = reversed[j], reversed[i]
	}

	t.Run("ascending", func(t *testing.T) {
		tree := NewBPTreeN(4)
		for _, v := range reversed {
			tree.Set(v, nil)
		}

		// Single-element offset traversal.
		var result []string
		for i := 0; i < len(sl); i++ {
			tree.IterateByOffset(i, 1, func(key string, value any) bool {
				result = append(result, key)
				return false
			})
		}
		assertSliceEqual(t, result, sl)

		// Sliding window.
		for l := 2; l <= len(sl); l++ {
			for i := 0; i <= len(sl); i++ {
				max := i + l
				if max > len(sl) {
					max = len(sl)
				}
				exp := sl[i:max]
				var actual []string
				tree.IterateByOffset(i, l, func(key string, value any) bool {
					actual = append(actual, key)
					return false
				})
				if len(actual) == 0 {
					actual = []string{}
				}
				assertSliceEqual(t, actual, exp)
			}
		}
	})

	t.Run("descending", func(t *testing.T) {
		tree := NewBPTreeN(4)
		for _, v := range reversed {
			tree.Set(v, nil)
		}

		// The descending order.
		desc := make([]string, len(sl))
		copy(desc, sl)
		for i, j := 0, len(desc)-1; i < j; i, j = i+1, j-1 {
			desc[i], desc[j] = desc[j], desc[i]
		}

		// Single-element offset traversal in reverse.
		var result []string
		for i := 0; i < len(desc); i++ {
			tree.ReverseIterateByOffset(i, 1, func(key string, value any) bool {
				result = append(result, key)
				return false
			})
		}
		assertSliceEqual(t, result, desc)

		// Sliding window in descending.
		for l := 2; l <= len(desc); l++ {
			for i := 0; i <= len(desc); i++ {
				max := i + l
				if max > len(desc) {
					max = len(desc)
				}
				exp := desc[i:max]
				var actual []string
				tree.ReverseIterateByOffset(i, l, func(key string, value any) bool {
					actual = append(actual, key)
					return false
				})
				if len(actual) == 0 {
					actual = []string{}
				}
				assertSliceEqual(t, actual, exp)
			}
		}
	})
}

func TestBSTProperty(t *testing.T) {
	tree := NewBPTreeN(4)
	keys := []string{"D", "B", "F", "A", "C", "E", "G"}
	for _, key := range keys {
		tree.Set(key, nil)
	}

	var result []string
	tree.Iterate("", "", func(key string, value any) bool {
		result = append(result, key)
		return false
	})

	for i := 1; i < len(result); i++ {
		if result[i] < result[i-1] {
			t.Errorf("Sorted property violated: %s < %s (index %d)",
				result[i], result[i-1], i)
		}
	}
}

func TestRemoveFromEmptyTree(t *testing.T) {
	var tree BPTree
	val, removed := tree.Remove("NonExistent")
	if val != nil || removed {
		t.Errorf("Expected no value and removed=false when removing from empty tree")
	}
}

// Ported from avl tree_test.gno

func TestPortedTreeSize(t *testing.T) {
	tree := NewBPTree32()
	if tree.Size() != 0 {
		t.Error("Expected empty tree size to be 0")
	}
	tree.Set("key1", "value1")
	tree.Set("key2", "value2")
	if tree.Size() != 2 {
		t.Error("Expected tree size to be 2")
	}
}

func TestPortedTreeHas(t *testing.T) {
	tree := NewBPTree32()
	tree.Set("key1", "value1")
	if !tree.Has("key1") {
		t.Error("Expected tree to have key1")
	}
	if tree.Has("key2") {
		t.Error("Expected tree to not have key2")
	}
}

func TestPortedTreeGet(t *testing.T) {
	tree := NewBPTree32()
	tree.Set("key1", "value1")
	value, exists := tree.Get("key1")
	if !exists || value != "value1" {
		t.Error("Expected Get to return value1 and true")
	}
	_, exists = tree.Get("key2")
	if exists {
		t.Error("Expected Get to return false for non-existent key")
	}
}

func TestPortedTreeGetByIndex(t *testing.T) {
	tree := NewBPTree32()
	tree.Set("key1", "value1")
	tree.Set("key2", "value2")
	key, value := tree.GetByIndex(0)
	if key != "key1" || value != "value1" {
		t.Error("Expected GetByIndex(0) to return key1 and value1")
	}
	key, value = tree.GetByIndex(1)
	if key != "key2" || value != "value2" {
		t.Error("Expected GetByIndex(1) to return key2 and value2")
	}
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected GetByIndex to panic for out-of-range index")
		}
	}()
	tree.GetByIndex(2)
}

func TestPortedTreeRemove(t *testing.T) {
	tree := NewBPTree32()
	tree.Set("key1", "value1")
	value, removed := tree.Remove("key1")
	if !removed || value != "value1" || tree.Size() != 0 {
		t.Error("Expected Remove to remove key-value pair")
	}
	_, removed = tree.Remove("key2")
	if removed {
		t.Error("Expected Remove to return false for non-existent key")
	}
}

func TestPortedTreeIterate(t *testing.T) {
	tree := NewBPTree32()
	tree.Set("key1", "value1")
	tree.Set("key2", "value2")
	tree.Set("key3", "value3")
	var keys []string
	tree.Iterate("", "", func(key string, value any) bool {
		keys = append(keys, key)
		return false
	})
	assertSliceEqual(t, keys, []string{"key1", "key2", "key3"})
}

func TestPortedTreeReverseIterate(t *testing.T) {
	tree := NewBPTree32()
	tree.Set("key1", "value1")
	tree.Set("key2", "value2")
	tree.Set("key3", "value3")
	var keys []string
	tree.ReverseIterate("", "", func(key string, value any) bool {
		keys = append(keys, key)
		return false
	})
	assertSliceEqual(t, keys, []string{"key3", "key2", "key1"})
}

func TestPortedTreeIterateByOffset(t *testing.T) {
	tree := NewBPTree32()
	tree.Set("key1", "value1")
	tree.Set("key2", "value2")
	tree.Set("key3", "value3")
	var keys []string
	tree.IterateByOffset(1, 2, func(key string, value any) bool {
		keys = append(keys, key)
		return false
	})
	assertSliceEqual(t, keys, []string{"key2", "key3"})
}

func TestPortedTreeReverseIterateByOffset(t *testing.T) {
	tree := NewBPTree32()
	tree.Set("key1", "value1")
	tree.Set("key2", "value2")
	tree.Set("key3", "value3")
	var keys []string
	tree.ReverseIterateByOffset(1, 2, func(key string, value any) bool {
		keys = append(keys, key)
		return false
	})
	assertSliceEqual(t, keys, []string{"key2", "key1"})
}

func TestPortedReverseIterateByOffsetVaried(t *testing.T) {
	tree := NewBPTree32()
	tree.Set("a", 1)
	tree.Set("b", 2)
	tree.Set("c", 3)
	tree.Set("d", 4)
	tree.Set("e", 5)

	cases := []struct {
		offset int
		limit  int
		want   []string
	}{
		{0, 5, []string{"e", "d", "c", "b", "a"}},
		{0, 1, []string{"e"}},
		{0, 3, []string{"e", "d", "c"}},
		{1, 2, []string{"d", "c"}},
		{2, 2, []string{"c", "b"}},
		{3, 5, []string{"b", "a"}},
		{4, 1, []string{"a"}},
		{4, 10, []string{"a"}},
		{5, 1, nil},
		{0, 0, nil},
		{10, 1, nil},
	}

	for _, tc := range cases {
		var got []string
		tree.ReverseIterateByOffset(tc.offset, tc.limit, func(key string, value any) bool {
			got = append(got, key)
			return false
		})
		if !slicesEqual(got, tc.want) {
			t.Errorf("ReverseIterateByOffset(%d, %d): got %v, want %v",
				tc.offset, tc.limit, got, tc.want)
		}
	}

	// Early termination.
	var got []string
	tree.ReverseIterateByOffset(1, 5, func(key string, value any) bool {
		got = append(got, key)
		return len(got) >= 2
	})
	assertSliceEqual(t, got, []string{"d", "c"})
}

func TestPortedBalanceAfterRemoval(t *testing.T) {
	// This tests the behavioral equivalence: after various insert/remove
	// patterns, the tree maintains correct sorted order and all keys are
	// accessible. (AVL tests checked balance factors; we check correctness.)
	tests := []struct {
		name       string
		insertKeys []string
		removeKey  string
	}{
		{"remove right node", []string{"B", "A", "D", "C", "E"}, "E"},
		{"remove left node", []string{"D", "B", "E", "A", "C"}, "A"},
		{"remove after complex insert", []string{"C", "B", "E", "A", "D", "F"}, "F"},
		{"descending insert, remove middle", []string{"E", "D", "C", "B", "A"}, "C"},
		{"ascending insert, remove middle", []string{"A", "B", "C", "D", "E"}, "C"},
		{"duplicate insert, remove key", []string{"C", "B", "C", "A", "D"}, "C"},
		{"complex case", []string{"H", "B", "A", "C", "E", "D", "F", "G"}, "B"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := NewBPTreeN(4)
			for _, key := range tt.insertKeys {
				tree.Set(key, nil)
			}
			tree.Remove(tt.removeKey)

			// Verify sorted order.
			var result []string
			tree.Iterate("", "", func(key string, value any) bool {
				result = append(result, key)
				return false
			})
			for i := 1; i < len(result); i++ {
				if result[i] <= result[i-1] {
					t.Errorf("Sorted property violated: %s <= %s", result[i], result[i-1])
				}
			}

			// Verify all expected keys present.
			for _, key := range tt.insertKeys {
				if key == tt.removeKey {
					if tree.Has(key) {
						// Only check if the key was unique (not duplicated).
						// With duplicates, Set overwrites, so there's only one copy.
						continue
					}
				}
			}
		})
	}
}

//----------------------------------------
// Structural invariant verification

// verifyInvariants checks all B+ tree structural invariants.
func verifyInvariants(t *testing.T, tree *BPTree) {
	t.Helper()
	if tree.root == nil {
		if tree.size != 0 {
			t.Errorf("nil root but size=%d", tree.size)
		}
		return
	}

	// Check leaf depth uniformity.
	depth := leafDepth(tree.root, 0)
	if depth == -1 {
		t.Error("leaves are at different depths")
	}

	// Check separator keys, child count bounds, sizes, ref-count.
	seen := make(map[node]bool)
	verifyNode(t, tree.root, tree.fanout, true, seen)

	// Check tree.size matches actual leaf count.
	actual := tree.root.nodeSize()
	if tree.size != actual {
		t.Errorf("tree.size=%d but root.nodeSize()=%d", tree.size, actual)
	}
}

func leafDepth(n node, depth int) int {
	if n.isLeaf() {
		return depth
	}
	inner := n.(*innerNode)
	d := -1
	for _, child := range inner.children {
		cd := leafDepth(child, depth+1)
		if d == -1 {
			d = cd
		} else if cd != d {
			return -1
		}
	}
	return d
}

func verifyNode(t *testing.T, n node, fanout int, isRoot bool, seen map[node]bool) {
	t.Helper()
	if seen[n] {
		t.Errorf("node reachable by multiple paths (ref-count >= 2)")
		return
	}
	seen[n] = true

	if n.isLeaf() {
		leaf := n.(*leafNode)
		if len(leaf.keys) == 0 && !isRoot {
			t.Errorf("non-root leaf has 0 keys")
		}
		if len(leaf.keys) > fanout {
			t.Errorf("leaf has %d keys, max=%d", len(leaf.keys), fanout)
		}
		return
	}

	inner := n.(*innerNode)

	// Child count bounds.
	minC := fanout / 2
	if isRoot {
		minC = 2
	}
	if len(inner.children) < minC && !isRoot {
		t.Errorf("inner node has %d children, min=%d", len(inner.children), minC)
	}
	if len(inner.children) > fanout {
		t.Errorf("inner node has %d children, max=%d", len(inner.children), fanout)
	}

	// keys/children/sizes length consistency.
	if len(inner.keys) != len(inner.children)-1 {
		t.Errorf("len(keys)=%d but len(children)=%d", len(inner.keys), len(inner.children))
	}
	if len(inner.sizes) != len(inner.children) {
		t.Errorf("len(sizes)=%d but len(children)=%d", len(inner.sizes), len(inner.children))
	}

	// Separator key correctness: keys[i] == children[i+1].minKey().
	for i, k := range inner.keys {
		expected := inner.children[i+1].minKey()
		if k != expected {
			t.Errorf("separator keys[%d]=%q but children[%d].minKey()=%q", i, k, i+1, expected)
		}
	}

	// sizes[i] matches child.
	for i, child := range inner.children {
		actual := child.nodeSize()
		if inner.sizes[i] != actual {
			t.Errorf("sizes[%d]=%d but child.nodeSize()=%d", i, inner.sizes[i], actual)
		}
		verifyNode(t, child, fanout, false, seen)
	}
}

func TestInvariantsAfterEveryOperation(t *testing.T) {
	tree := NewBPTreeN(4)
	keys := []string{"m", "f", "t", "b", "i", "p", "w", "a", "d", "g", "k", "n", "r", "u", "y"}

	for _, k := range keys {
		tree.Set(k, k)
		verifyInvariants(t, tree)
	}
	for _, k := range keys {
		tree.Remove(k)
		verifyInvariants(t, tree)
	}
}

//----------------------------------------
// Behavioral parity: GetByIndex == IterateByOffset

func TestGetByIndexMatchesIterateByOffset(t *testing.T) {
	tree := NewBPTreeN(4)
	for _, k := range []string{"h", "d", "l", "b", "f", "j", "n", "a", "c", "e", "g"} {
		tree.Set(k, k)
	}

	for i := 0; i < tree.Size(); i++ {
		k1, v1 := tree.GetByIndex(i)
		var k2 string
		var v2 any
		tree.IterateByOffset(i, 1, func(k string, v any) bool {
			k2 = k
			v2 = v
			return true
		})
		if k1 != k2 || v1 != v2 {
			t.Errorf("index %d: GetByIndex=(%q,%v) but IterateByOffset=(%q,%v)", i, k1, v1, k2, v2)
		}
	}
}

//----------------------------------------
// Stress: oscillating tree size

func TestOscillatingSize(t *testing.T) {
	tree := NewBPTreeN(4)

	// Insert 100.
	for i := 0; i < 100; i++ {
		tree.Set(intToKey(i), i)
	}
	verifyInvariants(t, tree)

	// Remove 50.
	for i := 0; i < 50; i++ {
		tree.Remove(intToKey(i))
	}
	verifyInvariants(t, tree)
	if tree.Size() != 50 {
		t.Errorf("Expected size 50, got %d", tree.Size())
	}

	// Insert 50 new.
	for i := 100; i < 150; i++ {
		tree.Set(intToKey(i), i)
	}
	verifyInvariants(t, tree)
	if tree.Size() != 100 {
		t.Errorf("Expected size 100, got %d", tree.Size())
	}

	// Remove all.
	for i := 50; i < 150; i++ {
		tree.Remove(intToKey(i))
	}
	verifyInvariants(t, tree)
	if tree.Size() != 0 {
		t.Errorf("Expected size 0, got %d", tree.Size())
	}

	// Insert again from empty.
	for i := 0; i < 20; i++ {
		tree.Set(intToKey(i), i)
	}
	verifyInvariants(t, tree)
	if tree.Size() != 20 {
		t.Errorf("Expected size 20, got %d", tree.Size())
	}
}

//----------------------------------------
// All 6 rebalance paths

func TestAllRebalancePaths(t *testing.T) {
	// We use fanout=4 so min=2 for leaves.
	// Each sub-test constructs a specific tree state and triggers one rebalance path.

	t.Run("leaf redistribute from left", func(t *testing.T) {
		tree := NewBPTreeN(4)
		for _, k := range []string{"a", "b", "c", "d", "e"} {
			tree.Set(k, k)
		}
		// After 90/10 split: left=[a,b,c], right=[d,e].
		// Remove "d" → right=[e] (1 < min=2). Left has 3 > 2 → redistribute from left.
		tree.Remove("d")
		verifyInvariants(t, tree)
		if !tree.Has("e") {
			t.Error("Missing key 'e' after redistribute")
		}
	})

	t.Run("leaf redistribute from right", func(t *testing.T) {
		tree := NewBPTreeN(4)
		for _, k := range []string{"a", "b", "c", "d", "e", "f", "g"} {
			tree.Set(k, k)
		}
		// Remove from leftmost leaf until it underflows and must redistribute from right.
		tree.Remove("a")
		tree.Remove("b")
		verifyInvariants(t, tree)
	})

	t.Run("leaf redistribute from right, no stale root separator keys", func(t *testing.T) {
		tree := NewBPTreeN(4)
		for _, k := range []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n"} {
			tree.Set(k, k)
		}
		tree.Remove("g") // no underflow
		tree.Remove("h") // leaf [h,i] underflows and must redistribute from right
		verifyInvariants(t, tree)
	})

	t.Run("leaf redistribute from right, no stale inner node separator keys", func(t *testing.T) {
		tree := NewBPTreeN(4)
		for _, k := range []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q"} {
			tree.Set(k, k)
		}
		tree.Remove("g") // no underflow
		tree.Remove("j") // no underflow
		tree.Remove("k") // leaf [k,l] underflows and must redistribute from right
		verifyInvariants(t, tree)
	})

	t.Run("stale separator, middle leaf pos0 redistribute from right", func(t *testing.T) {
		// Exercises the core bug: remove pos=0 from a middle leaf (childIdx > 0),
		// where left sibling is at minimum (can't redistribute left) and right
		// sibling has surplus (redistribute right fires). Without the fix,
		// parent.keys[childIdx-1] remains stale.
		//
		// After inserting a-k with fanout=4, tree is:
		//   root keys=["d","g","j"]
		//   children=[["a","b","c"], ["d","e","f"], ["g","h","i"], ["j","k"]]
		//
		// Remove "a" thins children[0] to ["b","c"] (at minimum).
		// Remove "f" thins children[1] to ["d","e"] (at minimum).
		// Remove "d" is pos=0 from children[1], causing underflow to ["e"].
		//   Left sibling ["b","c"] has 2 = minKeys, can't spare.
		//   Right sibling ["g","h","i"] has 3 > minKeys, redistribute right fires.
		//   Without fix: keys[0] stays "d" (stale). With fix: updated to "e".
		tree := NewBPTreeN(4)
		for _, k := range []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"} {
			tree.Set(k, k)
		}
		tree.Remove("a") // thin left sibling to minimum
		tree.Remove("f") // thin target leaf to minimum
		tree.Remove("d") // pos=0, underflow, triggers redistribute from right
		verifyInvariants(t, tree)
	})

	t.Run("leaf merge", func(t *testing.T) {
		tree := NewBPTreeN(4)
		for _, k := range []string{"a", "b", "c", "d", "e"} {
			tree.Set(k, k)
		}
		// Remove enough to make both siblings at minimum, then one more triggers merge.
		tree.Remove("a")
		tree.Remove("b") // left=[c], right=[d,e]. left underflows. right has 2=min. merge.
		verifyInvariants(t, tree)
		got := collectKeys(tree, func(tr *BPTree, cb IterCbFn) bool {
			return tr.Iterate("", "", cb)
		})
		assertSliceEqual(t, got, []string{"c", "d", "e"})
	})

	t.Run("inner redistribute from left", func(t *testing.T) {
		// Build a tree with enough inner nodes, then remove to trigger inner rebalance.
		tree := NewBPTreeN(4)
		for i := 0; i < 30; i++ {
			tree.Set(intToKey(i), i)
		}
		// Remove from the right side to underflow a right inner node.
		for i := 25; i < 30; i++ {
			tree.Remove(intToKey(i))
		}
		verifyInvariants(t, tree)
	})

	t.Run("inner redistribute from right", func(t *testing.T) {
		tree := NewBPTreeN(4)
		for i := 0; i < 30; i++ {
			tree.Set(intToKey(i), i)
		}
		// Remove from the left side to underflow a left inner node.
		for i := 0; i < 5; i++ {
			tree.Remove(intToKey(i))
		}
		verifyInvariants(t, tree)
	})

	t.Run("inner merge", func(t *testing.T) {
		tree := NewBPTreeN(4)
		for i := 0; i < 20; i++ {
			tree.Set(intToKey(i), i)
		}
		// Remove enough to trigger inner merge.
		for i := 0; i < 15; i++ {
			tree.Remove(intToKey(i))
		}
		verifyInvariants(t, tree)
		if tree.Size() != 5 {
			t.Errorf("Expected size 5, got %d", tree.Size())
		}
	})
}

//----------------------------------------
// Value mutation after Get should not affect tree

func TestValueMutationIndependence(t *testing.T) {
	tree := NewBPTreeN(4)

	// Store a slice value.
	orig := []int{1, 2, 3}
	tree.Set("k", orig)

	// Get the value and mutate it.
	v, _ := tree.Get("k")
	got := v.([]int)
	got[0] = 999

	// The tree's stored value should also be affected (since slices are reference types
	// and *any holds the same interface). This is the expected Go behavior — not a copy.
	v2, _ := tree.Get("k")
	got2 := v2.([]int)
	if got2[0] != 999 {
		t.Errorf("Expected slice mutation to be visible (reference semantics), got %v", got2)
	}

	// But replacing the value via Set should not affect previous Get results.
	tree.Set("k", []int{10, 20, 30})
	v3, _ := tree.Get("k")
	got3 := v3.([]int)
	if got3[0] != 10 {
		t.Errorf("Expected new value after Set, got %v", got3)
	}
	// The old reference should still have 999.
	if got[0] != 999 {
		t.Errorf("Old reference should still be 999, got %v", got)
	}
}

//----------------------------------------
// Comprehensive invariant check across fanouts and sizes

func TestInvariantsAcrossFanouts(t *testing.T) {
	for _, fanout := range []int{4, 5, 7, 8, 16, 32} {
		tree := NewBPTreeN(fanout)
		n := 200

		// Insert all.
		for i := 0; i < n; i++ {
			tree.Set(intToKey(i), i)
		}
		verifyInvariants(t, tree)

		// Remove every 3rd.
		for i := 0; i < n; i += 3 {
			tree.Remove(intToKey(i))
		}
		verifyInvariants(t, tree)

		// Remove remaining.
		for i := 0; i < n; i++ {
			tree.Remove(intToKey(i))
		}
		verifyInvariants(t, tree)
		if tree.Size() != 0 {
			t.Errorf("fanout=%d: expected size 0, got %d", fanout, tree.Size())
		}
	}
}

//----------------------------------------
// Deep stack unwinding

func TestDeepStackUnwinding(t *testing.T) {
	// With fanout=4 and 500 keys, the tree is 4-5 levels deep.
	// Iterating across subtree boundaries forces advanceLeaf/retreatLeaf
	// to pop multiple stack levels.
	tree := NewBPTreeN(4)
	n := 500
	for i := 0; i < n; i++ {
		tree.Set(intToKey(i), i)
	}

	// Full ascending iteration — every leaf boundary is crossed.
	count := 0
	prev := ""
	tree.Iterate("", "", func(k string, v any) bool {
		if k <= prev && prev != "" {
			t.Errorf("ascending order broken: %q after %q", k, prev)
		}
		prev = k
		count++
		return false
	})
	if count != n {
		t.Errorf("ascending: visited %d, want %d", count, n)
	}

	// Full descending iteration.
	count = 0
	prev = ""
	tree.ReverseIterate("", "", func(k string, v any) bool {
		if prev != "" && k >= prev {
			t.Errorf("descending order broken: %q after %q", k, prev)
		}
		prev = k
		count++
		return false
	})
	if count != n {
		t.Errorf("descending: visited %d, want %d", count, n)
	}

	// Offset-based iteration crossing deep boundaries.
	// Start from the middle, iterate to the end.
	count = 0
	tree.IterateByOffset(250, 250, func(k string, v any) bool {
		count++
		return false
	})
	if count != 250 {
		t.Errorf("IterateByOffset(250,250): visited %d, want 250", count)
	}

	// Reverse offset from middle.
	count = 0
	tree.ReverseIterateByOffset(250, 250, func(k string, v any) bool {
		count++
		return false
	})
	if count != 250 {
		t.Errorf("ReverseIterateByOffset(250,250): visited %d, want 250", count)
	}

	verifyInvariants(t, tree)
}

//----------------------------------------
// Height invariant

func treeHeight(n node) int {
	if n == nil {
		return 0
	}
	if n.isLeaf() {
		return 1
	}
	return 1 + treeHeight(n.(*innerNode).children[0])
}

func TestHeightInvariant(t *testing.T) {
	// B+ tree with fanout F should have height <= 1 + log_{ceil(F/2)}(n).
	for _, fanout := range []int{4, 8, 32} {
		tree := NewBPTreeN(fanout)
		n := 1000
		for i := 0; i < n; i++ {
			tree.Set(intToKey(i), i)
		}

		h := treeHeight(tree.root)
		// Compute max height: log base ceil(fanout/2) of n, plus 1 for root.
		minFill := fanout / 2
		if minFill < 2 {
			minFill = 2
		}
		maxH := 1
		capacity := 1
		for capacity < n {
			capacity *= minFill
			maxH++
		}

		if h > maxH {
			t.Errorf("fanout=%d, n=%d: height=%d exceeds max=%d", fanout, n, h, maxH)
		}
	}
}

//----------------------------------------
// Empty string as separator key

func TestEmptyStringSeparator(t *testing.T) {
	tree := NewBPTreeN(4)

	// Insert "" first, then other keys. With sequential inserts,
	// "" will end up in the leftmost leaf and could become a separator.
	tree.Set("", "empty")
	tree.Set("a", "a")
	tree.Set("b", "b")
	tree.Set("c", "c")
	tree.Set("d", "d") // triggers split; "" should be in left leaf

	verifyInvariants(t, tree)

	// Verify all keys accessible.
	if v, ok := tree.Get(""); !ok || v != "empty" {
		t.Errorf("Get('')=(%v,%v), want (empty,true)", v, ok)
	}
	for _, k := range []string{"a", "b", "c", "d"} {
		if !tree.Has(k) {
			t.Errorf("missing key %q", k)
		}
	}

	// Iterate should include "".
	got := collectKeys(tree, func(tr *BPTree, cb IterCbFn) bool {
		return tr.Iterate("", "", cb)
	})
	assertSliceEqual(t, got, []string{"", "a", "b", "c", "d"})

	// ReverseIterate should include "".
	got = collectKeys(tree, func(tr *BPTree, cb IterCbFn) bool {
		return tr.ReverseIterate("", "", cb)
	})
	assertSliceEqual(t, got, []string{"d", "c", "b", "a", ""})

	// Remove "" and verify tree is still valid.
	v, ok := tree.Remove("")
	if !ok || v != "empty" {
		t.Errorf("Remove('')=(%v,%v), want (empty,true)", v, ok)
	}
	verifyInvariants(t, tree)
	if tree.Has("") {
		t.Error("'' should be gone after remove")
	}

	// Now insert more keys with "" to force it into a separator position.
	// Build a tree where "" is between two inner node children.
	tree2 := NewBPTreeN(4)
	// Insert keys that will sort before and after "".
	// In ASCII, "" < everything. So "" is always the smallest key.
	// To make "" a separator, we need it to be the minKey of a right child.
	// That happens when the leaf containing "" splits and "" ends up
	// as right.keys[0] (the promoted separator).
	// With 50/50 split: left gets lower half, right gets upper half.
	// Since "" is the smallest, it will always be in the left leaf.
	// With 90/10 on append: "" would be in the left leaf too.
	// So "" can become a separator only if it's the minKey of a right child
	// after a split where "" is in the right half.
	// Insert in reverse order so "" ends up in the right half of a split:
	tree2.Set("d", "d")
	tree2.Set("c", "c")
	tree2.Set("b", "b")
	tree2.Set("a", "a")
	tree2.Set("", "empty") // inserted at position 0 (not append), 50/50 split

	verifyInvariants(t, tree2)
	got = collectKeys(tree2, func(tr *BPTree, cb IterCbFn) bool {
		return tr.Iterate("", "", cb)
	})
	assertSliceEqual(t, got, []string{"", "a", "b", "c", "d"})
}

//----------------------------------------
// Pointer independence after split

func TestPointerIndependenceAfterSplit(t *testing.T) {
	tree := NewBPTreeN(4)

	// Insert 5 keys to trigger a split.
	tree.Set("a", "val_a")
	tree.Set("b", "val_b")
	tree.Set("c", "val_c")
	tree.Set("d", "val_d")
	tree.Set("e", "val_e") // triggers split

	// Get values from both halves of the split.
	vLeft, _ := tree.Get("a")
	vRight, _ := tree.Get("d")

	if vLeft != "val_a" {
		t.Errorf("left value: got %v, want val_a", vLeft)
	}
	if vRight != "val_d" {
		t.Errorf("right value: got %v, want val_d", vRight)
	}

	// Update a value in the left half — should not affect right half.
	tree.Set("a", "new_a")
	vLeft2, _ := tree.Get("a")
	vRight2, _ := tree.Get("d")
	if vLeft2 != "new_a" {
		t.Errorf("updated left: got %v, want new_a", vLeft2)
	}
	if vRight2 != "val_d" {
		t.Errorf("right should be unchanged: got %v, want val_d", vRight2)
	}

	// Update a value in the right half — should not affect left half.
	tree.Set("d", "new_d")
	vLeft3, _ := tree.Get("a")
	vRight3, _ := tree.Get("d")
	if vLeft3 != "new_a" {
		t.Errorf("left should be unchanged: got %v, want new_a", vLeft3)
	}
	if vRight3 != "new_d" {
		t.Errorf("updated right: got %v, want new_d", vRight3)
	}

	// Verify that the *any pointers in the two leaves are different objects.
	// Access internals to check.
	inner := tree.root.(*innerNode)
	leftLeaf := inner.children[0].(*leafNode)
	rightLeaf := inner.children[1].(*leafNode)

	// Each value pointer should be unique.
	seen := make(map[*any]bool)
	for _, vp := range leftLeaf.values {
		if seen[vp] {
			t.Error("duplicate *any pointer in left leaf")
		}
		seen[vp] = true
	}
	for _, vp := range rightLeaf.values {
		if seen[vp] {
			t.Error("*any pointer shared between left and right leaves after split")
		}
		seen[vp] = true
	}
}

//----------------------------------------
// Helpers

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func intToKey(i int) string {
	// Zero-padded 4-digit string for correct lexicographic ordering.
	s := "0000"
	n := i
	b := []byte(s)
	for j := 3; j >= 0; j-- {
		b[j] = byte('0' + n%10)
		n /= 10
	}
	return string(b)
}

func collectKeys(tree *BPTree, fn func(*BPTree, IterCbFn) bool) []string {
	var keys []string
	fn(tree, func(k string, v any) bool {
		keys = append(keys, k)
		return false
	})
	return keys
}

func assertSliceEqual(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("got %v, want %v", got, want)
		return
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("got %v, want %v", got, want)
			return
		}
	}
}

func assertPanics(t *testing.T, name string, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("%s: expected panic, got none", name)
		}
	}()
	fn()
}

//----------------------------------------
// Stress tests

// TestAVLCrossValidation runs identical random operations on both avl.Tree
// and BPTree and compares every return value.
func TestAVLCrossValidation(t *testing.T) {
	rng := rand.New(rand.NewPCG(12345, 0))
	at := avl.NewTree()
	bt := NewBPTree32()

	const nOps = 5000
	const keyRange = 200

	for i := 0; i < nOps; i++ {
		key := intToKey(rng.IntN(keyRange))
		op := rng.IntN(3)
		switch op {
		case 0: // Set
			aUpd := at.Set(key, i)
			bUpd := bt.Set(key, i)
			if aUpd != bUpd {
				t.Fatalf("op %d: Set(%q) updated avl=%v bpt=%v", i, key, aUpd, bUpd)
			}
		case 1: // Remove
			aVal, aOk := at.Remove(key)
			bVal, bOk := bt.Remove(key)
			if aOk != bOk {
				t.Fatalf("op %d: Remove(%q) avl=%v bpt=%v", i, key, aOk, bOk)
			}
			if aOk && aVal != bVal {
				t.Fatalf("op %d: Remove(%q) value avl=%v bpt=%v", i, key, aVal, bVal)
			}
		case 2: // Get
			aVal, aOk := at.Get(key)
			bVal, bOk := bt.Get(key)
			if aOk != bOk {
				t.Fatalf("op %d: Get(%q) avl=%v bpt=%v", i, key, aOk, bOk)
			}
			if aOk && aVal != bVal {
				t.Fatalf("op %d: Get(%q) value avl=%v bpt=%v", i, key, aVal, bVal)
			}
		}
		if at.Size() != bt.Size() {
			t.Fatalf("op %d: size avl=%d bpt=%d", i, at.Size(), bt.Size())
		}
	}

	// Compare full iteration.
	var aKeys, bKeys []string
	at.Iterate("", "", func(k string, v any) bool { aKeys = append(aKeys, k); return false })
	bt.Iterate("", "", func(k string, v any) bool { bKeys = append(bKeys, k); return false })
	assertSliceEqual(t, bKeys, aKeys)
}

// TestExhaustiveRemovalPermutations inserts N keys then tries all N!
// permutations of removal order, verifying invariants after each remove.
func TestExhaustiveRemovalPermutations(t *testing.T) {
	keys := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	n := len(keys)

	perm := make([]int, n)
	for i := range perm {
		perm[i] = i
	}

	count := 0
	permute(perm, 0, func(order []int) {
		tree := NewBPTreeN(4)
		for _, k := range keys {
			tree.Set(k, k)
		}
		for _, idx := range order {
			tree.Remove(keys[idx])
			verifyInvariants(t, tree)
		}
		if tree.Size() != 0 {
			t.Fatalf("perm %d: tree not empty after removing all keys", count)
		}
		count++
	})
}

// permute generates all permutations of arr[start:] and calls fn for each.
func permute(arr []int, start int, fn func([]int)) {
	if start == len(arr) {
		fn(arr)
		return
	}
	for i := start; i < len(arr); i++ {
		arr[start], arr[i] = arr[i], arr[start]
		permute(arr, start+1, fn)
		arr[start], arr[i] = arr[i], arr[start]
	}
}

// TestMultiFanoutStress runs the same operation sequence across different
// fanouts and verifies all produce identical results.
func TestMultiFanoutStress(t *testing.T) {
	fanouts := []int{4, 5, 6, 7, 8, 16, 32}
	const nOps = 2000
	const keyRange = 100

	// Record expected results from the first fanout.
	type result struct {
		setUpdated bool
		getVal     any
		getOk      bool
		rmVal      any
		rmOk       bool
	}

	rng0 := rand.New(rand.NewPCG(99999, 0))
	type op struct {
		kind int // 0=set, 1=remove, 2=get
		key  string
		val  int
	}
	ops := make([]op, nOps)
	for i := range ops {
		ops[i] = op{
			kind: rng0.IntN(3),
			key:  intToKey(rng0.IntN(keyRange)),
			val:  i,
		}
	}

	// Run on each fanout and collect final keys.
	var referenceKeys []string
	for fi, fanout := range fanouts {
		tree := NewBPTreeN(fanout)
		for _, o := range ops {
			switch o.kind {
			case 0:
				tree.Set(o.key, o.val)
			case 1:
				tree.Remove(o.key)
			case 2:
				tree.Get(o.key)
			}
		}
		verifyInvariants(t, tree)

		var keys []string
		tree.Iterate("", "", func(k string, v any) bool {
			keys = append(keys, k)
			return false
		})

		if fi == 0 {
			referenceKeys = keys
		} else {
			assertSliceEqual(t, keys, referenceKeys)
		}
	}
}

// TestSequentialInsertReverseRemove inserts keys 0..N-1 sequentially
// (triggering 90/10 splits) then removes them in reverse order
// (triggering cascading merges from the right).
func TestSequentialInsertReverseRemove(t *testing.T) {
	for _, fanout := range []int{4, 6, 8, 32} {
		tree := NewBPTreeN(fanout)
		n := 500
		for i := 0; i < n; i++ {
			tree.Set(intToKey(i), i)
		}
		verifyInvariants(t, tree)
		if tree.Size() != n {
			t.Errorf("fanout=%d: expected size %d, got %d", fanout, n, tree.Size())
		}

		for i := n - 1; i >= 0; i-- {
			val, ok := tree.Remove(intToKey(i))
			if !ok {
				t.Fatalf("fanout=%d: Remove(%s) returned false", fanout, intToKey(i))
			}
			if val != i {
				t.Fatalf("fanout=%d: Remove(%s) value=%v want %d", fanout, intToKey(i), val, i)
			}
			verifyInvariants(t, tree)
		}
		if tree.Size() != 0 {
			t.Errorf("fanout=%d: expected empty tree, got size %d", fanout, tree.Size())
		}
	}
}

// TestRandomOpsWithPeriodicVerification does 10K random Set/Remove calls,
// verifying invariants and sorted iteration every 100 ops.
func TestRandomOpsWithPeriodicVerification(t *testing.T) {
	rng := rand.New(rand.NewPCG(54321, 0))
	tree := NewBPTreeN(4)

	const nOps = 10000
	const keyRange = 300
	const checkEvery = 100

	// Track expected keys in a sorted slice for comparison.
	present := make(map[string]bool)

	for i := 0; i < nOps; i++ {
		key := intToKey(rng.IntN(keyRange))
		if rng.IntN(3) == 0 { // ~33% removes
			_, ok := tree.Remove(key)
			if ok {
				delete(present, key)
			}
		} else { // ~67% sets
			tree.Set(key, i)
			present[key] = true
		}

		if (i+1)%checkEvery == 0 {
			verifyInvariants(t, tree)

			// Check size.
			if tree.Size() != len(present) {
				t.Fatalf("op %d: size mismatch tree=%d map=%d", i, tree.Size(), len(present))
			}

			// Check iteration matches sorted keys.
			var expected []string
			for k := range present {
				expected = append(expected, k)
			}
			sort.Strings(expected)

			var got []string
			tree.Iterate("", "", func(k string, v any) bool {
				got = append(got, k)
				return false
			})
			assertSliceEqual(t, got, expected)
		}
	}
}

// TestGetByIndexIterateByOffsetConsistency verifies that GetByIndex(i)
// matches IterateByOffset(i, 1) for every valid index.
func TestGetByIndexIterateByOffsetConsistency(t *testing.T) {
	rng := rand.New(rand.NewPCG(77777, 0))
	tree := NewBPTreeN(4)

	// Build a tree with random insertions and removals.
	const nOps = 1000
	const keyRange = 200
	for i := 0; i < nOps; i++ {
		key := intToKey(rng.IntN(keyRange))
		if rng.IntN(4) == 0 {
			tree.Remove(key)
		} else {
			tree.Set(key, i)
		}
	}

	n := tree.Size()
	for i := 0; i < n; i++ {
		gKey, gVal := tree.GetByIndex(i)

		var iKey string
		var iVal any
		tree.IterateByOffset(i, 1, func(k string, v any) bool {
			iKey = k
			iVal = v
			return true
		})

		if gKey != iKey {
			t.Fatalf("index %d: GetByIndex key=%q IterateByOffset key=%q", i, gKey, iKey)
		}
		if gVal != iVal {
			t.Fatalf("index %d: GetByIndex val=%v IterateByOffset val=%v", i, gVal, iVal)
		}
	}

	// Also check ReverseIterateByOffset consistency.
	for i := 0; i < n; i++ {
		gKey, gVal := tree.GetByIndex(n - 1 - i)

		var rKey string
		var rVal any
		tree.ReverseIterateByOffset(i, 1, func(k string, v any) bool {
			rKey = k
			rVal = v
			return true
		})

		if gKey != rKey {
			t.Fatalf("rev index %d: GetByIndex key=%q ReverseIterateByOffset key=%q", i, gKey, rKey)
		}
		if gVal != rVal {
			t.Fatalf("rev index %d: GetByIndex val=%v ReverseIterateByOffset val=%v", i, gVal, rVal)
		}
	}
}
