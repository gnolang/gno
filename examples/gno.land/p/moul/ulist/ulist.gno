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

type List struct {
	tree        avl.ITree
	count       int
	maxNodeSize int
}

func New() *List {
	return &List{
		tree:        avl.NewTree(),
		maxNodeSize: 16,
	}
}

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

type nodeData struct {
	elements []interface{}
	size     int
}

func (l *List) Len() int {
	return l.count
}

func (l *List) Get(index int) (interface{}, error) {
	if index < 0 || index >= l.count {
		return nil, ErrIndexOutOfRange
	}

	nodeIndex := index / l.maxNodeSize
	key := seqid.ID(nodeIndex).String()

	value, exists := l.tree.Get(key)
	if !exists {
		return nil, ErrIndexOutOfRange
	}

	node := value.(*nodeData)
	localIndex := index % l.maxNodeSize
	return node.elements[localIndex], nil
}

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
	l.tree.Set(key, node)

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

func (l *List) Insert(index int, value interface{}) error {
	if index < 0 || index > l.count {
		return ErrIndexOutOfRange
	}

	nodeIndex := index / l.maxNodeSize
	key := seqid.ID(nodeIndex).String()
	localIndex := index % l.maxNodeSize

	if index == l.count && localIndex > 0 {
		lastNodeI, exists := l.tree.Get(key)
		if exists {
			lastNode := lastNodeI.(*nodeData)
			if lastNode.size < l.maxNodeSize {
				lastNode.elements = append(lastNode.elements, value)
				lastNode.size++
				l.tree.Set(key, lastNode)
				l.count++
				return nil
			}
		}
	}

	var node *nodeData
	nodeI, exists := l.tree.Get(key)
	if exists {
		node = nodeI.(*nodeData)
		if node.size < l.maxNodeSize {
			node.elements = append(node.elements[:localIndex], append([]interface{}{value}, node.elements[localIndex:]...)...)
			node.size++
			l.tree.Set(key, node)
			l.count++
			return nil
		}
	} else {
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

	if node.size == 1 {
		l.tree.Remove(key)
	} else {
		node.elements = append(node.elements[:localIndex], node.elements[localIndex+1:]...)
		node.size--
		l.tree.Set(key, node)
	}

	l.count--
	return value, nil
}

func (l *List) Range(start, end int) ([]interface{}, error) {
	if start < 0 || end > l.count || start > end {
		return nil, ErrInvalidArgument
	}

	if start == end {
		return []interface{}{}, nil
	}

	result := make([]interface{}, 0, end-start)

	startNode := start / l.maxNodeSize
	endNode := (end - 1) / l.maxNodeSize

	for nodeIndex := startNode; nodeIndex <= endNode; nodeIndex++ {
		key := seqid.ID(nodeIndex).String()
		nodeI, exists := l.tree.Get(key)
		if !exists {
			return nil, ErrInvalidArgument
		}

		node := nodeI.(*nodeData)

		nodeStart := 0
		if nodeIndex == startNode {
			nodeStart = start % l.maxNodeSize
		}

		nodeEnd := node.size
		if nodeIndex == endNode {
			nodeEnd = ((end - 1) % l.maxNodeSize) + 1
		}

		result = append(result, node.elements[nodeStart:nodeEnd]...)
	}

	return result, nil
}

type Iterator struct {
	list       *List
	nodeIndex  int
	localIndex int
	done       bool
}

func (l *List) NewIterator() *Iterator {
	return &Iterator{
		list:       l,
		nodeIndex:  0,
		localIndex: 0,

		done: l.count == 0,
	}
}

func (it *Iterator) Next() (interface{}, bool) {
	if it.done {
		return nil, false
	}

	key := seqid.ID(it.nodeIndex).String()
	nodeI, exists := it.list.tree.Get(key)
	if !exists {
		it.done = true
		return nil, false
	}

	node := nodeI.(*nodeData)
	value := node.elements[it.localIndex]

	it.localIndex++
	if it.localIndex >= node.size {
		it.nodeIndex++
		it.localIndex = 0
	}

	totalIndex := it.nodeIndex*it.list.maxNodeSize + it.localIndex
	if totalIndex >= it.list.count {
		it.done = true
	}

	return value, true
}

func (l *List) Clear() {
	l.tree = avl.NewTree()
	l.count = 0
}
