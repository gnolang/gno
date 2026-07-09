# ADR: Mount the B+32 (bptree) store with the fast index; reprice depth gas

## Status

Proposed (PR pending). Stacked on the tm2 PR that made `MutableTree.Get`
serve clean-working-tree reads from the fast index
(`tm2/adr/prxxxx_bptree_fastindex_working_tree.md`).

## Context

gno.land's main state store has been IAVL (`tm2/pkg/store/iavl`, mounted at
`gno.land/pkg/gnoland/app.go`), constructed with fast storage skipped — every
point read is a full binary-tree descent (~15 uncached reads at 100M keys)
while gas charged only 3.0, a ~5× wall-clock underprice.

The B+32 store (`tm2/pkg/store/bptree`) wins the write path decisively
(~4.5× fewer write ops, ~12× fewer COW-path reads, 4.5–4.9× lower per-write
latency; `tm2/pkg/bptree/PERFORMANCE.md` — like-for-like workloads, same
fixture and block shape across backends). Its historical weakness — point
reads (~3.7 reads vs IAVL-with-fast-nodes' ~1) — is closed by the fast index
now that it serves the consensus read path. The gas depth params
(`gno.land/pkg/sdk/vm/params.go`) were already B+32-calibrated in
anticipation; this PR makes the pricing real.

## Decision

1. **Mount** `mainKey` with `storebptree.FastStoreConstructor` (B+32 + fast
   index; the stacked tm2 PR replaced the former package-global
   `FastIndexEnabled` toggle with this per-mount constructor). The index
   is the reference validator config: it is unauthenticated, outside the
   Merkle commitment, and gas-invisible, so a node that patches it off only
   hurts its own wall-clock — no fork.
2. **Reprice depth gas** (consensus genesis defaults):
   - `FixedGetReadDepth100`: 300 → **100**. A present-key Get on committed
     state is one flat DB read via the fast index, independent of tree size —
     so GET stays *pinned*, like IAVL fast nodes would be.
   - `FixedSetReadDepth100` stays **200**; `FixedWriteDepth100`: 440 →
     **540** (+1.0 for the per-mutation index write). Both remain PINNED at
     the measured-with-cache calibration rather than estimator-driven
     (Fixed=0) — the live estimator ignores the node LRU and overcharges
     mid-range sizes ~2× (see Alternatives). Revisit Fixed=0 once the
     estimator is cache-aware (follow-up 2).
   - Min floors set to match (100/200/540; `NewParams` pins Fixed = Min):
     inert behind the pins, they are the documented fallback if governance
     ever zeroes a Fixed pin (0 = "use tree estimate, floored by Min" — a
     legitimate value; params are stored per-field with explicit zeros and
     nothing backfills them). `TestDefaultParams` pins the default values
     and the Fixed = Min mirror, including a keeper round-trip.
3. **Fork carryover** (`contribs/gnogenesis`): a genesis exported from a
   chain carrying the untuned legacy defaults (Fixed == Min == 300/200/440,
   IterNextCostFlat == 1000 — the exact pre-mount fingerprint) is rewritten
   to the new defaults; any deviation means operator tuning and carries over
   verbatim. Without this, forked chains would run the bptree store under
   IAVL-era pricing (GET 3× overcharged, writes missing the index cost).
   Note: historical governance txs that
   explicitly set a param re-apply during genesis replay and can re-pin old
   values — correct behavior (explicit settings win), worth knowing.

## Consequences

- **Consensus-breaking for existing chains**: the commitment structure (app
  hash) changes with the backend. Fresh chains and export/import forks only;
  `gnogenesis fork` + `InitialHeight`/`SetInitialVersion` support this. The
  fork tool's *source* reader deliberately keeps the IAVL constructor
  (sources are legacy-format data dirs); forking a bptree-era chain from a
  data dir will need a bptree source reader later.
- **Gas shifts** (measured; identical txtar workloads before and after — only
  the price params changed): typical addpkg ~2.78M → ~2.48M (−11%), calls
  ~1.21M → ~1.02M (−15.5%) — GET repricing dominates; writes are dearer
  (write depth 4.4 → 5.4). Golden updates in this PR: `restart_gas`, `gc`,
  `gnokey_gasfee` (re-derived simulate→broadcast chain), `simulate_gas`,
  `stdlib_restart_compare`, `stdlib_ibc_crypto_determinism` txtars, the
  crossrealm38 multistore hash, and comment-only loosening in
  `addpkg_outofgas`.
- **Gas stays size-independent** (all three depths pinned): the only price
  direction change vs the old pins is Sets +~11% flat (write depth 4.4 →
  5.4 = +24,000 gas for the index write) against GETs −67%; a typical
  read-modify-write is ~−23% at every chain size. The trade accepted with
  pins: they drift gradually underpriced past the 100M calibration point
  (the write pin ~−13% at 1.6G keys; the SET-read pin faster on a measured
  basis) until governance re-tunes them — the self-correcting alternative
  (estimator pricing) is deferred until the estimator models the node
  cache.
- **Accepted imprecisions** (canonical here; `params.go` points at this ADR): absent-key GETs still
  walk (~3.7 reads at 100M) but are charged 1.0 — the designated fix is a
  post-fetch nil-result surcharge in `cacheStore.Get` (it already sees the
  nil result); the index's duplicated value bytes are not captured by
  `WriteCostPerByte`; `IterNextCostFlat=1000` undercharges bptree's
  out-of-line per-step value read (~1 random read/step; parity with the
  previous IAVL mount, which read a node per step — follow-up
  recalibration). Non-index-served reads (old-version queries) also walk;
  they are off the metered tx path. Each ABCI query-height store open also
  pays one flat stamp read (`ensureFastIndex` no-op check) plus tm2's
  pre-existing per-open version discovery — and the latter is a LINEAR scan
  over retained root records, which the mount promotes to the production
  query path: at the default PruneSyncable strategy (KeepRecent=705,600)
  that is ~705K iterator steps per query at steady state, unbounded on
  archive nodes, executed under the shared ABCI mutex. The IAVL store's
  discovery was O(log n). This is the highest-priority follow-up below.
- **Operator notes**: the index duplicates every live value on disk (bounded
  ~2× value bytes — note this also halves the physical bytes covered per
  deposited storage byte, since deposits price the logical value once); the
  first `Load` over an existing *non-indexed* bptree DB performs a full index
  rebuild inside startup — silently, in the shipped mount (the constructor
  passes a NopLogger; accepted because fresh chains and fork flows never
  rebuild) — and a rebuild error fails the node loudly. A CORRUPT index
  stamp is fail-stop (a missing one rebuilds); the escape hatch today is
  deleting the stamp key by hand. Normal restarts never rebuild — the stamp
  is maintained transactionally.
- Test scaffolds intentionally left on IAVL: `tm2/pkg/sdk/{auth,bank,params}`
  test_commons and `tm2/pkg/sdk/baseapp_test.go` (backend-agnostic tm2
  scope). All gno.land scaffolds that
  mirror the app's mounts were swapped (`gnoland/test_common.go`,
  `app_test.go` incl. the shared-DB `TestPruneStrategyNothing`,
  `sdk/vm/common_test.go` — the vm gas suite now runs the production code
  path; charged gas is index-independent).

