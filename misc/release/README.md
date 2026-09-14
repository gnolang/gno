# Release tooling

Two scripts. The policy they enforce is in [`RELEASING.md`](../../RELEASING.md);
this page is what to type.

## `cut-release.sh`

Tags a release, after checking it is one.

```sh
# dry run — every check runs, nothing is pushed
misc/release/cut-release.sh v1.3.0

# tag a specific commit rather than the branch tip
misc/release/cut-release.sh v1.2.0 --commit 9c8eb132e

# a coordinated upgrade: also writes the GovDAO halt proposal, then pushes
misc/release/cut-release.sh v1.3.0 --halt-height 120000 --push
```

| Option | |
|---|---|
| `--chain <name>` | chain branch to cut from (default `mainnet` → `chain/mainnet`) |
| `--commit <ref>` | commit to tag (default: the branch tip) |
| `--previous <version>` | previous release, for the change summary (default: newest `v*` tag) |
| `--halt-height <H>` | emit a GovDAO halt proposal gating the restart on this release |
| `--push` | push the tag (otherwise it stops after creating it locally) |
| `--allow-dirty` | skip the clean-worktree check — local rehearsal only |

Without `--push` nothing leaves the machine, and the tag is trivially undone
with `git tag -d`.

Run `git fetch --tags origin` first. The script compares against `origin/master`
and `origin/chain/*`, and uses local tags to find the previous release.

### What it refuses, and why

- **A version that does not parse.** The node understands `vX.Y.Z` and betanet's
  `chain/gnolandX.Y`, and nothing else. Name an unparseable tag in a halt
  proposal and `halt_min_version` falls back to byte equality, which refuses the
  upgraded binary and leaves the chain unable to restart.
- **A tag that already exists**, locally or on origin. Tags are immutable.
- **Protocol-version constants that disagree** — see below.
- **A binary that does not report the tag.** The script builds `gnoland` the way
  CI does and asks it. `chain/mainnet`'s published binaries answer `develop`.

It *warns*, rather than refusing, when the commit is not on `master` (listing
what is missing, since a squash-merge shows up here even though its content
landed), and when a range tagged as a PATCH contains commits marked
`feat!:`/`BREAKING` — a consensus change is a MINOR, and getting that wrong is
how a coordinated upgrade goes out as a rolling one.

## `bump-protocol-version.sh`

Moves the six protocol-version constants together. This is **not** the release
version; see the table at the top of `RELEASING.md`.

```sh
misc/release/bump-protocol-version.sh --check   # do they agree?
misc/release/bump-protocol-version.sh v1.0.0    # move all six
```

It edits the constants, verifies each write, and runs
`TestProtocolVersionsAgree` before reporting success. It does not commit — it
prints the commit command.

A MAJOR bump here is a network partition: `VersionSet.CompatibleWith` refuses a
peer whose major differs, so nodes on either side cannot gossip at all. The
script says so, loudly, before doing it.
