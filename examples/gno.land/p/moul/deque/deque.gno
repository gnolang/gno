// Package deque provides a doubly-linked deque with optional size limits,
// optimized for minimal storage updates and O(1) operations in Gno environments.
//
// This implementation uses a doubly-linked list structure that provides:
// - O(1) push back operations
// - O(1) pop front operations
// - O(1) access to first/last elements
// - No array resizing or slice shifting
// - Minimal allocations (one node per element)
// - Minimal storage updates (typically 1-2 pointer updates per operation)
// - Optional max-size eviction policy
//
// Optimized for Gno blockchain storage: each operation updates only specific nodes
// rather than entire data structures, making it ideal for storage-efficient lists
// that need simple read operations like first/last access, size checks, list
// enumeration, and iteration with minimal storage update overhead.
//
// Example usage:
//
//	// Unbounded deque
//	d := deque.New()
//	d.PushBack("a", "b", "c")
//
//	// Bounded deque with automatic eviction
//	d = deque.NewBounded(3)
//	d.PushBack("a", "b", "c", "d")  // "a" gets evicted
//
//	first := d.PopFront()  // Returns "b"
//	last := d.Last()       // Returns "d"
//
//	d.Enumerate(func(i int, val any) bool {
//	    println(i, val)
//	    return false  // continue
//	})
package deque

// Node represents a single element in the doubly-linked list
type Node struct {
	value any
	prev  *Node
	next  *Node
}

// Deque represents a doubly-linked deque with optional size limits
type Deque struct {
	head *Node
	tail *Node
	size int
	max  int // 0 = unlimited
}

// New creates a new unbounded deque.
func New() *Deque {
	return &Deque{
		max: 0, // unlimited
	}
}

// NewBounded creates a new bounded deque with the specified maximum size.
// When the deque exceeds maxSize, elements are removed from the front.
func NewBounded(maxSize int) *Deque {
	if maxSize <= 0 {
		maxSize = 1
	}
	return &Deque{
		max: maxSize,
	}
}

// PushBack adds one or more elements to the back of the deque.
// If the deque exceeds maxSize, elements are removed from the front.
func (d *Deque) PushBack(items ...any) {
	for _, item := range items {
		d.pushBackSingle(item)
	}
}

// pushBackSingle adds a single element to the back
func (d *Deque) pushBackSingle(item any) {
	node := &Node{value: item}

	if d.tail == nil {
		// First element
		d.head = node
		d.tail = node
	} else {
		// Link to existing tail
		d.tail.next = node
		node.prev = d.tail
		d.tail = node
	}

	d.size++

	// Handle size limit by removing from front
	if d.max > 0 && d.size > d.max {
		d.popFrontNode()
	}
}

// PopFront removes and returns the front element.
// Returns nil if the deque is empty.
func (d *Deque) PopFront() any {
	if d.head == nil {
		return nil
	}

	value := d.head.value
	d.popFrontNode()
	return value
}

// popFrontNode removes the front node
func (d *Deque) popFrontNode() {
	if d.head == nil {
		return
	}

	if d.head == d.tail {
		// Only one element
		d.head = nil
		d.tail = nil
	} else {
		// Move head forward
		d.head = d.head.next
		d.head.prev = nil
	}

	d.size--
}

// PopBack removes and returns the back element.
// Returns nil if the deque is empty.
func (d *Deque) PopBack() any {
	if d.tail == nil {
		return nil
	}

	value := d.tail.value
	d.popBackNode()
	return value
}

// popBackNode removes the back node
func (d *Deque) popBackNode() {
	if d.tail == nil {
		return
	}

	if d.head == d.tail {
		// Only one element
		d.head = nil
		d.tail = nil
	} else {
		// Move tail backward
		d.tail = d.tail.prev
		d.tail.next = nil
	}

	d.size--
}

