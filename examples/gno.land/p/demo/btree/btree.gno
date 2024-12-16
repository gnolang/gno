//////////
//
// Copyright 2014 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
//
// Copyright 2024 New Tendermint
//
// This Gno port of the original Go BTree is substantially rewritten/reimplemented
// from the original, primarily for clarity of code, clarity of documentation,
// and for compatibility with Gno.
//
// Authors:
//   Original version authors -- https://github.com/google/btree/graphs/contributors
//   Kirk Haines <wyhaines@gmail.com>
//
//////////

// Package btree implements in-memory B-Trees of arbitrary degree.
//
// It has a flatter structure than an equivalent red-black or other binary tree,
// which may yield better memory usage and/or performance.
package btree

import "sort"

//////////
//
// Types
//
//////////

// BTreeOption is a function interface for setting options on a btree with `New()`.
type BTreeOption func(*BTree)

// BTree is an implementation of a B-Tree.
//
// BTree stores Record instances in an ordered structure, allowing easy insertion,
// removal, and iteration.
type BTree struct {
	degree int
	length int
	root   *node
	cowCtx *copyOnWriteContext
}

//	Any type that implements this interface can be stored in the BTree. This allows considerable
//
// flexiblity in storage within the BTree.
type Record interface {
	// Less compares self to `than`, returning true if self is less than `than`
	Less(than Record) bool
}

// records is the storage within a node. It is expressed as a slice of Record, where a Record
// is any struct that implements the Record interface.
type records []Record

// node is an internal node in a tree.
//
// It must at all times maintain on of the two conditions:
//   - len(children) == 0, len(records) unconstrained
//   - len(children) == len(records) + 1
type node struct {
	records  records
	children children
	cowCtx   *copyOnWriteContext
}

// children is the list of child nodes below the current node. It is a slice of nodes.
type children []*node

// FreeNodeList represents a slice of nodes which are available for reuse. The default
// behavior of New() is for each BTree instance to have its own FreeNodeList. However,
// it is possible for multiple instances of BTree to share the same tree. If one uses
// New(WithFreeNodeList()) to create a tree, one may pass an existing FreeNodeList, allowing
// multiple trees to use a single list. In an application with multiple trees, it might
// be more efficient to allocate a single FreeNodeList with a significant initial capacity,
// and then have all of the trees use that same large FreeNodeList.
type FreeNodeList struct {
	nodes []*node
}

// copyOnWriteContext manages node ownership and ensures that cloned trees
// maintain isolation from each other when a node is changed.
//
// Ownership Rules:
//   - Each node is associated with a specific copyOnWriteContext.
//   - A tree can modify a node directly only if the tree's context matches the node's context.
//   - If a tree attempts to modify a node with a different context, it must create a
//     new, writable copy of that node (i.e., perform a clone) before making changes.
//
// Write Operation Invariant:
//   - During any write operation, the current node being modified must have the same
//     context as the tree requesting the write.
//   - To maintain this invariant, before descending into a child node, the system checks
//     if the child’s context matches the tree's context.
//   - If the contexts match, the node can be modified in place.
//   - If the contexts do not match, a mutable copy of the child node is created with the
//     correct context before proceeding.
//
// Practical Implications:
//   - The node currently being modified inherits the requesting tree's context, allowing
//     in-place modifications.
//   - Child nodes may initially have different contexts. Before any modification, these
//     children are copied to ensure they share the correct context, enabling safe and
//     isolated updates without affecting other trees that might be referencing the original nodes.
//
// Example Usage:
// When a tree performs a write operation (e.g., inserting or deleting a node), it uses
// its copyOnWriteContext to determine whether it can modify nodes directly or needs to
// create copies. This mechanism ensures that trees can share nodes efficiently while
// maintaining data integrity.
type copyOnWriteContext struct {
	nodes *FreeNodeList
}

// Record implements an interface with a single function, Less. Any type that implements
// RecordIterator allows callers of all of the iteration functions for the BTree
// to evaluate an element of the tree as it is traversed. The function will receive
// a stored element from the tree. The function must return either a true or a false value.
// True indicates that iteration should continue, while false indicates that it should halt.
type RecordIterator func(i Record) bool

//////////
//
// Functions
//
//////////

// NewFreeNodeList creates a new free list.
// size is the maximum size of the returned free list.
func NewFreeNodeList(size int) *FreeNodeList {
	return &FreeNodeList{nodes: make([]*node, 0, size)}
}

func (freeList *FreeNodeList) newNode() (nodeInstance *node) {
	index := len(freeList.nodes) - 1
	if index < 0 {
		return new(node)
	}
	nodeInstance = freeList.nodes[index]
	freeList.nodes[index] = nil
	freeList.nodes = freeList.nodes[:index]

	return nodeInstance
}

