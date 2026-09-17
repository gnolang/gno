package bptree

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"sync"

	lru "github.com/hashicorp/golang-lru/v2"
	"golang.org/x/sync/singleflight"

	dbm "github.com/gnolang/gno/tm2/pkg/db"
)

// nodeDB handles persistence: reading/writing nodes, values, and root
// references to the underlying key-value store.
type nodeDB struct {
	db    dbm.DB
	batch dbm.Batch
	opts  Options

	nodeCache *lru.Cache[string, Node] // keyed by serialized NodeKey

	// loadGroup coalesces concurrent cache-miss GetNode loads of the same
	// NodeKey so only one goroutine deserializes and caches it, and the rest
	// share that instance (no duplicate deserialize, single object identity).
	// See M13.
	loadGroup singleflight.Group

	// pendingVals buffers values staged since the last Commit, keyed by
	// string(ValueKey). Each staged value is written into the batch (flushed
	// atomically with nodes/root at Commit) AND mirrored here so GetValue
	// resolves reads issued before SaveVersion (read-your-writes). Cleared by
	// Commit and DiscardBatch. Single-writer by design — like batch and
	// nextNonce, it is mutated only on the Set/SaveVersion/Rollback path, which
	// the ABCI connection mutex serialises against query reads.
	pendingVals map[string][]byte

	mtx            sync.Mutex
	pruneMu        sync.RWMutex // serializes prune vs version-reader registration (H3)
	latestVersion  int64
	firstVersion   int64
	versionReaders map[int64]uint32

	logger Logger

	nextNonce uint32 // per-SaveVersion nonce counter
}

func newNodeDB(db dbm.DB, cacheSize int, logger Logger, opts Options) *nodeDB {
	var cache *lru.Cache[string, Node]
	if cacheSize > 0 {
		var err error
		cache, err = lru.New[string, Node](cacheSize)
		if err != nil {
			panic(err)
		}
	}
	return &nodeDB{
		db:             db,
		batch:          db.NewBatch(),
		opts:           opts,
		nodeCache:      cache,
		pendingVals:    make(map[string][]byte),
		versionReaders: make(map[int64]uint32),
		logger:         logger,
	}
}

// --- DB key construction ---

func nodeDBKey(nk []byte) []byte {
	key := make([]byte, 1+len(nk))
	key[0] = PrefixNode
	copy(key[1:], nk)
	return key
}

func valueDBKey(vk []byte) []byte {
	key := make([]byte, 1+len(vk))
	key[0] = PrefixVal
	copy(key[1:], vk)
	return key
}

func orphanDBKey(version int64) []byte {
	key := make([]byte, 1+8)
	key[0] = PrefixOrphan
	binary.BigEndian.PutUint64(key[1:], uint64(version))
	return key
}

func rootDBKey(version int64) []byte {
	key := make([]byte, 1+8)
	key[0] = PrefixRoot
	binary.BigEndian.PutUint64(key[1:], uint64(version))
	return key
}

// --- Node operations ---

// SaveNode writes a node to the batch and adds it to the cache.
func (ndb *nodeDB) SaveNode(node Node) error {
	nk := node.GetNodeKey()
	if nk == nil {
		return ErrNodeMissingNodeKey
	}
	nkBytes := nk.GetKey()

	var buf bytes.Buffer
	switch n := node.(type) {
	case *InnerNode:
		if err := n.Serialize(&buf); err != nil {
			return fmt.Errorf("serializing inner node: %w", err)
		}
		// Give in-memory-built inner nodes their ndb so they can lazy-load
		// children after the working tree drops the in-memory child pointers.
		n.ndb = ndb
	case *LeafNode:
		if err := n.Serialize(&buf); err != nil {
			return fmt.Errorf("serializing leaf node: %w", err)
		}
	default:
		return fmt.Errorf("unknown node type")
	}

	if err := ndb.batch.Set(nodeDBKey(nkBytes), stampChecksum(buf.Bytes())); err != nil {
		return err
	}

	if ndb.nodeCache != nil {
		ndb.nodeCache.Add(string(nkBytes), node)
	}
	return nil
}

