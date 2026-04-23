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

	// inlineThreshold: values of this size or smaller are stored inline
	// in the leaf (no ValueKey indirection / no separate PrefixVal
	// record). Larger values continue to use the external path. A
	// negative value disables inlining entirely.
	inlineThreshold int

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
	// fastNodesCap mirrors the LRU's configured capacity. The lru
	// library does not expose this back, so we record it at construction
	// to drive the working-set-vs-capacity comparison in
	// reconcileFastNodeState.
	fastNodesCap int
	// fastNodesOff is set by reconcileFastNodeState when t.size grows
	// past fastNodesCap * fastNodeWorkingSetMultiplier; while true,
	// the cache is treated as nil for read and write purposes.
	fastNodesOff bool
}

// fastNodeWorkingSetMultiplier governs the size-vs-capacity ratio at
// which the fast-node cache is suspended. Once the tree holds more
// than capacity*multiplier keys the LRU's hit ratio collapses (random
// access into a working set N× larger than the cache evicts before
// re-use), so cache lookups become pure overhead per Get and cache
// populates become churn per Set. 4 keeps the cache active across
// modestly-larger-than-capacity workloads while suspending it when
// it would otherwise thrash.
const fastNodeWorkingSetMultiplier = 4

// NewMutableTreeMem creates an in-memory MutableTree (no DB). Inline
// value storage is off by default; enable via a future explicit option
// if an in-memory constructor variant requires it.
func NewMutableTreeMem() *MutableTree {
	t := &MutableTree{
		logger:          NewNopLogger(),
		memValues:       make(map[string][]byte),
		nextValueNonce:  1, // nonce=0 is reserved to avoid collision with the "missing" sentinel (Finding #6)
		inlineThreshold: -1,
	}
	t.initFastNodeCache(DefaultFastNodeCacheSize)
	return t
}

// resolveInlineThreshold turns an Options value into the effective
// threshold. The zero value (and any negative value) maps to -1
// (disabled, every value goes external). Callers opt into inline
// storage by passing a positive threshold via
// InlineValueThresholdOption — DefaultInlineValueThreshold is the
// recommended starting value.
//
// Disabling-by-default preserves the external-ValueKey invariant for
// existing callers and tests; a downstream consumer (e.g. the
// store wrapper at tm2/pkg/store/bptree) opts in explicitly when it
// wants the inline-storage performance win.
func resolveInlineThreshold(opt int) int {
	if opt <= 0 {
		return -1
	}
	return opt
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
	t.fastNodesCap = size
}

// fastNodeActive reports whether the fast-node cache is currently
// usable for reads/writes. Inactive when the cache was never
// constructed or when reconcileFastNodeState has suspended it
// because the working set exceeded capacity * multiplier.
func (t *MutableTree) fastNodeActive() bool {
	return t.fastNodes != nil && !t.fastNodesOff
}

