package bptree

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"

	lru "github.com/hashicorp/golang-lru/v2"

	dbm "github.com/gnolang/gno/tm2/pkg/db"
)

// MutableTree is the working tree supporting Set, Get, Has, Remove,
// SaveVersion, LoadVersion, and Rollback.
//
// Concurrency: MutableTree is NOT safe for concurrent use. All methods —
// including read-only ones like Get, Has, Iterate, and proof
// generation — mutate tree-level caches (cachedRootHash, lazy childNodes
// on InnerNode) and therefore must be serialised externally. Proof
// generation in particular walks the current root without snapshotting
// miniTree state and will tear under concurrent Set/Remove. Callers that
// need concurrent reads should obtain an ImmutableTree via
// GetImmutable / Snapshot and read through that instead. See Findings
// #7 and #9.
type MutableTree struct {
	root      Node   // nil for empty tree
	lastSaved Node   // snapshot for rollback (set by SaveVersion)
	size      int64  // total key count in working tree
	version   int64  // last saved version

	ndb            *nodeDB // nil for in-memory only (no persistence)
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

	// Cached root-hash slices. Hash() / WorkingHash() each return a
	// []byte view of a [32]byte Hash value; the naive `h := x.Hash(); return h[:]`
	// pattern forces h to escape to the heap on every call. Holding the
	// slice here keeps the allocation to one per invalidation boundary
	// instead of one per call. nil means "recompute on next call".
	// See Finding #21.
	cachedRootHash  []byte // invalidated on t.root change (Set, Remove, LoadVersion, Rollback)
	cachedSavedHash []byte // invalidated on t.lastSaved change (SaveVersion, LoadVersion, Rollback)

	// fastNodes is a latest-view key→value LRU that short-circuits
	// Get/Has before the tree walk. Populated on Get hits; invalidated
	// per-key on Set/Remove and wholesale on Rollback / LoadVersion.
	// Covers only MutableTree reads — ImmutableTree snapshots walk the
	// tree unconditionally, since fast-node entries correspond to the
	// current root only. nil = disabled.
	fastNodes *lru.Cache[string, []byte]
}

// NewMutableTreeMem creates an in-memory MutableTree (no DB).
func NewMutableTreeMem() *MutableTree {
	t := &MutableTree{
		logger:         NewNopLogger(),
		memValues:      make(map[string][]byte),
		nextValueNonce: 1, // nonce=0 is reserved to avoid collision with the "missing" sentinel (Finding #6)
	}
	t.initFastNodeCache(DefaultFastNodeCacheSize)
	return t
}

// initFastNodeCache constructs the latest-view fast-node LRU according
// to `size`: > 0 uses that capacity, 0 uses DefaultFastNodeCacheSize,
// < 0 leaves the cache disabled. A panic at LRU construction indicates
// a bug (negative-size passed through the guard) — bail loudly.
func (t *MutableTree) initFastNodeCache(size int) {
	if size < 0 {
		return
	}
	if size == 0 {
		size = DefaultFastNodeCacheSize
	}
	c, err := lru.New[string, []byte](size)
	if err != nil {
		panic(fmt.Sprintf("bptree: fast-node cache init: %v", err))
	}
	t.fastNodes = c
}

// cowRoot ensures t.root is a mutable, working-version clone. It clones
// the root if it is shared with the lastSaved snapshot or came from the
// persistent layer (has a durable NodeKey). Once cloned, subsequent
// mutations within the same working version can skip the clone because
// t.root is neither == t.lastSaved nor has a NodeKey. See Finding #17.
func (t *MutableTree) cowRoot() {
	if t.root == nil {
		return
	}
	if t.root == t.lastSaved || t.root.GetNodeKey() != nil {
		t.root = cloneNode(t.root)
	}
}

// allocValueKey allocates a unique ValueKey for the current working session.
// The NodeKey struct is bypassed — only the serialized bytes are needed,
// so skipping the wrapper saves one heap allocation per Set. See Finding #21.
func (t *MutableTree) allocValueKey() []byte {
	vk := encodeNodeKeyBytes(t.WorkingVersion(), t.nextValueNonce)
	t.nextValueNonce++
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
	t := &MutableTree{
		ndb:            ndb,
		logger:         logger,
		initialVersion: opts.InitialVersion,
		// nonce=0 is reserved to avoid collision with the "missing" sentinel
		// in LeafNode.Serialize (12 zero bytes). See Finding #6.
		nextValueNonce: 1,
	}
	t.initFastNodeCache(opts.FastNodeCacheSize)
	return t
}