// GetNode loads a node from cache or DB.
func (ndb *nodeDB) GetNode(nkBytes []byte) (Node, error) {
	key := string(nkBytes)

	// Fast path: cache hit.
	if ndb.nodeCache != nil {
		if node, ok := ndb.nodeCache.Get(key); ok {
			return node, nil
		}
	}

	// Slow path: coalesce concurrent cache-miss loads of the same NodeKey via
	// singleflight, so only one goroutine does the DB read + deserialize + cache
	// Add and the rest share that instance (M13).
	v, err, _ := ndb.loadGroup.Do(key, func() (any, error) {
		// Re-check the cache: another caller may have populated it while we
		// waited on the singleflight.
		if ndb.nodeCache != nil {
			if node, ok := ndb.nodeCache.Get(key); ok {
				return node, nil
			}
		}

		data, err := ndb.db.Get(nodeDBKey(nkBytes))
		if err != nil {
			return nil, fmt.Errorf("db get node: %w", err)
		}
		if data == nil {
			return nil, fmt.Errorf("node not found: %x", nkBytes)
		}
		payload, err := verifyChecksum(data)
		if err != nil {
			return nil, fmt.Errorf("node record %x: %w", nkBytes, err)
		}

		nk := GetNodeKey(nkBytes)
		node, err := ReadNode(nk, payload)
		if err != nil {
			return nil, fmt.Errorf("deserializing node: %w", err)
		}

		// Set ndb on inner nodes for lazy child loading.
		if inner, ok := node.(*InnerNode); ok {
			inner.ndb = ndb
		}

		if ndb.nodeCache != nil {
			ndb.nodeCache.Add(key, node)
		}
		return node, nil
	})
	if err != nil {
		return nil, err
	}
	return v.(Node), nil
}

// --- Value operations ---

// SaveValue stages a value for the current session: it is buffered in
// pendingVals (so GetValue resolves it before SaveVersion) and written into
// the batch (so it is flushed atomically with the nodes/root at Commit).
// Nothing is written to the DB until Commit; Rollback/DiscardBatch drop it.
//
// The pendingVals buffer and the staged batch record are INDEPENDENT
// allocations: some backends (memdb, boltdb) retain the staged slice by
// reference all the way into the committed store, so sharing one buffer would
// alias the committed record to the read-your-writes buffer.
func (ndb *nodeDB) SaveValue(value, vk []byte) error {
	valCopy := make([]byte, len(value))
	copy(valCopy, value)
	ndb.pendingVals[string(vk)] = valCopy
	return ndb.batch.Set(valueDBKey(vk), stampChecksum(value))
}

// GetValue loads a value by its ValueKey, checking the uncommitted session
// buffer first (read-your-writes before SaveVersion), then the DB.
//
// pendingVals is the single-writer working-session buffer, so GetValue serves
// the working tree's OWN read-your-writes ONLY (MutableTree.resolveValue and
// working-tree iteration), all on the writer goroutine. Concurrent
// committed-snapshot readers must use getCommittedValue, which never touches
// the map and so cannot race SaveValue.
func (ndb *nodeDB) GetValue(vk []byte) ([]byte, error) {
	if v, ok := ndb.pendingVals[string(vk)]; ok {
		// Return a copy: a caller mutating the result must not change what
		// later reads of the staged value observe.
		return copyKey(v), nil
	}
	return ndb.getCommittedValue(vk)
}

