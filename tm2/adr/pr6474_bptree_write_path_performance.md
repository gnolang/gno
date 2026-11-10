# PR6474: B+32 write-path performance — serialization, mini merkle, node layout, prefixed leaf keys

## Status

Proposed.

## Context

[PR #5571](https://github.com/gnolang/gno/pull/5571) (stacked on the also-closed
[#5570](https://github.com/gnolang/gno/pull/5570)) proposed a large B+32
performance pass against `feat/alex/bp32tree-advanced`. Neither merged. What
landed on `master` instead was #5438, followed by #5877, #5937, #6018 and
#6383, so the tree those PRs optimised and the tree in the repo have diverged
substantially: `master` now has record checksums, a persistent latest-version
fast index, snapshot-isolated query views, a non-memoising `getChild`, and a
different prune.

This ADR records which of #5571's ideas still apply to that codebase, which
were superseded or are actively wrong here, and what replaced them. The work
was profile-driven rather than a port: a CPU and allocation profile of a
100-write block against a 100k-key tree put ~50% of CPU in the garbage
collector and attributed allocation as follows.

| source | % of allocated bytes | % of allocated objects |
|---|---|---|
| `InnerNode.Clone` | 56.2 | 3.6 |
| `LeafNode.Clone` | 15.7 | 1.3 |
| `bytes.growSlice` (SaveNode's `bytes.Buffer`) | 12.4 | 6.2 |
| `stampChecksum` | 5.0 | 2.4 |
| `walkAndPrune` (`map[string]bool` per node) | 4.1 | 18.1 |
| `writeUvarint` / `writeVarint` | 1.7 | 43.2 |
| `nodeDB.GetNode` (per-key, per-ref `make`) | — | 6.0 |

## Decision

### 1. Serialization appends into one exactly-sized buffer

`writeUvarint`/`writeVarint` took an `io.Writer`. The compiler cannot prove an
interface method does not retain its argument, so their
`[binary.MaxVarintLen64]byte` scratch escaped to the heap on every call —
43% of all objects a block allocated, for a value that is at most 10 bytes.

`InnerNode`/`LeafNode` grow an `appendSerialized([]byte) ([]byte, error)` built
on `binary.AppendUvarint`, plus an exact `serializedSize()`. `SaveNode`
allocates the record once — payload capacity plus `checksumSize` — appends into
it and calls `sealChecksum`, replacing a geometrically-growing `bytes.Buffer`
followed by `stampChecksum`'s second full copy. `Serialize(io.Writer)` stays as
a thin wrapper for tests and external callers.

`serializedSize` must stay exact or the single allocation silently becomes two.
`TestSerializedSizeIsExact` pins it against `appendSerialized` over randomised
node shapes including the varint edge cases, and `TestVarintLenMatchesStdlib`
pins the two length helpers against the stdlib encoders they predict.

This is #5571's "pooled node serialization" (a `sync.Pool` of
`bytes.Buffer`s), with the pool dropped: exact sizing removes the intermediate
buffer rather than recycling it, and the record allocation itself cannot be
pooled because `dbm.Batch` retains the slice.

### 2. Decoding reads from a cursor and slab-allocates

`ReadNode` wrapped its payload in a `bytes.Reader` and allocated once per key
and once per 12-byte ref — up to 64 allocations for a single full leaf. It now
walks a `byteCursor` over the payload directly and backs all of a node's keys
with one allocation and all of its refs with another. A full leaf decodes in 4
allocations instead of ~67.

Decoded keys and refs are copied out of the payload rather than aliasing it:
some backends return their internal storage from `Get`.
`TestDecodedKeysDoNotAliasPayload` pins that by scribbling over the payload
after the decode.

This is #5571's "all-external leaf decode does one bulk `io.ReadFull`",
generalised to the keys as well.

### 3. Mini merkle maintenance is incremental

A leaf slot hash is `HashLeafSlotFromValueHash(key, valueHash)` — a function of
the entry, **not of the slot index**. An insert or delete that shifts entries
within a node can therefore shift their slot hashes alongside them. The slot
hashes already live in the node: they are `miniTree.tree[B:2B]`.

So `leafInsert`'s shift path shifts the slot hashes, hashes only the new entry,
and folds only the ancestors of the dirtied range (`MiniMerkle.buildRange`)
instead of recomputing all `numKeys + B-1` hashes. The split path carries the
`B+1` slot hashes into both halves. `leafRemove` and `innerInsert`'s shift path
do the same. For a sorted-key append (`pos == numKeys`) a leaf insert's fold
drops from 31 hashes to 5.

`saveNode` also stopped calling `RebuildMiniMerkle` unconditionally on every
dirty node. Saving a node does not change its hash (the NodeKey is not hashed),
and every mutating path leaves the mini merkle consistent with its slot inputs,
so the call recomputed a value it had just been given — `numKeys + B-1` hashes
per dirty leaf and `B-1` per dirty inner, per block, on top of what the write
path already paid. It now re-folds only the slots whose child hash actually
moved.

`TestMiniMerkleInvariant_RandomMutations` is what makes dropping that
safety net affordable: over randomised Set/Remove sequences it asserts, before
every save, that every in-memory node's cached hash equals a from-scratch
recomputation and that each inner's `childHashes` and mini-merkle slots agree
with its children — then saves, reloads from disk, and compares.

This is #5571's "incremental mini-merkle rebuild" without its `slotsDirty`
bitmap and parallel `slotHashes` array: the mini merkle already stores the slot
hashes, so the second copy was redundant.

### 4. Leaf mini merkles are deferred; inner mini merkles stay eager

Building a full leaf's mini merkle costs 63 SHA256 calls, which was ~95% of the
time to decode one — and no read path needs it. A Get/Has descent searches
keys, an iterator walks keys and valueKeys, and a leaf's hash is already cached
in its parent's `childHashes`. Decoding a leaf therefore no longer builds it;
`LeafNode.miniBuilt` records that.

The unbuilt case is handled **without mutating the node**. `Clone` builds on
the copy (every mutating path clones first, and a leaf is cloned about once per
write that touches it, so the write path pays the same build it paid at decode
time); `Hash` and `nodeMiniMerkle` fold into a local buffer and return a value.
Not mutating is the whole point: `nodeDB.GetNode` publishes decoded nodes into
a shared LRU and hands the same instance to the writer and to concurrent
committed-snapshot readers, so a node is immutable from the moment it is
decoded. #5571's version built lazily on first `Hash()` observation, which in
this codebase is a data race on 32-byte hash writes — a torn read yields a
wrong hash rather than a detected conflict.

**Inner nodes deliberately stay eager.** Their build is only `B-1` hashes, but
a block clones the same inner node once per write that descends through it — up
to every write in the block, for the root. Building at decode time lets the
node cache amortise that across all of them. Deferring inner builds as well was
measured and made a 100-write block **19% slower** even though it made decoding
faster; a leaf, cloned about once, has nothing to amortise.

Where the leaf deferral pays off follows from the same arithmetic. Eager builds
once per leaf cache MISS; deferred builds once per leaf CLONE, which is once
per write. So for a block with `W` writes and `R` read-only leaf loads,

    eager builds   = leaf cache misses over W + R
    deferred builds = W

Deferred wins by the read-only loads and loses by the cache hits. gno.land is
firmly on the winning side: the node cache is 10 000 entries
(`store/bptree.defaultCacheSize`) against a tree of millions of keys, and
pruning walks a version `KeepRecent` blocks old — 705 600 under the default
`syncable` strategy — so its leaf loads are always cold and never hashed.

Measured on memdb against the same tree with the deferral removed (median of
6; `tm2/pkg/bptree/benchmarks`, whose cache-to-tree ratio matches production):

| benchmark | eager | deferred leaf |
|---|---|---|
| `Block/block-100` (100 writes + save + prune) | 2752 µs | **2119 µs (1.30×)** |
| `Block/block-500` | 10 957 µs | **8284 µs (1.32×)** |
| `IterationFull/100k` | 36.0 ms | **20.4 ms (1.76×)** |
| `GetMiss/100k` | 4934 ns | **1960 ns (2.52×)** |
| `Has/100k`, fast index off | 5230 ns | **2172 ns (2.41×)** |
| `SaveVersion/100k` (writes only, no reads, no prune) | 233 µs | 243 µs (+4.5%) |

`SaveVersion` is the one case with no read-only loads and no prune to amortise
against, and it is the only regression.

`TestDeferredMiniMerkle_PoisonedCache` enforces the immutability rule by making
a violation loud: it reloads a tree so every node comes from the decoder,
overwrites every deferred leaf mini merkle with garbage, and then checks root
hashes, membership and non-membership proofs, further writes, deletes and a
reload against a tree that was never poisoned.

Proof generation is the one read path that does need a leaf's mini merkle, for
its sibling path, so it folds one per proof. `builtMiniMerkleInto` takes an
out-pointer (a `MiniMerkle` is 2 KiB; returning it by value copied that again),
and a `leafMiniCache` memoises the last leaf folded within one proof call,
because a non-existence proof builds existence proofs for two ADJACENT keys —
usually the same leaf. Without those two, `NonMembershipProof/100k` regressed
25%; with them it is 1.06× faster than before and `MembershipProof/1k` is
1.62× faster.

### 5. Node structs group their pointer fields first

Go's collector scans an object up to the last pointer in its type
(`_type.PtrBytes`). With `ndb` declared after `miniTree`, `InnerNode`'s
`PtrBytes` covered all 5376 bytes, so the GC walked every word of every COW
clone. Grouping the pointer-bearing fields first cuts the scanned range to
2040 bytes for `InnerNode` (−62%) and 1528 for `LeafNode` (−67%), with no
semantic change at all. `TestNodePointerBytes` pins the grouping, since a
pointer field drifting back to the end is a pure throughput regression with no
other symptom.

### 6. Prune compares child refs without allocating

`walkAndPrune` built a `map[string]bool` of child refs for every node it
visited, with a `string(ref)` conversion per entry — 18% of a block's allocated
objects, for a set of at most 32 twelve-byte values. The successor node's own
`children` array already **is** that set, so `InnerNode.hasChildRef` scans it
directly: no map, no copy, nothing built. And since a ref is a fixed 12 bytes,
it compares two scalars rather than calling `bytes.Equal` — which matters
because `walkAndPrune` does up to B×B of these per visited node pair, and the
scalar form measured 3.4× faster in isolation and 16% on the whole prune.

`sameRecord` compares `NodeKey` structs instead of `GetKey()` byte slices (two
allocations per comparison), and `nodeDB.DeleteNodeKey` builds the DB key in
one allocation instead of two, through the `NodeKey.putKey` that `GetKey` also
uses so the wire layout has one definition.

Together: allocations per pruned version drop ~84%, and retiring one version
from a 1M-key tree goes from 2.58 ms to 1.12 ms.

### 7. `Has` consults the fast index

`MutableTree.Get` and `ImmutableTree.Get` already serve a present key from the
fast index in one DB read. `Has` did not, and fell through to a full descent:
at 100k keys with the production cache size that was 5963 ns and 60
allocations, **slower than IAVL**. `fastHas` applies the same trust rules as
`fastGet` (an entry newer than the reader's snapshot is untrusted; a miss is
never proof of absence) without copying the value out.

The version guard is load-bearing in a way it is not for `Get`: the index holds
the latest value for a key, so a snapshot older than the key's first write must
not be told the key exists. `TestFastIndex_HasVersionGuard` pins that, and the
pre-existing `TestFastIndex_Differential` already compares `Has` with the index
on and off across 4000 randomised operations.

**This trade is not free on small trees, and it is the one knowingly accepted
regression in this change.** `Get` wins from the index unconditionally because
the index saves the out-of-line value read as well as the descent. `Has` reads
no value, so the index can only win when the descent would miss the node
cache — and on pebbledb a `Get` costs far more than a cache-resident walk:

| pebbledb, `bptree-fast` | before | after |
|---|---|---|
| `Has/1k` | 140 ns | 336 ns (0.42×) |
| `Has/10k` | 208 ns | 445 ns (0.47×) |
| `Has/100k` | 6613 ns | **1115 ns (5.93×)** |
| `HasMiss/1k` | 127 ns | 500 ns (0.25×) |
| `HasMiss/100k` | 6369 ns | 3738 ns (1.70×) |

The crossover is where the tree stops fitting in the node cache — between 10k
and 100k keys at the benchmark's 500-entry cache. gno.land's main store runs a
10 000-entry cache against a keyspace that will be orders of magnitude larger,
so it sits well past the crossover; the losing regime is a tree small enough to
be wholly cache-resident, where the absolute cost is 200–400 ns and
`bptree-fast` is still 4× faster than IAVL (`HasMiss/1k`: 500 ns against
IAVL's 2023 ns). A size- or cache-ratio gate was considered and rejected: every
threshold that separates the measured cases is a tuned constant, the node cache
is shared with block execution so residency is not predictable from tree size,
and a cliff in `Has` latency is harder to reason about than a uniform rule.
Operators who do not want the trade can mount the store without the index
(`StoreConstructor` rather than `FastStoreConstructor`).

A second, sharper edge is inherent to the record layout: an index entry is
`version‖value‖crc32c`, so verifying it is O(value size) even for a boolean
answer. At 40-byte values that is nothing against the 2288 ns walk it replaces,
and crc32c runs at ~10 GB/s here, so the crossover is around 24 KB of value for
the CRC alone — before the backend's own read of those bytes. gno.land stores
package sources and serialized realm objects, so values past that size exist.
`Has` cannot skip the CRC without breaking the trust rule it documents; making
it uniformly cheap would need a separately-checksummed 8-byte version header,
or a value-less existence entry. `Get` has had the same shape since the index
landed, so this is a property to know about rather than a regression.

This supersedes #5571's `fastNodes` latest-view LRU, which did the same job in
memory. See "Rejected" below.

### 8. Leaf keys are prefix-compressed (`TypeLeafPrefixed`)

Leaf keys are sorted and, at `B=32`, a leaf holds 32 **adjacent** keys — which
in a structured keyspace means they share a long prefix. gno.land's main store
keys are `"oid:" + 40 hex chars of PkgID + ":" + NewTime`
(`gnovm/pkg/gnolang/store.go`), so a leaf covering one realm's objects repeated
the same 44 bytes 32 times.

A new node type byte, `TypeLeafPrefixed` (0x03), stores one common prefix plus
a per-slot suffix. The type byte doubles as the format discriminator: readers
accept `TypeLeaf` (0x02) as well, so existing DBs mount unchanged and upgrade
record by record as leaves are rewritten. The type byte is **not** in the
Merkle commitment — node hashes come from the mini merkle over keys and value
hashes, never from the serialized bytes — so no app hash, proof or gas
parameter changes.

Measured over a 20k-key tree (`TestLeafPrefixed_OnDiskSaving`), total node
record bytes:

| keyspace | before | after | bytes/key |
|---|---|---|---|
| `oid:<pkgid>:<n>`, 4096 objects per realm | 1 984 419 | 1 104 180 (55.6%) | 99.2 → 55.2 |
| one realm, all 20k objects | 1 999 945 | 1 105 812 (55.3%) | 100.0 → 55.3 |
| random 16-byte keys | 1 280 651 | 1 267 955 (99.0%) | 64.0 → 63.4 |

The random case is the floor: adjacent random keys still share about a byte, so
the format never costs more than the one-byte prefix-length field it adds.

The common prefix is folded over **every** key rather than inferred from the
first and last. The sorted-keys shortcut (`LCP(keys[0], keys[last])`) is
correct, but if the ordering invariant were ever broken upstream it would emit
a wrong prefix and corrupt every key in the leaf on reload; a word-wise scan
makes the full fold a couple of hundred comparisons, and encoding a full leaf
costs 364 ns with zero allocations.

The decoder's allocation bounds needed care, because a prefix is charged once
per key: unlike the unprefixed block the payload is no longer a lower bound on
the slab. Per key, `prefixLen + suffixLen` must stay within `maxReadBytesLen`
(so no decoded key exceeds what `MaxKeyLen` allows on the write side), and the
block total must stay within `B * maxReadBytesLen`, the structural maximum for
`B` keys. A corrupt header can therefore demand at most the slab a legitimate
leaf of the same declared shape would, and nothing legal is rejected.

This is #5571's "leaf format v3", with its sub-version byte replaced by a node
type byte and its `commonPrefixLenCached` field dropped (the LCP fold is now
cheap enough that caching it is not worth the invalidation obligations on every
key-mutating path).

### 9. A non-membership proof walks the tree once

`GetNonMembershipProof` called `Has(key)` to reject an existing key and then
`GetWithIndex(key)` to locate the gap: two full descents, and the second
resolved the key's value on the exists-path purely to throw it away. One
`treeGetWithIndex` answers both. With the index on, it also drops the probe
decision 7 added, which could never save a descent this function performs
regardless — that probe was the entire reason `NonMembershipProof/1k` on
pebbledb regressed 12% before this.

## Rejected

### Mark-and-sweep prune (#5570's `4258d0b71`, carried into #5571)

#5570 replaced the dual-tree-walk with a mark-and-sweep over the retained
version set and measured prune at 33.18 ms → 7.69 ms (−77%) on memdb;
#5571 inherited the algorithm and reported it 12.6× faster than IAVL. The
benchmark behind those numbers prunes **50 versions in one call** against a
20k-key tree. **Neither result holds on this codebase at gno.land's scale, in
that shape or any other, and the current algorithm should stay.**

Mark-and-sweep marks every node reachable from the first retained version
before deleting anything, so its cost is Θ(retained tree) per
`PruneVersionsTo` **call**. The dual-tree-walk descends only where two adjacent
versions differ, so its cost is Θ(divergence) per version pruned. Which wins
depends on

    (versions pruned per call) × (divergence per version)   vs   (tree size)

and gno.land sits at the far left of that ratio: `store/bptree.Store.Commit`
calls `DeleteVersionsTo(version - 1 - KeepRecent)`, which advances by one each
block, so one version is retired per block with one block's writes of
divergence against a tree of millions of keys.

`prune_abbench_test.go` ports the mark-and-sweep algorithm onto the current
`nodeDB` API and benchmarks the two side by side (both using
`DeleteNodeKey`, so the comparison is not measuring a delete-path allocation
difference). The whole block cycle — apply 100 writes, `SaveVersion`, retire
one version — on memdb with the production 10 000-entry node cache:

| tree size | dual-tree-walk | mark-and-sweep | mark set |
|---|---|---|---|
| 100k | 1.04 ms | 1.09 ms (1.05× slower) | 4 813 |
| 1M | 1.72 ms | 4.45 ms (2.59× slower) | 47 842 |
| 3M | 2.17 ms | 11.72 ms (**5.40× slower**) | 143 447 |

The mark set grows linearly with the tree, exactly as the model predicts, and
the gap grows with it. Measured against the dual-walk *before* decision 6
sped it up, mark-and-sweep was ahead below ~100k keys (1.84 ms vs 1.98 ms at
100k, but 3.03 vs 5.36 at 1M and 3.35 vs 12.05 at 3M); with the allocation and
comparison fixes the dual-walk is ahead at every size measured, and the
crossover is gone.

Retiring one version alone, without the block's writes, says the same thing:
0.60 ms vs 0.61 ms at 100k, and 1.12 ms vs 3.11 ms at 1M.

Two further costs are not in those numbers. The mark walk is a full traversal
of the retained tree's inner nodes, so once the tree outgrows the node cache it
evicts the block's working set on every block — visible at 3M as 407 DB reads
per block against the dual-walk's 185, on a tree whose divergence is identical.
And the mark set is one map entry per node in the retained tree — ~37 MB at 10M
keys, allocated and discarded every block.

Mark-and-sweep is also quadratic in the number of versions retired per call,
which is the one shape it was supposed to win: it holds no record of what it
already deleted, so sweeping version `v` re-descends and re-deletes nodes
sweeping `v-1` already staged. Retiring 32 versions in one call on a 100k tree
issues 10 387 batch deletes against the dual-walk's 1 845, and takes 24.4 ms
against 8.76 ms.

The micro-optimisations inside the mark-and-sweep code do apply to the
dual-walk, and decision 6 above takes them.

### The `fastNodes` latest-view LRU (#5571 feature 1)

A key→value LRU in front of the tree walk. `master`'s persistent fast index
(#5937, #6018) already answers the same queries in one DB read, is shared
across restarts, serves committed snapshots as well as the working tree, and
costs no resident memory. #5571 itself had to add an "adaptive suspend" layer
to stop the LRU from regressing `GetHit/100K` by 105% once the working set
thrashed it. Decision 7 extends the existing index to `Has` instead, which is
where the gap actually was.

### Inline small values (#5571 feature 2, "leaf format v2")

Values up to a threshold stored in the leaf payload rather than under their own
`PrefixVal` record. This is the one idea that is actively **wrong** for
gno.land, not merely superseded.

Values are currently written once and shared by every version that references
them; only the 12-byte ValueKey is duplicated when a leaf is copied on write.
Inlining makes a leaf's values part of the leaf, so every CoW of a leaf
rewrites all 32 of its values. A leaf touched in 1000 blocks stores its values
1000 times. #5571 measured this as "known regression #4: multi-version disk
amplification, 2.5× larger than IAVL at 100 retained versions"; gno.land's
default `syncable` strategy retains **705 600** versions, where the
amplification is bounded only by how often each leaf is written.

Inlining also works against the fast index, which already holds the latest
value for every key, and it enlarges the node struct whose cloning is already
the single largest allocation source in a block.

### A shared `emptyHash()` slice (#5571's `f7b731d16`)

`master` returns a fresh copy deliberately, so a caller cannot mutate the
package-level empty-tree hash. It is only reached on empty-tree paths.

### `valueResolver` bound once, `getChild` explicit unlocks, LoadVersion fast path

Already true on `master`: resolvers are bound method values assigned at
construction (`ndb.getCommittedValue`, `ndb.GetValue`), `getChild` no longer
memoises and takes no lock, and #6383 removed the version-discovery scan from
query-view loads.

## Consequences

### Benchmarks

`tm2/pkg/bptree/benchmarks`, memdb, `-count=6`, median; `bptree-fast` is the
production configuration (`store/bptree.FastStoreConstructor`). IAVL is shown
for reference and is unchanged by this work.

| benchmark | before | after | speedup | IAVL |
|---|---|---|---|---|
| `Has/100k` | 5950 ns | **126 ns** | 47.3× | 3432 ns |
| `SaveVersion/100k` | 950.7 µs | **271.1 µs** | 3.51× | 966.6 µs |
| `SaveVersion/1k` | 770.1 µs | 228.4 µs | 3.37× | 771.8 µs |
| `Prune` (50 versions, 20k tree) | 36.9 ms | **13.1 ms** | 2.82× | 44.9 ms |
| `HasMiss/100k` | 5788 ns | 2176 ns | 2.66× | 4182 ns |
| `GetMiss/100k` | 5872 ns | 2212 ns | 2.65× | 82 ns¹ |
| `Block/block-100` | 4.07 ms | **2.29 ms** | 1.77× | 3.15 ms |
| `Block/block-500` | 15.00 ms | 8.75 ms | 1.71× | 11.37 ms |
| `IterationFull/100k` | 44.4 ms | 24.7 ms | 1.80× | 30.5 ms |
| `MultiVersionCreate/100-versions` | 90.6 ms | 51.4 ms | 1.76× | 69.5 ms |
| `NonMembershipProof/100k` | 10.6 µs | 10.0 µs | 1.06× | 16.9 µs |
| `MembershipProof/1k` | 2.45 µs | 1.51 µs | 1.62× | 4.61 µs |
| `MembershipProof/100k` | 7.87 µs | 7.42 µs | 1.06× | 10.81 µs |
| `Remove/100k` | 6.89 µs | 4.50 µs | 1.53× | 1.51 µs |
| `SetInsert/100k` | 8.94 µs | 6.29 µs | 1.42× | 3.19 µs |
| `WorkingHash/100k` | 89.6 µs | 64.3 µs | 1.39× | 62.7 µs |
| `SetUpdate/100k` | 5.08 µs | 4.98 µs | 1.02× | 1.88 µs |

¹ IAVL keeps a bloom-style miss short-circuit the B+32 tree has no equivalent
of; a B+32 `GetMiss` still walks. The fast index cannot answer a miss (it is
never authoritative for absence), which is also why `HasMiss/1k` is 37 ns
slower than before — it now pays an index probe first. That trade is 2.66× at
100k and is what `BenchmarkHasMiss` (added here) exists to keep visible.

Allocations per operation drop across the board: `SaveVersion/100k` 10 023 →
1576 (−84%), `Block/block-100` 37 533 → 7087 (−81%), `Prune` 420 714 → 72 486
(−83%), `IterationFull/100k` 542 714 → 338 093 (−38%), `Has/100k` 60 → 1.

On pebbledb, the production backend (`-count=5`, median):

| benchmark | before | after | speedup | IAVL |
|---|---|---|---|---|
| `Has/100k` | 6613 ns | **1115 ns** | 5.93× | 7499 ns |
| `Prune` | 46.9 ms | **24.0 ms** | 1.96× | 148.3 ms |
| `GetMiss/100k` | 7129 ns | 3738 ns | 1.91× | 1021 ns |
| `HasMiss/100k` | 6369 ns | 3738 ns | 1.70× | 9041 ns |
| `Remove/100k` | 6490 ns | 4160 ns | 1.56× | 1599 ns |
| `SetInsert/100k` | 8911 ns | 6191 ns | 1.44× | 4193 ns |
| `SaveVersion/100k` | 6.39 ms | 4.47 ms | 1.43× | 7.13 ms |
| `WorkingHash/100k` | 87.3 µs | 62.2 µs | 1.40× | 81.8 µs |
| `Block/block-500` | 20.66 ms | 16.17 ms | 1.28× | 29.01 ms |
| `Block/block-100` | 14.19 ms | **11.74 ms** | 1.21× | 12.45 ms |
| `IterationFull/100k` | 134.8 ms | 116.7 ms | 1.15× | 14.3 ms |
| `NonMembershipProof/100k` | 15.67 µs | 15.01 µs | 1.04× | 28.03 µs |

Two things stand out there and neither is addressed here. `GetMiss` is the one
operation where IAVL is structurally ahead: it short-circuits a miss, while a
B+32 miss always walks (and the fast index cannot answer a miss, because it is
never authoritative for absence). And `IterationFull` on pebbledb is 8× behind
IAVL, for a reason the profile makes plain: IAVL stores values inside its
leaves, while B+32 resolves each one through its own `PrefixVal` record — 100k
random point reads for a 100k-key scan. #5571 solved that by inlining values,
which this ADR rejects for retention reasons; the shape that would work here is
an ordered scan of the fast index, which already holds the latest value for
every key **in user-key order**, with a fallback to the authoritative walk when
the stamp does not cover the iterator's version. That is a feature, not a
tuning change, and belongs in its own PR.

The three write benchmarks still behind IAVL — `SetInsert`, `SetUpdate`,
`Remove` — are all bounded by the same thing, and it is not addressed here: a COW write
clones one ~5 KiB node per level, so a single Set allocates ~24 KiB against
IAVL's ~3 KiB. Reusing a node that is already unsaved in the current session
instead of cloning it again would cut that by about half, but it changes a core
COW invariant (today a failed descent cannot have touched the published tree)
and belongs in its own change.

### On-disk format

New writes emit `TypeLeafPrefixed`. Readers accept both leaf types, so an
existing DB mounts unchanged and converts record by record as leaves are
rewritten; `TestLeafPrefixed_MixedFormatTreeOnDisk` mounts a downgraded DB,
writes to it, and reloads the resulting mixed-format tree.

A binary **older** than this change cannot read records written after it. A
rollback across this change therefore needs a resync, like any other store
format change. No app hash, proof, or gas parameter is affected.

### Deliberately out of scope

Each of these is a deeper fix for something this change only worked around.
They are recorded here so the next performance pass removes a cause rather
than adding a fifth workaround to it.

**Fixed-width node refs instead of slices.** `children [B][]byte` and
`valueKeys [B][]byte` hold 32 twelve-byte values as 32 slice headers (768
bytes) plus 32 heap payloads. Four of this change's optimisations exist only
because of that: `readRefs` slabs the 32 allocations the decoder would
otherwise make, `DeleteNodeKey` exists to avoid `GetKey` + `nodeDBKey`'s two,
`InnerNode.hasChildRef` exists to avoid a `map[string]bool` over them, and
`sameRecord` now compares `NodeKey` structs because comparing the serialized
form allocated twice. They are also most of what keeps `PtrBytes` at
2040/1528 — 32 slice headers are 32 pointer words. `children [B]NodeKey` /
`valueKeys [B]ValueKey`, with the zero value as the "not yet saved" sentinel
(version 0 is never a valid saved version), would delete all four and take
`InnerNode` to ~1270 pointer bytes and `LeafNode` to ~780. It touches
`nodeDB.GetNode`'s signature, import, export and prune, so it is its own PR.

**An inner node's mini-merkle slot level duplicates `childHashes`.** For an
inner node `miniTree.tree[B:B+nc]` is always exactly `childHashes[:nc]`, so
the hottest struct in the process carries 1 KiB of a persisted field twice and
five mutating sites carry a hand-maintained sync obligation (`saveNode`'s
two-clause dirty test is the tell). Having `MiniMerkle` store only the
internal levels and take the slots as a parameter — `Build(slots)`,
`buildRange(slots, lo, hi)`, `SiblingPath(slots, i)` — would delete every
`copy(slots, childHashes)`, take 1 KiB off `InnerNode` (~19% of the single
largest allocation source in a block), and make the leaf/inner asymmetry
structural rather than a comment. It changes exported `MiniMerkle` API and
every mutating path.

The leaf side has a storage twin: deferring the leaf build removed the
*computation* on the read path but not the *storage*. A DB-decoded leaf still
carries 2048 inline bytes of mini merkle that nothing reads — at the 10 000-entry
default cache that is up to ~20 MB of live heap holding zeroes, which is what
drives GC pacing, and `Clone` copies those 2 KiB immediately before
`RebuildMiniMerkle` overwrites them. Making `LeafNode.miniTree` a pointer
allocated only where it is built trades that for one allocation per mutated
leaf and an explicit copy in `Clone`; it is a residency win paid for on the
write path, so it needs its own measurement.

**COW clones an already-dirty node again.** A write clones one node per level,
so a Set allocates ~24 KiB against IAVL's ~3 KiB, and within one block the
root is cloned once per write even though only the last clone survives.
Mutating a node that is already unsaved in the current session would halve
that, but it gives up the present guarantee that a failed mid-descent load
cannot have touched the published tree.

**Iteration resolves one value per key.** On pebbledb a 100k-key scan issues
100k random `PrefixVal` point reads and is 8× behind IAVL. The fast index
already holds the latest value for every key in **user-key order**, so an
ordered scan of it — falling back to the authoritative walk when the stamp
does not cover the iterator's version — would close that without the
retention cost of inlining values into leaves.

**Prune rediscovers orphans by walking.** `walkAndPrune` costs one `GetNode`
per divergent old child plus a `findCorrespondingChild` re-descent from
`newRoot` for each, without first trying the corresponding child of the
`newNode` it already holds — O(H²) loads where O(H) would do. Deeper still:
record orphaned NODE records at write time, as the tree already does for
valueKeys (`t.versionOrphans`), and prune becomes O(orphans) with no tree walk
at all — which retires the mark-and-sweep question structurally instead of by
benchmark.

**`Has` could locate its own crossover.** Rather than the uniform probe
decision 7 settles on, a cache-peek-only descent that abandons to the index
the moment a node would need a DB read would answer in ~140 ns whenever the
path is resident and probe only when it is not, removing the small-tree
regression without a tuned threshold. It needs a peek-only `getChild` and a
second walk mode.

### Invariants a future change must respect

- `serializedSize` must stay exactly equal to what `appendSerialized` writes.
- Every mutating path must leave a node's mini merkle consistent with its slot
  inputs; `saveNode` no longer fixes it up.
- A node is immutable from the moment `GetNode` decodes it. A deferred leaf
  mini merkle is built on a clone or folded into a local, never in place.
- Pointer-bearing fields stay grouped at the top of `InnerNode`/`LeafNode`.
- `TypeLeafPrefixed`'s common prefix must be a prefix of every key in the leaf,
  not just the first and last.

Each of these has a test that fails loudly when broken; they are listed next to
their decisions above.
