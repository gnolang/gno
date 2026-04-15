package bptree

import (
	"bytes"
	"crypto/sha256"
	"fmt"

	dbm "github.com/gnolang/gno/tm2/pkg/db"
)

// MutableTree is the working tree supporting Set, Get, Has, Remove,
// SaveVersion, LoadVersion, and Rollback.
type MutableTree struct {
	root      Node   // nil for empty tree
	lastSaved Node   // snapshot for rollback (set by SaveVersion)
	size      int64  // total key count in working tree
	version   int64  // last saved version

	ndb            *nodeDB // nil for in-memory only (Phase 2 compat)
	initialVersion uint64
	logger         Logger

	// In-memory value store for Phase 2 compat (no ndb).
	// Maps SHA256(value) hex -> raw value bytes.
	memValues map[Hash][]byte
}

// NewMutableTreeMem creates an in-memory MutableTree (no DB).
func NewMutableTreeMem() *MutableTree {
	return &MutableTree{logger: NewNopLogger(), memValues: make(map[Hash][]byte)}
}

// NewMutableTreeWithDB creates a DB-backed MutableTree.
func NewMutableTreeWithDB(db dbm.DB, cacheSize int, logger Logger, options ...Option) *MutableTree {
	opts := DefaultOptions()
	for _, o := range options {
		o(&opts)
	}
	if logger == nil {
		logger = NewNopLogger()
	}
	ndb := newNodeDB(db, cacheSize, logger, opts)
	return &MutableTree{
		ndb:            ndb,
		logger:         logger,
		initialVersion: opts.InitialVersion,
	}
}

// Set inserts or updates a key-value pair. Returns true if the key
// already existed (update), false if it was a new insert.
func (t *MutableTree) Set(key, value []byte) (updated bool, err error) {
	if len(key) == 0 {
		return false, ErrEmptyKey
	}
	if value == nil {
		return false, fmt.Errorf("value must not be nil")
	}

	if t.root == nil {
		leaf := &LeafNode{miniTree: NewMiniMerkle()}
		leaf.keys[0] = copyKey(key)
		valueHash := sha256.Sum256(value)
		leaf.valueHashes[0] = valueHash
		leaf.numKeys = 1
		leaf.RebuildMiniMerkle()
		t.root = leaf
		t.size = 1

		// Save value out-of-line
		if t.ndb != nil {
			if err := t.ndb.SaveValue(value, valueHash); err != nil {
				return false, err
			}
		} else if t.memValues != nil {
			valCopy := make([]byte, len(value))
			copy(valCopy, value)
			t.memValues[valueHash] = valCopy
		}
		return false, nil
	}

	valueHash := sha256.Sum256(value)
	newRoot, updated := treeInsert(t.root, key, valueHash)
	t.root = newRoot
	if !updated {
		t.size++
	}

	// Save value out-of-line
	if t.ndb != nil {
		if err := t.ndb.SaveValue(value, valueHash); err != nil {
			return updated, err
		}
	} else if t.memValues != nil {
		valCopy := make([]byte, len(value))
		copy(valCopy, value)
		t.memValues[valueHash] = valCopy
	}
	return updated, nil
}

// Get retrieves the value for a key.
func (t *MutableTree) Get(key []byte) ([]byte, error) {
	if t.root == nil {
		return nil, nil
	}
	_, vh, found := treeLookup(t.root, key)
	if !found {
		return nil, nil
	}
	return t.resolveValue(vh)
}

// resolveValue resolves a value hash to actual bytes.
func (t *MutableTree) resolveValue(vh Hash) ([]byte, error) {
	if t.ndb != nil {
		return t.ndb.GetValue(vh)
	}
	if t.memValues != nil {
		if val, ok := t.memValues[vh]; ok {
			return val, nil
		}
	}
	return vh[:], nil
}

// Has returns true if the key exists in the tree.
func (t *MutableTree) Has(key []byte) (bool, error) {
	if t.root == nil {
		return false, nil
	}
	_, _, found := treeLookup(t.root, key)
	return found, nil
}

