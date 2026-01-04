# Various upgrade pattern explorations

This repository explores different upgrade patterns for Gno smart contracts.

## `upgrade_a`

- Versions are independent.
- Versions are not pausable; users can interact with them independently.
- New versions wrap the previous one (can be recursive) to extend the logic and optionally the storage.
- There is no consistency between versions; updating a version will impact the more recent ones but won't affect the older ones.
- Users and contracts interacting with non-latest versions won't have the latest state.

## `upgrade_b`

- Versions include a `SetNextVersion` function which pauses the current implementation and invites users interacting with a deprecated version to switch to the most recent one.
- Since only one version can be used at a time, the latest version can safely recycle the previous version's state in read-only mode.
- These logics can be applied recursively.
- Users and contracts interacting with non-latest versions will switch to a more restricted version (read-only).

## `upgrade_c`

- `root` is the storage contract with simple logic.
- Versions implement the logic and rely on `root` to manage the state.
- In the current example, only one version can write to `root` (the latest); in practice, it could be possible to support various logics concurrently relying on `root` for storage.

## `upgrade_d` -- "Lazy Migration"

- Demonstrates lazy migrations from v1 to v2 of a data structure in Gno.
- Uses AVL trees, but storage can vary since public `Get` functions are used.
- v1 can be made pausable and read-only during migration.

## `upgrade_e`

- `home` is the front-facing contract, focusing on exposing a consistent API to users.
- Versions implement an interface that `home` looks for and self-register themselves, which instantly makes `home` use the new logic implementation for ongoing calls.

## `upgrade_f`

- Similar to `upgrade_e`.
- Replaces self-registration with manual registration by an admin.
