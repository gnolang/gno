# Immutable B+ Tree — Drop-in Replacement for tm2/pkg/iavl

## Design Decisions

- **Branching factor**: B=32 (parameterized; B=64 also supported)
- **Hash**: Full SHA256 (32 bytes). ICS23 natively supports `HashOp_SHA256`.
- **Values**: Out-of-line. Leaf nodes store `key + value_hash(32B)`.
  Values stored separately, content-addressed by their SHA256 hash.
  Always written (no dedup `Has()` check — idempotent writes).
  Never garbage collected (dead values are harmless noise).
- **Node hash**: Mini merkle tree over ALL b slots (sentinel hash for empties).
  No version/height/size in hash — pure merkle, unlike IAVL.
  Incremental updates: in-memory `[2*B]Hash` binary tree per dirty node,
  recompute only the log₂(B)=5 path when a single slot changes.
  Mini merkle intermediate hashes are NOT serialized to disk — only
  childHashes (the leaf level) are persisted. Intermediates are
  recomputed on load when needed (~1.9μs per node, 31 SHA256 calls).
- **Proofs**: Custom ICS23 ProofSpec. log₂(b) sibling hashes per tree level.
  Maps to a uniform chain of binary merkle InnerOps — single InnerSpec.
- **Pruning**: Orphan-less, dual-tree-walk adapted for B+ tree fan-out.
- **Iteration**: Stack-based traversal for both ascending and descending.
  No leaf sibling pointers (avoids COW cascade on splits).
  Contract: no writes while iterating (matches IAVL).
- **Splits**: 90/10 split for append-only patterns (sequential keys).
  Standard 50/50 split for random inserts.
- **Package**: `tm2/pkg/bptree` + store wrapper at `tm2/pkg/store/bptree`.

## Why

| Metric (100M items, 100MB cache) | IAVL | B+32 |
|---|---|---|
| Disk reads per GET/proof | 9-15 | 1-2 |
| Proof size | ~1.2-1.7 KB | ~0.9-1.4 KB |
| 3.5 week history | ~1 TB | ~2-3 TB |
| Base tree size | ~26 GB | ~19 GB |

B+ tree trades ~2-3x more history storage for ~5-10x fewer disk reads.
Out-of-line values keep the history overhead manageable.

### Disk I/O Comparison (cold cache, single operation)

| Items | GET reads (IAVL / B+32) | SET reads | SET writes |
|-------|------------------------|-----------|------------|
| 100M  | 28 / 7                | 28 / 6    | 29 / 8     |
| 1G    | 31 / 8                | 31 / 7    | 32 / 9     |
| 10G   | 35 / 9                | 35 / 8    | 36 / 10    |

With 100MB cache, SET writes are unchanged (COW nodes always written).
Reads drop: IAVL 28→9, B+32 7→2.7 at 100M items.

## Node Layout

### InnerNode

```
// In-memory fields:
nodeKey     *NodeKey            // nil if unsaved (new/cloned)
numKeys     int16               // occupied separator keys (children = numKeys+1)
size        int64               // total leaf count in subtree
keys        [B-1][]byte         // separator keys: keys[i] = first key of child[i+1]
children    [B][]byte           // NodeKey references to children
childHashes [B][32]byte         // hash of each child (mini merkle leaf level)
height      int16               // levels above leaf (leaves=0, parent=1)
miniTree    [2*B][32]byte       // in-memory only; NOT serialized to disk
```

**On-disk format** (only childHashes serialized, not miniTree):
`type(1B) + numKeys(varint) + size(varint) + height(varint)`
`+ keys (varint-len-prefixed each) + childNodeKeys (12B each) + childHashes (32B each)`

On-disk size at 69% fill (eff_b=22):
`~8B header + 22 × 33B keys + 23 × 12B refs + 23 × 32B hashes ≈ 1,746B`

### LeafNode

