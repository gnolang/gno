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
| **Protocol version** | six constants under `tm2/`, see [Protocol versions](#protocol-versions) | the versionset exchanged in the p2p handshake; peers whose MAJOR differs cannot connect | rarely, and never on its own |
| **App version** | `baseApp.SetAppVersion` in `gno.land/pkg/gnoland/app.go` | the `app` entry of the versionset, persisted onto consensus state at the ABCI handshake | rarely |

Only the first is what people usually mean by "the version". The tooling in
[`misc/release/`](misc/release) covers the first two: `cut-release.sh` for the
release version, `bump-protocol-version.sh` for the protocol constants. The app
version is the literal `SetAppVersion("dev")` in `app.go` and is edited by hand.

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
upgrade; a thousand-line refactor of `gnoweb` is a patch. `cut-release.sh` warns
when a range tagged as a patch contains commits marked `feat!:`/`BREAKING`, but
the judgement is yours.

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
| `chain/mainnet` | `v1.x.x`, plus the `chain/mainnet` launch tag | Mainnet (`gnoland-1`) — coordinated upgrades only, never rebased |
| `chain/gnoland1` | `chain/gnoland1.0`, `chain/gnoland1.1` | Retired. Frozen at what betanet ran; the tags keep their original names because they are compiled into those binaries. The branch is to be renamed `chain/betanet` — until it is, `--chain betanet` has no branch to resolve |
| `chain/pearl`, `chain/test13`, … | `chain/<name>` | Testnets. One launch tag each |

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
New tags are annotated (`git tag -a`), so plain `git describe` finds them;
`v1.0.0` predates the rule and is lightweight, which is one reason every
`describe` in the tree passes `--tags --match 'v*'` — the other being that a
release commit also carries the chain's launch tag, which is not a version.
Pre-release tags (`v1.3.0-rc.1`) are allowed; they sort *below* the release they
lead to, and are published as GitHub pre-releases so they do not become the
"Latest" release operators and `misc/install.sh` land on.

### Chain branches are not master

`master` is the development tree. A `chain/*` branch is what a network actually
runs, and the two are not the same thing:

- Everything on a chain branch must also be on `master`. The chain must not run
  code the development tree has never seen. `cut-release.sh` checks this.
- `master` may be ahead, and usually is. That is fine and expected.
- **Do not merge `master` wholesale into a chain branch.** It pulls in
  everything unreleased, including consensus changes the network has not agreed
  to, and the branch stops describing the network. Cherry-pick what the upgrade
  actually contains.

## Cutting a release

```sh
# dry run: every check, no push
misc/release/cut-release.sh v1.3.0

# a coordinated upgrade: also prints the GovDAO halt proposal to go with it
misc/release/cut-release.sh v1.3.0 --halt-height 120000 --push
```

The script checks that a tag would work as a release, and each check
corresponds to something that has gone wrong before. All but one are refusals:

1. **The version shape parses.** A tag the node cannot parse degrades
   `halt_min_version` to byte equality, which refuses the very binary the
   upgrade was cut for — and the chain cannot restart at all.
2. **The tag is free**, locally and on origin.
3. **The commit is on `master`**, or differs only under `misc/deployments/`.
   A *warning*, not a refusal: a squash-merged commit shows up as missing even
   though its content landed, so the operator has to read the list.
4. **The protocol-version constants agree** (see below).
5. **The built binary reports the tag.** Built the way CI builds it, then asked.
   `chain/mainnet`'s published binaries report `develop`, satisfy no
   `halt_min_version`, and are not reproducible from the tag; this check is what
   catches that class of mistake before validators download it.

Pushing the tag triggers
[`release / chain-tag`](.github/workflows/release-chain-tag.yml), which builds
the four platforms, asserts each native binary carries the tag, and attaches
them to the release. It also triggers
[`release / docker`](.github/workflows/release-docker.yml), which publishes the
matching images to `ghcr.io/gnolang/gno/*` — the place every
`misc/deployments/*/VALIDATOR.md` points operators at.

**Do not hand-upload release binaries** — an artifact built outside CI is not
reproducible from the tag and generally lacks the `-ldflags` that give it a
version at all. Note that `-ldflags` is only recorded in `go version -m` output
for builds without `-trimpath`, so its absence there is not on its own evidence
of a hand-built binary; the binary's own `version` output is.

### Coordinated upgrades

For a MINOR bump, the release is only half of it: validators have to stop at the
same height and come back on the new binary. `--halt-height` prints the GovDAO
proposal to go with the tag, with `halt_min_version` set to the tag being cut,
and the command that creates it:

```sh
misc/deployments/mainnet.gno.land/govdao set-halt 120000 v1.3.0
```

The mechanism, the failure modes, and what a halt looks like in the logs are
documented in [`gno.land/cmd/gnoland/UPGRADES.md`](gno.land/cmd/gnoland/UPGRADES.md).

### Hotfix

A critical fix that cannot wait for the `master`-first flow: commit on the
`chain/` branch, tag, deploy, and **back-port to `master` immediately**. The
back-port is the part that gets skipped, and it is the part that matters — until
it lands, the chain is running code nobody can review on `master`.

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

Each guards the next in an `init()`, so a bump that misses one panics every node
at startup. Move them together:

```sh
misc/release/bump-protocol-version.sh --check   # are they consistent?
misc/release/bump-protocol-version.sh v1.0.0    # move all six
```

`TestProtocolVersionsAgree` in `tm2/pkg/bft/version/version_test.go` is what
fails in CI if they drift.

**A MAJOR bump here partitions the network.** `VersionSet.CompatibleWith`
compares major.minor and refuses a peer whose major differs, so old and new
nodes cannot gossip — independently of any halt height. It must ride a
coordinated upgrade, with every validator switching at the same block.
`TestVersionSetCompatibleWith` covers the refusal and the negotiated minor;
treat a MAJOR bump as needing a rehearsal on a testnet anyway.

These constants have never been bumped: they still read `v1.0.0-rc.0`, the value
they were given in 2023. Mainnet launched on it.

## New network

A chain reset with an incompatible genesis. Create `chain/<name>` from the new
genesis commit, tag `vMAJOR.0.0`, and cut a `chain/<name>` launch tag carrying
`genesis.json`. The previous line keeps receiving patches on its own branch,
LTS-style, until it is retired.
