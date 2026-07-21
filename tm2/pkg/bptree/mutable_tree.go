package bptree

import (
	"bytes"
	"crypto/sha256"
	"fmt"

	dbm "github.com/gnolang/gno/tm2/pkg/db"
)

// MutableTree is the working (writable) B+ tree: Set, Get, Has, Remove,
// SaveVersion, LoadVersion, Rollback, pruning, proofs, and iteration.
//
// # Concurrency
//
// A MutableTree is SINGLE-GOROUTINE. Its mutators (Set, Remove, SaveVersion,
// LoadVersion, Load, Rollback, DeleteVersionsTo, PruneVersionsTo, Import,
// SetInitialVersion) and its working-tree reads (Get, Has, Hash, WorkingHash,
// Version, WorkingVersion, Size, IsEmpty, Height, Iterate, Iterator,
// GetByIndex, GetWithIndex, GetValueByKey, GetMembershipProof,
// GetNonMembershipProof, Snapshot) read and write the working-tree fields
// (root, lastSaved, version, size, nextValueNonce, versionOrphans) WITHOUT
// locking, so none of them may be called concurrently with each other or with
// a mutator.
//
// For concurrent reads at a COMMITTED version, call GetImmutable(version) — it
// is safe to call concurrently with the writer (it reads only the
// internally-synchronized nodeDB, never the working-tree fields) — and read the
// returned ImmutableTree, which is safe for concurrent reads against an active
// writer; Close it when done (it holds a version-reader reservation that blocks
// pruning of that version until released). GetVersioned, GetCommittedValueByKey,
// VersionExists, and AvailableVersions are likewise safe to call concurrently
// with the writer.
//
// The gno ABCI path satisfies this contract by serializing all store access
// through the connection mutex.
type MutableTree struct {
	root      Node  // nil for empty tree
	lastSaved Node  // committed root: rollback target, and the clean-session witness gating fast-index reads (see fastReadable)
	size      int64 // total key count in working tree
	version   int64 // last saved version

	ndb            *nodeDB
	initialVersion uint64
	logger         Logger

	// Value nonce counter for allocating unique ValueKeys.
	nextValueNonce uint32

	// Tier 2: cross-version orphaned ValueKeys (from prior committed versions).
	// Persisted to DB at SaveVersion, consumed during PruneVersionsTo.
	versionOrphans [][]byte

	// poisoned records the first error that left the session inconsistent
	// (a mutation published before its DB staging failed, or a SaveVersion
	// whose deferred DiscardBatch destroyed staged values the working tree
	// still references). While set, Set/Remove/SaveVersion refuse with
	// ErrSessionPoisoned — committing the session would silently persist
	// dangling references or an unloadable version. Rollback fully restores
	// the session and clears it; a successful LoadVersion (full session
	// replacement) clears it too.
	poisoned error
}

// allocValueKey allocates a unique ValueKey for the current working session.
func (t *MutableTree) allocValueKey() []byte {
	nk := &NodeKey{Version: t.WorkingVersion(), Nonce: t.nextValueNonce}
	t.nextValueNonce++
	return nk.GetKey()
}

