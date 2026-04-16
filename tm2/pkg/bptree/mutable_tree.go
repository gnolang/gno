package bptree

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
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

	// In-memory value store (no ndb). Keyed by string(valueKey).
	memValues map[string][]byte

	// Value nonce counter for allocating unique ValueKeys.
	nextValueNonce uint32

	// Tier 1: all ValueKeys allocated in the current working session.
	// On Rollback, these are deleted from DB. On SaveVersion, cleared.
	sessionValues [][]byte

	// Tier 2: cross-version orphaned ValueKeys (from prior committed versions).
	// Persisted to DB at SaveVersion, consumed during PruneVersionsTo.
	versionOrphans [][]byte
}

// NewMutableTreeMem creates an in-memory MutableTree (no DB).
func NewMutableTreeMem() *MutableTree {
	return &MutableTree{
		logger:         NewNopLogger(),
		memValues:      make(map[string][]byte),
		nextValueNonce: 1, // nonce=0 is reserved to avoid collision with the "missing" sentinel (Finding #6)
	}
}

// allocValueKey allocates a unique ValueKey for the current working session.
func (t *MutableTree) allocValueKey() []byte {
	nk := &NodeKey{Version: t.WorkingVersion(), Nonce: t.nextValueNonce}
	t.nextValueNonce++
	vk := nk.GetKey()
	t.sessionValues = append(t.sessionValues, vk)
	return vk
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
		// nonce=0 is reserved to avoid collision with the "missing" sentinel
		// in LeafNode.Serialize (12 zero bytes). See Finding #6.
		nextValueNonce: 1,
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
		vk := t.allocValueKey()
		leaf.valueKeys[0] = vk
		leaf.numKeys = 1
		leaf.RebuildMiniMerkle()
		t.root = leaf
		t.size = 1

		// Save value out-of-line
		if t.ndb != nil {
			if err := t.ndb.SaveValue(value, vk); err != nil {
				return false, err
			}
		} else if t.memValues != nil {
			valCopy := make([]byte, len(value))
			copy(valCopy, value)
			t.memValues[string(vk)] = valCopy
		}
		return false, nil
	}

	valueHash := sha256.Sum256(value)
	vk := t.allocValueKey()
	newRoot, updated, oldValueKey := treeInsert(t.root, key, valueHash, vk)
	t.root = newRoot
	if !updated {
		t.size++
	}

	// Handle orphaned old valueKey on update
	if updated && oldValueKey != nil {
		t.orphanValueKey(oldValueKey)
	}

	// Save value out-of-line
	if t.ndb != nil {
		if err := t.ndb.SaveValue(value, vk); err != nil {
			return updated, err
		}
	} else if t.memValues != nil {
		valCopy := make([]byte, len(value))
		copy(valCopy, value)
		t.memValues[string(vk)] = valCopy
	}
	return updated, nil
}

// Get retrieves the value for a key.
func (t *MutableTree) Get(key []byte) ([]byte, error) {
	if t.root == nil {
		return nil, nil
	}
	_, _, vk, found := treeLookup(t.root, key)
	if !found {
		return nil, nil
	}
	return t.resolveValue(vk)
}

// resolveValue resolves a valueKey to actual bytes.
func (t *MutableTree) resolveValue(vk []byte) ([]byte, error) {
	if t.ndb != nil {
		return t.ndb.GetValue(vk)
	}
	if t.memValues != nil {
		if val, ok := t.memValues[string(vk)]; ok {
			return val, nil
		}
	}
	return nil, fmt.Errorf("value not found for key %x", vk)
}

// orphanValueKey handles an orphaned valueKey from an overwrite or remove.
// Tier 1 (same working version): delete eagerly from DB.
// Tier 2 (prior version): defer to orphan list for prune-time deletion.
func (t *MutableTree) orphanValueKey(vk []byte) {
	// Decode version from the first 8 bytes of the valueKey
	vkVersion := int64(binary.BigEndian.Uint64(vk[:8]))
	if vkVersion == t.WorkingVersion() {
		// Tier 1: intra-version orphan — delete eagerly
		if t.ndb != nil {
			t.ndb.DeleteValueDirect(vk)
		} else if t.memValues != nil {
			delete(t.memValues, string(vk))
		}
	} else {
		// Tier 2: cross-version orphan — defer to prune
		t.versionOrphans = append(t.versionOrphans, vk)
	}
}

// Has returns true if the key exists in the tree.
func (t *MutableTree) Has(key []byte) (bool, error) {
	if t.root == nil {
		return false, nil
	}
	_, _, _, found := treeLookup(t.root, key)
	return found, nil
}

