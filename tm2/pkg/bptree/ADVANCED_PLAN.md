# BP32 advanced performance plan

Branch: `feat/alex/bp32tree-advanced`. Baseline commit: `aa27d6920` on
`feat/alex/bp32tree-second-pass`. Benchmark comparison files live under
`tm2/pkg/bptree/benchmarks/final/` — overwritten after each phase;
always commit them alongside the phase's code change so before/after
is reproducible from git history.

## Status

| Phase | Status | Commit |
|---|---|---|
| 1.1 Fast-node cache | ✅ landed | `213301508` |
| 1.2 Pooled serialization (arena-style) | ✅ landed | `213301508` |
| 1.3 Deferred mini-merkle rebuild | ✅ landed | `213301508` |
| 1.4 Incremental slot-hash cache | ✅ landed | `213301508` |
| 1.5 Prefetch-ahead iterator | ⏭ skipped — goroutine overhead exceeded goleveldb parallelism benefit; inline values (2.1) solves the same problem more directly |
| 2.1 Inline small values | ⬜ open |
| 2.2 Leaf sibling pointer | ⬜ open |
| 2.3 Prefix compression | ⬜ open |
| 3.1 Bloom filter per inner node | ⬜ open |
| 4 Delta-encoded persistence | ⬜ open (gated) |

## Rules for every phase

- **No new benchmarks.** The existing suite in
  `tm2/pkg/bptree/benchmarks/bench_test.go` is the reference; every
  phase is judged by its effect on rows already there.
- **Re-baseline after each landing.** Overwrite
  `benchmarks/final/{memdb,goleveldb}.txt` with a fresh
  `-benchtime=3s -count=3` run before calling the phase done.
