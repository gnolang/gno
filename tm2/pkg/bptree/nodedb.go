package bptree

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"sync"

	lru "github.com/hashicorp/golang-lru/v2"
	"golang.org/x/sync/singleflight"

	dbm "github.com/gnolang/gno/tm2/pkg/db"
)

// nodeDB handles persistence: reading/writing nodes, values, and root
// references to the underlying key-value store.
//
// Lock discipline:
//
//   - `mtx` guards `latestVersion`, `firstVersion`, and `versionReaders`.
//   - `pruneMu` serialises prune with reader registration (see
//     Findings #15 and #40). Acquired BEFORE `mtx` on any path that
//     takes both.
//   - `nodeCache` is internally thread-safe (hashicorp/lru).
//   - `batch`, `nextNonce`, and `pendingEvicts` are NOT guarded by a
//     lock. They are only mutated by the single writer goroutine
//     (SaveVersion / PruneVersionsTo). A future `AsyncPruning` option
//     that introduces a second writer goroutine MUST add a save-path
//     mutex covering these fields before landing. See Findings #13
//     and #42.
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
	// beginPruning (exclusive) and by incrVersionReaders (shared). See
	// Finding #15.
	pruneMu sync.RWMutex

	logger Logger

	nextNonce uint32 // per-SaveVersion nonce counter

	// pendingEvicts accumulates node cache keys for nodes queued for
	// deletion in the current batch. Evictions are deferred to Commit
	// (after the batch write succeeds) so a reader racing a DeleteNode
	// cannot cache-miss, reload from the still-populated DB, and then
	// re-cache a node whose deletion has been flushed. See Finding #44.
	// The prune-reader invariant makes this theoretical today (pruned
	// versions have no readers), but deferring the eviction codifies
	// the contract instead of relying on a cross-file invariant.
	pendingEvicts []string

	// loadGroup deduplicates concurrent cache-miss GetNode calls for
	// the same NodeKey. Without it, two readers that miss the cache
	// on the same key both do the DB read + deserialisation +
	// mini-merkle rebuild and both Add to the cache; the second
	// overwrites the first, leaving each reader holding a distinct
	// in-memory instance. With singleflight only one goroutine does
	// the work; the rest wait and share the result, keeping
	// per-NodeKey instances coherent across readers and halving DB
	// reads under hot-cold iteration workloads. See Finding #3.2.
	loadGroup singleflight.Group
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

// saveBufPool holds a pool of *bytes.Buffer used by SaveNode to
// serialize each node. Reusing buffers across saves (arena-style)
// eliminates the grow-then-discard churn a fresh bytes.Buffer incurs
// per call. Buffers over saveBufCapCap (see below) are dropped on Put
// rather than returned, so a single oversize node does not inflate
// every pooled buffer.
var saveBufPool = sync.Pool{
	New: func() any {
		b := bytes.NewBuffer(make([]byte, 0, 512))
		return b
	},
}

// saveBufCapCap caps the retained pool buffer capacity so a single
// abnormally large node doesn't inflate every pooled buffer. 8 KiB
// fits any realistic inner/leaf node (full inner: ~2.4 KiB;
// full leaf: ~2 KiB of fixed-width + variable key bytes).
const saveBufCapCap = 8 << 10