// Set inserts or updates a key-value pair. Returns true if the key
// already existed (update), false if it was a new insert.
//
// Finding #28: the value is persisted BEFORE any tree mutation. If the
// DB write fails, the in-memory tree is untouched and no leaf ends up
// referencing a ValueKey that was never written. The previous order
// (mutate-then-save) left the tree pointing at a dangling ValueKey on
// SaveValue errors; a subsequent Get returned "value not found" and
// SaveVersion would persist the inconsistent leaf.
func (t *MutableTree) Set(key, value []byte) (updated bool, err error) {
	if len(key) == 0 {
		return false, ErrEmptyKey
	}
	if value == nil {
		return false, fmt.Errorf("value must not be nil")
	}
	t.cachedRootHash = nil // working tree about to mutate — invalidate cache (Finding #21)

	// Allocate the ValueKey and persist the value FIRST. allocValueKey
	// also appends vk to sessionValues so Rollback can clean up if this
	// Set was part of a working session that later rolls back. On
	// SaveValue failure we return before mutating the tree; the dangling
	// sessionValues entry becomes a harmless no-op delete on Rollback
	// (DB drivers treat delete-of-absent as success).
	vk := t.allocValueKey()
	if t.ndb != nil {
		if err := t.ndb.SaveValue(value, vk); err != nil {
			return false, err
		}
	} else if t.memValues != nil {
		valCopy := make([]byte, len(value))
		copy(valCopy, value)
		t.memValues[string(vk)] = valCopy
	}

	valueHash := sha256.Sum256(value)

	if t.root == nil {
		leaf := &LeafNode{miniTree: NewMiniMerkle()}
		leaf.keys[0] = copyKey(key)
		leaf.valueHashes[0] = valueHash
		leaf.valueKeys[0] = vk
		leaf.numKeys = 1
		// Fresh leaf: slotHashes cache is cold, do a full build now.
		leaf.RebuildMiniMerkle()
		t.root = leaf
		t.size = 1
		return false, nil
	}

	t.cowRoot()
	newRoot, updated, oldValueKey := treeInsert(t.root, key, valueHash, vk)
	t.root = newRoot
	if !updated {
		t.size++
	}

	// Handle orphaned old valueKey on update
	if updated && oldValueKey != nil {
		t.orphanValueKey(oldValueKey)
	}

	// Refresh the fast-node cache with the fresh value. The stored
	// slice is the value passed to this Set; the caller retains
	// ownership, but the contract is read-only so sharing is safe.
	if t.fastNodes != nil {
		t.fastNodes.Add(string(key), value)
	}
	return updated, nil
}

// Get retrieves the value for a key. The latest-view fast-node cache
// short-circuits the tree walk on a hit; misses fall through to the
// regular lookup and populate the cache for next time. The cache
// lookup happens ahead of the root-nil check so that a cached value
// from the previous working session remains readable after a
// hypothetical root swap — paranoia against mis-ordered cache
// invalidation elsewhere.
func (t *MutableTree) Get(key []byte) ([]byte, error) {
	if t.fastNodes != nil {
		if v, ok := t.fastNodes.Get(string(key)); ok {
			return v, nil
		}
	}
	if t.root == nil {
		return nil, nil
	}
	_, _, vk, found := treeLookup(t.root, key)
	if !found {
		return nil, nil
	}
	val, err := t.resolveValue(vk)
	if err == nil && val != nil && t.fastNodes != nil {
		t.fastNodes.Add(string(key), val)
	}
	return val, err
}

// resolveValue resolves a valueKey to actual bytes via the tree's backing
// store. A tree with neither ndb nor memValues configured returns
// ErrNoValueResolver so callers can distinguish "misconfigured tree" from
// "lookup miss". A hit against memValues returns nil without error; a
// miss falls through to the not-found path and returns ErrKeyDoesNotExist
// rather than a bespoke formatted error, keeping the error surface
// consistent with Get/Has/resolve from ImmutableTree. See Findings #10
// and #11.
func (t *MutableTree) resolveValue(vk []byte) ([]byte, error) {
	if t.ndb != nil {
		return t.ndb.GetValue(vk)
	}
	if t.memValues == nil {
		return nil, ErrNoValueResolver
	}
	if val, ok := t.memValues[string(vk)]; ok {
		return val, nil
	}
	return nil, ErrKeyDoesNotExist
}