// freeNode adds the given node to the list, returning true if it was added
// and false if it was discarded.

func (freeList *FreeNodeList) freeNode(nodeInstance *node) (nodeWasAdded bool) {
	if len(freeList.nodes) < cap(freeList.nodes) {
		freeList.nodes = append(freeList.nodes, nodeInstance)
		nodeWasAdded = true
	}
	return
}

// A default size for the free node list. We might want to run some benchmarks to see if
// there are any pros or cons to this size versus other sizes. This seems to be a reasonable
// compromise to reduce GC pressure by reusing nodes where possible, without stacking up too
// much baggage in a given tree.
const DefaultFreeNodeListSize = 32

// WithDegree sets the degree of the B-Tree.
func WithDegree(degree int) BTreeOption {
	return func(bt *BTree) {
		if degree <= 1 {
			panic("Degrees less than 1 do not make any sense for a BTree. Please provide a degree of 1 or greater.")
		}
		bt.degree = degree
	}
}

// WithFreeNodeList sets a custom free node list for the B-Tree.
func WithFreeNodeList(freeList *FreeNodeList) BTreeOption {
	return func(bt *BTree) {
		bt.cowCtx = &copyOnWriteContext{nodes: freeList}
	}
}

// New creates a new B-Tree with optional configurations. If configuration is not provided,
// it will default to 16 element nodes. Degree may not be less than 1 (which effectively
// makes the tree into a binary tree).
//
// `New(WithDegree(2))`, for example, will create a 2-3-4 tree (each node contains 1-3 records
// and 2-4 children).
//
// `New(WithFreeNodeList(NewFreeNodeList(64)))` will create a tree with a degree of 16, and
// with a free node list with a size of 64.
func New(options ...BTreeOption) *BTree {
	btree := &BTree{
		degree: 16, // default degree
		cowCtx: &copyOnWriteContext{nodes: NewFreeNodeList(DefaultFreeNodeListSize)},
	}
	for _, opt := range options {
		opt(btree)
	}
	return btree
}

// insertAt inserts a value into the given index, pushing all subsequent values
// forward.
func (recordsSlice *records) insertAt(index int, newRecord Record) {
	originalLength := len(*recordsSlice)

	// Extend the slice by one element
	*recordsSlice = append(*recordsSlice, nil)

	// Move elements from the end to avoid overwriting during the copy
	// TODO: Make this work with slice appends, instead. It should be faster?
	if index < originalLength {
		for position := originalLength; position > index; position-- {
			(*recordsSlice)[position] = (*recordsSlice)[position-1]
		}
	}

	// Insert the new record
	(*recordsSlice)[index] = newRecord
}

// removeAt removes a Record from the records slice at the specified index.
// It shifts subsequent records to fill the gap and returns the removed Record.
func (recordSlicePointer *records) removeAt(index int) Record {
	recordSlice := *recordSlicePointer
	removedRecord := recordSlice[index]
	copy(recordSlice[index:], recordSlice[index+1:])
	recordSlice[len(recordSlice)-1] = nil
	*recordSlicePointer = recordSlice[:len(recordSlice)-1]

	return removedRecord
}

// Pop removes and returns the last Record from the records slice.
// It also clears the reference to the removed Record to aid garbage collection.
func (r *records) pop() Record {
	recordSlice := *r
	lastIndex := len(recordSlice) - 1
	removedRecord := recordSlice[lastIndex]
	recordSlice[lastIndex] = nil
	*r = recordSlice[:lastIndex]
	return removedRecord
}

// This slice is intended only as a supply of records for the truncate function
// that follows, and it should not be changed or altered.
var emptyRecords = make(records, 32)

// truncate reduces the length of the slice to the specified index,
// and clears the elements beyond that index to prevent memory leaks.
// The index must be less than or equal to the current length of the slice.
func (originalSlice *records) truncate(index int) {
	// Split the slice into the part to keep and the part to clear.
	recordsToKeep := (*originalSlice)[:index]
	recordsToClear := (*originalSlice)[index:]

	// Update the original slice to only contain the records to keep.
	*originalSlice = recordsToKeep

	// Clear the memory of the part that was truncated.
	for len(recordsToClear) > 0 {
		// Copy empty values from `emptyRecords` to the recordsToClear slice.
		// This effectively "clears" the memory by overwriting elements.
		numCleared := copy(recordsToClear, emptyRecords)
		recordsToClear = recordsToClear[numCleared:]
	}
}