// getCommittedValue loads a value by its ValueKey from the DB ONLY (no
// pendingVals). It is the race-free read path for committed snapshots
// (ImmutableTree / store query / proof / Export / snapshot iterators), which
// run concurrently with the writer and never legitimately need the uncommitted
// buffer (a committed version resolves only valueKeys < workingVersion).
//
// A stored empty value ([]byte{}) returns a non-nil empty slice, while a
// ValueKey absent from the DB (corruption or an already-pruned value — neither
// expected when called via the tree) returns a wrapped ErrKeyDoesNotExist, so
// callers can distinguish missing from empty.
func (ndb *nodeDB) getCommittedValue(vk []byte) ([]byte, error) {
	key := valueDBKey(vk)
	data, err := ndb.db.Get(key)
	if err != nil {
		return nil, fmt.Errorf("db get value: %w", err)
	}
	if data != nil {
		payload, err := verifyChecksum(data)
		if err != nil {
			return nil, fmt.Errorf("value record %x: %w", vk, err)
		}
		// Copy, never re-slice: some backends (memdb) return their internal
		// storage from Get, so handing out the payload would let a caller
		// mutate the committed record in place.
		return copyKey(payload), nil
	}
	// Every stored record carries a checksum (>= 4 bytes), so nil from Get
	// means "missing" on every backend; the Has() check just keeps the error
	// distinct if a backend ever returns nil for a present key.
	has, herr := ndb.db.Has(key)
	if herr != nil {
		return nil, fmt.Errorf("db has value: %w", herr)
	}
	if !has {
		return nil, fmt.Errorf("%w: valueKey %x", ErrKeyDoesNotExist, vk)
	}
	return nil, fmt.Errorf("value record %x: %w: present but empty", vk, ErrChecksumMismatch)
}

// DeleteValue adds a value deletion to the batch (committed at prune time).
func (ndb *nodeDB) DeleteValue(vk []byte) error {
	return ndb.batch.Delete(valueDBKey(vk))
}

// DeleteValueDirect drops a value staged earlier this session (Tier 1
// intra-version orphan): remove it from the buffer and stage a batch delete.
// In the batch the earlier Set then this Delete for the same key net to
// "absent" on Write (every dbm backend replays ops in order, later wins).
func (ndb *nodeDB) DeleteValueDirect(vk []byte) error {
	delete(ndb.pendingVals, string(vk))
	return ndb.batch.Delete(valueDBKey(vk))
}

// --- Orphan list operations ---

// SaveOrphans persists a list of orphaned ValueKeys for a version.
// Written to batch (committed atomically with nodes and root).
func (ndb *nodeDB) SaveOrphans(version int64, orphans [][]byte) error {
	if len(orphans) == 0 {
		return nil // don't write empty orphan records
	}
	// Encode: count(uvarint) + N * NodeKeySize bytes
	size := binary.MaxVarintLen64 + len(orphans)*NodeKeySize
	buf := make([]byte, 0, size)
	var vbuf [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(vbuf[:], uint64(len(orphans)))
	buf = append(buf, vbuf[:n]...)
	for _, vk := range orphans {
		buf = append(buf, vk...)
	}
	return ndb.batch.Set(orphanDBKey(version), stampChecksum(buf))
}

// LoadOrphans loads the orphan list for a version from the DB.
func (ndb *nodeDB) LoadOrphans(version int64) ([][]byte, error) {
	data, err := ndb.db.Get(orphanDBKey(version))
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}
	// A 0-length record is corrupt, not "no orphans": SaveOrphans never
	// writes empty lists, and every legitimate record carries a checksum.
	payload, err := verifyChecksum(data)
	if err != nil {
		return nil, fmt.Errorf("orphan record v%d: %w", version, err)
	}
	r := bytes.NewReader(payload)
	count, err := binary.ReadUvarint(r)
	if err != nil {
		return nil, fmt.Errorf("reading orphan count: %w", err)
	}
	// Validate the untrusted count against the bytes actually present before
	// allocating (a corrupt record could claim a huge count and OOM).
	if count > uint64(r.Len())/NodeKeySize {
		return nil, fmt.Errorf("orphan count %d exceeds record size %d", count, len(data))
	}
	orphans := make([][]byte, count)
	for i := range orphans {
		orphans[i] = make([]byte, NodeKeySize)
		if _, err := io.ReadFull(r, orphans[i]); err != nil {
			return nil, fmt.Errorf("reading orphan %d: %w", i, err)
		}
	}
	return orphans, nil
}

