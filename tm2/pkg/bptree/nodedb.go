package bptree

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"sync"

	lru "github.com/hashicorp/golang-lru/v2"

	dbm "github.com/gnolang/gno/tm2/pkg/db"
)

// nodeDB handles persistence: reading/writing nodes, values, and root
// references to the underlying key-value store.
type nodeDB struct {
	db    dbm.DB
	batch dbm.Batch
	opts  Options

	nodeCache *lru.Cache[string, Node] // keyed by serialized NodeKey

	mtx            sync.Mutex
	latestVersion  int64
	firstVersion   int64
	versionReaders map[int64]uint32

	// pruneMu serialises prune with version-reader registration so that no
	// new reader can register while a prune is in flight. Acquired by
	// beginPruning and by incrVersionReaders. Always acquired BEFORE mtx
	// to avoid lock-order inversion. See Finding #15.
	pruneMu sync.Mutex

	isCommitting bool
	logger       Logger

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
	case *LeafNode:
		if err := n.Serialize(&buf); err != nil {
			return fmt.Errorf("serializing leaf node: %w", err)
		}
	default:
		return fmt.Errorf("unknown node type")
	}

	if err := ndb.batch.Set(nodeDBKey(nkBytes), buf.Bytes()); err != nil {
		return err
	}

	if ndb.nodeCache != nil {
		ndb.nodeCache.Add(string(nkBytes), node)
	}
	return nil
}

// GetNode loads a node from cache or DB.
//
// Error semantics (Finding #5):
//   - Returns (nil, ErrNodeNotFound) when the node's DB record is
//     absent. Callers (e.g. GetImmutable, the mark-and-sweep prune
//     path, loadNode) treat this as "pruned or never persisted" and
//     handle it as a recoverable condition.
//   - Panics on any other failure mode — a raw DB read error or a
//     deserialization error both indicate unrecoverable storage
//     corruption; continuing would propagate inconsistent state
//     through every hot read path. This mirrors getChild's invariant.
func (ndb *nodeDB) GetNode(nkBytes []byte) (Node, error) {
	// Check cache
	if ndb.nodeCache != nil {
		if node, ok := ndb.nodeCache.Get(string(nkBytes)); ok {
			return node, nil
		}
	}

	// Load from DB. A raw IO error here is unrecoverable — the tree
	// depends on reads succeeding to maintain its shape invariant.
	data, err := ndb.db.Get(nodeDBKey(nkBytes))
	if err != nil {
		panic(fmt.Sprintf("bptree: db get node %x: %v", nkBytes, err))
	}
	if data == nil {
		// Legitimate "missing" signal (pruned / never persisted).
		return nil, ErrNodeNotFound
	}

	nk := GetNodeKey(nkBytes)
	node, err := ReadNode(nk, data)
	if err != nil {
		// On-disk corruption — the tree's invariants are already
		// violated; panic rather than return garbage nodes.
		panic(fmt.Sprintf("bptree: deserializing node %x: %v", nkBytes, err))
	}

	// Set ndb on inner nodes for lazy child loading
	if inner, ok := node.(*InnerNode); ok {
		inner.ndb = ndb
	}

	if ndb.nodeCache != nil {
		ndb.nodeCache.Add(string(nkBytes), node)
	}
	return node, nil
}

// --- Value operations ---

// SaveValue writes a value to the DB directly (not via batch), keyed by
// ValueKey. Writing early allows Get to work before SaveVersion/Commit.
func (ndb *nodeDB) SaveValue(value, vk []byte) error {
	key := valueDBKey(vk)
	valCopy := make([]byte, len(value))
	copy(valCopy, value)
	return ndb.db.Set(key, valCopy)
}

// GetValue loads a value by its ValueKey from the DB.
func (ndb *nodeDB) GetValue(vk []byte) ([]byte, error) {
	data, err := ndb.db.Get(valueDBKey(vk))
	if err != nil {
		return nil, fmt.Errorf("db get value: %w", err)
	}
	return data, nil
}

// DeleteValue adds a value deletion to the batch (committed at prune time).
func (ndb *nodeDB) DeleteValue(vk []byte) error {
	return ndb.batch.Delete(valueDBKey(vk))
}

// DeleteValueDirect deletes a value directly from the DB (for Tier 1
// intra-version orphans and Rollback — matching SaveValue's eager writes).
func (ndb *nodeDB) DeleteValueDirect(vk []byte) error {
	return ndb.db.Delete(valueDBKey(vk))
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
	return ndb.batch.Set(orphanDBKey(version), buf)
}