// Find determines the appropriate index at which a given Record should be inserted
// into the sorted records slice. If the Record already exists in the slice,
// the method returns its index and sets found to true.
//
// Parameters:
// - record: The Record to search for within the records slice.
//
// Returns:
// - insertIndex: The index at which the Record should be inserted.
// - found: A boolean indicating whether the Record already exists in the slice.
func (recordsSlice records) find(record Record) (insertIndex int, found bool) {
	totalRecords := len(recordsSlice)

	// Perform a binary search to find the insertion point for the record
	insertionPoint := sort.Search(totalRecords, func(currentIndex int) bool {
		return record.Less(recordsSlice[currentIndex])
	})

	if insertionPoint > 0 {
		previousRecord := recordsSlice[insertionPoint-1]

		if !previousRecord.Less(record) {
			return insertionPoint - 1, true
		}
	}

	return insertionPoint, false
}

// insertAt inserts a value into the given index, pushing all subsequent values
// forward.
func (childSlice *children) insertAt(index int, n *node) {
	originalLength := len(*childSlice)

	// Extend the slice by one element
	*childSlice = append(*childSlice, nil)

	// Move elements from the end to avoid overwriting during the copy
	if index < originalLength {
		for i := originalLength; i > index; i-- {
			(*childSlice)[i] = (*childSlice)[i-1]
		}
	}

	// Insert the new record
	(*childSlice)[index] = n
}

// removeAt removes a Record from the records slice at the specified index.
// It shifts subsequent records to fill the gap and returns the removed Record.
func (childSlicePointer *children) removeAt(index int) *node {
	childSlice := *childSlicePointer
	removedChild := childSlice[index]
	copy(childSlice[index:], childSlice[index+1:])
	childSlice[len(childSlice)-1] = nil
	*childSlicePointer = childSlice[:len(childSlice)-1]

	return removedChild
}

// Pop removes and returns the last Record from the records slice.
// It also clears the reference to the removed Record to aid garbage collection.
func (childSlicePointer *children) pop() *node {
	childSlice := *childSlicePointer
	lastIndex := len(childSlice) - 1
	removedChild := childSlice[lastIndex]
	childSlice[lastIndex] = nil
	*childSlicePointer = childSlice[:lastIndex]
	return removedChild
}

// This slice is intended only as a supply of records for the truncate function
// that follows, and it should not be changed or altered.
var emptyChildren = make(children, 32)

// truncate reduces the length of the slice to the specified index,
// and clears the elements beyond that index to prevent memory leaks.
// The index must be less than or equal to the current length of the slice.
func (originalSlice *children) truncate(index int) {
	// Split the slice into the part to keep and the part to clear.
	childrenToKeep := (*originalSlice)[:index]
	childrenToClear := (*originalSlice)[index:]

	// Update the original slice to only contain the records to keep.
	*originalSlice = childrenToKeep

	// Clear the memory of the part that was truncated.
	for len(childrenToClear) > 0 {
		// Copy empty values from `emptyChildren` to the recordsToClear slice.
		// This effectively "clears" the memory by overwriting elements.
		numCleared := copy(childrenToClear, emptyChildren)

		// Slice recordsToClear to exclude the elements that were just cleared.
		childrenToClear = childrenToClear[numCleared:]
	}
}

// mutableFor creates a mutable copy of the node if the current node does not
// already belong to the provided copy-on-write context (COW). If the node is
// already associated with the given COW context, it returns the current node.
//
// Parameters:
// - cowCtx: The copy-on-write context that should own the returned node.
//
// Returns:
// - A pointer to the mutable node associated with the given COW context.
//
// If the current node belongs to a different COW context, this function:
// - Allocates a new node using the provided context.
// - Copies the node’s records and children slices into the newly allocated node.
// - Returns the new node which is now owned by the given COW context.
func (n *node) mutableFor(cowCtx *copyOnWriteContext) *node {
	// If the current node is already owned by the provided context, return it as-is.
	if n.cowCtx == cowCtx {
		return n
	}

	// Create a new node in the provided context.
	newNode := cowCtx.newNode()

	// Copy the records from the current node into the new node.
	newNode.records = append(newNode.records[:0], n.records...)

	// Copy the children from the current node into the new node.
	newNode.children = append(newNode.children[:0], n.children...)

	return newNode
}

// mutableChild ensures that the child node at the given index is mutable and
// associated with the same COW context as the parent node. If the child node
// belongs to a different context, a copy of the child is created and stored in the
// parent node.
//
// Parameters:
// - i: The index of the child node to be made mutable.
//
// Returns:
// - A pointer to the mutable child node.
func (n *node) mutableChild(i int) *node {
	// Ensure that the child at index `i` is mutable and belongs to the same context as the parent.
	mutableChildNode := n.children[i].mutableFor(n.cowCtx)
	// Update the child node reference in the current node to the mutable version.
	n.children[i] = mutableChildNode
	return mutableChildNode
}

