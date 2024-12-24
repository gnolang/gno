// Package ulist implements an ordered list using a hybrid data structure that combines
// aspects of B-trees and AVL trees. It stores elements in fixed-size nodes (chunks)
// organized in an AVL tree, where each node can hold up to maxNodeSize elements.
//
// Implementation details:
//   - Elements are stored in chunks of size maxNodeSize (default: 16)
//   - Chunks are indexed using seqid-formatted keys in an AVL tree
//   - Each chunk maintains a size counter for O(1) capacity checks
//   - Elements maintain strict sequential ordering
//
// Performance characteristics:
//   - Get/Set: O(log n) for tree traversal + O(1) for chunk access
//   - Insert/Delete: O(log n) for tree traversal + O(k) for chunk manipulation
//   - Append: O(1) amortized when last chunk has space, O(log n) when creating new chunk
//   - Memory: Better locality than AVL tree, more overhead than slice
//   - Space: O(n) elements + O(n/maxNodeSize) tree nodes
//
// Storage costs:
//   - Each modification to a chunk requires rewriting the entire chunk
//   - Smaller chunk sizes reduce the cost of modifications but increase tree overhead
//   - Larger chunk sizes improve read performance but increase write amplification
//   - Default chunk size (16) balances between read performance and write costs
//
// When to use:
//   - Over AVL tree: when needing index-based access or sequential iteration
//   - Over slice: when needing efficient insertion/deletion at arbitrary positions
//     or when frequent modifications to small portions would cause full slice rewrites
//   - Over linked list: when needing fast random access
//
// Example usage:
//
//	list := ulist.New()
//
//	// Add elements
//	list.Append(1)
//	list.Append(2)
//	list.Insert(1, 1.5)
//
//	// Access elements
//	first, _ := list.Get(0)    // returns 1
//	middle, _ := list.Get(1)   // returns 1.5
//	last, _ := list.Get(2)     // returns 2
//
//	// Delete elements
//	list.Delete(1)             // removes 1.5
//
//	// Get a range
//	elements, _ := list.Range(0, list.Len())
package ulist

import (
	"errors"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/seqid"
)

// List represents an ordered list using a hybrid B-tree/AVL structure.
// It stores elements in fixed-size nodes (chunks) organized in an AVL tree.
type List struct {
	tree        avl.ITree
	count       int
	maxNodeSize int
}

// New creates a new List with the default node size of 16.
// This size provides a good balance between read performance and write amplification.
func New() *List {
	return &List{
		tree:        avl.NewTree(),
		maxNodeSize: 16,
	}
}

// NewWithSize creates a new List with a custom node size.
// The size must be even and at least 4 to maintain efficient operations.
// Larger sizes improve read performance but increase write amplification.
// Smaller sizes reduce write amplification but increase tree overhead.
func NewWithSize(size int) *List {
	// ensure size is even
	if size%2 != 0 {
		size++
	}

	// ensure size is at least 4
	if size < 4 {
		size = 4
	}
	return &List{
		tree:        avl.NewTree(),
		maxNodeSize: size,
	}
}

var (
	ErrIndexOutOfRange = errors.New("index out of range")
	ErrInvalidArgument = errors.New("invalid argument")
)

// nodeData represents a chunk of elements in the list.
// Elements are stored in a slice to provide efficient sequential access.
type nodeData struct {
	elements []interface{}
	size     int
}

func (l *List) Len() int {
	return l.count
}

// Get retrieves the element at the specified index.
// Performance: O(log n) for tree traversal + O(1) for chunk access.
func (l *List) Get(index int) (interface{}, error) {
	if index < 0 || index >= l.count {
		return nil, ErrIndexOutOfRange
	}

	// Calculate which node contains our index
	nodeIndex := index / l.maxNodeSize
	key := seqid.ID(nodeIndex).String()

	value, exists := l.tree.Get(key)
	if !exists {
		return nil, ErrIndexOutOfRange
	}

	// Calculate the local index within the node
	node := value.(*nodeData)
	localIndex := index % l.maxNodeSize
	return node.elements[localIndex], nil
}

// Set updates the element at the specified index.
// Performance: O(log n) for tree traversal + O(maxNodeSize) for chunk update.
// Storage: Requires rewriting the entire chunk even for a single element update.
func (l *List) Set(index int, value interface{}) error {
	if index < 0 || index >= l.count {
		return ErrIndexOutOfRange
	}

	nodeIndex := index / l.maxNodeSize
	key := seqid.ID(nodeIndex).String()

	nodeI, exists := l.tree.Get(key)
	if !exists {
		return ErrIndexOutOfRange
	}

	node := nodeI.(*nodeData)
	localIndex := index % l.maxNodeSize
	node.elements[localIndex] = value
	l.tree.Set(key, node) // Requires rewriting the entire chunk

	return nil
}

func (l *List) Append(value interface{}) {
	l.Insert(l.count, value)
}

func (l *List) AppendMany(values ...interface{}) {
	for _, v := range values {
		l.Append(v)
	}
}

