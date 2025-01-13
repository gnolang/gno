package admindao

// TODO: Remove when a read only tree interface is available officially

import "gno.land/p/demo/avl"

type IReadOnlyTree interface {
	Size() int
	Has(key string) bool
	Get(key string) (value interface{}, exists bool)
	GetByIndex(index int) (key string, value interface{})
	Iterate(start, end string, cb avl.IterCbFn) bool
	ReverseIterate(start, end string, cb avl.IterCbFn) bool
	IterateByOffset(offset int, count int, cb avl.IterCbFn) bool
	ReverseIterateByOffset(offset int, count int, cb avl.IterCbFn) bool
}
