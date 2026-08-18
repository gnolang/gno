// Package list implements a dynamic list data structure backed by an AVL tree.
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
//	value := l.Get(1)  // returns 2
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
//	// Create a list of specific size
//	l = list.Make(3, "default")  // creates [default, default, default]
//
//	// Create a list using a variable declaration
//	var l2 list.List
//	l2.Append(4, 5, 6)
//	println(l2.Len())  // Output: 3
package list

import (
	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/avl/v0/rotree"
	"gno.land/p/nt/seqid/v0"
)

// IList defines the interface for list operations
type IList interface {
	Len() int
	Append(values ...any)
	Get(index int) any
	Set(index int, value any) bool
	Delete(index int) (any, bool)
	Slice(startIndex, endIndex int) []any
	ForEach(fn func(index int, value any) bool)
	Clone() *List
	DeleteRange(startIndex, endIndex int) int
}

// Verify List implements IList interface
var _ IList = (*List)(nil)

// List represents an ordered sequence of items backed by an AVL tree
type List struct {
	tree  avl.Tree
	idGen seqid.ID
}

// Len returns the number of elements in the list.
//
// Example:
//
//	var l list.List
//	l.Append(1, 2, 3)
//	println(l.Len()) // Output: 3
func (l *List) Len() int {
	return l.tree.Size()
}

// Append adds one or more values to the end of the list.
//
// Example:
//
//	var l list.List
//	l.Append(1)        // adds single value
//	l.Append(2, 3, 4)  // adds multiple values
//	println(l.Len()) // Output: 4
func (l *List) Append(values ...any) {
	for _, v := range values {
		l.tree.Set(l.idGen.Next().String(), v)
	}
}

// Get returns the value at the specified index.
// Returns nil if index is out of bounds.
//
// Example:
//
//	var l list.List
//	l.Append(1, 2, 3)
//	println(l.Get(1))    // Output: 2
//	println(l.Get(-1))   // Output: nil
//	println(l.Get(999))  // Output: nil
func (l *List) Get(index int) any {
	if index < 0 || index >= l.tree.Size() {
		return nil
	}
	_, value := l.tree.GetByIndex(index)
	return value
}

// Set updates or appends a value at the specified index.
// Returns true if the operation was successful, false otherwise.
// For empty lists, only index 0 is valid (append case).
//
// Example:
//
//	var l list.List
//	l.Append(1, 2, 3)
//
//	l.Set(1, 42)      // updates existing index
//	println(l.Get(1)) // Output: 42
//
//	l.Set(3, 4)       // appends at end
//	println(l.Get(3)) // Output: 4
//
//	l.Set(-1, 5)      // invalid index
//	println(l.Len()) // Output: 4 (list unchanged)
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
//
// Example:
//
//	var l list.List
//	l.Append(1, 2, 3)
//
//	val, ok := l.Delete(1)
//	println(val, ok)  // Output: 2 true
//	println(l.Len())  // Output: 2
//
//	val, ok = l.Delete(-1)
//	println(val, ok)  // Output: nil false
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
//
// Example:
//
//	var l list.List
//	l.Append(1, 2, 3, 4, 5)
//
//	println(l.Slice(1, 4))   // Output: [2 3 4]
//	println(l.Slice(-1, 2))  // Output: [1 2]
//	println(l.Slice(3, 999)) // Output: [4 5]
//	println(l.Slice(3, 2))   // Output: nil
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
//
// Example:
//
//	var l list.List
//	l.Append(1, 2, 3)
//
//	clone := l.Clone()
//	clone.Set(0, 42)
//
//	println(l.Get(0))    // Output: 1
//	println(clone.Get(0)) // Output: 42
func (l *List) Clone() *List {
	newList := &List{
		tree:  avl.Tree{},
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
//
// Example:
//
//	var l list.List
//	l.Append(1, 2, 3, 4, 5)
//
//	deleted := l.DeleteRange(1, 4)
//	println(deleted)     // Output: 3
//	println(l.Range(0, l.Len())) // Output: [1 5]
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

// Tree returns a read-only pointer to the underlying AVL tree.
//
// Example:
//
//	var l list.List
//	l.Append(1, 2, 3)
//
//	rotree := l.Tree()
//	rotree.ReverseIterateByOffset(0, rotree.Size(), func(key string, value any) bool {
//	  println(value) // Output: 3 2 1
//	  return false
//	})
func (l *List) Tree() *rotree.ReadOnlyTree {
	return rotree.Wrap(&l.tree, nil)
}