// Remove removes a key from the tree. Returns the old value and
// whether the key was found.
func (t *MutableTree) Remove(key []byte) ([]byte, bool, error) {
	if t.root == nil {
		return nil, false, nil
	}
	newRoot, _, oldVK, found := treeRemove(t.root, key)
	if !found {
		return nil, false, nil
	}
	t.root = newRoot
	t.size--

	// Resolve old value BEFORE orphaning (Tier 1 may delete it from DB)
	var val []byte
	if oldVK != nil {
		val, _ = t.resolveValue(oldVK)
		t.orphanValueKey(oldVK)
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
		t.sessionValues = t.sessionValues[:0]
		t.versionOrphans = t.versionOrphans[:0]
		t.nextValueNonce = 1
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
		// Legacy empty-tree blobs (zero-length stored value) deserialize
		// as (nil, nil) from GetRoot. Normalize to the canonical empty
		// hash so an idempotent re-save of an empty tree is recognised
		// as equivalent instead of producing a false "hash mismatch".
		// See Finding #26.
		existingEmpty := existingNK == nil
		if existingEmpty && existingHash == nil {
			existingHash = emptyHash()
		}
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

	// Persist cross-version orphan list (Tier 2)
	if err := t.ndb.SaveOrphans(version, t.versionOrphans); err != nil {
		return nil, 0, err
	}

	// Commit batch (nodes + root + orphan list, atomically)
	if err := t.ndb.Commit(); err != nil {
		return nil, 0, err
	}

	t.version = version
	t.lastSaved = t.root
	t.ndb.setLatestVersion(version)
	if t.ndb.getFirstVersion() == 0 {
		t.ndb.setFirstVersion(version)
	}

	// Clear session state
	t.sessionValues = t.sessionValues[:0]
	t.versionOrphans = t.versionOrphans[:0]
	t.nextValueNonce = 1

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
		// Version <= 0 means "load latest", matching IAVL behavior.
		return t.Load()
	}

	// Discover the DB's latest version before loading, to return it
	// (matching IAVL behavior which returns latestVersion, not targetVersion).
	if err := t.ndb.discoverVersions(); err != nil {
		return 0, err
	}
	latestVersion := t.ndb.getLatestVersion()

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
		return latestVersion, nil
	}

	root, err := t.loadNode(nkBytes)
	if err != nil {
		return 0, fmt.Errorf("loading root: %w", err)
	}

	t.root = root
	t.size = nodeSize(root)
	t.version = version
	t.lastSaved = root
	return latestVersion, nil
}

