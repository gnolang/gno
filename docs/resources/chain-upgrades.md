# Chain upgrades

This page describes how a gno.land chain moves from one binary — and one state
— to the next: what the chain does on its own, what operators have to do by
hand, and which parts of the design exist today versus which are still only
written down.

## Three kinds of upgrade

| Kind | State-breaking | Coordination needed | How it happens |
|---|---|---|---|
| **Rolling** | No | None | Operators switch binaries whenever convenient. The chain never stops. |
| **Coordinated** | Yes | Halt height | Every node stops at the same block, then restarts on the new binary — replaying history into a fresh genesis if the state format changed. |
| **Hotfix** | Yes | Fully manual | No governance vote. Last resort. |

Only the coordinated path has chain-level machinery behind it. The rest of this
page is about that path.

## Halting the chain

A coordinated upgrade starts by stopping every node at the *same* height, so
each one ends up with byte-identical state for the new binary to start from.

### Setting the halt height

The normal path is governance. The config field is the fallback for the cases
governance cannot reach. Use one **or** the other — never both on the same node;
see [Do not mix the two](#do-not-mix-the-two) below.

#### Chain-wide, via governance

A GovDAO proposal writes the height into chain state, so every node picks it up
on its own and no operator has to touch a config file:

```gno
package main

import (
    "gno.land/r/gov/dao"
    "gno.land/r/sys/params"
)

func main(cur realm) {
    // Halt at block 704052; require binary >= chain/gnoland1.1 to resume.
    preq := params.NewSetHaltRequest(cross(cur), 704052, "chain/gnoland1.1")
    dao.MustCreateProposal(cross(cur), preq)
}
```

This sets two params, readable by anyone:

| Param | Type | Meaning |
|---|---|---|
| `node:p:halt_height` | `int64` | Height to halt at. `0` cancels a scheduled halt. |
| `node:p:halt_min_version` | `string` | Minimum binary version required to restart afterwards. `""` means no version gate. |

```bash
gnokey query params/node:p:halt_height
gnokey query params/node:p:halt_min_version
```

The height must be strictly greater than the height at which the proposal
executes, but nothing enforces a *minimum* notice period — a proposal executing
at height H can legally set H+1.

Nothing has to be cleaned up afterwards. The params stay in state and the node
still resumes, because the arming check fires only on the exact halt height and
that height is behind it by then. (`halt_min_version` does stay in force as a
permanent version floor — see [Restart gates](#restart-gates).)

#### Per node, via config — when governance is not an option

Some halts cannot go through a vote: a local fork or devnet with no GovDAO, a
chain already halted or otherwise unable to execute a proposal, replay tooling,
or an emergency stop on a single node. For those, `halt_height` in `config.toml`
does the same job without chain state:

```bash
gnoland config set halt_height 704052
```

Each operator sets it themselves, which is also its weakness: coordination is
out-of-band, so one operator who misses the message keeps producing blocks.
That is why it is the fallback rather than the default for a network-wide halt.

`gnoland config set` only edits the file. The node reads `halt_height` once at
startup, so a running node has to be restarted before a new value takes effect.

**Clearing it afterwards is required, not just hygiene.** Unlike the governance
params, this value is re-applied at every startup — so once the node has halted
at `halt_height` it cannot resume: the next block is already past the limit, and
`BeginBlock` panics again on the first block it tries to process. Reset it
before restarting:

```bash
gnoland config set halt_height 0
```

### Do not mix the two

The governance param and the config field write the *same* field on the node —
there is one halt height in memory, and whoever writes it last wins. The config
value is applied at every startup; the governance value is applied by the block
that reaches it. A node carrying both halts at the wrong block, in one of two
directions:

- **Config below the governance height.** The node halts at its own value and
  never reaches the agreed one. It stops out of step with the network — exactly
  the divergence the governance halt exists to prevent.
- **Config above the governance height.** Governance wins at the time, but the
  restart re-applies the config value, so the node halts a second time at a
  height nobody voted for.

Both of those failures need a config value that is still in the *future*. A
value the chain has already passed simply blocks the restart, as described
above — loud, and found immediately. What survives unnoticed in a `config.toml`
is a future-dated height, which sits there silently until the halt it causes.

In practice: before a governance halt, every operator should confirm their own
config is clear.

```bash
gnoland config get halt_height   # expect 0
```

### What a halt actually looks like

The block **at** `halt_height` is fully committed. The halt is armed, and the
panic fires when the *next* block begins, so the last committed block is
`halt_height` itself.

The stop itself is identical whether the halt was scheduled by governance or set
in `config.toml`: each writes the same in-memory halt height, so each reaches the
same check in `BeginBlock`, where the panic is caught by the consensus routine:

```
CONSENSUS FAILURE!!!  err="halt height 704052 reached, node shutting down"  stack=...
```

What differs is the warning you get beforehand.

**A governance halt announces itself at the halt block.** The EndBlocker reads
`node:p:halt_height` on every block and logs when it matches, one block before
the stop:

```
GovDAO halt height reached, will halt after this block  height=704052  halt_height=704052
```

**A config halt announces itself only at startup.** The value is read once when
the node boots, which is where its single log line appears — possibly days
before the halt:

```
Halt height configured  height=704052
```

After that there is nothing until the crash. Nothing runs at the halt block,
because the config path never consults chain state, so no "will halt after this
block" line is ever emitted. If you are watching for a config-driven halt, watch
the height, not the log.

**The process does not exit.** `gnoland start` waits on its signal context, not
on consensus health, so the node stops producing blocks while the RPC server
stays up and the process idles. Stop it yourself. Process supervisors configured
to restart on exit will not fire.

That is a deliberate trade, not an oversight. The governance halt originally
called `osm.Kill()`, which sends the node `SIGTERM` and does end the process.
Review replaced it with `BaseApp.SetHaltHeight()` so that both ways of setting a
halt height share one deterministic path with no async signals: every node stops
at the same block because the same check runs during block execution, not because
a signal arrived at some moment. A process that outlives its own halt is the cost
of that guarantee.

### Restart gates

If `halt_min_version` is set, the node checks the running binary against it at
startup — before any blocks are processed — and refuses to start in two
directions:

- **After the halt** (`lastBlockHeight >= halt_height`): a binary *below* the
  minimum version is refused. This is what keeps a retired binary from resuming
  the chain and diverging from everyone else.
- **Before the halt** (`lastBlockHeight < halt_height`): a binary that *already
  meets* the minimum version is refused. The upgrade binary is for after the
  cut; running it early is the same divergence from the other side.

Both checks are skipped entirely when `halt_height` is `0` or
`halt_min_version` is empty. A halt proposal that leaves the version empty is a
coordinated *pause*, not an upgrade gate.

An operator who has already migrated out-of-band can bypass both:

```bash
gnoland config set skip_upgrade_height 704052   # must equal halt_height
```

#### Where the binary's version comes from

The value being compared is the `tm2/pkg/version.Version` variable compiled into
the binary. It is a build-time constant, not something the node reads from
config or chain state, so it is fixed the moment the binary is produced:

| How it was built | Reported version |
|---|---|
| `go build ./gno.land/cmd/gnoland` | `develop` — the hardcoded default, since nothing overrides it |
| `make build.gnoland` on a release tag | the tag, e.g. `chain/gnoland1.1` |
| `make build.gnoland` off a tag | `<branch>.<commits>+<hash>`, e.g. `master.3335+bc43a5fb7` |
| `go build -ldflags "-X github.com/gnolang/gno/tm2/pkg/version.Version=..."` | whatever you pass |

The Makefile injects it with `-ldflags -X`, deriving the value from
`git describe --tags --exact-match` and falling back to the branch/count/hash
form when the commit is not tagged. `chain/gnoland1.0` and `chain/gnoland1.1`
are real tags in this repository — that is where the
`chain/gnoland<major>.<minor>` format comes from, and why a binary built on a
release tag reports exactly the shape the comparison understands.

Versions are compared by parsing that shape: different majors compare as majors,
otherwise the minor must be greater or equal. **Anything that does not parse
falls back to exact string equality**, so `develop` and `master.3335+bc43a5fb7`
satisfy no `chain/gnolandX.Y` floor at all. A release build passes the gate; an
ad-hoc build of the same code does not.

To read a version back:

```bash
gnoland version                              # gnoland version: chain/gnoland1.1
curl -s localhost:26657/status | grep -i build_version
```

The RPC `/status` endpoint reports it as `BuildVersion`, so you can survey what
the rest of the network is running — worth doing before choosing a
`halt_min_version` that some validators cannot satisfy.

**The halt params are never cleared.** Nothing in the node resets them after a
successful resume, so `halt_min_version` stays in force as a permanent
minimum-version floor for every later restart. That is usually what you want.
To lift it, pass a new proposal with `NewSetHaltRequest(cross(cur), 0, "")`.

## Changing state: genesis replay

Halting and restarting is enough when the new binary can read the old state.
When it cannot, the state has to be rebuilt — and gno.land does that by
**replaying history into a new genesis**, not by transforming the old database
in place.

### How it works

`gnogenesis fork generate` reads the halted chain (an RPC endpoint, a halted
data directory, or a pre-exported archive) and writes a genesis file containing
the full transaction history plus the metadata needed to replay it:

```bash
gnogenesis fork generate \
    --source-txs-data-dir /path/to/halted/gnoland/data \
    --source-genesis-file /path/to/source/genesis.json \
    --chain-id gnoland-1 \
    --output genesis.json
```

On the new chain, `InitChain` replays every historical transaction through the
full ante handler. Three genesis-level fields make that possible:

| Field | Purpose |
|---|---|
| `PastChainIDs` | Chain IDs whose signatures stay valid during replay. Historical txs were signed against the source chain and cannot be re-signed. |
| `InitialHeight` | The new chain's first block. Set to `halt_height + 1`, so block numbering continues instead of restarting at 1. |
| `GasReplayMode` | `strict` (default) meters historical txs with the new VM's gas rules; `source` bypasses metering so txs that changed cost still reproduce their original outcome. |

At the end of replay the node emits a report counting each transaction as `ok`,
`ok_gas_differs`, `failed`, or `skipped_failed`, with a warn line per failure.

A new chain ID is **not** required. `PastChainIDs` may contain the current chain
ID, which is the right shape for a fork that changes state without changing the
chain's external identity.

### Expressing the migration

Since replay re-executes the original history, anything the upgrade *adds* is
expressed as extra input to genesis rather than as code in the node:

| Flag | What it does |
|---|---|
| `--migration-tx FILE` | Appends transactions to the end of `appState.Txs`, after all historical replay. Repeatable. This is where new deploys, param changes, and data seeding go. |
| `--patch-realm PKGPATH=SRCDIR` | Rewrites a genesis-mode `addpkg` transaction in place with files from `SRCDIR`. The only way to land a realm code change as part of a fork, since you cannot re-`addpkg` a path that already exists. |
| `--patch-txs FILE` | Patches specific historical transactions, matched by block height and signer. |

Two helpers build `--migration-tx` inputs: `gnogenesis fork valoper-seed` (from
a CSV of validators) and `gnogenesis fork addpkg` (from local package
directories). `gnogenesis fork test` boots the result in-process as a smoke
test before you ship it, and `gnogenesis fork inspect` prints a provenance
report over a finished genesis — how many transactions came from history, from
migration files, and from patches, and why.

### Known limits

- Fork genesis files are large — roughly 192 MB for a chain halted at height
  ~704k, since they carry the whole transaction history.
- The RPC source has no retry or resume; a single transient error aborts the
  fetch. Prefer `--source-txs-data-dir` against a halted data directory.
- All transactions are held in memory during generation, so very large chains
  can exhaust RAM.

### Worked examples

The deployment scripts are the real reference, and they are considerably more
detailed than this page:

- [`misc/deployments/test13.gno.land/gen-genesis.sh`](../../misc/deployments/test13.gno.land/gen-genesis.sh)
  — the fullest example, including `--migration-tx` for valoper seeding
- [`misc/deployments/pearl.gno.land/gen-genesis.sh`](../../misc/deployments/pearl.gno.land/gen-genesis.sh)

## What does not exist

Worth stating plainly, because the design documents describe more than the code
implements:

- **There is no migrate function or upgrade-handler framework.** A binary
  cannot register a migration keyed to `halt_min_version` and have the node run
  it on resume. `halt_min_version` gates startup and does nothing else. The
  idea was raised in review on
  [#5368](https://github.com/gnolang/gno/pull/5368) and deferred; the follow-ups
  that would have built it were closed unmerged, and
  [#5377](https://github.com/gnolang/gno/pull/5377) (`gnoland start --migrate`,
  which rebuilds app state by re-executing the block store in place) is still
  open. Migration happens on the new chain during replay, not in the binary
  that halted.
- **`skip_upgrade_height` is narrower than its name suggests.** It was added to
  skip a migrate function for operators who had already migrated out-of-band.
  With no migrate function, it only bypasses the two version checks.

## A coordinated upgrade, end to end

1. Agree the halt height, leaving room for the proposal to be voted and
   executed.
2. Pass a GovDAO `NewSetHaltRequest(cross(cur), H, minVersion)` proposal.
   Operators should confirm `halt_height` is unset in their own `config.toml`;
   a future-dated value there halts the node at the wrong block — early if it is
   below H, or a second time after the upgrade if above.
3. The chain commits block H and stops. Operators stop the idle processes.
4. **If the state format did not change**: restart on the new binary. It must
   satisfy `minVersion`. The chain continues at H+1.
5. **If it did change**: run `gnogenesis fork generate` against a halted data
   directory, add `--migration-tx` / `--patch-realm` for anything the upgrade
   introduces, verify with `gnogenesis fork test`, distribute the new
   `genesis.json`, and start the new binary against it. Replay rebuilds state
   and the chain produces its first fresh block at `InitialHeight`.

## Reference

| Topic | Where |
|---|---|
| Halt params, arming, startup checks | [`gno.land/adr/pr5368_govdao_halt_height.md`](../../gno.land/adr/pr5368_govdao_halt_height.md) |
| Genesis replay, tx metadata, fork tooling | [`gno.land/adr/pr5511_chain_upgrade_genesis_replay.md`](../../gno.land/adr/pr5511_chain_upgrade_genesis_replay.md) |
| `InitialHeight` in consensus and the block store | [`tm2/adr/pr5511_initial_height.md`](../../tm2/adr/pr5511_initial_height.md) |
| Manual test plan for the halt mechanism | [Testing the halt height feature](test-halt-height.md) |
| Node configuration and operation | [`gno.land/cmd/gnoland/README.md`](../../gno.land/cmd/gnoland/README.md) |
