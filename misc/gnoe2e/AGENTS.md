# AGENTS.md — misc/gnoe2e

gnoe2e is a Go module that runs txtar scenarios against real multi-validator gnoland clusters built from this
checkout, with the package-approver oracle (`contribs/gpao`) supervised from the script. It has its own CI
workflow, `.github/workflows/ci-gnoe2e.yml`. The repository-wide guide at the root `AGENTS.md` still applies here.

## Read first

| Task | Read |
| --- | --- |
| any work in this directory | `README.md` |
| writing or changing a scenario | `skills/writing-gnoe2e-scenarios/SKILL.md`, then `docs/writing-scenarios.md` and `docs/dialect.md` |
| changing the harness (verbs, cluster, builder, CLI) | `docs/architecture.md`, then the package you touch and its tests |

## Commands

```bash
cd misc/gnoe2e
make test                                             # unit tests, seconds
make test-scenarios                                   # every scenario, a few minutes
make test-all                                         # both lanes
go run . run -verbose testdata/oracle/<file>.txtar    # one scenario with every line echoed
make test-master                                      # the scenarios against gnoland and gpao built from master
make lint                                             # the repository's linter on this module
```

`GNOROOT` must not be set in the environment unless the intent is to point the run at another checkout; that is
what `make test-master` does for you.

## Rules for this module

- A new verb, key or exported variable is not done until `docs/dialect.md` documents it and
  `testdata/tour.txtar` uses it. The tour is the dialect's regression test.
- Harness changes follow test-first: a failing test in the package before the code. Scenario changes are proven by
  running the scenario three times.
- A scenario never uses `sleep` to wait for something that will happen; see `docs/writing-scenarios.md`.
- Scenarios read back every setting they declare. A `config.` or `genesis.` override that nothing reads is not
  accepted.
- Every scenario has one line in the index in `README.md`.
- Comments in scenarios and Go say what a line proves or why it is shaped that way. No change history, no
  references to plans or tasks, no descriptions of neighbouring code.
- Do not commit scratch scenarios or files under `testdata/` that are not part of the suite; the `go test` driver
  runs every `testdata/*.txtar` and `testdata/*/*.txtar`.
