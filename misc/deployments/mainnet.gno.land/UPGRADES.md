# gnoland-1 upgrades

Every binary the gno.land mainnet has run, one row per version. The source of truth is [`upgrades.json`](./upgrades.json); the table below is generated from it (`./render-upgrades.sh`) — do not edit it by hand. An entry is added when the release is cut (see [`RELEASING.md`](../../../RELEASING.md)); the image digest is filled in once CI has built the image, the halt time once the halt has happened.

- **To run a node today:** the version in the last row, pinned — `ghcr.io/gnolang/gno/gnoland:<version>` or the binaries attached to its release page. Never a floating tag. Configuration: [`VALIDATOR.md`](./VALIDATOR.md).
- **Joining from genesis:** the node stops at every halt height below (the historical governance halts fire during replay too); restart it with the same binary if it satisfies that row's `halt_min_version`, otherwise with that row's version.
- **How a coordinated upgrade works** (halt height, `halt_min_version`, what the node refuses): [`gno.land/cmd/gnoland/UPGRADES.md`](../../../gno.land/cmd/gnoland/UPGRADES.md).
- A version runs the blocks from its halt height + 1 up to the next row's halt height. Halt time is the block time of the halt block.

<!-- BEGIN GENERATED (render-upgrades.sh) -->
| Version | Commit | Halt height | Halt time (UTC) | halt_min_version | GovDAO proposal | gnoland image digest | Consensus-relevant changes |
|---|---|---|---|---|---|---|---|
| [v1.2.0](https://github.com/gnolang/gno/releases/tag/v1.2.0) | [9c8eb132e](https://github.com/gnolang/gno/commit/9c8eb132e483d6fd324d92c193e629ad65a98a37) | genesis | 2026-09-12T15:00:00Z | — | — | `sha256:4b161a2b5d5f…` | launch |
| [v1.3.0](https://github.com/gnolang/gno/releases/tag/v1.3.0) | [31b6650a1](https://github.com/gnolang/gno/commit/31b6650a100d9baf14e7669f8f0df924f1f841e0) | 36300 | 2026-09-14T09:17:06Z | *(empty)* | [#0](https://gno.land/r/gov/dao:0) | `sha256:897a5c0f75e6…` | #6173 sign-doc fee rendering (`tm2!`), #6114 gpao |
| [v1.4.0](https://github.com/gnolang/gno/releases/tag/v1.4.0) | [00417a1be](https://github.com/gnolang/gno/commit/00417a1be97b9a311d9669ae7aa9585b277ee594) | 113000 | 2026-09-17T13:24:04Z | *(empty)* | [#5](https://gno.land/r/gov/dao:5) | `sha256:a41395a8cf39…` | #6193 crossing `cur` fixed binding, #6177 version gate |
| [v1.5.0](https://github.com/gnolang/gno/releases/tag/v1.5.0) | [e75fef82c](https://github.com/gnolang/gno/commit/e75fef82c02876a4df92ad6e325c5479b9532168) | 162200 | 2026-09-19T10:37:10Z | *(empty)* | [#7](https://gno.land/r/gov/dao:7) | `sha256:6ae099c2828c…` | #6211 origin-call anchoring, #5981 iota shadowing |
<!-- END GENERATED -->

Notes

- `v1.3.0`, `v1.4.0` and `v1.5.0` were tagged after the fact, on 2026-09-22. At each halt, validators pulled the `chain-mainnet` image current at the time (the `ran_as` digests in `upgrades.json`): the same code as the `:v1.x.0` images, reporting an unparseable version.
- `halt_min_version` was empty for those three halts. From `v1.6.0` on, every halt proposal names the version.