// split splits the given node at the given index.  The current node shrinks,
// and this function returns the record that existed at that index and a new node
// containing all records/children after it.
func (n *node) split(i int) (Record, *node) {
	record := n.records[i]
	next := n.cowCtx.newNode()
	next.records = append(next.records, n.records[i+1:]...)
	n.records.truncate(i)
	if len(n.children) > 0 {
		next.children = append(next.children, n.children[i+1:]...)
		n.children.truncate(i + 1)
	}
	return record, next
}

// maybeSplitChild checks if a child should be split, and if so splits it.
// Returns whether or not a split occurred.
func (n *node) maybeSplitChild(i, maxRecords int) bool {
	if len(n.children[i].records) < maxRecords {
		return false
	}
	first := n.mutableChild(i)
	record, second := first.split(maxRecords / 2)
	n.records.insertAt(i, record)
	n.children.insertAt(i+1, second)
	return true
}

// insert adds a record to the subtree rooted at the current node, ensuring that no node in the subtree
// exceeds the maximum number of allowed records (`maxRecords`). If an equivalent record is already present,
// it replaces the existing one and returns it; otherwise, it returns nil.
//
// Parameters:
// - record: The record to be inserted.
// - maxRecords: The maximum number of records allowed per node.
//
// Returns:
// - The record that was replaced if an equivalent record already existed, otherwise nil.
func (n *node) insert(record Record, maxRecords int) Record {
	// Find the position where the new record should be inserted and check if an equivalent record already exists.
	insertionIndex, recordExists := n.records.find(record)

	if recordExists {
		// If an equivalent record is found, replace it and return the old record.
		existingRecord := n.records[insertionIndex]
		n.records[insertionIndex] = record
		return existingRecord
	}

	// If the current node is a leaf (has no children), insert the new record at the calculated index.
	if len(n.children) == 0 {
		n.records.insertAt(insertionIndex, record)
		return nil
	}

	// Check if the child node at the insertion index needs to be split due to exceeding maxRecords.
	if n.maybeSplitChild(insertionIndex, maxRecords) {
		// If a split occurred, compare the new record with the record moved up to the current node.
		splitRecord := n.records[insertionIndex]
		switch {
		case record.Less(splitRecord):
			// The new record belongs to the first (left) split node; no change to insertion index.
		case splitRecord.Less(record):
			// The new record belongs to the second (right) split node; move the insertion index to the next position.
			insertionIndex++
		default:
			// If the record is equivalent to the split record, replace it and return the old record.
			existingRecord := n.records[insertionIndex]
			n.records[insertionIndex] = record
			return existingRecord
		}
	}

	// Recursively insert the record into the appropriate child node, now guaranteed to have space.
	return n.mutableChild(insertionIndex).insert(record, maxRecords)
}

// get finds the given key in the subtree and returns it.
func (n *node) get(key Record) Record {
	i, found := n.records.find(key)
	if found {
		return n.records[i]
	} else if len(n.children) > 0 {
		return n.children[i].get(key)
	}
	return nil
}

// min returns the first record in the subtree.
func min(n *node) Record {
	if n == nil {
		return nil
	}
	for len(n.children) > 0 {
		n = n.children[0]
	}
	if len(n.records) == 0 {
		return nil
	}
	return n.records[0]
}

// max returns the last record in the subtree.
func max(n *node) Record {
	if n == nil {
		return nil
	}
	for len(n.children) > 0 {
		n = n.children[len(n.children)-1]
	}
	if len(n.records) == 0 {
		return nil
	}
	return n.records[len(n.records)-1]
}

// toRemove details what record to remove in a node.remove call.
type toRemove int

const (
	removeRecord toRemove = iota // removes the given record
	removeMin                    // removes smallest record in the subtree
	removeMax                    // removes largest record in the subtree
)

