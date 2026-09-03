# gpao

**gpao** (gno package-approver oracle) is a small off-chain approval daemon for
gno.land chains running with the `inert` code-submission policy (see
[PR #5888](https://github.com/gnolang/gno/pull/5888)).

Under the `inert` policy, anyone may submit a package with `MsgAddPackage`, but
it is stored **inert** — not typechecked, not executed, not importable. A
package only becomes active once an address in the chain's `PkgApprovers` param
sends `MsgEnablePackage`.

`gpao` automates that approver role:

1. **Watches** new blocks over RPC.
2. **Extracts** `MsgAddPackage` transactions from each block.
3. **Verifies** the submitted package off-chain — typecheck *and* preprocess,
   the same two stages the chain re-runs at `MsgEnablePackage` — under one
   wall-clock budget. Imports resolve from the local disk store (stdlibs +
   `examples/`) first, falling back to `vm/qfile` RPC queries against the
   watched node for on-chain-only packages.
4. If it passes **and finishes in time**, **broadcasts** a `MsgEnablePackage`
   signed by the approver key, activating the package on-chain.

> The oracle proposes, the chain enforces. gpao is untrusted for correctness:
> the validator re-runs `TypeCheckMemPackage` at `MsgEnablePackage` time and
> rejects ill-typed code. gpao only decides *which* pending packages get
> proposed for activation, and *when* — keeping the typechecker off the critical
> block-execution path.
>
> Which is why the time budget, not the correctness check, is the part that
> earns its keep: the chain will re-check correctness regardless, but it cannot
> bound how long that takes.

## Install

From this directory:

```sh
make install   # go install . — puts gpao on your $PATH
make build     # go build -o build/gpao . — leaves it here instead
```

## Usage

The approver key lives in a local [gnokey](../../gno.land/cmd/gnokey) keystore.
Create one, fund it, add its address to the chain's `PkgApprovers` param, then:

```sh
gpao \
  --remote http://127.0.0.1:26657 \
  --chain-id dev \
  --home ~/.gnokey \
  --key approver
```

gpao unlocks the key at startup: it reads the password from `$GPAO_PASSWORD` if
set (for unattended/service deployments), otherwise prompts once interactively.

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--remote` | `http://127.0.0.1:26657` | RPC address of the node to watch |
| `--chain-id` | *(required)* | Chain ID used to sign approval transactions |
| `--home` | gnokey home (`$GNOHOME`) | Keystore directory holding the approver key |
| `--key` | *(required)* | Name or bech32 address of the approver key |
| `--gno-root` | auto-detected | gno repo root, used to resolve stdlibs and examples for typechecking |
| `--gas-fee` | `1000000ugnot` | Gas fee for approval transactions |
| `--max-spend` | `100000000ugnot` | Total fees this run will pay for approvals before it stops approving |
| `--gas-wanted` | `20000000` | Fallback gas wanted, used only when the node will not simulate an approval |
| `--poll-interval` | `1s` | How often to poll for new blocks |
| `--start-height` | `0` | Height to start watching from (0 = current tip) |
| `--verify-budget` | `10s` | Withhold approval from a package that takes longer than this to verify |
| `--status-listen` | *(off)* | Address to serve the read-only status API on, e.g. `127.0.0.1:8546` |

### About `--status-listen`

gpao decides things nobody else can see. That a package failed to typecheck,
or that its enable was simulated and the chain would reject it, is the oracle's
knowledge alone -- the chain can say a package is parked, and why no enable
could succeed right now, but not why *this* one was refused. Without somewhere
to put it the only record is this process's stderr, so a submitter pays the
submission charge and then hears nothing.

With `--status-listen` set, two read-only JSON endpoints answer for it:

```sh
curl http://127.0.0.1:8546/status                        # every verdict
curl http://127.0.0.1:8546/status/gno.land/r/you/yours   # one package
```

```json
{
  "path": "gno.land/r/you/yours",
  "status": "rejected",
  "reason": "typecheck failed: undefined: Foo"
}
```

`status` is one of `rejected` (the code did not pass), `pending` (will be
retried), `gave_up` (retried to the cap, needs a human), `blocked` (nothing
wrong with the package -- the oracle has hit `--max-spend`), `approved`, or
`unknown` (never seen). `blocked` is the one worth separating: it means to go
and ask the operator, not to go and fix your code.

Off by default, and unauthenticated when on. Everything it reports concerns a
package submitted in a public transaction, so there is nothing here its
submitter could not already see -- but it does tell the world what this oracle
is doing, so bind it where you want that read.

### About `--verify-budget`

This is what the oracle is for. The chain re-runs the same type check *and*
preprocess when the package is enabled, so an oracle that only checked
correctness would repeat a check the validator already performs. What only an
off-chain actor can judge is whether verification *finishes quickly*, because
wall-clock time is not a consensus quantity.

Both stages are measured, since a package that type-checks quickly but
preprocesses slowly costs the chain just as much. Verification runs in a child
process (`gpao verify-one`, invoked by gpao on itself), which is what lets the
budget be *enforced* — a goroutine cannot be killed, but a process can. It also
means a package that crashes the typechecker takes down only its own child, and
that the approver key is never loaded in the process handling untrusted code.

The budget starts once the child has everything the compile needs: the standard
library and examples from disk, and every chain package the candidate imports,
fetched from the node. Fetching is the oracle's cost, not the package's, so a
slow node cannot turn a fast package into an overrun. That phase has a ceiling
of its own, one minute, and each request to the node is given the budget as its
timeout; either expiring leaves the package pending as unavailable rather than
counting against it.

The child's type-check options mirror `MsgEnablePackage`'s exactly (production
files only, no test-file evaluation), because the whole point is to predict what
the validator will do — any divergence is a way to approve something the chain
then rejects, or to reject something it would have accepted. Preprocess is skipped, with a
log line, for a package whose imports this oracle cannot resolve — approving on
the type check alone is what it did before, and refusing would penalise the
package for a limitation of the oracle.

The default is deliberately generous: a real package verifies in milliseconds,
and a borderline one should pass rather than lose a race with whatever else the
machine is doing.

Exceeding the budget is **not** a rejection. The package is left pending and
neither approved nor recorded as bad. Nothing re-offers it automatically — block
heights are read once and only move forward — so retrying it means restarting
with `--start-height` at or below the block that submitted it.

A failed **enable** works the same way. A package that verified but whose
approval failed on chain is left pending through a bounded number of attempts,
because the usual causes are not the package: an unfunded storage deposit, a
dependency not live yet, a namespace or governance param that moved, a block out
of gas. Those clear on their own. After the last attempt the path is recorded
and the log says a human is needed.

The key's address **must** be listed in the chain's vm `PkgApprovers` param, and
`code_submission_policy` must be `inert`, otherwise the `MsgEnablePackage`
transactions are rejected.

### Signing options

- **Local gnokey keystore (default, recommended)** — the encrypted key stays on
  disk and is unlocked at startup. This is the same keystore `gnokey` uses.
- **Mnemonic (dev only)** — set `$GPAO_MNEMONIC` to sign from a raw mnemonic
  without a keystore. Convenient for local devnets; not for production.
- **tmkms / gnokms are NOT supported** — those are *consensus* key managers that
  sign block votes over the privval protocol. gpao signs application
  *transactions* (`MsgEnablePackage`), which they cannot do. Use the gnokey
  keystore (or, in future, an HSM/KMS-backed keystore that can sign txs).

## Where imports come from

Standard library packages are read from local disk. They ship with the binary
and are not chain state, so disk is the only place to get them.

Everything else — `/p/` and `/r/` packages — is read from the chain, and disk is
not consulted for them at all. This is the point of the daemon: the verdict has
to describe what the validator will see when it runs the enable, and the
validator resolves imports from chain state. A package importing something that
exists in the operator's `examples/` but not on the chain must not verify clean;
if it did, the approval would fail its own type-check on chain, burning a fee and
blaming the code for the operator's local tree.

With no `--remote` there is nothing to ask, so disk answers everything. That is
a development mode, and the verdict then describes the operator's tree rather
than the chain.

## Import cache

Packages fetched via `vm/qfile` are cached for the process lifetime. This is
safe: on-chain package paths are write-once (re-adding an existing path fails),
so a fetched package never changes. Only successful fetches are cached — a miss
(a package still inert, or enabled later in the run) is re-queried on the next
lookup rather than pinned to "not found".

### About `--gas-wanted`

Each approval is sized by simulating it first: the measured gas plus 20%,
bounded by the chain's `Block.MaxGas`. The flag is only the fallback for when
the node will not answer a simulation.

Sizing it by hand is hard to get right, which is why it is measured. The worst
case is nearer 40,000,000 than the default — a 1 MB parked package, plus the
namespace and CLA realm calls, all of which run on the approver's meter — and
that ceiling depends on what the submitter sent, so no fixed number is safe for
every package.

Two details, in case the numbers look odd in the logs. The probe transaction is
signed at the chain's block ceiling rather than at the fallback, because a
simulation executes under the transaction's own limit — sizing the probe at the
fallback would run out of gas on exactly the packages worth measuring. And the
ceiling is read from the chain at startup rather than assumed, because the ante
REFUSES a gas-wanted above `Block.MaxGas` instead of clamping it, so a chain
configured below the tm2 default would reject every probe.

A failed simulation does not withhold approval. It logs, falls back, and sends.
Refusing to approve whenever the query path is unavailable would let anyone who
can disturb it stall approvals for the whole chain.

### About `--max-spend`

Every approval costs the full gas fee, whether or not the message succeeds. The
daemon decides on its own when to send one, so anything that makes approvals
fail repeatedly will drain the approver key. The bound stops that.

Two things reduce how often it is reached. Before approving, the daemon checks
whether the package is already deployed and skips it if so, which is the common
case when catching up with `--start-height` over blocks that were already
approved. And it ignores transactions that failed on chain, so a submission the
chain rejected never leads to an approval.

When the bound is reached the daemon says so and stops approving. It keeps
watching blocks. Raise the bound or restart to continue.

## Limitations

- **No catch-up persistence**: `--start-height` lets you replay from a given
  height, but gpao keeps no on-disk cursor between runs.
