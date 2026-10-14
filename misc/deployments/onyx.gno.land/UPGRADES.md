# onyx-1 upgrades

Every binary onyx has run, one row per version. Onyx runs mainnet's code, one release candidate ahead, and is upgraded whenever mainnet is, so this ledger mirrors [mainnet's](../mainnet.gno.land/UPGRADES.md) one release ahead: every mainnet release is rehearsed here first, as a release candidate on the same commit. The source of truth is [`upgrades.json`](./upgrades.json) (format: [`misc/deployments/upgrades/upgrades.schema.json`](../upgrades/upgrades.schema.json), rules in `gno.land/pkg/upgrades`); the table below is generated from it and never edited by hand:

```shell
go run ./misc/deployments/upgrades render misc/deployments/onyx.gno.land   # after editing upgrades.json
go run ./misc/deployments/upgrades check  misc/deployments/onyx.gno.land   # what CI runs
```

An entry is added before the release is cut (see [`RELEASING.md`](../../../RELEASING.md)); the image digest and the binaries are filled in together once CI has built the release, the proposal once it exists, the halt time once the halt has happened. Until then they are null, and the table shows the digest, proposal and halt time as *(pending)*.

- **To run a node today:** the version in the last row with no *(pending)* cell, pinned — the native binary from its release page (checksums in `binaries`), or `ghcr.io/gnolang/gno/gnoland:<version>`. Never a floating tag. Configuration: [`VALIDATOR.md`](./VALIDATOR.md).
- **Joining from genesis:** the node stops at every halt height below (the historical governance halts fire during replay too); restart it with the same binary if it satisfies that row's `halt_min_version`, otherwise with that row's version.
- **How a coordinated upgrade works** (halt height, `halt_min_version`, what the node refuses): [`gno.land/cmd/gnoland/UPGRADES.md`](../../../gno.land/cmd/gnoland/UPGRADES.md).
- A version runs the blocks from its own row's halt height + 1 (1 for genesis) up to and including the next row's halt height. A *rolling* row is a patch with no halt: it can serve the same blocks as the row before it. Halt time is the block time of the halt block.

<!-- BEGIN GENERATED (gno.land/pkg/upgrades) -->
| Version | Commit | Halt height | Halt time (UTC) | halt_min_version | GovDAO proposal | gnoland image digest |
|---|---|---|---|---|---|---|
| [v1.5.0](https://github.com/gnolang/gno/releases/tag/v1.5.0) | [e75fef82c](https://github.com/gnolang/gno/commit/e75fef82c02876a4df92ad6e325c5479b9532168) | genesis | 2026-09-28T00:00:00Z | — | — | `sha256:6ae099c2828c…` |
| [v1.5.1-rc.1](https://github.com/gnolang/gno/releases/tag/v1.5.1-rc.1) | [07e44776c](https://github.com/gnolang/gno/commit/07e44776c26187830b7dcbd2454991b5a1960378) | rolling | — | — | — | `sha256:231eedbda83b…` |
<!-- END GENERATED -->

Notes

- Onyx launched on `v1.5.0`, the version mainnet was running on launch day, with no release candidate: there was no new code to rehearse. Since `v1.5.1`, the row here names the candidate onyx actually ran (`v1.5.1-rc.1`), and mainnet's names the final tag on the same commit.