// Remove removes a key from the tree. Returns the old value and
// whether the key was found.
func (t *MutableTree) Remove(key []byte) ([]byte, bool, error) {
	if t.root == nil {
		return nil, false, nil
	}
	newRoot, oldVH, found := treeRemove(t.root, key)
	if !found {
		return nil, false, nil
	}
	t.root = newRoot
	t.size--

	val, err := t.resolveValue(oldVH)
	if err != nil {
		return nil, true, err
	}
	return val, true, nil
}

// SaveVersion persists the current tree state as a new version.
// Returns (rootHash, version, error).
func (t *MutableTree) SaveVersion() ([]byte, int64, error) {
	version := t.WorkingVersion()

	if t.ndb == nil {
		// In-memory only: just snapshot
		t.lastSaved = t.root
		t.version = version
		if t.root == nil {
			return emptyHash(), version, nil
		}
		h := t.root.Hash()
		return h[:], version, nil
	}

	// If this version already exists, verify the hash matches.
	// This prevents accidentally overwriting a version with different data.
	if t.ndb.VersionExists(version) {
		existingNK, existingHash, err := t.ndb.GetRoot(version)
		if err != nil {
			return nil, 0, err
		}
		var newHash []byte
		if t.root != nil {
			// Need to compute the working hash to compare
			h := t.root.Hash()
			newHash = h[:]
		} else {
			newHash = emptyHash()
		}
		// Compare: existing empty vs new non-empty, or hash mismatch
		existingEmpty := existingNK == nil
		newEmpty := t.root == nil
		if existingEmpty != newEmpty || !bytes.Equal(existingHash, newHash) {
			return nil, 0, fmt.Errorf("version %d already exists with a different hash", version)
		}
		// Same hash — idempotent save, skip
		t.version = version
		t.lastSaved = t.root
		return newHash, version, nil
	}

	t.ndb.ResetNonce()

	// Assign NodeKeys and save all dirty nodes
	if t.root != nil {
		if err := t.saveNode(t.root, version); err != nil {
			return nil, 0, err
		}
	}

	// Save root reference
	var rootHash []byte
	if t.root != nil {
		h := t.root.Hash()
		rootHash = h[:]
		if err := t.ndb.SaveRoot(version, t.root.GetNodeKey(), rootHash); err != nil {
			return nil, 0, err
		}
	} else {
		rootHash = emptyHash()
		if err := t.ndb.SaveRoot(version, nil, rootHash); err != nil {
			return nil, 0, err
		}
	}

	// Commit batch
	if err := t.ndb.Commit(); err != nil {
		return nil, 0, err
	}

	t.version = version
	t.lastSaved = t.root
	t.ndb.setLatestVersion(version)
	if t.ndb.getFirstVersion() == 0 {
		t.ndb.setFirstVersion(version)
	}

	return rootHash, version, nil
}

// saveNode recursively assigns NodeKeys and saves dirty nodes.
func (t *MutableTree) saveNode(node Node, version int64) error {
	if node.GetNodeKey() != nil {
		return nil // already saved
	}

	// For inner nodes, save children first (bottom-up)
	if inner, ok := node.(*InnerNode); ok {
		for i := 0; i < inner.NumChildren(); i++ {
			child := inner.getChild(i)
			if child != nil {
				if err := t.saveNode(child, version); err != nil {
					return err
				}
				// Update child reference and hash after save
				inner.children[i] = child.GetNodeKey().GetKey()
				inner.childHashes[i] = child.Hash()
			}
		}
		inner.RebuildMiniMerkle()
	}

	// Rebuild leaf mini merkle (may already be done, but ensure correctness)
	if leaf, ok := node.(*LeafNode); ok {
		leaf.RebuildMiniMerkle()
	}

	// Assign NodeKey
	nk := t.ndb.NextNodeKey(version)
	node.SetNodeKey(nk)

	return t.ndb.SaveNode(node)
}

