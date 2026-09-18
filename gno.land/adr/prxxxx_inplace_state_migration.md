# In-place state migration for coordinated chain upgrades

## Status

Draft. Proposed design, nothing implemented.

## Context

### Where the chain is today

`gnoland-1` is live. The halt mechanism from [#5368](https://github.com/gnolang/gno/pull/5368)
works and has been exercised once in production: block 36170 executed a
`NewSetHaltRequest(cross(cur), 36300, "")`, the chain stopped after committing
36302, and validators brought it back 6m12s later. The halt shipped with an
**empty** `halt_min_version`, i.e. no restart gate at all — which was the only
safe choice, because on the binary of the day no version string would have
parsed. [#6177](https://github.com/gnolang/gno/pull/6177) fixes that.

So the chain can stop together. It has no way to *change state* while stopped
other than rebuilding it.

### What exists

| Piece | Where |
|---|---|
| `node:p:halt_height`, `node:p:halt_min_version` | `gno.land/pkg/gnoland/node_params.go` |
| Arming the halt | `app.go:1151-1162` (EndBlocker, `req.Height == haltHeight`) |
| Stopping | `tm2/pkg/sdk/baseapp.go:596` (BeginBlock panics at `haltHeight+1`) |
| Startup gates | `node_params.go:133-181` (`checkNodeStartupParams`) |
| Governance entry point | `examples/gno.land/r/sys/params/halt.gno:28` |
| Genesis replay | `gnogenesis fork *`, ADR `pr5511` |

### What does not exist

`gno.land/cmd/gnoland/UPGRADES.md` states it plainly: there is no migrate
function and no upgrade-handler framework. `halt_min_version` gates startup and
does nothing else. There is no way to change state at all.

### Not a complement to replay upgrades

Replay upgrades have two existing paths today, both rebuilding state by
re-executing history rather than transforming it:

- **into a fresh genesis** — built by `gnogenesis fork`: a new chain carrying the
  whole transaction history, re-executed under the new rules at `InitChain`.
  Never used on mainnet nor testnet.
- **in place** — [#5377](https://github.com/gnolang/gno/pull/5377)
  (`gnoland start --migrate`), which replays the local block store under the new
  binary and swaps the app DB. Still open.

This ADR does not treat either as an existing mechanism to extend, fall back on,
or interoperate with. In-place migration is the upgrade path; the rest of this
document assumes nothing else exists.

### Why Cosmos SDK's design does not port wholesale

Cosmos splits upgrades into three layers: a named plan (`x/upgrade`), named
handlers, and per-module `ConsensusVersion` + `VersionMap` migrations, plus
`StoreUpgrades` applied at store-load time.

Two of those exist for reasons gno does not share.

**Per-module consensus versions exist because modules are shipped to strangers.**
`x/bank`'s author cannot know what state any given chain holds, so the module
carries its own version and its own migration chain. gno has one `go.mod`
(`module github.com/gnolang/gno`) — `tm2/pkg/sdk/bank` cannot be imported
without the whole repository, and nothing does. Every keeper ships in one binary,
in one release, cut by the people who changed them.

**Store upgrades operate on module substores.** gno mounts exactly two, both
permanent:

```go
mainKey := store.NewStoreKey("main")   // params, auth, bank, gasprice, vm's iavl side
baseKey := store.NewStoreKey("base")   // the VM object graph, raw dbadapter
```

`params`, `auth`, `bank`, `gasprice` and `vm` all share `main`. There is no
per-module partition for a `VersionMap` to key on, and no substore to add,
rename or delete. `StoreUpgrades`, `StoreLoader` and `upgrade-info.json` have
nothing to do here.

What does carry over is the part that has nothing to do with modularity: a
binary that opens a store written by an older binary needs to know what it is
reading.

## Decision

Three pieces, in order of independence:

1. **No new param.** The existing `halt_min_version` carries both roles: the
   version floor it already is, and the key an upgrade handler is registered
   under.
2. **Handlers** — arbitrary Go, registered under the release version that
   introduced them, compiled in, run inside a block.
3. **A single state-format version** — one integer in the binary, one in `main`,
   compared at startup.

Explicit non-goals, each dropped for the reasons above:

- per-module `ConsensusVersion`, `VersionMap`, `module.Manager`, `Configurator`,
  `RegisterMigration`;
- `StoreUpgrades`, `SetStoreLoader`, `upgrade-info.json`;
- multi-step catch-up across skipped releases (upgrades apply in sequence);
- `InitGenesis`-for-absent-module (new keeper state is seeded explicitly in a
  handler).

## Design

### 1. The plan

The two params that exist today, unchanged in name, type and wire format:

| Param | Type | Meaning |
|---|---|---|
| `node:p:halt_height` | `int64` | height to stop at; `0` cancels |
| `node:p:halt_min_version` | `string` | version floor **and** upgrade registry key |

`NewSetHaltRequest(cur, height, minVersion)` keeps its signature. Nothing about
the public surface moves.

An empty `halt_min_version` leaves no key to look up, so an upgrade on an
unversioned binary carries no migration.

`WillSetParam` (`node_params.go:69-73`) currently type-checks the string and
nothing else. It should additionally reject a value the node cannot parse, so a
floor that can only ever be met by byte equality cannot reach state.

### 2. Handlers

```go
// gno.land/pkg/gnoland/upgrades

// Env is everything a handler may reach. No such aggregate exists today: the
// keepers and store keys are locals in NewAppWithOptions, so building this and
// threading it to the BeginBlocker is part of the work.
type Env struct {
	Params   params.ParamsKeeperI
	Account  auth.AccountKeeperI
	Bank     bank.BankKeeperI
	GasPrice auth.GasPriceKeeperI
	VM       *vm.VMKeeper

	// Store keys, for what the keepers do not expose. Keepers alone are not
	// enough: ctx.Store(BaseKey) is the VM object graph, and walking it to
	// re-encode objects has no keeper method behind it.
	MainKey store.StoreKey
	BaseKey store.StoreKey
}

type Handler func(ctx sdk.Context, env Env) error

type Upgrade struct {
	Version          string  // the release tag; what halt_min_version carries
	Handler          Handler // the code of the migration
	StateFormatAfter int64   // the format version this upgrade leaves behind
}
```

Suggested layout, for clarity and isolation — each upgrade declares itself in
its own package, so its handler and whatever it touches stay reviewable on
their own:

```go
// gno.land/pkg/gnoland/upgrades/v1_3_0/upgrade.go

const Version = "v1.3.0"

var Upgrade = upgrades.Upgrade{
	Version:          Version,
	Handler:          migrate,
	StateFormatAfter: 4,
}

func migrate(ctx sdk.Context, env upgrades.Env) error { ... }
```

Registered in one place, as a package-level literal:

```go
// gno.land/pkg/gnoland/app.go

// A slice, not a map keyed by version. Declaration order is release order,
// which matters because upgrades apply in sequence and because the newest
// entry is always the last one. An init() somewhere could assert these
// invariants (version parses, strictly increasing, no duplicates...).
var Upgrades = []upgrades.Upgrade{
	v130.Upgrade,
	v140.Upgrade,
}
```

Read in two places, both covered below: `checkNodeStartupParams` (§3), to derive
the expected state-format version and to resolve a pending `halt_min_version`;
and the BeginBlocker (*Where it hooks*), to find the handler to run at `H+1`.

An entry may be dropped once nothing can replay the block that ran its handler.

Only upgrades that migrate state register anything. A coordinated upgrade that
breaks consensus without touching state needs a floor and nothing else, so
`halt_min_version` is set and no entry exists — the state-format versions agree,
so nothing looks one up.

A handler is arbitrary Go with `Env` in hand, so its reach is whatever the store
allows — deploy a package, rewrite params, seed data, replace the valoper set,
walk and re-encode objects. There is no fixed list, and no attempt to define
one. The only known limit is changing the code and/or the state of an
already-deployed realm, for the reason in Open questions.

### 3. State format version

- a value in the binary, *derived* as the maximum `StateFormatAfter` across the
  upgrade registry of §2 rather than written by hand — #6177 exists largely
  because six protocol constants had to be kept equal by hand, and a bump that
  missed one panicked every node at startup. Do not repeat that shape;
- an `int64` in `main` under a reserved key, written by the handler;
- compared in `checkNodeStartupParams`, which already reads committed state
  before any block runs.

| stored vs binary | action |
|---|---|
| equal | proceed |
| stored < binary | proceed **only if** a pending `halt_min_version` names an entry in the upgrade registry closing the gap; otherwise refuse |
| stored > binary | refuse — downgrade onto newer data |

This is the piece that has no substitute today. `halt_min_version` gates which
*binary* may start; even carrying the upgrade registry key, it says nothing
about the *data*. Nothing compares what the binary expects against what is on
disk.

`checkNodeStartupParams` (`node_params.go:129-181`, called once at `app.go:288`)
therefore grows one check beside the two it has:

| Check | When | Bypassed by `skip_upgrade_height`? |
|---|---|---|
| binary meets the floor (`:158-166`) | post-halt | yes, as today |
| binary does *not* meet the floor (`:170-178`) | pre-halt | yes, as today |
| stored state-format version vs the binary's | always | **no** |

The new check sits deliberately outside the escape hatch:

```bash
gnoland config set skip_upgrade_height 704052
```

That flag exists so an operator can assert "I already migrated out-of-band."
Skipping the *version* gate is a claim about which binary is running, and the
operator is entitled to make it. Skipping the *migration* means running new code
on unmigrated data — the divergence the format version exists to catch. So the
flag keeps its meaning for the version checks, and the operator's claim becomes
verified by the format check rather than trusted. Same for a node restored from a
pre-upgrade backup, or one simply offline across the upgrade.

### Where it hooks

`baseApp.SetBeginBlocker` (`tm2/pkg/sdk/options.go:68`) exists and gno.land does
not use it — `app.go` sets only `InitChainer`, `AnteHandler`, tx hooks and
`EndBlocker`. `BaseApp.BeginBlock` calls it at `baseapp.go:622`, after the halt
check (`:596`) and after `deliverState` is prepared. 

The new `BeginBlocker` goes in `app.go` beside the existing `EndBlocker`, built
the same way: a constructor taking what it needs — `prmk` and the `Env` — and
returning the `sdk.BeginBlocker` closure (`tm2/pkg/sdk/abci.go:12`), wired with
`baseApp.SetBeginBlocker(...)` next to the `SetEndBlocker` call at `app.go:262`.

It runs at `req.Height == haltHeight+1`, mirroring the EndBlocker's existing
exact-height arming at `app.go:1151-1162`:

```
read node:p:halt_min_version
  ""       → nothing to look up; no migration possible
  entry    → run the handler; write state_format_version;
             record (version, height); clear both params
  no entry → nothing to migrate; clear both params
```

**The BeginBlocker needs to clear both params**. Left set, they are read
forever: `halt_min_version` keeps naming an entry in the upgrade registry
that every later startup looks up, which pins that entry in the binary for the
life of the chain and turns any later pruning into a refused startup on every
node at once. It also leaves a version floor standing that nobody chose to
keep.

It also reverses documented behaviour, so `UPGRADES.md` has to be fixed in the
same change: it currently states that the halt params are never cleared and that
`halt_min_version` therefore "stays in force as a permanent minimum-version floor
for every later restart".

### Timeline

```
  H      EndBlocker arms the halt (existing code, unchanged)
  H+1    BeginBlock panics; nodes stop            [existing]
  ---    operators swap binaries
  boot   checkNodeStartupParams: version gate [existing]
                              + registry gate + format gate [new]
  H+1    BeginBlocker runs the handler, in consensus
```

### Why the handler must run inside a block

It is tempting to migrate at startup, outside consensus, so that block history
stays replayable by any binary. It does not work: a node syncing from genesis
rebuilds state by executing blocks, so it arrives at `H` in the *old* format and
must apply the same transformation at the same boundary to reach the same
AppHash. The handler is needed either way — and running it outside consensus
means a node that migrates differently diverges silently instead of failing at a
block.

So the handler runs in-block, and the cost below is not a choice.

## Consequences

**No dry run exists.** A handler is arbitrary code that commits inside a block
on every validator, and there is no way to rehearse one first. Something like
`gnoland upgrade dry-run --at <height>` against a copy of the data dir has to be
built alongside the first real handler.

**Determinism and bounded execution.** Every validator runs the handler inside a
block: byte-identical output, no OOM, no block-timeout blowout.

**There is no rollback.** A bad migration commits. Recovery is
restore-from-backup at the halt height, on every validator.

**The permanent version floor goes away.** `UPGRADES.md` documents today's
behaviour — the params are never cleared, so `halt_min_version` "stays in force
as a permanent minimum-version floor for every later restart" — and this reverses
it. The protection is not lost, it moves: a stale binary is now caught by the
state-format check, which compares what the code expects against what is on disk
instead of against a string someone typed into a proposal.

## Syncing past an upgrade height

A node joining the network replays history, and an upgrade height is where that
breaks. Two distinct problems:

**1. The encoding changed.** Once the encoding of a stored value moves, only code
that knows the old encoding can replay the blocks that came before it, because
the encoding impacts the AppHash. A node replaying block 5 with a binary that
encodes differently computes an AppHash the committed header does not agree with,
and sync fails.

**2. The halt is still armed.** The EndBlocker arms on `req.Height == haltHeight`
from committed state, and `halt_height` is never cleared. A node syncing
`gnoland-1` from genesis therefore executes block 36170 (which sets
`halt_height = 36300`), arms at 36300, and stops entering 36301 — the same halt
the original upgrade produced. Block 36300 is committed first and
`app.haltHeight` is in-memory, so a restart gets past it: one manual restart per
historical halt height.

### How Cosmos handles it

Twice:

| Node | Mechanism |
|---|---|
| ordinary | **state sync** — restore a recent snapshot, never execute a pre-upgrade block |
| archive, from genesis | **cosmovisor** — read `upgrade-info.json`, swap to `upgrades/<name>/bin/<daemon>`, restart; blocks 1..H1 run under v1, H1..H2 under v2 |

The binary boundary does the versioning. Note what this makes of problem 2: the
stop is **the signal, not a fault** — it is what tells cosmovisor which binary to
run next.

### How gno could handle it

gno has neither mechanism: no ABCI state-sync surface in `tm2` (`ListSnapshots`,
`OfferSnapshot`, `LoadSnapshotChunk`, `ApplySnapshotChunk` are absent, so there
is nothing a `gnoland snapshot dump` could dump), and nothing cosmovisor-shaped
in `contribs/`. One of the following must exist before any upgrade may change a
stored encoding, in increasing order of cost:

- **Manual snapshots.** No code at all: archive a stopped node's data dir — app
  DB, TM state DB, block store — and a joining node unpacks it and syncs forward,
  never executing a pre-upgrade block. Several Cosmos chains run exactly this,
  outside the protocol. The cost is trust: the tarball is unverified, so
  operators trust whoever published it, where state sync verifies chunks against
  the AppHash.
- **A supervisor.** Preserves sync-from-genesis, which is the only way to join
  today. It needs the halt to be machine-detectable and a per-upgrade binary
  layout.
- **State sync.** The largest build, and the only one that makes
  sync-from-genesis optional rather than mandatory — which is in turn the only
  thing that would let historical encodings be dropped rather than kept.

Whichever is chosen, problem 2 must not be "fixed" by suppressing the arming
during catch-up. That lets a single binary sync straight through, producing
exactly the replay that cannot reach the right AppHash. What the stop needs is to
become machine-readable, so a supervisor can act on it rather than the node
idling until an operator notices.

*Read from `app.go:1151-1162` and `baseapp.go:596`; confirm against a real sync
of `gnoland-1` before acting on it.*

## Alternatives considered

**Port Cosmos's module layer wholesale.** Rejected: builds a module manager,
configurator and per-module registry to serve four keepers sharing one store
key, for a modularity property gno does not have.

**Two state-format versions, `main` and `base` separately.** Rejected, for the
same reason as per-module consensus versions: there is no independent axis. Both
stores are written by one binary, migrated by one handler that commits atomically
inside a block, and released together — so they cannot drift apart, and no
consumer exists to ask about one without the other. Nor does a second counter
save work: a handler already knows which store it touched.

**A separate `node:p:upgrade_name` param**, as Cosmos has. Rejected. It would
break `NewSetHaltRequest`'s public realm signature and add a second source of
truth for the question `halt_min_version` already answers.

Cosmos's `Plan` is `{Name, Height, Info}` and has never had a version field, but
that is unavailability rather than preference: `x/upgrade` ships to hundreds of
chains with their own binaries and tag schemes, so the SDK has no equivalent of
`tm2/pkg/version.Version` to read, and no guarantee any chain's string is even
semver. A name is opaque, so it works everywhere. gno has exactly what the SDK
lacks — one binary, one release process, one version string it controls and can
parse — so the name would carry no information the version does not.

### The release version as the format version

Undecided — recorded to be settled before implementation. Drop
`StateFormatAfter` and the derived integer entirely; store the `Version` of the
last applied upgrade, and take the expected value as the newest `Version` in the
upgrade registry.

Orthogonal to the rest of the design: the check tables and the post-halt scoping
are identical either way.

#### What it removes

1. **The `StateFormatAfter` field, entirely.** Every upgrade implicitly advances
   the format to its own version. Nothing to choose per entry.
2. **The derivation step.** The expected value is just the newest `Version` in
   the upgrade registry.
3. **A whole numbering scheme.** Nobody has to answer "does this upgrade bump the
   format, or leave it where it was?"

#### What it improves

4. **The tripwire stops being forgettable.** The strongest argument.
   `StateFormatAfter` is hand-chosen per entry — write an upgrade that changes the
   VM object encoding, leave `StateFormatAfter` at its predecessor's value, and the
   check silently never fires. Deriving the constant from the upgrade registry
   fixes drift
   between constant and entries; it does nothing about a wrong value *on* an entry.
   With versions there is no value to get wrong, so the error class disappears.
5. **It merges with the applied-upgrade record.** The handler already records
   `(version, height)`. If the stored value *is* the version, those are one piece
   of state instead of two — the same consolidation applied to the params.
6. **Legibility.** A refused startup reads `stored v1.3.0, expected v1.4.0`
   rather than `stored 3, expected 4`, and the operator can match it directly
   against `gnoland version`.
7. **Stricter by default.** An upgrade that changes no data layout still advances
   the pointer, so "you skipped an upgrade" becomes detectable, not only "your
   data shape is wrong".

#### What it costs

8. **The format check gains a parser dependency.** Today an integer compare in
   the startup path; it becomes `meetsMinVersion`-style ordering. Fine post-#6177,
   but it adds a failure mode — an unparseable stored value has no safe
   interpretation, where `0` always did.
9. **Registry order must equal semver order.** True today, but it becomes
   load-bearing rather than incidental.
10. **"This upgrade changes nothing on disk" becomes inexpressible.** With
    integers, reusing the predecessor's number states it explicitly. With versions
    every entry advances.
11. **Two copies of the same string in state.** After the v1.3.0 upgrade,
    `halt_min_version` is `"v1.3.0"` and the stored format is `"v1.3.0"`. Harmless,
    and they are meaningfully different facts — what governance asked for versus
    what actually ran, which diverge exactly when a migration half-ran — but it
    reads as redundancy.

#### What needs deciding

12. **The zero value.** `0` is a natural "no upgrades applied" for a fresh chain.
    `""` has to be defined, and must sort below every real version including
    betanet's `chain/gnolandX.Y` shape.
13. **What an upgrade with no entry writes.** With integers it writes nothing —
    the formats already agree. With versions the stored value would have to
    advance anyway, so the BeginBlocker writes it even when there is no handler
    to run.

## Open questions

1. **Can a handler change the code of a deployed realm?** `addpkg` refuses a
   path that already exists. If that cannot be worked around from inside a
   handler, realm code changes are blocked on a VM change rather than on upgrade
   plumbing — and that is most of what an upgrade wants to do. This needs
   answering before anything else here is built.
2. **Where the reserved `main` key lives** so it cannot collide with a realm or
   a module prefix.

## Phasing

1. Pick and build one of the three answers in *Syncing past an upgrade height*,
   starting with manual snapshots. Everything else is blocked on it.
2. Answer open question 1. Everything downstream depends on it.
3. State-format version + startup gate. Useful alone: it makes any future
   mismatch loud, and commits to nothing else.
4. The upgrade registry, its startup gate, and the BeginBlocker. No param
   changes — `halt_min_version` already carries what is needed.
5. First real handler, with a dry-run tool.

Worth folding into whichever phase touches the params: the halt params have **no
`.txtar` coverage at all** — zero hits across `gno.land/pkg/integration/testdata/`
and `misc/gnoe2e/testdata/` for `halt_height`, `halt_min_version`, `node:p:`,
`NewSetHaltRequest` or `set_halt`, while the valset half of the same keeper has
around ten (`params_valset_*.txtar`). Existing coverage is Go unit tests only:
`app_test.go:2476-2560` (`WillSetParam`), `:3442-3519`
(`TestCheckNodeStartupParams`, including the skip-bypass case) and `:3520-3592`
(`TestEndBlockerHalt`).
