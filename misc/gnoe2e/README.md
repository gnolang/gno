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

`testdata/tour/every_verb.txtar` is the worked example: every verb, every `-- cluster --` key and every variable a
script can name, each read back from the chain, with a comment per line saying what the line proves.

## Running

A run needs the `go` toolchain: `gnoland` is built on demand from the enclosing checkout, and `gpao` on the first
`gpao start` of the run. That checkout is found through `go list`, so `GNOROOT` must not be set in the environment
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

Four `run` flags override what a scenario declared:

| Flag | Overrides |
| --- | --- |
| `-validators N` | `validators` |
| `-code-submission-policy <policy>` | `code-submission-policy` |
| `-pkg-approver <address>` | `pkg-approver` |
| `-block-max-gas N` | `block-max-gas` |

A flag counts as an override only when it is actually given, so leaving `-validators` alone is not the same as
passing `-validators 2`. A given flag beats every scenario's declaration, including scenarios that cannot pass
without theirs: running a four-validator outage scenario with `-validators 2` turns it red, and that is what an
override is for. The remaining flags (`-chain-id`, `-max-tx-bytes`, `-max-data-bytes`, `-block-time-iota`,
`-load-examples`) set the base every cluster in the run starts from; no scenario declares them.

## Documentation

| Page | Read it when |
| --- | --- |
| [`docs/dialect.md`](docs/dialect.md) | you need the exact syntax of a verb, a cluster key, an exported variable, or the two accounts |
| [`docs/writing-scenarios.md`](docs/writing-scenarios.md) | you are writing or reviewing a scenario: what to wait on, what to negate, what to read back |
| [`docs/architecture.md`](docs/architecture.md) | you are changing the harness itself, or wondering how a run boots and tears down a cluster |
| [`AGENTS.md`](AGENTS.md) | you are an AI agent working in this directory |

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

- `every_verb.txtar`: the worked example. Every verb, every cluster key, one four-validator inert chain: every
  setting read back out of the chain, a package parked and enabled, a validator down and back, the oracle down and
  back, and last an enable by hand with no oracle running.

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
