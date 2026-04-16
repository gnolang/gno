package bptree

import (
	"bytes"
	"fmt"
)

// Iterator traverses key-value pairs in a B+ tree within a [start, end) range.
// Implements the same contract as tm2/pkg/db.Iterator.
// Uses a stack of (innerNode, childIndex) pairs for traversal — no leaf
// sibling pointers needed.
type Iterator struct {
	// Configuration
	start         []byte
	end           []byte
	ascending     bool
	ndb           *nodeDB       // for value resolution via DB; nil for in-memory
	valueResolver ValueResolver // alternative value resolution (from ImmutableTree)

	// State
	stack     []stackEntry
	leaf      *LeafNode
	leafIdx   int // current position within leaf
	valid     bool
	err       error
	closed    bool

	// Version reader tracking
	version int64 // 0 if no version reader
}

type stackEntry struct {
	inner    *InnerNode
	childIdx int
}

// newIterator creates an iterator over the tree rooted at root.
// start is inclusive, end is exclusive. If start is nil, starts from the beginning.
// If end is nil, iterates to the end.
func newIterator(root Node, start, end []byte, ascending bool, ndb *nodeDB, version int64) *Iterator {
	it := &Iterator{
		start:     start,
		end:       end,
		ascending: ascending,
		ndb:       ndb,
		version:   version,
	}
	// Register as an active reader of this version so that a concurrent
	// PruneVersionsTo(version) returns ErrActiveReaders until the iterator
	// is closed. Only meaningful for DB-backed snapshots of saved versions.
	// See Finding #1.
	if ndb != nil && version > 0 {
		ndb.incrVersionReaders(version)
	}
	if root == nil {
		it.valid = false
		return it
	}

	if ascending {
		it.seekFirst(root)
	} else {
		it.seekLast(root)
	}
	return it
}

// seekFirst positions the iterator at the first key >= start.
func (it *Iterator) seekFirst(node Node) {
	for {
		switch n := node.(type) {
		case *InnerNode:
			var childIdx int
			if it.start != nil {
				childIdx = searchInner(n, it.start)
			} else {
				childIdx = 0
			}
			it.stack = append(it.stack, stackEntry{inner: n, childIdx: childIdx})
			node = n.getChild(childIdx)
		case *LeafNode:
			it.leaf = n
			if it.start != nil {
				pos, _ := searchLeaf(n, it.start)
				it.leafIdx = pos
			} else {
				it.leafIdx = 0
			}
			// Advance to first valid position
			if it.leafIdx >= int(n.numKeys) {
				it.nextLeaf()
			} else {
				it.valid = true
				it.checkEnd()
			}
			return
		default:
			it.valid = false
			return
		}
	}
}

// seekLast positions the iterator at the last key < end (or the very last key if end is nil).
func (it *Iterator) seekLast(node Node) {
	for {
		switch n := node.(type) {
		case *InnerNode:
			var childIdx int
			if it.end != nil {
				childIdx = searchInner(n, it.end)
				// searchInner returns the child for keys >= end.
				// We want the child containing keys < end.
				// If end would be in child[childIdx], that child may have keys < end.
				// But we need the rightmost key < end, so start from childIdx
				// and let the leaf positioning handle the exact boundary.
				if childIdx >= n.NumChildren() {
					childIdx = n.NumChildren() - 1
				}
			} else {
				childIdx = n.NumChildren() - 1
			}
			it.stack = append(it.stack, stackEntry{inner: n, childIdx: childIdx})
			node = n.getChild(childIdx)
		case *LeafNode:
			it.leaf = n
			if it.end != nil {
				// end is exclusive for both branches: if end matches a
				// key, pos-1 is the last key < end; if end would be
				// inserted at pos, pos-1 is still the last key < end.
				pos, _ := searchLeaf(n, it.end)
				it.leafIdx = pos - 1
			} else {
				it.leafIdx = int(n.numKeys) - 1
			}
			if it.leafIdx < 0 {
				it.prevLeaf()
			} else {
				it.valid = true
				it.checkStart()
			}
			return
		default:
			it.valid = false
			return
		}
	}
}

// nextLeaf advances to the next leaf in ascending order using the stack.
func (it *Iterator) nextLeaf() {
	for len(it.stack) > 0 {
		top := &it.stack[len(it.stack)-1]
		top.childIdx++
		if top.childIdx < top.inner.NumChildren() {
			// Descend to leftmost leaf of next child
			node := top.inner.getChild(top.childIdx)
			it.descendLeft(node)
			return
		}
		// Pop exhausted inner node
		it.stack = it.stack[:len(it.stack)-1]
	}
	// Stack empty — no more leaves
	it.valid = false
}

// prevLeaf moves to the previous leaf in descending order using the stack.
func (it *Iterator) prevLeaf() {
	for len(it.stack) > 0 {
		top := &it.stack[len(it.stack)-1]
		top.childIdx--
		if top.childIdx >= 0 {
			// Descend to rightmost leaf of previous child
			node := top.inner.getChild(top.childIdx)
			it.descendRight(node)
			return
		}
		// Pop exhausted inner node
		it.stack = it.stack[:len(it.stack)-1]
	}
	it.valid = false
}

// descendLeft descends from node to the leftmost leaf, pushing inner nodes onto the stack.
func (it *Iterator) descendLeft(node Node) {
	for {
		switch n := node.(type) {
		case *InnerNode:
			it.stack = append(it.stack, stackEntry{inner: n, childIdx: 0})
			node = n.getChild(0)
		case *LeafNode:
			it.leaf = n
			it.leafIdx = 0
			it.valid = true
			it.checkEnd()
			return
		default:
			it.valid = false
			return
		}
	}
}

