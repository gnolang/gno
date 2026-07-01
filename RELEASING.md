# RELEASING.md

> For contributors, see [CONTRIBUTING.md](CONTRIBUTING.md). This document
> covers internal processes for those with merge access.

---

## Versioning & Branching Strategy

Gno uses a versioning scheme that reflects both the **client
state** and the **network state**. Releases are cut when they
provide meaningful value for users to install/upgrade or when the
network needs to coordinate an upgrade — not on every commit.

**Semver:** `vMAJOR.MINOR.PATCH`

| Component | Meaning |
|-----------|---------|
| **MAJOR** | New network (incompatible chain reset). `v1` = gnoland1, `v2` = gnoland2, etc. |
| **MINOR** | Chain upgrade on the current network (state migration, new module, protocol change). |
| **PATCH** | Backward-compatible change that does not require a coordinated chain upgrade. |

**Global tags** (`vMAJOR.MINOR.PATCH`) track the current mainnet
(`gnoland1`). They are the canonical versions users install.

**Branches and chain-specific tags:**

| Branch | Tags | Purpose |
|--------|------|---------|
| `master` | none | Continuous integration, `gno.land` staging environment |
| `chain/gnoland1` | `v1.x.x`, `chain/gnoland1.0`, `chain/gnoland1.1`, … | Mainnet (betanet initially) — coordinated upgrades only, never rebased |
| `chain/test12`, … | `chain/test12.0`, `chain/test12.1`, … | Testnets, same model as mainnet |

Chain-specific tags (`chain/<name>.<N>`) identify each
chain upgrade without a patch component — they map 1:1
to on-chain upgrade heights. Global `vX.Y.Z` tags are
created alongside the mainnet chain tag.

## Release Workflow

**Chain upgrade (minor bump, e.g. v1.0.0 → v1.1.0):** Any
change requiring a coordinated network upgrade — state
migrations, new built-in modules, consensus parameter changes.
Cherry-pick or merge the relevant commits from `master` into
`chain/gnoland1`, tag, and coordinate the upgrade height with
validators.

**Patch release (e.g. v1.1.0 → v1.1.1):** Backward-compatible
changes (client tools, SDK, server-side fixes that don't affect
consensus). Cherry-pick or merge from `master` into
`chain/gnoland1` and tag. No validator coordination needed.

**Hotfix:** Critical fix that cannot wait for the normal
`master`-first flow. Commit directly on the `chain/` branch,
tag, and deploy. Back-port to `master` afterwards. Chain
branches reflect the exact state of each network, which also
makes it possible to ship urgent fixes without pulling in
unrelated work from `master`.

**New network (major bump, e.g. v1.x.x → v2.0.0):** Chain
reset with incompatible genesis. Create `chain/gnoland2` from
the new genesis commit, tag as `v2.0.0`. The previous line
continues receiving patches on its branch (LTS-style).

**Tagging rules:** Tags are immutable — never move or delete.
All tags follow `vMAJOR.MINOR.PATCH`. Pre-release tags (e.g.
`v1.0.0-rc.1`) are allowed for mainnet/testnets only. Tag
messages should reference the upgrade proposal or change
summary.