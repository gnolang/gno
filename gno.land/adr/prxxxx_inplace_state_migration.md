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
the public surface moves: the two `gnokey query params/node:p:*` paths documented
at `UPGRADES.md:58-59` and scripted throughout `UPGRADES-TESTING.md`, the
`set_halt` event payload keys `height` / `min_version`, and the realm API pinned
by `examples/gno.land/r/sys/params/params_test.gno:21,28,35`.

That matters more than it looks. An `upgrade_name` param would have been a
breaking change to a public realm API for no capability that `halt_min_version`
cannot express — see Alternatives.

An empty value leaves no key to look up, so an upgrade on an unversioned binary
carries no migration.

`WillSetParam` (`node_params.go:69-73`) currently type-checks the string and
nothing else. It should additionally reject a value the node cannot parse, so a
floor that can only ever be met by byte equality cannot reach state.

### 2. Handlers

```go
// gno.land/pkg/gnoland/upgrades

type Handler func(ctx sdk.Context, k Keepers) error

type Upgrade struct {
	Version          string  // the release tag; what halt_min_version carries
	Handler          Handler // nil for a version-gate-only upgrade
	StateFormatAfter int64   // the format version this upgrade leaves behind
}
```

Registered in one place, compiled in, **never removed** (see Consequences).

`Version` is the release that **introduced** the upgrade — not "the version you
must be running". The param's value is used by two different operations:

```go
meetsMinVersion(tmver.Version, minVersion)  // ordered: binary vs param
registry[minVersion]                        // exact: param only
```

The binary's own version never enters the lookup. A v1.3.1 binary resolves
`"v1.3.0"` because the registry is a source-level literal and v1.3.1 is cut from
a commit that still declares that entry. So an already-voted proposal survives a
patch release cut before the halt height.

**That covers the proposal, not the binaries.** If the patch fixed the handler
itself, two nodes clear the floor and hold the key but run different code under
it, compute different state at `H+1`, and split on AppHash. Nothing here detects
that — validators must agree on the exact binary, not just the floor.

A `nil` Handler is legitimate: a coordinated upgrade that breaks consensus
without changing state format still needs a floor, and registering it keeps a
mistyped version a *refused startup* rather than a silently skipped migration.

A handler is arbitrary Go with the keepers in hand, so its reach is whatever the
store allows — deploy a package, rewrite params, seed data, replace the valoper
set, walk and re-encode objects. There is no fixed list, and no attempt to define
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
| stored < binary | proceed **only if** a pending `halt_min_version` names a registry entry closing the gap; otherwise refuse |
| stored > binary | refuse — downgrade onto newer data |

This is the piece that has no substitute today. `halt_min_version` gates which
*binary* may start; even carrying the registry key, it says nothing about the
*data*. Nothing compares what the binary expects against what is on disk.

`checkNodeStartupParams` (`node_params.go:129-181`, called once at `app.go:288`
after `LoadLatestVersion` and before the app serves anything) therefore grows two
checks beside the two it has:

| Check | When | Bypassed by `skip_upgrade_height`? |
|---|---|---|
| binary meets the floor (`:158-166`) | post-halt | yes, as today |
| binary does *not* meet the floor (`:170-178`) | pre-halt | yes, as today |
| non-empty `halt_min_version` names a registry entry | **post-halt only** | **no** |
| stored state-format version vs the binary's | always | **no** |

The two new checks sit deliberately outside the escape hatch:

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
check (`:596`) and after `deliverState` is prepared. That is the same position
Cosmos gives its `PreBlocker`.

At `req.Height == haltHeight+1` — mirroring the EndBlocker's existing
exact-height arming at `app.go:1151-1162`:

```
read node:p:halt_min_version
  ""                       → nothing to look up; no migration possible
  entry, Handler non-nil   → run it; write state_format_version;
                             record (version, height)
  entry, Handler nil       → write state_format_version; record
  no entry                 → unreachable: startup already refused
```

