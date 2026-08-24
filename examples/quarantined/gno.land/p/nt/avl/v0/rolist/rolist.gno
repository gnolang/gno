// Package rolist provides a read-only wrapper for list.List with safe value transformation.
//
// It is useful when you want to expose a read-only view of a list while ensuring that
// the sensitive data cannot be modified.
//
// Example:
//
//	// Define a user structure with sensitive data
//	type User struct {
//	    Name     string
//	    Balance  int
//	    Internal string // sensitive field
//	}
//
//	// Create and populate the original list
//	var privateList list.List
//	privateList.Append(&User{
//	    Name:     "Alice",
//	    Balance:  100,
//	    Internal: "sensitive",
//	})
//
//	// Create a safe transformation function that copies the struct
//	// while excluding sensitive data
//	makeEntrySafeFn := func(v any) any {
//	    u := v.(*User)
//	    return &User{
//	        Name:     u.Name,
//	        Balance:  u.Balance,
//	        Internal: "", // omit sensitive data
//	    }
//	}
//
//	// Create a read-only view of the list
//	publicList := rolist.Wrap(privateList, makeEntrySafeFn)
//
//	// Safely access the data
//	value := publicList.Get(0)
//	user := value.(*User)
//	// user.Name == "Alice"
//	// user.Balance == 100
//	// user.Internal == "" (sensitive data is filtered)
package rolist

import (
	"gno.land/p/nt/avl/v0/list"
	"gno.land/p/nt/avl/v0/rotree"
)

// IReadOnlyList defines the read-only operations available on a list.
type IReadOnlyList interface {
	Len() int
	Get(index int) any
	Slice(startIndex, endIndex int) []any
	ForEach(fn func(index int, value any) bool)
}

// ReadOnlyList wraps a list.List and provides read-only access.
type ReadOnlyList struct {
	list            *list.List
	makeEntrySafeFn func(any) any
}

// Verify interface implementations
var (
	_ IReadOnlyList = (*ReadOnlyList)(nil)
	_ IReadOnlyList = (interface{ list.IList })(nil) // is subset of list.IList
)

// Wrap creates a new ReadOnlyList from an existing list.List and a safety transformation function.
// If makeEntrySafeFn is nil, values will be returned as-is without transformation.
func Wrap(list *list.List, makeEntrySafeFn func(any) any) *ReadOnlyList {
	return &ReadOnlyList{
		list:            list,
		makeEntrySafeFn: makeEntrySafeFn,
	}
}

// getSafeValue applies the makeEntrySafeFn if it exists, otherwise returns the original value
func (rol *ReadOnlyList) getSafeValue(value any) any {
	if rol.makeEntrySafeFn == nil {
		return value
	}
	return rol.makeEntrySafeFn(value)
}

// Len returns the number of elements in the list.
func (rol *ReadOnlyList) Len() int {
	return rol.list.Len()
}

// Get returns the value at the specified index, converted to a safe format.
// Returns nil if index is out of bounds.
func (rol *ReadOnlyList) Get(index int) any {
	value := rol.list.Get(index)
	if value == nil {
		return nil
	}
	return rol.getSafeValue(value)
}

// Slice returns a slice of values from startIndex (inclusive) to endIndex (exclusive),
// with all values converted to a safe format.
func (rol *ReadOnlyList) Slice(startIndex, endIndex int) []any {
	values := rol.list.Slice(startIndex, endIndex)
	if values == nil {
		return nil
	}

	result := make([]any, len(values))
	for i, v := range values {
		result[i] = rol.getSafeValue(v)
	}
	return result
}

// ForEach iterates through all elements in the list, providing safe versions of the values.
func (rol *ReadOnlyList) ForEach(fn func(index int, value any) bool) {
	rol.list.ForEach(func(index int, value any) bool {
		return fn(index, rol.getSafeValue(value))
	})
}

// Tree returns a read-only pointer to the underlying AVL tree.
func (rol *ReadOnlyList) Tree() *rotree.ReadOnlyTree {
	return rol.list.Tree()
}