// remove removes a record from the subtree rooted at the current node.
//
// Parameters:
// - record: The record to be removed (can be nil when the removal type indicates min or max).
// - minRecords: The minimum number of records a node should have after removal.
// - typ: The type of removal operation to perform (removeMin, removeMax, or removeRecord).
//
// Returns:
// - The record that was removed, or nil if no such record was found.
func (n *node) remove(record Record, minRecords int, removalType toRemove) Record {
	var targetIndex int
	var recordFound bool

	// Determine the index of the record to remove based on the removal type.
	switch removalType {
	case removeMax:
		// If this node is a leaf, remove and return the last record.
		if len(n.children) == 0 {
			return n.records.pop()
		}
		targetIndex = len(n.records) // The last record index for removing max.

	case removeMin:
		// If this node is a leaf, remove and return the first record.
		if len(n.children) == 0 {
			return n.records.removeAt(0)
		}
		targetIndex = 0 // The first record index for removing min.

	case removeRecord:
		// Locate the index of the record to be removed.
		targetIndex, recordFound = n.records.find(record)
		if len(n.children) == 0 {
			if recordFound {
				return n.records.removeAt(targetIndex)
			}
			return nil // The record was not found in the leaf node.
		}

	default:
		panic("invalid removal type")
	}

	// If the current node has children, handle the removal recursively.
	if len(n.children[targetIndex].records) <= minRecords {
		// If the target child node has too few records, grow it before proceeding with removal.
		return n.growChildAndRemove(targetIndex, record, minRecords, removalType)
	}

	// Get a mutable reference to the child node at the target index.
	targetChild := n.mutableChild(targetIndex)

	// If the record to be removed was found in the current node:
	if recordFound {
		// Replace the current record with its predecessor from the child node, and return the removed record.
		replacedRecord := n.records[targetIndex]
		n.records[targetIndex] = targetChild.remove(nil, minRecords, removeMax)
		return replacedRecord
	}

	// Recursively remove the record from the child node.
	return targetChild.remove(record, minRecords, removalType)
}

// growChildAndRemove grows child 'i' to make sure it's possible to remove an
// record from it while keeping it at minRecords, then calls remove to actually
// remove it.
//
// Most documentation says we have to do two sets of special casing:
//  1. record is in this node
//  2. record is in child
//
// In both cases, we need to handle the two subcases:
//
//	A) node has enough values that it can spare one
//	B) node doesn't have enough values
//
// For the latter, we have to check:
//
//	a) left sibling has node to spare
//	b) right sibling has node to spare
//	c) we must merge
//
// To simplify our code here, we handle cases #1 and #2 the same:
// If a node doesn't have enough records, we make sure it does (using a,b,c).
// We then simply redo our remove call, and the second time (regardless of
// whether we're in case 1 or 2), we'll have enough records and can guarantee
// that we hit case A.
func (n *node) growChildAndRemove(i int, record Record, minRecords int, typ toRemove) Record {
	if i > 0 && len(n.children[i-1].records) > minRecords {
		// Steal from left child
		child := n.mutableChild(i)
		stealFrom := n.mutableChild(i - 1)
		stolenRecord := stealFrom.records.pop()
		child.records.insertAt(0, n.records[i-1])
		n.records[i-1] = stolenRecord
		if len(stealFrom.children) > 0 {
			child.children.insertAt(0, stealFrom.children.pop())
		}
	} else if i < len(n.records) && len(n.children[i+1].records) > minRecords {
		// steal from right child
		child := n.mutableChild(i)
		stealFrom := n.mutableChild(i + 1)
		stolenRecord := stealFrom.records.removeAt(0)
		child.records = append(child.records, n.records[i])
		n.records[i] = stolenRecord
		if len(stealFrom.children) > 0 {
			child.children = append(child.children, stealFrom.children.removeAt(0))
		}
	} else {
		if i >= len(n.records) {
			i--
		}
		child := n.mutableChild(i)
		// merge with right child
		mergeRecord := n.records.removeAt(i)
		mergeChild := n.children.removeAt(i + 1).mutableFor(n.cowCtx)
		child.records = append(child.records, mergeRecord)
		child.records = append(child.records, mergeChild.records...)
		child.children = append(child.children, mergeChild.children...)
		n.cowCtx.freeNode(mergeChild)
	}
	return n.remove(record, minRecords, typ)
}

type direction int

const (
	descend = direction(-1)
	ascend  = direction(+1)
)