// PushFront adds one or more elements to the front of the deque.
// If bounded and size limit is exceeded, elements are removed from the back.
func (d *Deque) PushFront(items ...any) {
	for _, item := range items {
		d.pushFrontSingle(item)
	}
}

// pushFrontSingle adds a single element to the front
func (d *Deque) pushFrontSingle(item any) {
	node := &Node{value: item}

	if d.head == nil {
		// First element
		d.head = node
		d.tail = node
	} else {
		// Link to existing head
		d.head.prev = node
		node.next = d.head
		d.head = node
	}

	d.size++

	// Handle size limit by removing from back
	if d.max > 0 && d.size > d.max {
		d.popBackNode()
	}
}

// Size returns the current number of elements
func (d *Deque) Size() int {
	return d.size
}

// MaxSize returns the maximum size limit (0 = unlimited)
func (d *Deque) MaxSize() int {
	return d.max
}

// IsEmpty returns true if the deque is empty
func (d *Deque) IsEmpty() bool {
	return d.size == 0
}

// IsBounded returns true if the deque has a size limit
func (d *Deque) IsBounded() bool {
	return d.max > 0
}

// First returns the front element without removing it.
// Returns nil if the deque is empty.
func (d *Deque) First() any {
	if d.head == nil {
		return nil
	}
	return d.head.value
}

// Last returns the back element without removing it.
// Returns nil if the deque is empty.
func (d *Deque) Last() any {
	if d.tail == nil {
		return nil
	}
	return d.tail.value
}

// Get returns the element at the specified index (0 = front).
// Returns nil if index is out of bounds.
func (d *Deque) Get(index int) any {
	if index < 0 || index >= d.size {
		return nil
	}

	node := d.head
	for i := 0; i < index; i++ {
		node = node.next
	}
	return node.value
}

// List returns all elements as a slice, ordered from front to back
func (d *Deque) List() []any {
	if d.size == 0 {
		return nil
	}

	result := make([]any, d.size)
	node := d.head
	for i := 0; i < d.size; i++ {
		result[i] = node.value
		node = node.next
	}
	return result
}

// Enumerate calls fn for each element with its index (0-based from front)
func (d *Deque) Enumerate(fn func(index int, value any) bool) {
	node := d.head
	for i := 0; i < d.size; i++ {
		if fn(i, node.value) {
			break
		}
		node = node.next
	}
}

// Clear removes all elements from the deque
func (d *Deque) Clear() {
	d.head = nil
	d.tail = nil
	d.size = 0
}

// SetMaxSize changes the maximum size limit.
// If the new limit is smaller than current size, elements are removed from front.
func (d *Deque) SetMaxSize(maxSize int) {
	d.max = maxSize

	if maxSize > 0 {
		for d.size > maxSize {
			d.popFrontNode()
		}
	}
}

// Iterator returns a new iterator for forward traversal
func (d *Deque) Iterator() *Iterator {
	return &Iterator{
		current: d.head,
	}
}

// ReverseIterator returns a new iterator for backward traversal
func (d *Deque) ReverseIterator() *ReverseIterator {
	return &ReverseIterator{
		current: d.tail,
	}
}

// Iterator provides forward traversal from front to back
type Iterator struct {
	current *Node
}

// Next returns the next element or nil when complete
func (it *Iterator) Next() any {
	if it.current == nil {
		return nil
	}

	value := it.current.value
	it.current = it.current.next
	return value
}

// HasNext returns true if more elements are available
func (it *Iterator) HasNext() bool {
	return it.current != nil
}

// ReverseIterator provides backward traversal from back to front
type ReverseIterator struct {
	current *Node
}

// Next returns the next element (moving backward) or nil when complete
func (it *ReverseIterator) Next() any {
	if it.current == nil {
		return nil
	}

	value := it.current.value
	it.current = it.current.prev
	return value
}

// HasNext returns true if more elements are available
func (it *ReverseIterator) HasNext() bool {
	return it.current != nil
}
