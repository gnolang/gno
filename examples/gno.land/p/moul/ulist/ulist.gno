// Package ulist implements an ordered list using a modified AVL tree structure.
// While it maintains the self-balancing properties of an AVL tree for O(log n)
// operations, it differs in several key ways:
//
//  1. Each node can store multiple elements (like a B-tree), reducing the total
//     number of nodes and improving memory locality
//  2. Elements maintain a strict sequential order and can be accessed by index,
//     unlike a traditional AVL tree which is ordered but not indexed
//  3. The implementation provides list-like operations (append, insert at index)
//     rather than key-based operations
//
// When to use this over other data structures:
// - Over an AVL tree: when you need index-based access rather than key-based lookup
// - Over a slice: when you need efficient insertion/deletion at arbitrary positions
// - Over a linked list: when you need fast random access by index
//
// Performance characteristics:
// - O(log n) for random access, insertion, and deletion
// - O(1) for append operations
// - Better memory locality than a standard binary tree due to multiple elements per node
package ulist

import "errors"

// List represents an ordered sequence of elements with efficient random access,
// insertion, and deletion operations.
type List struct {
	root  *node
	count int
}

type node struct {
	elements []interface{}
	size_    int  // total elements in this subtree
	height_  int8 // for maintaining balance
	left     *node
	right    *node
	parent   *node // Added parent reference
}

// Common errors
var (
	ErrIndexOutOfRange = errors.New("index out of range")
	ErrInvalidArgument = errors.New("invalid argument")
)

// New creates an empty List.
func New() *List {
	return &List{}
}

// NewFromSlice creates a new List containing the provided elements.
func NewFromSlice(values []interface{}) *List {
	l := New()
	l.AppendMany(values...)
	return l
}

// Len returns the number of elements in the list.
func (l *List) Len() int {
	return l.count
}

// Get returns the element at the specified index.
func (l *List) Get(index int) (interface{}, error) {
	if index < 0 || index >= l.count {
		return nil, ErrIndexOutOfRange
	}
	node, localIndex := l.findNodeByIndex(index)
	return node.elements[localIndex], nil
}

// Set updates the element at the specified index.
func (l *List) Set(index int, value interface{}) error {
	if index < 0 || index >= l.count {
		return ErrIndexOutOfRange
	}
	node, localIndex := l.findNodeByIndex(index)
	node.elements[localIndex] = value
	return nil
}

// Append adds an element to the end of the list.
func (l *List) Append(value interface{}) {
	l.Insert(l.count, value)
}

// AppendMany adds multiple elements to the end of the list.
func (l *List) AppendMany(values ...interface{}) {
	for _, v := range values {
		l.Append(v)
	}
}

// Insert adds an element at the specified index.
func (l *List) Insert(index int, value interface{}) error {
	if index < 0 || index > l.count {
		return ErrIndexOutOfRange
	}
	l.root = l.insert(l.root, index, value)
	l.count++
	return nil
}

// Delete removes and returns the element at the specified index.
func (l *List) Delete(index int) (interface{}, error) {
	if index < 0 || index >= l.count {
		return nil, ErrIndexOutOfRange
	}
	var value interface{}
	l.root, value = l.delete(l.root, index)
	l.count--
	return value, nil
}

// Clear removes all elements from the list.
func (l *List) Clear() {
	l.root = nil
	l.count = 0
}

// Range returns a slice containing elements from start (inclusive) to end (exclusive).
func (l *List) Range(start, end int) ([]interface{}, error) {
	if start < 0 || end > l.count || start > end {
		return nil, ErrInvalidArgument
	}

	result := make([]interface{}, end-start)
	idx := 0

	l.ForEach(start, end, func(_ int, value interface{}) bool {
		result[idx] = value
		idx++
		return false
	})

	return result, nil
}

// ForEach executes fn for each element in the specified range.
func (l *List) ForEach(start, end int, fn func(index int, value interface{}) bool) error {
	if start < 0 || end > l.count || start > end {
		return ErrInvalidArgument
	}

	it := l.NewIterator()
	for i := 0; i < start; i++ {
		it.Next()
	}

	for i := start; i < end; i++ {
		value, ok := it.Next()
		if !ok {
			break
		}
		if fn(i, value) {
			break
		}
	}
	return nil
}

func max(a, b int8) int8 {
	if a > b {
		return a
	}
	return b
}

// Iterator provides sequential access to list elements.
type Iterator struct {
	list    *List
	current *node
	index   int
	pos     int
}

// NewIterator creates a new iterator for the list.
func (l *List) NewIterator() *Iterator {
	if l.root == nil {
		return &Iterator{list: l}
	}
	return &Iterator{
		list:    l,
		current: l.leftmost(l.root),
		index:   0,
		pos:     0,
	}
}

// Next advances to and returns the next element.
// Returns nil and false when iteration is complete.
func (it *Iterator) Next() (interface{}, bool) {
	if it.index >= it.list.count {
		return nil, false
	}

	if it.pos >= len(it.current.elements) {
		it.current = it.findNextNode()
		it.pos = 0
	}

	value := it.current.elements[it.pos]
	it.index++
	it.pos++
	return value, true
}

const (
	maxNodeSize = 16 // tunable parameter for node size
	minNodeSize = maxNodeSize / 2
)

// size returns the total number of elements in the subtree rooted at n
func (n *node) size() int {
	if n == nil {
		return 0
	}
	return n.size_
}

// height returns the height of the node in the tree
func (n *node) height() int8 {
	if n == nil {
		return 0
	}
	return n.height_
}