// iterate provides a simple method for iterating over elements in the tree.
//
// When ascending, the 'start' should be less than 'stop' and when descending,
// the 'start' should be greater than 'stop'. Setting 'includeStart' to true
// will force the iterator to include the first record when it equals 'start',
// thus creating a "greaterOrEqual" or "lessThanEqual" rather than just a
// "greaterThan" or "lessThan" queries.
func (n *node) iterate(dir direction, start, stop Record, includeStart bool, hit bool, iter RecordIterator) (bool, bool) {
	var ok, found bool
	var index int
	switch dir {
	case ascend:
		if start != nil {
			index, _ = n.records.find(start)
		}
		for i := index; i < len(n.records); i++ {
			if len(n.children) > 0 {
				if hit, ok = n.children[i].iterate(dir, start, stop, includeStart, hit, iter); !ok {
					return hit, false
				}
			}
			if !includeStart && !hit && start != nil && !start.Less(n.records[i]) {
				hit = true
				continue
			}
			hit = true
			if stop != nil && !n.records[i].Less(stop) {
				return hit, false
			}
			if !iter(n.records[i]) {
				return hit, false
			}
		}
		if len(n.children) > 0 {
			if hit, ok = n.children[len(n.children)-1].iterate(dir, start, stop, includeStart, hit, iter); !ok {
				return hit, false
			}
		}
	case descend:
		if start != nil {
			index, found = n.records.find(start)
			if !found {
				index = index - 1
			}
		} else {
			index = len(n.records) - 1
		}
		for i := index; i >= 0; i-- {
			if start != nil && !n.records[i].Less(start) {
				if !includeStart || hit || start.Less(n.records[i]) {
					continue
				}
			}
			if len(n.children) > 0 {
				if hit, ok = n.children[i+1].iterate(dir, start, stop, includeStart, hit, iter); !ok {
					return hit, false
				}
			}
			if stop != nil && !stop.Less(n.records[i]) {
				return hit, false //	continue
			}
			hit = true
			if !iter(n.records[i]) {
				return hit, false
			}
		}
		if len(n.children) > 0 {
			if hit, ok = n.children[0].iterate(dir, start, stop, includeStart, hit, iter); !ok {
				return hit, false
			}
		}
	}
	return hit, true
}

func (tree *BTree) Iterate(dir direction, start, stop Record, includeStart bool, hit bool, iter RecordIterator) (bool, bool) {
	return tree.root.iterate(dir, start, stop, includeStart, hit, iter)
}

// Clone creates a new BTree instance that shares the current tree's structure using a copy-on-write (COW) approach.
//
// How Cloning Works:
//   - The cloned tree (`clonedTree`) shares the current tree’s nodes in a read-only state. This means that no additional memory
//     is allocated for shared nodes, and read operations on the cloned tree are as fast as on the original tree.
//   - When either the original tree (`t`) or the cloned tree (`clonedTree`) needs to perform a write operation (such as an insert, delete, etc.),
//     a new copy of the affected nodes is created on-demand. This ensures that modifications to one tree do not affect the other.
//
// Performance Implications:
//   - **Clone Creation:** The creation of a clone is inexpensive since it only involves copying references to the original tree's nodes
//     and creating new copy-on-write contexts.
//   - **Read Operations:** Reading from either the original tree or the cloned tree has no additional performance overhead compared to the original tree.
//   - **Write Operations:** The first write operation on either tree may experience a slight slow-down due to the allocation of new nodes,
//     but subsequent write operations will perform at the same speed as if the tree were not cloned.
//
// Returns:
// - A new BTree instance (`clonedTree`) that shares the original tree's structure.
func (t *BTree) Clone() *BTree {
	// Create two independent copy-on-write contexts, one for the original tree (`t`) and one for the cloned tree.
	originalContext := *t.cowCtx
	clonedContext := *t.cowCtx

	// Create a shallow copy of the current tree, which will be the new cloned tree.
	clonedTree := *t

	// Assign the new contexts to their respective trees.
	t.cowCtx = &originalContext
	clonedTree.cowCtx = &clonedContext

	return &clonedTree
}

// maxRecords returns the max number of records to allow per node.
func (t *BTree) maxRecords() int {
	return t.degree*2 - 1
}

// minRecords returns the min number of records to allow per node (ignored for the
// root node).
func (t *BTree) minRecords() int {
	return t.degree - 1
}

func (c *copyOnWriteContext) newNode() (n *node) {
	n = c.nodes.newNode()
	n.cowCtx = c
	return
}

type freeType int

const (
	ftFreelistFull freeType = iota // node was freed (available for GC, not stored in nodes)
	ftStored                       // node was stored in the nodes for later use
	ftNotOwned                     // node was ignored by COW, since it's owned by another one
)

// freeNode frees a node within a given COW context, if it's owned by that
// context.  It returns what happened to the node (see freeType const
// documentation).
func (c *copyOnWriteContext) freeNode(n *node) freeType {
	if n.cowCtx == c {
		// clear to allow GC
		n.records.truncate(0)
		n.children.truncate(0)
		n.cowCtx = nil
		if c.nodes.freeNode(n) {
			return ftStored
		} else {
			return ftFreelistFull
		}
	} else {
		return ftNotOwned
	}
}