// orphanValueKey handles an orphaned valueKey from an overwrite or remove.
// Tier 1 (same working version): delete eagerly from DB.
// Tier 2 (prior version): defer to orphan list for prune-time deletion.
//
// Errors from the eager DB delete are logged rather than returned: the
// caller (Set / Remove) has already committed the tree mutation, so a
// failure here represents a space leak — not an inconsistency in the
// tree itself — and short-circuiting would leave the tree in a more
// confusing half-mutated state. See Finding #31.
func (t *MutableTree) orphanValueKey(vk []byte) {
	// Finding #23: a malformed or truncated valueKey would panic on the
	// slice bound. Bail out early with a log line so a corrupt inner ref
	// can't propagate into a slice-bounds runtime error.
	if len(vk) < 8 {
		t.logger.Error("bptree: orphanValueKey: short valueKey", "len", len(vk))
		return
	}
	// Decode version from the first 8 bytes of the valueKey
	vkVersion := int64(binary.BigEndian.Uint64(vk[:8]))
	if vkVersion == t.WorkingVersion() {
		// Tier 1: intra-version orphan — delete eagerly
		if t.ndb != nil {
			if err := t.ndb.DeleteValueDirect(vk); err != nil {
				// Finding #25 / #31: don't discard silently.
				t.logger.Error("bptree: DeleteValueDirect failed in orphanValueKey", "vk", fmt.Sprintf("%x", vk), "err", err)
			}
		} else if t.memValues != nil {
			delete(t.memValues, string(vk))
		}
	} else {
		// Tier 2: cross-version orphan — defer to prune
		t.versionOrphans = append(t.versionOrphans, vk)
	}
}