// LoadOrphans loads the orphan list for a version from the DB.
func (ndb *nodeDB) LoadOrphans(version int64) ([][]byte, error) {
	data, err := ndb.db.Get(orphanDBKey(version))
	if err != nil {
		return nil, err
	}
	if data == nil || len(data) == 0 {
		return nil, nil
	}
	r := bytes.NewReader(data)
	count, err := binary.ReadUvarint(r)
	if err != nil {
		return nil, fmt.Errorf("reading orphan count: %w", err)
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
	var val []byte
	if nk != nil {
		val = make([]byte, 0, NodeKeySize+HashSize)
		val = append(val, nk.GetKey()...)
		val = append(val, hash...)
	} else if len(hash) > 0 {
		// Empty tree: store just the hash (no NodeKey prefix)
		val = make([]byte, 0, HashSize)
		val = append(val, hash...)
	}
	// val is nil/empty only if both nk and hash are nil/empty
	return ndb.batch.Set(rootDBKey(version), val)
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
	if len(data) == 0 {
		// Legacy empty tree (no hash stored)
		return nil, nil, nil
	}
	if len(data) == HashSize {
		// Empty tree with hash only (no NodeKey)
		return nil, data, nil
	}
	if len(data) != NodeKeySize+HashSize {
		return nil, nil, fmt.Errorf("corrupt root ref: len=%d", len(data))
	}
	return data[:NodeKeySize], data[NodeKeySize:], nil
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
func (ndb *nodeDB) VersionExists(version int64) bool {
	has, _ := ndb.db.Has(rootDBKey(version))
	return has
}

// AvailableVersions returns all versions that have root references,
// in ascending order. Uses a single prefix iterator over PrefixRoot so
// the cost is O(n) sequential DB reads rather than O(latest-first)
// random point lookups — the gap matters for long-lived chains that
// retain many versions. See Finding #14.
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
		// Keys are big-endian encoded versions; ascending iteration
		// yields them in version order (no sort needed).
		v := int64(binary.BigEndian.Uint64(key[1:]))
		versions = append(versions, int(v))
	}
	return versions
}

// discoverVersions scans the DB for root references to find
// the first and latest versions. Called during Load.
func (ndb *nodeDB) discoverVersions() error {
	prefix := []byte{PrefixRoot}
	end := make([]byte, len(prefix))
	copy(end, prefix)
	end[0]++

	itr, err := ndb.db.Iterator(prefix, end)
	if err != nil {
		return err
	}
	defer itr.Close()

	first := int64(0)
	latest := int64(0)
	for ; itr.Valid(); itr.Next() {
		key := itr.Key()
		if len(key) != 9 { // prefix(1) + version(8)
			continue
		}
		v := int64(binary.BigEndian.Uint64(key[1:]))
		if first == 0 || v < first {
			first = v
		}
		if v > latest {
			latest = v
		}
	}
	if err := itr.Error(); err != nil {
		return err
	}

	ndb.mtx.Lock()
	ndb.firstVersion = first
	ndb.latestVersion = latest
	ndb.mtx.Unlock()
	return nil
}

// --- Version readers ---

func (ndb *nodeDB) incrVersionReaders(version int64) {
	// Take pruneMu FIRST so that registrations block while a prune holds
	// it. This closes the check-vs-register TOCTOU window: prune verifies
	// reader counts under pruneMu and keeps it held for the duration of
	// the prune, so any new reader attempting to register waits until
	// prune completes — at which point the version will no longer exist
	// in the DB and subsequent callers will observe that naturally.
	ndb.pruneMu.Lock()
	defer ndb.pruneMu.Unlock()
	ndb.mtx.Lock()
	defer ndb.mtx.Unlock()
	ndb.versionReaders[version]++
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

// beginPruning atomically verifies that no version in [first, to] has
// active readers and claims pruneMu for the duration of the prune. If any
// version has readers, it returns an ErrActiveReaders wrapping the first
// such version and releases the lock.
//
// Callers MUST call endPruning when done to release pruneMu; the typical
// pattern is `defer ndb.endPruning()` immediately after a successful
// beginPruning call. See Finding #15.
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

// endPruning releases pruneMu acquired by a prior beginPruning call.
func (ndb *nodeDB) endPruning() {
	ndb.pruneMu.Unlock()
}

// --- Commit coordination ---

func (ndb *nodeDB) SetCommitting() {
	ndb.mtx.Lock()
	defer ndb.mtx.Unlock()
	ndb.isCommitting = true
}

func (ndb *nodeDB) UnsetCommitting() {
	ndb.mtx.Lock()
	defer ndb.mtx.Unlock()
	ndb.isCommitting = false
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
	ndb.batch.Close()
	ndb.batch = ndb.db.NewBatch()
	return err
}

// ResetNonce resets the per-version nonce counter.
func (ndb *nodeDB) ResetNonce() {
	ndb.nextNonce = 0
}

// NextNodeKey returns a new NodeKey for the given version with an
// auto-incrementing nonce.
func (ndb *nodeDB) NextNodeKey(version int64) *NodeKey {
	ndb.nextNonce++
	return NewNodeKey(version, ndb.nextNonce)
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

// DeleteVersion removes a root reference and all nodes for that version
// from the batch. Used by DeleteVersionsFrom.
func (ndb *nodeDB) DeleteRoot(version int64) error {
	return ndb.batch.Delete(rootDBKey(version))
}

// DeleteNode removes a node from the batch (used during pruning).
func (ndb *nodeDB) DeleteNode(nkBytes []byte) error {
	if ndb.nodeCache != nil {
		ndb.nodeCache.Remove(string(nkBytes))
	}
	return ndb.batch.Delete(nodeDBKey(nkBytes))
}
