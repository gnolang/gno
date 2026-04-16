package bptree

// ValueResolver resolves a valueKey to the raw value bytes.
type ValueResolver func(vk []byte) ([]byte, error)

// ImmutableTree is a read-only snapshot of the tree at a specific version.
// It is safe for concurrent reads. Created by MutableTree.GetImmutable()
// (Phase 3) or by snapshotting the root after SaveVersion.
type ImmutableTree struct {
	root          Node
	version       int64
	valueResolver ValueResolver // resolves valueKeys to raw values

	// ndb is retained so Close() can decrement this version's reader count
	// and to prevent pruning while the snapshot is in use. nil for snapshots
	// that do not correspond to a saved version in a nodeDB (e.g. proof
	// scratch trees or purely in-memory snapshots).
	ndb *nodeDB
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
// It is safe to call multiple times; subsequent calls are no-ops. Callers
// that hold an ImmutableTree obtained from GetImmutable or immutableForProof
// MUST call Close when done; otherwise PruneVersionsTo on that version will
// return ErrActiveReaders indefinitely. See Finding #30.
func (t *ImmutableTree) Close() error {
	if t.ndb != nil && t.version > 0 {
		t.ndb.decrVersionReaders(t.version)
	}
	t.ndb = nil
	return nil
}

// resolveValue resolves a valueKey to raw bytes if a resolver is set.
func (t *ImmutableTree) resolveValue(vk []byte) ([]byte, error) {
	if t.valueResolver != nil {
		return t.valueResolver(vk)
	}
	return nil, ErrKeyDoesNotExist
}

// Get returns the value for a key, or nil if not found.
func (t *ImmutableTree) Get(key []byte) ([]byte, error) {
	if t.root == nil {
		return nil, nil
	}
	_, _, vk, found := treeLookup(t.root, key)
	if !found {
		return nil, nil
	}
	return t.resolveValue(vk)
}

// Has returns true if the key exists.
func (t *ImmutableTree) Has(key []byte) (bool, error) {
	if t.root == nil {
		return false, nil
	}
	_, _, _, found := treeLookup(t.root, key)
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
	key, _, vk := treeGetByIndex(t.root, index)
	val, err := t.resolveValue(vk)
	return key, val, err
}

// GetWithIndex returns the index, value, and whether the key was found.
func (t *ImmutableTree) GetWithIndex(key []byte) (int64, []byte, error) {
	if t.root == nil {
		return 0, nil, nil
	}
	idx, _, vk, found := treeGetWithIndex(t.root, key)
	if !found {
		return idx, nil, nil
	}
	val, err := t.resolveValue(vk)
	return idx, val, err
}

// Iterate calls fn for each key-value pair in sorted order.
// If a value resolver is set, values are resolved to actual bytes; a
// resolver error stops iteration and is returned. If no resolver is
// set, this falls back to yielding value hashes (legacy behavior).
func (t *ImmutableTree) Iterate(fn func(key []byte, value []byte) bool) (bool, error) {
	if t.root == nil {
		return false, nil
	}
	if t.valueResolver != nil {
		var resolveErr error
		stopped := iterateNodeResolved(t.root, func(key, vk []byte) bool {
			val, err := t.valueResolver(vk)
			if err != nil {
				resolveErr = err
				return true // stop
			}
			return fn(key, val)
		})
		return stopped, resolveErr
	}
	return iterateNode(t.root, fn), nil
}
