# ADR: GovDAO-Scheduled Chain Halt via `r/sys/params`

## Context

A coordinated chain upgrade needs every validator to stop at the *same* height.
If some nodes keep producing blocks past the agreed cut, the new binary starts
from a state the rest of the network never saw, and the upgrade becomes a fork.

### What existed before

#5334 added a per-node `halt_height` field to `config.toml`
(`tm2/pkg/bft/config/config.go:310`), wired in `gno.land/cmd/gnoland/start.go:287`
and enforced in `tm2/pkg/sdk/baseapp.go`. It works, but it has two gaps:

1. **Coordination is out-of-band.** Every operator has to edit their own config
   with the same number. One operator who misses the message keeps producing
   blocks past the cut; one who fat-fingers the number halts early. There is no
   on-chain record of the agreed height, so there is nothing to check against.

2. **Nothing keeps the wrong binary out.** After the halt, an operator can
   restart the *old* binary and rejoin. That is precisely the divergence the
   halt was meant to prevent — the halt buys a synchronized stopping point and
   then immediately gives it away.

### Why not just tell operators the number

Because the number is a consensus-relevant decision and the network already has
a mechanism for those. Any solution that lives only in each operator's config
file is advisory; the chain cannot enforce it, and post-mortem it cannot even
tell you what was agreed.

## Decision

Make the halt height **governance-owned chain state**, and add a binary-version
floor that gates restart. Keep the config field as a per-node override.

### Params keys

Two scalars under the existing `node` module in the params store:

| Key | Type | Meaning |
|---|---|---|
| `node:p:halt_height` | `int64` | Height to halt at. `0` = no halt / cancel. |
| `node:p:halt_min_version` | `string` | Minimum `tm2/pkg/version.Version` required to restart once the halt is reached. `""` = no version gate. |

They are written by `r/sys/params.NewSetHaltRequest(cur, height, minVersion)`
(`examples/gno.land/r/sys/params/halt.gno`), which builds a GovDAO
`ProposalRequest` whose executor sets both keys and emits a `set_halt` event.
`height == 0` is the cancel sentinel.

`nodeParamsKeeper.WillSetParam` (`gno.land/pkg/gnoland/node_params.go:53`)
validates at write time: `halt_height` must be an `int64`, non-negative, and —
unless it is the `0` sentinel — strictly greater than the current block height.
`halt_min_version` must be a `string`. Unknown `p:` keys panic, so governance
cannot write arbitrary keys into the `node` namespace.

### Arming and halting

The halt is a two-step handoff from chain state to the ABCI app:

1. **EndBlocker arms it** (`gno.land/pkg/gnoland/app.go:1102`). At the end of
   every block it reads `node:p:halt_height` and, if
   `req.Height == haltHeight`, calls `baseApp.SetHaltHeight(haltHeight)`.
   `SetHaltHeight` is the only method the `endBlockerApp` interface needs for
   this (`app.go:1022`).

2. **BeginBlock fires it** (`tm2/pkg/sdk/baseapp.go:590`). It panics when
   `header.Height > haltHeight`.

So the block **at** `halt_height` is fully committed, and the panic lands at the
start of `halt_height + 1`. The last committed block is `halt_height`.

The comparison in step 1 is `==`, not `>=`, and that matters: with `>=`, a node
restarted after the halt would re-arm on its very first block and panic again,
forever. With `==`, the condition is false for every `req.Height > haltHeight`,
so the halt fires exactly once and a compliant binary can resume.

The panic is caught by the consensus routine's recover
(`tm2/pkg/bft/consensus/state.go:639`), which logs `CONSENSUS FAILURE!!!`,
stops the WAL, and exits `receiveRoutine`. The process itself keeps running —
`gnoland start` blocks on its signal context (`start.go:316`), not on consensus
health — so the node stops making blocks but the operator still has to stop it.

### The two startup checks

`checkNodeStartupParams` (`node_params.go:134`) runs once in
`NewAppWithOptions`, after `LoadLatestVersion`, against committed state
(`app.go:284`). It returns early — no checks at all — when `halt_height == 0`
or `halt_min_version == ""`. Otherwise it compares `tm2/pkg/version.Version`
against `halt_min_version` in one of two directions:

