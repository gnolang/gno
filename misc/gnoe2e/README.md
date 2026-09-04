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

`testdata/tour.txtar` is the worked example: every verb, every `-- cluster --` key and every variable a
script can name, each read back from the chain, with a comment per line saying what the line proves.

## Running

A run needs the `go` toolchain: `gnoland` is built on demand from the enclosing checkout, and `gpao` on the first
`gpao start` of the run. That checkout is found through `go list`, so `GNOROOT` must not be set in the environment
unless it deliberately points at another tree.

```bash
# One lane of scenarios through the CLI
cd misc/gnoe2e && go run . run testdata/oracle

# Named files and directories mix, and the argument order is the run order
go run . run testdata/oracle/first_light.txtar testdata/tour.txtar

# No argument runs the tour
go run . run

# Every scenario, coloured, verbose
make scenarios

# Every -- cluster -- key with the value a cluster boots with
make defaults

# The unit tests, in seconds, under the race detector, booting nothing
make test

# The scenarios through go test, then both lanes one after the other
make test-scenarios
make test-all

# One scenario, named by its path under testdata/ without the extension
go test -timeout 30m -run 'TestScenarios/oracle/first_light' .

# The scenarios against gnoland and gpao built from master, so a scenario
# written for a fix can be shown red before the fix lands
make test-master
SCENARIO_FLAGS='-timeout 30m -run TestScenarios/oracle/first_light' make test-master

# A cluster to poke at by hand: prints its RPC address and blocks until Ctrl-C
go run . serve -validators 3 -verbose

# The repository's linter, on this module
make lint
```

`make build` puts the binary in `build/gnoe2e` and `make install` installs it, for anyone who would rather not go
through `go run` each time.

`make test-master` unpacks master with `git archive` into a temp directory and points `GNOROOT` at it, which is the
one case where setting `GNOROOT` is the point. `GOTEST_FLAGS` (`-v -race`) carries the unit lane and
`SCENARIO_FLAGS` (`-v -timeout 30m -run TestScenarios`) the scenario lane; giving either replaces the whole value,
timeout included. `make help` lists every target.

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
override is for. `-code-submission-policy inert` needs no approver on the command line: every cluster in a run gets
the oracle as its approver unless the scenario names somebody else. The remaining flags (`-chain-id`, `-max-tx-bytes`, `-max-data-bytes`, `-block-time-iota`,
`-load-examples`) set the base every cluster in the run starts from; no scenario declares them.

## Documentation

| Page | Read it when |
| --- | --- |
| [`docs/dialect.md`](docs/dialect.md) | you need the exact syntax of a verb, a cluster key, an exported variable, or the two accounts |
| [`docs/writing-scenarios.md`](docs/writing-scenarios.md) | you are writing or reviewing a scenario: what to wait on, what to negate, what to read back |
| [`docs/architecture.md`](docs/architecture.md) | you are changing the harness itself, or wondering how a run boots and tears down a cluster |
| [`AGENTS.md`](AGENTS.md) | you are an AI agent working in this directory |

## Scenarios

`testdata/tour.txtar`: the worked example. Every verb, every cluster key, one four-validator inert chain: every
setting read back out of the chain, a package parked and enabled, a validator down and back, the oracle down and
back, and last an enable by hand with no oracle running.

`testdata/oracle`

- `first_light.txtar`: a package submitted after genesis parks, the oracle activates it, and all three nodes serve
  the result. Submission enters through one node and the oracle works through another.
- `patient_oracle.txtar`: losing two validators of four halts consensus while the survivors keep serving RPC. The
  oracle waits on a frozen tip without spinning, exiting or losing its place.

## Limits

Scenarios run one at a time, on both routes. Every script in a run shares one keybase, and a four-validator cluster
plus the oracle already fills a four-vCPU runner.

The timeouts are sized for a workstation: 60s for the cluster's first block, 90s for a node to answer RPC, 30s for
the oracle's status board, 30s for a default `eventually`. The whole suite takes about a minute. A
machine slow enough to miss one of these fails the scenario rather than waiting longer.

`run -timeout` bounds the whole run rather than each scenario, and defaults to 10 minutes. The validator processes
are started from that deadline's context, so reaching it takes the nodes out from under whichever scenario is
running and the run ends reporting that scenario as failed. The scenarios that had not started are not attempted,
and the count that ends the run is out of what it got to. Ctrl-C ends it the same way. The `go test` route has no
such flag and is bounded by `go test -timeout` instead, which is why the scenario targets pass `-timeout 30m`.

A run removes what it made when it unwinds, and Ctrl-C, `kill` and a cancelled CI job all unwind it. `kill -9` does
not: it leaves the temp directories named `e2e-cluster-*`, `gnoe2e-bin-*`, `gnoe2e-home-*` and `go-test-script*` in
`TMPDIR`, together with the processes of whichever scenario was in flight, up to four gnoland validators and one
gpao. `pkill -f e2e-cluster` clears those nodes. Node logs live in the cluster directory as
`validator_N/stdout.log` and `validator_N/stderr.log`, which is where to look afterwards.

Port allocation binds an ephemeral port, reads its number and closes the listener before the node binds it, so two
clusters starting at the same moment can be handed the same port. The window is small and the failure is a node
that will not start, not a silent wrong result.
