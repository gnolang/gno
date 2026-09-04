# How a run works

gnoe2e is a Go module with one binary and five internal packages. A run turns a list of txtar files into one
booted cluster per scenario, drives each script through testscript, and tears the cluster down again.

| Package | Owns |
| --- | --- |
| `internal/builder` | building `gnoland` and `gpao` from the enclosing checkout into a directory the run owns |
| `internal/cluster` | genesis, per-node config, ports, starting and stopping validator processes, reading their logs |
| `internal/daemon` | supervising one long-running child (readiness probe, captured output, stop); knows nothing about gpao |
| `internal/integration` | the txtar verbs, the `-- cluster --` parser, the environment exported to scripts, the testscript runner |
| `internal/termlog` | the coloured slog handler the CLI logs through |

The `main` package at the module root is the CLI (`run`, `serve`) and the `go test` driver (`scenarios_test.go`).
Both routes call the same two functions, so a scenario behaves the same under `go run . run` and under
`go test -run TestScenarios`.

## One run

1. **Resolve** the arguments to scenario files, in the order given; a directory contributes its files sorted.
2. **Parse** every file's `-- cluster --` section up front, resolving the `config.` and `genesis.` paths
   against the defaults, so a misspelled key or a value its field cannot hold fails before anything boots.
3. **Prepare the suite**, once: a keybase in a temp directory holding the test user and the oracle key, and a
   `gnoland` binary built from the checkout `gnoenv.RootDir()` names. `gpao` is not built here.
4. **Per scenario**: apply the section onto a cluster config (validators, policy, approver, gas ceiling, overrides),
   fund both identities in genesis, write the genesis document, and for each validator write its config in four
   passes: defaults with the harness's ports, peer topology, fast consensus timing, then the scenario's
   `config.<key>` overrides. Start the processes, wait for the first committed block, run the script, stop
   everything, remove the cluster directory.
5. **Report** every scenario's result; one failure does not stop the others.

The consensus pass sets every timeout to 10ms and keeps empty blocks coming on a 10ms interval, so a cluster
commits continuously rather than only when a transaction arrives. A restarted validator catches up almost at once,
and a `sleep` window runs on a chain that is demonstrably live. The price is paid in determinism: height becomes a
function of how long the cluster has been up, and so does block time, which runs ahead of the wall clock because
each vote time is at least the previous block's plus `TimeIotaMS`, 100ms, while commits are 10ms apart
(`consensus/state.go:1761`). A scenario that needs either to be a function of its own transactions declares
`config.consensus.create_empty_blocks: false` and pays for liveness explicitly, with a transaction rather than a
wait.

`gpao` is built on the first `gpao start` of the run and reused by every later one, so a run whose scenarios never
start the oracle never pays for the build.

## The checkout

Every path into the checkout goes through `gnoenv.RootDir()`: the builder's source directories, the `examples/`
tree genesis packages come from, the `GNOROOT` exported to scripts, and gpao's `-gno-root`. `RootDir()` reads
`GNOROOT` first and otherwise asks `go list` for the `github.com/gnolang/gno` module directory, which the module's
`replace github.com/gnolang/gno => ../..` resolves to the enclosing checkout.

That single knob is what `make test-master` uses: it unpacks `master` with `git archive` into a temp directory and
runs the suite with `GNOROOT` pointing there, so the scenarios and the harness come from the working tree while
`gnoland` and `gpao` come from master.

`GNOROOT` reaches the two spawned binaries and the `examples/` tree, and nothing else. The module's
`replace github.com/gnolang/gno => ../..` links the harness against the working tree's own gno packages, so
in-process `gnokey`, the genesis document `BuildGenesis` assembles, and the packages `LoadPackagesFromDir` reads
are the working tree's under `test-master` as well. A claim that lives in one of those is green either way, and
proving it red needs a checkout of master rather than this knob.

## Ports, addresses, processes

Each node gets two free ports from the OS, one for RPC and one for P2P, and the RPC address is exported as
`RPC_ADDR_<index>`. Validators run as child processes of the run with their context cancelled when the run's
deadline passes. A stopped validator keeps its data directory and identity; `validator restart` starts the same
process again against the same chain. Node logs are `validator_N/stdout.log` and `validator_N/stderr.log` in the
cluster's temp directory, and the tail of a node's stderr is attached to the error when a node fails to come up.

## Where things run in CI

`.github/workflows/ci-gnoe2e.yml` runs lint, the `go fix` check and `go test -timeout 30m ./...` for this module on
every pull request that touches it or the code it exercises: `contribs/gpao`, `examples`, `gno.land`, `gnovm`, `tm2`.
The suite runs sequentially and takes a few minutes.
