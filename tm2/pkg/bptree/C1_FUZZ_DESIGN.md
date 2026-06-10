# C1 fuzzer — design and semantics

## What this is, in plain English

Both fuzzers throw endless random sequences of operations at the tree and
constantly check that nothing breaks — they just stress different things:

- **`FuzzTreeOps`** is like hiring millions of short-lived chaos monkeys.
  Each one gets a brand-new tree and a random script — write some keys,
  delete some, save a version, prune old versions, take a snapshot, restart,
  inject a fake disk error — runs it for a few thousand steps, and then the
  wreckage is audited: is every saved version still perfectly readable? Did
  the prune delete exactly the right records — nothing it shouldn't have
  (corruption), nothing left behind (leaks)? Do the cryptographic proofs
  still verify? Go's fuzzer also *learns*: a random script that reaches new
  code paths is kept and mutated, so over time it hunts the weird corners a
  human wouldn't think to write.

- **`TestSoak_TreeOps`** is one monkey with one tree, forever. It simulates
  a node that just keeps running — saving a version, pruning old ones, over
  and over, hundreds of thousands of times — auditing after every prune. Its
  job is to catch anything that only shows up with *age*: slow leaks,
  drifting bookkeeping, anything that accumulates across a long lifetime
  rather than within one short script.

Short version: the fuzzer asks *"is there any weird sequence of operations
that breaks the tree?"* — the soak asks *"does the tree survive running
forever?"* Both have proven teeth: pointed at the pre-fix code, they catch
the M19 twin leak and the missing M20 guard within the first second.

## How to run

```sh
# Fuzzer — runs forever until Ctrl-C (or bound it with -fuzztime=10m):
go test ./tm2/pkg/bptree/ -run '^$' -fuzz FuzzTreeOps

# Soak — one persistent tree, indefinitely (or BPTREE_SOAK=2h etc.):
BPTREE_SOAK=forever go test ./tm2/pkg/bptree/ -run TestSoak_TreeOps -timeout=0 -v

# Concurrent-reader stress (runs in the normal suite; best under -race):
go test -race ./tm2/pkg/bptree/ -run TestStress_ConcurrentSanctionedReaders

# The ten seed programs also run as plain subtests in every normal `go test`.
```

---

> What follows is the specification behind `c1fuzz_test.go`: the op set, the
> model, the `expectedPrune` precedence table, and the oracles. Read this
> before changing the engine — most clauses here exist because the
> alternative was shown to false-positive or miss bugs.

## Objective
Continuously-accumulating assurance for the prune (C1): coverage-guided op
programs interleaving prune with every other operation, checked by exact
oracles (over-deletion, leak, per-version integrity, proofs, bookkeeping).

## Architecture: one engine, three entry points

### The engine
Two layers:
- `runOpChunk(tb testing.TB, st *fuzzState, data []byte)` — decodes and
  executes ops against a PERSISTENT state (tree, model, holds, error-injector),
  running oracles per the cadence. `fuzzState` carries the tree, the wrapped
  DB (failingGetDB around memdb — constructed ONCE, since nodeDB captures the
  handle), per-version model snapshots + recorded root hashes, outstanding
  holds, the op counter (value derivation), `maxImportVersion`, and cfg
  (keyspace size, window W, hold budget, session-ops cap, maxOps, cacheSize).
- `runOpProgram(tb, data, cfg)` — fuzz-entry wrapper: fresh `fuzzState`, one
  chunk, final full oracle.
Everything derives from `data`/seed (no wall clock, no global rand). The
engine never calls `t.Parallel()` and never asserts inside goroutines.

### Entry 1 — `FuzzTreeOps(f *testing.F)`
Seeds = known-nasty programs (below). Per iteration: `runOpProgram` with
`maxOps` ≈ 2048 decoded ops. Full oracle after every prune + at end. CI uses
`-fuzztime`; seed corpus entries also run as plain subtests (engine is
re-entrant, no package-level mutable state).

