package bptree

import "bytes"

// Iterator traverses key-value pairs in a B+ tree within a [start, end) range.
// Implements the same contract as tm2/pkg/db.Iterator.
// Uses a stack of (innerNode, childIndex) pairs for traversal — no leaf
// sibling pointers needed.
type Iterator struct {
	// Configuration
	start         []byte
	end           []byte
	ascending     bool
	ndb           *nodeDB       // for value resolution via DB; nil for snapshot-only iterators
	valueResolver ValueResolver // alternative value resolution (from ImmutableTree)

	// State
	stack   []stackEntry
	leaf    *LeafNode
	leafIdx int // current position within leaf
	valid   bool
	err     error
	closed  bool

	// Version reader tracking
	version int64 // 0 if no version reader
}

type stackEntry struct {
	inner    *InnerNode
	childIdx int
}

// copyBound returns a nil-preserving copy of an iterator bound: nil stays nil
// (nil means "unbounded" — an empty slice would invert that semantics).
func copyBound(b []byte) []byte {
	if b == nil {
		return nil
	}
	c := make([]byte, len(b))
	copy(c, b)
	return c
}

// newIterator creates an iterator over the tree rooted at root.
// start is inclusive, end is exclusive. If start is nil, starts from the beginning.
// If end is nil, iterates to the end.
//
// The bounds are copied at construction: checkStart/checkEnd consult them on
// every Next, so a caller mutating its slices mid-iteration must not shift
// the range.
func newIterator(root Node, start, end []byte, ascending bool, ndb *nodeDB, version int64) *Iterator {
	it := &Iterator{
		start:     copyBound(start),
		end:       copyBound(end),
		ascending: ascending,
		ndb:       ndb,
		version:   version,
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
	// A failed construction seek holds no reservation: release the version
	// reader now (the caller may never Close a never-valid iterator) and zero
	// the version so a deferred Close — which decrements only when version>0
	// — cannot double-release.
	if it.err != nil && it.ndb != nil && it.version > 0 {
		it.ndb.decrVersionReaders(it.version)
		it.version = 0
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
			child, err := n.getChild(childIdx)
			if err != nil {
				it.err = err
				it.valid = false
				return
			}
			node = child
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
			child, err := n.getChild(childIdx)
			if err != nil {
				it.err = err
				it.valid = false
				return
			}
			node = child
		case *LeafNode:
			it.leaf = n
			if it.end != nil {
				pos, found := searchLeaf(n, it.end)
				if found {
					it.leafIdx = pos - 1 // end is exclusive
				} else {
					it.leafIdx = pos - 1 // pos is where end would be inserted
				}
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
			node, err := top.inner.getChild(top.childIdx)
			if err != nil {
				it.err = err
				it.valid = false
				return
			}
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
			node, err := top.inner.getChild(top.childIdx)
			if err != nil {
				it.err = err
				it.valid = false
				return
			}
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
			child, err := n.getChild(0)
			if err != nil {
				it.err = err
				it.valid = false
				return
			}
			node = child
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
			child, err := n.getChild(idx)
			if err != nil {
				it.err = err
				it.valid = false
				return
			}
			node = child
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
	// Copy per the dbm iterator contract ("safe for modification"): the raw
	// slice belongs to a live leaf shared with the tree and the node cache —
	// mutating it would corrupt committed state.
	return copyKey(it.leaf.keys[it.leafIdx])
}

func (it *Iterator) Value() []byte {
	if !it.valid {
		panic("iterator invalid")
	}
	vk := it.leaf.valueKeys[it.leafIdx]
	// Resolve via the per-source resolver. A working-tree iterator carries a
	// pendingVals-aware resolver (read-your-writes, single-writer); a
	// committed-snapshot iterator carries a DB-only resolver, so it never
	// touches the writer's pendingVals map and cannot race SaveValue.
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
// When the immutable tree carries a DB-backed ndb, the iterator registers as
// a version reader on imm.version so that concurrent PruneVersionsTo cannot
// delete nodes this iterator still walks. The reader is released on Close.
func NewIteratorWithNDB(imm *ImmutableTree, start, end []byte, ascending bool, mtree *MutableTree) *Iterator {
	// Prefer the immutable tree's own ndb (set by GetImmutable) so we can
	// track the specific version being read. Fall back to the mutable tree's
	// ndb only for value resolution (no version tracking in that case).
	ndb := imm.ndb
	var trackVersion int64
	if ndb != nil {
		// version > 0 mirrors Close's decrement guard: registering version 0
		// would never be released.
		if imm.version > 0 {
			trackVersion = imm.version
			ndb.incrVersionReaders(trackVersion)
		}
	} else if mtree != nil {
		ndb = mtree.ndb
	}
	itr := newIterator(imm.root, start, end, ascending, ndb, trackVersion)
	// Committed snapshot: resolve values DB-only (never the writer's pendingVals
	// buffer), so iterating concurrently with the writer cannot race SaveValue.
	if ndb != nil {
		itr.valueResolver = ndb.getCommittedValue
	}
	return itr
}

// Iterator returns an iterator over [start, end) in the given direction.
// MutableTree iterators walk the in-memory working tree and do not take a
// version reader: pruning rejects toVersion >= latest, and a working tree
// loaded at an older version is protected by the working-tree-reader guard
// (PruneVersionsTo rejects t.version <= toVersion with ErrActiveReaders).
func (t *MutableTree) Iterator(start, end []byte, ascending bool) (*Iterator, error) {
	itr := newIterator(t.root, start, end, ascending, t.ndb, 0)
	// Working-tree iterator: resolve through GetValue (pendingVals first) so a
	// Set issued earlier this session is visible (read-your-writes). Single
	// writer goroutine, so the pendingVals access does not race.
	itr.valueResolver = t.ndb.GetValue
	return itr, nil
}

// ImmutableTree.Iterator returns an iterator over [start, end).
//
// When the immutable tree is DB-backed (t.ndb != nil), the iterator registers
// as a version reader so concurrent pruning cannot delete the nodes it walks.
// The reader is released on Close. For snapshot-only trees without an ndb,
// values are resolved via t.valueResolver (no version tracking needed).
func (t *ImmutableTree) Iterator(start, end []byte, ascending bool) (*Iterator, error) {
	var trackVersion int64
	// version > 0 mirrors Close's decrement guard: registering version 0
	// would never be released.
	if t.ndb != nil && t.version > 0 {
		trackVersion = t.version
		t.ndb.incrVersionReaders(trackVersion)
	}
	itr := newIterator(t.root, start, end, ascending, t.ndb, trackVersion)
	// Use the snapshot's own resolver (DB-only for newImmutable-derived trees);
	// if unset but DB-backed, fall back to DB-only committed resolution. Either
	// way a committed snapshot never touches the writer's pendingVals buffer.
	if t.valueResolver != nil {
		itr.valueResolver = t.valueResolver
	} else if t.ndb != nil {
		itr.valueResolver = t.ndb.getCommittedValue
	}
	return itr, nil
}

// IterateRange iterates over [start, end) calling fn. Stops early if fn
// returns true. A value-resolution or node-load failure is returned as err —
// fn is never called with the failing row, and stopped is meaningless when
// err != nil.
func (t *ImmutableTree) IterateRange(start, end []byte, ascending bool, fn func(key, value []byte) bool) (stopped bool, err error) {
	itr, err := t.Iterator(start, end, ascending)
	if err != nil {
		return false, err
	}
	defer itr.Close()
	for itr.Valid() {
		key := itr.Key()
		value := itr.Value()
		if err := itr.Error(); err != nil {
			return false, err
		}
		if fn(key, value) {
			return true, nil
		}
		itr.Next()
	}
	return false, itr.Error()
}

// IterateRange on MutableTree. Same contract as ImmutableTree.IterateRange.
func (t *MutableTree) IterateRange(start, end []byte, ascending bool, fn func(key, value []byte) bool) (stopped bool, err error) {
	itr, err := t.Iterator(start, end, ascending)
	if err != nil {
		return false, err
	}
	defer itr.Close()
	for itr.Valid() {
		key := itr.Key()
		value := itr.Value()
		if err := itr.Error(); err != nil {
			return false, err
		}
		if fn(key, value) {
			return true, nil
		}
		itr.Next()
	}
	return false, itr.Error()
}
