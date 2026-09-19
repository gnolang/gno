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

`params`, `auth`, `bank`, `gasprice` and `vm` all share the `main` store. There
is no per-module partition for a `VersionMap` to key on, and no substore to add,
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
3. **A record of completed upgrades** in the `main` store, checked at startup
   against the upgrade registry.

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

The code that performs one upgrade's state change, written once for a specific
release and run once, inside the block that follows the halt. Each is registered
under the version that introduced it — the value `halt_min_version` carries — so
naming that version in a proposal is what selects the code that runs.

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
	Version string  // the release tag; what halt_min_version carries
	Handler Handler // the code of the migration
}
```

Suggested layout, for clarity and isolation — each upgrade declares itself in
its own package, so its handler and whatever it touches stay reviewable on
their own:

```go
// gno.land/pkg/gnoland/upgrades/v1_3_0/upgrade.go

const Version = "v1.3.0"

var Upgrade = upgrades.Upgrade{
	Version: Version,
	Handler: migrate,
}

func migrate(ctx sdk.Context, env upgrades.Env) error { ... }
```

Registered in one place, as a package-level literal:

```go
// gno.land/pkg/gnoland/app.go

// A slice, not a map keyed by version: the key would restate Version, and
// declaration order is release order, which matters because upgrades apply in
// sequence. An init() somewhere could assert these invariants (version parses,
// strictly increasing, no duplicates...).
var Upgrades = []upgrades.Upgrade{
	v130.Upgrade,
	v140.Upgrade,
}
```

Read in two places, both covered below: `checkNodeStartupParams` (§3), to look up
the chain's last completed upgrade; and the BeginBlocker (*Where it hooks*), to
find the handler named by a pending `halt_min_version`.

An entry may be dropped once no chain the binary serves still sits at it — which
in practice means keeping the most recent, since that is what §3 looks up. More
may be needed: a binary able to replay across several upgrade heights must carry
the handler for each one it crosses. Whether it can is decided by whether those
upgrades changed the stored encoding — which nothing here marks, so it stays a
judgement for whoever prunes.

Only upgrades that migrate state register anything. A coordinated upgrade that
breaks consensus without touching state needs a floor and nothing else, so
`halt_min_version` is set and no entry exists.

A handler is arbitrary Go with `Env` in hand, so its reach is whatever the store
allows — deploy a package, rewrite params, seed data, replace the valoper set,
walk and re-encode objects. There is no fixed list, and no attempt to define
one. The only known limit is changing the code and/or the state of an
already-deployed realm, for the reason in Open questions.

### 3. Completed upgrades

The BeginBlocker records each upgrade it applies — version and height — under a
reserved key in the `main` store.

The record exists to answer one question at startup: **does this binary know the
shape the chain is already in?** Read the most recent completed upgrade and look
it up in the upgrade registry.

| last completed | binary | entry in the registry? | result |
|---|---|---|---|
| v1.5.0 | v1.3.0 | no — predates it | **refuse** |
| v1.5.0 | v1.5.0 | yes | proceed |
| v1.5.0 | v1.6.0 | yes, inherited | proceed |

An upgrade with an empty `halt_min_version` records nothing: there is no version
to record, and no handler ran. The chain's last completed upgrade stays whatever
it was before — which is correct, since nothing about the stored data changed.
A chain that has only ever halted this way, `gnoland-1` today, has no record at
all, and the check has nothing to compare, so every binary passes it. That is
the honest answer rather than a gap: without a version there is nothing to say
which binaries understand that chain.

`checkNodeStartupParams` (`node_params.go:129-181`, called once at `app.go:288`)
therefore grows one check beside the two it has:

| Check | When | Bypassed by `skip_upgrade_height`? |
|---|---|---|
| binary meets the floor (`:158-166`) | post-halt | yes, as today |
| binary does *not* meet the floor (`:170-178`) | pre-halt | yes, as today |
| the last completed upgrade names a registry entry | always | **no** |

The new check sits deliberately outside the escape hatch:

```bash
gnoland config set skip_upgrade_height 704052
```

That flag exists so an operator can assert "I already migrated out-of-band."
Skipping the *version* gates is a claim about which binary is running, and the
operator is entitled to make it. The new check is not a claim about the binary
but about the chain — which upgrades it has already applied — so it holds
regardless. Same for a node restored from a pre-upgrade backup, or one simply
offline across the upgrade.

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
  entry    → run the handler; record (version, height) as completed
  no entry → nothing to migrate
```

### Timeline

```
  H      EndBlocker arms the halt (existing code, unchanged)
  H+1    BeginBlock panics; nodes stop            [existing]
  ---    operators swap binaries
  boot   checkNodeStartupParams: version gates [existing]
                              + completed-upgrades check [new]
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

So the handler runs in-block, and everything that follows from that — no dry
run, determinism, no rollback — is a consequence rather than a choice.

## Consequences

**No dry run exists.** A handler is arbitrary code that commits inside a block
on every validator, and there is no way to rehearse one first. Something like
`gnoland upgrade dry-run --at <height>` against a copy of the data dir has to be
built alongside the first real handler.

**Determinism and bounded execution.** Every validator runs the handler inside a
block: byte-identical output, no OOM, no block-timeout blowout.

**There is no rollback.** A bad migration commits. Recovery is
restore-from-backup at the halt height, on every validator.

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

**A state-format counter**, an integer describing the shape of the stored data,
carried per upgrade and compared against a value derived from the upgrade
registry.
Rejected: the derived value is the maximum over the entries, so pruning one
lowers it and every node reads the chain as newer than the binary — refusing to
start, all at once. Avoiding that means a hand-maintained high-water constant
whose only failure mode bricks the chain, or keeping pruned entries as
headstones. Recording completed upgrades needs neither, and the presence test it
enables answers the same question.

Two counters, one per store, were considered alongside and rejected for the same
reason as per-module consensus versions: there is no independent axis. Both
stores are written by one binary, migrated by one handler that commits atomically
inside a block, and released together.

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

## Open questions

1. **Can a handler change the code of a deployed realm?** `addpkg` refuses a
   path that already exists. If that cannot be worked around from inside a
   handler, realm code changes are blocked on a VM change rather than on upgrade
   plumbing — and that is most of what an upgrade wants to do. This needs
   answering before anything else here is built.
2. **How the completed-upgrades record is read from outside.** Keeping it out of
   the params keeper costs the `gnokey query params/...` path, so there is no way
   to ask a node which upgrades it has applied. The `app` entry of the p2p
   `VersionSet` is the natural place, but it is fed by `state.AppVersion`, which
   nothing populates today.

## Phasing

1. Pick and build one of the three answers in *Syncing past an upgrade height*,
   starting with manual snapshots. Everything else is blocked on it.
2. Answer open question 1. Everything downstream depends on it.
3. The upgrade registry, the completed-upgrades record, the startup check and the
   BeginBlocker. No param changes — `halt_min_version` already carries what is
   needed.
4. First real handler, with a dry-run tool.

Worth folding into whichever phase touches the params: the halt params have **no
`.txtar` coverage at all** — zero hits across `gno.land/pkg/integration/testdata/`
and `misc/gnoe2e/testdata/` for `halt_height`, `halt_min_version`, `node:p:`,
`NewSetHaltRequest` or `set_halt`, while the valset half of the same keeper has
around ten (`params_valset_*.txtar`). Existing coverage is Go unit tests only:
`app_test.go:2476-2560` (`WillSetParam`), `:3442-3519`
(`TestCheckNodeStartupParams`, including the skip-bypass case) and `:3520-3592`
(`TestEndBlockerHalt`).