// descendRight descends from node to the rightmost leaf.
func (it *Iterator) descendRight(node Node) {
	for {
		switch n := node.(type) {
		case *InnerNode:
			idx := n.NumChildren() - 1
			it.stack = append(it.stack, stackEntry{inner: n, childIdx: idx})
			node = n.getChild(idx)
		case *LeafNode:
			it.leaf = n
			it.leafIdx = int(n.numKeys) - 1
			it.valid = true
			it.checkStart()
			return
		default:
			it.valid = false
			return
		}
	}
}

// checkEnd invalidates the iterator if the current key is >= end.
func (it *Iterator) checkEnd() {
	if !it.valid || it.end == nil {
		return
	}
	if bytes.Compare(it.leaf.keys[it.leafIdx], it.end) >= 0 {
		it.valid = false
	}
}

// checkStart invalidates the iterator if the current key is < start.
func (it *Iterator) checkStart() {
	if !it.valid || it.start == nil {
		return
	}
	if bytes.Compare(it.leaf.keys[it.leafIdx], it.start) < 0 {
		it.valid = false
	}
}

// --- db.Iterator interface ---

func (it *Iterator) Domain() (start, end []byte) {
	return it.start, it.end
}

func (it *Iterator) Valid() bool {
	return it.valid && !it.closed
}

func (it *Iterator) Next() {
	if !it.valid {
		return
	}
	if it.ascending {
		it.leafIdx++
		if it.leafIdx >= int(it.leaf.numKeys) {
			it.nextLeaf()
		} else {
			it.checkEnd()
		}
	} else {
		it.leafIdx--
		if it.leafIdx < 0 {
			it.prevLeaf()
		} else {
			it.checkStart()
		}
	}
}

func (it *Iterator) Key() []byte {
	if !it.valid {
		panic("iterator invalid")
	}
	return it.leaf.keys[it.leafIdx]
}

func (it *Iterator) Value() []byte {
	if !it.valid {
		panic("iterator invalid")
	}
	vk := it.leaf.valueKeys[it.leafIdx]
	if it.ndb != nil {
		val, err := it.ndb.GetValue(vk)
		if err != nil {
			it.err = err
			it.valid = false
			return nil
		}
		return val
	}
	if it.valueResolver != nil {
		val, err := it.valueResolver(vk)
		if err != nil {
			it.err = err
			it.valid = false
			return nil
		}
		return val
	}
	return nil // no resolver available
}

func (it *Iterator) Error() error {
	return it.err
}

func (it *Iterator) Close() error {
	if it.closed {
		return nil
	}
	it.closed = true
	it.valid = false
	if it.ndb != nil && it.version > 0 {
		it.ndb.decrVersionReaders(it.version)
	}
	return nil
}

// --- Tree integration ---

// NewIteratorWithNDB creates an iterator over an immutable tree with value
// resolution via the mutable tree's nodeDB. Used by the store wrapper.
//
// The iterator registers as a version reader on imm.version (when DB-backed)
// so that a concurrent PruneVersionsTo(imm.version) is rejected until the
// iterator is closed. Callers MUST call Close() on the returned iterator.
func NewIteratorWithNDB(imm *ImmutableTree, start, end []byte, ascending bool, mtree *MutableTree) *Iterator {
	var ndb *nodeDB
	if mtree != nil {
		ndb = mtree.ndb
	}
	itr := newIterator(imm.root, start, end, ascending, ndb, imm.version)
	if ndb == nil && mtree != nil && mtree.memValues != nil {
		itr.valueResolver = func(vk []byte) ([]byte, error) {
			val, ok := mtree.memValues[string(vk)]
			if !ok {
				return nil, fmt.Errorf("value not found in memValues for key %x", vk)
			}
			return val, nil
		}
	}
	return itr
}

// Iterator returns an iterator over [start, end) in the given direction.
func (t *MutableTree) Iterator(start, end []byte, ascending bool) (*Iterator, error) {
	itr := newIterator(t.root, start, end, ascending, t.ndb, 0)
	if t.ndb == nil && t.memValues != nil {
		itr.valueResolver = func(vk []byte) ([]byte, error) {
			val, ok := t.memValues[string(vk)]
			if !ok {
				return nil, fmt.Errorf("value not found in memValues for key %x", vk)
			}
			return val, nil
		}
	}
	return itr, nil
}

// ImmutableTree.Iterator returns an iterator. If a valueResolver is set,
// the iterator resolves values via a wrapping ndb-like mechanism.
// For DB-backed trees, use NewIteratorWithNDB instead.
func (t *ImmutableTree) Iterator(start, end []byte, ascending bool) (*Iterator, error) {
	itr := newIterator(t.root, start, end, ascending, nil, 0)
	itr.valueResolver = t.valueResolver
	return itr, nil
}

// IterateRange iterates over [start, end) calling fn. Stops if fn returns true.
func (t *ImmutableTree) IterateRange(start, end []byte, ascending bool, fn func(key, value []byte) bool) bool {
	itr, _ := t.Iterator(start, end, ascending)
	defer itr.Close()
	for itr.Valid() {
		if fn(itr.Key(), itr.Value()) {
			return true
		}
		itr.Next()
	}
	return false
}

// IterateRange on MutableTree.
func (t *MutableTree) IterateRange(start, end []byte, ascending bool, fn func(key, value []byte) bool) bool {
	itr, _ := t.Iterator(start, end, ascending)
	defer itr.Close()
	for itr.Valid() {
		if fn(itr.Key(), itr.Value()) {
			return true
		}
		itr.Next()
	}
	return false
}
