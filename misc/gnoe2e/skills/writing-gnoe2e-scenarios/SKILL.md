---
name: writing-gnoe2e-scenarios
description: Use when writing, changing or reviewing a txtar scenario under misc/gnoe2e/testdata, when a change to gnoland, gpao or the harness needs an end-to-end proof across several validators or through the package-approver oracle, or when a scenario is flaky, passes by luck, or waits with sleep.
---

# Writing gnoe2e scenarios

A gnoe2e scenario is one claim about the chain, the oracle or the harness, written as a txtar script that fails
when the claim stops being true. This skill is the entry point; the pages below own the method and the reference.
Read them from the module root, `misc/gnoe2e/`.

1. `docs/writing-scenarios.md`: the method. Shape, choosing the cluster, one-shot versus wait, negation,
   comments, determinism, the pre-commit checklist. Read it whole before writing a line.
2. `docs/dialect.md`: the reference. Every verb's usage, every `-- cluster --` key, every exported variable, the two
   accounts. Read the row for each verb you use.
3. `testdata/tour.txtar`: the worked example. Copy its shape, not its content.
4. `docs/architecture.md`: only when the scenario needs something the harness does not offer yet.

Before declaring a `config.` or `genesis.` key, run `make defaults`: it lists every key a `-- cluster --` section
takes and the value a cluster boots with, so a key is checked rather than guessed.

## Non-negotiable

- Every declared cluster setting is read back by a line in the script.
- Anything the oracle does is an `eventually`; a read of a write this script committed, on the same node, is one
  shot; `sleep` only observes an absence, and the comment says why the window is long enough.
- Every `!` line is pinned by the `stderr` or `stdout` match that follows it.
- Three consecutive passes with `go run . run -verbose testdata/oracle/<file>.txtar` before committing, and one line
  added to the index in `README.md`.

## When the claim is a chain bug

Write the scenario so it is red on `master` and green with the fix (`make test-master` runs the working tree's
scenarios against binaries built from master). It lands in the same change as the fix, asserting the fixed
behaviour, with a header that says what used to fail.
