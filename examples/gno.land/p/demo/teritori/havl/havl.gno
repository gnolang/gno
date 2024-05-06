package havl

import (
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/avl"
)

type Tree struct {
	root          avl.Tree // height -> *avl.Tree
	initialHeight int64
}

var Latest = int64(0)

// FIXME: this is not optimized at all, we make a full copy on write

func NewTree() *Tree {
	return &Tree{initialHeight: std.GetHeight()}
}

func (t *Tree) Size(height int64) int {
	snapshot, _ := t.GetSnapshot(height)
	return snapshot.Size()
}

func (t *Tree) Has(key string, height int64) (has bool) {
	snapshot, _ := t.GetSnapshot(height)
	return snapshot.Has(key)
}

func (t *Tree) Get(key string, height int64) (value interface{}, exists bool) {
	snapshot, _ := t.GetSnapshot(height)
	return snapshot.Get(key)
}

func (t *Tree) GetByIndex(index int, height int64) (key string, value interface{}) {
	snapshot, _ := t.GetSnapshot(height)
	return snapshot.GetByIndex(index)
}

func (t *Tree) Set(key string, value interface{}) (updated bool) {
	root := t.getOrCreateCurrentRoot()
	return root.Set(key, value)
}

func (t *Tree) Remove(key string) (value interface{}, removed bool) {
	root := t.getOrCreateCurrentRoot()
	return root.Remove(key)
}

// Shortcut for TraverseInRange.
func (t *Tree) Iterate(start, end string, height int64, cb avl.IterCbFn) bool {
	snapshot, _ := t.GetSnapshot(height)
	return snapshot.Iterate(start, end, cb)
}

// Shortcut for TraverseInRange.
func (t *Tree) ReverseIterate(start, end string, height int64, cb avl.IterCbFn) bool {
	snapshot, _ := t.GetSnapshot(height)
	return snapshot.ReverseIterate(start, end, cb)
}

// Shortcut for TraverseByOffset.
func (t *Tree) IterateByOffset(offset int, count int, height int64, cb avl.IterCbFn) bool {
	snapshot, _ := t.GetSnapshot(height)
	return snapshot.IterateByOffset(offset, count, cb)
}

// Shortcut for TraverseByOffset.
func (t *Tree) ReverseIterateByOffset(offset int, count int, height int64, cb avl.IterCbFn) bool {
	snapshot, _ := t.GetSnapshot(height)
	return snapshot.ReverseIterateByOffset(offset, count, cb)
}

func (t *Tree) GetSnapshot(height int64) (*avl.Tree, int64) {
	key := getPaddedKey(height)
	var snapshot *avl.Tree
	snapshotHeight := int(t.initialHeight)
	t.root.ReverseIterate("", key, func(key string, value interface{}) bool {
		snapshot = value.(*avl.Tree)
		var err error
		snapshotHeight, err = strconv.Atoi(key)
		if err != nil {
			panic("internal error: failed to unmarshal key")
		}
		return true
	})

	if snapshot == nil {
		snapshot = avl.NewTree()
	}

	return snapshot, int64(snapshotHeight)
}

// utils

func getPaddedKey(height int64) string {
	if height <= 0 {
		height = std.GetHeight()
	}

	val := strconv.Itoa(int(height))
	return strings.Repeat("0", len("9223372036854775807")-len(val)) + val
}

func clone(t *avl.Tree) *avl.Tree {
	r := avl.NewTree()
	t.Iterate("", "", func(key string, value interface{}) bool {
		r.Set(key, value)
		return false
	})

	return r
}

func (t *Tree) getOrCreateCurrentRoot() *avl.Tree {
	key := getPaddedKey(0)
	iroot, ok := t.root.Get(key)
	var root *avl.Tree
	if ok {
		root = iroot.(*avl.Tree)
	} else {
		snapshot, _ := t.GetSnapshot(0)
		root = clone(snapshot)
		t.root.Set(key, root)
	}

	return root
}