// balance returns the balance factor of the node
func (n *node) balance() int {
	return int(n.left.height() - n.right.height())
}

// updateMetrics updates the size and height of the node
func (n *node) updateMetrics() {
	n.size_ = len(n.elements) + n.left.size() + n.right.size()
	n.height_ = 1 + max(n.left.height(), n.right.height())
}

// findNodeByIndex locates the node containing the element at the given index
// and returns both the node and the local index within that node's elements.
func (l *List) findNodeByIndex(index int) (*node, int) {
	current := l.root
	pos := index

	for current != nil {
		leftSize := current.left.size()

		// Element is in left subtree
		if pos < leftSize {
			current = current.left
			continue
		}

		// Element is in current node
		localPos := pos - leftSize
		if localPos < len(current.elements) {
			return current, localPos
		}

		// Element is in right subtree
		pos = pos - leftSize - len(current.elements)
		current = current.right
	}

	return nil, 0 // Should never reach here if index is valid
}

// leftmost returns the leftmost node in the subtree
func (l *List) leftmost(n *node) *node {
	current := n
	for current != nil && current.left != nil {
		current = current.left
	}
	return current
}

// findNextNode returns the next node in in-order traversal
func (it *Iterator) findNextNode() *node {
	if it.current == nil {
		return nil
	}

	// If we have a right child, go to its leftmost node
	if it.current.right != nil {
		return it.list.leftmost(it.current.right)
	}

	// Otherwise, go up until we find a parent we haven't visited
	current := it.current
	for current.parent != nil && current == current.parent.right {
		current = current.parent
	}
	return current.parent
}

// insert adds a value at the specified index in the subtree rooted at n.
// Returns the new root of the subtree after insertion.
func (l *List) insert(n *node, index int, value interface{}) *node {
	if n == nil {
		newNode := &node{
			elements: []interface{}{value},
			size_:    1,
			height_:  1,
		}
		return newNode
	}

	leftSize := n.left.size()

	// Insert into left subtree
	if index < leftSize {
		n.left = l.insert(n.left, index, value)
		n.left.parent = n
	} else if index-leftSize < len(n.elements) {
		// Insert into current node
		pos := index - leftSize
		n.elements = append(n.elements[:pos], append([]interface{}{value}, n.elements[pos:]...)...)
	} else {
		// Insert into right subtree
		n.right = l.insert(n.right, index-leftSize-len(n.elements), value)
		n.right.parent = n
	}

	n.updateMetrics()
	return l.balance(n)
}

// balance maintains the AVL tree property by performing rotations if needed
func (l *List) balance(n *node) *node {
	if n == nil {
		return nil
	}

	balance := n.balance()

	// Left heavy
	if balance > 1 {
		if n.left.balance() < 0 {
			n.left = l.rotateLeft(n.left)
		}
		return l.rotateRight(n)
	}

	// Right heavy
	if balance < -1 {
		if n.right.balance() > 0 {
			n.right = l.rotateRight(n.right)
		}
		return l.rotateLeft(n)
	}

	return n
}

// rotateLeft performs a left rotation around node n
func (l *List) rotateLeft(n *node) *node {
	if n == nil || n.right == nil {
		return n
	}

	newRoot := n.right
	n.right = newRoot.left
	if n.right != nil {
		n.right.parent = n
	}
	newRoot.left = n
	newRoot.parent = n.parent
	n.parent = newRoot

	n.updateMetrics()
	newRoot.updateMetrics()

	return newRoot
}

// rotateRight performs a right rotation around node n
func (l *List) rotateRight(n *node) *node {
	if n == nil || n.left == nil {
		return n
	}

	newRoot := n.left
	n.left = newRoot.right
	if n.left != nil {
		n.left.parent = n
	}
	newRoot.right = n
	newRoot.parent = n.parent
	n.parent = newRoot

	n.updateMetrics()
	newRoot.updateMetrics()

	return newRoot
}

// delete removes and returns the element at the specified index in the subtree rooted at n.
// Returns the new root of the subtree after deletion and the deleted value.
func (l *List) delete(n *node, index int) (*node, interface{}) {
	if n == nil {
		return nil, nil
	}

	leftSize := n.left.size()

	// Value is in left subtree
	if index < leftSize {
		var value interface{}
		n.left, value = l.delete(n.left, index)
		if n.left != nil {
			n.left.parent = n
		}
		n.updateMetrics()
		return l.balance(n), value
	}

	// Value is in current node
	localIndex := index - leftSize
	if localIndex < len(n.elements) {
		value := n.elements[localIndex]
		n.elements = append(n.elements[:localIndex], n.elements[localIndex+1:]...)

		// If node is empty after deletion, remove it
		if len(n.elements) == 0 {
			if n.left == nil {
				if n.right != nil {
					n.right.parent = n.parent
				}
				return n.right, value
			}
			if n.right == nil {
				if n.left != nil {
					n.left.parent = n.parent
				}
				return n.left, value
			}

			// Node has both children, replace with successor
			successor := l.leftmost(n.right)
			n.elements = successor.elements
			n.right, _ = l.delete(n.right, 0)
			if n.right != nil {
				n.right.parent = n
			}
		}

		n.updateMetrics()
		return l.balance(n), value
	}

	// Value is in right subtree
	var value interface{}
	n.right, value = l.delete(n.right, index-leftSize-len(n.elements))
	if n.right != nil {
		n.right.parent = n
	}
	n.updateMetrics()
	return l.balance(n), value
}
