# gnoe2e

gnoe2e runs txtar scenarios against real gnoland clusters. Each scenario declares the chain it needs in a
`-- cluster --` section, and the harness boots that chain from genesis, runs the script against it, and throws it
away: one cluster per scenario, one to four gnoland processes running a binary built from the enclosing checkout,
with an off-chain daemon the script can start, stop and restart alongside them. That is what the lane is for. A
scenario can stop a validator mid-run, watch the chain carry on without it, and bring it back; it can watch the
package-approver oracle activate a parked package and then take the oracle away.

Two other lanes in this repository test gnoland, and they answer different questions.

| Lane | Chain | Script dialect | Packages |
| --- | --- | --- | --- |
| `misc/gnoe2e` | one to N gnoland processes per scenario, from genesis | txtar plus `gnokey`, `gpao`, `validator`, `eventually`, `repeat`, `http_get`, `sleep` | deployed by the script with `gnokey maketx addpkg` |
| `gno.land/pkg/integration` | one in-process node the script starts itself | txtar plus `gnoland start\|stop\|restart`, `gnokey`, `loadpkg`, `adduser`, `patchpkg` | loaded into genesis with `loadpkg`, before the node starts |
| `misc/e2e` | one gnoland container from docker compose, started `-lazy` | a `gnokey` shell script, `run_tests.sh` | none |

`gno.land/pkg/integration` boots in-process, allocates no ports and puts packages in genesis, so it is the cheap
lane. What it cannot express is a second node, a node dying, or a process outside the chain reacting to what the
chain committed. Those are what gnoe2e is here for.

`testdata/tour/every_verb.txtar` is the worked example. It uses every verb, every `-- cluster --` key and all but
two of the exported variables, with a comment on each line saying why that line is shaped the way it is.

## Running

A run needs the `go` toolchain: `gnoland`, and `gpao` when a scenario asks for the oracle, are built on demand from
the enclosing checkout. That checkout is found through `go list`, so `GNOROOT` must not be set in the environment
unless it deliberately points at another tree.

```bash
# One directory of scenarios through the CLI
cd misc/gnoe2e && go run . run testdata/oracle

# Named files and directories mix, and the argument order is the run order
go run . run testdata/oracle/first_light.txtar testdata/tour

# No argument runs testdata/integration, the suite that ships with the checkout
go run . run

# Every scenario directory, coloured, verbose
make scenarios

# The whole suite through go test
go test -timeout 30m ./...

# Unit tests only: TestScenarios skips under -short
go test -short ./...

# One scenario, named by its path under testdata/ without the extension
go test -timeout 30m -run 'TestScenarios/oracle/first_light' .

# The suite against gnoland and gpao built from master, so a scenario written
# for a fix can be shown red before the fix lands
make test-master
GOTEST_FLAGS='-timeout 30m -run TestScenarios/oracle/first_light' make test-master

# A cluster to poke at by hand: prints its RPC address and blocks until Ctrl-C
go run . serve -validators 3 -verbose

# The repository's linter, on this module
make lint
```

`make build` puts the binary in `build/gnoe2e` and `make install` installs it, for anyone who would rather not go
through `go run` each time.

`make test-master` unpacks master with `git archive` into a temp directory and points `GNOROOT` at it, which is the
one case where setting `GNOROOT` is the point. `GOTEST_FLAGS` defaults to `-v -timeout 30m`, and giving it replaces
both.

Five `run` flags override what a scenario declared:

| Flag | Overrides |
| --- | --- |
| `-validators N` | `validators` |
| `-oracle` | `oracle` |
| `-code-submission-policy <policy>` | `code-submission-policy` |
| `-pkg-approver <address>` | `pkg-approver` |
| `-block-max-gas N` | `block-max-gas` |

A flag counts as an override only when it is actually given, so leaving `-validators` alone is not the same as
passing `-validators 2`. A given flag beats every scenario's declaration, including scenarios that cannot pass
without theirs: running a four-validator outage scenario with `-validators 2` turns it red, and that is what an
override is for. The remaining flags (`-chain-id`, `-max-tx-bytes`, `-max-data-bytes`, `-block-time-iota`,
`-load-examples`) set the base every cluster in the run starts from; no scenario declares them.

## The cluster section

Every script must carry a `-- cluster --` file section, and a script without one is an error rather than a default
cluster. A section is `key: value` lines; blank lines and `#` comments are skipped, and an unknown key is an error
naming the key, so a typo fails before anything boots.

