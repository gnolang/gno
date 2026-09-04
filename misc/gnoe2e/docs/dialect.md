# The gnoe2e dialect

A scenario is a [txtar](https://pkg.go.dev/github.com/rogpeppe/go-internal/txtar) file run by
[testscript](https://pkg.go.dev/github.com/rogpeppe/go-internal/testscript): the comment block at the top is the
script, the file sections below it are unpacked into `$WORK`. One section, `-- cluster --`, is not a file for the
script but the declaration of the chain the harness boots for it. This page is the reference for that section, for
the verbs the harness adds to testscript's builtins, for the variables a script can name, and for the two accounts
every cluster carries. `testdata/tour.txtar` uses all of it, with a comment per line.

## The cluster section

Every script must carry a `-- cluster --` file section, and a script without one is an error rather than a default
cluster. A section is `key: value` lines; blank lines and `#` comments are skipped, and an unknown key is an error
naming the key, so a typo fails before anything boots.

| Key | Value | Effect |
| --- | --- | --- |
| `validators` | integer, at least 1 | Required. Number of validator processes. |
| `code-submission-policy` | `permissionless`, `permissioned` or `inert` | Empty leaves the chain default. |
| `pkg-approver` | empty, `user`, or a bech32 address | Who may send `MsgEnablePackage`. Empty means the oracle, `user` means the test account, which is what leaves the oracle unauthorized. |
| `block-max-gas` | integer | Per-block gas limit. Default 3000000000. |

An inert chain needs somebody who can enable, and `pkg-approver` left unset names the oracle: its key is
provisioned for every run, so an inert chain always boots with an approver rather than as a chain where nothing
submitted after genesis could ever go live. `pkg-approver: user` and `pkg-approver: <address>` name somebody else
instead, and naming the user is what leaves the oracle unauthorized.

Two generic families reach the settings that have no key of their own.

| Prefix | Vocabulary | Applied to |
| --- | --- | --- |
| `config.<path>` | the keys `gnoland config set` takes, which are the config.toml spellings | every validator's config.toml, after the harness's own configuration passes and before any node starts |
| `genesis.<path>` | the keys `gnogenesis params set` takes | the `auth`, `vm` and `bank` params in the genesis app state |

```
config.moniker: tour
genesis.vm.chain_domain: tour.gno.land
```

`make defaults` prints every key of both families with the value a cluster boots with, written in this section's
own syntax; `go run . defaults config` and `go run . defaults genesis` print one family. The values are the
harness's rather than the chain's, which is the point of reading them there: a local cluster commits on 10ms
consensus timeouts, so `config.consensus.timeout_commit` starts at `10ms` and not at tm2's default.

A setting that holds a list is stated as one comma-separated value with no spaces, the way the genesis and config
CLIs read theirs: `genesis.bank.restricted_denoms: ugnot,gnot`.

The two vocabularies spell things differently and neither is negotiable. `config.` paths are resolved by toml tag,
because the top-level section of the node config carries no json tags at all: `config.moniker` and
`config.consensus.timeout_commit` are the spellings, not their json equivalents. `genesis.` paths are resolved by
json tag, matching `gnogenesis params set`: `genesis.vm.chain_domain`, `genesis.auth.max_memo_bytes`,
`genesis.bank.restricted_denoms`.

Three `config.` keys are refused outright, with an error saying so:

```
config.rpc.laddr
config.p2p.laddr
config.p2p.persistent_peers
```

The harness picks a free port per node and hands the resulting addresses to the script as `RPC_ADDR_N`. A scenario
that set a listen address would take the cluster away from its own commands. It also writes each validator a peer
list naming the others, and a section is applied once to every node: a scenario that set the peer list would hand
all of them the same peers, and a validator set that cannot reach quorum commits no block at all.

A path that runs past a leaf is an error naming the key and the type it ran into, so
`config.consensus.timeout_commit.seconds` reports that `config.consensus.timeout_commit` is a
`time.Duration` rather than a section. A value that will not parse into the field's type is an error too, and for
`genesis.` paths so is a value that parses but leaves the module's params invalid.

Order within the section does not matter for the named keys, but the two generic families are applied last, so a
path that covers the same field as a named key wins over it: `genesis.vm.code_submission_policy` beats
`code-submission-policy`, and `genesis.vm.pkg_approvers` beats `pkg-approver`. Within a family the lines apply in
the order they are written.

Consensus parameters are not reachable through `genesis.`. They live in the genesis document rather than in the app
state the `genesis.` family writes to, and the app state is where the `auth`, `vm` and `bank` modules keep theirs.
`block-max-gas` is the one consensus parameter a scenario can set; the rest are `run` flags for the whole run.

## Verbs

Standard testscript builtins work as usual: `stdout`, `stderr`, `cmp`, `cp`, `exec`, `!` negation and the rest. On
top of them:

| Verb | Usage | What it does |
| --- | --- | --- |
| `gnokey` | `gnokey [root flags] <subcommand> [flags]` | Runs gnokey in-process against the cluster. |
| `sleep` | `sleep <duration>` | Waits. Any `time.ParseDuration` string. Not negatable. |
| `repeat` | `repeat [-all] N <cmd> [args...]` | Runs a custom verb N times. |
| `eventually` | `eventually [timeout [interval]] [-stdout regex] <cmd> [args...]` | Reruns a custom verb until it succeeds. Not negatable. |
| `http_get` | `http_get <url> [regex]` | Fetches a URL, writes the body to stdout. A `tcp://` address is read as `http://`, so `$RPC_ADDR/status` works. |
| `gpao` | `gpao start\|stop\|restart [flags...]` | Supervises the package-approver oracle. |
| `validator` | `validator stop\|restart <index>` | Stops and restarts one validator of the cluster. |

**`gnokey`** gets `-home`, `-remote`, `-insecure-password-stdin` injected, and the caller's arguments come last, so a
line naming its own `-remote` wins. `-remote` is a root flag and must precede the subcommand:
`gnokey -remote $RPC_ADDR_2 query ...` addresses validator 2, while a line naming none addresses the first
validator. `-chainid` is a `maketx` flag rather than a root one and is not injected, so a transaction needs
`-chainid $CHAIN_ID` written out.

**`repeat`** stops at the first failing iteration and reports which one it was. With `-all` it runs the full count
and prints a pass/fail summary. Iterations share testscript's output buffers, so a following `! stdout <pattern>`
checks the pattern against every iteration's output at once, which is stricter than one read. `! repeat` asserts
that at least one iteration failed.

**`eventually`** defaults to a 30s budget and a 1s interval; a leading duration sets the budget and a second sets
the interval. The deadline is checked between attempts, so a sub-command that hangs can overrun it. Each attempt
starts with the output buffers emptied, so an assertion after the wait sees only the attempt that succeeded, which
is what makes a negation or a `cp stdout` safe to retry. Without a gate the wait ends as soon as the sub-command
exits 0, and a following `stdout <pattern>` then runs once against whatever that first success produced. For a
query that answers an empty result without erring, `vm/qinertpaths` being the usual one, that is unsound and
`-stdout <regex>` is the answer: the pattern is checked inside the attempt, so an exit 0 whose output does not
match is not yet an answer and the wait runs the command again.

`repeat` and `eventually` dispatch only to the verbs in this table. A testscript builtin cannot be wrapped in
either, because the builtin table is not reachable from a custom command.

**`gpao`** owns the oracle's lifetime, so a script that fails halfway leaves no daemon behind. `start` picks a free
port for the status board, exports `GPAO_STATUS`, and returns once that board answers, up to 30s. The mnemonic,
`-chain-id`, `-status-listen` and `-gno-root` are supplied by the harness; everything on the line is passed through
to the binary, and `-remote` there replaces the run's default node. `! gpao start` is allowed, for a scenario
asserting the oracle refuses to come up. `stop` and `restart` are not negatable, and `restart` is a stop followed by
a start that takes the flags on its own line. Whatever the oracle wrote is logged when it stops, so a failed
assertion still comes with the oracle's own account of events.

**`validator`** indexes the same way the scripts already do, so `validator stop 3` stops the node behind
`RPC_ADDR_3`. A stopped node keeps its data directory and its identity, and `restart` brings the same validator
back to the same chain and returns once its RPC answers, which is earlier than it having caught up. A restarted
node appends to the log of the run that stopped rather than replacing it, so `validator_N/stderr.log` holds both.

Only `restart` is negatable, and only for the node itself: `! validator restart N` asserts the node cannot come
back, and the error carries a tail of the node's stderr so the scenario can name the reason it died. Naming a
validator the cluster does not have, or one the script never stopped, fails the scenario in either mode -- a
negation standing on either would pass a scenario in which no node ever went away.

`validator stop N` fails the scenario when the node had already exited on its own, rather than reporting a stop it
did not perform.

## Environment

| Variable | Value | Set |
| --- | --- | --- |
| `RPC_ADDR` | the first validator's RPC address, `tcp://127.0.0.1:<port>` | always |
| `RPC_ADDR_0` through `RPC_ADDR_<N-1>` | one per validator, indexed the way `validator stop` and `validator restart` index them | always |
| `CHAIN_ID` | the chain id, `test-e2e` unless `-chain-id` says otherwise | always |
| `USER_ADDR` | the test account's address | always |
| `USER_NAME` | the test account's keybase name, `test1` | always |
| `GNOHOME` | the run's keybase directory, which `gnokey` reads as its `-home` | always |
| `GNOROOT` | the checkout the binaries were built from | always |
| `GPAO_KEY_NAME` | the keybase name of the oracle's key, `gpao` | always |
| `GPAO_ADDR` | the oracle's address | always |
| `GPAO_STATUS` | base URL of the running oracle's status board | after `gpao start` |
| `WORK` | testscript's own working directory, where the script's file sections are unpacked | always |

## Identities

Two accounts, imported once into a keybase every cluster in the run shares, and funded with 1000 GNOT on each of
those clusters.

| Account | Keybase name | Key | Role |
| --- | --- | --- | --- |
| test user | `test1` | derived at account 0 index 0 from the test1 seed `gno.land/pkg/integration` uses, the `-mnemonic` default | signs everything a script submits |
| oracle | `gpao` | derived at account 0 index 0 from the BIP39 test vector (`abandon` eleven times, then `about`), the `-gpao-mnemonic` default | signs the oracle's approvals |

The oracle's key exists for every run, whether or not a script starts the oracle: what shapes a chain is its
submission policy and its approver set, and running `gpao` is a decision the script makes. The binary is the part
that is deferred, built on the first `gpao start` of a run and shared by every later one. The key is derived once
and used twice, for the genesis approver entry and for the keybase import, so the chain cannot end up with an
approver nobody can sign for. Being in the keybase is what lets a script do by hand what the oracle does, with
`gnokey maketx enablepkg ... $GPAO_KEY_NAME`.

The oracle needs a separate mnemonic rather than another index in the test user's, because it derives its signer at
account 0 index 0 with no way to change that.

An oracle that is not an approver is the one misconfiguration that leaves it looking healthy. It starts, follows
blocks, verifies packages and reports on them, and every package it approves stays inert. The run logs a warning
per scenario whose chain is inert and whose approver set does not hold the oracle's key, because nothing else about
the run will say so: the only symptom is deploys that never activate. A chain that is not inert parks nothing, so
its approver set says nothing about the oracle and no warning is due.
