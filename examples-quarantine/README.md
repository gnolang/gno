# examples-quarantine

Packages from the legacy `examples/` tree that are not part of the **test-13
genesis** (see [PR #5653](https://github.com/gnolang/gno/pull/5653)).

These packages remain in the repository as integration-test fodder and
history but are **not** shipped to genesis and have **not** been audited or
modernized to the current interrealm spec. Do not deploy them and do not rely
on their behavior.

## What goes here vs `examples/`

- `examples/` keeps the audited set that ships at test-13 genesis, plus
  packages that `gnovm/` tests or the `gno.land/pkg/integration/testdata/`
  txtar harness reference via hardcoded import paths (transitively closed
  over deps). Both stay in `examples/` because their consumers resolve
  packages out of `examples/<importpath>` directly and would break otherwise.
- `examples-quarantine/` holds everything else from the legacy tree.

Module import paths are preserved via unchanged `gnomod.toml` entries, so
on-chain identity is unaffected by the move.

## Testing

```sh
make test         # full Go driver (gno tests + integration-node load)
make test.load    # only the genesis-load smoke test
```

The driver lives in `quarantine_test.go`. `TestQuarantineRealms` builds a
unified package list spanning both trees and runs gno tests against each
quarantined package. `TestQuarantineRealmsLoad` boots an in-memory gnoland
node and deploys every quarantined package at genesis, reporting all
`AddPackage` failures (not just the first).
