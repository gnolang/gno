# Valset checkpoint replay

## Context

Hardfork ceremonies seed the new chain's genesis state by replaying a
migration `.jsonl` against gnoland's `InitChainer`. The migration mixes
two kinds of entries:

- **Genesis-mode entries** (`metadata == nil` or `BlockHeight == 0`)
  run against the InitChain ctx (`runtime.ChainHeight() == 0`), so
  realm guards keyed on "post-genesis" are bypassed.
- **Historical replay entries** (`metadata.BlockHeight > 0`) run
  against a per-tx ctx with the original block height set, so realm
  guards see the historical height.

For the source chains we hardfork from today, this works fine because
the source has neither `r/gnops/valopers` nor `r/sys/validators/v3`.
There are no historical txs that touch valoper state. The migration
just deploys v3+valopers via `addpkg` and Registers profiles fresh via
`gnogenesis fork valoper-seed` (genesis-mode txs).

For a **future** hardfork from a chain that already has v3+valopers in
its history, faithful replay would need to walk through every
historical `Register`, `UpdateSigningKey`, and v3 proposal-execute. The
current `InitChainer` design breaks this:

1. `InitChainer` seeds `valset:current` from `req.Validators` (the
   FINAL valset of the source chain) at line ~361 of `app.go`.
2. Historical `Register` at H=t for an addr that survives to FINAL
   trips the front-run guard (`IsValidator(signingAddr) == true`
   because the addr is already in the seeded final valset).
3. Historical `UpdateSigningKey` rotates against a valset that
   doesn't reflect the historical state at t.

The mitigation in place today is the documented pattern: hardfork
producers should NOT include historical valoper/v3 txs in the
migration `.jsonl`. Re-bootstrap via `valoper-seed` instead. This
loses pre-fork signingRegistry attribution but works for a single
fork boundary.

For the second hardfork — and any future chain that wants
cross-fork signing-key history (slashing attribution, key-reuse
blocking across forks) — the re-bootstrap pattern is too lossy.

## Decision (deferred)

Introduce a **valset checkpoint** primitive in the migration stream:
bare-metadata entries that hard-set valset state at known historical
boundaries, plus a per-realm `ImportValidators(valset)` entrypoint
gated by an ExecContext sentinel.

### Pieces

1. **Bare-metadata checkpoint tx**
   - Schema: `gnoland.TxWithMetadata{Tx: nil, Metadata: &GnoTxMetadata{Valset, Valrealm, ...}}`
   - `Tx` becomes pointer-typed (currently value); `nil` signals
     "checkpoint-only, apply via the runtime, do not Deliver".
   - New metadata fields:
     - `Valset []ValsetEntry` — the snapshot to install (each entry
       carries pubkey + power; format mirrors what's in
       `valset:current`).
     - `Valrealm string` — pkgpath of the realm whose
       `ImportValidators` is called with the snapshot. Empty means
       "params-only update; no realm sync."

2. **Effect of applying a checkpoint**
   - `valset:current` is hard-set to the snapshot via the existing
     `internalWriteCtxKey` sentinel (so `node:valset:*` validators
     accept the write).
   - If `Valrealm != ""`, `<Valrealm>.ImportValidators(snapshot)` is
     invoked via `vmk.Call` with a special ExecContext flag set.
   - The realm's `ImportValidators` is gated on the flag — it cannot
     be called from a normal tx, so the privileged-import semantics
     are unforgeable.

3. **ExecContext sentinel**
   - Add `GenesisValsetOverride bool` to `stdlibs.ExecContext` (or a
     dedicated key). Set true only inside the checkpoint-apply path.
   - Expose a stdlib helper: `runtime.IsGenesisValsetOverride()`.
   - `<valrealm>.ImportValidators` panics unless this returns true.

4. **`ImportValidators` realm function**
   - Each version of the validator realm (v1, v2, v3, ...) implements
     its own `ImportValidators(valset)`. The function syncs the
     realm's local state (valoperCache, signingRegistry,
     valset:proposed/dirty, etc.) to match the snapshot.
   - Idempotent: calling it twice with the same snapshot is a no-op.
   - Migration-only: gated on `runtime.IsGenesisValsetOverride()`.

