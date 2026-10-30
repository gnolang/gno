# RELEASING.md

> For contributors, see [CONTRIBUTING.md](CONTRIBUTING.md). This document
> covers internal processes for those with merge access.

---

## Three versions, not one

A gno.land release moves three independent version surfaces. Confusing them is
the source of most release incidents, so they are named separately here.

| Surface | Where it lives | What it controls | When it moves |
|---|---|---|---|
| **Release version** | a git tag, compiled into `tm2/pkg/version.Version` via `-ldflags` | what `gnoland version` reports, and what a governance `halt_min_version` is compared against | every release |
| **Protocol version** | six constants under `tm2/`, see [Protocol versions](#protocol-versions) | the versionset exchanged in the p2p handshake, where peers whose MAJOR differs cannot connect; and the version in every block header, which nodes check for equality | rarely, never on its own, and always as a coordinated upgrade |
| **App version** | `baseApp.SetAppVersion` in `gno.land/pkg/gnoland/app.go` | the `app` entry of the versionset, persisted onto consensus state at the ABCI handshake; also in every block header, which nodes check for equality | rarely, and always as a coordinated upgrade |

Only the first is what people usually mean by "the version". The tooling in [`misc/release/`](misc/release) covers the first two: `cut-release.sh` for the release version, `promote-release.sh` for the release operators land on, `bump-protocol-version.sh` for the protocol constants. The app version is the literal `SetAppVersion("dev")` in `app.go` and is edited by hand.

## Versioning & branching

Releases are cut when they provide meaningful value for users to
install/upgrade, or when the network needs to coordinate an upgrade — not on
every commit.

**Semver:** `vMAJOR.MINOR.PATCH`

| Component | Meaning |
|-----------|---------|
| **MAJOR** | A new network: a chain reset with an incompatible genesis, where state does not carry over. |
| **MINOR** | A coordinated upgrade of the current network — consensus change, state migration, new module, protocol change. Validators must halt together. |
| **PATCH** | A backward-compatible change that needs no coordination. Operators switch binaries whenever convenient. |

**If a change alters the bytes nodes agree on, it is at least a MINOR**, whatever
it looks like otherwise. A one-line fix to the signing payload is a coordinated
upgrade; a thousand-line refactor of `gnoweb` is a patch. `cut-release.sh`
refuses a range tagged as a patch that contains commits marked
`feat!:`/`BREAKING`, but an unmarked consensus change passes it: the judgement
is yours.

### The `v` line is continuous across networks

`v1.0.0` and `v1.1.0` are betanet's — the same two commits as
`chain/gnoland1.0` and `chain/gnoland1.1`. They were deleted upstream for a
while and have been restored at their original commits.

The line continues rather than restarting: mainnet's launch is `v1.2.0`. A tag
name is never reused for different content, which is also why restoring a
deleted tag is only safe at the commit it originally pointed to — anyone still
holding the old ref then has an identical one.

### Branches and tags

| Branch | Tags | Purpose |
|--------|------|---------|
| `master` | none | Continuous integration; the development tree |
| `chain/mainnet` | `v1.x.x`, plus the `chain/mainnet` launch tag | Mainnet (`gnoland-1`) — releases, coordinated or rolling, plus the documentation and release-tooling commits between them; never rebased |
| `chain/gnoland1` | `chain/gnoland1.0`, `chain/gnoland1.1` | Retired. Frozen at what betanet ran; the tags keep their original names because they are compiled into those binaries. The branch is to be renamed `chain/betanet` — until it is, `--chain betanet` has no branch to resolve |
| *(none)* | `chain/<name>` | A testnet on the mainnet line, onyx first: it runs mainnet's `vX.Y.Z` binaries, so it has no branch. Its launch tag sits on `chain/mainnet` at the commit that adds its folder `misc/deployments/<name>.gno.land/`, on top of the version it launched on (for onyx, `v1.5.0`), and the tag's release carries its `genesis.json` |
| `chain/pearl`, `chain/test13`, … | `chain/<name>` | Earlier testnets, each on its own code line. Frozen |

Two tag shapes exist, and the node parses both
(`gno.land/pkg/gnoland.meetsMinVersion`):

- **`vMAJOR.MINOR.PATCH`** — the release shape. Ordered, so it can gate an
  upgrade. This is what new releases use.
- **`chain/gnolandMAJOR.MINOR`** — betanet's retired shape. Still parsed, because
  those two tags are the version strings inside binaries that once ran a chain.

A bare `chain/<name>` tag — `chain/mainnet`, `chain/pearl` — marks where a chain
launched and carries its `genesis.json` as a release asset. It is **not a
version**: it does not order, so it cannot be used as a `halt_min_version`. Name
a `vX.Y.Z` tag there instead.

**Tagging rules.** Tags are immutable: never move one, and never reuse a name.
New `v*` tags are annotated (`git tag -a`), so plain `git describe` finds them;
launch tags (`chain/<name>`) are lightweight. `v1.0.0` predates the rule and is
lightweight too, which is one reason every `describe` that stamps a binary's
version passes `--tags --match 'v*'` — the other being that a release commit
can also carry the chain's launch tag, which is not a version.
Pre-release tags (`v1.3.0-rc.1`) are allowed; they sort *below* the release they lead to. Every release is published as a GitHub pre-release, the final one included, and a final becomes GitHub's "Latest" release and `misc/install.sh`'s default (the highest `v*` release that is not a pre-release) only once it is promoted (see [Publishing early, promoting once the switch is set](#publishing-early-promoting-once-the-switch-is-set)).

### Chain branches are not master

`master` is the development tree. A `chain/*` branch is what a network actually
runs, and the two are not the same thing:

- Everything on a chain branch must also be on `master`. The chain must not run
  code the development tree has never seen. `cut-release.sh` checks this.
- `master` may be ahead, and usually is. That is fine and expected.
- **A chain branch moves only when a release is cut, and every move is
  tagged.** Whether a release merges `master` or cherry-picks a set of PRs is
  decided per release and stated in its release notes. A merge that touches
  consensus code is refused by `cut-release.sh` unless `--allow-merge` says the
  operator knows it ships everything `master` had. Between releases the branch
  tip is the latest tag, plus at most documentation or release-tooling commits
  that change no node code (the ledger, `VALIDATOR.md`, the release workflows,
  `cut-release.sh`). Operators still run the tag, never the tip: a build of
  anything but the tag reports a version no `halt_min_version` accepts.
- **The commits a release is cut from reach the chain branch through a pull request that is never merged.** They are a subset of `master` on an older base, a tree no CI run has seen as a whole. Push them to a branch `chain-upgrade/vX.Y.Z-rc.N`, open a pull request into the chain branch labelled `don't merge`, and once CI is green fast-forward the chain branch to it with a plain push; GitHub then marks the pull request merged. Never squash-merge it: that folds the cherry-picks into one commit, so the chain branch no longer shows which `master` commits it carries. A fix for a failed candidate goes the same way, on its own `chain-upgrade/vX.Y.Z-rc.N+1` branch and pull request.
- **Where this ends up.** Merging `master` is the natural default while mainnet
  tracks it closely. The day `master` carries work mainnet must not ship yet,
  the default flips to cherry-picking and the branch becomes a maintained
  release line in the Cosmos SDK / CometBFT sense: `release/vX.Y.x`, never
  merged from `master`, every change a labelled backport. Set that up the
  first time a hotfix must ship while `master` is mid-refactor — before, not
  during.

## Cutting a release

```sh
# dry run: every check, no push. The branch tip may carry deployment docs on
# top of the code, so every cut names the code commit
misc/release/cut-release.sh v1.6.0 --commit <code-commit>

# the cut itself. A final is cut once its last candidate has passed on the
# testnet, before mainnet's halt height is chosen: the tag records none
misc/release/cut-release.sh v1.6.0 --commit <code-commit> --push
```

The script checks that a tag would work as a release, and each check
corresponds to something that has gone wrong before. All but one are refusals:

1. **The version shape parses.** A tag the node cannot parse degrades
   `halt_min_version` to byte equality, which refuses the very binary the
   upgrade was cut for — and the chain cannot restart at all.
2. **The tag is free**, locally and on origin.
3. **The commit is on `master`**, or differs only under `misc/deployments/`.
   A *warning*, not a refusal: a squash-merged commit shows up as missing even
   though its content landed, so the operator has to read the list. Two
   refusals sit next to it: a **merge of `master` that touches consensus code**
   needs `--allow-merge`, and a **breaking commit in a PATCH range** is refused
   outright — a consensus change is a MINOR.
4. **The protocol-version constants agree** (see below).
5. **The built binary reports the tag.** Built the way CI builds it, then asked.
   Mainnet's launch-day binaries were hand-built without the `-ldflags`: they
   reported `develop`, satisfied no `halt_min_version`, and were not
   reproducible from the tag, until the `v1.2.0` release replaced them; this
   check is what catches that class of mistake before validators download it.
6. **The upgrade ledger passes `upgrades check`, and an entry it has for the version names the commit being tagged** (`misc/deployments/<chain>.gno.land/upgrades.json`). The check is the one CI runs: the ledger validates and its `UPGRADES.md` is current. The entry is written once the chain runs the release (see [Release pages and the upgrade ledger](#release-pages-and-the-upgrade-ledger)), so a version without one only warns. The script reads the ledger from the checkout it runs in, not from the commit it tags. A release candidate only warns, since the final release cut after it, on its commit, is checked the same way; a final release with a ledger that fails the check, or whose entry names another commit, is refused. A chain with no `upgrades.json` skips this check.

Pushing a `v*` tag triggers
[`release / chain-tag`](.github/workflows/release-chain-tag.yml), which builds
the four platforms, asserts that `gnoland`, `gnokey` and `gno` carry the tag on
the two platforms its runners can execute, and attaches them to the release.
It also triggers [`release / docker`](.github/workflows/release-docker.yml),
which publishes the matching images to `ghcr.io/gnolang/gno/*`. A
`chain/<name>` launch tag triggers nothing: its release is the genesis record,
created by hand with the `genesis.json` and its checksums attached, and no
binaries.

**Do not hand-upload release binaries** — an artifact built outside CI is not
reproducible from the tag and generally lacks the `-ldflags` that give it a
version at all. Note that `-ldflags` is only recorded in `go version -m` output
for builds without `-trimpath`, so its absence there is not on its own evidence
of a hand-built binary; the binary's own `version` output is.

**A published release is never replaced.** The `release / chain-tag` and
`release / docker` workflows run on `v*` tag pushes only, with no manual
trigger, and both refuse to replace what is already out: `release / chain-tag`
stops before attesting when the release is published, `release / docker`
before building when the registry already serves the version tag. Runs for one
tag never overlap: GitHub can start two for a single push, and the second waits
for the first, then meets those checks. These
builds are not reproducible, so a replacement would swap the binaries,
`CHECKSUMS.txt` and images the upgrade ledgers record for others under the
same version. The release is created as a draft and published only once every
asset is attached, so a run that failed before publishing can be re-run, and
the re-run publishes the draft. Images differ: *Re-run failed jobs* builds
only the images whose job failed before its push (with *Re-run all jobs*, the
images already pushed fail their check and the run ends red), and an image
pushed but left unsigned stays unsigned: the next candidate or patch
supersedes it. A fault in either workflow surfaces on the release candidate,
cut first on the same commit, and is fixed in the next candidate. A manual
trigger would also attest and sign under a branch's identity, which the
verify commands below refuse.

### Signing and provenance

The binaries `release / chain-tag` attaches, and the `CHECKSUMS.txt` beside
them, carry a build-provenance attestation (which workflow built them, from
which commit and tag), and the images `release / docker` pushes are signed
keylessly by digest, the multi-arch index and each platform image in it. Both
rely on the workflow's identity, not on a key anyone holds. `v1.5.0` and every
earlier release carry neither: their tagged commits predate these steps. To
verify what you are about to run:

```sh
gh attestation verify gnoland_linux_amd64 --repo gnolang/gno \
  --signer-workflow gnolang/gno/.github/workflows/release-chain-tag.yml \
  --source-ref refs/tags/v1.6.0

# cosign v3 or later: the signatures are in the Sigstore bundle format
cosign verify ghcr.io/gnolang/gno/gnoland:v1.6.0 \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  --certificate-identity https://github.com/gnolang/gno/.github/workflows/release-docker.yml@refs/tags/v1.6.0
```

The release binaries are not reproducible from the tag (they are built without `-trimpath`, so `gnoland` can find `GNOROOT`), which is why the attestation, the checksums and the ledger's recorded digests are the operator's anchors. `ci / upgrades-ledger-digests` checks, daily, what the ledgers record for each version: its commit against the commit its version's tag points to, each image digest against the registry, each `ran_as` image still resolving, each binary checksum against its release asset. It reports every difference and fails at the end; a value still pending the build is listed, not checked.

### Rehearse on the testnet first

The testnet runs the same binary as mainnet; only the genesis balances and the
on-chain transactions differ. Every mainnet release is therefore first cut as
a release candidate on the same commit
(`cut-release.sh v1.6.0-rc.1 --commit <code-commit> --push`), deployed to the
testnet the way mainnet will take it, and checked there. A patch is switched to
node by node; a coordinated upgrade goes through `set-halt` naming the rc, and
the rehearsal checks that the halt fires, an early start is refused and blocks
are produced after the restart. Either way, the nodes report the rc and a fresh
node syncs from genesis with it. A fix means a new commit, through its own
pull request, and `rc.2`. Only then is the final tag cut, on the same commit as
the last rc.

### Publishing early, promoting once the switch is set

A final release is cut and published as soon as its last candidate has passed on the testnet, days before mainnet switches to it, so operators can fetch and check its binaries and images ahead of the switch. Nothing a release publishes names a halt height: not the tag, the binaries or the images. The height goes only where it can be edited (the release notes, the release's tracking issue, the announcement to operators): marked as an estimate until the halt proposal has executed, then replaced by the height the chain holds.

Published that early, a final must not be what operators land on by default. `release / chain-tag` therefore publishes every release as a GitHub pre-release, the final one included, and `release / docker` pushes no `latest` image tag: GitHub's "Latest" release, `misc/install.sh`'s default and the CLI images' `latest` stay on the release mainnet runs. Once mainnet's switch is set (for a coordinated upgrade, once the halt proposal has executed; for a rolling one, on the agreed day), promote the final:

```sh
misc/release/promote-release.sh v1.6.0          # dry run: every check, nothing changed
misc/release/promote-release.sh v1.6.0 --push
```

It points `latest` at the release's images for the four CLI tools, then makes the release a full release and GitHub's "Latest". A promotion only moves forward from the highest full release, which is not always the highest tag: a patch to the running line is promoted while the next MINOR waits for its halt. The script refuses a release candidate, and a MINOR until mainnet's halt proposal names it (read from mainnet's `halt_min_version`), so a final mainnet skips stays a pre-release for good; when to promote a PATCH is the release manager's call.

### Container images

`release / docker` publishes `ghcr.io/gnolang/gno/<tool>:vX.Y.Z` for every `v*` tag, release candidates included, and never re-pushes a version tag. The four CLI tools — `gno`, `gnokey`, `gnodev`, `gnoweb` — also answer to `latest`, which `release / docker` never sets: `promote-release.sh` moves it when it makes a final a full release, the kind `misc/install.sh`'s default picks, and GitHub's "Latest" (see [Publishing early, promoting once the switch is set](#publishing-early-promoting-once-the-switch-is-set)). `gnoland` deliberately has no floating tag: a node image that moves by itself is incompatible with a coordinated upgrade, because a validator restarting for an unrelated reason after the release is published would pull the new binary: until the halt proposal executes nothing stops it running on the old chain, and once it has, the node refuses to start it until the halt. Operators pin the version the halt proposal names. Images are built from tags, not from branch pushes; `:master` images are development builds whose version does not parse. A release image is pushed with its version tag only, never a tag named after its commit: a final release shares its commit with its last candidate, so such a tag would move from the candidate's images to the final's. Some earlier images also answer to a `sha-<commit>` tag, which can name another build of that commit: pin a `vX.Y.Z` tag or a digest.

### Release pages and the upgrade ledger

Two kinds of release page, one job each:

- `chain/<name>` is the genesis record: `genesis.json`, its `.gz`, their
  checksums, and a description of the chain at birth. Frozen; no binaries.
- `vX.Y.Z` is a binary you can run: the CI-built binaries and `CHECKSUMS.txt`, for a coordinated upgrade the `halt_min_version` and, once the halt proposal has executed, the halt height that made it live, the eight image URLs, and the changelog. The workflow attaches the binaries with generated notes; the rest is written by hand.

`misc/deployments/<chain>.gno.land/upgrades.json` is the ledger: one entry per
version the network has run — the genesis, one per coordinated halt, and one
per rolling patch (`kind: rolling`: no halt, no floor, no proposal, same MINOR
as its predecessor) — with commit, halt height and time, `halt_min_version`,
GovDAO proposal, image digest, and a per-platform map of binary downloads with
checksums. The format is
[`misc/deployments/upgrades/upgrades.schema.json`](misc/deployments/upgrades/upgrades.schema.json);
the rules — chain order, one genesis, versions the node can order, `null` for a
fact that has not happened yet and never `""` — are enforced by
`gno.land/pkg/upgrades`, whose tests validate every ledger in the tree and
check that the `UPGRADES.md` next to it is current:

```sh
go run ./misc/deployments/upgrades render misc/deployments/mainnet.gno.land   # after editing the JSON
go run ./misc/deployments/upgrades check  misc/deployments/mainnet.gno.land   # what CI runs
```

Add the entry once the chain runs the release — after the halt for a coordinated upgrade, once validators run it for a rolling one — when everything it records is known: the halt height and the proposal once the proposal has executed, the halt time once the chain has halted, the image digest and binaries once CI has built the release. A complete entry renders no *(pending)* cell. Render, and port both files to `master`. A testnet's ledger names what it ran, release candidates included. Every halt proposal names the version being released — `set-halt.sh` refuses a value the node cannot parse and an empty one, which would leave the restart gate off, and a floor the version `REMOTE` runs already meets, which would refuse that version's restarts before the halt and gate nothing after it.

### Coordinated upgrades

For a MINOR bump, the release is only half of it: validators have to stop at the same height and come back on the new binary. The height is chosen once the release is published, and a GovDAO proposal sets it, with `halt_min_version` naming the release:

```sh
misc/deployments/mainnet.gno.land/govdao-exec.sh set-halt 120000 v1.6.0
```

Once the proposal has executed, promote the release (see [Publishing early, promoting once the switch is set](#publishing-early-promoting-once-the-switch-is-set)) and write the height into its release notes.

The mechanism, the failure modes, and what a halt looks like in the logs are
documented in [`gno.land/cmd/gnoland/UPGRADES.md`](gno.land/cmd/gnoland/UPGRADES.md).

### Hotfix

A critical fix that cannot wait for the `master`-first flow: commit it on its own `chain-upgrade/…` branch off the chain branch, open the CI-only pull request and fast-forward once green (see "Chain branches are not master"), then tag, deploy, promote (see [Publishing early, promoting once the switch is set](#publishing-early-promoting-once-the-switch-is-set)), and **back-port to `master` immediately**. The back-port is the part that gets skipped, and it is the part that matters — until it lands, the chain is running code nobody can review on `master`.

## Protocol versions

The protocol version is one value living in six files, split that way to avoid
import cycles rather than because the parts may differ:

| File | Constant |
|---|---|
| `tm2/pkg/crypto/version.go` | `Version` |
| `tm2/pkg/bft/abci/version/version.go` | `Version` |
| `tm2/pkg/bft/blockchain/version/version.go` | `Version` |
| `tm2/pkg/bft/types/version/version.go` | `BlockVersion` |
| `tm2/pkg/p2p/version/version.go` | `Version` |
| `tm2/pkg/bft/version/version.go` | `Version` |

Three `init()` checks chain them together (`types` checks `crypto`,
`blockchain` checks `types`, and `bft` checks `blockchain`, `abci` and `p2p`),
so a bump that misses one panics every node at startup. Move them together:

```sh
misc/release/bump-protocol-version.sh --check   # are they consistent?
misc/release/bump-protocol-version.sh v1.0.0    # move all six
```

`TestProtocolVersionsAgree` in `tm2/pkg/bft/version/version_test.go` is what
fails in CI if they drift.

**Any bump here splits the network, not only a MAJOR one.** `BlockVersion` is
written into every block header, and a node refuses a block whose
`Header.Version` differs from the binary's own constant
(`tm2/pkg/bft/state/validation.go`), in consensus and on fast sync alike: nodes
on either side of a bump reject each other's blocks, and a bumped binary cannot
sync blocks produced before it. Nothing gates that check by height, so a bump
needs such code first, then a coordinated upgrade with every validator
switching at the same block, rehearsed on a testnet. A MAJOR bump also stops
old and new nodes from peering: `VersionSet.CompatibleWith` refuses a peer
whose major differs (`TestVersionSetCompatibleWith` covers it).

These constants have never been bumped: they still read `v1.0.0-rc.0`,
unchanged since before the 2023 reorganisation that created `tm2/`. Mainnet
launched on it.

## New network

A chain reset with an incompatible genesis. Create `chain/<name>` from the new
genesis commit, tag `vMAJOR.0.0`, and cut a `chain/<name>` launch tag carrying
`genesis.json`. The previous line keeps receiving patches on its own branch,
LTS-style, until it is retired.