### Entry 2 — `TestSoak_TreeOps` (env-gated)
Skip unless `BPTREE_SOAK` is set (duration or "forever"; document
`-timeout=0`). ONE `fuzzState` forever: seeded rng generates chunks for
`runOpChunk`. Boundedness by construction:
- **Forced prune cadence + catch-up**: when retained window > W, enter
  catch-up — Rollback a dirty session, **Close ALL harness holds**, suppress
  new holds at/below the target, prune to latest−W, and **assert it
  succeeds**. The `window > 2×W ⇒ fail` check remains only as a backstop for
  genuine reader leaks (it cannot false-positive once catch-up closes every
  harness hold — any survivor is a harness/library leak).
- **Session-ops cap**: ≥ N mutations without a SaveVersion forces a
  SaveVersion (pendingVals/versionOrphans grow per-mutation and clear only at
  commit/rollback).
- **Import disabled by default** in soak (M21 — repeated imports leak
  unboundedly); enable-able with a count cap.
- **Bounded failure context**: failures report the op counter and engine
  state; the full program regenerates deterministically from the seed (no
  unbounded trace is kept).
- Height-4 (40k-key) config: follow-up, not v1.

### Entry 3 — `TestStress_ConcurrentSanctionedReaders` (seeded `-race`)
Not a fuzz target (schedules aren't input-reproducible). Seeded writer runs a
mutator-only chunk stream; N readers hammer ONLY the sanctioned surface
(`GetImmutable(committed)`+Get/Has/GetMembershipProof/Iterator+Close,
`GetVersioned`, `VersionExists`, `AvailableVersions`) with random hold-times.
Reader-side expectations are NON-exact (R-pin): a prune may return
`ErrActiveReaders` or nil depending on timing, and a reader's `GetImmutable`
may race to `ErrVersionDoesNotExist` — assert absence of corruption/panic/race,
not exact returns. Small in CI, env-scalable.

## Op set (decoder)
One opcode byte + parameter bytes; key indices map into `cfg.keys` (default
800, `"fz%04d"`); values derive from the monotone op counter.

Mutators:
- `Set(k)` — fresh value.
- `SetSame(k)` — rewrite existing key with its CURRENT value (twin-maker);
  no-op byte if absent.
- `Remove(k)` — no-op byte if absent.
- `NetZero(j)` — Set-then-Remove of a key from a DISJOINT sub-keyspace
  (`"nz%04d"`, never used by Set/GrowWave) so it is always genuinely fresh
  (A net-zero on an existing key is just Remove+orphan — different
  transition, covered by Remove.)
- `GrowWave(n)` — sequential-key run (90/10 split path).
- `DrainAll` — remove every live key. On an already-empty tree: pure no-op,
  does NOT dirty the session (Remove of an absent key returns early).
- `SaveVersion` — records {model snapshot, root hash}. NOTE (pinned):
  consecutive saves always create NEW versions (no no-change short-circuit;
  WorkingVersion = version+1 unconditionally); a clean-session save of a
  non-empty tree creates an "alias version" whose root ref is the SAME record
  (saveNode early-returns on non-nil NodeKey) — prune handles it via
  sameRecord; the model snapshots are simply duplicated.
- `EmptySave` — SaveVersion with no preceding mutations (same pinned note).
- `Rollback` — model overlay dropped.
- `LoadOld(v)` — LoadVersion of a retained version < latest. SKIPPED when only
  one version is retained (see precedence: a covering prune would then
  hit the "cannot prune latest" plain error, not ErrActiveReaders).
  **Entry semantics (pinned): `LoadVersion` itself discards any staged
  session at entry (`DiscardBatch` + `resetSession`, mutable_tree.go:404-409)
  — the model drops its uncommitted overlay at the op's START (Rollback
  semantics, same as ColdRestart) and the working view becomes the v
  snapshot.** While loaded: a covering prune (toVersion ≥ v) and a below-prune
  (toVersion < v) are both asserted **per `expectedPrune`** (covering ⇒
  `ErrActiveReaders`; below ⇒ nil unless a hold covers `[first, toVersion]`,
  then cell-5 `ErrActiveReaders`); after a below-prune the loaded view must
  still read correctly. A SaveVersion while loaded is DECIDABLE:
  working hash == recorded hashes[v+1] ⇒ idempotent adoption (nil error,
  returns v+1; the working tree becomes the persisted v+1 — model overlay
  replaced by the v+1 snapshot); else ⇒ plain error `"already exists with a
  different hash"` (deferred DiscardBatch already dropped staged writes; the
  model drops the overlay at the op's closing recovery). The op always ends
  with `Load()` (latest) so the program continues normally.
- `ColdRestart` — **Close all outstanding holds FIRST and assert
  versionReaders is empty** (registrations live in the old nodeDB's
  memory; dropping the tree would orphan them and void oracle 4), then drop
  the tree, reopen from the same db, `Load()`, verify latest vs model.
  **The model drops its uncommitted working overlay** — staged session state
  lived only in the old batch/pendingVals and does not survive the restart
  (ColdRestart with a dirty session has Rollback semantics for the model).
- `ExportImport` — export a retained NON-EMPTY version (Export of an empty
  tree errors — pinned), import at exactly `latest+1`. `Import()` itself
  Rollbacks the session first and `Importer.Commit` leaves the working tree =
  the imported content (pinned — the model must mirror both). Sets
  `maxImportVersion`. Disabled in soak by default.
- `InjectError(n)` — restricted to ops that actually perform PrefixNode
  reads: **prune** (and optionally a LoadOld-triggered idempotent adoption).
  Preconditions and semantics:
  - **Purge the node cache before arming** (`nodeCache.Purge()`, or run
    injection-bearing programs with cacheSize=0 as l2_prune_error_test.go
    does) — `GetNode` is cache-first, so a warm cache serves every prune
    read and the injected branch is otherwise unreachable.
  - The read-count branching applies ONLY when `expectedPrune(toVersion)`
    (below) is nil — i.e. an otherwise-valid cell-6 prune. If the target
    fails a precedence cell 1–5, the prune errors with possibly ZERO node
    reads; assert the predicate's outcome instead.
  - Branch on whether the injector actually FIRED (it exposes a counter):
    fired ⇒ expect an error, assert all retained versions intact, and a
    disarmed retry succeeds (the continuous L2 property); not fired ⇒ the
    prune followed the predicate's normal expectation and the model applies
    the corresponding transition.
  - **Disarm before any oracle runs** (the oracles read through the same
    wrapped DB), and disarm regardless of which branch was taken.
  - Never arm for an ORDINARY SaveVersion (its create path performs zero
    PrefixNode reads — saveNode only writes; the existence check is a Has on
    PrefixRoot; only the idempotent-adoption path does one node read).

Prune (the pinned precedence table). **Implement the table once as a
model-side predicate `expectedPrune(toVersion) → {nil | sentinel | plain}`
over (floor, latest, session-dirty, loadedVersion, outstanding holds), and
have EVERY prune-return assertion consult it** — the standalone prune op,
LoadOld's internal covering/below prunes (a "below" prune is nil ONLY if no
hold covers `[first, toVersion]`; with such a hold it is cell-5
`ErrActiveReaders`), InjectError's not-fired branch, and soak catch-up. Do
not restate expectations inline; that is how contradictions crept in.
`PruneVersionsTo(toVersion)` checks IN ORDER:
1. `toVersion >= latest` ⇒ PLAIN error "cannot prune latest version %d"
   (NOT a sentinel — match by string or just expect non-nil; includes
   `Prune(0)` before any save, since latest=0).