// reconcileFastNodeState toggles the fast-node cache off when the
// tree has grown past the working-set threshold and back on when it
// has shrunk below. The on→off transition purges all entries: while
// suspended the cache observes neither Sets nor Removes, so an entry
// from before the suspension cannot safely be served on a future
// re-activation. A re-activated cache therefore always starts empty
// and refills naturally from subsequent Sets and Get-hit populates.
func (t *MutableTree) reconcileFastNodeState() {
	if t.fastNodes == nil {
		return
	}
	threshold := int64(t.fastNodesCap) * fastNodeWorkingSetMultiplier
	over := t.size > threshold
	switch {
	case over && !t.fastNodesOff:
		t.fastNodes.Purge()
		t.fastNodesOff = true
	case !over && t.fastNodesOff:
		t.fastNodesOff = false
	}
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
		nextValueNonce:  1,
		inlineThreshold: resolveInlineThreshold(opts.InlineValueThreshold),
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

	// Small values inline into the leaf; large values take the external
	// ValueKey path. Only the external path allocates a ValueKey,
	// appends to sessionValues, and calls SaveValue — inline bytes live
	// directly on the node and ride along with leaf serialization at
	// SaveVersion time. The external path still persists before any
	// tree mutation (Finding #28).
	valueHash := sha256.Sum256(value)
	payload := slotPayload{valueHash: valueHash}
	if t.inlineThreshold >= 0 && len(value) <= t.inlineThreshold {
		// Inline: copy bytes so caller can retain/modify their slice.
		cp := make([]byte, len(value))
		copy(cp, value)
		payload.inline = cp
	} else {
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
		payload.valueKey = vk
	}

	if t.root == nil {
		leaf := &LeafNode{miniTree: NewMiniMerkle()}
		leaf.keys[0] = copyKey(key)
		leaf.valueHashes[0] = valueHash
		leaf.numKeys = 1
		if payload.inline != nil {
			leaf.inlineValues[0] = payload.inline
			leaf.inlineMask = 1
		} else {
			leaf.valueKeys[0] = payload.valueKey
		}
		// Fresh leaf: slotHashes cache is cold, do a full build now.
		leaf.RebuildMiniMerkle()
		t.root = leaf
		t.size = 1
		t.reconcileFastNodeState()
		t.cacheValueForKey(key, value, payload.inline)
		return false, nil
	}

	t.cowRoot()
	newRoot, updated, oldPayload := treeInsert(t.root, key, payload)
	t.root = newRoot
	if !updated {
		t.size++
		t.reconcileFastNodeState()
	}

	// Handle orphaned old valueKey on update. Only the external-slot
	// case produces an orphan — inline-slot displacements just drop
	// their bytes (the replaced leaf version carries them).
	if updated && oldPayload.valueKey != nil {
		t.orphanValueKey(oldPayload.valueKey)
	}

	// Refresh the fast-node cache with a private copy of the fresh
	// value — the caller retains ownership of the original slice and
	// may mutate it after Set returns.
	t.cacheValueForKey(key, value, payload.inline)
	return updated, nil
}

// cacheValueForKey populates the latest-view cache for `key`. When
// `owned` is non-nil it is an already-private slice (the inline-storage
// path's payload copy made at Set time) and is stored directly. When
// `owned` is nil the function copies `value` so caller-side mutation
// cannot corrupt the cache.
func (t *MutableTree) cacheValueForKey(key, value, owned []byte) {
	if !t.fastNodeActive() {
		return
	}
	cp := owned
	if cp == nil {
		cp = make([]byte, len(value))
		copy(cp, value)
	}
	t.fastNodes.Add(string(key), cp)
}

// Get retrieves the value for a key. The latest-view fast-node cache
// short-circuits the tree walk on a hit; misses fall through to the
// regular lookup and populate the cache for next time. The cache
// lookup happens ahead of the root-nil check so that a cached value
// from the previous working session remains readable after a
// hypothetical root swap — paranoia against mis-ordered cache
// invalidation elsewhere.
func (t *MutableTree) Get(key []byte) ([]byte, error) {
	if t.fastNodeActive() {
		if v, ok := t.fastNodes.Get(string(key)); ok {
			// Defensive copy on the hit path: returning the cached
			// slice directly would let a caller mutate it and corrupt
			// every future Get hit on the same key. Mirrors the
			// populate-path discipline (cacheValueForKey on Set, and
			// the copy below on the Get-miss-then-cache-add path).
			out := make([]byte, len(v))
			copy(out, v)
			return out, nil
		}
	}
	if t.root == nil {
		return nil, nil
	}
	leaf, slot, found := treeLookup(t.root, key)
	if !found {
		return nil, nil
	}
	val, err := leaf.valueAt(slot, t.valueResolverFn())
	if err == nil && val != nil && t.fastNodeActive() {
		// Defensive copy before caching: for inline slots,
		// leaf.valueAt returns a slice aliased to leaf.inlineValues[i].
		// Adding that shared slice straight into the cache would let a
		// caller-side mutation of the returned bytes (or a future
		// in-place mutation in this package, were the key-ownership
		// invariant ever extended to values) corrupt both the leaf and
		// the cache. Mirrors cacheValueForKey on the Set path.
		cached := make([]byte, len(val))
		copy(cached, val)
		t.fastNodes.Add(string(key), cached)
	}
	return val, err
}