// Load loads the latest version from the DB.
func (t *MutableTree) Load() (int64, error) {
	if t.ndb == nil {
		return 0, nil
	}
	if err := t.ndb.discoverVersions(); err != nil {
		return 0, err
	}
	latest := t.ndb.getLatestVersion()
	if latest == 0 {
		return 0, nil
	}
	return t.LoadVersion(latest)
}

// LoadVersion loads a specific version from the DB.
func (t *MutableTree) LoadVersion(version int64) (int64, error) {
	if t.ndb == nil {
		return 0, nil
	}
	if version <= 0 {
		return 0, nil
	}

	nkBytes, _, err := t.ndb.GetRoot(version)
	if err != nil {
		return 0, err
	}

	if nkBytes == nil {
		// Empty tree at this version
		t.root = nil
		t.size = 0
		t.version = version
		t.lastSaved = nil
		return version, nil
	}

	root, err := t.loadNode(nkBytes)
	if err != nil {
		return 0, fmt.Errorf("loading root: %w", err)
	}

	t.root = root
	t.size = nodeSize(root)
	t.version = version
	t.lastSaved = root
	t.ndb.setLatestVersion(version)
	return version, nil
}

// LoadVersionForOverwriting loads a version and deletes all newer versions.
func (t *MutableTree) LoadVersionForOverwriting(version int64) error {
	if t.ndb == nil {
		return nil
	}
	// Remember the old latest version before LoadVersion overwrites it
	oldLatest := t.ndb.getLatestVersion()

	_, err := t.LoadVersion(version)
	if err != nil {
		return err
	}

	// Delete all versions after the target, using the old latest as the upper bound
	for v := version + 1; v <= oldLatest; v++ {
		if t.ndb.VersionExists(v) {
			if err := t.ndb.DeleteRoot(v); err != nil {
				return err
			}
		}
	}
	if err := t.ndb.Commit(); err != nil {
		return err
	}
	t.ndb.setLatestVersion(version)
	return nil
}

// loadNode loads a node from the DB. Children are loaded lazily via
// getChild (which uses the ndb reference set during deserialization).
func (t *MutableTree) loadNode(nkBytes []byte) (Node, error) {
	return t.ndb.GetNode(nkBytes)
}

// GetImmutable returns an ImmutableTree for the given version.
func (t *MutableTree) GetImmutable(version int64) (*ImmutableTree, error) {
	if t.ndb == nil {
		if version == t.version && t.lastSaved != nil {
			imm := NewImmutableTree(t.lastSaved, version)
			if t.memValues != nil {
				imm.valueResolver = func(vh Hash) ([]byte, error) {
					val, ok := t.memValues[vh]
					if !ok {
						return nil, fmt.Errorf("value not found in memValues")
					}
					return val, nil
				}
			}
			return imm, nil
		}
		return nil, ErrVersionDoesNotExist
	}

	nkBytes, _, err := t.ndb.GetRoot(version)
	if err != nil {
		return nil, err
	}
	if nkBytes == nil {
		return NewImmutableTree(nil, version), nil
	}

	root, err := t.loadNode(nkBytes)
	if err != nil {
		return nil, err
	}
	imm := NewImmutableTree(root, version)
	imm.valueResolver = func(vh Hash) ([]byte, error) {
		return t.ndb.GetValue(vh)
	}
	return imm, nil
}

// GetVersioned returns the value for a key at a specific version.
func (t *MutableTree) GetVersioned(key []byte, version int64) ([]byte, error) {
	imm, err := t.GetImmutable(version)
	if err != nil {
		return nil, nil // match IAVL behavior: silent nil for missing version
	}
	return imm.Get(key)
}

// DeleteVersionsTo deletes versions from first to toVersion (inclusive),
// including orphaned nodes via dual-tree-walk pruning.
func (t *MutableTree) DeleteVersionsTo(toVersion int64) error {
	return t.PruneVersionsTo(toVersion)
}

