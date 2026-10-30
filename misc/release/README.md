# Release tooling

Three scripts. The policy they enforce is in [`RELEASING.md`](../../RELEASING.md); this page is what to type.

## `cut-release.sh`

Tags a release, after checking it is one.

```sh
# dry run — every check runs, nothing is pushed. The branch tip may carry
# deployment docs on top of the code: name the code commit
misc/release/cut-release.sh v1.6.0 --commit <code-commit>

# the cut itself. A final is cut once its last candidate has passed on the
# testnet, before mainnet's halt height is chosen: the tag records none
misc/release/cut-release.sh v1.6.0 --commit <code-commit> --push
```

| Option | |
|---|---|
| `--chain <name>` | chain branch to cut from (default `mainnet` → `chain/mainnet`) |
| `--commit <ref>` | commit to tag (default: the branch tip) |
| `--previous <version>` | previous release, for the change summary (default: the newest final `v*` tag reachable from the commit) |
| `--halt-height <H>` | print the GovDAO halt proposal gating the restart on this release, and the `govdao-exec.sh set-halt` command that creates it (mechanics: [`UPGRADES.md`](../../gno.land/cmd/gnoland/UPGRADES.md)). For a release candidate the command names `misc/deployments/<chain>.gno.land/`, the testnet's folder, since candidates are cut from `chain/mainnet` too. Only before the tag exists, which for a release published early is before the height is chosen: once it is chosen, run `govdao-exec.sh set-halt <H> <version>` from the network's folder directly |
| `--push` | create and push the tag (without it, nothing is created at all) |
| `--allow-dirty` | skip the clean-worktree check — local rehearsal only |
| `--allow-merge` | cut a range that merges `master` and touches consensus code, which is otherwise refused |

Without `--push` nothing is created — not even locally. A local tag left behind
by a rehearsal would make the next run refuse the version as already released,
so a dry run prints the tag it would make and stops.

Run `git fetch --tags origin` first. The script compares against `origin/master`
and `origin/chain/*`, and uses local tags to find the previous release.

### What it refuses, and why

- **A version that does not parse.** The script accepts `vX.Y.Z[-pre]` only;
  the node also parses betanet's retired `chain/gnolandX.Y`, which no new
  release uses. Name an unparseable tag in a halt proposal and
  `halt_min_version` falls back to byte equality, which refuses the upgraded
  binary and leaves the chain unable to restart.
- **A tag that already exists**, locally or on origin. Tags are immutable.
- **Protocol-version constants that disagree** — see below. Read from a worktree of the commit being tagged, not from your checkout: that is the code commit `--commit` names, below any deployment docs you have checked out, and drift that exists only there is drift validators run.
- **A binary that does not report the tag.** The script builds `gnoland` the way
  CI does and asks it. Mainnet's launch-day binaries answered `develop`.
- **A final release whose upgrade ledger fails `upgrades check`**, the check CI runs (`misc/deployments/<chain>.gno.land/upgrades.json` and its `UPGRADES.md`), or whose entry names another commit than the one being tagged. A release candidate only warns.
- **A commit marked breaking in a PATCH range** (`feat!:`/`BREAKING`). A
  consensus change is a MINOR, and getting that wrong is how a coordinated
  upgrade goes out as a rolling one.
- **A range that merges `master` and touches consensus code**, unless
  `--allow-merge` says the operator knows it ships everything `master` had.

It *warns*, rather than refusing, when the commit is not on `master` (listing what is missing, since a squash-merge shows up here even though its content landed), and when the ledger has no entry for the version yet: the entry is written once the chain runs the release, when everything it records, the halt time included, is known.

## `promote-release.sh`

Makes a published final release the one operators land on: points `latest` at its `gno`, `gnokey`, `gnodev` and `gnoweb` images, byte for byte, then makes it a full release, the kind `misc/install.sh` installs by default (the highest `v*` release that is not a pre-release), and GitHub's "Latest". Run it once mainnet's switch is set: for a coordinated upgrade, once the halt proposal has executed; for a rolling one, on the agreed day.

```sh
misc/release/promote-release.sh v1.7.0          # dry run: every check, nothing changed
misc/release/promote-release.sh v1.7.0 --push
```

It needs `curl`, `jq`, `sha256sum` or `shasum`, and `gh` logged in with the `write:packages` scope. It refuses a release candidate, a release still a draft (its `release / chain-tag` run has not finished), a release older than the highest full release (a promotion only moves forward), a MINOR whose halt mainnet has not scheduled (mainnet's `halt_min_version` must name it; a PATCH to the highest full release's line, or the release mainnet runs, needs none), a MAJOR (a new network, which mainnet never schedules), and an image it cannot read, or whose bytes or media type are not what release / docker pushed: every image is read and checked before anything is written. A missing write permission shows only at the first write, where ghcr.io checks it. A run stopped halfway leaves the release a pre-release, though some images' `latest` may have moved; run it again to finish.

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

Any bump here splits the network: the version is in every block header, and
nodes refuse blocks whose version differs from theirs, so nodes on either side
reject each other's blocks and a bumped binary cannot sync older ones. A MAJOR
bump also stops them peering. The script says so, loudly, before doing it
(RELEASING.md, "Protocol versions").