- **Check 1 — `lastBlockHeight >= haltHeight`** (the halt has happened). Refuse
  to start unless the binary meets the minimum version. This is what keeps a
  retired binary from resuming the chain.
- **Check 2 — `lastBlockHeight < haltHeight`** (the halt has not happened yet).
  Refuse to start *any* binary that already meets the minimum version. The
  upgrade binary is for after the cut; running it early is the same divergence
  from the other side.

`skip_upgrade_height` in `config.toml` (`config.go:314`, threaded through
`AppOptions.SkipUpgradeHeight`) bypasses both when it equals `haltHeight` — the
escape hatch for an operator who has already migrated out-of-band.

### Version comparison

`meetsMinVersion` (`node_params.go:193`) parses `chain/gnoland<major>.<minor>`:
if the majors differ it compares majors, otherwise it requires
`minor >= minMinor`. If *either* side fails to parse, it falls back to exact
string equality.

The fallback is deliberately strict rather than permissive, but it has a sharp
edge worth knowing: `gno.land/Makefile` builds inject a `git describe`-derived
version (`master.12345+abc1234`), which never parses as a chain version. Such a
binary satisfies only a byte-identical `halt_min_version`, so in practice it
fails any `chain/gnolandX.Y` requirement. Testing this feature requires
building with an explicit `-ldflags` version; see
[`docs/resources/test-halt-height.md`](../../docs/resources/test-halt-height.md).

## Alternatives Considered

- **Keep the config-only halt from #5334.** Rejected for the two gaps above.
  It is *kept*, but as an alternative rather than a complement: both write the
  same `BaseApp.haltHeight` field, so a node with both set halts at whichever
  was written last, and a stale config value re-applied at startup can strand
  the node entirely. It is the lever for cases governance cannot reach — local
  forks, replay tooling, stopping a single node in an emergency — and should be
  cleared afterwards.

- **Halt inside `Commit` instead of the next `BeginBlock`.** Cutting after
  writing block N inside `Commit` is the same logical cut point, but panicking
  mid-`Commit` risks stopping between `deliverState.ms.MultiWrite()` and
  `cms.Commit()` — an ordering the code explicitly depends on. Checking at the
  next `BeginBlock` puts the panic outside every write path. (The doc comment
  on `BaseApp.Commit` still describes the older deferred-in-`Commit` design and
  is stale.)

- **Halt *at* `halt_height` rather than after it** — i.e. never commit the
  block at the halt height. Raised in review. Rejected because it re-introduces
  the restart loop from the other direction: if the block never commits, a
  restarted node sees the same last height, arms again, and panics again, with
  no way forward. Committing `halt_height` and panicking at `halt_height + 1`
  leaves the chain cleanly stopped on a committed block, which is also the state
  the replay tooling expects.

- **`osm.Kill()` — send the node's own process `SIGTERM`.** The governance
  halt's first implementation. It ends the process, which is friendlier to
  operators and to process supervisors, but it stops the node through an async
  signal delivered outside block execution. Review replaced it with
  `BaseApp.SetHaltHeight()` so the governance halt shares the deterministic
  `BeginBlock` path from #5334: nodes stop at the same block because the same
  check runs during block execution, not because a signal landed at some point.
  The process surviving its own halt is the cost of that (see Consequences).

- **`>=` in the EndBlocker check.** Simpler to read, and wrong: it makes the
  halt unrecoverable, as described above.

- **Clear the halt params automatically once a compliant binary resumes.**
  Considered and not implemented. Auto-clearing would make `halt_min_version` a
  one-shot gate; leaving it set makes it a persistent floor, which is the more
  useful property (see Consequences). Clearing also means a chain-internal
  write to a governance-owned key, which the `node` module otherwise only
  permits for the `valset:` mirrors.

