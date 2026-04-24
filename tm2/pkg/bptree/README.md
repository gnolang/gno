# bptree — Immutable B+ Tree for Gno

A drop-in replacement for `tm2/pkg/iavl`, implementing a versioned,
persistent B+ tree with ICS23 merkle proofs.

## Why replace IAVL?

IAVL is a binary AVL+ tree. Every read traverses ~28 levels for 100M
items. Every write COW-clones ~28 nodes. The B+ tree with branching
factor 32 reduces this to ~6 levels — fewer disk reads, fewer writes,
fewer round-trips to the underlying KV store.

| Metric (100M items, 10K cache) | IAVL | B+32 |
|---|---|---|
| GET disk reads | 15 | 3 |
| SET disk writes | 28 | 7 |
| SET total ops | 43 | 10 |
| Proof size | ~1.4 KB | ~1.0 KB |

See [PERFORMANCE.md](PERFORMANCE.md) for full analysis across tree
sizes, cache configurations, and batched workloads.

## Design choices

### Branching factor B=32

A compile-time constant. Each inner node holds up to 31 separator keys
and 32 child references. Each leaf holds up to 32 key-value pairs.
The value 32 balances:
- **Read performance**: tree height is log₃₂(n) ≈ 6 for 100M items
- **Write amplification**: each COW'd node is ~1.7KB (vs IAVL's ~100B)
- **Proof size**: log₂(32)=5 mini-merkle siblings per tree level
- **Cache efficiency**: 10K nodes × ~4.3KB ≈ 43MB caches all inner nodes up to ~200M items

### Out-of-line values

Leaf nodes store `key + SHA256(value) + ValueKey(12B)`, not the raw value.
Values are stored separately in the DB under their ValueKey (a
per-allocation `(version, nonce)` identifier), with the SHA256(value) kept
in the leaf for Merkle proofs. This means:
- **Smaller nodes**: leaf nodes are ~1.4KB regardless of value size
- **Less COW amplification**: modifying one key copies only hash references, not sibling values
- **Copy safety**: callers cannot corrupt stored values by mutating their byte slices

Values are written directly to the DB (not via batch) so `Get` works
before `SaveVersion`. Values are never garbage collected — dead values
after pruning are harmless noise. Eagerly-written values left behind by
a crashed working session (process died before `SaveVersion` or
`Rollback`) are cleaned up on the next `Load()` by scanning for
ValueKeys with a version greater than the latest persisted version.

**No content-addressed deduplication.** Two `Set` calls with identical
values each get a fresh ValueKey and a separate DB entry — there is no
hash-based lookup table. This keeps the design simple (no reference
counting, no "unreferenced hash" cleanup) at the cost of storage
overhead when identical values are common. If your workload has highly
duplicated values, dedupe at the application layer before `Set`.

### Full SHA256 (32 bytes)

No hash truncation. ICS23 natively supports `HashOp_SHA256`. This
avoids any need to fork the ICS23 library or define custom hash ops.
Proof sizes are slightly larger than a 20-byte hash would give, but
the simplicity and ICS23 compatibility are worth it.

### Mini merkle tree within each node

Each node's hash is the root of a binary merkle tree over all B=32
slots. This collapses the B+ tree into what ICS23 sees as a uniform
chain of binary `InnerOp`s — a single `InnerSpec` works for all levels.

Domain separation follows RFC 6962:
- **Leaf slot**: `SHA256(0x00 || varint(len(key)) || key || 0x20 || SHA256(value))`
- **Inner**: `SHA256(0x01 || left || right)`
- **Sentinel**: `SHA256(0x02)` for empty slots

The sentinel uses a **short-circuit rule**: if both children are the
sentinel, the result is the sentinel (not `SHA256(0x01 || sentinel || sentinel)`).
This ensures `EmptyChild` in the ICS23 spec works correctly at all
mini-merkle depths for non-membership proofs.

The leaf slot hash includes varint length prefixes to match ICS23's
`LengthOp_VAR_PROTO` exactly. Without this, proofs fail verification.

Mini merkle intermediates are **not serialized to disk** — only the
leaf-level hashes (childHashes/valueHashes) are persisted. The full
binary tree is cached in memory (`miniTree [2*B][32]byte`) and
recomputed on load (~31 SHA256 calls per node, ~1.9μs). Incremental
updates via `SetSlot` recompute only 5 hashes per slot change.

### No leaf sibling pointers

B+ trees traditionally link leaves for fast range scans. In an
immutable/COW tree, updating a sibling pointer on split requires
cloning the neighbor leaf and its entire ancestor path — a cascade
that doubles the write cost of every split.

Instead, both ascending and descending iteration use a **stack-based
traversal**: a stack of `(innerNode, childIndex)` pairs. When a leaf
is exhausted, pop the stack to find the parent, advance (or retreat)
the child index, and descend to the next leaf. The amortized cost is
O(1) node loads per leaf transition — the same as sibling pointers.

### 90/10 split for sequential keys

When the inserted key is larger than all existing keys in the leaf
(append pattern), the leaf splits asymmetrically: left gets B-1=31
keys, right gets 2 keys. Without this, sequential inserts produce
~50% fill (each left half freezes at 16/32). With 90/10 splits,
leaves stay ~97% full. Detection: `insertPos == B`.

