package pager

import (
	"gno.land/p/demo/avl"
	"gno.land/p/moul/pageable"
)

// NewPager creates a new Pager with default values.
func NewPager(tree avl.ITree, defaultPageSize int, reversed bool) *pageable.Pager {
	wrappedTree := NewAVLWrapper(tree)
	return pageable.NewPager(wrappedTree, defaultPageSize, reversed)
}

// AVLWrapper adapts an avl.ITree to implement Pageable
type AVLWrapper struct {
	tree avl.ITree
}

func NewAVLWrapper(tree avl.ITree) *AVLWrapper {
	return &AVLWrapper{tree: tree}
}

func (w *AVLWrapper) Size() int {
	return w.tree.Size()
}

func (w *AVLWrapper) IterateByOffset(offset int, count int, cb func(index interface{}, value interface{}) bool) bool {
	return w.tree.IterateByOffset(offset, count, func(key string, value interface{}) bool {
		return cb(key, value)
	})
}
