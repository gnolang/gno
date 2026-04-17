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
	stack   []stackEntry
	leaf    *LeafNode
	leafIdx int // current position within leaf
	valid   bool
	err     error
	closed  bool

	// Per-leaf value cache (Finding #16). When the iterator enters a leaf
	// and the caller invokes Value() for the first time, every value in
	// the leaf is resolved up front and stashed here. Subsequent Value()
	// calls against the same leaf return from cache, trading one DB
	// round-trip per Next() for one batched resolution per leaf. The
	// cache is cleared by setLeaf() on every leaf transition so each
	// cache entry is valid only for the leaf currently pointed at by
	// `it.leaf`.
	leafValues       [B][]byte
	leafValuesLoaded bool

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

// setLeaf switches the iterator to leaf n, invalidating the per-leaf
// value cache so a subsequent Value() call reloads from the resolver
// for the new leaf (Finding #16).
func (it *Iterator) setLeaf(n *LeafNode) {
	it.leaf = n
	it.leafValuesLoaded = false
	for i := range it.leafValues {
		it.leafValues[i] = nil
	}
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
			it.setLeaf(n)
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
				// searchInner returns the first j where keys[j] > end, so
				// `end` itself lives in child[childIdx]. When end matches
				// a separator exactly (end == keys[childIdx-1], which is
				// the smallest key of child[childIdx]), the rightmost key
				// strictly less than end is in child[childIdx-1], not
				// child[childIdx]. Without the adjustment we would descend
				// into child[childIdx], find the leaf's first key equals
				// end, land on leafIdx = -1, and climb back via
				// prevLeaf() — correct but one extra DB load per seek.
				// See Finding #33.
				if childIdx > 0 && bytes.Equal(n.keys[childIdx-1], it.end) {
					childIdx--
				}
				if childIdx >= n.NumChildren() {
					childIdx = n.NumChildren() - 1
				}
			} else {
				childIdx = n.NumChildren() - 1
			}
			it.stack = append(it.stack, stackEntry{inner: n, childIdx: childIdx})
			node = n.getChild(childIdx)
		case *LeafNode:
			it.setLeaf(n)
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
			it.setLeaf(n)
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
			it.setLeaf(n)
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
	// Resolver wiring is validated per-slot inside valueAt: inline
	// slots don't need one, external slots surface ErrNoValueResolver
	// (captured via loadLeafValues on the first Value() call).
	// Leaf-level value prefetch (Finding #16). The first Value() call
	// after entering a leaf populates the cache for every occupied slot
	// in a single pass; subsequent calls return the cached slice. For
	// the common scan pattern (Value() once per Next()) this amortises
	// per-call DB / resolver overhead at the cost of at most one eager
	// resolution per leaf. Error on any slot invalidates the iterator
	// and propagates via Error().
	if !it.leafValuesLoaded {
		if !it.loadLeafValues() {
			return nil
		}
	}
	return it.leafValues[it.leafIdx]
}

// loadLeafValues resolves values for the slots the iterator will actually
// visit on the current leaf and caches them in it.leafValues. Returns
// false and invalidates the iterator on the first resolution error.
//
// The visit window is computed from the iterator's direction and bounds:
// for a range scan that only touches part of a leaf, this avoids the
// wasted resolves (one DB round-trip each on disk-backed backends) that
// the earlier "prefetch every occupied slot" implementation paid.
// Full iteration (start and end both nil) still fetches the whole
// leaf — the window collapses to [0, numKeys).
func (it *Iterator) loadLeafValues() bool {
	lo, hi := it.leafVisitWindow()
	resolver := it.valueResolver
	if resolver == nil && it.ndb != nil {
		resolver = it.ndb.GetValue
	}
	for i := lo; i < hi; i++ {
		val, err := it.leaf.valueAt(i, resolver)
		if err != nil {
			it.err = err
			it.valid = false
			return false
		}
		it.leafValues[i] = val
	}
	it.leafValuesLoaded = true
	return true
}

// leafVisitWindow returns the [lo, hi) range of leaf slot indices that
// the iterator will visit on the current leaf, given its direction and
// start/end bounds.
//
// Ascending: lo is the current leafIdx (earliest slot to be visited —
// Next only moves forward); hi is numKeys clipped by end (exclusive).
// Descending: hi is leafIdx+1 (leafIdx is the highest slot — Next moves
// backward); lo is 0 clipped up by start (inclusive).
//
// The clip consults searchLeaf, costing log₂B = 5 comparisons per leaf
// transition — strictly dominated by the DB resolves it eliminates.
func (it *Iterator) leafVisitWindow() (lo, hi int) {
	n := int(it.leaf.numKeys)
	if it.ascending {
		lo = it.leafIdx
		hi = n
		if it.end != nil {
			// searchLeaf returns the first index whose key is >= end.
			// end is exclusive, so that index is the first slot NOT
			// visited; everything strictly below it is in-range.
			p, _ := searchLeaf(it.leaf, it.end)
			if p < hi {
				hi = p
			}
		}
	} else {
		hi = it.leafIdx + 1
		lo = 0
		if it.start != nil {
			// searchLeaf returns the first index whose key is >= start.
			// start is inclusive, so that index is the first (lowest)
			// slot we'll still visit on the descent.
			p, _ := searchLeaf(it.leaf, it.start)
			if p > lo {
				lo = p
			}
		}
	}
	return lo, hi
}

// Error reports the first error encountered during iteration or value
// resolution. Callers MUST check Error() after Valid() returns false to
// distinguish a clean end-of-range walk from a resolver / DB failure that
// truncated iteration silently. See Finding #34.
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
// When the tree has no value-resolution path configured (no ndb and no
// memValues — e.g. a MutableTree built via a bare struct literal), the
// iterator is returned with err = ErrNoValueResolver and valid = false;
// Error() reports the misconfiguration and the caller avoids silent
// nil-value reads. See Finding #35.
//
// No version-reader registration: MutableTree is contracted as
// single-goroutine (see the struct doc). Protecting against a
// concurrent PruneVersionsTo on the same tree would defend against a
// contract violation the caller has already made; callers that need a
// snapshot stable across threads must obtain an ImmutableTree via
// GetImmutable instead, which does register a reader.
func (t *MutableTree) Iterator(start, end []byte, ascending bool) (*Iterator, error) {
	itr := newIterator(t.root, start, end, ascending, t.ndb, 0)
	if t.ndb == nil {
		if t.memValues == nil {
			itr.err = ErrNoValueResolver
			itr.valid = false
			return itr, ErrNoValueResolver
		}
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
// the iterator resolves values via that resolver. If none is set, the
// iterator is returned with err = ErrNoValueResolver and valid = false
// (see Finding #35) — do not silently yield hashes or nils.
// For DB-backed trees, use NewIteratorWithNDB instead.
func (t *ImmutableTree) Iterator(start, end []byte, ascending bool) (*Iterator, error) {
	itr := newIterator(t.root, start, end, ascending, nil, 0)
	if t.valueResolver == nil {
		itr.err = ErrNoValueResolver
		itr.valid = false
		return itr, ErrNoValueResolver
	}
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
