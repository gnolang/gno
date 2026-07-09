# ADR: Serve clean-working-tree reads from the bptree fast index

## Status

Proposed (PR pending).

## Context

The bptree fast index (`tm2/pkg/bptree/fast_index.go`) is an optional,
advisory, latest-version accelerator: a flat `user-key → version‖value` map
that resolves a point Get of a present key in 1 DB read instead of a full
tree descent plus the out-of-line value read (~3 reads at 100M keys with the
default 10K node cache; measured 3.66, see `PERFORMANCE.md`).

Until this change it was consulted only by `ImmutableTree.Get` on committed
snapshots — the ABCI-query surface. The consensus execution read path
(`store/bptree.Store.Get → mutableTreeAdapter → MutableTree.Get`) always did
the full walk, so the index never helped where reads are gas-priced and
wall-clock-critical.

Key observation: tm2 buffers all block writes in cache layers above the tree.
`BaseApp.Commit` runs `deliverState.ms.MultiWrite()` immediately before
`cms.Commit()` (`tm2/pkg/sdk/baseapp.go`), so during all of
DeliverTx/CheckTx/Simulate the working tree is byte-identical to the last
committed version — exactly the state the fast index describes.

## Decision

`MutableTree.Get` consults `fastGet` when `fastReadable()` holds: the feature
is on AND `t.root == t.lastSaved` (pointer identity with the committed root).
Every published mutation COW-clones the root (`treeInsert`/`treeRemove` clone
at entry), so pointer identity exactly witnesses "no staged mutations"; the
same predicate is the pointer-identity component of `PruneVersionsTo`'s
clean-session check. Any staged mutation routes reads back to the
authoritative walk, so the gate degrades to the status quo whenever the
buffering assumption doesn't hold (e.g. stores written directly outside the
cache layers stay dirty until their next SaveVersion and simply walk).

Placement: the gate sits after the nil-root early return. That avoids probing
an empty tree, and covers the one pointer-reunion case (a committed-empty tree
with a staged Set+Remove round-trip has `root == lastSaved == nil`) as defense
in depth.

`Importer.Commit` bypasses per-entry index maintenance, so `Import()` now
drops the index stamp (own commit) and clears all `'F'` entries up front,
where the batch is empty — pre-existing entries are stamped ≤ the import
version and would otherwise pass the version guard and serve pre-import
values after Commit. An abort after the clear costs only a missing index
(the next `Load` rebuilds); it can never cause a wrong read.

### The trade this makes (read this first)

An unauthenticated, non-Merkle index now feeds consensus-execution reads on
nodes that enable it. A checksum-valid-but-wrong entry (disk tampering, a
rebuild bug, out-of-contract staleness) no longer just corrupts a query
answer — that node computes a different app hash and forks itself off the
network. We accept this because:

- entries are maintained in the SAME atomic batch as the tree
  (Set/Remove/SaveVersion), so they cannot disagree with the committed tree,
  even across a crash;
- every entry is checksummed, and any miss/too-new/corrupt entry falls back
  to the authoritative walk (the index can never fabricate absence — only a
  trusted-present hit is served). Known limitation: the CRC covers
  version‖value, not the user key (which lives in the DB key), so key↔payload
  cross-wiring below the backend's own block checksums would not be detected —
  a follow-up is to fold the key into the CRC (index format bump +
  rebuild-on-upgrade);