For random inserts, the standard 50/50 split is used.

### Copy-on-write versioning

Every `Set`/`Remove` clones the root-to-leaf path. Unchanged subtrees
are shared between versions. `SaveVersion` assigns `NodeKey(version, nonce)`
to each dirty node and batch-writes them to the DB. The root reference
`R<version>` stores the root's NodeKey + hash (44 bytes).

If a version already exists with a different hash, `SaveVersion` returns
an error (matching IAVL behavior — prevents accidental overwrites).

`Rollback` discards COW'd nodes and reverts to the last saved snapshot.

### Pruning via dual-tree-walk

No orphan index. To prune version V, walk V's tree and V+1's tree
simultaneously. At each inner node, compare child hash **sets** (not
positions — children may shift due to splits/merges). Matching hashes
mean shared subtrees — skip. Unmatched hashes mean orphaned nodes —
delete recursively. Cost: O(changed nodes per version).

### Lazy node loading

`LoadVersion` loads only the root node from the DB. Children are loaded
on demand by `getChild`, which checks `childNodes[idx]`, and if nil,
loads from DB via the `ndb` reference stored on each `InnerNode`. The
10K-node LRU cache prevents repeated DB hits for hot inner nodes.

### ICS23 proof system

The `BptreeSpec` defines:
- `LeafOp`: prefix `0x00`, `PrehashValue=SHA256`, `Length=VAR_PROTO`
- `InnerSpec`: `ChildOrder=[0,1]`, prefix length 1 (`0x01`), `EmptyChild=SHA256(0x02)`
- `MinDepth=5` (at least one mini-merkle traversal), `MaxDepth=60`

Membership proofs: collect the path from root to leaf, then for each
node emit log₂(B)=5 `InnerOp`s from the mini merkle sibling path.
Total: 5 × tree_height InnerOps + 1 LeafOp.

Non-membership proofs: find the two adjacent keys bracketing the
missing key, produce existence proofs for both. ICS23's `IsLeftNeighbor`
verifies adjacency using the `EmptyChild` sentinel for boundary checks.

### Nil values rejected

`Set(key, nil)` returns an error, matching IAVL behavior. Use
`[]byte{}` for empty values.

### Store integration

`tm2/pkg/store/bptree` provides a `CommitStore` wrapper that satisfies
the same interfaces as `tm2/pkg/store/iavl`:
- `types.Store`, `types.CommitStore`, `types.Queryable`, `types.DepthEstimator`
- `StoreConstructor` is a drop-in replacement for `iavl.StoreConstructor`

The bptree proof decoder is registered alongside IAVL and simple merkle
in `DefaultProofRuntime` — both proof types coexist.

## Package structure

```
tm2/pkg/bptree/
  const.go           B=32, HashSize=32, domain separators, sentinel
  errors.go          Sentinel errors (ErrVersionDoesNotExist, etc.)
  logger.go          Logger interface and NopLogger
  options.go         Options struct and functional option constructors
  hash.go            SHA256 helpers, leaf slot hash, inner hash with short-circuit
  mini_merkle.go     Binary merkle over B slots, incremental updates, sibling path
  node.go            InnerNode, LeafNode, serialization, lazy getChild
  node_key.go        NodeKey (version + nonce), encoding
  search.go          Binary search within sorted node arrays
  insert.go          Insert with COW, split propagation
  split.go           50/50 and 90/10 leaf/inner splits
  remove.go          Remove with COW, merge/redistribute
  mutable_tree.go    MutableTree: Set, Get, Remove, SaveVersion, LoadVersion, ...
  immutable_tree.go  ImmutableTree: read-only snapshot with value resolution
  iterator.go        Stack-based ascending/descending iterator
  nodedb.go          DB persistence, LRU cache, version tracking, batch writes
  export.go          Post-order tree export with value inlining
  import.go          Tree reconstruction from export stream
  proof_spec.go      BptreeSpec for ICS23
  proof.go           Membership and non-membership proof generation/verification
  prune.go           Dual-tree-walk pruning

tm2/pkg/store/bptree/
  store.go           CommitStore wrapper, Query, Iterator, proof integration
  tree.go            Tree interface adapters (mutable/immutable)
```

## Differences from IAVL

| Aspect | IAVL | B+32 |
|--------|------|------|
| Structure | Binary AVL+ | B+ tree, B=32 |
| Height (100M items) | ~28 | ~6 |
| Node size | ~100B | ~1.7KB |
| Value storage | Inline in leaf | Out-of-line by hash |
| Fast node index | Yes (separate KV index) | No (tree is fast enough) |
| Proof hash includes | height, size, version | Nothing (pure merkle) |
| Proof type | `ics23:iavl` | `ics23:bptree` |
| Leaf iteration | Goroutine + channel | Synchronous stack |
| Orphan tracking | Explicit orphan index | Dual-tree-walk (no index) |
| Node loading | Eager (full tree) | Lazy (on demand) |
| Copy semantics | Values shared by reference | Values copied (content-addressed) |

## Testing

314 tests covering:
- 202 B+ tree specific tests (internals, edge cases, golden vectors)
- 112 ported IAVL behavioral tests (identical function names)

```
go test ./tm2/pkg/bptree/ ./tm2/pkg/store/bptree/
```