// Insert adds an element at the specified index, shifting existing elements right.
// Performance: O(log n) for tree traversal + O(maxNodeSize) for chunk manipulation.
// Storage: May require rewriting multiple chunks if the target chunk is full.
func (l *List) Insert(index int, value interface{}) error {
	if index < 0 || index > l.count {
		return ErrIndexOutOfRange
	}

	nodeIndex := index / l.maxNodeSize
	key := seqid.ID(nodeIndex).String()
	localIndex := index % l.maxNodeSize

	// Optimization: If inserting at the end and last chunk has space
	if index == l.count && localIndex > 0 {
		lastNodeI, exists := l.tree.Get(key)
		if exists {
			lastNode := lastNodeI.(*nodeData)
			if lastNode.size < l.maxNodeSize {
				// Fast path: append to existing chunk
				lastNode.elements = append(lastNode.elements, value)
				lastNode.size++
				l.tree.Set(key, lastNode)
				l.count++
				return nil
			}
		}
	}

	// Normal insertion path
	var node *nodeData
	nodeI, exists := l.tree.Get(key)
	if exists {
		node = nodeI.(*nodeData)
		if node.size < l.maxNodeSize {
			// Insert into existing chunk with space
			node.elements = append(node.elements[:localIndex], append([]interface{}{value}, node.elements[localIndex:]...)...)
			node.size++
			l.tree.Set(key, node)
			l.count++
			return nil
		}
	} else {
		// Create new chunk for sparse insertions
		node = &nodeData{
			elements: []interface{}{value},
			size:     1,
		}
		l.tree.Set(key, node)
		l.count++
		return nil
	}

	return ErrInvalidArgument
}

// Delete removes and returns the element at the specified index.
// Performance: O(log n) for tree traversal + O(maxNodeSize) for chunk update.
// Storage: Requires rewriting the chunk even when removing a single element.
// If the chunk becomes empty, it is removed from the tree entirely.
func (l *List) Delete(index int) (interface{}, error) {
	if index < 0 || index >= l.count {
		return nil, ErrIndexOutOfRange
	}

	nodeIndex := index / l.maxNodeSize
	key := seqid.ID(nodeIndex).String()
	localIndex := index % l.maxNodeSize

	nodeI, exists := l.tree.Get(key)
	if !exists {
		return nil, ErrIndexOutOfRange
	}

	node := nodeI.(*nodeData)
	value := node.elements[localIndex]

	// If this is the last element in the chunk, remove the entire chunk
	if node.size == 1 {
		l.tree.Remove(key)
	} else {
		// Otherwise, remove the element and shift remaining elements left
		node.elements = append(node.elements[:localIndex], node.elements[localIndex+1:]...)
		node.size--
		l.tree.Set(key, node)
	}

	l.count--
	return value, nil
}

// Range returns a slice of elements from start (inclusive) to end (exclusive).
// Performance: O(log n) for tree traversal + O(end-start) for copying elements.
// Storage: Creates a new slice containing the requested range of elements.
func (l *List) Range(start, end int) ([]interface{}, error) {
	if start < 0 || end > l.count || start > end {
		return nil, ErrInvalidArgument
	}

	if start == end {
		return []interface{}{}, nil
	}

	// Pre-allocate slice with exact size needed
	result := make([]interface{}, 0, end-start)

	// Calculate start and end nodes
	startNode := start / l.maxNodeSize
	endNode := (end - 1) / l.maxNodeSize

	// Iterate through all nodes that contain our range
	for nodeIndex := startNode; nodeIndex <= endNode; nodeIndex++ {
		key := seqid.ID(nodeIndex).String()
		nodeI, exists := l.tree.Get(key)
		if !exists {
			return nil, ErrInvalidArgument
		}

		node := nodeI.(*nodeData)

		// Calculate start and end indices within this node
		nodeStart := 0
		if nodeIndex == startNode {
			nodeStart = start % l.maxNodeSize
		}

		nodeEnd := node.size
		if nodeIndex == endNode {
			nodeEnd = ((end - 1) % l.maxNodeSize) + 1
		}

		// Append the relevant portion of this node's elements
		result = append(result, node.elements[nodeStart:nodeEnd]...)
	}

	return result, nil
}

// Iterator provides sequential access to list elements.
// It maintains its position using nodeIndex and localIndex,
// traversing through chunks in order.
type Iterator struct {
	list       *List
	nodeIndex  int  // Current chunk index
	localIndex int  // Current position within chunk
	done       bool // Indicates if iteration is complete
}

// NewIterator creates an iterator starting at the beginning of the list.
// The iterator will traverse elements in order from index 0 to Len()-1.
func (l *List) NewIterator() *Iterator {
	return &Iterator{
		list:       l,
		nodeIndex:  0,
		localIndex: 0,
		done:       l.count == 0,
	}
}

// Next returns the next element in the iteration and a boolean indicating
// if the iteration is complete. Returns (nil, false) when done.
// Performance: O(1) amortized, as tree traversal only occurs when moving
// to a new chunk.
func (it *Iterator) Next() (interface{}, bool) {
	if it.done {
		return nil, false
	}

	// Get current chunk
	key := seqid.ID(it.nodeIndex).String()
	nodeI, exists := it.list.tree.Get(key)
	if !exists {
		it.done = true
		return nil, false
	}

	node := nodeI.(*nodeData)
	value := node.elements[it.localIndex]

	// Move to next position
	it.localIndex++
	if it.localIndex >= node.size {
		it.nodeIndex++
		it.localIndex = 0
	}

	// Check if we've reached the end
	totalIndex := it.nodeIndex*it.list.maxNodeSize + it.localIndex
	if totalIndex >= it.list.count {
		it.done = true
	}

	return value, true
}

// Clear removes all elements from the list.
// Performance: O(1) as it simply creates a new empty tree.
// Storage: Allows garbage collection of all nodes and elements.
func (l *List) Clear() {
	l.tree = avl.NewTree()
	l.count = 0
}
