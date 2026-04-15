// Package list implements a dynamic list data structure backed by a B+ tree.
// It provides O(log n) operations for most list operations while maintaining
// order stability.
//
// The list supports various operations including append, get, set, delete,
// range queries, and iteration. It can store values of any type.
//
// Example usage:
//
//	// Create a new list and add elements
//	var l list.List
//	l.Append(1, 2, 3)
//
//	// Get and set elements
//	value, _ := l.Get(1)  // returns 2
//	l.Set(1, 42)      // updates index 1 to 42
//
//	// Delete elements
//	l.Delete(0)       // removes first element
//
//	// Iterate over elements
//	l.ForEach(func(index int, value any) bool {
//	    ufmt.Printf("index %d: %v\n", index, value)
//	    return false  // continue iteration
//	})
//	// Output:
//	// index 0: 42
//	// index 1: 3
//
//	// Create a list using a variable declaration
//	var l2 list.List
//	l2.Append(4, 5, 6)
//	println(l2.Len())  // Output: 3
package list

import (
	"gno.land/p/nt/bptree/v0"
	"gno.land/p/nt/bptree/v0/rotree"
	"gno.land/p/nt/seqid/v0"
)

// IList defines the interface for list operations
type IList interface {
	Len() int
	Append(values ...any)
	Get(index int) (any, bool)
	Set(index int, value any) bool
	Delete(index int) (any, bool)
	Slice(startIndex, endIndex int) []any
	ForEach(fn func(index int, value any) bool)
	Clone() *List
	DeleteRange(startIndex, endIndex int) int
}

// Verify List implements IList interface
var _ IList = (*List)(nil)

// List represents an ordered sequence of items backed by a B+ tree
type List struct {
	tree  bptree.BPTree
	idGen seqid.ID
}

// Len returns the number of elements in the list.
func (l *List) Len() int {
	return l.tree.Size()
}

// Append adds one or more values to the end of the list.
func (l *List) Append(values ...any) {
	for _, v := range values {
		l.tree.Set(l.idGen.Next().String(), v)
	}
}

// Get returns the value at the specified index and true if the index is valid.
// Returns (nil, false) if index is out of bounds.
func (l *List) Get(index int) (any, bool) {
	if index < 0 || index >= l.tree.Size() {
		return nil, false
	}
	_, value := l.tree.GetByIndex(index)
	return value, true
}

// Set updates or appends a value at the specified index.
// Returns true if the operation was successful, false otherwise.
// For empty lists, only index 0 is valid (append case).
func (l *List) Set(index int, value any) bool {
	size := l.tree.Size()

	// Handle empty list case - only allow index 0
	if size == 0 {
		if index == 0 {
			l.Append(value)
			return true
		}
		return false
	}

	if index < 0 || index > size {
		return false
	}

	// If setting at the end (append case)
	if index == size {
		l.Append(value)
		return true
	}

	// Get the key at the specified index
	key, _ := l.tree.GetByIndex(index)
	if key == "" {
		return false
	}

	// Update the value at the existing key
	l.tree.Set(key, value)
	return true
}

// Delete removes the element at the specified index.
// Returns the deleted value and true if successful, nil and false otherwise.
func (l *List) Delete(index int) (any, bool) {
	size := l.tree.Size()
	// Always return nil, false for empty list
	if size == 0 {
		return nil, false
	}

	if index < 0 || index >= size {
		return nil, false
	}

	key, value := l.tree.GetByIndex(index)
	if key == "" {
		return nil, false
	}

	l.tree.Remove(key)
	return value, true
}

// Slice returns a slice of values from startIndex (inclusive) to endIndex (exclusive).
// Returns nil if the range is invalid.
func (l *List) Slice(startIndex, endIndex int) []any {
	size := l.tree.Size()

	// Normalize bounds
	if startIndex < 0 {
		startIndex = 0
	}
	if endIndex > size {
		endIndex = size
	}
	if startIndex >= endIndex {
		return nil
	}

	count := endIndex - startIndex
	result := make([]any, count)

	i := 0
	l.tree.IterateByOffset(startIndex, count, func(_ string, value any) bool {
		result[i] = value
		i++
		return false
	})
	return result
}

// ForEach iterates through all elements in the list.
func (l *List) ForEach(fn func(index int, value any) bool) {
	if l.tree.Size() == 0 {
		return
	}

	index := 0
	l.tree.IterateByOffset(0, l.tree.Size(), func(_ string, value any) bool {
		result := fn(index, value)
		index++
		return result
	})
}

// Clone creates a shallow copy of the list.
func (l *List) Clone() *List {
	newList := &List{
		tree:  bptree.BPTree{},
		idGen: l.idGen,
	}

	size := l.tree.Size()
	if size == 0 {
		return newList
	}

	l.tree.IterateByOffset(0, size, func(_ string, value any) bool {
		newList.Append(value)
		return false
	})

	return newList
}

// DeleteRange removes elements from startIndex (inclusive) to endIndex (exclusive).
// Returns the number of elements deleted.
func (l *List) DeleteRange(startIndex, endIndex int) int {
	size := l.tree.Size()

	// Normalize bounds
	if startIndex < 0 {
		startIndex = 0
	}
	if endIndex > size {
		endIndex = size
	}
	if startIndex >= endIndex {
		return 0
	}

	// Collect keys to delete
	keysToDelete := make([]string, 0, endIndex-startIndex)
	l.tree.IterateByOffset(startIndex, endIndex-startIndex, func(key string, _ any) bool {
		keysToDelete = append(keysToDelete, key)
		return false
	})

	// Delete collected keys
	for _, key := range keysToDelete {
		l.tree.Remove(key)
	}

	return len(keysToDelete)
}

// Tree returns a read-only pointer to the underlying B+ tree.
func (l *List) Tree() *rotree.ReadOnlyTree {
	return rotree.Wrap(&l.tree, nil)
}
