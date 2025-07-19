package pager

import (
	"gno.land/p/moul/pageable"
	"gno.land/p/moul/ulist"
)

func NewPager(list ulist.IList, defaultPageSize int, reversed bool) *pageable.Pager {
	wrappedList := NewWrapper(list)
	return pageable.NewPager(wrappedList, defaultPageSize, reversed)
}

// Wrapper adapts a ulist.IList to implement Pageable
type Wrapper struct {
	list ulist.IList
}

func NewWrapper(list ulist.IList) *Wrapper {
	return &Wrapper{list: list}
}

func (w *Wrapper) Size() int {
	return w.list.TotalSize()
}

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

	return w.list.IteratorByOffset(offset, count, func(index int, value interface{}) bool {
		return cb(index, value)
	})
}