| Key | Value | Effect |
| --- | --- | --- |
| `validators` | integer, at least 1 | Required. Number of validator processes. |
| `oracle` | `true` or `false` | Builds `gpao`, derives its key and funds it. Default `false`. |
| `code-submission-policy` | `permissionless`, `permissioned` or `inert` | Empty leaves the chain default. |
| `pkg-approver` | empty, `user`, or a bech32 address | Who may send `MsgEnablePackage`. Empty means the oracle when one is declared, `user` means the test account, which is what leaves the oracle unauthorized. |
| `block-max-gas` | integer | Per-block gas limit. Default 3000000000. |

An inert chain needs somebody who can enable, so `code-submission-policy: inert` is refused when `pkg-approver` is
left unset and no oracle is declared: nothing submitted after genesis on such a chain could ever go live. Any one
of `oracle: true`, `pkg-approver: user` and `pkg-approver: <address>` satisfies it, though the refusal message
names only the first two.

Two generic families reach the settings that have no key of their own.

| Prefix | Vocabulary | Applied to |
| --- | --- | --- |
| `config.<path>` | the keys `gnoland config set` takes, which are the config.toml spellings | every validator's config.toml, after the harness's own configuration passes and before any node starts |
| `genesis.<path>` | the keys `gnogenesis params set` takes | the `auth`, `vm` and `bank` params in the genesis app state |

```
config.mempool.max_pending_txs_bytes: 4096
genesis.vm.chain_domain: tour.gno.land
```

The two vocabularies spell things differently and neither is negotiable. `config.` paths are resolved by toml tag,
because the top-level section of the node config carries no json tags at all: `config.moniker` and
`config.consensus.timeout_commit` are the spellings, not their json equivalents. `genesis.` paths are resolved by
json tag, matching `gnogenesis params set`: `genesis.vm.chain_domain`, `genesis.auth.max_memo_bytes`,
`genesis.bank.restricted_denoms`.

Two `config.` keys are refused outright, with an error saying so:

```
config.rpc.laddr
config.p2p.laddr
```

The harness picks a free port per node and hands the resulting addresses to the script as `RPC_ADDR_N`. A scenario
that set a listen address would take the cluster away from its own commands.

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
| `http_get` | `http_get <url> [regex]` | Fetches a URL, writes the body to stdout. |
| `gpao` | `gpao start\|stop\|restart [flags...]` | Supervises the package-approver oracle. |
| `validator` | `validator stop\|restart <index>` | Stops and restarts one validator of the cluster. |

**`gnokey`** gets `-home`, `-remote`, `-insecure-password-stdin` injected, and the caller's arguments come last, so a
line naming its own `-remote` wins. `-remote` is a root flag and must precede the subcommand:
`gnokey -remote $RPC_ADDR_2 query ...` addresses validator 2, while a line naming none addresses the first
validator. `-chainid` is a `maketx` flag rather than a root one and is not injected, so a transaction needs
`-chainid $CHAIN_ID` written out.

**`repeat`** stops at the first failing iteration and reports which one it was. With `-all` it runs the full count
and prints a pass/fail summary. Iterations share testscript's output buffers, so a following `! stdout <pattern>`
checks the pattern against every iteration's output at once, which is stricter than one read.

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
back to the same chain and returns once its RPC answers, which is earlier than it having caught up. Only `restart`
is negatable: `! validator restart N` asserts the node cannot come back, and the error carries a tail of the node's
stderr so the scenario can name the reason it died.

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
| `GPAO_ADDR` | the oracle's address, or the zero address when no scenario in the run declared `oracle: true` | always |
| `GPAO_STATUS` | base URL of the running oracle's status board | after `gpao start` |
| `WORK` | testscript's own working directory, where the script's file sections are unpacked | always |

## Identities

Two accounts, imported once into a keybase every cluster in the run shares, and funded with 1000 GNOT on each of
those clusters.

| Account | Keybase name | Key | Role |
| --- | --- | --- | --- |
| test user | `test1` | derived at account 0 index 0 from the test1 seed `gno.land/pkg/integration` uses, the `-mnemonic` default | signs everything a script submits |
| oracle | `gpao` | derived at account 0 index 0 from the BIP39 test vector (`abandon` eleven times, then `about`), the `-gpao-mnemonic` default | signs the oracle's approvals |

The oracle's key exists only when some scenario in the run declared `oracle: true`, since deriving it means building
`gpao` and funding an account. It is derived once and used twice, for the genesis approver entry and for the keybase
import, so the chain cannot end up with an approver nobody can sign for. Being in the keybase is what lets a script
do by hand what the oracle does, with `gnokey maketx enablepkg ... $GPAO_KEY_NAME`.