// DeleteOrphans removes the orphan list for a version from the batch.
func (ndb *nodeDB) DeleteOrphans(version int64) error {
	return ndb.batch.Delete(orphanDBKey(version))
}

// --- Root operations ---

// SaveRoot writes a root reference: NodeKey (12B) + root hash (32B) = 44B.
// For empty trees, nk is nil but hash is still stored (just the 32-byte hash).
func (ndb *nodeDB) SaveRoot(version int64, nk *NodeKey, hash []byte) error {
	if nk == nil && len(hash) == 0 {
		// Both callers always supply a 32-byte hash (an empty tree stores
		// emptyHash()); a hashless record would be indistinguishable from
		// corruption on reload.
		return fmt.Errorf("bptree: SaveRoot v%d: nil node key and empty hash", version)
	}
	var val []byte
	if nk != nil {
		val = make([]byte, 0, NodeKeySize+HashSize)
		val = append(val, nk.GetKey()...)
		val = append(val, hash...)
	} else {
		// Empty tree: store just the hash (no NodeKey prefix)
		val = make([]byte, 0, HashSize)
		val = append(val, hash...)
	}
	return ndb.batch.Set(rootDBKey(version), stampChecksum(val))
}

// GetRoot loads the root reference for a version.
// Returns (nodeKey bytes, root hash, error). nodeKey is nil for empty tree.
func (ndb *nodeDB) GetRoot(version int64) ([]byte, []byte, error) {
	data, err := ndb.db.Get(rootDBKey(version))
	if err != nil {
		return nil, nil, fmt.Errorf("db get root: %w", err)
	}
	if data == nil {
		return nil, nil, ErrVersionDoesNotExist
	}
	payload, err := verifyChecksum(data)
	if err != nil {
		return nil, nil, fmt.Errorf("root record v%d: %w", version, err)
	}
	if len(payload) == HashSize {
		// Empty tree with hash only (no NodeKey)
		return nil, payload, nil
	}
	if len(payload) != NodeKeySize+HashSize {
		return nil, nil, fmt.Errorf("corrupt root ref: len=%d", len(payload))
	}
	return payload[:NodeKeySize], payload[NodeKeySize:], nil
}

// --- Version management ---

func (ndb *nodeDB) getLatestVersion() int64 {
	ndb.mtx.Lock()
	defer ndb.mtx.Unlock()
	return ndb.latestVersion
}

func (ndb *nodeDB) setLatestVersion(v int64) {
	ndb.mtx.Lock()
	defer ndb.mtx.Unlock()
	ndb.latestVersion = v
}

func (ndb *nodeDB) getFirstVersion() int64 {
	ndb.mtx.Lock()
	defer ndb.mtx.Unlock()
	return ndb.firstVersion
}

func (ndb *nodeDB) setFirstVersion(v int64) {
	ndb.mtx.Lock()
	defer ndb.mtx.Unlock()
	ndb.firstVersion = v
}

// VersionExists checks if a root reference exists for the given version.
//
// This boolean form reports a DB error as "does not exist" (logged at Error
// level so the failure is observable). It is fine for callers that are robust
// to that — the store-layer compat surface and informational queries. Callers
// that must distinguish "absent" from "DB error" — notably SaveVersion, where
// treating a transient failure as absent would overwrite an existing version
// with unverified data — MUST use versionExistsE instead.
func (ndb *nodeDB) VersionExists(version int64) bool {
	has, err := ndb.db.Has(rootDBKey(version))
	if err != nil {
		ndb.logger.Error("bptree: VersionExists DB error", "version", version, "err", err)
		return false
	}
	return has
}