// resetSession clears the state accumulated during the current working session
// (the value-nonce counter and the cross-version orphan list). Called whenever a
// session is committed, rolled back, or abandoned by loading a different
// version, so nothing carries into the next working view.
func (t *MutableTree) resetSession() {
	t.versionOrphans = t.versionOrphans[:0]
	t.nextValueNonce = 0
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
	if t.poisoned != nil {
		return false, fmt.Errorf("%w: %w", ErrSessionPoisoned, t.poisoned)
	}
	if len(key) == 0 {
		return false, ErrEmptyKey
	}
	if len(key) > MaxKeyLen {
		return false, fmt.Errorf("%w: %d > %d", ErrKeyTooLong, len(key), MaxKeyLen)
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

		// Save value out-of-line.
		if err := t.ndb.SaveValue(value, vk); err != nil {
			// The root/size were already published referencing this value —
			// a later SaveVersion would commit a dangling valueKey.
			t.poisoned = err
			return false, err
		}
		// Maintain the fast index (no-op when disabled). Staged after SaveValue
		// so a poisoned value never leaves a live index entry behind.
		if err := t.ndb.setFastIndex(key, vk, value); err != nil {
			t.poisoned = err
			return false, err
		}
		return false, nil
	}

	valueHash := sha256.Sum256(value)
	vk := t.allocValueKey()
	newRoot, updated, oldValueKey, err := treeInsert(t.root, key, valueHash, vk)
	if err != nil {
		// Mid-descent load failure: every mutation was on unpublished clones,
		// so the tree is untouched — do NOT publish the root.
		return false, err
	}
	t.root = newRoot
	if !updated {
		t.size++
	}

	// Handle orphaned old valueKey on update
	if updated && oldValueKey != nil {
		if err := t.orphanValueKey(oldValueKey); err != nil {
			// Post-publication failure: the tree references the new value
			// while the old one's staged delete failed.
			t.poisoned = err
			return updated, err
		}
	}

	// Save value out-of-line.
	if err := t.ndb.SaveValue(value, vk); err != nil {
		// The tree already references this valueKey — a later SaveVersion
		// would commit it dangling.
		t.poisoned = err
		return updated, err
	}
	// Maintain the fast index (no-op when disabled). Overwrites any prior entry
	// for key with the new value; the old vk is already orphaned above.
	if err := t.ndb.setFastIndex(key, vk, value); err != nil {
		t.poisoned = err
		return updated, err
	}
	return updated, nil
}

// fastReadable reports whether Get may serve from the committed fast index:
// the feature is on and the working root IS the committed root, i.e. the
// session has no staged mutations. Pointer identity is exact because every
// published mutation COW-clones the root (treeInsert/treeRemove clone at
// entry); this is the pointer-identity component of PruneVersionsTo's
// clean-session check (prune.go). No staged-batch/pendingVals checks are
// needed: fastGet reads committed DB state only, and a clean root means no
// staged write can affect a read's answer.
func (t *MutableTree) fastReadable() bool {
	return t.ndb.opts.FastIndex && t.root == t.lastSaved
}

// Get retrieves the value for a key.
func (t *MutableTree) Get(key []byte) ([]byte, error) {
	if t.root == nil {
		return nil, nil
	}
	// Advisory fast path — trust rule in fast_index.go (a clean working tree
	// IS the committed snapshot at t.version). Keep this after the nil-root
	// return: it avoids probing on an empty tree, and a committed-empty tree
	// with a staged Set+Remove round-trip has root == lastSaved == nil
	// (defense in depth — in-contract, a clean committed-empty tree has an
	// empty index anyway).
	if t.fastReadable() {
		if val, ok := t.ndb.fastGet(key, t.version); ok {
			return val, nil
		}
	}
	_, _, vk, found, err := treeLookup(t.root, key)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	return t.resolveValue(vk)
}

// resolveValue resolves a valueKey to actual bytes.
func (t *MutableTree) resolveValue(vk []byte) ([]byte, error) {
	return t.ndb.GetValue(vk)
}

// orphanValueKey handles an orphaned valueKey from an overwrite or remove.
// Tier 1 (same working version): drop the staged write before it is committed.
// Tier 2 (prior version): defer to orphan list for prune-time deletion.
func (t *MutableTree) orphanValueKey(vk []byte) error {
	// Decode version from the first 8 bytes of the valueKey
	if vkVersion(vk) == t.WorkingVersion() {
		// Tier 1: intra-version orphan — drop the staged value
		return t.ndb.DeleteValueDirect(vk)
	}
	// Tier 2: cross-version orphan — defer to prune
	t.versionOrphans = append(t.versionOrphans, vk)
	return nil
}

