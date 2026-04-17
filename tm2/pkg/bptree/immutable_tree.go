package bptree

import "sync"

// ValueResolver resolves a valueKey to the raw value bytes.
type ValueResolver func(vk []byte) ([]byte, error)

// ImmutableTree is a read-only snapshot of the tree at a specific version.
//
// Concurrency: ImmutableTree is safe for concurrent reads across
// goroutines so long as no mutator on the originating MutableTree
// still holds a reference to any of the snapshot's nodes. Snapshots
// returned by GetImmutable are freshly loaded from the DB (children
// begin lazy, guarded by InnerNode.childMu). Snapshots returned by
// MutableTree.Snapshot clone the root so subsequent COW descents on
// the MutableTree cannot reach back into the snapshot. See Finding #7.
//
// Callers that obtain an ImmutableTree from GetImmutable or
// immutableForProof MUST call Close when done — the tree registers as
// an active version reader at construction and a missing Close blocks
// PruneVersionsTo with ErrActiveReaders until the process exits.
type ImmutableTree struct {
	root          Node
	version       int64
	valueResolver ValueResolver // resolves valueKeys to raw values

	// ndb is retained so Close() can decrement this version's reader count
	// and to prevent pruning while the snapshot is in use. nil for snapshots
	// that do not correspond to a saved version in a nodeDB (e.g. proof
	// scratch trees or purely in-memory snapshots).
	ndb *nodeDB

	// closeOnce guards decrVersionReaders so double-Close or concurrent
	// Close calls do not over-decrement the reader count (which would
	// allow a prune to proceed against a version that still has live
	// snapshots). See Finding #45.
	closeOnce sync.Once
}

// NewImmutableTree creates an ImmutableTree from a root node and version.
func NewImmutableTree(root Node, version int64) *ImmutableTree {
	return &ImmutableTree{root: root, version: version}
}

// SetValueResolver sets the function used to resolve valueKeys to raw values.
func (t *ImmutableTree) SetValueResolver(resolver ValueResolver) {
	t.valueResolver = resolver
}

// Close releases the version-reader reservation held by this snapshot.
// Guarded by sync.Once so repeated or concurrent Close calls decrement
// the reader count exactly once — a double decrement would allow a
// prune to proceed while other snapshots of the same version are still
// live. Callers that hold an ImmutableTree obtained from GetImmutable or
// immutableForProof MUST call Close when done; otherwise
// PruneVersionsTo on that version will return ErrActiveReaders
// indefinitely. See Findings #30 and #45.
func (t *ImmutableTree) Close() error {
	t.closeOnce.Do(func() {
		if t.ndb != nil && t.version > 0 {
			t.ndb.decrVersionReaders(t.version)
		}
		t.ndb = nil
	})
	return nil
}

// resolveValue resolves a valueKey to raw bytes via the configured resolver.
// Returns ErrNoValueResolver when no resolver is set; this is distinct from
// ErrKeyDoesNotExist so callers can tell a missing key from a misconfigured
// tree (e.g. an ImmutableTree constructed outside GetImmutable/Snapshot).
// See Findings #10 and #11.
func (t *ImmutableTree) resolveValue(vk []byte) ([]byte, error) {
	if t.valueResolver != nil {
		return t.valueResolver(vk)
	}
	return nil, ErrNoValueResolver
}

// Get returns the value for a key, or nil if not found.
func (t *ImmutableTree) Get(key []byte) ([]byte, error) {
	if t.root == nil {
		return nil, nil
	}
	leaf, slot, found := treeLookup(t.root, key)
	if !found {
		return nil, nil
	}
	return leaf.valueAt(slot, t.valueResolver)
}

// Has returns true if the key exists.
func (t *ImmutableTree) Has(key []byte) (bool, error) {
	if t.root == nil {
		return false, nil
	}
	_, _, found := treeLookup(t.root, key)
	return found, nil
}

// Size returns the total number of key-value pairs.
func (t *ImmutableTree) Size() int64 {
	if t.root == nil {
		return 0
	}
	return nodeSize(t.root)
}

// Height returns the tree height.
func (t *ImmutableTree) Height() int8 {
	if t.root == nil {
		return 0
	}
	return int8(nodeHeight(t.root))
}

// Hash returns the root hash. Returns SHA256("") for empty trees, matching IAVL.
func (t *ImmutableTree) Hash() []byte {
	if t.root == nil {
		return emptyHash()
	}
	h := t.root.Hash()
	return h[:]
}

// Version returns the version of this snapshot.
func (t *ImmutableTree) Version() int64 {
	return t.version
}

// IsEmpty returns true if the tree has no keys.
func (t *ImmutableTree) IsEmpty() bool {
	return t.root == nil
}

// GetByIndex returns the key and value at the given index.
func (t *ImmutableTree) GetByIndex(index int64) ([]byte, []byte, error) {
	if t.root == nil || index < 0 || index >= t.Size() {
		return nil, nil, ErrKeyDoesNotExist
	}
	leaf, slot := treeGetByIndex(t.root, index)
	val, err := leaf.valueAt(slot, t.valueResolver)
	return leaf.keys[slot], val, err
}

// GetWithIndex returns the index, value, and whether the key was found.
func (t *ImmutableTree) GetWithIndex(key []byte) (int64, []byte, error) {
	if t.root == nil {
		return 0, nil, nil
	}
	idx, leaf, slot, found := treeGetWithIndex(t.root, key)
	if !found {
		return idx, nil, nil
	}
	val, err := leaf.valueAt(slot, t.valueResolver)
	return idx, val, err
}

// Iterate calls fn for each key-value pair in sorted order. Values are
// resolved to actual bytes via the configured resolver; a resolver error
// stops iteration and is returned. If no resolver is set AND the tree
// contains any external slots, Iterate returns (false, ErrNoValueResolver);
// an all-inline tree can iterate without a resolver. See Finding #11.
func (t *ImmutableTree) Iterate(fn func(key []byte, value []byte) bool) (bool, error) {
	if t.root == nil {
		return false, nil
	}
	var resolveErr error
	stopped := iterateNodeResolved(t.root, func(key []byte, leaf *LeafNode, slot int) bool {
		val, err := leaf.valueAt(slot, t.valueResolver)
		if err != nil {
			resolveErr = err
			return true // stop
		}
		return fn(key, val)
	})
	return stopped, resolveErr
}
