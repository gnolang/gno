package ulist

import (
	"errors"
)

// MigratorFn is a function type that lazily converts values from source to target
type MigratorFn func(interface{}) interface{}

// LayeredProxyList represents a wrapper around an existing List that handles migration
type LayeredProxyList struct {
	source       IList
	target       *List
	migrator     MigratorFn
	sourceHeight int // Store initial source size to optimize lookups
}

// NewLayeredProxyList creates a new LayeredProxyList instance that wraps an existing List
func NewLayeredProxyList(source IList, migrator MigratorFn) *LayeredProxyList {
	sourceHeight := source.TotalSize()
	target := New()
	target.totalSize = sourceHeight
	return &LayeredProxyList{
		source:       source,
		target:       target,
		migrator:     migrator,
		sourceHeight: sourceHeight,
	}
}

// Get retrieves the value at the specified index
// Uses sourceHeight to efficiently route requests
func (l *LayeredProxyList) Get(index int) interface{} {
	if index < l.sourceHeight {
		// Direct access to source for indices below sourceHeight
		val := l.source.Get(index)
		if val == nil {
			return nil
		}
		// Only apply migrator if it exists
		if l.migrator != nil {
			return l.migrator(val)
		}
		return val
	}
	// Access target list directly for new indices
	return l.target.Get(index)
}

// Append adds one or more values to the target list
func (l *LayeredProxyList) Append(values ...interface{}) {
	l.target.Append(values...)
}

// Delete marks elements as deleted in the appropriate list
func (l *LayeredProxyList) Delete(indices ...int) error {
	for _, index := range indices {
		if index < l.sourceHeight {
			return errors.New("cannot delete from source list")
		}
	}
	return l.target.Delete(indices...)
}

// Size returns the total number of active elements
func (l *LayeredProxyList) Size() int {
	return l.source.Size() + l.target.Size()
}

// TotalSize returns the total number of elements ever added
func (l *LayeredProxyList) TotalSize() int {
	return l.target.TotalSize()
}

// MustDelete deletes elements, panicking on error
func (l *LayeredProxyList) MustDelete(indices ...int) {
	if err := l.Delete(indices...); err != nil {
		panic(err)
	}
}

// MustGet retrieves a value, panicking if not found
func (l *LayeredProxyList) MustGet(index int) interface{} {
	val := l.Get(index)
	if val == nil {
		panic(ErrDeleted)
	}
	return val
}

// GetRange returns elements between start and end indices
func (l *LayeredProxyList) GetRange(start, end int) []Entry {
	var entries []Entry
	l.Iterator(start, end, func(index int, value interface{}) bool {
		entries = append(entries, Entry{Index: index, Value: value})
		return false
	})
	return entries
}

// GetRangeByOffset returns elements starting from offset
func (l *LayeredProxyList) GetRangeByOffset(offset int, count int) []Entry {
	var entries []Entry
	l.IteratorByOffset(offset, count, func(index int, value interface{}) bool {
		entries = append(entries, Entry{Index: index, Value: value})
		return false
	})
	return entries
}

// Iterator performs iteration between start and end indices
func (l *LayeredProxyList) Iterator(start, end int, cb IterCbFn) bool {
	// For empty list or invalid range
	if start < 0 && end < 0 {
		return false
	}

	// Normalize indices
	if start < 0 {
		start = 0
	}
	if end < 0 {
		end = 0
	}

	totalSize := l.TotalSize()
	if end >= totalSize {
		end = totalSize - 1
	}
	if start >= totalSize {
		start = totalSize - 1
	}

	// Handle reverse iteration
	if start > end {
		for i := start; i >= end; i-- {
			val := l.Get(i)
			if val != nil {
				if cb(i, val) {
					return true
				}
			}
		}
		return false
	}

	// Handle forward iteration
	for i := start; i <= end; i++ {
		val := l.Get(i)
		if val != nil {
			if cb(i, val) {
				return true
			}
		}
	}
	return false
}

// IteratorByOffset performs iteration starting from offset
func (l *LayeredProxyList) IteratorByOffset(offset int, count int, cb IterCbFn) bool {
	if count == 0 {
		return false
	}

	// Normalize offset
	if offset < 0 {
		offset = 0
	}
	totalSize := l.TotalSize()
	if offset >= totalSize {
		offset = totalSize - 1
	}

	// Determine end based on count direction
	var end int
	if count > 0 {
		end = totalSize - 1
	} else {
		end = 0
	}

	wrapperReturned := false
	remaining := abs(count)
	wrapper := func(index int, value interface{}) bool {
		if remaining <= 0 {
			wrapperReturned = true
			return true
		}
		remaining--
		return cb(index, value)
	}

	ret := l.Iterator(offset, end, wrapper)
	if wrapperReturned {
		return false
	}
	return ret
}

// Verify that LayeredProxyList implements IList
var _ IList = (*LayeredProxyList)(nil)