// Has returns true if the key exists in the tree.
func (t *MutableTree) Has(key []byte) (bool, error) {
	if t.root == nil {
		return false, nil
	}
	return treeHas(t.root, key)
}

// Remove removes a key from the tree. Returns the old value and
// whether the key was found.
func (t *MutableTree) Remove(key []byte) ([]byte, bool, error) {
	if t.poisoned != nil {
		return nil, false, fmt.Errorf("%w: %w", ErrSessionPoisoned, t.poisoned)
	}
	if t.root == nil {
		return nil, false, nil
	}
	newRoot, _, oldVK, found, err := treeRemove(t.root, key)
	if err != nil {
		// Mid-descent load failure: every mutation was on unpublished clones,
		// so the tree is untouched — do NOT publish the root.
		return nil, false, err
	}
	if !found {
		return nil, false, nil
	}
	t.root = newRoot
	t.size--

	// Resolve old value BEFORE orphaning (Tier 1 may delete it from DB)
	var val []byte
	if oldVK != nil {
		val, _ = t.resolveValue(oldVK)
		if err := t.orphanValueKey(oldVK); err != nil {
			// Post-publication failure: the removal is published but the
			// displaced value's staged delete failed — committing would leak
			// the record permanently (it appears in no orphan list).
			t.poisoned = err
			return val, true, err
		}
	}
	// Drop the fast-index entry (no-op when disabled). Load-bearing: a leftover
	// entry would be wrongly trusted for a snapshot at vkVersion(vk_old) or later.
	if err := t.ndb.deleteFastIndex(key); err != nil {
		t.poisoned = err
		return val, true, err
	}
	return val, true, nil
}