// Insert adds the given record to the B-tree. If a record already exists in the tree with the same value,
// it is replaced, and the old record is returned. Otherwise, it returns nil.
//
// Notes:
// - The function panics if a nil record is provided as input.
// - If the root node is empty, a new root node is created and the record is inserted.
//
// Parameters:
// - record: The record to be inserted into the B-tree.
//
// Returns:
// - The replaced record if an equivalent record already exists, or nil if no replacement occurred.
func (t *BTree) Insert(record Record) Record {
	if record == nil {
		panic("nil record cannot be added to BTree")
	}

	// If the tree is empty (no root), create a new root node and insert the record.
	if t.root == nil {
		t.root = t.cowCtx.newNode()
		t.root.records = append(t.root.records, record)
		t.length++
		return nil
	}

	// Ensure that the root node is mutable (associated with the current tree's copy-on-write context).
	t.root = t.root.mutableFor(t.cowCtx)

	// If the root node is full (contains the maximum number of records), split the root.
	if len(t.root.records) >= t.maxRecords() {
		// Split the root node, promoting the middle record and creating a new child node.
		middleRecord, newChildNode := t.root.split(t.maxRecords() / 2)

		// Create a new root node to hold the promoted middle record.
		oldRoot := t.root
		t.root = t.cowCtx.newNode()
		t.root.records = append(t.root.records, middleRecord)
		t.root.children = append(t.root.children, oldRoot, newChildNode)
	}

	// Insert the new record into the subtree rooted at the current root node.
	replacedRecord := t.root.insert(record, t.maxRecords())

	// If no record was replaced, increase the tree's length.
	if replacedRecord == nil {
		t.length++
	}

	return replacedRecord
}

// Delete removes an record equal to the passed in record from the tree, returning
// it.  If no such record exists, returns nil.
func (t *BTree) Delete(record Record) Record {
	return t.deleteRecord(record, removeRecord)
}

// DeleteMin removes the smallest record in the tree and returns it.
// If no such record exists, returns nil.
func (t *BTree) DeleteMin() Record {
	return t.deleteRecord(nil, removeMin)
}

// Shift is identical to DeleteMin. If the tree is thought of as an ordered list, then Shift()
// removes the element at the start of the list, the smallest element, and returns it.
func (t *BTree) Shift() Record {
	return t.deleteRecord(nil, removeMin)
}

// DeleteMax removes the largest record in the tree and returns it.
// If no such record exists, returns nil.
func (t *BTree) DeleteMax() Record {
	return t.deleteRecord(nil, removeMax)
}

// Pop is identical to DeleteMax. If the tree is thought of as an ordered list, then Shift()
// removes the element at the end of the list, the largest element, and returns it.
func (t *BTree) Pop() Record {
	return t.deleteRecord(nil, removeMax)
}

// deleteRecord removes a record from the B-tree based on the specified removal type (removeMin, removeMax, or removeRecord).
// It returns the removed record if it was found, or nil if no matching record was found.
//
// Parameters:
// - record: The record to be removed (can be nil if the removal type indicates min or max).
// - removalType: The type of removal operation to perform (removeMin, removeMax, or removeRecord).
//
// Returns:
// - The removed record if it existed in the tree, or nil if it was not found.
func (t *BTree) deleteRecord(record Record, removalType toRemove) Record {
	// If the tree is empty or the root has no records, return nil.
	if t.root == nil || len(t.root.records) == 0 {
		return nil
	}

	// Ensure the root node is mutable (associated with the tree's copy-on-write context).
	t.root = t.root.mutableFor(t.cowCtx)

	// Attempt to remove the specified record from the root node.
	removedRecord := t.root.remove(record, t.minRecords(), removalType)

	// Check if the root node has become empty but still has children.
	// In this case, the tree height should be reduced, making the first child the new root.
	if len(t.root.records) == 0 && len(t.root.children) > 0 {
		oldRoot := t.root
		t.root = t.root.children[0]
		// Free the old root node, as it is no longer needed.
		t.cowCtx.freeNode(oldRoot)
	}

	// If a record was successfully removed, decrease the tree's length.
	if removedRecord != nil {
		t.length--
	}

	return removedRecord
}

// AscendRange calls the iterator for every value in the tree within the range
// [greaterOrEqual, lessThan), until iterator returns false.
func (t *BTree) AscendRange(greaterOrEqual, lessThan Record, iterator RecordIterator) {
	if t.root == nil {
		return
	}
	t.root.iterate(ascend, greaterOrEqual, lessThan, true, false, iterator)
}

// AscendLessThan calls the iterator for every value in the tree within the range
// [first, pivot), until iterator returns false.
func (t *BTree) AscendLessThan(pivot Record, iterator RecordIterator) {
	if t.root == nil {
		return
	}
	t.root.iterate(ascend, nil, pivot, false, false, iterator)
}

// AscendGreaterOrEqual calls the iterator for every value in the tree within
// the range [pivot, last], until iterator returns false.
func (t *BTree) AscendGreaterOrEqual(pivot Record, iterator RecordIterator) {
	if t.root == nil {
		return
	}
	t.root.iterate(ascend, pivot, nil, true, false, iterator)
}