## Alternatives considered

- **Keep IAVL, enable its fast nodes**: point reads match, but SET-reads
  ~16× and writes ~4× costlier (BENCHMARKS.md "If gno.land used IAVL"
  table), and write costs scale with log₂N — the workload is write-dominant.
- **Mount bptree without the index**: honest GET price would be ~300 and
  size-coupled; point reads stay ~4× slower than IAVL+fast. The index costs
  disk, not correctness (checksummed, same-batch-atomic, advisory).
- **Estimator-driven SET/WRITE (Fixed=0, Min floors)**: self-corrects as
  the tree grows — but the current estimator ignores the node LRU and
  overcharges mid-range sizes ~2× (4.0 charged vs ~2.0 measured at 1M keys),
  and makes gas stepwise size-dependent (client gas-wanted caching breaks at
  power-of-two boundaries). Deferred until the estimator is cache-aware;
  the pins price what was measured.
- **Hedge GET at 150–200 for absent-key misses**: taxes every honest read
  +50–100% to shave a tail that is already better than the IAVL-era status
  quo (worst-case forced I/O per gas improves ~27%), and iteration — not
  GET-miss — is the cheapest adversarial read primitive either way.

## Follow-ups (tracked here until issues are filed)

In rough priority order:

1. Seek-based (or cached) version discovery for per-query immutable store
   opens — removes the O(retained-versions) scan per ABCI query (tm2
   `discoverVersions`).
2. Cache-aware depth estimator (model the node LRU absorbing the top tree
   levels in `expectedDepth100`), then revisit Fixed=0 estimator pricing for
   SET/WRITE so gas self-corrects with state growth instead of requiring
   governance re-tunes.
3. Post-fetch nil-result surcharge in `cacheStore.Get` — honest absent-key
   pricing without touching present-key reads.
4. `IterNextCostFlat` recalibration for bptree's out-of-line per-step value
   read.
5. Read-only `ensureFastIndex` on immutable DBs (nil-batch rebuild hazard)
   and rebuild-on-corrupt-stamp instead of fail-stop.
6. Fold the user key into the fast-index entry CRC (format bump +
   rebuild-on-upgrade).
7. Post-import stamp re-arm (the first post-import SaveVersion stamps a
   near-empty index complete; perf only).
8. bptree-era fork source reader for `gnogenesis fork --source-txs-data-dir`.
9. Index hit/fallback/rebuild metrics; node-config toggle for the index;
   operator migration guide under docs/; resumable index rebuild.
10. Benchmarks: index-ON Get/GetMiss run on the existing ~101M fixture;
   flat-cost (ReadCostFlat/WriteCostFlat) validation; steady-state disk
   footprint.

## AI assistance

Implemented with AI assistance (plan and diff reviewed through multi-agent
review rounds to convergence); the human author reviewed and owns the change.
