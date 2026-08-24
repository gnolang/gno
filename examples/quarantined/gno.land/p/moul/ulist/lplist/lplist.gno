// Package lplist provides a layered proxy implementation for lists that allows transparent
// migration of data between different schema versions.
//
// LayeredProxyList wraps an existing list (source) with a new list (target) and optionally
// applies migrations to source data when it's accessed. This enables schema evolution without
// requiring upfront migration of all data, making it ideal for large datasets or when
// preserving original data is important.
//
// Key features:
// - Lazy migration: Data is only transformed when accessed, not stored in migrated form
// - Append-only source: Source data is treated as immutable to preserve original data
// - Chaining: Multiple LayeredProxyLists can be stacked for multi-step migrations
//
// Example usage:
//
//	// Define data types for different schema versions
//	type UserV1 struct {
//	    Name string
//	    Age int
//	}
//
//	type UserV2 struct {
//	    FullName string
//	    Age int
//	    Active bool
//	}
//
//	// Create source list with old schema
//	sourceList := ulist.New()
//	sourceList.Append(
//	    UserV1{Name: "Alice", Age: 30},
//	    UserV1{Name: "Bob", Age: 25},
//	)
//
//	// Define migration function from V1 to V2
//	migrateUserV1ToV2 := func(v any) any {
//	    user := v.(UserV1)
//	    return UserV2{
//	        FullName: user.Name,  // Name field renamed to FullName
//	        Age: user.Age,
//	        Active: true,         // New field with default value
//	    }
//	}
//
//	// Create layered proxy with migration
//	proxy := NewLayeredProxyList(sourceList, migrateUserV1ToV2)
//
//	// Add new data directly in V2 format
//	proxy.Append(UserV2{FullName: "Charlie", Age: 40, Active: false})
//
//	// All access through proxy returns data in V2 format
//	for i := 0; i < proxy.Size(); i++ {
//	    user := proxy.Get(i).(UserV2)
//	    fmt.Printf("User: %s, Age: %d, Active: %t\n", user.FullName, user.Age, user.Active)
//	}
package lplist

import (
	"errors"

	"gno.land/p/moul/ulist"
)

// MigratorFn is a function type that lazily converts values from source to target
type MigratorFn func(any) any

// LayeredProxyList represents a wrapper around an existing List that handles migration
type LayeredProxyList struct {
	source       ulist.IList
	target       *ulist.List
	migrator     MigratorFn
	sourceHeight int // Store initial source size to optimize lookups
}

// NewLayeredProxyList creates a new LayeredProxyList instance that wraps an existing List
func NewLayeredProxyList(source ulist.IList, migrator MigratorFn) *LayeredProxyList {
	sourceHeight := source.TotalSize()
	target := ulist.New()
	return &LayeredProxyList{
		source:       source,
		target:       target,
		migrator:     migrator,
		sourceHeight: sourceHeight,
	}
}

// Get retrieves the value at the specified index
// Uses sourceHeight to efficiently route requests
func (l *LayeredProxyList) Get(index int) any {
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
	// For indices >= sourceHeight, adjust index to be relative to target list starting at 0
	targetIndex := index - l.sourceHeight
	return l.target.Get(targetIndex)
}

// Append adds one or more values to the target list
func (l *LayeredProxyList) Append(values ...any) {
	l.target.Append(values...)
}

// Delete marks elements as deleted in the appropriate list
func (l *LayeredProxyList) Delete(indices ...int) error {
	for _, index := range indices {
		if index < l.sourceHeight {
			err := l.source.Delete(index)
			if err != nil {
				return err
			}
		}
	}

	for _, index := range indices {
		targetIndex := index - l.sourceHeight
		err := l.target.Delete(targetIndex)
		if err != nil {
			return err
		}
	}
	return nil
}

// Size returns the total number of active elements
func (l *LayeredProxyList) Size() int {
	return l.source.Size() + l.target.Size()
}

// TotalSize returns the total number of elements in the list
func (l *LayeredProxyList) TotalSize() int {
	return l.sourceHeight + l.target.TotalSize()
}

// MustDelete deletes elements, panicking on error
func (l *LayeredProxyList) MustDelete(indices ...int) {
	if err := l.Delete(indices...); err != nil {
		panic(err)
	}
}

// MustGet retrieves a value, panicking if not found
func (l *LayeredProxyList) MustGet(index int) any {
	val := l.Get(index)
	if val == nil {
		panic(ulist.ErrDeleted)
	}
	return val
}

// GetRange returns elements between start and end indices
func (l *LayeredProxyList) GetRange(start, end int) []ulist.Entry {
	var entries []ulist.Entry
	l.Iterator(start, end, func(index int, value any) bool {
		entries = append(entries, ulist.Entry{Index: index, Value: value})
		return false
	})
	return entries
}

// GetRangeByOffset returns elements starting from offset
func (l *LayeredProxyList) GetRangeByOffset(offset int, count int) []ulist.Entry {
	var entries []ulist.Entry
	l.IteratorByOffset(offset, count, func(index int, value any) bool {
		entries = append(entries, ulist.Entry{Index: index, Value: value})
		return false
	})
	return entries
}

// Iterator performs iteration between start and end indices
func (l *LayeredProxyList) Iterator(start, end int, cb ulist.IterCbFn) bool {
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
func (l *LayeredProxyList) IteratorByOffset(offset int, count int, cb ulist.IterCbFn) bool {
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

	// Calculate absolute value manually instead of using abs function
	remaining := count
	if remaining < 0 {
		remaining = -remaining
	}

	wrapper := func(index int, value any) bool {
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

// Set updates the value at the specified index
func (l *LayeredProxyList) Set(index int, value any) error {
	if index < l.sourceHeight {
		// Cannot modify source list directly
		return errors.New("cannot modify source list directly")
	}

	// Adjust index to be relative to target list starting at 0
	targetIndex := index - l.sourceHeight
	return l.target.Set(targetIndex, value)
}

// MustSet updates the value at the specified index, panicking on error
func (l *LayeredProxyList) MustSet(index int, value any) {
	if err := l.Set(index, value); err != nil {
		panic(err)
	}
}

// GetByOffset returns elements starting from offset with count determining direction
func (l *LayeredProxyList) GetByOffset(offset int, count int) []ulist.Entry {
	var entries []ulist.Entry
	l.IteratorByOffset(offset, count, func(index int, value any) bool {
		entries = append(entries, ulist.Entry{Index: index, Value: value})
		return false
	})
	return entries
}

// Verify that LayeredProxyList implements IList
var _ ulist.IList = (*LayeredProxyList)(nil)
