# B+32 Tree (`tm2/pkg/bptree`) — Issue Tally

Consolidated tally of **bugs / correctness / robustness / concurrency issues** raised
against the B+32 tree (PR [#5438](https://github.com/gnolang/gno/pull/5438)) across its
review and related fix PRs.

> Scope note: per request this focuses on **bugs**, not optimizations. Pure performance
> findings from the related PRs are listed compactly in [Appendix I](#appendix-i--performance-findings-not-bugs)
> for completeness only.

Compiled from PR descriptions, inline review comments, reproducer tests, and verification
against the current `feat/jae/bp32tree` working tree (2026-06-06).

> ⚠️ **Statuses in §2 below were the FIRST-PASS (inferred) statuses.** They have since been
> verified line-by-line against the code — see **[§V. Verification results](#v-verification-results--recommendations-verified-against-code)**,
> which supersedes them. Several first-pass calls were wrong (notably C1, M1, M2, M5 — corrected below).

---

## V. Verification results & recommendations (verified against code)

**Method.** Every claim was re-checked against the actual code on `feat/jae/bp32tree` (2026-06-06)
by five parallel adversarial code-reading passes (each told to *disprove* the bug and quote
exact code) plus a manual read of `prune.go` and an **adversarial runtime test**: 200 versions
of random insert/delete churn on a height-3 DB-backed tree (≈3,400 live keys), pruning a
15-version sliding window every round, verifying every live key after each prune, plus a
cold-restart reload. **All existing prune tests + the adversarial test pass.**

**The single most important framing fact:** in the standard gno node, every ABCI `Query`,
`DeliverTx`, and `Commit` is serialized by one shared `localClientCreator.mtx`, so store reads
and writes do **not** run concurrently today. That makes the entire *concurrency* class
(Tier C) **real in code but not triggerable in the current deployment** — they become live
only if that serialization is relaxed (async pruning, concurrent/historical query connections,
network ABCI, state-sync, or direct library reuse). The *reachable-today* bugs are the
**value-key lifecycle** bugs (Tier A).

### Tier A — CONFIRMED, reachable today, correctness / data-loss (priority)

| ID | Bug (verified) | Why it bites | Fix source |
|---|---|---|---|
| ~~**H8**~~ ✅ | **FIXED (commit `568bc03b6`).** Was: `LoadVersion(V<latest)`+`Set` did an eager **out-of-batch** `db.Set` into the already-persisted `V+1` namespace, silently overwriting live values. Now values stage in `pendingVals`+batch and `LoadVersion`/`SaveVersion`-non-commit exits `DiscardBatch`, so a stale staged write can't flush over a committed version. | **Bite-proven**: `TestLoadVersion_FailedSaveDoesNotLeakStagedValue` ("LEAK" without the fix). | **commit `568bc03b6`** (root-cause; supersedes the #5570/#5591 mitigations) |
| ~~**H9**~~ ✅ | **FIXED (commit `568bc03b6`).** Was: idempotent `SaveVersion` left session state set; `Rollback` then `DeleteValueDirect`'d those vks, wiping live data. Now `Rollback` discards the *uncommitted* batch (never deletes committed values) and the idempotent path resets session state — closed three independent ways. | Validated: `TestStaging_IdempotentSaveRollbackKeepsData`; Rollback issues **zero** committed-value deletes (structural). | **commit `568bc03b6`** |
| ~~**H6**~~ ✅ | **FIXED (commit `568bc03b6`).** Was: `Set` mutated the tree before the eager value write; a `SaveValue` DB error left a dangling ref. Now `SaveValue` only buffers (map + in-memory batch, **no I/O at stage time**), so it cannot fail mid-`Set`; persistence failure is confined to the atomic `Commit`. | Closed by construction (no stage-time I/O); not independently reproducible with memdb. | **commit `568bc03b6`** |
| ~~**H10**~~ ✅ | **FIXED (commit `74b7ca21b`, adopted from #5591).** `SaveVersion` now uses `versionExistsE`, which propagates the `db.Has` error instead of reading it as "absent"; `VersionExists` logs + returns false for the store-compat surface. Bite-proven: `TestSaveVersion_PropagatesVersionExistsError`. | Transient DB read error on pebbledb/goleveldb. | **commit `74b7ca21b`** (#5591) |
| ~~**H12**~~ ✅ | ~~`immutableForProof` builds proofs from `t.root` (**uncommitted** working tree), not `t.lastSaved`; proofs verify only against `WorkingHash()`, not the committed `Hash()`. No `ErrNoCommittedState` guard.~~ **FIXED TODAY** — cherry-picked #5468 + nits + `newImmutable` dedup. | **Only the `MutableTree` proof wrappers (tests/bench).** The store `Query` path proves via a committed `ImmutableTree` gated by `VersionExists`, so **production is unaffected** — a `MutableTree`-API/test-hygiene fix, not a query-security bug. | **#5468** (cherry-picked to branch) |
| ~~**M6**~~ ✅ | **FIXED (commit `74b7ca21b`, adopted from #5570 finding:8).** `MutableTree.Iterate` and `ImmutableTree.Iterate` now capture and return the resolver error instead of swallowing it. Bite-proven: `TestIterate_PropagatesResolverError`. *(`IterateRange` keeps its IAVL-parity `bool` signature per #5570 — test-only caller; the store's query path uses the streaming `Iterator`, which records `Error()`.)* | Any value-store read fault mid-iteration. | **commit `74b7ca21b`** (#5570) |

### Tier B — CONFIRMED, reachable, space leaks / robustness (not honest-state corruption)

| ID | Bug (verified) | Fix source |
|---|---|---|
| ~~H4~~ | **NOT A BUG (corrected).** The `vRootNK==nil` early-return skips `LoadOrphans(nextV)`, but orphans are recorded *only* on overwrite/remove of an existing value (`orphanValueKey`, `mutable_tree.go:115,201`). An empty V→V+1 is pure insertion → `orphans[V+1]` is provably empty → nothing to leak. #5570 finding:2's premise is unreachable. | — |
| ~~H11~~ ✅ | **FIXED (commit `568bc03b6`)** — values are no longer written to the DB until `Commit`, so an uncommitted `Set` persists nothing and a crash leaks nothing (no startup scan needed). Bite-proven: `TestStaging_UncommittedSetDoesNotPersist`. | commit `568bc03b6` |
| ~~H13~~ ✅ | **FIXED (commit `568bc03b6`)** — `SaveValue` now stages into the batch (flushed atomically at `Commit`), not a direct out-of-batch `db.Set`. Bite-proven via the eager-revert experiment + updated `TestRollback_CleansUpValues`. | commit `568bc03b6` |
| ~~L1~~ ✅ | **FIXED (`1db08015f`).** `PruneVersionsTo` now returns early when `toVersion < first`, so an out-of-spec arg can no longer rewind the version floor (or spuriously flush mid-session staged writes via the old fall-through `Commit`). Bite-tested: `TestPrune_BelowFloorDoesNotRewindFirstVersion`. | #5570 finding:41 |
| ~~L2~~ ✅ | **FIXED (`fc5451526`).** Prune now `DiscardBatch`es on any error (no partial deletes survive for a later `Commit` to flush — previously a mid-prune DB error → half-deleted version with a live root ref → unreadable AND unprunable → `Store.Commit` panic crash-loop). Guarded by a clean-session precondition (`ErrUncommittedChanges`; four-term check incl. `nextValueNonce` for net-zero sessions — a guard hole found+proven in review). Bonus: `findCorrespondingChild` now propagates node-load errors instead of panicking or swallowing them into "orphaned" (which could delete subtrees shared with the successor). Bite-tested with DB-error injection. | #5570 finding:52 |
| ~~L3~~ ✅ | **FIXED (`1c5d72670`)** — `deleteAllNodesForVersion` removed entirely; the `nextV==0` case now returns an error (unreachable by invariant: `toVersion < latest` guarantees a successor), so no path deletes nodes without processing value-orphan lists. **Verified complete by the H1+L3+M9 adversarial pass**, which found the adjacent **M22** (below) — fixed in the same pass. | #5570 finding:43 |
| ~~**M22**~~ ✅ | **NEW (found by the L3 verification pass) + FIXED same pass.** Pruning decided version existence via the error-swallowing `VersionExists` at two **destructive** sites, violating the H10 norm: (a) the pruneRange skip — a transient `db.Has` error read an existing version as absent, skipping it while its successor was pruned against a later version, deleting records the skipped version still shares (bricked: unreadable AND unprunable); (b) `findNextVersion` — an error on the true successor silently pruned against a later version, deleting records and orphan-listed values a RETAINED version still uses. Both now use `versionExistsE` and abort through the L2 DiscardBatch funnel (nothing staged, floor unadvanced, retry clean). Bite-proven both modes: `TestPrune_PropagatesVersionExistsError` fails on the unfixed parent. All other `VersionExists` callers audited non-destructive. | this review |
| ~~L5~~ | **DOWNGRADED — not a bug post-H2/H3** (3-agent adversarial pass, tripwire-verified). Under the contract no `GetNode` can observe the evict→flush window: the prune never re-reads keys it deleted (whole suite run under a panic-tripwire: zero hits), and registered readers of retained versions never reference deleted keys (dual-walk/COW). A stale entry can form only via the accepted unregistered-store-view gap, is unreachable by any retained/future tree (`SaveNode`'s cache `Add` overwrites every reused key on Import), ages out of the LRU, and serving it to that out-of-contract reader beats "node not found". Decisive: #5570's `pendingEvicts` would NOT fix the one constructible wrong-data case (a straddling reader's cache-Add lands post-flush regardless of eviction timing) — only serialization does. | #5570 finding:44 |
| ~~**M19**~~ ✅ | **FIXED (`c15626f6e`) — the "twin" node-record leak (the leak half of C1, found by the C1 fuzzer-design review).** `walkAndPrune` decided "shared" by HASH, but COW shares by NodeKey record: content-identical records under new NodeKeys (net-zero sessions, same-value rewrites, Import twinning a whole tree) were skipped-but-unreferenced → permanently unreachable node records (~linear growth; import-then-prune leaked 100%). Fix: sharing by NodeKey identity at all 3 comparison sites; soundness via the ref-contiguity invariant, made unconditional by the idempotent-SaveVersion adoption fix (below). Bite-proven (exact garbage oracle); zero over-deletions across churn oracles. | this review |
| ~~**M20**~~ ✅ | **FIXED (`c15626f6e`).** (a) Clean `LoadVersion(old)` + prune covering `old` succeeded and bricked the working view (next Get panics) — the working tree is an unregistered reader of its loaded version; now rejected with `ErrActiveReaders` (`t.version <= toVersion`). (b) The idempotent `SaveVersion` path kept the in-memory REPLAYED tree after a hash-only check — a divergent same-hash replay forked the record lineage (would make the M19 fix OVER-delete) and already corrupted HEAD three ways (foreign/never-persisted valueKey resolution, replayed orphan lists deleting live values); it now ADOPTS the persisted version's root. Plus: Import rejects zero-key markers. Regressions: divergent-replay values/nodes variants, collision-delete replay, import-gap replay. | this review |
| **M21** ⚠️ | **NEW (filed by the twin-fix review): import-then-prune leaks the pre-import versions' VALUE records.** Values are reclaimed only via orphan lists, and an Import displaces nothing, so pre-import values appear in no list — **bounded, one-time per import, non-compounding** (proven: 200→200 after 5 further prune cycles). Latent (no production Import caller). Fix at state-sync wiring (value reconciliation or import-into-fresh-DB requirement), together with the M18 precondition (`version == latest+1`, which also covers the Import/SetInitialVersion version-gap + `latestVersion`-rewind quirks). | this review |
| **M23** ⚠️ | **NEW (filed by the H1+L3+M9 verification pass): Import performs no separator ORDERING/consistency validation.** `Importer.Add` validates separator count and length but not that separators are strictly increasing and consistent with their children's key ranges — a malicious export stream can build a structurally search-broken tree that still hashes. Latent (no production Import caller); fix with the M18/M21 import-hardening cluster when wiring state-sync. | this review |
| **M18** ⚠️ | **NEW (found by the L5 pass): `Import` to a PRUNED version number corrupts retained versions.** `Import(version)` only checks `!VersionExists(version)` (import.go:32) — a pruned version qualifies — and the import then reuses that version's node/value key namespaces `{v, nonce}`, overwriting nodes/values still **shared into retained versions**: empirically 99/100 wrong keys in a retained version from a fresh handle (silent; reads don't verify valueHash), plus `latestVersion` rewound below `firstVersion`. **Latent** — `Import` has no production caller (no state-sync). Fix when wiring state-sync: require `version > latestVersion` (or an empty version space) in `Import`. | this review |
| ~~L6~~ ✅ | **FIXED (`1db08015f`).** `orphanValueKey` propagates `DeleteValueDirect`'s error through `Set`/`Remove` (was swallowed). Near-unreachable post-staging (no I/O at stage time), but no longer silent. | #5570 finding:25 |
| ~~L7~~ ✅ | **FIXED (`8208eb3c1`).** `Import` now calls `t.Rollback()` first, discarding any uncommitted working-session state (pending batch, orphan list, value-nonce, working root) so it can't leak into the import's `SaveVersion`. (#5570's "reserve nonce 0" was confirmed unnecessary on HEAD — versions are always ≥1, so a value key is never the all-zero sentinel; `allocValueKey` already uses nonce 0 routinely.) Bite-tested: `TestImport_RollsBackPendingSession`. | #5570 finding:39 |
| ~~M9~~ ✅ | **FIXED (`06de00540`)** — `MaxKeyLen` (= `maxReadBytesLen`, 1 MiB) enforced at every write ingress: `Set` (`ErrKeyTooLong`) and `Importer.Add` (leaf keys AND inner separator keys; empty separators also rejected). Separators only ever derive from capped leaf keys (splits/redistributions copy, never lengthen), so the cap is transitive. **Verified complete by the H1+L3+M9 adversarial pass**; added `TestSet_MaxKeyLenRoundTripsFromDisk` (pins the `MaxKeyLen == maxReadBytesLen` boundary through SaveVersion + cold reload) and `TestImport_RejectsSeparatorOverMax`. Bonus hardening from the pass: `LoadOrphans` validates the untrusted count against bytes present before allocating. | #5591 (5f091db7e) |
| M15 ⚠️ | `Exporter`: abandon-without-`Close()` **permanently leaks the version reader** (decr is only in `Close`) + leaks the goroutine (>32 nodes). **Escalated by H2/H3**: the pinned reader now makes the store's auto-prune *panic* (`ErrActiveReaders`). **Mitigated (option b):** must-`Close` lifecycle documented on `Exporter`/`Export`; `Export` has **no production caller** (no state-sync), so it's contract-only. Structural fix (pull-based, no goroutine — option a) deferred until state-sync exists. | #5442 bug5 |
| ~~M8~~ ✅ | **FIXED (`1db08015f`).** `LoadVersionForOverwriting`/`DeleteVersionsFrom` return a wrapped `ErrUnsupported` (new sentinel) instead of panicking. Still unsupported by design (would leak nodes/values). | #5570 finding:12 |
| ~~M17~~ ✅ | **FIXED (`0de551f17`).** `getChild` memoized every loaded child on the node and `saveNode` never cleared it → the working-tree graph pinned every node touched since the last reload, growing unbounded toward the whole tree (**OOM at scale**, single-threaded, independent of the bounded LRU). Now getChild doesn't memoize and `saveNode` clears `childNodes` post-persist → working tree bounded to root + LRU. Bite-proven: `TestGetChild_WorkingTreeBoundedAfterSave` (pinned count = 1). | this review |

### Tier C — CONFIRMED in code, NOT triggerable under the current ABCI single-mutex (latent)

| ID | Bug (verified) | Fix source |
|---|---|---|
| ~~H1~~ ✅ | **FIXED (`ac3d2756f`)** — DB-backed snapshot iterators (`ImmutableTree.Iterator`, `NewIteratorWithNDB`) register as version readers at construction and release on idempotent `Close`; working-tree iterators stay unregistered by design (protected by `toVersion >= latest` + the M20 working-tree guard). **Verified complete by the H1+L3+M9 adversarial pass**: all 3 constructor sites, incr/decr balance (incl. tree+iterator double-hold via the store Query path, both released), bite re-proven (`TestPrune_IteratorBlocksPrune`/`TestPrune_StoreIteratorBlocksPrune` fail with the incr calls removed). Pass also unified the `version > 0` registration guard across iterators/Export (a version-0 registration was never released — stale map entry) and made the store `/subspace` query Close deferred. | #5450; #5570 finding:1; #5591 |
| ~~H2~~ ✅ | **FIXED (`a07f4b3ce`).** `GetImmutable`/`immutableForProof` now `incrVersionReaders` (before `GetRoot`) and the caller / proof wrapper `Close`s (idempotent `ImmutableTree.Close`); `GetVersioned` defers Close. Long-lived store snapshots use non-registering `GetImmutableUnregistered` (gap accepted, dormant). | #5570 finding:30/40 |
| ~~H3~~ ✅ | **FIXED (`a07f4b3ce`).** `pruneMu` RWMutex: `incrVersionReaders` RLocks it; `beginPruning` holds it exclusively across the whole prune, so no reader can register a to-be-deleted version. Replaces the `hasVersionReaders` TOCTOU. | #5450; #5570 finding:15 |
| ~~M11~~ ✅ | **FIXED (`0de551f17`).** `getChild`'s write-back mutated shared nodes on reads (racing the writer's COW `Clone`). Now reads never memoize; writes clone-first → shared nodes immutable; `childMu` removed. `-race`-verified (concurrent `Has`+`Set`). | #5570 finding:7 |
| ~~M12~~ ✅ | **FIXED (`0de551f17`).** Same mechanism — the aliased root is no longer mutated in place (`miniTree`/`childHashes` rebuilt only on clones), so a concurrent proof reads stable fields. | #5570 finding:9 |
| ~~M13~~ ✅ | **FIXED (`27c6ebbe5`).** `GetNode` cache-miss loads coalesce via a `singleflight.Group` keyed on the NodeKey — one deserialize + cache Add, shared instance. (Benign post-getChild-fix: read-path nodes are immutable + unmemoized, so it was dedup/efficiency, not a race.) | #5570 arch |

### Tier D — Hardening gaps (corrupt-DB / malicious-input only; not honest-path bugs)

| ID | Bug (verified) | Fix source |
|---|---|---|
| ~~M3~~ ✅ | **FIXED (`83f5ee041`)** — `ReadNode` rejects payloads with trailing bytes (fail-fast on corruption). | #5591 (c43bf6a4b) |
| ~~M4~~ ✅ | **FIXED (`27090f5b0`)** — `Serialize` fails fast on nil valueKey / child ref (write-side size checks before persisting). | #5570 finding:18; #5591 |

### Tier E — NOT bugs / already fixed / overstated (corrections to first-pass §2)

| ID | First-pass said | **Verified reality** |
|---|---|---|
| **C1** | ❌ critical prune bug "still present" | **Original #5442 bug6 is FIXED by #5451** (from-root `findCorrespondingChild`); **the residual was then settled by the twin-fix work (`c15626f6e`)**: the LEAK half of the heuristic was real (M19, hash-vs-record-identity — fixed), and the OVER-DELETION half now rests on a proven invariant instead of a heuristic — sharing is record identity, sound by ref-contiguity (made unconditional by the idempotent-adoption fix) + routing completeness (verified by review: first-key routing to intrinsic height lands exactly on any record the successor references). Still O(changed-nodes/block); do **not** swap in #5570's O(tree-size) mark-and-sweep. **The C1 fuzzer is now LANDED (`cdd739f32`)**: coverage-guided op programs interleaving prune with everything (twin-makers, waves, holds, LoadOld, cold restarts, imports, DB-error injection), exact garbage/integrity/proof oracles after every prune, an env-gated indefinite soak (551k ops/10s smoke, bounded by construction), and a seeded `-race` reader stress — validated to catch M19+M20 instantly on the pre-fix parent and quiet on HEAD (104k execs/60s). C1 is **closed**: proven invariant + fix + continuously-runnable assurance. |
| M1 | ❌ numKeys bounds unchecked | **NOT A BUG** — range check `numKeys<0 \|\| >B-1`(inner)/`>B`(leaf) is present right after the cast, before any array use (`node.go:265,314`). |
| M2 | ⚠️ readBytes OOM | **NOT A BUG** — `length > maxReadBytesLen` (1 MiB) guard precedes `make` (`node.go:372`). |
| M5 | ❌ GetNode errors swallowed in prune | **ALREADY FIXED** (commit `585ae1e34`) — all delete-decision sites propagate wrapped errors; since L2 (`fc5451526`) `findCorrespondingChild` also propagates node-load errors (nil is returned only for genuinely-orphaned routing). |
| M10 | ✅ import bounds | **Confirmed FIXED** (#5451) — `nk>B`/`>B-1` guards at `import.go:70,93`; later strengthened to also reject `numKeys < 1` zero-key markers (M20). |
| M14 | ⚠️ resolveValue returns hash | **Dangerous core FIXED** (commit `cc2c7a7a6`): both `resolveValue` now return errors, not `vh[:]`. Only residue: no dedicated `ErrNoValueResolver` (conflated with `ErrKeyDoesNotExist`), surfaces only for an existing key with no resolver wired (misconfig, not corruption). |
| L4 | ❌ nodeKeyBytesToArr zero-pads | **NOT A BUG** — no such helper exists; `GetNodeKey` returns nil on wrong length. |
| L8 | ❌ ImmutableTree.Close not idempotent | **NOT A BUG** — at first-pass time `ImmutableTree.Close` did not exist; H2 (`a07f4b3ce`) later added it, idempotent via `sync.Once` + `registered` gate (double/concurrent Close cannot over-decrement). |
| H5 | H — all-zero ValueKey collision | **Overstated/PARTIAL** — needs `Version==0` (non-default; default first version is 1) AND nothing actually reads the zero placeholder as "missing" (effectively dead branch). |
| H7 | H — SaveVersion resets counters before Commit | **NOT A BUG** — reset happens strictly **after** the `Commit()` success check; a failed Commit correctly leaves counters intact for retry (claim was backwards). |
| M7 | M — DeleteVersionsFrom leaks nodes | **NOT A BUG (as stated)** — unsupported by design: it returns a wrapped `ErrUnsupported` (panic→error via M8, `1db08015f`) rather than leaking; see M8. |
| Lat1 | latent — Clone shares slices | Confirmed **latent only** — `c:=*n` aliases key slices, but no path mutates key bytes in place (all writes are slot-reassign/`copyKey`). COW-safe today. |

### Score

- **Reachable correctness bugs remaining: 0.** ✅ H6/H8/H9 (`568bc03b6`); H10/M6 (`74b7ca21b`); H12 (#5468, cherry-picked `cebeecf9a`).
- **Reachable leaks/robustness remaining: 0 of the original Tier B.** ✅ H11/H13 (`568bc03b6`); ✅ **M9** (`06de00540`) + ✅ **L3** (`1c5d72670`), both verified complete by the H1+L3+M9 adversarial pass; ✅ **L7** (`8208eb3c1`); ✅ **L1/L6/M8** (`1db08015f`); ✅ **L2** (`fc5451526`); ⚠️ **L5** downgraded (not-a-bug post-H2/H3); ⚠️ **M15** documented-contract (structural fix at state-sync). **NEW from the fuzzer-design + twin-fix reviews:** ✅ **M19** (twin node-record leak — the real leak half of C1) and ✅ **M20** (loaded-version prune brick + divergent-idempotent-replay corruption) fixed (`c15626f6e`); ✅ **M22** (prune decided existence via error-swallowing `VersionExists` at two destructive sites — found and fixed by the H1+L3+M9 pass); ⚠️ **M18** + ⚠️ **M21** + ⚠️ **M23** latent Import issues (key-namespace reuse; pre-import value wall; separator ordering unvalidated) — guard `version == latest+1` + value reconciliation + structural validation when wiring state-sync; no production Import caller today.
- **Latent concurrency (dormant under ABCI mutex): 0 of the original Tier C.** ✅ **H1** (`ac3d2756f`, verified complete by the H1+L3+M9 adversarial pass); ✅ **M11/M12** (`0de551f17`); ✅ **pendingVals** (`75c946820`, §VIII); ✅ **H2/H3** (`a07f4b3ce`, §IX); ✅ **M13** (`27c6ebbe5`, `GetNode` singleflight). Single-tree node reads, value resolution, prune-vs-reader, AND concurrent node loads are now safe against a writer. **Residual (addressed by contract — §X):** the pre-existing `MutableTree.version`/`lastSaved`/`root` read-vs-write field race is now documented as a single-goroutine contract + named `-race` guard (`446e4a6ad`, option a); an `RWMutex` on `MutableTree` (option b) is the only remaining promotion-time work.
- **Memory / liveness: 0.** ✅ **M17** (unbounded working-tree memory → OOM at scale) fixed (`0de551f17`): getChild memoization removed + clear-on-save bound the working tree to the nodeDB LRU, matching IAVL.
- **Hardening: 0.** ✅ **M3** (ReadNode trailing bytes) + **M4** (Serialize nil-ref) landed via #5591 cherry-pick.
- **Disproven / already-fixed / overstated: 13** (the 12 Tier E rows incl. the C1 ship-blocker and M1/M2/M5, plus Tier B's H4).

### Where things stand

1. **The headline "ship-blocker" (C1) is not reproducible.** #5451's fix holds against heavy adversarial churn; the remaining concern is *provability*, not a known failure.
2. **The value-key subsystem — the main reachable risk — is now fixed at the root** (commit `568bc03b6`). The eager, out-of-batch `db.Set` (keyed by a per-version nonce that reset to 0 and was never reseeded on load) was the common cause of H6/H8/H11/H13 and H9's teeth. Values now stage in `pendingVals`+batch and flush atomically at `Commit`; `Rollback`/`LoadVersion`/`SaveVersion`-non-commit paths `DiscardBatch`. ✅ H10 (`VersionExists` error) and M6 (`Iterate` errors) are now fixed too (commit `74b7ca21b`, adopted from #5591/#5570) — **all reachable value-key correctness bugs are now closed.**
3. **All the concurrency findings are dormant** because of the global ABCI mutex — correct today, fragile for any future relaxation, and the package documents no single-goroutine contract.
4. **Most fixes exist in PRs, but cherry-pick — don't take #5570 wholesale**: #5591 + #5468 cover essentially all of Tier A. #5570 covers Tier A+B+C+D, **but its prune rewrite (mark-and-sweep) is O(tree size) per prune and will not scale** (its benchmark is only 10K keys); lift its prune *locking* + *leak* fixes (H3/L2/L3/L5) and the non-prune findings, but **keep the existing dual-tree-walk**. It's also large, stacked, unmerged, with a noted `LoadVersion` perf regression.

### What should be done — options

**Option 1 — Land the existing fix PRs (most complete).** Rebase + merge **#5591** (15 correctness/safety fixes) and **#5468** (proof root), then **#5570** (the 52-issue pass, incl. mark-and-sweep prune + reader registration). Pros: closes every confirmed item incl. the latent concurrency class and de-risks C1 with a *provable* algorithm. Cons: large surface to review; stacked on diverged branches; #5570's `LoadVersion` perf regression needs resolution first.

**Option 2 — Targeted minimal patch (reachable-only, fastest to safe).** ✅ **Done.** `568bc03b6` (value staging) closed H6/H8/H11/H13 + neutered H9; `74b7ca21b` closed H10 + M6; #5468 closed H12; `06de00540` closed M9 (write-side key-length cap). No reachable items remain; Tier C was subsequently fixed anyway (§VIII–§X) and the C1 pruning question settled (twin fix + fuzzer).

**Option 3 — Harden the existing pruner; do NOT swap in mark-and-sweep.** ⚠️ #5570's content-addressed mark-and-sweep builds the full reachable NodeKey set of the retained version on every prune → **O(tree size)** memory + disk reads (millions of node KVs at 100M keys), vs. the current dual-tree-walk's **O(changed-nodes/block)** from its same-hash subtree short-circuit (`prune.go:134`). #5570's prune benchmark (10K keys / 50 versions) is too small to expose this; it would not scale in production. Since C1 isn't reproducible, **keep the dual-tree-walk and prove/fuzz it** (random + adversarial split/merge/prune/restart, like the test in §V), and lift only #5570's prune **locking** (H3, `pruneMu`) and **leak** fixes (L2/L3/L5) — not the algorithm swap.

**Option 4 — Document the concurrency contract.** If the ABCI single-mutex invariant is guaranteed, add explicit "MutableTree is single-goroutine; ImmutableTree concurrent-read only under serialization" docs and a `-race` guard test, and consciously defer Tier C. Cheapest; valid only while the invariant holds — revisit before async prune / concurrent queries / state-sync.

**Recommended sequence:** Option 2 now (stop the reachable data-loss) → Option 3 (harden — don't replace — the pruner) → Option 1 to absorb the non-prune fixes → Option 4 docs as the safety net for Tier C. Add the adversarial prune test + value-lifecycle regression tests alongside. **Cherry-pick #5570's findings individually; do not take its mark-and-sweep prune.**

---

## VI. #5591 cherry-pick landed (commits `ac3d2756f..d63d02387`)

Cherry-picked the 12 still-wanted #5591 commits onto the branch; **dropped 3**:
`536f84a72` (redundant — H10 already on branch), `82eebc957` (redundant — branch
clears session state via `resetSession`; see §1 of this doc), `388068894`
(moot — value staging removed eager writes; its test would fail). Three conflicts
resolved by hand (all in the SaveValue/GetValue/errors area the staging+H10 work
rewrote): `newImmutable` now sets `imm.ndb` so snapshot iterators register as
readers; `GetValue` keeps the `pendingVals` read-your-writes check then layers
the missing-vs-empty `Has` disambiguation; `errors.go` keeps both
`ErrNoCommittedState` and `ErrKeyTooLong`. One cherry-picked test
(`TestGetValue_MissingValueReturnsError`) was adapted to commit before simulating
corruption (staging keeps the value in `pendingVals` until `SaveVersion`).

Closed by this cherry-pick: **H1** (iterators register as version readers —
`ac3d2756f`, iterator-blocks-prune tests pass), **M3** (`ReadNode` trailing
bytes — `83f5ee041`), **M4** (`Serialize` nil-ref — `27090f5b0`), **M9**
(`MaxKeyLen` on Set/Import — `06de00540`), **L3** (unreachable
`deleteAllNodesForVersion` removed — `1c5d72670`). Plus robustness/perf:
GetValue missing-vs-empty, orphans[v] first-version edge, separator `copyKey`,
`saveNode` no-force-load, prune batch-memory bound, deferred key copy. Full
suite + benchmarks + store wrapper pass; full `-race` (33s) clean.

Caveat carried forward: **H3** (prune-reader TOCTOU) and the rest of Tier C are
**not** addressed here (they need #5570's `pruneMu`); **H2** is only partially
addressed (snapshot *iterators* register, but a Get-only/proof snapshot still
does not). `4dd84b894`'s commit message advertises a 4 MiB flush default that is
dead under `DefaultOptions`' 100 KiB — behavior is correct, message is stale.

---

## VII. getChild non-memoization + clear-on-save (commit `0de551f17`)

Adopts IAVL's bounded-memory model and removes the getChild read-path data race
(M11/M12) in one change:

- **`getChild` no longer memoizes loaded children** (returns them without
  storing back) → reads never mutate the node; the working tree is bounded by
  the nodeDB LRU instead of pinning every node ever touched (closes the
  unbounded-memory / OOM issue, **M17**).
- **`saveNode` clears `childNodes[i]`** once `children[i]`/`childHashes[i]` are
  durable.
- **`SaveNode` sets `inner.ndb`** so in-memory-built nodes (root splits,
  `splitInner`, merge, import) can lazy-load after the clear; without it the
  next `Set`/`Remove` on a saved-but-not-reloaded tree panics
  ("inner node has nil child"). The load-bearing companion change.
- **`childMu` removed** — reads are pure and writes clone-first, so shared nodes
  are immutable; the mutex guarded a write that no longer exists.

Closes **M11** (concurrent `Get`+`Set` node-field race — `-race`-verified with a
concurrent `Has`+`Set` test) and **M12** (the `immutableForProof` aliased root
can no longer be torn, since shared nodes are never mutated in place).

**Trade-off (benchstat, identical workload):** reads re-fetch loaded children
from the cache rather than following a memoized pointer — `Get` +44% worst-case
(synthetic random-read loop, tree fully cached), `Proof` +6%, `Iterate` +31%;
**`BlockCommit` unchanged** (the dirty path is still memoized via `setChild`).
bptree's shallower tree (~6 levels vs IAVL's ~28) keeps it ahead of IAVL on
reads even without memoization, and IAVL never had this memoization.

**Residual (now closed — see §VIII):** this getChild change was a *partial*
concurrency step; a concurrent value-resolving read (`Get`/proof) still raced on
the `nodeDB.pendingVals` map (`SaveValue` write vs `GetValue` read). That
residual is fixed in §VIII (`75c946820`).

---

## VIII. pendingVals value-resolution race (commit `75c946820`)

Closes the residual from §VII. A concurrent value-resolving reader on a
**committed** snapshot (`Get` / proof / `Export` / snapshot iterator) read the
`nodeDB.pendingVals` map (via `GetValue`) while the writer's `SaveValue` wrote it
→ `fatal error: concurrent map read and map write`.

`pendingVals` is the uncommitted working-session buffer, keyed by the **working**
version (`allocValueKey`), so a committed snapshot's lookups never legitimately
hit it. Fix: confine `pendingVals` to the single-writer working session and
resolve committed snapshots **DB-only**:

- `nodedb.go`: `getCommittedValue` (DB read + missing-vs-empty `Has`
  disambiguation); `GetValue` = pendingVals check → `getCommittedValue`, now used
  only by the working tree's read-your-writes.
- `newImmutable(root, version, committed)`: `GetImmutable`/`immutableForProof`
  DB-only; `Snapshot` (wraps the LIVE working tree) keeps read-your-writes;
  exported `GetCommittedValueByKey`.
- `iterator.go` `Value()` resolves via the per-source `valueResolver`
  (working-tree → pendingVals, committed snapshot → DB-only); `export.go`,
  `proof.go`, `store/bptree` route committed reads DB-only.

**No locks, no perf hit** — committed reads drop an always-missing map probe and
do the same `db.Get`; the DB handles reader/writer concurrency. Working-tree
read-your-writes unchanged (single goroutine). `-race`-verified (concurrent
writer + committed Get/proof/iterator readers; fails on the prior HEAD).

With this, **single-tree node AND value reads are concurrent-safe against a
writer.** The LRU `nodeCache` is internally `RWMutex`-locked (not a data race).
Remaining Tier C: H2/H3 (below) and M13 (`GetNode` cache-miss singleflight).

---

## IX. H2 + H3 — snapshot version-reader registration + atomic prune (commit `a07f4b3ce`)

Closes the read-vs-prune gap: a prune could delete nodes a concurrent `Get`/proof
snapshot was still walking. Dormant under the ABCI single-mutex; before-promotion
hardening. Mechanism mirrors #5570 (whose stale base we did NOT cherry-pick).

**H2 — register Get/proof snapshots.** `ImmutableTree.Close()` (idempotent via
`sync.Once`, gated on a `registered` flag); `GetImmutable`/`immutableForProof`
`incrVersionReaders` **before** `GetRoot` (closes the reader-side TOCTOU), with
decr on error paths; the proof wrappers and `GetVersioned` `defer imm.Close()`.
A new `GetImmutableUnregistered` serves the store's **long-lived** immutable
`LoadVersion` view, which has no Close hook — registering it would pin the
version against pruning forever (and `Store.Commit`'s prune **panics** on
`ErrActiveReaders`, so a leak is a hard crash, not a silent gap). The store-level
read-vs-prune gap there is accepted (dormant) and documented; **no `Store.Close`**
added — the `Store`/`MultiStore` abstraction has no `Close` to wire it through
(rootmulti builds immutable views via `LoadVersion`, never closed).

**H3 — atomic prune.** `nodeDB.pruneMu sync.RWMutex`: `incrVersionReaders` takes
`pruneMu.RLock()` (stays `void` — no caller ripple); `beginPruning(first,to)`
takes `pruneMu.Lock()`, checks reader counts under `mtx`, then holds `pruneMu`
for the whole prune so no new reader can register a to-be-deleted version;
`endPruning` releases it. A **separate** `pruneMu` (not `mtx`) is required because
the prune body itself takes `mtx` via `getFirstVersion`/`setFirstVersion` (holding
`mtx` across the prune would self-deadlock). Replaces the old `hasVersionReaders`
TOCTOU check.

Keeps HEAD's dual-tree-walk prune + `getCommittedValue` (NOT #5570's
mark-and-sweep). `-race`-verified: snapshot-blocks-prune-until-Close, no leaked
reader after `GetVersioned`/proof, `Snapshot.Close` doesn't under-decrement a live
reader, unregistered snapshot doesn't block prune, concurrent prune-vs-reader.
Adapted 7 existing tests that opened an iterator/exporter via a now-registering
`GetImmutable` to also Close it.

**Residual:** none in the original Tier C — M13 (`GetNode` cache-miss singleflight)
is now also fixed (`27c6ebbe5`). The one before-promotion item — the pre-existing
`MutableTree.version`/`lastSaved`/`root` read-vs-write field race — is addressed
by contract in §X.

---

## X. MutableTree single-goroutine contract (commit `446e4a6ad`, Option 4 / option a)

The pre-existing `MutableTree` field race — `immutableForProof`/`Hash`/`Version`/
`Get`/etc. read `t.version`/`t.lastSaved`/`t.root` while `SaveVersion` writes them
— is the one concurrency item that is NOT a node/value/prune race closed by the
earlier fixes. It manifests only if a caller violates the single-writer contract;
the gno ABCI connection mutex already serializes Query vs Commit, so it is dormant.
No PR fixes it (verified: #5570's 52-issue pass adds no `MutableTree` mutex either).

Closed by **contract, not a lock** (option a): the `MutableTree` type now documents
that it is single-goroutine — its mutators and working-tree reads touch the
unlocked working-tree fields and must not run concurrently — and that concurrent
reads of a committed version must use `GetImmutable` (safe to call concurrently;
returns a concurrent-read-safe `ImmutableTree`) or `GetVersioned`/
`GetCommittedValueByKey`/`VersionExists`/`AvailableVersions`. A named `-race` guard
(`TestContract_ConcurrentSnapshotReadsVsWriter_NoRace`) encodes the sanctioned
pattern; the surface is also covered by the getChild/pendingVals/H2/M13 race tests.
Reviewers confirmed the partition both ways empirically (safe set `-race`-clean;
`tree.Hash()`/`tree.Version()` called directly DO race).

**Deferred (option b):** an `RWMutex` on `MutableTree` guarding `root`/`version`/
`lastSaved` — only needed when bptree comes off the ABCI mutex (concurrent
queries / async prune / network ABCI); it adds lock overhead to the single-writer
hot path, so it waits for promotion.

---

## 0. Two different trees — don't conflate

| Tree | Package | Purpose | PRs |
|---|---|---|---|
| **Consensus B+32 tree** (this doc) | `tm2/pkg/bptree/`, `tm2/pkg/store/bptree/` | Versioned, Merkle, ICS23 — drop-in IAVL replacement for chain state | #5438, #5442, #5451, #5450, #5468, #5570, #5591, #5571 |
| Userspace Gno tree (out of scope) | `examples/.../p/.../bptree` | In-realm `avl`-style container | #5475, #5644 |

All bugs below are in the **consensus** tree. #5475 (merged, the userspace tree) and #5644
(API change `Get` → `nil`) are a *separate package* and carry no bug reports — noted only to
avoid confusion.

---

## 1. Related-PR map

| PR | Title | Author | State | Base → Head | Role |
|---|---|---|---|---|---|
| [#5438](https://github.com/gnolang/gno/pull/5438) | immutable B+32 tree — drop-in IAVL replacement | jaekwon | **open** | master ← feat/jae/bp32tree | The implementation under review |
| [#5442](https://github.com/gnolang/gno/pull/5442) | tests showcasing bugs — **DO NOT MERGE** | clockworkgr | closed | bp32tree ← test/bp32tree-review | **6 canonical bugs w/ reproducers** |
| [#5451](https://github.com/gnolang/gno/pull/5451) | Fixes and improvements | clockworkgr | **merged ✅** | bp32tree ← feat/alex/bptree-fixes | Bug-6 partial fix, import bounds, round-trip |
| [#5450](https://github.com/gnolang/gno/pull/5450) | bptree validation | notJoon | closed (unmerged) | bp32tree ← feat/bptree-validation | 3 robustness bugs |
| [#5468](https://github.com/gnolang/gno/pull/5468) | use lastSaved in immutableForProof | notJoon | **open** | bp32tree ← fix/bptree-proof-committed-state | Proof-from-uncommitted-state bug |
| [#5570](https://github.com/gnolang/gno/pull/5570) | correctness/robustness/perf — **52 issues** | clockworkgr | **open** | bp32tree ← bp32tree-second-pass | 52 findings + 2 second-pass |
| [#5591](https://github.com/gnolang/gno/pull/5591) | deep-dive: correctness/safety/perf | clockworkgr | **open** | bp32tree ← bp32tree-deep-dive | 15 findings |
| [#5571](https://github.com/gnolang/gno/pull/5571) | leaf v2/v3 + cache + 13-fix pass | clockworkgr | **open** | second-pass ← bp32tree-advanced | Features + self-review bugs |

**Merge reality:** only **#5451** and **#5475** are merged. The large correctness PRs
(#5570, #5591, #5450, #5468) are **open/unmerged**, so most issues below are **still present**
on `feat/jae/bp32tree`. See [§4 Branch status](#4-status-on-current-branch-verified).

---

## 2. Master issue tally (deduplicated)

Severity: **C**=Critical, **H**=High, **M**=Medium, **L**=Low, **Lat**=Latent (not yet
reachable). "Status" verified against the working tree where marked ✅/❌; otherwise inferred
from merge state.

### Correctness — data loss / corruption / chain halt

| ID | Issue | Sev | Location | Source PR(s) | Status on branch |
|---|---|---|---|---|---|
| **C1** | **Pruning deletes nodes shared across versions after an inner-node split** (positional descent picks one "corresponding" node; split siblings deleted while live → next prune panics, `Get`/`Has` panic in `DeliverTx`, node bricked on cold `LoadVersion`). Triggers on first prune (~block 705,601 under `PruneSyncable`). | **C** | `prune.go` `walkAndPrune`/`findCorrespondingChild` | #5442 bug6; partial fix #5451; full rewrite #5570 finding:3 | ❌ **Open** — positional `findCorrespondingChild` still in tree (prune.go:183,203); #5570 mark-and-sweep unmerged |
| H1 | **Iterators never register as version readers** (`newIterator` hard-codes `version=0`; `incrVersionReaders` never fires) → a concurrent `PruneVersionsTo` deletes nodes a live iterator/snapshot is walking | H | `iterator.go` (all 3 call sites) | #5450; #5570 finding:1; #5591 #1 | ❌ **Open** — all 3 `newIterator(...,0)` still present (iterator.go:329,344,361) |
| ~~H2~~ | `GetImmutable`/`immutableForProof` don't register readers → snapshot torn out by concurrent prune | H | `mutable_tree.go`, `proof.go` | #5570 finding:30, finding:40 | ✅ **FIXED `a07f4b3ce`** (§IX) — register + idempotent `Close`; long-lived store views use `GetImmutableUnregistered` |
| ~~H3~~ | Pruning active-readers check is **TOCTOU** (check then delete in separate critical sections; reader can register in between) | M→H | `prune.go`, `nodedb.go` | #5450; #5570 finding:15 | ✅ **FIXED `a07f4b3ce`** (§IX) — `pruneMu` held across the prune |
| H4 | Pruning empty-tree branch (`vRootNK == nil`) skips orphan/value cleanup → value records leak permanently | H | `prune.go` | #5570 finding:2 | ❌ Open |
| H5 | **All-zero `ValueKey` collides with the "missing" sentinel** (nonce starts at 0; first ValueKey in v0 = 12 zero bytes = the missing-value placeholder) | H | `node_key.go`, `mutable_tree.go`, `import.go` | #5570 finding:6 | ❌ Open |
| H6 | **`Set` not atomic with value save** — tree mutated first, then `SaveValue`; on `SaveValue` error the leaf references a never-persisted ValueKey (dangling ref persisted on next `SaveVersion`) | H | `mutable_tree.go` | #5570 finding:28 | ❌ Open — `treeInsert` (mutable_tree.go:107) runs before `SaveValue` (:120) |
| H7 | `SaveVersion` partial-failure leaves `nextValueNonce`/session slices dirty → **nonce reuse** overwrites live values (compounds H5) | M→H | `mutable_tree.go` | #5570 finding:36; #5591 #2 | ❌ Open |
| H8 | **`LoadVersion(non-latest)` + `Set` corrupts values** — new ValueKeys allocated in the `V+1` namespace; `SaveValue` is an out-of-batch direct write that silently overwrites live values before the version hash-check can reject | H | `mutable_tree.go`, `nodedb.go` | #5570 second-pass | ❌ Open |
| H9 | **Idempotent `SaveVersion` leaks session state** — on the "version already exists, same hash" replay path `sessionValues` survives the early return; a later `Rollback` `DeleteValueDirect`s vks that now collide with live entries → **wipes real data** | H | `mutable_tree.go` | #5591 #2 | ❌ Open |
| H10 | **`VersionExists` swallows DB errors** (`has,_ := db.Has(...)`) → transient `Has` failure reads as "does not exist", letting `SaveVersion` overwrite an existing version | M→H | `nodedb.go:303` | #5591 #3 | ❌ Open — confirmed `has, _ := ndb.db.Has(...)` at nodedb.go:304 |
| H11 | **Crashed-session value leak** — `SaveValue` writes eagerly, `sessionValues` is in-memory only; a crash before `SaveVersion`/`Rollback` leaks values permanently (no recovery scan) | M | `nodedb.go`, `mutable_tree.go` | #5442 bug3; #5591 #5 | ❌ Open |
| ~~H12~~ ✅ | ~~**`immutableForProof` builds proofs from the *unsaved* working root** (`t.root` not `t.lastSaved`) → ICS23 proofs that can't verify against the committed root hash.~~ **FIXED TODAY.** (Scope: only the `MutableTree` proof wrappers; store `Query` proves via a committed `ImmutableTree` gated by `VersionExists`, so production was never affected.) | M→Low (prod: none) | `proof.go` / `mutable_tree.go` | #5468 | ✅ **Fixed today** (#5468 cherry-picked + nits + dedup) |

### Robustness — corruption detection / leaks / panics on bad input

| ID | Issue | Sev | Location | Source PR(s) | Status on branch |
|---|---|---|---|---|---|
| H13 | **`SaveValue` bypasses the write batch** (`db.Set`, not `batch.Set`); `Rollback` only restores the root pointer → orphaned values accumulate (unbounded DB bloat). "Certain" to manifest each block. | M | `nodedb.go:157` | #5442 bug3; #5570 finding:25/31 | ❌ Open — confirmed `return ndb.db.Set(key, valCopy)` |
| M1 | **No bounds check on deserialized `numKeys`** — uvarint cast to `int16` with no `[0,B-1]` range check; `0xFFFF`→`-1` slips past negative-checks; `32`→array-index panic. Single corrupt byte can brick a node. | M | `node.go` `readInnerNode`/`readLeafNode` | #5442 bug4; #5570 finding:23 | ❌ Open — uvarint read with no `> B` guard before cast |
| M2 | **`readBytes` OOM** — decodes a uvarint length and `make([]byte, length)` before `io.ReadFull`; bogus `1<<40` OOM-kills the process | M | `node.go` `readBytes` | #5450; #5442 bug4 | ⚠️ **Mitigated** — `length > maxReadBytesLen` cap now present (node.go:355); #5591 adds cumulative `maxLeafReadBytes` |
| M3 | `ReadNode` doesn't reject **trailing bytes** → corrupt payloads with extra bytes decode as silently-truncated nodes | M | `node.go` | #5591 #node-framing | ❌ Open |
| M4 | **`Serialize` silently writes nil child / valueKey refs** — leaf path writes a 12-byte zero placeholder (round-trips to "value not found"); inner path `w.Write(nil)` shifts every subsequent field read | M | `node.go` | #5570 finding:18; #5591 #framing | ❌ Open |
| M5 | **`GetNode` error surface is one bucket** — callers can't tell a legitimately-pruned node from a corrupt DB; errors silently swallowed (`continue`) in prune & elsewhere | M | `nodedb.go`, `prune.go` | #5438 review; #5570 finding:5 | ❌ Open — `child,err := GetNode(...); ...; continue` at prune.go:189 |
| M6 | **`Iterate`/iterator swallows resolver errors** — closure returns `(true,nil)` on error; caller sees "iteration complete" instead of a DB failure (silent data loss); `Valid()==false` indistinguishable from real error | M | `immutable_tree.go`, `mutable_tree.go`, `iterator.go` | #5438 review; #5570 finding:8/34/35 | ❌ Open |
| M7 | **`DeleteVersionsFrom` / `LoadVersionForOverwriting` leak node data** — delete only root refs (`R` prefix), never node entries (`B` prefix); stub with no orphan analysis | M | `nodedb.go` | #5450 | ❌ Open (#5450 unmerged) |
| ~~M8~~ | **Public API methods panic** — `LoadVersionForOverwriting` / `DeleteVersionsFrom` are IAVL-compat surface with no safe impl | M | `mutable_tree.go` | #5570 finding:12 | ✅ **FIXED `1db08015f`** — return wrapped `ErrUnsupported` instead of panicking |
| M9 | **Write side missing `MaxKeyLen` cap** — read side caps length-prefixed fields, write side unbounded; an oversize key serializes but fails to deserialize → version permanently un-mountable | M | `mutable_tree.go`, `import.go` | #5591 #6 | ❌ Open |
| M10 | **Import bounds not validated** — `Importer.Add` doesn't check `NumKeys` vs fixed-array bounds → OOB panic on malformed export stream | M | `import.go` | #5442-adjacent; **fixed #5451** | ✅ **Fixed** — `nk > B` / `> B-1` guards present (import.go:61,84) |
| ~~L1~~ | `PruneVersionsTo` can **rewind `firstVersion`** below the true first retained version → `AvailableVersions`/`discoverVersions` return wrong ranges | M | `prune.go` | #5570 finding:41 | ✅ **FIXED `1db08015f`** — below-floor prune is an early-return no-op |
| ~~L2~~ | Mid-loop prune error leaves **partial batch state** (deletes flushed by a later `Commit` while `firstVersion` is unadvanced → corrupt, unprunable version → crash loop) | L→M | `prune.go` | #5570 finding:52 | ✅ **FIXED `fc5451526`** — DiscardBatch on prune error + clean-session guard (`ErrUncommittedChanges`) + `findCorrespondingChild` error propagation |
| L3 | `deleteAllNodesForVersion` / unreachable `nextV==0` branch **skips orphan/value cleanup** (dormant value-leak timebomb) | L | `prune.go` | #5570 finding:43; #5591 #unreachable | ❌ Open |
| L4 | `nodeKeyBytesToArr` silently **zero-pads short slices** → miscompare vs reachable-set could delete live nodes | L | `prune.go` | #5570 finding:47 | ❌ Open |
| ~~L5~~ | `DeleteNode` **evicts from cache before batch commit** → a concurrent miss reloads from disk & re-caches, then batch flush leaves cache holding a deleted node | L | `nodedb.go` | #5570 finding:44 | ⚠️ **DOWNGRADED** — not a bug post-H2/H3 (see Tier B entry); `pendingEvicts` would not help; no fix |
| ~~L6~~ | `Rollback`/`orphanValueKey` **ignore `DeleteValueDirect` errors** → silent space leak with no diagnostic | L | `mutable_tree.go` | #5570 finding:25/31 | ✅ **FIXED `1db08015f`** — error propagated through `Set`/`Remove` |
| ~~L7~~ | `Importer` doesn't reset host-tree counters on reuse → stale pending session state could leak into the import's `SaveVersion` | L | `import.go` | #5570 finding:39 | ✅ **FIXED `8208eb3c1`** — `Import` calls `t.Rollback()` first (clean slate); nonce-1 reservation found unnecessary (versions ≥1) |

### Concurrency / thread-safety

| ID | Issue | Sev | Location | Source PR(s) | Status |
|---|---|---|---|---|---|
| ~~M11~~ | **Thread-safety under-specified**; `childMu` guarded only the lazy-load path → concurrent `Get`+`Set` race on the `getChild` write-back vs COW `Clone` | M | `node.go`, `mutable_tree.go` | #5570 finding:7 | ✅ **FIXED `0de551f17`** — getChild no longer memoizes; `childMu` removed; `-race`-clean. The separate value-resolution `pendingVals` race is also now fixed (`75c946820`, §VIII). |
| ~~M12~~ | **`immutableForProof` shares the mutable root** — proof walks `miniTree`/`childHashes` | M | `proof.go` | #5570 finding:9 | ✅ **FIXED `0de551f17`** — shared nodes now immutable (reads pure; miniTree rebuilt only on clones), so the aliased root can't be torn. |
| ~~M13~~ | No **`singleflight` on cache-miss** — two readers missing the same NodeKey deserialize independently and overwrite each other's `Add` | M | `nodedb.go` | #5570 finding:5/arch | ✅ **FIXED `27c6ebbe5`** — `loadGroup singleflight.Group` coalesces cache-miss loads |
| L8 | `ImmutableTree.Close` **not idempotent** — double/concurrent Close decrements reader count twice → corrupts count | L | `immutable_tree.go` | #5570 finding:45 | ❌ Open |

### Latent / API ergonomics

| ID | Issue | Sev | Location | Source PR(s) | Status |
|---|---|---|---|---|---|
| Lat1 | **Shallow `Clone()` shares mutable `[]byte` slices** (`c := *n`) — one `append()`/in-place write away from silent cross-version corruption | Lat | `node.go` `Clone` | #5442 bug1; #5570 finding:24/20/27 | ❌ Open (latent; audit shows no in-place mutation today) |
| M14 | **`resolveValue` returns the 32-byte hash as the value** when no resolver is set (silent data corruption); `ErrKeyDoesNotExist` conflates missing-key with no-resolver | M | `immutable_tree.go`, `mutable_tree.go` | #5442 bug2; #5570 finding:10/11 | ⚠️ **Partially fixed** — now returns `ErrKeyDoesNotExist` (no longer the hash); resolver/missing conflation remains (#5570 `ErrNoValueResolver` unmerged) |
| M15 | **Exporter goroutine leak + permanent version-reader lock** — abandoned `Exporter` blocks forever (>32 nodes) and pins its version (now → prune panic, post-H2/H3) | M | `export.go` | #5442 bug5 | ⚠️ **Mitigated** — must-`Close` contract documented (option b); no production caller (no state-sync); structural pull-based fix (a) deferred to state-sync |
| M16 | **`InlineValueThreshold` unbounded** (#5571 feature) — oversize inline value produces a leaf exceeding the reader budget → permanently un-mountable; field overloads `-1/0/positive` meanings | M | `options.go` (v2 leaves) | #5571 hardening | n/a — only relevant if #5571 inline-value feature lands |
| Lat2 | `slotsDirty`/`miniTreeDirty` invariants (#5571 feature) — `InnerNode.Clone` cleared `miniTreeDirty` (stale-merkle hazard); `splitLeaf` missed a dirty mark | M | `node.go`, `split.go` | #5571 self-review | n/a — within #5571's new incremental-merkle feature |

---

## 3. The 6 canonical demonstrated bugs (#5442, with reproducer tests)

These are the original review bugs, each with a **failing reproducer test** in `bugs_test.go`.
Maps to master IDs in brackets.

| # | Bug | Sev | Likelihood | Reproducer tests | Master |
|---|---|---|---|---|---|
| 1 | Shallow clone shares mutable slices | Low (latent) | None today | `TestBug1_CloneSharesSlices`, `TestBug1_COWSafety`, `TestBug1_COWRegressionTest` | Lat1 |
| 2 | `resolveValue` returns hash as value | Medium | None on store path | `TestBug2_ResolveValueReturnsHash`, `TestBug2_ImmutableResolveValue`, `TestBug2_GetReturnsHashNotValue`, `TestBug2_StoreLayerSetsResolver` | M14 |
| 3 | `SaveValue` bypasses batch; Rollback leaks | Medium | **Certain** | `TestBug3_SaveValueBypassesBatch`, `TestBug3_RollbackLeavesOrphanedValues`, `TestBug3_ValueVisibleBeforeCommit`, `TestBug3_OrphanAccumulation` | H13/H11 |
| 4 | No bounds check on deserialized `numKeys` | Medium | Low | `TestBug4_NumKeysOverflow`, `TestBug4_NegativeNumKeysOverflow`, `TestBug4_ZeroNumKeys`, `TestBug4_MaxValidNumKeys` | M1/M2 |
| 5 | Exporter goroutine leak + version-reader lock | Medium | Low (dormant) | `TestBug5_ExporterGoroutineLeak`, `TestBug5_VersionReaderLeak` | M15 |
| 6 | **Pruning deletes shared nodes across versions** | **Critical** | **Certain** | `TestBug6_SingleVersionPruneCorruptsTree` (panic ~block 98, 302 cascading errors), `TestBug6_PruneCorruptsNewerVersions` (only 4,686/18,000 keys readable after prune), `TestBug6_PruneBricksNodeOnRestart` | **C1** |

> #5442's bottom line: **"Bug #6 is a ship-blocker. Any chain running B+32 with default
> pruning will eventually panic and brick"** — the 705,600-block delay before first prune
> (~12 days @ 1s blocks) means it passes all testing but fails in production.

---

## 4. Status on current branch (verified)

Greps against `feat/jae/bp32tree` working tree (2026-06-06):

| Check | Result | Implication |
|---|---|---|
| `prune.go` algorithm | `walkAndPrune` + `findCorrespondingChild` (positional) still present | **C1 critical pruning bug class still in tree** — #5451 added `findCorrespondingChild`-from-root (mitigates the *single*-node variant) but #5570 finding:3 argues the positional approach is still wrong under nested split/merge; mark-and-sweep rewrite unmerged |
| `newIterator(...)` calls | all 3 pass `version=0` | **H1 unfixed** — iterators don't register as readers |
| `resolveValue` (immutable) | returns `ErrKeyDoesNotExist` (not `vh[:]`) | M14 **no longer returns hash**; resolver/missing conflation remains |
| `SaveValue` | `return ndb.db.Set(...)` (direct) | **H13/H11 present** — values bypass batch |
| `readBytes` | `length > maxReadBytesLen` cap present | **M2 mitigated** by constant cap |
| `import.go` numKeys | `nk > B` / `> B-1` guards present | **M10 fixed** (merged #5451) |
| `VersionExists` | `has, _ := db.Has(...)` | **H10 present** — swallows error |
| `export.go` | `go e.run()`, no context | **M15 present** — leak path open |

**Merged into branch:** #5451 (M10 import bounds, C1 *partial*, round-trip determinism), #5475.
**Everything else (#5570, #5591, #5450, #5468) is unmerged → those fixes are NOT in the branch.**

---

## 5. Full appendices (per-PR source lists)

### Appendix A — #5442: 6 bugs
See [§3](#3-the-6-canonical-demonstrated-bugs-5442-with-reproducer-tests).

### Appendix B — #5570: 52 findings + 2 second-pass

Severity from PR: H/M/L/C. Bracketed master IDs where deduped.

**Correctness (14):**
- finding:1 (H) Iterators didn't register as version readers `[H1]`
- finding:2 (H) Pruning empty-tree branch leaked orphans `[H4]`
- finding:3 (H) Cascading prune corruption under splits/merges — positional descent removed, replaced by content-addressed mark-and-sweep `[C1]`
- finding:6 (H) All-zero ValueKey collided with "missing" sentinel `[H5]`
- finding:8 (M) `Iterate` swallowed resolver errors `[M6]`
- finding:15 (M) Pruning active-readers check was TOCTOU `[H3]`
- finding:26 (L) Idempotent `SaveVersion` mismatched legacy empty-tree blob
- finding:28 (H) `Set` not atomic with value save `[H6]`
- finding:30 (H) `GetImmutable`/`immutableForProof` didn't register readers `[H2]`
- finding:36 (M) `SaveVersion` partial-failure left counters dirty `[H7]`
- finding:40 (H) readers registered AFTER loading root `[H2]`
- finding:43 (L) `deleteAllNodesForVersion` skipped orphan processing `[L3]`
- *second-pass:* `LoadVersion(non-latest)`+`Set` value corruption `[H8]`
- *second-pass:* `SetCommitting`/`UnsetCommitting` dead flag removed

**Robustness / error handling (20):**
- finding:5 (H) `GetNode` error surface single bucket `[M5]`
- finding:12 (M) Public API methods panicked (`LoadVersionForOverwriting`, `DeleteVersionsFrom`) `[M8]`
- finding:13 (M) `nodeDB` lock-discipline gaps undocumented
- finding:18 (M) `InnerNode.Serialize` silently wrote nil child refs `[M4]`
- finding:23 (L) Bounds checks ran AFTER unchecked casts (`numKeys`, `orphanValueKey`) `[M1]`
- finding:25 (L) `Rollback`/`orphanValueKey` ignored `DeleteValueDirect` errors `[L6]`
- finding:31 (L) same as 25 (Tier 1)
- finding:34 (L) Iterator invalidation lost the underlying error `[M6]`
- finding:35 (L) `MutableTree.Iterator` silently returned nil values (no resolver) `[M6]`
- finding:37 (L) `treeRemove` root-collapse nodeKey expectations (doc only)
- finding:38 (L) `Commit` lifecycle on error silently discarded writes
- finding:39 (L) `Importer` didn't reset host-tree counters on reuse `[L7]`
- finding:41 (M) `PruneVersionsTo` could rewind `firstVersion` `[L1]`
- finding:42 (M) Unwired `AsyncPruning` option removed (latent racing writer)
- finding:44 (L) `DeleteNode` evicted from cache before batch commit `[L5]`
- finding:45 (L) `ImmutableTree.Close` no idempotency/synchronisation `[L8]`
- finding:46 (L) Mark-and-sweep leaf-skip relied on unchecked `height==1` invariant
- finding:47 (L) `nodeKeyBytesToArr` silently zero-padded short slices `[L4]`
- finding:50 (L) `deleteSubtree` ignored `childNodes[i]` (in-memory children leak)
- finding:52 (L) Mid-loop prune error left partial batch state `[L2]`

**Code quality / API (12):** finding:7 thread-safety under-specified `[M11]`; finding:9 `immutableForProof` shared mutable root `[M12]`; finding:10 `ErrKeyDoesNotExist` conflated missing/no-resolver `[M14]`; finding:11 `Iterate` w/o resolver returned hash bytes `[M14]`; finding:19 dead code/config; finding:20 name/identifier inconsistencies + separator `copyKey` `[Lat1]`; finding:22 magic numbers/brittle constants (`maxReadBytesLen` 1MiB→64KiB); finding:24 internal contract fragility (`Clone` struct-copy); finding:27 cosmetic; finding:32 panic-on-default in type switches (resolved as designed); finding:48 stale comments; finding:51 leaf-skip comment.

**Performance (8) — see Appendix I:** finding:4, 14, 16, 17, 21 (8 sub-items), 29, 33, 49.

### Appendix C — #5591: 15 deep-dive findings

**Correctness / safety:**
- Iterator version readers (`version=0`) `[H1]`
- Idempotent `SaveVersion` leaked session state → Rollback wipes live data `[H9]`
- `VersionExists` swallowed DB errors `[H10]`
- `GetValue` couldn't distinguish missing from empty (`(nil,nil)`) — masks corruption as empty
- Crashed-session value leak — `Load()` now scans value keyspace above `latestVersion` `[H11]`
- `MaxKeyLen` cap missing on Set/Import write side `[M9]`
- Unreachable `deleteAllNodesForVersion` value-leak branch `[L3]`

**Fail-fast guards:**
- `ReadNode` trailing-bytes check `[M3]`
- `Serialize` rejects nil `valueKey`/child-ref `[M4]`

**Performance (Appendix I):** `saveNode` force-loaded unchanged siblings; `PruneVersionsTo` batch memory (`FlushThreshold`); deferred `copyKey` on update path.

**Defensive / docs:** orphans of first pruned version; `redistributeLeft/Right` inner-case `copyKey` `[Lat1]`; README dedup-claim correction (code does **not** content-address/dedup values — every `Set` allocates a fresh `ValueKey`).

### Appendix D — #5450 (notJoon): 3 bugs
1. `readBytes` OOM from corrupt length — `length > r.Len()` guard `[M2]`
2. `DeleteVersionsFrom` / `LoadVersionForOverwriting` leak node data (root-ref-only delete) `[M7]`
3. `NewIteratorWithNDB` doesn't register as version reader + `PruneVersionsTo` TOCTOU (`beginPruning`/`endPruning` atomic check-and-mark, `ErrVersionBeingPruned`) `[H1/H3]`

### Appendix E — #5468 (notJoon): 1 bug
- ~~`immutableForProof()` used `t.root` (unsaved working tree) → proofs from uncommitted state, unverifiable against committed `Hash()`.~~ ✅ **FIXED TODAY.** Scope: only the `MutableTree.Get{,Non}MembershipProof` wrappers (tests/bench) — the store `Query` path proves via a committed `ImmutableTree` (gated by `VersionExists`), so production was never affected; a `MutableTree`-API/test-hygiene fix, not a query-security bug. Fix: use `t.lastSaved`, return `ErrNoCommittedState` only when nothing was *ever* committed (committed-but-empty → `ErrEmptyTree`). Cherry-picked #5468 (`cebeecf9a`) + follow-up nits (NonMembership guard assertion, committed-empty test, DB-backed Snapshot test, doc comments, `newImmutable` resolver dedup). `[H12]`

### Appendix F — #5451 (clockworkgr) — MERGED ✅
- Round-trip determinism (export→import identical structure/rootHash)
- **Critical pruning bug #6 from #5442** (partial — positional `findCorrespondingChild`-from-root; #5570 argues still insufficient) `[C1]`
- Comprehensive IAVL-comparison benchmark suite
- Import bounds not validated → OOB panic `[M10]`

### Appendix G — #5571 (clockworkgr): correctness items
(Feature PR — v2 inline values / v3 prefix keys / fastnode cache / 13-fix perf. Bugs *introduced & closed within its own features*:)
- `InnerNode.Clone` cleared `miniTreeDirty` → stale-merkle hazard `[Lat2]`
- `splitLeaf` missed a dirty-mark on one path `[Lat2]`
- Fastnode cache served **stale values** on root swap → wholesale purge on Rollback/LoadVersion; defensive copy on Get-hit
- `InlineValueThreshold` unbounded → un-mountable leaf; 3-layer clamp + named `InlineThreshold` type `[M16]`
- `maxLeafReadBytes` cumulative cap vs v2/v3 amplification

### Appendix H — #5438 inline review comments (15)
Seeds of the fix PRs above:

| # | Reviewer | File | Comment | Master |
|---|---|---|---|---|
| 1 | notJoon | store.go:143 | `LoadVersion` immutable path silently ignores error (mutable path at :153 handles it) | M5 |
| 2 | notJoon | mutable_tree.go:82 | `treeInsert` recomputes sha256 for same value — pass the precomputed hash | perf |
| 3-6 | clockworkgr | prune.go (×4) | `GetNode` errors silently ignored — `continue` only if actually deleted, else bubble up | M5 |
| 7 | clockworkgr | nodedb.go:304 | should check the error here | M5/H10 |
| 8 | clockworkgr | immutable_tree.go | Iterator returns `(true,nil)` on error → error lost = silent data loss (same in MutableTree) | M6 |
| 9 | clockworkgr | mutable_tree.go | same iterator issue | M6 |
| 10 | clockworkgr | node.go:62 | document `getChild` panics on DB-load failure (unrecoverable) | M5 |
| 11 | clockworkgr | nodedb.go:313 | loop over all versions w/ `VersionExists` vs iterator scan (use `discoverVersions` pattern) | perf (finding:14) |
| 12 | clockworkgr | options.go:8 | are these options used anywhere? (dead config) | finding:19/42 |
| 13 | clockworkgr | mini_merkle.go:88 | `B` is compile-time const — don't recompute depth at runtime | perf (finding:21) |
| 14 | clockworkgr | remove.go:326 | use `copyKey` for separator consistency | Lat1 |
| 15 | notJoon | split.go:69 | `splitInner` uses direct reference vs `splitLeaf`'s copy — keep consistent | Lat1 |

Plus a design thread (clockworkgr/notJoon): make `getNode` return `ErrNodeNotFound` for deleted nodes and panic for other failures (→ #5570 finding:5).

### Appendix I — Performance findings (NOT bugs)
Listed for completeness; excluded from the bug tally per request.
- **#5570:** finding:4 `saveNode` reloads unchanged subtrees (H, 192 reads vs 6 writes/update); finding:14 `AvailableVersions` O(versions) DB lookups; finding:16 iterator 1 DB read per `Value()`; finding:17 unconditional root clone per `Set` (~4.3 KB); finding:21 hot-path allocations (8 sub-items); finding:29 `Clone()` copies 2 KB mini-merkle + rebuild; finding:33 `seekLast` over-descent; finding:49 `sweepOld` redundant recursion.
- **#5591:** `saveNode` force-loaded unchanged siblings (~18K reads/block @ scale); `PruneVersionsTo` batch memory; deferred `copyKey` on update.
- **#5571:** fastnode cache, pooled serialization (allocs −85%), incremental mini-merkle, v2 inline values, v3 prefix compression, mark-and-sweep + LoadVersion fast path.

---

## 6. Bottom line

- **One ship-blocker:** **C1** — the positional pruning algorithm corrupts the tree after
  inner-node splits and **is still present** on `feat/jae/bp32tree`. #5451 only partially
  addressed it; the definitive mark-and-sweep rewrite (#5570) is unmerged.
- **High-severity correctness still open** (all in unmerged PRs): H1–H13 — iterator/snapshot
  reader registration & prune TOCTOU (H1–H3), value-key nonce collisions & non-atomic value
  save & version-overwrite races (H5–H10), value/orphan leaks (H4, H11, H13).
  (**H12 uncommitted-state proofs: fixed on branch — and was never reachable via the
  store `Query` path, which proves against a committed `ImmutableTree`.**)
- **Already fixed on branch:** M10 (import bounds, #5451), M2 (readBytes cap), M14 (no longer
  returns hash-as-value — partial).
- **Robustness/concurrency hardening** (M1, M3–M13, L1–L8, M15) awaits #5570/#5591/#5450.
- Two reviewers independently flagged the **iterator-version-reader** (H1) and
  **prune-reader TOCTOU** (H3) bugs (#5450 notJoon, #5570/#5591 clockworkgr), and #5442's
  reproducers confirm the **critical prune** (C1) bug — high confidence these are real.
