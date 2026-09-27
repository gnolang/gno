# gnoland-1 upgrades

Every binary the gno.land mainnet has run, one row per version. The source of truth is [`upgrades.json`](./upgrades.json) (format: [`misc/deployments/upgrades/upgrades.schema.json`](../upgrades/upgrades.schema.json), rules in `gno.land/pkg/upgrades`); the table below is generated from it and never edited by hand:

```shell
go run ./misc/deployments/upgrades render misc/deployments/mainnet.gno.land   # after editing upgrades.json
go run ./misc/deployments/upgrades check  misc/deployments/mainnet.gno.land   # what CI runs
```

An entry is added when the release is cut (see [`RELEASING.md`](../../../RELEASING.md)); the image digest is filled in once CI has built the image, the proposal once it exists, the halt time once the halt has happened — each shows as *(pending)* until then.

- **To run a node today:** the version in the last row, pinned — the native binary from its release page (checksums in `binaries`), or `ghcr.io/gnolang/gno/gnoland:<version>`. Never a floating tag. Configuration: [`VALIDATOR.md`](./VALIDATOR.md).
- **Joining from genesis:** the node stops at every halt height below (the historical governance halts fire during replay too); restart it with the same binary if it satisfies that row's `halt_min_version`, otherwise with that row's version. Verified on 2026-09-22 with `v1.5.0` alone, genesis to tip.
- **How a coordinated upgrade works** (halt height, `halt_min_version`, what the node refuses): [`gno.land/cmd/gnoland/UPGRADES.md`](../../../gno.land/cmd/gnoland/UPGRADES.md).
- A version runs the blocks from the previous row's halt height + 1 (1 for genesis) up to and including the next row's halt height. Halt time is the block time of the halt block.

<!-- BEGIN GENERATED (gno.land/pkg/upgrades) -->
| Version | Commit | Halt height | Halt time (UTC) | halt_min_version | GovDAO proposal | gnoland image digest |
|---|---|---|---|---|---|---|
| [v1.2.0](https://github.com/gnolang/gno/releases/tag/v1.2.0) | [9c8eb132e](https://github.com/gnolang/gno/commit/9c8eb132e483d6fd324d92c193e629ad65a98a37) | genesis | 2026-09-12T15:00:00Z | — | — | `sha256:4b161a2b5d5f…` |
| [v1.3.0](https://github.com/gnolang/gno/releases/tag/v1.3.0) | [31b6650a1](https://github.com/gnolang/gno/commit/31b6650a100d9baf14e7669f8f0df924f1f841e0) | 36300 | 2026-09-14T09:17:06Z | *(not set)* | [#0](https://gno.land/r/gov/dao:0) | `sha256:897a5c0f75e6…` |
| [v1.4.0](https://github.com/gnolang/gno/releases/tag/v1.4.0) | [00417a1be](https://github.com/gnolang/gno/commit/00417a1be97b9a311d9669ae7aa9585b277ee594) | 113000 | 2026-09-17T13:24:04Z | *(not set)* | [#5](https://gno.land/r/gov/dao:5) | `sha256:a41395a8cf39…` |
| [v1.5.0](https://github.com/gnolang/gno/releases/tag/v1.5.0) | [e75fef82c](https://github.com/gnolang/gno/commit/e75fef82c02876a4df92ad6e325c5479b9532168) | 162200 | 2026-09-19T10:37:10Z | *(not set)* | [#7](https://gno.land/r/gov/dao:7) | `sha256:6ae099c2828c…` |
<!-- END GENERATED -->

Notes

- `v1.3.0`, `v1.4.0` and `v1.5.0` were tagged after the fact, on 2026-09-22. At each halt, validators pulled the `chain-mainnet` image current at the time (`ran_as` in `upgrades.json`): the same code as the `:v1.x.0` images, reporting an unparseable version.
- `halt_min_version` is *(not set)* for those three halts: the proposals set no floor. From `v1.6.0` on, every halt proposal names the version.
