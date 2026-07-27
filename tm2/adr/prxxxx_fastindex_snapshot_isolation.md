# ADR: Query-path snapshot isolation and fast-index write-proofing (gno#6011)

## Status

Proposed (fixes https://github.com/gnolang/gno/issues/6011).

## Context

A topaz-1 node rejected the proposal for block 227783 with an AppHash
mismatch. Investigation traced it to a persisted bptree fast-index entry that
held an account value from block 224859 while the authoritative tree carried
the block-227773 update — under a fast-index stamp that claimed currency at
227782. The stale entry was served by `MutableTree.Get`'s clean-working-tree
fast path (PR #5937) during block execution, diverging the node's state.

The root cause is a chain of four interacting design gaps, reproduced both
deterministically and under natural concurrency
(`tm2/pkg/store/rootmulti/fastindex_*_test.go`):

1. **Snapshot bypass.** `rootmulti.constructStore` routed mounts with a
   dedicated db (`params.db != nil`) to that LIVE db even when building the
   "immutable" query multistore of `MultiImmutableCacheWrapWithVersion`,
   whose `ims.db` snapshot then only covered rootmulti metadata. gno.land
   mounts both stores with `cfg.DB`, so every query-path store read bypassed
   the snapshot. (`Simulate` was already routed through
   `MultiImmutableCacheWrapWithVersion` — it was a *victim* of this bypass,
   not a separate unsafe path.)
2. **Concurrent queries.** The query ABCI connection has an independent mutex
   (`proxy.NewReadOnlyABCIClient`), so query loads run concurrently with the
   consensus commit.
3. **Write-capable query loads.** Immutable multistores skip the CollectingDB
   wrap, so the throwaway query store held a REAL writable batch over the
   live DB; `bptree.Store.LoadVersion` (Immutable branch) ran `Load()` →
   `ensureFastIndex()`, a maintenance path that can WRITE.
4. **Load TOCTOU.** `Load()` reads "latest" via a `discoverVersions` iterator
   (a pre-commit point-in-time view) and the fast-index stamp via a later
   `Get`. When the commit drain for block N landed between the two reads, the
   query load saw latest = N−1 with stamp = N, concluded the index was stale,
   and `rebuildFastIndex()` silently rewrote EVERY fast-index entry from
   root@N−1 — old values, old versions — into the live DB. Later commits
   re-advanced the stamp, making the poisoning undetectable at restart. Keys
   updated in block N and never touched again stayed stale; the incident's
   account was one (entry vk 224859 = its last write before 227772).

Adjacent bugs found during the investigation:

- `batchHandle` (CollectingDB batches) forwarded ops to the shared collector
  at `Set` time with no-op `Write`/`Close`, so bptree's `DiscardBatch` did not
  actually discard under production wiring (latent: the affected error paths
  currently panic the process, which drops the collector).
- `clearFastIndex` re-opened its chunk iterator from the range start and
  relied on the chunk commit applying its deletes — under CollectingDB the
  commit only moves ops to the collector and the iterator reads the
  underlying DB, so clearing ≥ 65536 existing entries looped forever,
  re-staging the same keys unboundedly (sampled runs: 262k ops for 70k
  entries within a second; 68.8M within 30s). This also affected
  `Import` → `dropFastIndex`.
- `ImmutableDB.NewBatch()` returned nil, a latent panic for any consumer that
  constructs batches eagerly (bptree, IAVL's BatchWithFlusher).
- The restart window: `querySnapshot` was only installed by `Commit()`, so a
  restarted node's queries fell back to the live DB until its first block.

## Decision

Layered defense — any one of the first three layers stops the incident:

1. **Read-only loads on immutable stores** (`bptree.MutableTree.LoadReadonly`,
   used by `store/bptree` Immutable branches): query-path loads perform NO
   fast-index maintenance. Read paths cannot write, on any DB routing. Fast
   READS keep working (advisory `fastGet` with its version guard).
2. **Stale-reader guard** (`ensureFastIndex`): rebuild only when the stamp is
   missing or BEHIND the loaded version. The stamp commits atomically with
   each version's records, so a stamp AHEAD of the loaded version means the
   observer is a stale reader racing a newer commit; rebuilding from its
   older root is exactly the poisoning. Skips log a warning. This also
   protects direct library users outside the store layer.
3. **Snapshot routing** (`rootmulti.constructStore`): immutable multistores
   read `ms.db` — the frozen post-commit SnapshotDB (or ImmutableDB fallback)
   — even for dedicated-db mounts. Constraint: a dedicated `params.db` must
   be the same physical DB as the multistore's (true for every in-repo
   mount); a genuinely separate merkleized store DB fails loudly at
   LoadVersion's commit-id check (a dbadapter mount would not — its CommitID
   is always zero — so separate physical DBs must not be mounted).
   - Companion (hard prerequisite): `ImmutableDB.NewBatch` returns a shared
     read-only no-op batch (Set/Delete discarded, Write panics) instead of
     nil.
   - Companion: `rootmulti.LoadVersion` seeds `querySnapshot` (non-immutable
     multistores, both the ver==0 and normal exits), giving restarted nodes
     snapshot isolation before their first commit.
4. **`.store` query isolation** (`rootmulti.QueryImmutable` +
   `baseapp.handleQueryStore`): store queries are served from the query
   snapshot via the extracted `immutableAtVersion`; on error (height ≤ 0
   pre-first-commit, pruned heights) the handler falls back to the legacy
   live path, preserving the existing response surface for those corners.
5. **Real batch discard semantics** (`db/collecting.go`): `batchHandle`
   buffers ops locally; `Write` moves them into the collector (order
   preserved, handle reusable); `Close` without `Write` discards — restoring
   the standard `dbm.Batch` contract that `DiscardBatch` and rollback paths
   rely on. `GetByteSize` now measures the handle's own buffer (fixes
   whole-collector readings in flush-threshold loops). Required companion:
   `rootmulti.Commit` calls `metaBatch.Write()` BEFORE draining, keeping
   commitInfo/latestVersion in the same atomic WriteSync as store data.
