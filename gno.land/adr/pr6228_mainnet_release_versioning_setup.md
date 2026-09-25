# ADR: One version per binary the network runs — mainnet releases, tags, images and the upgrade ledger

## Context

Mainnet (`gnoland-1`) launched on 2026-09-12 at `9c8eb132e`, tagged both
`chain/mainnet` and `v1.2.0`. In its first week it was upgraded three times
through GovDAO halts (heights 36300, 113000, 162200). Each upgrade was done by
merging `master` into `chain/mainnet`, waiting for the docker workflow, and
having validators pull the floating `ghcr.io/gnolang/gno/gnoland:chain-mainnet`
image. None of the three had a version tag, a release page or a fixed image
reference; the binaries reported `heads/chain/mainnet.<count>+<sha>`, which
`gno.land/pkg/gnoland.meetsMinVersion` does not parse, so `halt_min_version`
was left empty on every proposal and the restart gate from #6177 never
protected a restart.

Meanwhile `RELEASING.md` (#5275, #6177) described a different process:
`vMAJOR.MINOR.PATCH` tags cut with `misc/release/cut-release.sh`, no wholesale
merges of `master`, no hand-uploaded assets. Both ran side by side. The
`chain/mainnet` release page carried hand-built binaries from a post-launch
commit, `docker/metadata-action`'s default `latest=auto` had silently pointed
`gnoland:latest` at the launch-day build, and the goreleaser `master` images
reported `v0.0.0`, a version that parses.

## Decision

1. **`chain/mainnet` moves only when a release is cut, and every move is
   tagged.** Whether a release merges `master` or cherry-picks is a per-release
   choice written in the tag message. `RELEASING.md`'s "never merge `master`"
   becomes this rule.
2. **The three past upgrades get their versions after the fact:** `v1.3.0` @
   `31b6650a1`, `v1.4.0` @ `00417a1be`, `v1.5.0` @ `e75fef82c`, cut with
   `cut-release.sh --commit`, their binaries and images built by CI from the
   tags. The code tree at each commit is identical to the `master` commit it
   merged, which `cut-release.sh`'s preflight and a direct diff both confirmed.
3. **Two release pages, one job each.** `chain/mainnet` is the genesis record
   (`genesis.json`, `.gz`, checksums; no binaries). `vX.Y.Z` is a binary you
   can run: CI-built assets, the halt data, the image URLs, the changelog.
4. **Images and binaries are built from `vX.Y.Z` tags only.** A `chain/<name>`
   launch tag builds nothing. No floating tag for `gnoland`; the four CLI
   tools get `latest` only when the tag is the highest final release, and
   GitHub's "Latest" marker and `misc/install.sh` follow the same rule, so a
   backport never moves any of them backwards. `chain-mainnet` is retired
   after a transition pointing it at `v1.5.0`. Dispatch builds are labelled
   with the built commit's sha, not the event's.
5. **`misc/deployments/mainnet.gno.land/upgrades.json` is the ledger, and
   `gno.land/pkg/upgrades` owns its format.** One entry per version the
   network has run: kind (genesis or upgrade), commit, halt height and time,
   the proposal, `halt_min_version` (`null` when the proposal set none, never
   `""`), the image digest, what actually ran, and a per-platform map of
   binary downloads with checksums in the cosmovisor form. Facts that have not
   happened yet are `null` and render as *pending*. The package parses,
   validates (chain order, one genesis, versions the node can order, digest
   and checksum shapes) and renders `UPGRADES.md`; its tests validate every
   ledger in the tree, so a wrong or stale ledger fails the gno.land CI job. A
   JSON Schema next to the command documents the shape for third parties. A
   version runs blocks `halt_height + 1` through the next entry's
   `halt_height` — the contract a replaying supervisor relies on.
6. **Every release is rehearsed on the testnet as a release candidate on the
   same commit**, and every halt proposal names the version. The testnet runs
   mainnet's binaries, so it has no code branch: a `chain/<name>` launch tag on
   the commit of the version it launched on, a deployment folder on
   `chain/mainnet`, and a ledger of its own that names what it ran, release
   candidates included.
7. **`cut-release.sh` refuses rather than warns** where a warning is how the
   three untagged upgrades happened: a merge of `master` that touches
   consensus code needs `--allow-merge`, a breaking commit in a PATCH range is
   refused, and a final release without a ledger entry is refused.
8. Off-tag builds report an unparseable version everywhere, computed by one
   shell expression — `<branch>.<count>+<sha>`, `HEAD` on a detached checkout —
   in the Makefile and both release workflows; goreleaser reads it from the
   workflow instead of its dummy `{{ .Tag }}`, and
   `TestReleaseToolingMatchesTheParser` pins the three to that shape and
   asserts the node refuses it.

## Alternatives considered

- **Keep merging `master` and treat only tags as authoritative** (the branch
  as "master plus the deployment folder"). Rejected: the branch tip is what
  operators check out and what the floating image was built from; a branch
  that does not describe the network is a trap only documentation guards.
- **Strict cherry-pick only**, as `RELEASING.md` said. Deferred rather than
  rejected: it is the right default once `master` holds work mainnet must not
  ship yet; today every release has been "all of `master`", and the invariant
  that matters — moves only at release time, every move tagged — holds either
  way.
- **Tag only the current tip and record the two older upgrades by commit and
  image digest.** Rejected: a replay from genesis may need each version in
  sequence, and the `sha-*` image tags those rows would point at are not
  guaranteed to persist.
- **A floating `chain-mainnet` pointing at the latest release.** Rejected: a
  validator restarting for an unrelated reason between "release built" and
  "halt height" would pull the new binary, and `checkNodeStartupParams` would
  refuse to start it until the halt.
- **A single hand-maintained `UPGRADES.md` table.** Rejected after review: a
  supervisor needs a parsable source, and two hand-maintained copies drift.
- **The ledger tooling as a shell script per chain** (the PR's first shape:
  bash, jq and awk, rendering the table and checking it is not stale).
  Replaced after review: it validated nothing — a reordered or duplicated
  ledger rendered cleanly — printed `null` on the documented happy path, and
  would have been forked into every `misc/deployments/<chain>/`. One Go
  package with table-driven tests is what a supervisor can import.
- **A code branch per testnet** (`chain/onyx`). Rejected: a testnet that runs
  mainnet's binaries has nothing of its own but a genesis; a branch that never
  differs in code is a trap for whoever checks it out.

## Consequences

- Operators pin `:vX.Y.Z` and change it once per upgrade; `VALIDATOR.md`,
  `install.md` and `gnoland-networks.md` say so.
- Syncing from genesis stops at every past halt (the governance halts fire
  during replay); with `halt_min_version` set from `v1.6.0`, a restart between
  a proposal's execution and its halt height while already on the new binary
  is refused and needs `skip_upgrade_height`. Documented; a supervisor that
  reads `upgrades.json` is the tool that removes the manual step.
- `cut-release.sh` refuses a final release without a ledger entry (an rc only
  warns); the entry, the digest and the halt time are three separate edits, in
  that order, rendered with `go run ./misc/deployments/upgrades render`, and
  ported to `master`.
- Whether the `v*` tags carry a Go API promise is not decided here.
  `github.com/gnolang/gno` has no `/vN` module suffix, so a network-driven
  MAJOR would force a module path change; the maintainers are settling that
  separately and `RELEASING.md` will say which way.
- `v1.5.0` alone replays the whole chain from genesis: verified on 2026-09-22
  with the published genesis, the `VALIDATOR.md` configuration and the
  `ghcr.io/gnolang/gno/gnoland:v1.5.0` image — three stops at the historical
  halt heights, a plain restart each time, block ids and app hashes identical
  to the network's up to the tip. A future upgrade may break that property;
  the ledger is what a joiner falls back on when it does.