// versionExistsE is the error-propagating variant of VersionExists, for code
// paths where a DB failure must not be silently read as "does not exist".
func (ndb *nodeDB) versionExistsE(version int64) (bool, error) {
	return ndb.db.Has(rootDBKey(version))
}

// AvailableVersions returns all versions that have root references, ascending.
// A single PrefixRoot scan (root keys are PrefixRoot‖version-BE, so iteration is
// already version-ordered) rather than one db.Has per version in [first, latest].
func (ndb *nodeDB) AvailableVersions() []int {
	prefix := []byte{PrefixRoot}
	end := make([]byte, len(prefix))
	copy(end, prefix)
	end[0]++

	itr, err := ndb.db.Iterator(prefix, end)
	if err != nil {
		return nil
	}
	defer itr.Close()

	var versions []int
	for ; itr.Valid(); itr.Next() {
		key := itr.Key()
		if len(key) != 9 { // prefix(1) + version(8)
			continue
		}
		versions = append(versions, int(binary.BigEndian.Uint64(key[1:])))
	}
	if itr.Error() != nil {
		// Don't return a silently-truncated list on a mid-scan DB error; mirror
		// the Iterator()-open error path. (discoverVersions propagates instead.)
		return nil
	}
	return versions
}

// discoverVersions sets the oldest and newest saved versions.
func (ndb *nodeDB) discoverVersions() error {
	first, err := ndb.edgeRootVersion(false)
	if err != nil {
		return err
	}
	latest, err := ndb.edgeRootVersion(true)
	if err != nil {
		return err
	}

	ndb.mtx.Lock()
	ndb.firstVersion = first
	ndb.latestVersion = latest
	ndb.mtx.Unlock()
	return nil
}

// edgeRootVersion returns the oldest (reverse=false) or newest (reverse=true)
// saved version, or 0 if none. Root keys are 'R' followed by the version in
// big-endian, so sorting them sorts the versions and each end is one seek:
//
//	R 00 00 00 00 00 00 00 03   <- oldest, reverse=false returns 3
//	R 00 00 00 00 00 00 00 04
//	R 00 00 00 00 00 00 00 05   <- newest, reverse=true returns 5
func (ndb *nodeDB) edgeRootVersion(reverse bool) (int64, error) {
	prefix := []byte{PrefixRoot}
	end := make([]byte, len(prefix))
	copy(end, prefix)
	end[0]++

	var (
		itr dbm.Iterator
		err error
	)
	if reverse {
		itr, err = ndb.db.ReverseIterator(prefix, end)
	} else {
		itr, err = ndb.db.Iterator(prefix, end)
	}
	if err != nil {
		return 0, err
	}
	defer itr.Close()

	for ; itr.Valid(); itr.Next() {
		key := itr.Key()
		if len(key) != 9 { // prefix(1) + version(8)
			continue
		}
		return int64(binary.BigEndian.Uint64(key[1:])), itr.Error()
	}
	return 0, itr.Error()
}

// --- Version readers ---

func (ndb *nodeDB) incrVersionReaders(version int64) {
	// Take pruneMu (shared) FIRST: registrations block while a prune holds it
	// exclusively, closing the check-vs-register TOCTOU. A reader that races an
	// in-flight prune waits here until the prune finishes, by which point the
	// version is gone and the caller's GetRoot/GetNode fails naturally.
	// Concurrent registrations don't block each other.
	ndb.pruneMu.RLock()
	defer ndb.pruneMu.RUnlock()
	ndb.mtx.Lock()
	defer ndb.mtx.Unlock()
	ndb.versionReaders[version]++
}