// Ascend calls the iterator for every value in the tree within the range
// [first, last], until iterator returns false.
func (t *BTree) Ascend(iterator RecordIterator) {
	if t.root == nil {
		return
	}
	t.root.iterate(ascend, nil, nil, false, false, iterator)
}

// DescendRange calls the iterator for every value in the tree within the range
// [lessOrEqual, greaterThan), until iterator returns false.
func (t *BTree) DescendRange(lessOrEqual, greaterThan Record, iterator RecordIterator) {
	if t.root == nil {
		return
	}
	t.root.iterate(descend, lessOrEqual, greaterThan, true, false, iterator)
}

// DescendLessOrEqual calls the iterator for every value in the tree within the range
// [pivot, first], until iterator returns false.
func (t *BTree) DescendLessOrEqual(pivot Record, iterator RecordIterator) {
	if t.root == nil {
		return
	}
	t.root.iterate(descend, pivot, nil, true, false, iterator)
}

// DescendGreaterThan calls the iterator for every value in the tree within
// the range [last, pivot), until iterator returns false.
func (t *BTree) DescendGreaterThan(pivot Record, iterator RecordIterator) {
	if t.root == nil {
		return
	}
	t.root.iterate(descend, nil, pivot, false, false, iterator)
}

// Descend calls the iterator for every value in the tree within the range
// [last, first], until iterator returns false.
func (t *BTree) Descend(iterator RecordIterator) {
	if t.root == nil {
		return
	}
	t.root.iterate(descend, nil, nil, false, false, iterator)
}

// Get looks for the key record in the tree, returning it.  It returns nil if
// unable to find that record.
func (t *BTree) Get(key Record) Record {
	if t.root == nil {
		return nil
	}
	return t.root.get(key)
}

// Min returns the smallest record in the tree, or nil if the tree is empty.
func (t *BTree) Min() Record {
	return min(t.root)
}

// Max returns the largest record in the tree, or nil if the tree is empty.
func (t *BTree) Max() Record {
	return max(t.root)
}

// Has returns true if the given key is in the tree.
func (t *BTree) Has(key Record) bool {
	return t.Get(key) != nil
}

// Len returns the number of records currently in the tree.
func (t *BTree) Len() int {
	return t.length
}

// Clear removes all elements from the B-tree.
//
// Parameters:
// - addNodesToFreelist:
//     - If true, the tree's nodes are added to the freelist during the clearing process,
//       up to the freelist's capacity.
//     - If false, the root node is simply dereferenced, allowing Go's garbage collector
//       to reclaim the memory.
//
// Benefits:
// - **Performance:**
//     - Significantly faster than deleting each element individually, as it avoids the overhead
//       of searching and updating the tree structure for each deletion.
//     - More efficient than creating a new tree, since it reuses existing nodes by adding them
//       to the freelist instead of discarding them to the garbage collector.
//
// Time Complexity:
// - **O(1):**
//     - When `addNodesToFreelist` is false.
//     - When `addNodesToFreelist` is true but the freelist is already full.
// - **O(freelist size):**
//     - When adding nodes to the freelist up to its capacity.
// - **O(tree size):**
//     - When iterating through all nodes to add to the freelist, but none can be added due to
//       ownership by another tree.

func (tree *BTree) Clear(addNodesToFreelist bool) {
	if tree.root != nil && addNodesToFreelist {
		tree.root.reset(tree.cowCtx)
	}
	tree.root = nil
	tree.length = 0
}

// reset adds all nodes in the current subtree to the freelist.
//
// The function operates recursively:
// - It first attempts to reset all child nodes.
// - If the freelist becomes full at any point, the process stops immediately.
//
// Parameters:
// - copyOnWriteCtx: The copy-on-write context managing the freelist.
//
// Returns:
// - true: Indicates that the parent node should continue attempting to reset its nodes.
// - false: Indicates that the freelist is full and no further nodes should be added.
//
// Usage:
// This method is called during the `Clear` operation of the B-tree to efficiently reuse
// nodes by adding them to the freelist, thereby avoiding unnecessary allocations and reducing
// garbage collection overhead.
func (currentNode *node) reset(copyOnWriteCtx *copyOnWriteContext) bool {
	// Iterate through each child node and attempt to reset it.
	for _, childNode := range currentNode.children {
		// If any child reset operation signals that the freelist is full, stop the process.
		if !childNode.reset(copyOnWriteCtx) {
			return false
		}
	}

	// Attempt to add the current node to the freelist.
	// If the freelist is full after this operation, indicate to the parent to stop.
	freelistStatus := copyOnWriteCtx.freeNode(currentNode)
	return freelistStatus != ftFreelistFull
}