// DeleteVersionsFrom deletes all versions >= fromVersion.
// Stub for Phase 3.
func (t *MutableTree) DeleteVersionsFrom(fromVersion int64) error {
	if t.ndb == nil {
		return nil
	}
	latest := t.ndb.getLatestVersion()
	for v := fromVersion; v <= latest; v++ {
		if t.ndb.hasVersionReaders(v) {
			return fmt.Errorf("%w: version %d", ErrActiveReaders, v)
		}
	}
	for v := fromVersion; v <= latest; v++ {
		if t.ndb.VersionExists(v) {
			if err := t.ndb.DeleteRoot(v); err != nil {
				return err
			}
		}
	}
	if err := t.ndb.Commit(); err != nil {
		return err
	}
	if fromVersion <= latest {
		newLatest := fromVersion - 1
		t.ndb.setLatestVersion(newLatest)

		// If the working tree's version was deleted, reset to the new latest.
		if t.version >= fromVersion {
			if newLatest > 0 && t.ndb.VersionExists(newLatest) {
				if _, err := t.LoadVersion(newLatest); err != nil {
					return err
				}
			} else {
				t.root = nil
				t.lastSaved = nil
				t.size = 0
				t.version = newLatest
			}
		}
	}
	return nil
}

// Size returns the total number of key-value pairs.
func (t *MutableTree) Size() int64 { return t.size }

// IsEmpty returns true if the tree has no keys.
func (t *MutableTree) IsEmpty() bool { return t.root == nil }

// Hash returns the root hash of the last saved version.
// Returns SHA256("") for empty trees, matching IAVL behavior.
func (t *MutableTree) Hash() []byte {
	if t.lastSaved == nil {
		return emptyHash()
	}
	h := t.lastSaved.Hash()
	return h[:]
}

// WorkingHash computes the hash of the current unsaved working tree.
// Returns SHA256("") for empty trees, matching IAVL behavior.
func (t *MutableTree) WorkingHash() []byte {
	if t.root == nil {
		return emptyHash()
	}
	h := t.root.Hash()
	return h[:]
}

// WorkingVersion returns the version that will be used by the next SaveVersion.
func (t *MutableTree) WorkingVersion() int64 {
	if t.version == 0 && t.initialVersion > 0 {
		return int64(t.initialVersion)
	}
	return t.version + 1
}

// Version returns the last saved version.
func (t *MutableTree) Version() int64 { return t.version }

// VersionExists returns true if the given version exists.
func (t *MutableTree) VersionExists(version int64) bool {
	if t.ndb != nil {
		return t.ndb.VersionExists(version)
	}
	return version == t.version && t.lastSaved != nil
}

// AvailableVersions returns all available version numbers.
func (t *MutableTree) AvailableVersions() []int {
	if t.ndb != nil {
		return t.ndb.AvailableVersions()
	}
	if t.lastSaved != nil {
		return []int{int(t.version)}
	}
	return nil
}

// SetInitialVersion sets the version number for the first SaveVersion.
func (t *MutableTree) SetInitialVersion(version uint64) {
	t.initialVersion = version
}

// SetCommitting signals that a commit is in progress.
func (t *MutableTree) SetCommitting() {
	if t.ndb != nil {
		t.ndb.SetCommitting()
	}
}

// UnsetCommitting signals that a commit has finished.
func (t *MutableTree) UnsetCommitting() {
	if t.ndb != nil {
		t.ndb.UnsetCommitting()
	}
}

// Rollback discards all mutations since the last save.
func (t *MutableTree) Rollback() {
	t.root = t.lastSaved
	if t.root != nil {
		t.size = nodeSize(t.root)
	} else {
		t.size = 0
	}
}

// Height returns the tree height.
func (t *MutableTree) Height() int8 {
	if t.root == nil {
		return 0
	}
	return int8(nodeHeight(t.root))
}

// GetValueByHash resolves a value hash to the raw value bytes.
func (t *MutableTree) GetValueByHash(vh Hash) ([]byte, error) {
	if t.ndb != nil {
		return t.ndb.GetValue(vh)
	}
	if t.memValues != nil {
		val, ok := t.memValues[vh]
		if !ok {
			return nil, fmt.Errorf("value not found")
		}
		return val, nil
	}
	return vh[:], nil
}

