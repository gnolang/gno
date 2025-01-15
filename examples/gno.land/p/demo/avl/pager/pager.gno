package pager

import (
	"gno.land/p/demo/avl"
	"gno.land/p/moul/pageable"
)

// NewPager creates a new Pager with default values.
func NewPager(tree avl.ITree, defaultPageSize int, reversed bool) *pageable.Pager {
	wrappedTree := NewWrapper(tree)
	return pageable.NewPager(wrappedTree, defaultPageSize, reversed)
}

// Wrapper adapts an avl.ITree to implement Pageable
type Wrapper struct {
	tree avl.ITree
}

func NewWrapper(tree avl.ITree) *Wrapper {
	return &Wrapper{tree: tree}
}

func (w *Wrapper) Size() int {
	return w.tree.Size()
}

func (w *Wrapper) IterateByOffset(offset int, count int, cb func(index interface{}, value interface{}) bool) bool {
	if count == 0 {
		return false
	}

	size := w.tree.Size()
	if size == 0 {
		return false
	}

	if offset >= size {
		return false
	}

	// Use the tree's IterateByOffset with the full count
	// (positive for forward, negative for reverse)
	return w.tree.IterateByOffset(offset, count, func(key string, value interface{}) bool {
		return cb(key, value)
	})
}

// Helper functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