// Has returns true if the key exists in the tree. A hit in the
// fast-node cache answers without the tree walk; the absence of an
// entry is not a negative signal (miss-caching is deliberately not
// implemented) so a cache-miss falls through to the regular lookup.
func (t *MutableTree) Has(key []byte) (bool, error) {
	if t.fastNodes != nil {
		if _, ok := t.fastNodes.Get(string(key)); ok {
			return true, nil
		}
	}
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
	t.cachedRootHash = nil // working tree about to mutate — invalidate cache (Finding #21)
	t.cowRoot()
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
	if t.fastNodes != nil {
		t.fastNodes.Remove(string(key))
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
		// lastSaved just became root — saved-hash cache is now the
		// working-hash cache. Defer to WorkingHash() so the first
		// caller populates both entries in one step.
		t.cachedSavedHash = nil
		return t.WorkingHash(), version, nil
	}

	// If this version already exists, verify the hash matches.
	// This prevents accidentally overwriting a version with different data.
	if t.ndb.VersionExists(version) {
		existingNK, existingHash, err := t.ndb.GetRoot(version)
		if err != nil {
			return nil, 0, err
		}
		newHash := t.WorkingHash()
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
		t.cachedSavedHash = t.cachedRootHash
		return newHash, version, nil
	}

	t.ndb.ResetNonce()

	// On any failure between here and a successful Commit, the in-memory
	// tree may hold NodeKeys assigned by saveNode but never flushed to
	// disk (the batch is discarded). A straight return would leave the
	// caller with a tree whose dirty nodes look clean to a retry —
	// saveNode would skip them (`node.GetNodeKey() != nil` short-
	// circuits) and persist an incomplete version. Recover by rolling
	// back to the last saved snapshot and discarding the batch; the
	// caller sees the failure and can re-apply their mutations. See
	// Finding #36.
	failPartialSave := func(err error) ([]byte, int64, error) {
		t.ndb.discardBatch()
		t.Rollback()
		return nil, 0, err
	}

	// Assign NodeKeys and save all dirty nodes
	if t.root != nil {
		if err := t.saveNode(t.root, version); err != nil {
			return failPartialSave(err)
		}
	}

	// Save root reference
	rootHash := t.WorkingHash()
	if t.root != nil {
		if err := t.ndb.SaveRoot(version, t.root.GetNodeKey(), rootHash); err != nil {
			return failPartialSave(err)
		}
	} else {
		if err := t.ndb.SaveRoot(version, nil, rootHash); err != nil {
			return failPartialSave(err)
		}
	}

	// Persist cross-version orphan list (Tier 2)
	if err := t.ndb.SaveOrphans(version, t.versionOrphans); err != nil {
		return failPartialSave(err)
	}

	// Commit batch (nodes + root + orphan list, atomically)
	if err := t.ndb.Commit(); err != nil {
		return failPartialSave(err)
	}

	t.version = version
	t.lastSaved = t.root
	t.cachedSavedHash = t.cachedRootHash
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
//
// Finding #4: only slots whose in-memory child exists AND is dirty are
// visited. Slots with childNodes[i] == nil (never mutated since the
// parent was cloned or loaded) are skipped entirely — their serialized
// ref (children[i]) and hash (childHashes[i]) were preserved by the
// array-copy in Clone() and remain authoritative. Slots whose cached
// child has a NodeKey (clean — loaded from DB for a read but never
// mutated) are likewise skipped; re-saving a node that already has a
// durable NodeKey is a no-op but the old code still paid for the
// getChild() call, which triggers a DB read + ReadNode + 31-hash
// mini-merkle rebuild for any sibling of a mutated leaf.
func (t *MutableTree) saveNode(node Node, version int64) error {
	if node.GetNodeKey() != nil {
		return nil // already saved
	}

	// For inner nodes, save children first (bottom-up)
	if inner, ok := node.(*InnerNode); ok {
		for i := 0; i < inner.NumChildren(); i++ {
			child := inner.childNodes[i]
			if child == nil {
				// Unchanged subtree — serialized ref + hash already current.
				continue
			}
			if child.GetNodeKey() != nil {
				// Clean cached child (read-only touch); nothing to persist.
				continue
			}
			// Dirty child — save recursively and refresh the parent's
			// serialized ref + hash from the newly assigned NodeKey.
			if err := t.saveNode(child, version); err != nil {
				return err
			}
			inner.children[i] = child.GetNodeKey().GetKey()
			inner.childHashes[i] = child.Hash()
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
//
// ValueKey nonce recovery: `nextValueNonce` names slots in the
// (t.version+1) namespace. For a fresh / latest-loaded tree that
// namespace is unpopulated and starting at 1 is safe. But if the
// caller loads a non-latest version (e.g. to re-play history into a
// different v+1 state) and then Sets, a naive allocator would reuse
// nonces that already name persisted values under the real version
// v+1 — and because SaveValue writes directly to disk (outside the
// batch, so Get can observe intra-session writes), the collision
// silently overwrites a live value before SaveVersion's hash check
// can reject the stale save.
//
// Fix (Finding #1.1): scan the maximum nonce currently persisted for
// the working-version namespace and seed `nextValueNonce` past it.
// Subsequent Sets allocate fresh slots that cannot collide with any
// existing on-disk value. If SaveVersion later rejects the save
// (hash mismatch), Rollback cleans up these orphans via
// DeleteValueDirect.
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

	// Seed the value-nonce allocator past any persisted nonce in the
	// working-version namespace to prevent SaveValue from overwriting
	// a live on-disk value. See comment above and Finding #1.1.
	workingVersion := version + 1
	maxNonce, err := t.ndb.maxValueNonceForVersion(workingVersion)
	if err != nil {
		return 0, fmt.Errorf("scanning valueKey nonces for v%d: %w", workingVersion, err)
	}
	t.nextValueNonce = maxNonce + 1

	// Entries in the fast-node cache are for the previously-loaded
	// root; any key → value mapping we had is no longer authoritative.
	if t.fastNodes != nil {
		t.fastNodes.Purge()
	}

	if nkBytes == nil {
		// Empty tree at this version
		t.root = nil
		t.size = 0
		t.version = version
		t.lastSaved = nil
		t.cachedRootHash = nil
		t.cachedSavedHash = nil
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
	t.cachedRootHash = nil
	t.cachedSavedHash = nil
	return latestVersion, nil
}

// LoadVersionForOverwriting is not supported — it would leak values and nodes.
// Not called by gno.land, the SDK, or the store layer. Returns ErrUnsupported
// rather than panicking so callers that probe for IAVL compatibility can
// detect the gap without crashing the process. See Finding #12.
func (t *MutableTree) LoadVersionForOverwriting(_ int64) error {
	return fmt.Errorf("%w: LoadVersionForOverwriting; use PruneVersionsTo", ErrUnsupported)
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
// Not called by gno.land, the SDK, or the store layer. Returns ErrUnsupported
// rather than panicking so callers that probe for IAVL compatibility can
// detect the gap without crashing the process. See Finding #12.
func (t *MutableTree) DeleteVersionsFrom(_ int64) error {
	return fmt.Errorf("%w: DeleteVersionsFrom; use PruneVersionsTo", ErrUnsupported)
}

// Size returns the total number of key-value pairs.
func (t *MutableTree) Size() int64 { return t.size }

// IsEmpty returns true if the tree has no keys.
func (t *MutableTree) IsEmpty() bool { return t.root == nil }

// Hash returns the root hash of the last saved version.
// Returns SHA256("") for empty trees, matching IAVL behavior.
func (t *MutableTree) Hash() []byte {
	if t.cachedSavedHash != nil {
		return t.cachedSavedHash
	}
	if t.lastSaved == nil {
		t.cachedSavedHash = emptyHash()
		return t.cachedSavedHash
	}
	h := t.lastSaved.Hash()
	buf := make([]byte, HashSize)
	copy(buf, h[:])
	t.cachedSavedHash = buf
	return buf
}

// WorkingHash computes the hash of the current unsaved working tree.
// Returns SHA256("") for empty trees, matching IAVL behavior.
func (t *MutableTree) WorkingHash() []byte {
	if t.cachedRootHash != nil {
		return t.cachedRootHash
	}
	if t.root == nil {
		t.cachedRootHash = emptyHash()
		return t.cachedRootHash
	}
	h := t.root.Hash()
	buf := make([]byte, HashSize)
	copy(buf, h[:])
	t.cachedRootHash = buf
	return buf
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
//
// The root is cloned so subsequent mutations on the MutableTree cannot
// corrupt the snapshot's view. The child arrays are copied by value
// (fixed-size arrays), so subsequent COW descents that swap child
// pointers in the MutableTree's root do not affect the snapshot's root.
// See Finding #17.
func (t *MutableTree) Snapshot(version int64) *ImmutableTree {
	var snapRoot Node
	if t.root != nil {
		snapRoot = cloneNode(t.root)
	}
	imm := NewImmutableTree(snapRoot, version)
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

// Rollback discards all mutations since the last save.
// Eagerly-written values from the session are deleted from DB.
//
// Per-value delete errors are logged (Findings #25 / #31) and Rollback
// continues with the remaining entries. A single bad delete is a space
// leak, not a correctness issue, and returning early would leave the
// caller unable to restore the tree's in-memory state — which is the
// primary job of Rollback.
func (t *MutableTree) Rollback() {
	// Delete all eagerly-written values from this session
	if t.ndb != nil {
		for _, vk := range t.sessionValues {
			if err := t.ndb.DeleteValueDirect(vk); err != nil {
				t.logger.Error("bptree: DeleteValueDirect failed in Rollback", "vk", fmt.Sprintf("%x", vk), "err", err)
			}
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
	// root is now the saved root — the working-hash cache, if any, is
	// stale; reuse the saved-hash cache as-is (it is still valid).
	t.cachedRootHash = t.cachedSavedHash
	// Working-session Sets populated the fast-node cache with unsaved
	// values; those are no longer valid against the rolled-back root.
	// Preserving just the untouched entries would require per-session
	// write tracking we don't otherwise keep, so purge wholesale —
	// Rollback is a rare, user-initiated path.
	if t.fastNodes != nil {
		t.fastNodes.Purge()
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

// iterateNodeResolved walks node in key order, yielding each (key, valueKey)
// pair to fn. Returning true from fn stops iteration; iterateNodeResolved
// returns true in that case and false when the walk completes normally. The
// previous value-hash sibling (iterateNode) was removed alongside the
// legacy no-resolver Iterate fallback that silently handed callers hashes
// where values were expected. See Findings #11 and #19.
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