```
// In-memory fields:
nodeKey     *NodeKey
numKeys     int16               // occupied slots
keys        [B][]byte           // sorted keys
valueHashes [B][32]byte         // SHA256 of each value
miniTree    [2*B][32]byte       // in-memory only; NOT serialized to disk
```

**On-disk format** (only valueHashes serialized, not miniTree):
`type(1B) + numKeys(varint) + keys (varint-len-prefixed each) + valueHashes (32B each)`

Leaf node on-disk size at 69% fill:
`~3B header + 22 × 33B keys + 22 × 32B hashes ≈ 1,433B`

Note: `size` is stored only in InnerNode. A leaf's size is simply `numKeys`.
`ImmutableTree.Size()` reads `root.size` — one node load from the root.

### Node Hash (Mini Merkle)

Each node's hash is the root of a binary merkle tree over all B slots.
Domain separation (RFC 6962) prevents second-preimage attacks:

- **Leaf slot hash**: `SHA256(0x00 || varint(len(key)) || key || 0x20 || SHA256(value))`
  for occupied slots, sentinel for empty slots.
- **Mini merkle inner hash**: `SHA256(0x01 || left || right)`

The leaf slot hash includes varint length prefixes to match ICS23's
`LengthOp_VAR_PROTO`. The `0x20` byte is `varint(32)` — the fixed
length of a SHA256 output. Without length prefixes, the boundary
between key and value hash would be ambiguous.

The `0x00` / `0x01` prefix bytes make it cryptographically impossible to
confuse a leaf hash with an inner hash, regardless of key length.

Inner node slots use `childHash[i]` directly (already a hash, no re-prefix).

**Sentinel hash**: `SHA256(0x02)` — a third domain separator guaranteeing
the sentinel can never collide with any `0x00`-prefixed (leaf) or
`0x01`-prefixed (inner) hash.

**Sentinel short-circuit**: When computing the mini merkle, if both
children of an inner node are the sentinel, the result is the sentinel
itself — NOT `SHA256(0x01 || sentinel || sentinel)`. This ensures that
empty subtrees at ALL depths have the same hash value (the sentinel).

This is critical for ICS23 non-membership proofs. ICS23's
`leftBranchesAreEmpty`/`rightBranchesAreEmpty` compare each sibling
against a single flat `EmptyChild` value. Without short-circuiting,
empty subtrees at depth > 0 would have different hashes than
`EmptyChild`, causing `IsLeftMost`/`IsRightMost` to fail.

Short-circuiting is cryptographically safe: SHA256 preimage resistance
means `SHA256(0x01 || x || y)` cannot equal `SHA256(0x02)` for any x, y.
So the sentinel is a "reserved" value that no real inner hash can produce.

The full binary tree is stored in `miniTree[2*B]` (heap-style array,
in-memory only — not serialized to disk). When a single slot changes,
only the log₂(B)=5 ancestor hashes are recomputed — 5 SHA256 calls
instead of 31. Recomputation applies the short-circuit rule at each level.

This gives log₂(B)=5 sibling hashes per level in proofs (for B=32).

## Caching

