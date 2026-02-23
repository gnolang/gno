// Package pager provides pagination functionality for ulist by implementing
// the pageable.Pageable interface. It allows for:
//
//   - Forward and reverse pagination through ulist entries
//   - Configurable page sizes
//   - URL query parameter parsing for page navigation
//   - Markdown-based pagination UI
//   - Support for deleted/nil elements (they are skipped during iteration)
//
// The package works by wrapping a ulist.IList instance to make it compatible with
// the pageable interface. To use it:
//
//	list := ulist.New()
//	// ... add items to list ...
//
//	// Create a new pager with default page size of 10, forward ordering
//	pager := pager.NewPager(list, 10, false)
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
//
// Note: Unlike AVL-based paging which guarantees exactly the requested number of items
// per page (unless at the end), ulist paging may return fewer items if there are
// deleted/nil elements within the requested range.
package pager

import (
	"gno.land/p/moul/pageable"
	"gno.land/p/moul/ulist"
)

// NewPager creates a new Pager with default values.
func NewPager(list ulist.IList, defaultPageSize int, reversed bool) *pageable.Pager {
	wrappedList := NewWrapper(list)
	return pageable.NewPager(wrappedList, defaultPageSize, reversed)
}

// Wrapper adapts a ulist.IList to implement Pageable.
// It handles the mapping between ulist's iteration methods and the pageable interface.
type Wrapper struct {
	list ulist.IList
}

// NewWrapper creates a new wrapper around a ulist.IList.
func NewWrapper(list ulist.IList) *Wrapper {
	return &Wrapper{list: list}
}

// Size returns the total size of the list, including deleted elements.
// This is used to calculate page numbers and determine navigation bounds.
func (w *Wrapper) Size() int {
	return w.list.TotalSize()
}

// IterateByOffset performs iteration starting from offset for count elements.
// The callback receives an index and a value, returns true to stop iteration.
//
// For reverse iteration (negative count), the ulist.IteratorByOffset handles
// the direction internally. Deleted/nil elements are automatically skipped
// by the underlying ulist iterator.
func (w *Wrapper) IterateByOffset(offset int, count int, cb func(index interface{}, value interface{}) bool) bool {
	if count == 0 {
		return false
	}

	size := w.list.TotalSize()
	if size == 0 {
		return false
	}

	if offset >= size {
		return false
	}

	// ulist's IteratorByOffset already handles negative counts for reverse iteration
	return w.list.IteratorByOffset(offset, count, func(index int, value any) bool {
		return cb(index, value)
	})
}