Note the last line. Cosmos needs "panic if no handler" because the missing
handler *is* its halt. gno gates at startup instead, so a stale binary is
refused before any block with a real error message rather than as a consensus
panic on a running node. The in-block branch is a belt-and-braces assertion, not
the mechanism.

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

**Handlers are permanent.** Every handler must stay compiled in for as long as
the chain can be synced from genesis, because any syncing node re-executes it.
Cosmos chains accumulate `upgrades/v2/`, `v3/`, … forever and use cosmovisor to
swap binaries per height. gno has no supervisor (`contribs/` is gnodev, gnokms,
gnogenesis, gpao, …) and **no state sync** — there is no statesync reactor in
`tm2/pkg/bft`. So sync-from-genesis is the only way to join, and the registry is
append-only. This is the single largest cost of the decision.

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

Twice, and neither answer is conditional encoding inside one binary:

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

**No state-format version, rely on the plan record.** Rejected: with sequential
upgrades "which ran" is equivalent to "what format", but equivalence is not a
check. Nothing detects a binary expecting format N meeting data at N-1, and
`skip_upgrade_height` exists specifically to let an operator bypass the only
gate that exists.

**Two state-format versions, `main` and `base` separately.** Rejected, for the
same reason as per-module consensus versions: there is no independent axis. Both
stores are written by one binary, migrated by one handler that commits atomically
inside a block, and released together — so they cannot drift apart, and no
consumer exists to ask about one without the other. Nor does a second counter
save work: a handler already knows which store it touched.

**A separate `node:p:upgrade_name` param.** Rejected. It would break
`NewSetHaltRequest`'s public realm signature (three tests pin the arity at
`params_test.gno:21,28,35`), widen a public contract that already includes two
documented query paths and the `set_halt` event payload, and add a second source
of truth for the question `halt_min_version` already answers. Because the
registry is append-only, handler presence and the version floor are the same
predicate — so the second param buys nothing.

**Cosmos's name-only plan.** `x/upgrade`'s `Plan` is `{Name, Height, Info}` —
there has never been a version field. That is not a rejection of the version
approach so much as its unavailability: `x/upgrade` ships to hundreds of chains
with their own binaries, tag schemes and version strings, so the SDK has no
equivalent of `tm2/pkg/version.Version` to read, and no guarantee any chain's
string is even semver. A name is opaque, so it works everywhere. Cosmos then
gets the ordering free from its append-only registry, and a stable name survives
patch releases cut between the vote and the halt height.

The price Cosmos pays is that a wrong binary surfaces as a consensus panic at
the upgrade height rather than a refused startup — a large part of why cosmovisor
exists. gno has exactly the thing the SDK lacks: one binary, one release process,
one version string it controls and can parse. It should use it.

### The release version as the format version

Undecided — recorded to be settled before implementation. Drop
`StateFormatAfter` and the derived integer entirely; store the `Version` of the
last applied upgrade, and take the expected value as the newest `Version` in the
registry.

Orthogonal to the rest of the design: the check tables and the post-halt scoping
are identical either way.

#### What it removes

1. **The `StateFormatAfter` field, entirely.** Every upgrade implicitly advances
   the format to its own version. Nothing to choose per entry.
2. **The derivation step.** The expected value is just the newest `Version` in
   the registry.
3. **A whole numbering scheme.** Nobody has to answer "does this upgrade bump the
   format, or leave it where it was?"

#### What it improves

4. **The tripwire stops being forgettable.** The strongest argument.
   `StateFormatAfter` is hand-chosen per entry — write an upgrade that changes the
   VM object encoding, leave `StateFormatAfter` at its predecessor's value, and the
   check silently never fires. Deriving the constant from the registry fixes drift
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
13. **Who writes it for a `nil` handler.** The BeginBlocker already does this in
    the current design, so it carries over — worth confirming it stays
    unconditional.

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
4. The registry, the registry startup gate, and the BeginBlocker. No param
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