- **A migrate function linked to `halt_min_version`.** Raised in review on
  #5368: if the new binary carries a state migration, have the node look for a
  migration keyed to `halt_min_version` and run it on resume. Deferred at the
  time as needing "a more complete upgrade handler framework", which #5374
  (`meta: chain hardfork`) then specced — every hardfork defined by
  source + binary + overlay, where the overlay could be numbered shell scripts
  *or* "named Go migration functions registered in the binary".

  It never landed, and neither did the tooling #5374 listed alongside it:
  #5411, #5376 and #5369 were all closed unmerged, and #5374 itself is closed.

  #5377 (`--migrate`, in-place block-replay migration) is still open, but it is
  a different mechanism rather than this hook. It has the operator rebuild app
  state by re-executing the existing block store with the new binary — state
  rebuild by re-execution, not state transformation — and its extension point,
  `MigrationConfig.GenesisOverlay`, patches the *genesis doc* before replay
  instead of running off `halt_min_version`. Its own ADR proposes
  auto-triggering the rebuild when `node:p:halt_height` matches the block store
  height and defers that too, on the grounds that the params-based halt height
  was not yet merged. So #5368 deferred the trigger forward to an upgrade
  framework and #5377 deferred it back to these params; neither built it.

  #5377 then lost its rationale. It had rejected "export-then-genesis" because
  block heights would reset and indexers would lose continuity — but #5540's
  `GenesisDoc.InitialHeight` made a replayed chain resume at the source chain's
  halt height + 1, and #5511 shipped that route.

  The upgrade path that shipped instead is genesis replay
  (#5511, #5540 — see
  [pr5511_chain_upgrade_genesis_replay.md](./pr5511_chain_upgrade_genesis_replay.md)),
  which expresses migrations as *transactions appended to the new chain's
  genesis* (`gnogenesis fork generate --migration-tx`, plus `--patch-realm`)
  rather than as Go functions in the halting binary. Migration therefore
  happens on the new chain during replay, not in the binary that halted, so
  there is nothing left for `halt_min_version` to hook into. Anyone reviving
  the idea should start from the genesis-replay design rather than from this
  params pair.

- **A dedicated upgrade module (Cosmos SDK `x/upgrade` style).** Rejected as
  overkill for a coordinated stop: two scalars in the params store already
  cover it, without a new module, keeper, and store prefix. The migration
  half of what such a module would provide went to the genesis-replay tooling
  instead, as described in the previous bullet.

## Consequences

- **The halt params are never cleared.** Nothing in the node clears them; only
  a new proposal can. Three effects follow:
  - `halt_min_version` becomes a permanent minimum-version floor. Every later
    restart re-runs check 1, so binaries below it stay off the chain
    indefinitely. Useful; surprising if you expected a one-shot gate.
  - Check 2 stops applying once `lastBlockHeight >= haltHeight`, so newer
    binaries can be started freely after the halt.
  - Scheduling the next halt needs a fresh proposal, and `WillSetParam`
    guarantees the new height is in the future. Lifting the version floor needs
    `NewSetHaltRequest(cross(cur), 0, "")`.

- **A halt is a supermajority action with an immediate chain-wide effect.**
  `WillSetParam` bounds the height to the future but not to the *near* future:
  a proposal executing at height H can legally set `H+1`, halting the chain on
  the next block. There is no minimum notice period.

- **An empty `halt_min_version` disables both startup checks.** The halt still
  happens, but nothing stops the old binary from resuming. Setting a min
  version is what buys the anti-divergence protection; a halt proposal that
  leaves it empty is a coordinated pause, not an upgrade gate.

- **The stop is a crash, not a graceful shutdown.** Operators see
  `CONSENSUS FAILURE!!!` in the logs, the process idles with its RPC server
  still up, and they must stop it themselves. Process supervisors that restart
  on exit will not trigger, because the process does not exit. This follows
  directly from choosing the `BeginBlock` panic over `osm.Kill()` (see
  Alternatives) — determinism was preferred to a clean process exit.

- **`skip_upgrade_height` is narrower than its origin suggests.** It was added
  in #5368 to answer "skip the migrate function, for a validator that has
  already migrated its state out-of-band". Since that migrate function was
  never built (see Alternatives), the flag only bypasses the two version
  checks. Its `config.go:314` comment describes what it does; the broader
  intent is unimplemented.

- **Adding a node param now requires touching `node_params.go`**, since
  `WillSetParam` panics on unknown `p:` keys. That is the intended trade: an
  allowlist of node-level knobs, at the cost of a Go-side change per new key.
