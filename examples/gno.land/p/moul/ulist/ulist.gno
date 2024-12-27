// Package ulist provides an append-only list implementation using a binary tree structure,
// optimized for scenarios requiring sequential inserts with auto-incrementing indices.
//
// The implementation uses a binary tree where new elements are added by following a path
// determined by the binary representation of the index. This provides automatic balancing
// for append operations without requiring any balancing logic.
//
// Key characteristics:
// * O(log n) append and access operations
// * Perfect balance for power-of-2 sizes
// * No balancing needed
// * Memory efficient
// * Natural support for range queries
// * Support for soft deletion of elements
// * Forward and reverse iteration capabilities
// * Offset-based iteration with count control
package ulist

// TODO: Make avl/pager compatible in some way. Explain the limitations (not always 10 items because of nil ones).
// TODO: Use this ulist in moul/collection for the primary index.
// TODO: Consider adding a "compact" method that removes nil nodes.

import (
	"errors"
)

// List represents an append-only binary tree list
type List struct {
	root       *treeNode
	totalSize  int
	activeSize int
}

// Entry represents a key-value pair in the list, where Index is the position
// and Value is the stored data
type Entry struct {
	Index int
	Value interface{}
}

// treeNode represents a node in the binary tree
type treeNode struct {
	data  interface{}
	left  *treeNode
	right *treeNode
}

// Error variables
var (
	ErrOutOfBounds = errors.New("index out of bounds")
	ErrDeleted     = errors.New("element already deleted")
)

// New creates a new empty List instance
func New() *List {
	return &List{}
}

// Append adds one or more values to the end of the list.
// Values are added sequentially, and the list grows automatically.
func (l *List) Append(values ...interface{}) {
	for _, value := range values {
		index := l.totalSize
		node := l.findNode(index, true)
		node.data = value
		l.totalSize++
		l.activeSize++
	}
}

// Get retrieves the value at the specified index.
// Returns nil if the index is out of bounds or if the element was deleted.
func (l *List) Get(index int) interface{} {
	node := l.findNode(index, false)
	if node == nil {
		return nil
	}
	return node.data
}

// Delete marks the elements at the specified indices as deleted.
// Returns ErrOutOfBounds if any index is invalid or ErrDeleted if
// the element was already deleted.
func (l *List) Delete(indices ...int) error {
	if len(indices) == 0 {
		return nil
	}
	if l == nil || l.totalSize == 0 {
		return ErrOutOfBounds
	}

	for _, index := range indices {
		if index < 0 || index >= l.totalSize {
			return ErrOutOfBounds
		}

		node := l.findNode(index, false)
		if node == nil || node.data == nil {
			return ErrDeleted
		}
		node.data = nil
		l.activeSize--
	}

	return nil
}

// Size returns the number of active (non-deleted) elements in the list
func (l *List) Size() int {
	if l == nil {
		return 0
	}
	return l.activeSize
}

// TotalSize returns the total number of elements ever added to the list,
// including deleted elements
func (l *List) TotalSize() int {
	if l == nil {
		return 0
	}
	return l.totalSize
}

// IterCbFn is a callback function type used in iteration methods.
// Return true to stop iteration, false to continue.
type IterCbFn func(index int, value interface{}) bool