// beginPruning checks that no version in [first, to] has active readers and, if
// clear, holds the exclusive pruneMu for the duration of the prune so no new
// reader can register a to-be-deleted version. Callers MUST call endPruning to
// release it (typically `defer`). mtx is released before returning; only pruneMu
// is held across the prune (the prune body itself takes mtx via
// getFirstVersion/setFirstVersion, so holding mtx here would self-deadlock).
func (ndb *nodeDB) beginPruning(first, to int64) error {
	ndb.pruneMu.Lock()
	ndb.mtx.Lock()
	for v := first; v <= to; v++ {
		if ndb.versionReaders[v] > 0 {
			ndb.mtx.Unlock()
			ndb.pruneMu.Unlock()
			return fmt.Errorf("%w: version %d", ErrActiveReaders, v)
		}
	}
	ndb.mtx.Unlock()
	return nil
}

// endPruning releases the exclusive pruneMu acquired by beginPruning.
func (ndb *nodeDB) endPruning() {
	ndb.pruneMu.Unlock()
}

func (ndb *nodeDB) decrVersionReaders(version int64) {
	ndb.mtx.Lock()
	defer ndb.mtx.Unlock()
	if ndb.versionReaders[version] > 0 {
		ndb.versionReaders[version]--
		if ndb.versionReaders[version] == 0 {
			delete(ndb.versionReaders, version)
		}
	}
}

// --- Batch operations ---

// Commit flushes the current batch to disk and creates a new batch.
// Always closes the old batch and creates a new one, even on error,
// to avoid leaving the nodeDB in a broken state.
func (ndb *nodeDB) Commit() error {
	var err error
	if ndb.opts.Sync {
		err = ndb.batch.WriteSync()
	} else {
		err = ndb.batch.Write()
	}
	// Staged values are now durable in the DB; recycle the batch and clear the
	// session buffer (so subsequent GetValue reads from disk).
	ndb.DiscardBatch()
	return err
}

// DiscardBatch drops every write staged since the last Commit (values AND
// nodes) and starts a fresh batch. Used by Commit (after a successful Write),
// by Rollback, and by SaveVersion's non-committing exits (error/idempotent):
// staged writes must not survive to be flushed by a later Commit, which would
// corrupt an existing version.
func (ndb *nodeDB) DiscardBatch() {
	if ndb.batch != nil {
		ndb.batch.Close()
	}
	ndb.batch = ndb.db.NewBatch()
	clear(ndb.pendingVals)
}

// ResetNonce resets the per-version nonce counter.
func (ndb *nodeDB) ResetNonce() {
	ndb.nextNonce = 0
}

// NextNodeKey returns a new NodeKey for the given version with an
// auto-incrementing nonce.
func (ndb *nodeDB) NextNodeKey(version int64) *NodeKey {
	ndb.nextNonce++
	return &NodeKey{Version: version, Nonce: ndb.nextNonce}
}

// Close closes the nodeDB batch. The underlying DB is NOT closed
// because it may be shared by other trees.
func (ndb *nodeDB) Close() error {
	if ndb.batch != nil {
		ndb.batch.Close()
		ndb.batch = nil
	}
	return nil
}

// DeleteRoot stages deletion of a version's root reference in the batch.
// Used by PruneVersionsTo after the version's orphaned nodes are deleted.
func (ndb *nodeDB) DeleteRoot(version int64) error {
	return ndb.batch.Delete(rootDBKey(version))
}

// DeleteNode stages deletion of a node in the batch (used during pruning).
// Evicting the cache before the batch flushes is safe: the prune never re-reads
// keys it deleted, and registered readers of retained versions never reference
// them (dual-walk deletes only unshared nodes; beginPruning excludes readers of
// the pruned range), so no in-contract GetNode can observe the window.
func (ndb *nodeDB) DeleteNode(nkBytes []byte) error {
	if ndb.nodeCache != nil {
		ndb.nodeCache.Remove(string(nkBytes))
	}
	return ndb.batch.Delete(nodeDBKey(nkBytes))
}