// SaveVersion persists the current tree state as a new version.
// Returns (rootHash, version, error).
func (t *MutableTree) SaveVersion() (rootHash []byte, savedVersion int64, err error) {
	if t.poisoned != nil {
		return nil, 0, fmt.Errorf("%w: %w", ErrSessionPoisoned, t.poisoned)
	}

	version := t.WorkingVersion()

	// Values and nodes staged by Set/Remove since the last Commit live in the
	// batch + pendingVals; they become durable only if Commit succeeds below.
	// On every non-committing exit (error OR idempotent no-op) we MUST discard
	// them — a staged write left in the batch would be flushed by the next
	// Commit (a later SaveVersion or PruneVersionsTo), silently overwriting an
	// already-persisted version's value namespace (the LoadVersion(non-latest)
	// +Set hazard).
	//
	// That discard is also why EVERY error exit poisons the session: it
	// destroys the staged values while the working tree still references
	// them, so a retry (even after a transient fault) would commit dangling
	// valueKeys — or, after the save phase started, an unloadable version
	// (saveNode skips "already saved" nodes whose records were discarded).
	// Rollback restores the session and clears the poison. NOTE: the
	// poison-setter reads the named err — keep all returns explicit.
	committed := false
	defer func() {
		if !committed {
			t.ndb.DiscardBatch()
		}
		if err != nil {
			t.poisoned = err
		}
	}()

	// If this version already exists, verify the hash matches.
	// This prevents accidentally overwriting a version with different data.
	// Use the error-propagating variant: a transient DB error must NOT be read
	// as "does not exist" (which would overwrite the existing version with
	// unverified new data). The deferred DiscardBatch drops staged writes.
	exists, err := t.ndb.versionExistsE(version)
	if err != nil {
		return nil, 0, fmt.Errorf("checking version %d existence: %w", version, err)
	}
	if exists {
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
		// Same hash — idempotent save. Adopt the PERSISTED version's tree:
		// a replayed in-memory graph can reference different records than the
		// persisted version does (equal hash does not mean equal NodeKeys or
		// valueKeys), and carrying it forward would fork the record lineage —
		// later versions built from it would reference records the persisted
		// chain dropped (prune then over-deletes them), reference valueKeys
		// the deferred DiscardBatch is about to discard, or orphan-list
		// valueKeys the persisted version still needs. Loading the persisted
		// root makes the working tree exactly the persisted version.
		var root Node
		if existingNK != nil {
			root, err = t.loadNode(existingNK)
			if err != nil {
				return nil, 0, fmt.Errorf("adopting persisted version %d: %w", version, err)
			}
		}
		t.root = root
		t.lastSaved = root
		if root != nil {
			t.size = nodeSize(root)
		} else {
			t.size = 0
		}
		t.version = version
		// Drops the replayed session's value-nonce and orphan list (the
		// latter could name valueKeys the persisted version still references);
		// the deferred DiscardBatch drops its staged values and deletes.
		t.resetSession()
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

	// Stamp the fast index complete through this version (no-op when disabled),
	// so Load skips the rebuild for an eagerly-maintained index. Rides the same
	// atomic Commit as the index entries it describes.
	if err := t.ndb.setFastIndexVersion(version); err != nil {
		return nil, 0, err
	}

	// Commit batch (nodes + root + orphan list, atomically)
	if err := t.ndb.Commit(); err != nil {
		return nil, 0, err
	}
	committed = true

	t.version = version
	t.lastSaved = t.root
	t.ndb.setLatestVersion(version)
	if t.ndb.getFirstVersion() == 0 {
		t.ndb.setFirstVersion(version)
	}

	t.resetSession()

	return rootHash, version, nil
}

// saveNode recursively assigns NodeKeys and saves dirty nodes.
//
// Only in-memory child references (inner.childNodes[i]) are traversed: an
// unloaded child (childNodes[i] == nil) cannot have been mutated in this
// session, so its serialized reference (inner.children[i]) and its cached
// hash (inner.childHashes[i]) are still correct. Calling getChild here —
// which would force-load every sibling from DB just to early-return on
// "already has a NodeKey" — would cause O(B) wasted reads per COW'd inner.
// For a path-length-H insert with branching B, that's H*(B-1) unnecessary
// DB reads per SaveVersion, mostly cache-missing at blockchain scale.
func (t *MutableTree) saveNode(node Node, version int64) error {
	if node.GetNodeKey() != nil {
		return nil // already saved
	}

	// For inner nodes, save dirty children first (bottom-up).
	if inner, ok := node.(*InnerNode); ok {
		for i := 0; i < inner.NumChildren(); i++ {
			child := inner.childNodes[i]
			if child == nil {
				// Unloaded and therefore unchanged: children[i] (NodeKey ref)
				// and childHashes[i] are authoritative from the prior load.
				continue
			}
			if err := t.saveNode(child, version); err != nil {
				return err
			}
			// Update child reference and hash after save. For clean children
			// whose saveNode call early-returned, these assignments are
			// redundant but harmless (same NodeKey, same hash).
			inner.children[i] = child.GetNodeKey().GetKey()
			inner.childHashes[i] = child.Hash()
			// Now that children[i]/childHashes[i] are durable, drop the in-memory
			// child pointer: it reloads on demand via getChild, so the working
			// tree stays bounded by the cache instead of pinning every saved node.
			inner.childNodes[i] = nil
		}
		inner.RebuildMiniMerkle()
	}

	// Rebuild leaf mini merkle (may already be done, but ensure correctness).
	if leaf, ok := node.(*LeafNode); ok {
		leaf.RebuildMiniMerkle()
	}

	// Assign NodeKey.
	nk := t.ndb.NextNodeKey(version)
	node.SetNodeKey(nk)

	return t.ndb.SaveNode(node)
}

// Load loads the latest version from the DB.
func (t *MutableTree) Load() (int64, error) {
	if err := t.ndb.discoverVersions(); err != nil {
		return 0, err
	}
	latest := t.ndb.getLatestVersion()
	if latest == 0 {
		return 0, nil
	}
	v, err := t.LoadVersion(latest)
	if err != nil {
		return v, err
	}
	// Build the fast index from the loaded latest root if it is absent/stale
	// (e.g. enabling the feature on an existing DB, or post-import). No-op when
	// disabled or already current. A rebuild error is returned (surfacing an
	// index-write failure, or a value-read failure since the rebuild re-reads
	// every live value, at startup); the loaded tree itself is unaffected and a
	// retry Load re-attempts the rebuild.
	if err := t.ensureFastIndex(); err != nil {
		return v, err
	}
	return v, nil
}

// LoadVersion loads a specific version from the DB.
//
// All fallible reads happen BEFORE the working session is discarded: on any
// error the previous session survives fully intact (a failed load must not
// wipe staged values the working tree still references — committing that
// would persist dangling valueKeys).
func (t *MutableTree) LoadVersion(version int64) (int64, error) {
	if version <= 0 {
		// Version <= 0 means "load latest", matching IAVL behavior.
		return t.Load()
	}

	// Discover the DB's latest version before loading, to return it
	// (matching IAVL behavior which returns latestVersion, not targetVersion).
	// Refreshing first/latest on a path that later errors is harmless: they
	// are re-derived counters, not session state.
	if err := t.ndb.discoverVersions(); err != nil {
		return 0, err
	}
	latestVersion := t.ndb.getLatestVersion()

	nkBytes, _, err := t.ndb.GetRoot(version)
	if err != nil {
		return 0, err
	}

	if nkBytes == nil {
		// Empty tree at this version: replace the session.
		t.ndb.DiscardBatch()
		t.resetSession()
		t.root = nil
		t.size = 0
		t.version = version
		t.lastSaved = nil
		t.poisoned = nil // full session replacement
		return latestVersion, nil
	}

	root, err := t.loadNode(nkBytes)
	if err != nil {
		return 0, fmt.Errorf("loading root: %w", err)
	}

	// Reads succeeded — replace the session: drop any values/nodes staged
	// since the last Commit (they belong to the abandoned working session)
	// and reset session counters, so stale staged writes can't later flush
	// into the wrong version's value namespace.
	t.ndb.DiscardBatch()
	t.resetSession()
	t.root = root
	t.size = nodeSize(root)
	t.version = version
	t.lastSaved = root
	t.poisoned = nil // full session replacement
	return latestVersion, nil
}

// LoadVersionForOverwriting is not supported — it would leak values and nodes.
// Not called by gno.land, the SDK, or the store layer.
func (t *MutableTree) LoadVersionForOverwriting(_ int64) error {
	return fmt.Errorf("%w: LoadVersionForOverwriting (use PruneVersionsTo)", ErrUnsupported)
}

// loadNode loads a node from the DB. Children are loaded lazily via
// getChild (which uses the ndb reference set during deserialization).
func (t *MutableTree) loadNode(nkBytes []byte) (Node, error) {
	return t.ndb.GetNode(nkBytes)
}

// newImmutable builds an ImmutableTree for root/version with this tree's value
// resolver wired. Centralizes the resolver wiring shared by GetImmutable,
// Snapshot, and immutableForProof.
//
// committed selects the value-resolution policy:
//   - true  (GetImmutable / immutableForProof): the root is a DURABLE committed
//     version, read concurrently with the writer → resolve DB-only
//     (getCommittedValue), never the writer's pendingVals buffer, so reads can't
//     race SaveValue. A committed version only resolves valueKeys <
//     workingVersion, so its pendingVals lookups would always miss anyway.
//   - false (Snapshot): the root is the LIVE working tree, whose latest Sets
//     live only in pendingVals (not yet in the DB). Snapshot is a
//     single-writer-only convenience (no concurrent writer by contract), so it
//     resolves through GetValue for read-your-writes.
func (t *MutableTree) newImmutable(root Node, version int64, committed bool) *ImmutableTree {
	imm := NewImmutableTree(root, version)
	// Carry ndb so iterators created from this snapshot register as version
	// readers (incrVersionReaders), blocking a concurrent prune of `version`
	// until they Close.
	imm.ndb = t.ndb
	if committed {
		imm.valueResolver = t.ndb.getCommittedValue
		// The fast index reflects committed state, so only committed snapshots
		// may consult it here; a read-your-writes (Snapshot) tree must not.
		imm.fast = t.ndb.opts.FastIndex
	} else {
		imm.valueResolver = t.ndb.GetValue
	}
	return imm
}

// GetImmutable returns a committed read-only snapshot at version, REGISTERED as
// a version reader: a concurrent PruneVersionsTo(version) is blocked until the
// snapshot is Closed. Callers MUST Close it (else that version can never prune).
func (t *MutableTree) GetImmutable(version int64) (*ImmutableTree, error) {
	return t.getImmutable(version, true)
}

// GetImmutableUnregistered returns a committed read-only snapshot WITHOUT
// registering as a version reader. For long-lived snapshots that have no Close
// hook (e.g. the store's immutable LoadVersion view) — registering them would
// pin the version against pruning forever. Such a snapshot is not protected
// against a concurrent prune of its version. Queries are NOT serialized against
// prune: the query connection gets its own mutex in tm2/pkg/bft/proxy/client.go,
// separate from the one consensus and mempool share, so the two run
// concurrently by design. What keeps this safe is that a query does not reach
// the consensus tree at all — rootmulti builds a separate store over a snapshot,
// with its own version-reader map — and that this snapshot registers no reader,
// so it cannot block prune either.
func (t *MutableTree) GetImmutableUnregistered(version int64) (*ImmutableTree, error) {
	return t.getImmutable(version, false)
}

// getImmutable builds a committed snapshot at version. When register (and
// version > 0) it increments the version-reader count BEFORE loading the root —
// closing the reader-side check-vs-delete TOCTOU — and marks the snapshot so
// Close decrements it; the count is decremented on every error path.
func (t *MutableTree) getImmutable(version int64, register bool) (*ImmutableTree, error) {
	reg := register && version > 0
	if reg {
		t.ndb.incrVersionReaders(version)
	}
	nkBytes, _, err := t.ndb.GetRoot(version)
	if err != nil {
		if reg {
			t.ndb.decrVersionReaders(version)
		}
		return nil, err
	}
	if nkBytes == nil {
		// Empty saved version: hold the reservation (if any) until Close.
		imm := NewImmutableTree(nil, version)
		imm.ndb = t.ndb
		imm.registered = reg
		return imm, nil
	}
	root, err := t.loadNode(nkBytes)
	if err != nil {
		if reg {
			t.ndb.decrVersionReaders(version)
		}
		return nil, err
	}
	imm := t.newImmutable(root, version, true)
	imm.registered = reg
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
	return fmt.Errorf("%w: DeleteVersionsFrom (use PruneVersionsTo)", ErrUnsupported)
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
//
// The snapshot wraps the LIVE working tree, whose most recent Sets may live
// only in pendingVals (uncommitted), so it resolves with read-your-writes
// (committed=false). It is a single-writer-only convenience: unlike
// GetImmutable, it is NOT safe to read concurrently with the writer.
func (t *MutableTree) Snapshot(version int64) *ImmutableTree {
	return t.newImmutable(t.root, version, false)
}

// VersionExists returns true if the given version exists.
func (t *MutableTree) VersionExists(version int64) bool {
	return t.ndb.VersionExists(version)
}

// AvailableVersions returns all available version numbers.
func (t *MutableTree) AvailableVersions() []int {
	return t.ndb.AvailableVersions()
}

// SetInitialVersion sets the version number for the first SaveVersion.
func (t *MutableTree) SetInitialVersion(version uint64) {
	t.initialVersion = version
}

// Rollback discards all mutations since the last save. Values staged this
// session live only in the uncommitted batch, so discarding it drops them —
// nothing was written to the DB to clean up.
func (t *MutableTree) Rollback() {
	t.ndb.DiscardBatch()
	t.resetSession()

	t.root = t.lastSaved
	if t.root != nil {
		t.size = nodeSize(t.root)
	} else {
		t.size = 0
	}
	// Rollback restores every field a poisoning failure left inconsistent,
	// so the session is clean again.
	t.poisoned = nil
}

// Height returns the tree height.
func (t *MutableTree) Height() int8 {
	if t.root == nil {
		return 0
	}
	return int8(nodeHeight(t.root))
}

// GetValueByKey resolves a valueKey to the raw value bytes, consulting the
// uncommitted working-session buffer first (read-your-writes). Safe only on the
// single writer goroutine.
func (t *MutableTree) GetValueByKey(vk []byte) ([]byte, error) {
	return t.resolveValue(vk)
}

// GetCommittedValueByKey resolves a valueKey to the raw value bytes from the DB
// ONLY (never the uncommitted working-session buffer). It is the race-free read
// for cross-package committed-snapshot consumers (the store's GetImmutable /
// proof resolvers), which run concurrently with the writer.
func (t *MutableTree) GetCommittedValueByKey(vk []byte) ([]byte, error) {
	return t.ndb.getCommittedValue(vk)
}

// Close closes the tree and its underlying DB resources.
func (t *MutableTree) Close() error {
	return t.ndb.Close()
}

// GetByIndex returns the key and value at the given zero-based index.
func (t *MutableTree) GetByIndex(index int64) ([]byte, []byte, error) {
	if t.root == nil || index < 0 || index >= t.size {
		return nil, nil, ErrKeyDoesNotExist
	}
	key, _, vk, err := treeGetByIndex(t.root, index)
	if err != nil {
		return nil, nil, err
	}
	val, err := t.resolveValue(vk)
	return key, val, err
}

// GetWithIndex returns the index, value, and whether the key was found.
func (t *MutableTree) GetWithIndex(key []byte) (int64, []byte, error) {
	if t.root == nil {
		return 0, nil, nil
	}
	idx, _, vk, found, err := treeGetWithIndex(t.root, key)
	if err != nil {
		return 0, nil, err
	}
	if !found {
		return idx, nil, nil
	}
	val, err := t.resolveValue(vk)
	return idx, val, err
}

// Iterate calls fn for each key-value pair in sorted order.
// Values are resolved from the value store.
func (t *MutableTree) Iterate(fn func(key []byte, value []byte) bool) (bool, error) {
	if t.root == nil {
		return false, nil
	}
	var resolveErr error
	stopped, walkErr := iterateNodeResolved(t.root, func(key, vk []byte) bool {
		val, err := t.resolveValue(vk)
		if err != nil {
			resolveErr = err
			return true // stop
		}
		return fn(key, val)
	})
	if walkErr != nil {
		return stopped, walkErr
	}
	return stopped, resolveErr
}

// --- helpers ---

func treeLookup(node Node, key []byte) (*LeafNode, Hash, []byte, bool, error) {
	for {
		switch n := node.(type) {
		case *LeafNode:
			pos, found := searchLeaf(n, key)
			if !found {
				return n, Hash{}, nil, false, nil
			}
			return n, n.valueHashes[pos], n.valueKeys[pos], true, nil
		case *InnerNode:
			idx := searchInner(n, key)
			child, err := n.getChild(idx)
			if err != nil {
				return nil, Hash{}, nil, false, err
			}
			if child == nil {
				// Defensive; unreachable on healthy trees (in-range children
				// always have a node or a ref).
				return nil, Hash{}, nil, false, nil
			}
			node = child
		default:
			panic("unknown node type")
		}
	}
}

// treeHas reports whether key exists under node. It wraps treeLookup, which
// returns five values; Has only needs found and err.
func treeHas(node Node, key []byte) (bool, error) {
	_, _, _, found, err := treeLookup(node, key) //nolint:dogsled // treeLookup returns 5 values; only found+err are needed here
	if err != nil {
		return false, err
	}
	return found, nil
}

func treeGetByIndex(node Node, index int64) ([]byte, Hash, []byte, error) {
	switch n := node.(type) {
	case *LeafNode:
		// Copy the key: it is returned to external callers (GetByIndex) and
		// embedded in non-membership proofs; the raw slice belongs to a live
		// leaf shared with the tree and node cache.
		return copyKey(n.keys[index]), n.valueHashes[index], n.valueKeys[index], nil
	case *InnerNode:
		offset := int64(0)
		for i := 0; i < n.NumChildren(); i++ {
			childSize := n.childSizes[i]
			if index < offset+childSize {
				child, err := n.getChild(i)
				if err != nil {
					return nil, Hash{}, nil, err
				}
				return treeGetByIndex(child, index-offset)
			}
			offset += childSize
		}
		panic("index out of range in treeGetByIndex")
	default:
		panic("unknown node type")
	}
}

func treeGetWithIndex(node Node, key []byte) (int64, Hash, []byte, bool, error) {
	switch n := node.(type) {
	case *LeafNode:
		pos, found := searchLeaf(n, key)
		if !found {
			return int64(pos), Hash{}, nil, false, nil
		}
		return int64(pos), n.valueHashes[pos], n.valueKeys[pos], true, nil
	case *InnerNode:
		childIdx := searchInner(n, key)
		offset := int64(0)
		for i := range childIdx {
			offset += n.childSizes[i]
		}
		child, err := n.getChild(childIdx)
		if err != nil {
			return 0, Hash{}, nil, false, err
		}
		idx, vh, vk, found, err := treeGetWithIndex(child, key)
		return offset + idx, vh, vk, found, err
	default:
		panic("unknown node type")
	}
}

// iterateNode walks the subtree passing raw value HASHES to the callback. On
// a child-load error it returns (true, err) — "stopped abnormally"; stopped
// is meaningless when err != nil.
func iterateNode(node Node, fn func(key, value []byte) bool) (bool, error) {
	switch n := node.(type) {
	case *LeafNode:
		for i := 0; i < int(n.numKeys); i++ {
			if fn(n.keys[i], n.valueHashes[i][:]) {
				return true, nil
			}
		}
		return false, nil
	case *InnerNode:
		for i := 0; i < n.NumChildren(); i++ {
			child, err := n.getChild(i)
			if err != nil {
				return true, err
			}
			if child != nil {
				stopped, err := iterateNode(child, fn)
				if stopped || err != nil {
					return stopped, err
				}
			}
		}
		return false, nil
	default:
		panic("unknown node type")
	}
}

// iterateNodeResolved is like iterateNode but passes valueKeys to the callback
// instead of valueHashes, enabling value resolution via ValueKey. On a
// child-load error it returns (true, err); stopped is meaningless when
// err != nil.
func iterateNodeResolved(node Node, fn func(key, vk []byte) bool) (bool, error) {
	switch n := node.(type) {
	case *LeafNode:
		for i := 0; i < int(n.numKeys); i++ {
			// Copy the key: it reaches the caller's callback, and the raw
			// slice belongs to a live leaf shared with the tree and cache.
			if fn(copyKey(n.keys[i]), n.valueKeys[i]) {
				return true, nil
			}
		}
		return false, nil
	case *InnerNode:
		for i := 0; i < n.NumChildren(); i++ {
			child, err := n.getChild(i)
			if err != nil {
				return true, err
			}
			if child != nil {
				stopped, err := iterateNodeResolved(child, fn)
				if stopped || err != nil {
					return stopped, err
				}
			}
		}
		return false, nil
	default:
		panic("unknown node type")
	}
}