// Close closes the tree and its underlying DB resources.
func (t *MutableTree) Close() error {
	if t.ndb != nil {
		return t.ndb.Close()
	}
	return nil
}

// GetByIndex returns the key and value at the given zero-based index.
func (t *MutableTree) GetByIndex(index int64) ([]byte, []byte, error) {
	if t.root == nil || index < 0 || index >= t.size {
		return nil, nil, ErrKeyDoesNotExist
	}
	key, vh := treeGetByIndex(t.root, index)
	val, err := t.resolveValue(vh)
	return key, val, err
}

// GetWithIndex returns the index, value, and whether the key was found.
func (t *MutableTree) GetWithIndex(key []byte) (int64, []byte, error) {
	if t.root == nil {
		return 0, nil, nil
	}
	idx, vh, found := treeGetWithIndex(t.root, key)
	if !found {
		return idx, nil, nil
	}
	val, err := t.resolveValue(vh)
	return idx, val, err
}

// Iterate calls fn for each key-value pair in sorted order.
// Values are resolved from the value store (DB or memValues).
func (t *MutableTree) Iterate(fn func(key []byte, value []byte) bool) (bool, error) {
	if t.root == nil {
		return false, nil
	}
	if t.ndb != nil || t.memValues != nil {
		return iterateNode(t.root, func(key, valueHash []byte) bool {
			var vh Hash
			copy(vh[:], valueHash)
			val, err := t.resolveValue(vh)
			if err != nil {
				return true
			}
			return fn(key, val)
		}), nil
	}
	return iterateNode(t.root, fn), nil
}

// --- helpers ---

func treeLookup(node Node, key []byte) (*LeafNode, Hash, bool) {
	for {
		switch n := node.(type) {
		case *LeafNode:
			pos, found := searchLeaf(n, key)
			if !found {
				return n, Hash{}, false
			}
			return n, n.valueHashes[pos], true
		case *InnerNode:
			idx := searchInner(n, key)
			child := n.getChild(idx)
			if child == nil {
				return nil, Hash{}, false
			}
			node = child
		default:
			panic("unknown node type")
		}
	}
}

func treeGetByIndex(node Node, index int64) ([]byte, Hash) {
	switch n := node.(type) {
	case *LeafNode:
		return n.keys[index], n.valueHashes[index]
	case *InnerNode:
		offset := int64(0)
		for i := 0; i < n.NumChildren(); i++ {
			childSize := n.childSizes[i]
			if index < offset+childSize {
				child := n.getChild(i)
				return treeGetByIndex(child, index-offset)
			}
			offset += childSize
		}
		panic("index out of range in treeGetByIndex")
	default:
		panic("unknown node type")
	}
}

func treeGetWithIndex(node Node, key []byte) (int64, Hash, bool) {
	switch n := node.(type) {
	case *LeafNode:
		pos, found := searchLeaf(n, key)
		if !found {
			return int64(pos), Hash{}, false
		}
		return int64(pos), n.valueHashes[pos], true
	case *InnerNode:
		childIdx := searchInner(n, key)
		offset := int64(0)
		for i := 0; i < childIdx; i++ {
			offset += n.childSizes[i]
		}
		child := n.getChild(childIdx)
		idx, vh, found := treeGetWithIndex(child, key)
		return offset + idx, vh, found
	default:
		panic("unknown node type")
	}
}

func iterateNode(node Node, fn func(key, value []byte) bool) bool {
	switch n := node.(type) {
	case *LeafNode:
		for i := 0; i < int(n.numKeys); i++ {
			if fn(n.keys[i], n.valueHashes[i][:]) {
				return true
			}
		}
		return false
	case *InnerNode:
		for i := 0; i < n.NumChildren(); i++ {
			child := n.getChild(i)
			if child != nil {
				if iterateNode(child, fn) {
					return true
				}
			}
		}
		return false
	default:
		panic("unknown node type")
	}
}