### Replay shape

```
state.Txs = [
  checkpoint_initial,           // initial valset for the oldest chain in history
  ...replay_txs_chain_a...,     // historical txs from chain A
  checkpoint_chain_b_start,     // valset at A→B fork; valrealm=".../v2" if B introduced v2
  ...replay_txs_chain_b...,
  checkpoint_chain_c_start,     // valset at B→C fork; valrealm=".../v3" if C introduced v3
  ...replay_txs_chain_c...,
  checkpoint_final,             // matches GenesisDoc.Validators; valrealm=".../v3" (or current)
]
```

Multiple checkpoints can target the same `valrealm` if no realm
upgrade happened between forks (e.g., several `.../v3` syncs in a
row). The mechanism is per-snapshot, not per-realm-version.

After all `state.Txs` are consumed:
- `valset:current` equals `checkpoint_final.Valset`.
- `<latest valrealm>` realm state is synced to match.
- `assertGenesisValopersConsistent` (or its successor) verifies
  coverage one more time.

### Why this lifts the "no historical valoper txs" constraint

Without checkpoints, `valset:current` is seeded once with the FINAL
state at line 361 of `app.go`. Historical txs replay against an
already-final valset that doesn't reflect the historical timeline,
so guards trip.

With checkpoints, each segment of historical replay sees a
`valset:current` matching the historical state at that segment's
chain epoch. Front-run / squat / signingRegistry-uniqueness guards
pass naturally. signingRegistry accumulates the full cross-chain
history of every signing key ever rotated.

### Tradeoffs

- **Schema change**: `TxWithMetadata.Tx` becomes pointer-typed
  (nilable). All readers (gnogenesis, indexers, etc.) must handle
  the new shape.
- **Producer cost**: `gnogenesis fork generate` must query the
  source chain for valset state at every epoch boundary. Adds a
  dependency on RPC / archive access.
- **Realm contract**: every validator-realm version needs an
  `ImportValidators` function. Old versions (v1, v2 of valset
  proposal flow) need backports if the migration includes their
  segments.
- **Verification**: a misconfigured checkpoint stream could
  silently install a wrong valset. The post-replay assertion
  (matching `valset:current` against `GenesisDoc.Validators` and
  against valoperCache coverage) is the safety net.

## When to implement

Land this when the **second** hardfork is on the horizon — the one
where the source chain already has v3+valopers and operators have
rotated keys during its lifetime. For the immediate gnoland-1
hardfork, the simpler re-bootstrap pattern (`valoper-seed` Register
at genesis-mode, no historical valoper/v3 txs) is sufficient and
much smaller in scope.

Don't pre-build the schema or plumbing speculatively. The exact
shape of `Valset`/`Valrealm` and the `ImportValidators` contract
should be designed against the concrete needs of that future
ceremony, not guessed at now.

## Relationship to existing code

- `gno.land/pkg/gnoland/app.go` — InitChainer's `loadAppState` loop
  needs a branch for `tx.Tx == nil` to apply a checkpoint instead of
  `Deliver`.
- `gnoland.TxWithMetadata` (in `gno.land/pkg/gnoland/types.go` and
  amino-encoded variants) — `Tx` field becomes `*std.Tx`.
- `stdlibs.ExecContext` — add the override flag.
- `examples/gno.land/r/sys/validators/v3/cache.gno` — add
  `ImportValidators(snapshot)`. Subsumes the `seen`-set construction
  used by `AssertGenesisValopersConsistent`.
- `contribs/gnogenesis/internal/fork/generate.go` — emit checkpoints
  alongside historical txs based on a snapshot-query against the
  source chain.

## Out of scope

- Cross-fork validation rewinds. If a historical tx panicked on the
  source chain, replay still skips it via the `metadata.Failed` path
  documented in `pr5511_chain_upgrade_genesis_replay.md`.
- Multi-realm checkpoints in a single `Metadata`. One realm per
  checkpoint; if multiple realms need syncing at the same epoch
  boundary, emit multiple consecutive checkpoints.
- Live (post-boot) `ImportValidators` invocation. The ExecContext
  sentinel makes this impossible by construction.
