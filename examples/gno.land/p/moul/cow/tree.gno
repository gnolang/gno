// Package cow provides a Copy-on-Write (CoW) AVL tree implementation.
// This is a fork of gno.land/p/nt/avl/v0 that adds CoW functionality
// while maintaining the original AVL tree interface and properties.
//
// Copy-on-Write creates a copy of a data structure only when it is modified,
// while still presenting the appearance of a full copy. When a tree is cloned,
// it initially shares all its nodes with the original tree. Only when a
// modification is made to either the original or the clone are new nodes created,
// and only along the path from the root to the modified node.
//
// Key features:
//   - O(1) cloning operation
//   - Minimal memory usage through structural sharing
//   - Full AVL tree functionality (self-balancing, ordered operations)
//   - Thread-safe for concurrent reads of shared structures
//
// While the CoW mechanism handles structural copying automatically, users need
// to consider how to handle the values stored in the tree:
//
//  1. Simple Values (int, string, etc.):
//     - These are copied by value automatically
//     - No additional handling needed
//
//  2. Complex Values (structs, pointers):
//     - Only the reference is copied by default
//     - Users must implement their own deep copy mechanism if needed
//
// Example:
//
//	// Create original tree
//	original := cow.NewTree()
//	original.Set("key1", "value1")
//
//	// Create a clone - O(1) operation
//	clone := original.Clone()
//
//	// Modify clone - only affected nodes are copied
//	clone.Set("key1", "modified")
//
//	// Original remains unchanged
//	val, _ := original.Get("key1") // Returns "value1"
package cow

type IterCbFn func(key string, value any) bool

//----------------------------------------
// Tree

// The zero struct can be used as an empty tree.
type Tree struct {
	node *Node
}

// NewTree creates a new empty AVL tree.
func NewTree() *Tree {
	return &Tree{
		node: nil,
	}
}

// Size returns the number of key-value pair in the tree.
func (tree *Tree) Size() int {
	return tree.node.Size()
}

// Has checks whether a key exists in the tree.
// It returns true if the key exists, otherwise false.
func (tree *Tree) Has(key string) (has bool) {
	return tree.node.Has(key)
}

// Get retrieves the value associated with the given key.
// It returns the value and a boolean indicating whether the key exists.
func (tree *Tree) Get(key string) (value any, exists bool) {
	_, value, exists = tree.node.Get(key)
	return
}

// GetByIndex retrieves the key-value pair at the specified index in the tree.
// It returns the key and value at the given index.
func (tree *Tree) GetByIndex(index int) (key string, value any) {
	return tree.node.GetByIndex(index)
}

// Set inserts a key-value pair into the tree.
// If the key already exists, the value will be updated.
// It returns a boolean indicating whether the key was newly inserted or updated.
func (tree *Tree) Set(key string, value any) (updated bool) {
	newnode, updated := tree.node.Set(key, value)
	tree.node = newnode
	return updated
}

// Remove removes a key-value pair from the tree.
// It returns the removed value and a boolean indicating whether the key was found and removed.
func (tree *Tree) Remove(key string) (value any, removed bool) {
	newnode, _, value, removed := tree.node.Remove(key)
	tree.node = newnode
	return value, removed
}

// Iterate performs an in-order traversal of the tree within the specified key range.
// It calls the provided callback function for each key-value pair encountered.
// If the callback returns true, the iteration is stopped.
func (tree *Tree) Iterate(start, end string, cb IterCbFn) bool {
	return tree.node.TraverseInRange(start, end, true, true,
		func(node *Node) bool {
			return cb(node.Key(), node.Value())
		},
	)
}

// ReverseIterate performs a reverse in-order traversal of the tree within the specified key range.
// It calls the provided callback function for each key-value pair encountered.
// If the callback returns true, the iteration is stopped.
func (tree *Tree) ReverseIterate(start, end string, cb IterCbFn) bool {
	return tree.node.TraverseInRange(start, end, false, true,
		func(node *Node) bool {
			return cb(node.Key(), node.Value())
		},
	)
}

// IterateByOffset performs an in-order traversal of the tree starting from the specified offset.
// It calls the provided callback function for each key-value pair encountered, up to the specified count.
// If the callback returns true, the iteration is stopped.
func (tree *Tree) IterateByOffset(offset int, count int, cb IterCbFn) bool {
	return tree.node.TraverseByOffset(offset, count, true, true,
		func(node *Node) bool {
			return cb(node.Key(), node.Value())
		},
	)
}

// ReverseIterateByOffset performs a reverse in-order traversal of the tree starting from the specified offset.
// It calls the provided callback function for each key-value pair encountered, up to the specified count.
// If the callback returns true, the iteration is stopped.
func (tree *Tree) ReverseIterateByOffset(offset int, count int, cb IterCbFn) bool {
	return tree.node.TraverseByOffset(offset, count, false, true,
		func(node *Node) bool {
			return cb(node.Key(), node.Value())
		},
	)
}

// Equal returns true if the two trees contain the same key-value pairs.
// WARNING: This is an expensive operation that recursively traverses the entire tree structure.
// It should only be used in tests or when absolutely necessary.
func (tree *Tree) Equal(other *Tree) bool {
	if tree == nil || other == nil {
		return tree == other
	}
	return tree.node.Equal(other.node)
}

// Clone creates a shallow copy of the tree
func (tree *Tree) Clone() *Tree {
	if tree == nil {
		return nil
	}
	return &Tree{
		node: tree.node,
	}
}