// LoadVersionForOverwriting is not supported — it would leak values and nodes.
// Not called by gno.land, the SDK, or the store layer.
func (t *MutableTree) LoadVersionForOverwriting(_ int64) error {
	panic("LoadVersionForOverwriting is not supported; use PruneVersionsTo")
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
				imm.valueResolver = func(vk []byte) ([]byte, error) {
					val, ok := t.memValues[string(vk)]
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

	// Register as a reader FIRST so that a concurrent PruneVersionsTo
	// cannot delete the root record or any node entries for this version
	// between our GetRoot call and the end of root loading. Finding #15
	// closed the check-vs-delete race from the prune side; this closes
	// the symmetric reader-side TOCTOU. See Findings #30 and #40.
	t.ndb.incrVersionReaders(version)

	nkBytes, _, err := t.ndb.GetRoot(version)
	if err != nil {
		t.ndb.decrVersionReaders(version)
		return nil, err
	}
	if nkBytes == nil {
		// Empty saved version — registration is still held so the caller
		// gets a consistent view until Close().
		imm := NewImmutableTree(nil, version)
		imm.ndb = t.ndb
		return imm, nil
	}

	root, err := t.loadNode(nkBytes)
	if err != nil {
		t.ndb.decrVersionReaders(version)
		return nil, err
	}
	imm := NewImmutableTree(root, version)
	imm.ndb = t.ndb
	imm.valueResolver = func(vk []byte) ([]byte, error) {
		return t.ndb.GetValue(vk)
	}
	return imm, nil
}

// GetVersioned returns the value for a key at a specific version.
func (t *MutableTree) GetVersioned(key []byte, version int64) ([]byte, error) {
	imm, err := t.GetImmutable(version)
	if err != nil {
		return nil, nil // match IAVL behavior: silent nil for missing version
	}
	defer imm.Close()
	return imm.Get(key)
}

// DeleteVersionsTo deletes versions from first to toVersion (inclusive),
// including orphaned nodes via dual-tree-walk pruning.
func (t *MutableTree) DeleteVersionsTo(toVersion int64) error {
	return t.PruneVersionsTo(toVersion)
}

// DeleteVersionsFrom is not supported — it would leak values and nodes.
// Not called by gno.land, the SDK, or the store layer.
func (t *MutableTree) DeleteVersionsFrom(_ int64) error {
	panic("DeleteVersionsFrom is not supported; use PruneVersionsTo")
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

// Snapshot creates an ImmutableTree snapshot of the current working tree
// with a properly wired value resolver. For tests and lightweight snapshots.
func (t *MutableTree) Snapshot(version int64) *ImmutableTree {
	imm := NewImmutableTree(t.root, version)
	if t.ndb != nil {
		imm.valueResolver = func(vk []byte) ([]byte, error) {
			return t.ndb.GetValue(vk)
		}
	} else if t.memValues != nil {
		imm.valueResolver = func(vk []byte) ([]byte, error) {
			val, ok := t.memValues[string(vk)]
			if !ok {
				return nil, fmt.Errorf("value not found in memValues")
			}
			return val, nil
		}
	}
	return imm
}

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
// Eagerly-written values from the session are deleted from DB.
func (t *MutableTree) Rollback() {
	// Delete all eagerly-written values from this session
	if t.ndb != nil {
		for _, vk := range t.sessionValues {
			t.ndb.DeleteValueDirect(vk)
		}
	} else if t.memValues != nil {
		for _, vk := range t.sessionValues {
			delete(t.memValues, string(vk))
		}
	}
	t.sessionValues = t.sessionValues[:0]
	t.versionOrphans = t.versionOrphans[:0]
	t.nextValueNonce = 1

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

// GetValueByKey resolves a valueKey to the raw value bytes.
func (t *MutableTree) GetValueByKey(vk []byte) ([]byte, error) {
	return t.resolveValue(vk)
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
	key, _, vk := treeGetByIndex(t.root, index)
	val, err := t.resolveValue(vk)
	return key, val, err
}

// GetWithIndex returns the index, value, and whether the key was found.
func (t *MutableTree) GetWithIndex(key []byte) (int64, []byte, error) {
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
// Values are resolved from the value store (DB or memValues).
// If value resolution fails, iteration stops and the error is returned.
func (t *MutableTree) Iterate(fn func(key []byte, value []byte) bool) (bool, error) {
	if t.root == nil {
		return false, nil
	}
	var resolveErr error
	stopped := iterateNodeResolved(t.root, func(key, vk []byte) bool {
		val, err := t.resolveValue(vk)
		if err != nil {
			resolveErr = err
			return true // stop
		}
		return fn(key, val)
	})
	return stopped, resolveErr
}

// --- helpers ---

func treeLookup(node Node, key []byte) (*LeafNode, Hash, []byte, bool) {
	for {
		switch n := node.(type) {
		case *LeafNode:
			pos, found := searchLeaf(n, key)
			if !found {
				return n, Hash{}, nil, false
			}
			return n, n.valueHashes[pos], n.valueKeys[pos], true
		case *InnerNode:
			idx := searchInner(n, key)
			child := n.getChild(idx)
			if child == nil {
				return nil, Hash{}, nil, false
			}
			node = child
		default:
			panic("unknown node type")
		}
	}
}

func treeGetByIndex(node Node, index int64) ([]byte, Hash, []byte) {
	switch n := node.(type) {
	case *LeafNode:
		return n.keys[index], n.valueHashes[index], n.valueKeys[index]
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

func treeGetWithIndex(node Node, key []byte) (int64, Hash, []byte, bool) {
	switch n := node.(type) {
	case *LeafNode:
		pos, found := searchLeaf(n, key)
		if !found {
			return int64(pos), Hash{}, nil, false
		}
		return int64(pos), n.valueHashes[pos], n.valueKeys[pos], true
	case *InnerNode:
		childIdx := searchInner(n, key)
		offset := int64(0)
		for i := 0; i < childIdx; i++ {
			offset += n.childSizes[i]
		}
		child := n.getChild(childIdx)
		idx, vh, vk, found := treeGetWithIndex(child, key)
		return offset + idx, vh, vk, found
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

// iterateNodeResolved is like iterateNode but passes valueKeys to the callback
// instead of valueHashes, enabling value resolution via ValueKey.
func iterateNodeResolved(node Node, fn func(key, vk []byte) bool) bool {
	switch n := node.(type) {
	case *LeafNode:
		for i := 0; i < int(n.numKeys); i++ {
			if fn(n.keys[i], n.valueKeys[i]) {
				return true
			}
		}
		return false
	case *InnerNode:
		for i := 0; i < n.NumChildren(); i++ {
			child := n.getChild(i)
			if child != nil {
				if iterateNodeResolved(child, fn) {
					return true
				}
			}
		}
		return false
	default:
		panic("unknown node type")
	}
}