2. `toVersion < first` ⇒ nil no-op (includes negative toVersion).
3. dirty session ⇒ `ErrUncommittedChanges` — REACHABLE ONLY when
   `first ≤ toVersion < latest`; the dirty-prune cell must arrange an
   in-range target (then assert, then Rollback).
4. `t.version <= toVersion` ⇒ `ErrActiveReaders` (loaded-version guard).
5. held reader in [first, toVersion] ⇒ `ErrActiveReaders`; after release the
   same prune succeeds.
6. else nil; floor becomes toVersion+1; the model drops ALL per-version
   bookkeeping for the pruned versions — snapshots AND recorded root hashes
   (a hashes-forever map is the last unbounded structure in a forever-soak;
   no oracle needs a pruned version's hash).
Selectors: floor−1 / negative (cell 2), width-1, wide catch-up, latest−1
(cell 6), ≥latest (cell 1), dirty (cell 3), covering-loaded (cell 4),
held (cell 5).

Readers:
- `HoldSnapshot(v)` — `GetImmutable(v)` held; auto-released after the hold
  budget (K ops). Works on empty versions (nil-root snapshot) — still must be
  Closed.
- `SnapshotReads(v)` — GetImmutable + Get/Has + Close immediately.
- `IteratePartial(v)` — open, consume some, Close.

## Oracles
Cheap (every op): the return-value expectations above.
Full (after EVERY prune — measured ~0.15–1ms at this scale — and at chunk end):
1. **Garbage/over-deletion**: `assertNoGarbage` (landed in twinfix_test.go;
   cache-bypassing). The helpers take
   `testing.TB` so the engine can call them. Value
   check: full until the first `ExportImport`; afterwards use the
   **vk-version wall** — an unreferenced PrefixVal record is tolerated
   iff its vk version (first 8 bytes) < `maxImportVersion`; any unreferenced
   value AT/ABOVE the wall is a leak. (The naive "baseline count" is wrong:
   pre-import values leak only when their versions are later pruned, so the
   count grows after import — the wall is exact and monotone.)
2. **Per-version integrity**: every retained version — `GetImmutable(v)`
   hash == recorded; full Iterate == model snapshot (count + every kv);
   Close. Empty versions: hash == emptyHash, zero iterations.
3. **Proof oracle** (pinned mechanics): per retained NON-EMPTY version
   (empty ⇒ both proof kinds return `ErrEmptyTree` — skip), via the SAME
   `GetImmutable(v)` handle as oracle 2: ≤3 sampled present keys →
   `imm.GetMembershipProof(key)` then
   `ics23.VerifyMembership(BptreeSpec, recordedHash[v], proof, key, rawValue)`
   (raw model value — LeafOp prehashes it); 1 absent key →
   `imm.GetNonMembershipProof` then `ics23.VerifyNonMembership(BptreeSpec,
   recordedHash[v], proof, absentKey)`. Do NOT use the MutableTree proof
   wrappers (they prove only against lastSaved). Single-key trees are fine.
4. **Bookkeeping**: `AvailableVersions` == model's retained set;
   `versionReaders` empty when no holds outstanding; `countPinned(root)==1`
   after each save.
5. **Floor census** after a drain-to-empty prune: one root record, zero
   nodes/values/orphans (modulo the R2 wall if an import happened).
6. **Cold-restart equivalence** on `ColdRestart`.

## Seed corpus
Programs encoding: net-zero twin → prune; same-value rewrite → prune;
drain → empty-saves → prune-through-empties → regrow; repeated width-1
prunes through churn; grow/shrink height oscillation with mid-oscillation
prunes; import-then-prune; LoadOld + covering-prune + idempotent-save +
re-Load; injected error mid-prune → retry; held-snapshot blocks → release →
prune; separator-shift deletes (each leaf's first key) → prune.

## Explicit non-goals
- Out-of-contract concurrency (two writers; unregistered working-tree reads).
- Import to gaps/pruned versions (M18/M21 — filed; precondition lands with
  state-sync).
- Height-4 soak config (follow-up).