- the feature is node-local and per-mount (`FastStoreConstructor` vs the plain
  `StoreConstructor` in `tm2/pkg/store/bptree`), the same trust model as
  cosmos IAVL's fast nodes, which likewise serve consensus reads when enabled
  (tm2's own IAVL mount currently constructs with fast storage skipped);
- gas is formula-based (depth params × tree size), so consensus gas is
  identical whether the index is on or off — this change is consensus-neutral.

### Trust contract

(The canonical living version of this contract is the `fast_index.go` header
comment; this section records it as of this PR.)

Index currency is verified by `Load` (`ensureFastIndex` rebuilds on a stamp
mismatch) and preserved by eager same-batch maintenance plus the Import-time
clear. A tree reached ONLY via `LoadVersion` — never `Load` — over a DB whose
later versions were committed with the feature off is outside the contract
(nothing re-verifies the stamp there). This is documented rather than
mechanized: the in-repo store layer always goes through `Load`, which fails
startup on a rebuild error rather than serving a stale index. For reads at an
older version (`LoadVersion(old)` or old snapshots), the per-entry version
guard suffices: an entry newer than the snapshot is rejected; an entry ≤ the
snapshot is provably the key's latest write ≤ that snapshot (later writes
re-stamp, removals delete the entry).

Two hardening follow-ups: `ensureFastIndex` should be read-only over
immutable DBs (their `NewBatch` returns nil, so a rebuild triggered on a
query-path open — unreachable today, since the stamp is maintained
transactionally and the writable startup Load rebuilds first — would
nil-deref rather than degrade); and a CORRUPT stamp currently fails the load
(fail-stop) where a MISSING stamp rebuilds — rebuilding on a corrupt stamp
is strictly safer for an advisory structure.

## Alternatives considered

- **Cosmos-IAVL-style unsaved-fastnode overlay** (in-memory adds/removes for
  read-your-writes on a dirty tree): solves a problem tm2 doesn't have — the
  tree is clean during execution, and the only dirty window (inside Commit)
  has no interleaved reads. The overlay is upstream IAVL's most bug-prone
  subsystem; rejected.
- **Explicit `dirty bool`**: derived state with ~8 maintenance sites
  (Set/Remove/SaveVersion/Load/LoadVersion/Rollback/Import); pointer identity
  is already load-bearing for Rollback and prune, needs no new state, and the
  read-your-writes tests fail loudly if a future refactor breaks it. Rejected.
- **Fast path for `Has` (both trees)**: no reachable caller benefits —
  `cacheStore.Has` is implemented as `Get(key) != nil`, so consensus existence
  checks already resolve through Get, and the query path never calls tree-level
  Has. Rejected as scope creep; `Has` stays walk-only.
- **Enabling the index by default for all stores**: at the time of this PR
  nothing mounted the bptree store in production (gno.land mounted IAVL),
  and enabling the index
  pre-commits a first-`Load` full rebuild (hours at 100M keys) plus doubled
  value bytes on disk. Belongs in the mount PR together with gas-depth
  repricing (present-key GET → ~1 flat op; each Set/Remove gains one index
  write). Deferred — the mount PR selects it per-mount via
  `FastStoreConstructor`.

## Consequences

- With the index on, a present-key point Get on the clean working tree is
  1 DB read (was ~3–3.7 at 100M keys) — matching IAVL+fast-node point-read
  performance while keeping B+32's ~4.5× write advantage.
- Absent-key Gets pay one extra flat read (the index miss) before the walk;
  gas is unchanged (formula-based).
- Benchmarks: for index-on fixtures the default working-tree read mode and
  `-disk-committed-read` now measure the same fast path; the plain `bptree`
  factory remains the index-free baseline (comments updated in
  `benchmarks/`).
- Concurrency: unchanged contract. The gate reads only working-tree fields
  Get already read; `fastGet` reads the internally-synchronized DB. New
  `-race` coverage exercises committed-snapshot readers (fastGet) against a
  committing writer with the index on.
- Tests pin the behavior with doctored-entry probes (valid-checksum entries
  with wrong values planted directly in the DB): a probe served proves the
  fast path fired, the real value proves the walk fired, and each probe
  self-verifies so a malformed probe cannot silently turn a fast-path
  assertion into a vacuous walk test.

## AI assistance

Implemented with AI assistance (plan and diff reviewed through multi-agent
review rounds to convergence); the human author reviewed and owns the change.
