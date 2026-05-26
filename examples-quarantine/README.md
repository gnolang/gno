# examples-quarantine

Packages from the legacy `examples/` tree that are not part of the **test-13
genesis** (see [PR #5653](https://github.com/gnolang/gno/pull/5653) and the
manifest in `misc/quarantine/`).

These packages remain in the repository as integration-test fodder and history
but are **not** shipped to genesis and have **not** been audited or modernized
to the current interrealm spec. Do not deploy them and do not rely on their
behavior.

## What goes here vs `examples/`

- `examples/` — the audited set that ships at test-13 genesis (88 packages),
  plus packages that gnovm tests or `gno.land/pkg/integration/testdata/`
  txtar scripts reference by hardcoded import path, transitively closed
  over their deps (the `gnovm-pinned` set in `misc/quarantine/`, 44
  packages). Both stay in `examples/`.
- `examples-quarantine/` — everything else from the legacy `examples/` tree
  (171 packages).

Concretely: if `examples/gno.land/r/tests/vm/crossrealm_b` is still in
`examples/` despite not being on `safe-list.txt`, it's because a
`gnovm/tests/files/*.gno` filetest imports it. Similarly,
`examples/gno.land/r/demo/defi/foo20` stays because a txtar integration
script does `loadpkg gno.land/r/demo/defi/foo20`. See
`misc/quarantine/gnovm-pinned.txt`.

The split is computed by `misc/quarantine/derive.sh` from the canonical
`safe-list.txt`. To re-derive, run that script and then `misc/quarantine/move.sh`.

## Testing

```sh
make test         # runs the Go driver: gno tests + integration-node load
make test.load    # only the genesis-load smoke test
```

The driver lives in `quarantine_test.go`. It builds a unified package list
spanning `examples/` and `examples-quarantine/` so cross-tree imports
resolve, then runs gno tests against each quarantined package.

`TestQuarantineRealms` runs the unit-test sweep.
`TestQuarantineRealmsLoad` boots an in-memory gnoland node and deploys every
quarantined package at genesis; it fails on the first AddPackage failure.

## Import paths are preserved

Every package's `gnomod.toml` keeps its original `module = "gno.land/..."`
line. Nothing else in the codebase needs to know whether a package lives in
`examples/` or `examples-quarantine/`.