// valueResolverFn returns a ValueResolver closure for external-slot
// lookups. Inline slots are handled by leaf.valueAt without invoking
// the resolver.
func (t *MutableTree) valueResolverFn() ValueResolver {
	return func(vk []byte) ([]byte, error) { return t.resolveValue(vk) }
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
	if t.fastNodeActive() {
		if _, ok := t.fastNodes.Get(string(key)); ok {
			return true, nil
		}
	}
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
	t.cachedRootHash = nil // working tree about to mutate — invalidate cache (Finding #21)
	t.cowRoot()
	newRoot, oldPayload, found := treeRemove(t.root, key)
	if !found {
		return nil, false, nil
	}
	t.root = newRoot
	t.size--
	t.reconcileFastNodeState()

	// Recover the displaced value — inline bytes come straight off the
	// leaf; external slots resolve via the value store and then orphan
	// the valueKey for prune-time cleanup.
	var val []byte
	if oldPayload.inline != nil {
		val = oldPayload.inline
	} else if oldPayload.valueKey != nil {
		var err error
		val, err = t.resolveValue(oldPayload.valueKey)
		if err != nil {
			// Log but don't fail the Remove: the slot has already been
			// shifted out of the leaf, so returning early would leave
			// the tree in a partially-mutated state with the caller
			// unable to retry coherently. Surfacing the error in the
			// log preserves the diagnostic trail without breaking the
			// Remove contract. The returned val is nil; callers that
			// need to distinguish resolution failure from a healthy
			// nil-old-value will need a wrapped error instead.
			t.logger.Error("bptree: resolveValue failed in Remove",
				"vk", fmt.Sprintf("%x", oldPayload.valueKey), "err", err)
		}
		t.orphanValueKey(oldPayload.valueKey)
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
		// Same hash — idempotent save. The session may have allocated
		// fresh ValueKeys via Set/SaveValue that the working tree's
		// leaves now reference. Those VKs are persisted on disk (see
		// Finding #28's atomicity rule) and will be carried into the
		// next SaveVersion via t.lastSaved = t.root. Clear the
		// per-session bookkeeping (sessionValues, versionOrphans,
		// nextValueNonce) so the state matches a normal successful
		// save — without this, a subsequent Rollback would attempt to
		// DeleteValueDirect VKs that t.lastSaved still references, and
		// the next SaveVersion would carry stale orphans/nonces forward.
		// See BUG-2 in PR-5750.md.
		t.sessionValues = t.sessionValues[:0]
		t.versionOrphans = t.versionOrphans[:0]
		t.nextValueNonce = 1
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
		// Mutation paths (innerInsert, redistribute, merge, split)
		// either keep the mini-merkle current via SetSlot or set
		// miniTreeDirty when they cannot. ensureMiniMerkleBuilt
		// rebuilds only when dirty; clean clones (no descendant
		// mutated through this inner) skip the 31-hash rebuild.
		inner.ensureMiniMerkleBuilt()
	}

	// Same dirty-bit gating for leaves: mutation paths set
	// miniTreeDirty + slotsDirty bits via markLeafSlotDirty /
	// markLeafSlotsDirtyRange. ensureMiniMerkleBuilt drives the
	// incremental rebuild that touches only changed slots.
	if leaf, ok := node.(*LeafNode); ok {
		leaf.ensureMiniMerkleBuilt()
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
	//
	// Snapshot the prior nonce so we can restore it if loadNode below
	// fails — leaving t.nextValueNonce seeded for a load that didn't
	// complete would let the next Set against the OLD t.version
	// allocate VKs in a different version's namespace, re-introducing
	// the value-overwrite hazard maxValueNonceForVersion exists to
	// prevent. See BUG-3 in PR-5750.md.
	priorNonce := t.nextValueNonce
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
		t.reconcileFastNodeState()
		t.version = version
		t.lastSaved = nil
		t.cachedRootHash = nil
		t.cachedSavedHash = nil
		return latestVersion, nil
	}

	// LoadVersion is a single-goroutine boot path on a freshly-constructed
	// MutableTree — no other reader can be racing on the same NodeKey.
	// Skip the singleflight wrapper around GetNode to save its per-call
	// `*call` + closure allocations on the cold-start hot path.
	root, err := t.ndb.getNodeUncontended(nkBytes)
	if err != nil {
		t.nextValueNonce = priorNonce
		return 0, fmt.Errorf("loading root: %w", err)
	}

	t.root = root
	t.size = nodeSize(root)
	t.reconcileFastNodeState()
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
	t.reconcileFastNodeState()
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
	leaf, slot := treeGetByIndex(t.root, index)
	val, err := leaf.valueAt(slot, t.valueResolverFn())
	return leaf.keys[slot], val, err
}

