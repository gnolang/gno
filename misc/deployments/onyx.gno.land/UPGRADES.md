# onyx-1 upgrades

Every binary onyx has run, one row per version. Onyx runs mainnet's code and takes every mainnet release first, as a release candidate on the same commit, so each row here mirrors a row of [mainnet's ledger](../mainnet.gno.land/UPGRADES.md), except a candidate whose release mainnet skips (see Notes). The source of truth is [`upgrades.json`](./upgrades.json) (format: [`misc/deployments/upgrades/upgrades.schema.json`](../upgrades/upgrades.schema.json), rules in `gno.land/pkg/upgrades`); the table below is generated from it and never edited by hand:

```shell
go run ./misc/deployments/upgrades render misc/deployments/onyx.gno.land   # after editing upgrades.json
go run ./misc/deployments/upgrades check  misc/deployments/onyx.gno.land   # what CI runs
```

An entry is added once the chain runs the release — after the halt for a coordinated upgrade, once validators run it for a rolling one (see [`RELEASING.md`](../../../RELEASING.md)) — when everything it records is known: the halt height and the proposal once the proposal has executed, the halt time once the chain has halted, the image digest and binaries once CI has built the release. A cell reads *(pending)* only for a fact that has not happened yet.

- **To run a node today:** the version in the last row with no *(pending)* cell, pinned — the native binary from its release page (checksums in `binaries`), or `ghcr.io/gnolang/gno/gnoland:<version>`. Never a floating tag. Configuration: [`VALIDATOR.md`](./VALIDATOR.md).
- **Joining from genesis:** the node stops at every halt height below (the historical governance halts fire during replay too); restart it with the same binary if it satisfies that row's `halt_min_version`, otherwise with that row's version.
- **How a coordinated upgrade works** (halt height, `halt_min_version`, what the node refuses): [`gno.land/cmd/gnoland/UPGRADES.md`](../../../gno.land/cmd/gnoland/UPGRADES.md).
- A version runs the blocks from its own row's halt height + 1 (1 for genesis) up to and including the next row's halt height. A *rolling* row is a patch with no halt: it can serve the same blocks as the row before it. Halt time is the block time of the halt block.

<!-- BEGIN GENERATED (gno.land/pkg/upgrades) -->
| Version | Commit | Halt height | Halt time (UTC) | halt_min_version | GovDAO proposal | gnoland image digest |
|---|---|---|---|---|---|---|
| [v1.5.0](https://github.com/gnolang/gno/releases/tag/v1.5.0) | [e75fef82c](https://github.com/gnolang/gno/commit/e75fef82c02876a4df92ad6e325c5479b9532168) | genesis | 2026-09-28T00:00:00Z | — | — | `sha256:6ae099c2828c…` |
| [v1.5.1-rc.1](https://github.com/gnolang/gno/releases/tag/v1.5.1-rc.1) | [07e44776c](https://github.com/gnolang/gno/commit/07e44776c26187830b7dcbd2454991b5a1960378) | rolling | — | — | — | `sha256:231eedbda83b…` |
| [v1.6.0-rc.1](https://github.com/gnolang/gno/releases/tag/v1.6.0-rc.1) | [c868ccab1](https://github.com/gnolang/gno/commit/c868ccab1f93f0542a4c10a9d9b49b6fabcc0d98) | 445000 | 2026-10-16T12:51:52Z | `v1.6.0-rc.1` | [#33](https://onyx.testnets.gno.land/r/gov/dao:33) | `sha256:4286ffcd805f…` |
<!-- END GENERATED -->

Notes

- Onyx launched on `v1.5.0`, the version mainnet was running on launch day, with no release candidate: there was no new code to rehearse. Since `v1.5.1`, the row here names the candidate onyx actually ran (`v1.5.1-rc.1`), and mainnet's names the final tag on the same commit.
- `v1.6.0-rc.1` has no counterpart in mainnet's ledger: mainnet skips `v1.6.0` and takes its changes with `v1.7.0`.