// Iterator performs iteration between start and end indices, calling cb for each entry.
// If start > end, iteration is performed in reverse order.
// Returns true if iteration was stopped early by the callback returning true.
// Skips deleted elements.
func (l *List) Iterator(start, end int, cb IterCbFn) bool {
	// For empty list or invalid range
	if l == nil || l.totalSize == 0 {
		return false
	}
	if start < 0 && end < 0 {
		return false
	}
	if start >= l.totalSize && end >= l.totalSize {
		return false
	}

	// Normalize indices
	if start < 0 {
		start = 0
	}
	if end < 0 {
		end = 0
	}
	if end >= l.totalSize {
		end = l.totalSize - 1
	}
	if start >= l.totalSize {
		start = l.totalSize - 1
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

// IteratorByOffset performs iteration starting from offset for count elements.
// If count is positive, iterates forward; if negative, iterates backward.
// The iteration stops after abs(count) elements or when reaching list bounds.
// Skips deleted elements.
func (l *List) IteratorByOffset(offset int, count int, cb IterCbFn) bool {
	if count == 0 || l == nil || l.totalSize == 0 {
		return false
	}

	// Normalize offset
	if offset < 0 {
		offset = 0
	}
	if offset >= l.totalSize {
		offset = l.totalSize - 1
	}

	// Determine end based on count direction
	var end int
	if count > 0 {
		end = l.totalSize - 1
	} else {
		end = 0
	}

	wrapperReturned := false

	// Wrap the callback to limit iterations
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

// abs returns the absolute value of x
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Add this helper method to the List struct
func (l *List) findNode(index int, create bool) *treeNode {
	// For read operations, check bounds strictly
	if !create && (l == nil || index < 0 || index >= l.totalSize) {
		return nil
	}

	// For create operations, allow index == totalSize for append
	if create && (l == nil || index < 0 || index > l.totalSize) {
		return nil
	}

	// Initialize root if needed
	if l.root == nil {
		if !create {
			return nil
		}
		l.root = &treeNode{}
		return l.root
	}

	node := l.root

	// Special case for root node
	if index == 0 {
		return node
	}

	// Calculate the number of bits needed (inline highestBit logic)
	bits := 0
	n := index + 1
	for n > 0 {
		n >>= 1
		bits++
	}

	// Start from the second highest bit
	for level := bits - 2; level >= 0; level-- {
		bit := (index & (1 << uint(level))) != 0

		if bit {
			if node.right == nil {
				if !create {
					return nil
				}
				node.right = &treeNode{}
			}
			node = node.right
		} else {
			if node.left == nil {
				if !create {
					return nil
				}
				node.left = &treeNode{}
			}
			node = node.left
		}
	}

	return node
}

// MustDelete deletes elements at the specified indices.
// Panics if any index is invalid or if any element was already deleted.
func (l *List) MustDelete(indices ...int) {
	if err := l.Delete(indices...); err != nil {
		panic(err)
	}
}

// MustGet retrieves the value at the specified index.
// Panics if the index is out of bounds or if the element was deleted.
func (l *List) MustGet(index int) interface{} {
	if l == nil || index < 0 || index >= l.totalSize {
		panic(ErrOutOfBounds)
	}
	value := l.Get(index)
	if value == nil {
		panic(ErrDeleted)
	}
	return value
}

// GetRange returns a slice of Entry containing elements between start and end indices.
// If start > end, elements are returned in reverse order.
// Deleted elements are skipped.
func (l *List) GetRange(start, end int) []Entry {
	var entries []Entry
	l.Iterator(start, end, func(index int, value interface{}) bool {
		entries = append(entries, Entry{Index: index, Value: value})
		return false
	})
	return entries
}

// GetRangeByOffset returns a slice of Entry starting from offset for count elements.
// If count is positive, returns elements forward; if negative, returns elements backward.
// The operation stops after abs(count) elements or when reaching list bounds.
// Deleted elements are skipped.
func (l *List) GetRangeByOffset(offset int, count int) []Entry {
	var entries []Entry
	l.IteratorByOffset(offset, count, func(index int, value interface{}) bool {
		entries = append(entries, Entry{Index: index, Value: value})
		return false
	})
	return entries
}

// IList defines the interface for list operations.
// It provides methods for appending, accessing, deleting, and iterating over elements.
type IList interface {
	// Basic operations
	Append(values ...interface{})
	Get(index int) interface{}
	Delete(indices ...int) error
	Size() int
	TotalSize() int

	// Must variants that panic instead of returning errors
	MustDelete(indices ...int)
	MustGet(index int) interface{}

	// Range operations
	GetRange(start, end int) []Entry
	GetRangeByOffset(offset int, count int) []Entry

	// Iterator operations
	Iterator(start, end int, cb IterCbFn) bool
	IteratorByOffset(offset int, count int, cb IterCbFn) bool
}

// Verify that List implements IList
var _ IList = (*List)(nil)