// GetWithIndex returns the index, value, and whether the key was found.
func (t *MutableTree) GetWithIndex(key []byte) (int64, []byte, error) {
	if t.root == nil {
		return 0, nil, nil
	}
	idx, leaf, slot, found := treeGetWithIndex(t.root, key)
	if !found {
		return idx, nil, nil
	}
	val, err := leaf.valueAt(slot, t.valueResolverFn())
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
	resolver := t.valueResolverFn()
	stopped := iterateNodeResolved(t.root, func(key []byte, leaf *LeafNode, slot int) bool {
		val, err := leaf.valueAt(slot, resolver)
		if err != nil {
			resolveErr = err
			return true // stop
		}
		return fn(key, val)
	})
	return stopped, resolveErr
}

// --- helpers ---

// treeLookup descends to the leaf containing `key`. Returns the leaf,
// the slot index within it, and whether the key exists. Callers
// resolve the value via leaf.valueAt(slot, resolver) so both inline
// and external payloads work through one path.
func treeLookup(node Node, key []byte) (*LeafNode, int, bool) {
	for {
		switch n := node.(type) {
		case *LeafNode:
			pos, found := searchLeaf(n, key)
			if !found {
				return n, pos, false
			}
			return n, pos, true
		case *InnerNode:
			idx := searchInner(n, key)
			child := n.getChild(idx)
			if child == nil {
				return nil, 0, false
			}
			node = child
		default:
			panic("unknown node type")
		}
	}
}

// treeGetByIndex returns (leaf, slotIdx) for the idx-th key in sort
// order. Caller reads key/value via leaf fields + valueAt.
func treeGetByIndex(node Node, index int64) (*LeafNode, int) {
	switch n := node.(type) {
	case *LeafNode:
		return n, int(index)
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

// treeGetWithIndex returns the sort-rank index of `key`, the leaf it
// would live in, the slot within that leaf, and whether it exists.
func treeGetWithIndex(node Node, key []byte) (int64, *LeafNode, int, bool) {
	switch n := node.(type) {
	case *LeafNode:
		pos, found := searchLeaf(n, key)
		return int64(pos), n, pos, found
	case *InnerNode:
		childIdx := searchInner(n, key)
		offset := int64(0)
		for i := 0; i < childIdx; i++ {
			offset += n.childSizes[i]
		}
		child := n.getChild(childIdx)
		idx, leaf, slot, found := treeGetWithIndex(child, key)
		return offset + idx, leaf, slot, found
	default:
		panic("unknown node type")
	}
}

// iterateNodeResolved walks node in key order, yielding each
// (key, leaf, slotIdx) triplet to fn. Returning true stops iteration.
// Callers resolve the value via leaf.valueAt(slotIdx, resolver).
func iterateNodeResolved(node Node, fn func(key []byte, leaf *LeafNode, slot int) bool) bool {
	switch n := node.(type) {
	case *LeafNode:
		for i := 0; i < int(n.numKeys); i++ {
			if fn(n.keys[i], n, i) {
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
