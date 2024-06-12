# Various upgrade pattern explorations.

## `upgrade_a`

- versions are independent.
- versions are not pausable and users can interact independently.
- new versions wrap the previous one (can be recursive) to extend the logic and optionally the storage too
- there is no consistency between versions; updating a version will impact the more recent ones but won't impact the older ones
- users and contracts interacting/importing non-latest version won't have the latest state

## `upgrade_b`

- versions embed a `SetNextVersion` which pauses the the current implementation and invite users interacting with a deprecated version to switch to the most recent one.
- since there is a single version that can be used at the same time; the latest version can safely decide to recycle the previous version state in read-only mode.
- these logics can be applied recursively.
- users and contracts interacting/importing non-latest versions will switch to a more restricted version (read-only)

## `upgrade_c`

- `root` is the storage contract with simple logic
- versions impelments the logic and rely on root to manage the state
- in the current example, only one version is able to write to `root` (the latest); in practice, it could be possible to support various logic to concurrently rely on `root` for the storage.

## `upgrade_d` -- "lazy migration"

- demonstrates lazy migrations from v1 to v2 of a data structure in Gno.
- using AVL trees, but storage can vary since public Get functions are used.
- v1 can be made pausable and readonly during migration.

## `upgrade_e`

TODO