- **Exit criterion = benchmark delta.** If the planned row doesn't
  improve (or regresses a row it shouldn't), rework or revert.
- **Format changes bump a type byte.** Reader handles both legacy and
  new types; writer emits new exclusively once migrated.
- **Single-writer invariants remain.** `MutableTree` is single
  goroutine; `ImmutableTree` is multi-reader; prune blocks readers via
  the existing `pruneMu` / version-reader machinery.

---

## Phase 2 — leaf-format v2 (single migration bump)

All three sub-items bundle into one format byte: `TypeLeafV2 byte = 0x12`
alongside `TypeLeaf byte = 0x02`. Reader dispatches on type byte. Writer
emits v2 once this phase lands. Old leaves auto-upgrade on the first
save that touches them.

### 2.1 Inline small values — CENTRAL WIN

Large values remain external (valueKey indirection). Small values live
inside the leaf. The goleveldb iteration regression (BP32 ~6× slower
than IAVL at 100k full scan) is entirely per-value DB round-trips;
inlining eliminates them for the common small-value case.

**Threshold**: `Options.InlineValueThreshold`, default **64 bytes**.
Values ≤ threshold inline; larger go external.

**In-memory layout additions** on `LeafNode`:

```go
type LeafNode struct {
    // ... existing fields ...
    inlineValues [B][]byte  // non-nil for inline slots
    inlineMask   uint32      // bit i set = slot i is inline
    // valueKeys[i] == nil when inline; valueHashes[i] still computed
    // (proofs still use the value hash path)
}
```

**Serialized layout (v2)**:

```
type (1)                      = 0x12
numKeys (uvarint)
keys... (see 2.3 prefix compression)
valueHashes (32 × numKeys bytes, fixed-width)
inlineMask (uvarint)          # 32-bit bitmap
for each slot i in 0..numKeys-1:
    if inlineMask & (1<<i):
        inlineLen (uvarint) + inline payload
    else:
        valueKey (12 bytes)
siblingFlag (1)               # see 2.2
if siblingFlag:
    siblingNodeKey (12 bytes)
```

**Mutation paths**:

- `Set(k, v)`: if `len(v) ≤ threshold`, set `leaf.inlineValues[pos] = v`
  (copy, not share), set `leaf.inlineMask |= 1<<pos`,
  `leaf.valueKeys[pos] = nil`. Do **not** call `SaveValue`; no session
  value bookkeeping for inline entries.
- Update: if the slot transitions inline↔external or inline↔inline of
  a different size, the previous storage is dropped. External → inline
  transition orphans the old valueKey. Inline → external requires a
  fresh valueKey via `SaveValue` as today.
- Remove: drop `leaf.inlineValues[pos]`; no session cleanup needed
  since nothing was saved externally.

**Read paths**:

- Add `valueAt(slot int, resolver ValueResolver) ([]byte, error)` on
  `LeafNode`: inline path returns `leaf.inlineValues[slot]`; external
  path delegates to `resolver(leaf.valueKeys[slot])`.
- Every current `resolveValue(leaf.valueKeys[i])` call site switches to
  `leaf.valueAt(i, t.valueResolver)`. Callers: Get, Has (doesn't need
  value, so only the existence check runs), Iterate, proof
  generation, export.
- `resolveValue` stays for the external path, but is only called
  through `valueAt`.

**Proof generation**: the leaf-slot hash is still
`HashLeafSlotFromValueHash(key, valueHash)`. Value hash is always
computed and stored, regardless of inline/external — proof format
unchanged. `ExistenceProof.Value` is the raw value; `valueAt` supplies
it.

**Fast-node cache**: orthogonal. Cache stores `value []byte` indexed by
key; populated on `Get` hit (inline or external, same path).

**Migration**:

- Reader: dispatch on type byte. `TypeLeaf` (0x02) uses old reader with
  inlineMask = 0 and inlineValues = nil. `TypeLeafV2` (0x12) uses new
  reader.
- Writer: always emits `TypeLeafV2`. Old leaves auto-upgrade on the
  first save that dirties them. Background `Compact()` helper walks
  every node and rewrites — optional, not required.

**Exit criterion**:

- `BenchmarkIterationFull/bptree/100k goleveldb`: currently **122 ms**,
  target **< 40 ms** (within 2× of IAVL's 19 ms).
- `BenchmarkIterationRange/bptree/100k goleveldb`: currently **798 µs**,
  target **< 250 µs**.
- `BenchmarkDiskSpace/bptree/100k goleveldb`: should shrink (values
  move inline, net per-key on-disk footprint drops for small values).
- `BenchmarkGetHit` memdb: should stay flat — inline path touches one
  more field but no DB.

**Risks / gotchas**:

- **sessionValues accounting**: inline Sets don't add to `sessionValues`.
  Rollback iterates `sessionValues` and deletes via
  `DeleteValueDirect`. No change needed; inline slots are wiped by
  reverting `t.root = t.lastSaved`.
- **orphanValueKey on update**: an inline-then-external update
  currently orphans the old valueKey. An external-then-inline update
  also needs to orphan the old valueKey. An inline-then-inline update
  just swaps the bytes, no orphan.
- **export / import**: `ExportNode.Value` is the raw value today. Same
  works for inline and external (Importer calls SaveValue for every
  value; if we want to preserve inlineness across import, we need a
  v2 ExportNode format — probably defer to a later phase).
- **Mini-merkle**: unchanged. Slot hash is still
  `HashLeafSlotFromValueHash(keys[i], valueHashes[i])` —
  `valueHashes[i]` is computed regardless.
- **slotHashes cache**: unchanged.

### 2.2 Leaf sibling pointer

Persisted `nextLeaf *NodeKey` (or zero) on each LeafNode. Set during
save; read during iteration as a shortcut past the stack climb.

**Non-authoritative design** — if `GetNode(nextLeaf)` returns
`ErrNodeNotFound` (pruned, for example), the iterator falls back to the
existing `nextLeaf()` stack walk. The sibling pointer is a hint, not a
contract.

**Iterator changes**:

```go
func (it *Iterator) nextLeaf() {
    // Try sibling hint first
    if next := it.leaf.nextLeaf; next != nil {
        if node, err := it.ndb.GetNode(next.GetKey()); err == nil {
            if leaf, ok := node.(*LeafNode); ok {
                it.setLeaf(leaf)
                it.leafIdx = 0
                it.valid = true
                it.checkEnd()
                return
            }
        }
        // Fall through on cache miss / wrong type / pruned
    }
    // Existing stack-based fallback
    for len(it.stack) > 0 { ... }
}
```

`prevLeaf` gets the same treatment with a `prevLeaf` pointer — or skip
descending for v1 and only wire the ascending shortcut.

**Writing the pointer**: during `SaveVersion`, walk the tree leaf-by-leaf
in key order and assign each leaf's `nextLeaf` to the next leaf's
`NodeKey`. This walk already happens in saveNode (post-order); we'd
need a second pass after NodeKey assignment to populate sibling
pointers, then re-serialize. Alternative: do a forward scan, queue
assignments, save in two phases.

**Risks**:

- **Prune invalidation**: a sibling pointer can reference a node in a
  pruned version. The non-authoritative design handles this — fallback
  restores correctness at some iteration cost.
- **COW cost**: mutating a leaf's sibling pointer when its neighbour
  rewrites requires cloning the leaf. If we update sibling pointers
  only during SaveVersion (not on every Set), this is already covered
  by the "leaf is dirty → saved" path.

**Exit criterion**:

- `BenchmarkIterationFull/bptree/100k goleveldb`: further improvement
  past 2.1's inline-values win, ideally approaching IAVL's 19 ms.
- `BenchmarkIterationRange` on goleveldb: another 10-20% win.
- `BenchmarkIterationFull/bptree/100k memdb`: stay flat (no DB reads
  avoided).

### 2.3 Prefix compression

Keys in a leaf are sorted, so adjacent keys often share a long common
prefix (typical for gno.land patterns like `vm:<realm>:<path>`).

**Serialized layout** of the keys block:

```
commonPrefixLen (uvarint)
commonPrefix (commonPrefixLen bytes)
for each slot:
    suffixLen (uvarint)
    suffix (suffixLen bytes)
```

**Writer**: compute common prefix of the first and last key (already
sorted, so first and last bound the prefix). Emit common prefix once,
suffixes per slot.

**Reader**: reconstruct full keys on deserialize. Store full keys
in-memory (no in-memory prefix compression — keep existing
`LeafNode.keys [B][]byte`). On-disk-only optimization.

**Exit criterion**:

- `BenchmarkDiskSpace/bptree/100k`: additional 10-30% disk reduction
  when key patterns share prefixes.
- No time-of-use regression — the prefix decompression adds one
  varint read + one copy per slot at deserialize time, amortised by
  the node cache.

---

## Phase 3 — inner-format v2

### 3.1 Bloom filter per inner node

Type bump: `TypeInnerV2 byte = 0x11`. Appends an 8-byte bloom filter
after the existing fields.

**Filter construction**: fold the first 8 bytes of each separator key
into a 64-bit bitmap via FNV or XOR-rotate. At 32 keys × 8 bytes
folded, ~0.1% FPR for uniformly-distributed keys.

**Query path**: `searchInner` on the non-existence path:

```go
func searchInner(n *InnerNode, key []byte) int {
    // Bloom check: if none of the separators could match, jump to the
    // known child without the bytes.Compare loop.
    if !n.bloomMightMatch(key) {
        return 0  // descend into first child (where < smallest-key lives)
    }
    // existing binary search
}
```

Actually the bloom helps less here than for Has/Get on random keys —
searchInner still needs to find WHICH child, not just whether a match
exists. Bloom only eliminates the bytes.Compare cost when zero matches
are possible.

**Better application**: inline the bloom check into the `GetMiss` /
`Has`-returning-false paths before descending into children for
confirmation. If bloom says "definitely not", return miss without
descending.

**Exit criterion**:

- `BenchmarkGetMiss/bptree/100k memdb`: currently **168 ns**, target
  **< 130 ns**.
- `BenchmarkNonMembershipProof`: modest improvement (half the descent
  is the proof walk, not a short-circuit candidate).

**Risks**:

- Bloom false-positives fall through to the existing loop — no
  correctness impact, only missed optimization.
- The 8-byte overhead per inner node grows on-disk size by ~0.2%.
  Negligible.

---

## Phase 4 — delta-encoded persistence (gated)

Ship last, behind `Options.DeltaEncoding`.

**Problem**: multi-version disk at 100 retained versions is ~1.88×
IAVL's. Every dirty node is persisted as a full node even if only one
slot changed.

**Format**: `TypeInnerDelta byte = 0x13`,
`TypeLeafDelta byte = 0x14`.

```
type (1)
parentNodeKey (12)            # the base version this delta applies to
changedSlotMask (uvarint)     # bitmap of modified slots
for each set bit:
    slot payload (slot-specific encoding)
```

**Reader**: walk the delta chain. For each `TypeInnerDelta` /
`TypeLeafDelta`, load the parent node, apply the changed slots,
recurse until a non-delta base is reached. **Cap depth at 4**; a 5th
save in the chain forces a full-node rewrite.

**Materialization cache**: an LRU mapping NodeKey → fully-materialized
`*InnerNode`/`*LeafNode`. Deltas never hit the main nodeCache — only
materialized nodes do, keyed by their own NodeKey. Singleflight
protects the materialization.

**Prune cooperation**: when sweeping a version, if we delete a node
that is the base of a still-reachable delta chain, we must first
materialise and rewrite the chain's earliest reachable ancestor as a
full node. This adds a "rebase" step to sweep.

**Exit criterion**:

- `BenchmarkDiskSpaceMultiVersion/bptree/100-versions goleveldb`:
  currently **10.75 MB**, target **≤ 6 MB** (parity with IAVL's 5.7 MB).
- `BenchmarkLoadVersion`, `BenchmarkMultiVersionCreate`, `BenchmarkPrune`:
  regress no more than 10%.

**Risks**:

- Reader complexity (chain walk, materialisation cache, concurrency).
- Prune becomes more elaborate.
- Wrong interaction with `nodeCache` — deltas must not be cached as
  if they were whole nodes.

---

## Cross-cutting

### Format versioning

One on-disk-format byte under `PrefixMeta` (add a new prefix `F` or
extend the existing meta record). Bumped per phase:

- After Phase 2 lands: format v2.
- After Phase 3 lands: format v3.
- After Phase 4 lands: format v4.

Reader on `Load()` checks the byte and refuses to mount if the binary
supports an older version than the stored format.

### Fuzz coverage

One test target: `Fuzz_NodeFormatRoundTrip`. For every supported type
byte, serialize-then-deserialize an arbitrary node and verify
`Hash()` and slot contents are identical. Catches every format
regression.

### Comparison discipline

After each phase lands:

1. Run `tm2/pkg/bptree/benchmarks/` with
   `go test -bench=. -benchmem -benchtime=3s -count=3` for memdb and
   goleveldb; write to `benchmarks/final/{memdb,goleveldb}.txt`.
2. Use `git show HEAD~1:tm2/pkg/bptree/benchmarks/final/<backend>.txt`
   for the pre-phase reference.
3. Confirm every row in the phase's exit criteria moved in the
   expected direction; no row marked "should stay flat" regressed
   by >10%.
4. Commit code + benchmark data together.

### Not in scope

- New benchmarks — comparison against the existing suite only.
- Iterator prefetch-ahead (tried in Phase 1.5, reverted). Inline values
  (2.1) solves the same read-latency problem more cleanly.
- B parameter changes (breaks proofs).
- Proof-format changes.
- SHA256 replacement (compat break).
- Lock-free write path (single-writer contract gives us the wins
  without the hazards).
- Batch proof generation — no existing benchmark covers it; defer.
