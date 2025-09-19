// Package pager provides pagination functionality for AVL trees by implementing
// the pageable.Pageable interface. It allows for:
//
//   - Forward and reverse pagination through AVL tree entries
//   - Configurable page sizes
//   - URL query parameter parsing for page navigation
//   - Markdown-based pagination UI
//
// The package works by wrapping an avl.ITree instance to make it compatible with
// the pageable interface. To use it:
//
//	tree := avl.NewTree()
//	// ... add items to tree ...
//
//	// Create a new pager with default page size of 10, forward ordering
//	pager := pager.NewPager(tree, 10, false)
//
//	// Get specific page
//	page := pager.GetPage(1)
//
//	// Or get page from URL query parameters
//	page = pager.MustGetPageByPath("/path?page=1&size=20")
//
//	// Generate pagination UI
//	ui := page.Picker()
//
// This package is used in conjunction with gno.land/p/moul/pageable which provides
// the core pagination functionality.
package pager

import (
	"gno.land/p/nt/avl"
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

	// Use ReverseIterateByOffset for negative count
	if count < 0 {
		return w.tree.ReverseIterateByOffset(offset, -count, func(key string, value interface{}) bool {
			return cb(key, value)
		})
	}

	return w.tree.IterateByOffset(offset, count, func(key string, value interface{}) bool {
		return cb(key, value)
	})
}