The oracle needs a separate mnemonic rather than another index in the test user's, because it derives its signer at
account 0 index 0 with no way to change that.

An oracle that is not an approver is the one misconfiguration that leaves it looking healthy. It starts, follows
blocks, verifies packages and reports on them, and every package it approves stays inert. The run logs a warning
per scenario where the oracle's key is absent from the chain's approver set, because nothing else about the run will
say so: the only symptom is deploys that never activate.

## Scenarios

`testdata/integration`

- `smoke.txtar`: the node is up and gnokey can reach it. One validator.

`testdata/oracle`

- `false_start.txtar`: a dead RPC endpoint is fatal at startup under the default `-start-height`, and merely
  retried inside the poll loop under `-start-height 1`. Same call, same endpoint, two outcomes.
- `first_light.txtar`: a package submitted after genesis parks, the oracle activates it, and all three nodes serve
  the result. Submission enters through one node and the oracle works through another.
- `phantom_approval.txtar`: a private realm redeployed over its own live version is reported approved though no
  enable was ever sent, because the check for "already active" cannot see a redeploy in the inert key space.
- `poisoned_dependent.txtar`: a dependent submitted before its dependency is rejected permanently by content hash,
  and resubmitting the same bytes is a silent no-op.
- `uncollected_toll.txtar`: the run budget is debited before an approval is attempted, so approvals that never
  reach the chain still cost the run its allowance.

`testdata/oracle-budget`

- `exhausted_purse.txtar`: `-max-spend` measured against fees the chain really charged: two approvals leave the
  approver exactly two fees down, and the third is refused.
- `starved_verifier.txtar`: a 1ms verification budget kills every child before it can finish, and what it leaves
  behind is the point: the packages stay parked.

`testdata/oracle-closure`

- `serialized_closure.txtar`: a real 26-package dependency chain where each link cannot be verified until the one
  before it has committed. Waiting on the last package is the whole proof.

`testdata/oracle-containment`

- `contained_blast.txtar`: a package built to kill the thing that inspects it kills only the verification child,
  and the oracle carries on.

`testdata/oracle-gasceiling`

- `borrowed_ceiling.txtar`: on a chain whose block gas ceiling is not the usual 3000000000, the oracle has to read
  the ceiling rather than fall back to its default, and has to clamp an operator's `-gas-wanted` to it.

`testdata/oracle-outage`

- `amnesiac_oracle.txtar`: an oracle restarted with no `-start-height` resumes past the node's tip, so it strands
  every submission made while it was away and nothing ever revisits them.
- `patient_oracle.txtar`: losing two validators of four halts consensus while the survivors keep serving RPC. The
  oracle waits on a frozen tip without spinning, exiting or losing its place.

`testdata/oracle-unauthorized`

- `rotated_out.txtar`: the oracle's key is not in the chain's approver set. It verifies everything and activates
  nothing, pays no gas fee because the simulate pre-flight stops each enable before it is broadcast, and still
  spends its own run budget.

`testdata/tour`

- `every_verb.txtar`: the worked example. Every verb, every cluster key, one four-validator inert chain: overrides
  read back out of the chain, a package parked and enabled, a validator down and back, the oracle down and back,
  and last an enable by hand with no oracle running.

## Limits

Scenarios run one at a time, on both routes. Every script in a run shares one keybase, and a four-validator cluster
plus the oracle already fills a four-vCPU runner.

The timeouts are sized for a workstation: 60s for the cluster's first block, 90s for a node to answer RPC, 30s for
the oracle's status board, 30s for a default `eventually`. The whole suite takes about two and a half minutes. A
machine slow enough to miss one of these fails the scenario rather than waiting longer.

`run -timeout` bounds the whole run rather than each scenario, and defaults to 10 minutes. The validator processes
are started from that deadline's context, so reaching it takes the nodes out from under whichever scenario is
running and the run ends reporting that scenario as failed. The `go test` route has no such flag and is bounded by
`go test -timeout` instead, which is why the Makefile passes `-timeout 30m`.

A run that unwinds normally removes what it made. A run killed outright does not, and leaves directories named
`e2e-cluster-*`, `gnoe2e-bin-*` and `gnoe2e-home-*` in `TMPDIR`. Node logs live in the cluster directory as
`validator_N/stdout.log` and `validator_N/stderr.log`, which is where to look after a kill.

Port allocation binds an ephemeral port, reads its number and closes the listener before the node binds it, so two
clusters starting at the same moment can be handed the same port. The window is small and the failure is a node
that will not start, not a silent wrong result.
