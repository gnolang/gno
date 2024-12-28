package pager

import (
	"gno.land/p/moul/pageable"
	"gno.land/p/moul/ulist"
)

func NewPager(list ulist.IList, defaultPageSize int, reversed bool) *pageable.Pager {
	wrappedList := NewUListWrapper(list)
	return pageable.NewPager(wrappedList, defaultPageSize, reversed)
}

// UListWrapper adapts a ulist.IList to implement Pageable
type UListWrapper struct {
	list ulist.IList
}

func NewUListWrapper(list ulist.IList) *UListWrapper {
	return &UListWrapper{list: list}
}

func (w *UListWrapper) Size() int {
	return w.list.TotalSize()
}

func (w *UListWrapper) IterateByOffset(offset int, count int, cb func(index interface{}, value interface{}) bool) bool {
	return w.list.IteratorByOffset(offset, count, func(index int, value interface{}) bool {
		return cb(index, value)
	})
}
