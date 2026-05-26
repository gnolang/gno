# Quarantine manifest

Inputs that govern which packages live under `examples/` vs `examples-quarantine/`.

## Files

- `safe-list.txt` — the 88 packages that ship at test-13 genesis. Sourced verbatim from PR [#5653 comment 4431729886](https://github.com/gnolang/gno/pull/5653#issuecomment-4431729886).
- `move-list.txt` — the 195 packages currently under `examples/gno.land/{p,r}/` that are NOT on the safe list AND are NOT referenced by hardcoded import paths in `gnovm/` tests. These are the packages that move to `examples-quarantine/`.
- `gnovm-pinned.txt` — the packages that are NOT on the safe list but must stay in `examples/` because they're referenced by hardcoded import paths in `gnovm/` tests OR in `gno.land/pkg/integration/testdata/` txtar scripts. Transitively closed: deps of pinned packages are also pinned. Moving them would break those harnesses, which resolve packages out of `examples/<importpath>` directly. Current count: 44.
- `derive.sh` — reproducible script that regenerates `move-list.txt` and `gnovm-pinned.txt` from `safe-list.txt` and the current state of `examples/`.

## Regenerating

```sh
./misc/quarantine/derive.sh
```

Run from repo root.