// SaveNode writes a node to the batch and adds it to the cache.
func (ndb *nodeDB) SaveNode(node Node) error {
	nk := node.GetNodeKey()
	if nk == nil {
		return ErrNodeMissingNodeKey
	}
	nkBytes := nk.GetKey()

	buf := saveBufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer func() {
		if buf.Cap() <= saveBufCapCap {
			saveBufPool.Put(buf)
		}
	}()
	switch n := node.(type) {
	case *InnerNode:
		if err := n.Serialize(buf); err != nil {
			return fmt.Errorf("serializing inner node: %w", err)
		}
	case *LeafNode:
		if err := n.Serialize(buf); err != nil {
			return fmt.Errorf("serializing leaf node: %w", err)
		}
	default:
		return fmt.Errorf("unknown node type")
	}

	// batch.Set retains the value slice (see tm2/pkg/db memdb/goleveldb
	// implementations). The pooled buffer is about to be reused, so
	// copy the bytes into a fresh allocation whose lifetime matches
	// the batch entry.
	data := make([]byte, buf.Len())
	copy(data, buf.Bytes())
	if err := ndb.batch.Set(nodeDBKey(nkBytes), data); err != nil {
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
//
// Concurrency: concurrent cache-miss readers share a single DB load
// via singleflight, keyed on the serialised NodeKey. The returned
// Node pointer is the same instance across all callers for a given
// key, so lazy-loaded children populated by one reader via getChild
// are observed by all.
func (ndb *nodeDB) GetNode(nkBytes []byte) (Node, error) {
	key := string(nkBytes)

	// Fast path: cache hit.
	if ndb.nodeCache != nil {
		if node, ok := ndb.nodeCache.Get(key); ok {
			return node, nil
		}
	}

	// Slow path: singleflight the DB load + deserialisation so
	// concurrent cache-miss readers share one instance.
	v, err, _ := ndb.loadGroup.Do(key, func() (any, error) {
		// Re-check the cache inside the singleflight in case another
		// caller populated it while we were waiting on the lock.
		if ndb.nodeCache != nil {
			if node, ok := ndb.nodeCache.Get(key); ok {
				return node, nil
			}
		}
		// A raw IO error here is unrecoverable — the tree depends on
		// reads succeeding to maintain its shape invariant.
		data, err := ndb.db.Get(nodeDBKey(nkBytes))
		if err != nil {
			panic(fmt.Sprintf("bptree: db get node %x: %v", nkBytes, err))
		}
		if data == nil {
			return nil, ErrNodeNotFound
		}
		nk := GetNodeKey(nkBytes)
		node, err := ReadNode(nk, data)
		if err != nil {
			// On-disk corruption — the tree's invariants are already
			// violated; panic rather than return garbage nodes.
			panic(fmt.Sprintf("bptree: deserializing node %x: %v", nkBytes, err))
		}
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

// SaveValue writes a value to the DB directly (not via batch), keyed by
// ValueKey. Writing early allows Get to observe freshly-Set values inside
// the same working session before SaveVersion flushes the surrounding
// node updates.
//
// Crash-recovery tradeoff: on a process crash between a Set and the
// subsequent SaveVersion's Commit, the in-memory sessionValues list is
// lost and the eagerly-written values become orphans on disk — no tree
// in any persisted version references them. The orphan is a pure space
// leak, never a correctness issue: ValueKeys embed the working version,
// whose nonce namespace LoadVersion seeds past any persisted slot via
// maxValueNonceForVersion, so a stale Set can never overwrite a live
// value. Rollback cleans them up in-process; a future restart-time
// scrub could walk PrefixVal entries whose version exceeds the
// persisted latestVersion to reclaim orphaned-on-crash values.
func (ndb *nodeDB) SaveValue(value, vk []byte) error {
	key := valueDBKey(vk)
	valCopy := make([]byte, len(value))
	copy(valCopy, value)
	return ndb.db.Set(key, valCopy)
}

// GetValue loads a value by its ValueKey from the DB.
//
// Return semantics:
//   - Present-and-empty values round-trip as a non-nil zero-length slice
//     (verified on both memdb and goleveldb backends) with err == nil.
//   - A missing record returns (nil, ErrValueMissing). The tree layer
//     should never see this in a healthy DB — every leaf.valueKeys[i]
//     references a ValueKey that SaveValue persisted — so an
//     ErrValueMissing return signals out-of-band corruption (e.g. the
//     value file was deleted externally) rather than a legitimate
//     Get miss. Callers should propagate the error rather than treat
//     it as a missing key. See ajnavarro PR #5571 review.
func (ndb *nodeDB) GetValue(vk []byte) ([]byte, error) {
	data, err := ndb.db.Get(valueDBKey(vk))
	if err != nil {
		return nil, fmt.Errorf("db get value: %w", err)
	}
	if data == nil {
		return nil, fmt.Errorf("%w: %x", ErrValueMissing, vk)
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

// maxValueNonceForVersion returns the largest nonce currently persisted
// for any ValueKey in the given version's namespace. Returns 0 when no
// ValueKey exists under that version.
//
// Used by LoadVersion to seed MutableTree.nextValueNonce past any
// on-disk slot in the working-version namespace, preventing SaveValue
// collisions after loading a non-latest version.
//
// Cost: one DB reverse-seek. On goleveldb/pebble this is O(log N); on
// memdb the current implementation materialises a sorted key list, so
// the scan is O(N log N) in DB size — acceptable because LoadVersion
// runs at most once per process lifecycle in production, and memdb is
// a test-only backend.
func (ndb *nodeDB) maxValueNonceForVersion(version int64) (uint32, error) {
	// math.MaxInt64 as the working version would wrap on the upper
	// bound; such a value is structurally unreachable (every
	// SaveVersion advances by 1 from 0) but guarding here makes the
	// helper safe for any int64 input.
	if version == math.MaxInt64 {
		return 0, fmt.Errorf("version %d out of range for nonce scan", version)
	}
	start := valueDBKey(encodeNodeKeyBytes(version, 0))
	end := valueDBKey(encodeNodeKeyBytes(version+1, 0))
	itr, err := ndb.db.ReverseIterator(start, end)
	if err != nil {
		return 0, err
	}
	defer itr.Close()
	if !itr.Valid() {
		return 0, nil
	}
	k := itr.Key()
	if len(k) != 1+NodeKeySize {
		return 0, nil
	}
	nk := GetNodeKey(k[1:])
	if nk == nil || nk.Version != version {
		return 0, nil
	}
	return nk.Nonce, nil
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
	if len(data) == 0 {
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
	// Take pruneMu (shared) FIRST so that registrations block while a
	// prune holds it exclusively. This closes the check-vs-register
	// TOCTOU window: prune verifies reader counts under the exclusive
	// lock and keeps it held for the duration of the prune, so any new
	// reader attempting to register waits until prune completes — at
	// which point the version will no longer exist in the DB and
	// subsequent callers will observe that naturally. Concurrent
	// registrations do not block each other. See Finding #15.
	ndb.pruneMu.RLock()
	defer ndb.pruneMu.RUnlock()
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

// endPruning releases the exclusive pruneMu acquired by a prior
// beginPruning call.
func (ndb *nodeDB) endPruning() {
	ndb.pruneMu.Unlock()
}

// --- Batch operations ---

// Commit flushes the current batch to disk and creates a new batch.
// Always closes the old batch and creates a new one, even on error,
// to avoid leaving the nodeDB in a broken state.
//
// On write error the pending batch is lost: most backends treat a batch
// write as atomic, so either every queued write landed on disk or none
// did. The caller cannot retry the same writes through this nodeDB; any
// state that was recorded in the batch (in particular: a half-built
// prune or save-version pass) must be reconstructed or the tree rolled
// back to a consistent checkpoint. See Finding #38.
//
// Pending cache evictions recorded by DeleteNode are applied only if
// the batch write succeeds. On error they are discarded because the
// deletions never reached disk (Finding #44).
func (ndb *nodeDB) Commit() error {
	var err error
	if ndb.opts.Sync {
		err = ndb.batch.WriteSync()
	} else {
		err = ndb.batch.Write()
	}
	ndb.batch.Close()
	ndb.batch = ndb.db.NewBatch()
	if err == nil {
		ndb.flushPendingEvicts()
	} else {
		ndb.pendingEvicts = ndb.pendingEvicts[:0]
	}
	return err
}

// discardBatch closes the current pending batch without writing it and
// starts a fresh one. Any cache evictions queued for the discarded
// batch are dropped since the underlying deletes never landed on disk.
// Callers that hit an error partway through queueing writes use this
// to drop the half-built batch so subsequent operations do not silently
// flush inconsistent state on the next Commit. See Findings #44 and #52.
func (ndb *nodeDB) discardBatch() {
	if ndb.batch != nil {
		ndb.batch.Close()
	}
	ndb.batch = ndb.db.NewBatch()
	ndb.pendingEvicts = ndb.pendingEvicts[:0]
}

// flushPendingEvicts applies deferred cache evictions accumulated by
// DeleteNode. Must only be called after the batch has been written
// successfully. See Finding #44.
func (ndb *nodeDB) flushPendingEvicts() {
	if ndb.nodeCache != nil {
		for _, k := range ndb.pendingEvicts {
			ndb.nodeCache.Remove(k)
		}
	}
	ndb.pendingEvicts = ndb.pendingEvicts[:0]
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

// DeleteRoot removes the root reference for the given version from the
// batch. It does NOT touch the nodes that were reachable from that root
// — the caller (PruneVersionsTo via the mark-and-sweep pass) is
// responsible for deleting any unreachable nodes separately.
func (ndb *nodeDB) DeleteRoot(version int64) error {
	return ndb.batch.Delete(rootDBKey(version))
}

// DeleteNode queues a node deletion on the current batch and records
// a pending cache eviction that Commit will apply if the batch write
// succeeds. Deferring the eviction closes a theoretical race where a
// cache-miss reader could reload the node after eviction but before
// the batch flushes, then re-cache a node whose deletion is now on
// disk. See Finding #44.
func (ndb *nodeDB) DeleteNode(nkBytes []byte) error {
	if err := ndb.batch.Delete(nodeDBKey(nkBytes)); err != nil {
		return err
	}
	if ndb.nodeCache != nil {
		ndb.pendingEvicts = append(ndb.pendingEvicts, string(nkBytes))
	}
	return nil
}
