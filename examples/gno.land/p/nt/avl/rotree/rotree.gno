// Package rotree provides a read-only wrapper for avl.Tree with safe value transformation.
//
// It is useful when you want to expose a read-only view of a tree while ensuring that
// the sensitive data cannot be modified.
//
// Example:
//
//	// Define a user structure with sensitive data
//	type User struct {
//		Name     string
//		Balance  int
//		Internal string // sensitive field
//	}
//
//	// Create and populate the original tree
//	privateTree := avl.NewTree()
//	privateTree.Set("alice", &User{
//		Name:     "Alice",
//		Balance:  100,
//		Internal: "sensitive",
//	})
//
//	// Create a safe transformation function that copies the struct
//	// while excluding sensitive data
//	makeEntrySafeFn := func(v any) any {
//		u := v.(*User)
//		return &User{
//			Name:     u.Name,
//			Balance:  u.Balance,
//			Internal: "", // omit sensitive data
//		}
//	}
//
//	// Create a read-only view of the tree
//	PublicTree := rotree.Wrap(tree, makeEntrySafeFn)
//
//	// Safely access the data
//	value, _ := roTree.Get("alice")
//	user := value.(*User)
//	// user.Name == "Alice"
//	// user.Balance == 100
//	// user.Internal == "" (sensitive data is filtered)
package rotree

import (
	"gno.land/p/nt/avl"
)

// Wrap creates a new ReadOnlyTree from an existing avl.Tree and a safety transformation function.
// If makeEntrySafeFn is nil, values will be returned as-is without transformation.
//
// makeEntrySafeFn is a function that transforms a tree entry into a safe version that can be exposed to external users.
// This function should be implemented based on the specific safety requirements of your use case:
//
//  1. No-op transformation: For primitive types (int, string, etc.) or already safe objects,
//     simply pass nil as the makeEntrySafeFn to return values as-is.
//
//  2. Defensive copying: For mutable types like slices or maps, you should create a deep copy
//     to prevent modification of the original data.
//     Example: func(v any) any { return append([]int{}, v.([]int)...) }
//
//  3. Read-only wrapper: Return a read-only version of the object that implements
//     a limited interface.
//     Example: func(v any) any { return NewReadOnlyObject(v) }
//
//  4. DAO transformation: Transform the object into a data access object that
//     controls how the underlying data can be accessed.
//     Example: func(v any) any { return NewDAO(v) }
//
// The function ensures that the returned object is safe to expose to untrusted code,
// preventing unauthorized modifications to the original data structure.
func Wrap(tree *avl.Tree, makeEntrySafeFn func(any) any) *ReadOnlyTree {
	return &ReadOnlyTree{
		tree:            tree,
		makeEntrySafeFn: makeEntrySafeFn,
	}
}

// ReadOnlyTree wraps an avl.Tree and provides read-only access.
type ReadOnlyTree struct {
	tree            *avl.Tree
	makeEntrySafeFn func(any) any
}

// IReadOnlyTree defines the read-only operations available on a tree.
type IReadOnlyTree interface {
	Size() int
	Has(key string) bool
	Get(key string) (any, bool)
	GetByIndex(index int) (string, any)
	Iterate(start, end string, cb avl.IterCbFn) bool
	ReverseIterate(start, end string, cb avl.IterCbFn) bool
	IterateByOffset(offset int, count int, cb avl.IterCbFn) bool
	ReverseIterateByOffset(offset int, count int, cb avl.IterCbFn) bool
}

// Verify that ReadOnlyTree implements both ITree and IReadOnlyTree
var (
	_ avl.ITree     = (*ReadOnlyTree)(nil)
	_ IReadOnlyTree = (*ReadOnlyTree)(nil)
)

// getSafeValue applies the makeEntrySafeFn if it exists, otherwise returns the original value
func (roTree *ReadOnlyTree) getSafeValue(value any) any {
	if roTree.makeEntrySafeFn == nil {
		return value
	}
	return roTree.makeEntrySafeFn(value)
}

// Size returns the number of key-value pairs in the tree.
func (roTree *ReadOnlyTree) Size() int {
	return roTree.tree.Size()
}

// Has checks whether a key exists in the tree.
func (roTree *ReadOnlyTree) Has(key string) bool {
	return roTree.tree.Has(key)
}

// Get retrieves the value associated with the given key, converted to a safe format.
func (roTree *ReadOnlyTree) Get(key string) (any, bool) {
	value, exists := roTree.tree.Get(key)
	if !exists {
		return nil, false
	}
	return roTree.getSafeValue(value), true
}

// GetByIndex retrieves the key-value pair at the specified index in the tree, with the value converted to a safe format.
func (roTree *ReadOnlyTree) GetByIndex(index int) (string, any) {
	key, value := roTree.tree.GetByIndex(index)
	return key, roTree.getSafeValue(value)
}

// Iterate performs an in-order traversal of the tree within the specified key range.
func (roTree *ReadOnlyTree) Iterate(start, end string, cb avl.IterCbFn) bool {
	return roTree.tree.Iterate(start, end, func(key string, value any) bool {
		return cb(key, roTree.getSafeValue(value))
	})
}

// ReverseIterate performs a reverse in-order traversal of the tree within the specified key range.
func (roTree *ReadOnlyTree) ReverseIterate(start, end string, cb avl.IterCbFn) bool {
	return roTree.tree.ReverseIterate(start, end, func(key string, value any) bool {
		return cb(key, roTree.getSafeValue(value))
	})
}

// IterateByOffset performs an in-order traversal of the tree starting from the specified offset.
func (roTree *ReadOnlyTree) IterateByOffset(offset int, count int, cb avl.IterCbFn) bool {
	return roTree.tree.IterateByOffset(offset, count, func(key string, value any) bool {
		return cb(key, roTree.getSafeValue(value))
	})
}

// ReverseIterateByOffset performs a reverse in-order traversal of the tree starting from the specified offset.
func (roTree *ReadOnlyTree) ReverseIterateByOffset(offset int, count int, cb avl.IterCbFn) bool {
	return roTree.tree.ReverseIterateByOffset(offset, count, func(key string, value any) bool {
		return cb(key, roTree.getSafeValue(value))
	})
}

// Set is not supported on ReadOnlyTree and will panic.
func (roTree *ReadOnlyTree) Set(key string, value any) bool {
	panic("Set operation not supported on ReadOnlyTree")
}

// Remove is not supported on ReadOnlyTree and will panic.
func (roTree *ReadOnlyTree) Remove(key string) (value any, removed bool) {
	panic("Remove operation not supported on ReadOnlyTree")
}

// RemoveByIndex is not supported on ReadOnlyTree and will panic.
func (roTree *ReadOnlyTree) RemoveByIndex(index int) (key string, value any) {
	panic("RemoveByIndex operation not supported on ReadOnlyTree")
}
