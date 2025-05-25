// Package fifo implements a fixed-size FIFO (First-In-First-Out) list data structure
// using a singly-linked list. The implementation prioritizes storage efficiency by minimizing
// storage operations - each add/remove operation only updates 1-2 pointers, regardless of
// list size.
//
// Key features:
// - Fixed-size with automatic removal of oldest entries when full
// - Support for both prepend (add at start) and append (add at end) operations
// - Constant storage usage through automatic pruning
// - O(1) append operations and latest element access
// - Iterator support for sequential access
// - Dynamic size adjustment via SetMaxSize
//
// This implementation is optimized for frequent updates, as insertions and deletions only
// require updating 1-2 pointers. However, random access operations are O(n) as they require
// traversing the list. For use cases where writes are rare, a slice-based
// implementation might be more suitable.
//
// The linked list structure is equally efficient for storing both small values (like pointers)
// and larger data structures, as each node maintains a single next-pointer regardless of the
// stored value's size.
//
// Example usage:
//
//	list := fifo.New(3)        // Create a new list with max size 3
//	list.Append("a")           // List: [a]
//	list.Append("b")           // List: [a b]
//	list.Append("c")           // List: [a b c]
//	list.Append("d")           // List: [b c d] (oldest element "a" was removed)
//	latest := list.Latest()    // Returns "d"
//	all := list.Entries()      // Returns ["b", "c", "d"]
package fifo

// node represents a single element in the linked list
type node struct {
	value any
	next  *node
}

// List represents a fixed-size FIFO list
type List struct {
	head    *node
	tail    *node
	size    int
	maxSize int
}

// New creates a new FIFO list with the specified maximum size
func New(maxSize int) *List {
	return &List{
		maxSize: maxSize,
	}
}

// Prepend adds a new entry at the start of the list. If the list exceeds maxSize,
// the last entry is automatically removed.
func (l *List) Prepend(entry any) {
	if l.maxSize == 0 {
		return
	}

	newNode := &node{value: entry}

	if l.head == nil {
		l.head = newNode
		l.tail = newNode
		l.size = 1
		return
	}

	newNode.next = l.head
	l.head = newNode

	if l.size < l.maxSize {
		l.size++
		return
	}

	// Remove last element by traversing to second-to-last
	if l.size == 1 {
		// Special case: if size is 1, just update both pointers
		l.head = newNode
		l.tail = newNode
		newNode.next = nil
		return

	}

	// Find second-to-last node
	current := l.head
	for current.next != l.tail {
		current = current.next
	}
	current.next = nil
	l.tail = current

}

// Append adds a new entry at the end of the list. If the list exceeds maxSize,
// the first entry is automatically removed.
func (l *List) Append(entry any) {
	if l.maxSize == 0 {
		return
	}

	newNode := &node{value: entry}

	if l.head == nil {
		l.head = newNode
		l.tail = newNode
		l.size = 1
		return
	}

	l.tail.next = newNode
	l.tail = newNode

	if l.size < l.maxSize {
		l.size++
	} else {
		l.head = l.head.next
	}
}

// Get returns the entry at the specified index.
// Index 0 is the oldest entry, Size()-1 is the newest.
func (l *List) Get(index int) any {
	if index < 0 || index >= l.size {
		return nil
	}

	current := l.head
	for i := 0; i < index; i++ {
		current = current.next
	}
	return current.value
}

// Size returns the current number of entries in the list
func (l *List) Size() int {
	return l.size
}

// MaxSize returns the maximum size configured for this list
func (l *List) MaxSize() int {
	return l.maxSize
}

// Entries returns all current entries as a slice
func (l *List) Entries() []any {
	entries := make([]any, l.size)
	current := l.head
	for i := 0; i < l.size; i++ {
		entries[i] = current.value
		current = current.next
	}
	return entries
}

// Iterator returns a function that can be used to iterate over the entries
// from oldest to newest. Returns nil when there are no more entries.
func (l *List) Iterator() func() any {
	current := l.head
	return func() any {
		if current == nil {
			return nil
		}
		value := current.value
		current = current.next
		return value
	}
}

// Latest returns the most recent entry.
// Returns nil if the list is empty.
func (l *List) Latest() any {
	if l.tail == nil {
		return nil
	}
	return l.tail.value
}

// SetMaxSize updates the maximum size of the list.
// If the new maxSize is smaller than the current size,
// the oldest entries are removed to fit the new size.
func (l *List) SetMaxSize(maxSize int) {
	if maxSize < 0 {
		maxSize = 0
	}

	// If new maxSize is smaller than current size,
	// remove oldest entries until we fit
	if maxSize < l.size {
		// Special case: if new maxSize is 0, clear the list
		if maxSize == 0 {
			l.head = nil
			l.tail = nil
			l.size = 0
		} else {
			// Keep the newest entries by moving head forward
			diff := l.size - maxSize
			for i := 0; i < diff; i++ {
				l.head = l.head.next
			}
			l.size = maxSize
		}
	}

	l.maxSize = maxSize
}

// Delete removes the element at the specified index.
// Returns true if an element was removed, false if the index was invalid.
func (l *List) Delete(index int) bool {
	if index < 0 || index >= l.size {
		return false
	}

	// Special case: deleting the only element
	if l.size == 1 {
		l.head = nil
		l.tail = nil
		l.size = 0
		return true
	}

	// Special case: deleting first element
	if index == 0 {
		l.head = l.head.next
		l.size--
		return true
	}

	// Find the node before the one to delete
	current := l.head
	for i := 0; i < index-1; i++ {
		current = current.next
	}

	// Special case: deleting last element
	if index == l.size-1 {
		l.tail = current
		current.next = nil
	} else {
		current.next = current.next.next
	}

	l.size--
	return true
}