Single count-based LRU cache for tree nodes (inner + leaf), keyed by
NodeKey bytes. Default size: **10,000 entries** (matching IAVL's default).

At ~4.3KB per in-memory B+32 node (including the `miniTree` array),
10K entries ≈ **~43MB**. This is enough to hold all inner nodes for
trees up to ~200M items, giving 1-2 disk reads per GET (leaf + value).

No fast node cache needed — unlike IAVL, the B+ tree's shallow height
makes the tree itself fast enough without a parallel flat index.

For larger trees, increase the cache size or switch to a byte-weighted
LRU. The constructor accepts `cacheSize int` matching IAVL's API.

## Persistence Format

All stored in the underlying `dbm.DB`:

| Prefix | Key format | Value | Purpose |
|--------|-----------|-------|---------|
| `B` | `B<version:8><nonce:4>` | serialized node | Inner/leaf node storage |
| `V` | `V<hash:32>` | raw value bytes | Out-of-line value store |
| `R` | `R<version:8>` | NodeKey (12B) + hash (32B), or empty | Root reference per version |
| `M` | `M<string>` | string | Metadata (storage version, etc.) |

Value writes are always unconditional (no `Has()` check). Content-addressed
writes are idempotent — duplicate writes overwrite with identical bytes.

## Copy-on-Write

When `Set()` or `Remove()` is called:
1. Traverse root → leaf, cloning every node on the path (set `nodeKey = nil`).
2. Modify the cloned leaf (insert/update/delete).
3. If leaf overflows (B+1 entries), split and propagate separator up.
   If the root splits, create a new inner node as root (height increases).
4. If leaf underflows (<B/2 entries), merge or redistribute with sibling.
   The sibling is found via the parent (already on the COW path).
   The sibling itself must be cloned. If merging causes the parent to
   underflow, cascade upward — at each level, the parent is already on
   the COW path, only the sibling at that level needs cloning.
   If the root inner node is left with one child after a merge, replace
   the root with that child (height decreases).
5. Update `size` on every cloned inner node:
   - Insert (new key): +1
   - Update (existing key, new value): unchanged
   - Delete: -1
   `Set()` returns `(updated bool, error)` — `true` if key existed.

### Empty Tree

When the last key is removed, the root becomes nil. `Size()` returns 0.
`Hash()` returns nil. `IsEmpty()` returns true. An empty root reference
`R<version>` with empty value is persisted to record the version exists.

On `SaveVersion()`:
1. Walk new nodes (nodeKey == nil), assign `NodeKey{version, nonce++}`.
2. Compute hashes bottom-up (recompute mini merkle for each dirty node).
3. Write all new nodes + values to DB batch.
4. Write root reference `R<version>` = NodeKey(12B) + rootHash(32B).
   ALWAYS written, even if tree unchanged. Storing the root hash here
   avoids recomputing 31 SHA256 calls on every `LoadVersion` + `Hash()`.
5. Commit batch.
6. Snapshot the current tree as `lastSaved` for ImmutableTree.
7. Return `(rootHash, version, nil)`. rootHash is the mini merkle root
   of the root node, or nil for an empty tree.

Old versions still reference old (persisted) nodes until pruned.

### Rollback

`Rollback()` discards the entire working tree (all COW'd nodes with
`nodeKey == nil`) and reverts the root to `lastSaved`. The working
`size` counter also reverts. Any values already written to the value
store via out-of-line writes are orphaned but harmless (content-addressed,
never GC'd). The tree-level `workingSize` resets to `lastSaved.Size()`.

### Split Strategy

**Standard split (random inserts):** When a leaf overflows, split 50/50.
Left gets ⌈(B+1)/2⌉ = 17 keys, right gets 16 keys. Optimal for random
workloads where both halves receive future inserts.

**90/10 split (sequential/append pattern):** When the inserted key is
larger than all existing keys in the leaf (append pattern), split
asymmetrically: left gets B-1=31 keys, right gets 2 keys (the last
existing key + the new key). The separator promoted to the parent is
a copy of the first key of the right leaf (standard B+ tree convention:
all keys remain in leaves, separator is a copy).

Detection: if `insertKey > leaf.keys[leaf.numKeys-1]`, use 90/10.

The left side freezes at ~97% full. The right side grows until it too
splits. Without 90/10 split, sequential inserts produce ~50% fill
(each left half freezes at 16/32). With it, leaves stay ~97% full.
This is a well-known B+ tree optimization sometimes called
"rightmost split."

For **standard 50/50 split**: B+1=33 keys split as left=17, right=16.
Separator = copy of right[0]. Total keys in leaves: 17+16=33. ✓

### Size Tracking

Each inner node stores `size int64` — the total leaf count of its subtree.
Updated on every COW clone along the modification path (+1 for insert,
-1 for delete). A leaf's size is `numKeys` (no separate field needed).

`ImmutableTree.Size()` returns `root.size` — a single node load.

`ExpectedDepth()` in the store layer must change from `log₂(size)` to
`log₂(size) / log₂(B)` for B+ trees.

### WorkingHash

`WorkingHash()` computes the hash of the current unsaved working tree.
This requires computing mini merkle hashes for all dirty (nodeKey==nil)
nodes bottom-up. This is potentially expensive but is only called when
needed (e.g., between transactions within a block). `Hash()` returns
the last saved version's cached hash (cheap).

## Pruning (Dual Tree Walk)

`DeleteVersionsTo(toVersion)`: for each version V from first to toVersion:
1. Walk tree V and tree V+1 simultaneously (DFS).
2. At each inner node, compare child hashes pairwise by hash value
   (not by position — children may have shifted due to splits/merges).
3. If child hash matches in both trees → skip that subtree (shared).
4. If child hash exists in V but not in V+1 → entire subtree is orphaned,
   delete all its nodes recursively.
5. If child hash exists in V but has a different hash in V+1 → descend
   into both children to find deeper orphans.
6. Delete root reference for version V.

The key difference from IAVL's binary dual-walk: instead of comparing
left/right children positionally, we compare child hash *sets*. Build
a set of child hashes from V+1's inner node, then for each child hash
in V's inner node, check membership. This handles split/merge shifts.

No explicit orphan index needed. Cost is O(changed nodes per version).

### Async Pruning Coordination

`SetCommitting()` / `UnsetCommitting()` prevent concurrent pruning from
interfering with `SaveVersion()`. When `SetCommitting()` is called,
background pruning pauses. When `UnsetCommitting()` is called, it resumes.
This matches IAVL's coordination mechanism. The `nodeDB` tracks this
state and the background pruning goroutine checks it before deleting.

### Version Readers

The `nodeDB` tracks active readers per version via
`versionReaders map[int64]uint32`. This prevents pruning a version
while an iterator or exporter is reading it.

- `ImmutableTree.Iterator()` calls `nodeDB.incrVersionReaders(version)`
- `iterator.Close()` calls `nodeDB.decrVersionReaders(version)`
- `Export()` similarly increments; `Exporter.Close()` decrements
- `DeleteVersionsTo/From` returns an error if a version has active readers
- `Close()` is idempotent (uses a `closed bool` flag)

## ICS23 Proof Design

### ProofSpec

```go
var BptreeSpec = &ics23.ProofSpec{
    LeafSpec: &ics23.LeafOp{
        Prefix:       []byte{0x00},        // leaf domain separator (RFC 6962)
        PrehashKey:   ics23.HashOp_NO_HASH,
        PrehashValue: ics23.HashOp_SHA256,
        Hash:         ics23.HashOp_SHA256,
        Length:       ics23.LengthOp_VAR_PROTO,
    },
    InnerSpec: &ics23.InnerSpec{
        ChildOrder:      []int32{0, 1},    // binary merkle
        MinPrefixLength: 1,                // the 0x01 domain separator
        MaxPrefixLength: 1,
        ChildSize:       32,               // SHA256 output
        EmptyChild:      sentinelHash[:],  // SHA256(0x02)
        Hash:            ics23.HashOp_SHA256,
    },
}
```

The `0x01` prefix in InnerSpec corresponds to `SHA256(0x01 || left || right)`.
ICS23 InnerOp.Apply computes `hash(prefix || child || suffix)`, so:
- Left child (idx=0): prefix=`0x01` (1B), suffix=`right_sibling` (32B)
- Right child (idx=1): prefix=`0x01 || left_sibling` (33B), suffix=empty

### Leaf Hash (must match ICS23 LeafOp exactly)

The tree computes leaf slot hashes as:
```
SHA256(0x00 || varint(len(key)) || key || varint(32) || SHA256(value))
```

This matches what ICS23 `LeafOp.Apply(key, value)` produces with
`Prefix=0x00, PrehashKey=NO_HASH, PrehashValue=SHA256, Length=VAR_PROTO`:
1. `pkey = varint(len(key)) || key`
2. `pvalue = varint(32) || SHA256(value)`   (varint(32) = 0x20)
3. `hash = SHA256(0x00 || pkey || pvalue)`

If the tree's internal hash differs from this, every proof fails.

### Proof Structure

The mini merkle approach collapses the B+ tree into what ICS23 sees as
a binary merkle tree. Each proof is a uniform chain of binary InnerOps:

- **Per tree level**: log₂(B) = 5 InnerOps (mini merkle siblings)
- **Leaf**: 1 LeafOp producing the leaf slot hash (see above)
- **Total InnerOps**: 5 × tree_height

For B=32 with 100M items (height ~6): 30 InnerOps + 1 LeafOp.
IAVL with 100M items: ~27 InnerOps + 1 LeafOp. Comparable depth.

Each InnerOp (with 0x01 domain separator):
- Child is left (idx=0): `prefix = 0x01` (1B), `suffix = sibling_hash` (32B)
- Child is right (idx=1): `prefix = 0x01 || sibling_hash` (33B), `suffix = empty`

### Non-Membership Proofs

Simpler than IAVL. All data is in sorted leaves:
1. Find the leaf where the key would be.
2. The left neighbor = previous slot (or last slot of previous leaf).
3. The right neighbor = current slot (or first slot of next leaf).
4. Produce existence proofs for both adjacent keys.
5. ICS23 `NonExistenceProof` verifies adjacency via `IsLeftNeighbor`.

The `EmptyChild` sentinel (`SHA256(0x02)`) in InnerSpec is essential for
the `IsLeftMost`/`IsRightMost`/`leftBranchesAreEmpty`/`rightBranchesAreEmpty`
checks in adjacency verification.

### Store Integration

Requires changes to `tm2/pkg/store/types/ics23.go`:
- New constant: `ProofOpBptreeCommitment = "ics23:bptree"`
- New constructor: `NewBptreeCommitmentOp(key, proof)`
- Add `"ics23:bptree"` case to `CommitmentOpDecoder`
- Register in `DefaultProofRuntime`

IBC light clients must be upgraded to recognize `BptreeSpec` at the
migration block height.

## API Surface

Must satisfy the same interface as `tm2/pkg/iavl`:

### MutableTree
- `NewMutableTree(db, cacheSize, skipFastStorageUpgrade, logger, ...options)`
- `IsEmpty() bool`
- `Set(key, value) (bool, error)`
- `Get(key) ([]byte, error)`
- `Has(key) (bool, error)`
- `Remove(key) ([]byte, bool, error)`
- `SaveVersion() ([]byte, int64, error)`
- `Load() (int64, error)`
- `LoadVersion(version) (int64, error)`
- `LoadVersionForOverwriting(version) error`
- `DeleteVersionsTo(version) error`
- `DeleteVersionsFrom(version) error`
- `GetImmutable(version) (*ImmutableTree, error)`
- `GetVersioned(key, version) ([]byte, error)`
- `Rollback()`
- `Hash() []byte`
- `WorkingHash() []byte`
- `WorkingVersion() int64`
- `AvailableVersions() []int`
- `VersionExists(version) bool`
- `SetInitialVersion(version)`
- `SetCommitting()`
- `UnsetCommitting()`
- `Iterator(start, end, ascending) (dbm.Iterator, error)`
- `Iterate(fn func(key, value []byte) bool) (bool, error)`
- `Close() error`
- `Import(version) (*Importer, error)`
- `Export() (*Exporter, error)` (via embedded ImmutableTree)
- `GetMembershipProof(key) (*ics23.CommitmentProof, error)`
- `GetNonMembershipProof(key) (*ics23.CommitmentProof, error)`

### ImmutableTree
- `Get(key) ([]byte, error)`
- `Has(key) (bool, error)`
- `GetWithIndex(key) (int64, []byte, error)`
- `GetByIndex(index) (key, value []byte, err error)`
- `Iterator(start, end, ascending) (dbm.Iterator, error)`
- `Iterate(fn func(key, value []byte) bool) (bool, error)`
- `IterateRange(start, end, ascending, fn) bool`
- `IterateRangeInclusive(start, end, ascending, fn) bool`
- `Hash() []byte`
- `Version() int64`
- `Size() int64`
- `Height() int8`
- `Export() (*Exporter, error)`
- `GetMembershipProof(key) (*ics23.CommitmentProof, error)`
- `GetNonMembershipProof(key) (*ics23.CommitmentProof, error)`
- `VerifyMembership(proof, key) (bool, error)`
- `VerifyNonMembership(proof, key) (bool, error)`

### GetByIndex / GetWithIndex

These use per-inner-node `size` fields for O(height) positional access.
At each inner node, scan children's sizes to find which child contains
the target index, then descend. O(B) work per level × height levels.

### GetVersioned

Load the immutable tree at the given version via `GetImmutable(version)`,
then call `Get(key)` on it. The out-of-line value store is content-addressed
and never GC'd, so values from any undeleted version are always retrievable.

### LoadVersionForOverwriting

Load the target version, then call `DeleteVersionsFrom(targetVersion + 1)`
to delete all newer versions. Allows history rewriting (e.g., after a
chain halt and restart from an earlier height).

## Export / Import

### ExportNode

```go
type ExportNode struct {
    Key      []byte
    Value    []byte  // actual value, not hash (inlined for export)
    Version  int64   // version this leaf was last modified
    Height   int8    // 0 for leaf, >0 for inner
}
```

### Export

Depth-first post-order traversal (children before parent).
Leaf nodes export with `Height=0`, `Key`, `Value` (fetched from value store),
and `Version` (from nodeKey.version).
Inner nodes export with `Height>0`, `Key` (first key of subtree), and
`Version`. Values are inlined in the export stream so the importer
doesn't need access to the value store.

### Import

Reconstructs the tree from the export stream. Nodes arrive in post-order
(leaves first). Use a stack: push leaf nodes, and when an inner node
arrives, pop its children from the stack and wire them up. Assign
NodeKeys with the target import version. Batch-write to DB.

The import format is NOT compatible with IAVL export format. State sync
must negotiate the format based on which tree type the chain uses.

## Iteration

### Stack-Based Traversal

Both ascending and descending iteration use a stack of
`(innerNode, childIndex)` pairs representing the path from root to
the current leaf.

**Ascending**: When the current leaf is exhausted, pop the stack to
find the parent, advance childIndex by 1, and descend to the leftmost
leaf of that child. Amortized O(1) node loads per leaf transition.

**Descending**: When the current leaf is exhausted going backward,
pop the stack, decrement childIndex, and descend to the rightmost leaf
of that child. Same amortized O(1) cost.

Worst case per leaf boundary: O(height) node loads (when crossing a
subtree boundary at every level). This happens once per ~B^k leaves
at depth k, so it amortizes away.

### Iterator Contract

**No writes while iterating.** Matches IAVL's existing contract.
The store layer wraps iterators with `CacheWrap`, which batches writes,
so in practice writes don't happen during iteration.

If `Set()` is called while an iterator is open, the COW mechanism
means the iterator's stack references pre-mutation nodes. The iterator
would see a stale (but consistent) view. This is safe but undocumented
— the contract simply forbids it.

### Thread Safety

- `MutableTree` is NOT safe for concurrent use (same as IAVL).
- `ImmutableTree` is safe for concurrent reads.
- `nodeDB` handles its own locking via `sync.Mutex`.
- Version readers prevent pruning of versions with open iterators.

## File Structure

```
tm2/pkg/bptree/
  PLAN.md               ← this file

  # Phase 1: Core data structures
  const.go              — B, HashSize, MinKeys, sentinel hash, precomputed empties
  errors.go             — sentinel errors
  logger.go             — Logger interface
  options.go            — Options, functional opts
  node_key.go           — NodeKey (version + nonce), encoding
  hash.go               — SHA256 helpers, leaf slot hash with length prefixes
  node.go               — InnerNode (with size), LeafNode structs + serialization
  mini_merkle.go        — binary merkle over b slots, incremental updates, sibling path
  search.go             — binary search within node slot arrays

  # Phase 2: Tree operations (in-memory, no persistence)
  insert.go             — insert path with COW + split propagation
  split.go              — leaf and inner node split (50/50 + 90/10)
  remove.go             — remove path with COW + merge propagation
  merge.go              — merge and redistribute logic
  mutable_tree.go       — MutableTree: Set, Get, Has, Remove, Rollback (in-memory)
  immutable_tree.go     — ImmutableTree (skeleton)
  bptree_test.go        — basic correctness tests

  # Phase 3: Persistence
  nodedb.go             — persistence layer, key formats, LRU cache, batch,
                          version readers, SetCommitting/UnsetCommitting
  value_store.go        — out-of-line value storage (content-addressed, always-write)
  # Wire SaveVersion, LoadVersion, GetImmutable, Rollback into mutable_tree.go
  export.go             — Exporter (depth-first post-order traversal)
  import.go             — Importer (reconstruct tree from stream)

  # Phase 4: Iteration
  iterator.go           — stack-based traversal for both ascending and descending

  # Phase 5: Proofs
  proof_spec.go         — BptreeSpec for ICS23 (with sentinel = SHA256(0x02))
  proof.go              — GetMembershipProof, GetNonMembershipProof
  proof_test.go         — proof verification round-trip tests

  # Phase 6: Pruning
  prune.go              — adapted dual-tree-walk for B+ fan-out,
                          async pruning with SetCommitting coordination

  # Phase 7: Store integration
  tm2/pkg/store/bptree/
    store.go            — CommitStore wrapping bptree.MutableTree
    tree.go             — Tree interface adapter (own interface, not iavl.Tree)
```

## Implementation Order

### Phase 1 — Core data structures
Get the foundational types compiling and tested in isolation.
- const, errors, logger, options
- NodeKey encoding
- SHA256 hash helpers, leaf slot hash with varint length prefixes
- InnerNode (with `size` field) / LeafNode structs with Serialize/Deserialize
- Mini merkle: full binary tree layout (in-memory), incremental slot update, sibling path
- Binary search within sorted node keys
- Sentinel = SHA256(0x02) with short-circuit rule (empty subtrees at all
  depths equal the sentinel — no recursive hashing of empty pairs)

### Phase 2 — Tree operations
Get Set/Get/Remove working with in-memory nodes (no DB).
- Insert into leaf, split leaf (50/50 + 90/10), propagate split up
- Remove from leaf, merge/redistribute, propagate down
- MutableTree wrapping a root node, COW path cloning
- Size tracking per inner node (+1/-1 on COW path)
- Rollback: discard COW nodes, revert root to lastSaved
- Test: random insert/delete/get sequences, verify ordering
- Test: sequential inserts, verify 90/10 split produces high fill factor
- Test: GetByIndex / GetWithIndex using per-node sizes

### Phase 3 — Persistence
Wire the tree to a real DB via nodeDB.
- Node serialization to/from bytes (use sync.Pool for byte buffers only,
  NOT for node structs — COW'd nodes may still be referenced by nodeDB
  cache or iterator stacks)
- DB key formats (B, V, R, M prefixes)
- SaveVersion: assign NodeKeys, compute hashes, batch write, snapshot lastSaved
- LoadVersion / LoadVersionForOverwriting
- Value store: always-write, content-addressed by SHA256
- Version reader tracking (incrVersionReaders / decrVersionReaders)
- SetCommitting / UnsetCommitting for async pruning coordination
- Export/Import (post-order traversal, values inlined)
- Test: save version, reload, verify state
- Test: rollback after mutations, verify revert

### Phase 4 — Iteration
- Ascending: descend to start leaf, use stack to walk forward through tree
- Descending: descend to end leaf, use stack to walk backward through tree
- Both directions use a stack of (innerNode, childIndex) pairs
- Iterator.Close() decrements version readers (idempotent via closed flag)
- IterateRange, IterateRangeInclusive callbacks
- Test: range queries, empty ranges, full scans, single-element ranges

### Phase 5 — Proofs
- Define BptreeSpec for ICS23 (binary InnerSpec, SHA256, EmptyChild=SHA256(0x02))
- Leaf hash: SHA256(0x00 || varint(len(key)) || key || 0x20 || SHA256(value))
- Build existence proofs: leaf mini-merkle path + inner mini-merkle paths
- Build non-existence proofs: bracket with adjacent keys
- Register new proof type in store layer
- Test: generate proof, verify against root hash
- Test: non-membership for keys before/after/between existing keys

### Phase 6 — Pruning
- Adapted dual-tree-walk using child hash set comparison
- Delete orphaned nodes and root references
- Handle split/merge position shifts between versions
- Async pruning with SetCommitting coordination
- Check version readers before deleting
- Test: save multiple versions, prune, verify remaining versions intact
- Test: prune after splits and merges
- Test: concurrent prune + iterator (version readers block pruning)

### Phase 7 — Store integration
- Create tm2/pkg/store/bptree/ with own Tree interface
- CommitStore wrapping bptree.MutableTree
- Update ExpectedDepth to use log₂(size)/log₂(B)
- Register ProofOpBptreeCommitment in CommitmentOpDecoder + DefaultProofRuntime
- Wire into gno.land app as alternative store constructor
- Integration tests with full gnoland app

## Resolved Questions

1. **Leaf sibling pointers**: NO. Stack-based traversal for both directions.
   Avoids COW cascade when a split requires updating the left neighbor.

2. **Value garbage collection**: NONE. Content-addressed values are never
   deleted. Dead values after pruning are harmless — bounded by total
   unique values ever written. No reference counting needed.

3. **Parameterizing B**: `const B = 32` with fixed arrays for MVP.
   B=64 support via build tags or code generation later.

4. **ICS23 hash**: Full SHA256 (32 bytes). Natively supported by ICS23
   as `HashOp_SHA256`. No truncation, no custom hash ops, no fork needed.

5. **Value dedup**: Always write unconditionally. No `Has()` check.
   Content-addressed writes are idempotent.

6. **Sentinel hash**: `SHA256(0x02)` — provably distinct from any
   `0x00`-prefixed (leaf) or `0x01`-prefixed (inner) hash.

7. **Leaf hash format**: Must include varint length prefixes to match
   ICS23 `LengthOp_VAR_PROTO` exactly. Without this, proofs fail.

8. **Iterator contract**: No writes while iterating. Matches IAVL.

9. **Version readers**: Track active readers per version. Pruning
   refuses to delete versions with active readers.

10. **Sentinel short-circuit**: Empty subtrees at all depths equal the
    sentinel (SHA256(0x02)). Do not compute SHA256(0x01 || sentinel ||
    sentinel). This is required for ICS23 EmptyChild to work at all
    mini merkle depths in non-membership proofs.

11. **miniTree not serialized**: The `[2*B][32]byte` intermediate mini
    merkle hashes are in-memory only. On-disk nodes store only
    childHashes/valueHashes (the leaf level). Intermediates are
    recomputed on demand. sync.Pool is for serialization byte buffers
    only — not node structs (which may be referenced by cache/iterators).

12. **Insert vs update**: `Set()` distinguishes new keys (size +1) from
    existing keys (size unchanged). Returns `(updated bool, error)`.

13. **Root creation/collapse**: Splitting the root creates a new inner
    root (height +1). Merging the root's last two children collapses
    the root to the remaining child (height -1). Removing the last key
    sets root to nil.

14. **SaveVersion always writes root ref**: Even with no mutations,
    `R<version>` is written to record the version exists.