6. **Terminating index clear** (`clearFastIndex`): each chunk resumes its
   iterator after the last staged key instead of the range start, making
   progress independent of whether staged deletes have been applied. This —
   not layer 5 — is what fixes the ≥64Ki-entry startup hang; `Import` →
   `dropFastIndex` inherits it.

## Alternatives considered

- **Per-store snapshots for genuinely-separate dedicated DBs**: rejected;
  no in-repo mount needs it, and the failure mode without it is loud.
- **Serializing queries with the consensus mutex**: rejected; reintroduces
  the query-blocks-consensus latency #5431 removed, and does not fix the
  design gap (write-capable read paths).
- **Making `MutableTree` thread-safe**: out of scope; the tree's
  single-writer contract is sound once no other goroutine can reach live
  mutable state.
- **Only fixing the TOCTOU (consistent version+stamp read)**: insufficient;
  the query path had no business writing at all, and other interleavings
  (e.g. racing an in-progress clear) would remain.

## Consequences

- The natural-concurrency repro (8 query goroutines vs 1900 commits, which
  persisted hundreds of stale entries per pre-fix run — 290 and 480 in two
  sampled runs, scheduling-dependent) runs clean, including under `-race`.
- `.store` queries at pruned or pre-commit heights now take the legacy
  fallback explicitly; served heights get snapshot isolation. A query for the
  just-committed height can also hit the fallback in the narrow window inside
  `Commit` between the WriteSync and the snapshot refresh (commitInfo durable,
  snapshot one behind) — read-only and no worse than the pre-fix behavior.
  Each `.store` query now pays the same per-query store construction that
  custom queries already paid (including a `discoverVersions` root scan — a
  pre-existing cost, now shared; a future optimization could cache the latest
  version).
- CollectingDB read-your-writes narrowed for BATCH ops: staged batch ops are
  visible to `CollectingDB.Get` only after `batch.Write()` (previously at
  `Set` time). No in-repo consumer read back unwritten batch ops (bptree
  serves its own staging from `pendingVals`/node cache; dbadapter and the
  cache flush write-then-Write); direct Set/Delete visibility is unchanged.
- Multistore lifecycle: anything that loads a multistore over PebbleDB must
  release it (`Close`) before closing the DB — previously only required
  after `Commit`, now also after load (the seeded snapshot). In-repo callers
  (baseapp, gnogenesis fork, tests) were audited and updated.
- Non-snapshot backends (goleveldb, boltdb, lmdb, mdbx) keep UNISOLATED
  query reads over the live DB (ImmutableDB fallback) — but write-proof, and
  layers 1–2 close the poisoning there too. The default backend (pebbledb)
  gets full isolation.
- A statesync/import flow that drops or rebuilds the index through the
  CollectingDB wiring is durable only after the next rootmulti drain; wiring
  that ends without a `Commit()` must flush explicitly.
- Startup rebuilds over CollectingDB still stage the whole index in memory
  until the next drain (unchanged; bounded by index size). Documented cost.
- **Operator remediation for already-poisoned nodes**: the fixes prevent new
  poisoning but cannot detect existing stale entries (their stamp claims
  currency). Delete the fast-index stamp (`s/_/M` ‖ `fastidx` in the app DB)
  to force a full rebuild at the next start, or resync. The same applies
  after any external rollback/DB surgery: rollback tooling must delete the
  stamp, since old-timeline entries become trusted again once the chain
  re-passes their versions.
