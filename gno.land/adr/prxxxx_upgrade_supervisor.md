# ADR: An upgrade supervisor driven by the ledger

## Context

A gno.land node runs one release per stretch of blocks. The upgrade ledger (`misc/deployments/<chain>/upgrades.json`, rules in `gno.land/pkg/upgrades`) records which: a version runs the blocks from its own halt height + 1 up to and including its successor's halt height (`BlockRanges`). Today an operator switches releases by hand at every governance halt, in two situations.

- **Joining from genesis.** A node replaying history stops at every past halt and must be restarted on the right release. On gnoland-1 this is no longer optional: v1.8.0 cannot replay the blocks before its halt (`wrong Block.Header.AppHash`), so a joiner runs v1.7.1 up to 1651500 and then v1.8.0 (`misc/deployments/mainnet.gno.land/UPGRADES.md`). The first release carrying #6208 will add another leg.
- **A live halt.** A GovDAO proposal sets `node:p:halt_height` and `node:p:halt_min_version`; there is no minimum notice (the 1183000 halt for v1.7.0 executed about 15 minutes before it fired). Each operator must notice the halt, stop the idle process and start the new release.

What a node does at a halt (`gno.land/cmd/gnoland/UPGRADES.md`, `gno.land/pkg/gnoland/app.go:1168-1182`, `tm2/pkg/sdk/baseapp.go:661-664`): the EndBlocker of block H arms the halt, whatever the binary's version; block H is committed; the next block's `BeginBlock` panics. A node at the tip catches the panic in consensus and stays up, RPC included, with no new blocks. A node catching up in block sync has nothing to catch it: the process exits with code 2, the Go runtime's code for an unrecovered panic. On restart, `checkNodeStartupParams` refuses a binary below the floor once H is committed, and a binary at or above the floor before H, unless `skip_upgrade_height` equals H.

Issue #6237 asks for "the cosmovisor equivalent, driven by the ledger", in the monorepo. Two other pieces bear on the same problem:

- #6528 (omarsy) proposes a post-halt handoff: once block H is committed, the live node writes a versioned file a supervisor can read, plus an opt-in parent mode. It leaves historical replay to #6237.
- #6239 proposes recording upgrades on chain, so that `upgrades.json` becomes a mirror. Since #6381 a ledger entry is written only after the chain runs the release, so before a future halt the only machine-readable source is the vote's `halt_min_version`.

## Decision

The design has three pieces. This change implements the first half of the first.

### 1. The supervisor (#6237)

A command, `gnosupervisor`, that runs `gnoland start` as its child, on the release the ledger assigns to the node's next block.

**This change: historical replay.** Every iteration of its loop:

1. Reads the node's last committed height from the stopped node's `state` database, opened read-only (`sm.State.LastBlockHeight`, as `contribs/gnobr` reads it offline). The supervisor resumes from the data directory, never from its own memory.
2. Resolves the leg for height + 1 through `upgrades.BlockRanges`: the range's version, or the last rolling patch released into it (`Range.Rolling`), which by definition changes no consensus code. The range rule is never copied.
3. Prepares that release's artifact (below), then starts preparing the next leg's in the background, so the switch does not wait on the network and a registry that is down does not delay the node now.
4. Starts `gnoland start --data-dir <dir> --genesis <file> --gnoroot-dir <artifact>/gnoroot [operator flags]`, in its own process group.
5. While the child runs, polls its RPC `/status`. If the node reports its range's last block for 2 minutes without moving, it has left block sync for consensus before the halt (block sync can switch early, `tm2/pkg/bft/blockchain/pool.go:183-185`) and caught the halt panic (`tm2/pkg/bft/consensus/state.go:651-665`): the supervisor stops it. Only the ledger's halt heights are watched, so stopping there is what the ledger asks for anyway. On every range, each time the node reports a new height, the supervisor reads the chain's `node:p:halt_height` over RPC (`abci_query`), and once more before it stops the node: if the node sits at a halt the ledger does not record, it records the stop as below (`gnosupervisor.halt`) before anyone restarts anything. A node that commits no block for 10 minutes is reported as an error (not stopped).
6. When the child exits, reads the committed height again:
   - equal to the range's end: the ledger's halt. The next iteration resolves the next release. Code 2 is expected; another code at exactly that height is logged and leads to the same step.
   - past the range's end: the node ran blocks the ledger assigns to another release. The supervisor stops; it never guesses.
   - code 2 after progress, anywhere else: a halt the ledger does not record. It may be an upgrade missing from this copy of the ledger (entries are written after the halt), and with no version floor nothing in the node would refuse the old release on the next release's blocks. The supervisor stops and says so, and writes the height to `gnosupervisor.halt` in the data directory, so that a restarted supervisor (by hand, or by a restart policy) refuses too. It goes on once the ledger records a halt there, or once the operator confirms a pause with `--continue-after-halt <H>`.
   - anything else: an unexpected exit. It restarts the same release with a doubling backoff, and gives up after `--max-restarts` exits in a row with no block committed in between.

Before the loop it validates the ledger with `Ledger.Validate`, checks that the genesis file is the one the ledger records (`chain_id` and `genesis_sha256`) and that the node's state, if any, is for the same chain (gnoland reads the genesis from its database once it has one), and refuses a node configuration with a non-zero `halt_height` or `skip_upgrade_height` (see below).

**Next: a live halt.** At the tip the process does not exit, so the loop above never sees the halt. The supervisor will learn of it from #6528's handoff once that lands (until then, by polling `/status` against `params/node:p:halt_height`), stop the idle child, and start the release named by the vote's `halt_min_version`, prepared and verified ahead of H. An empty floor is a pause: the node stays stopped and the supervisor says so. That release is not in the ledger yet (#6381), which is why the source of truth for a future halt is the open question below.

### 2. The live signal (#6528)

The handoff file replaces RPC polling: it is written once H is committed, by the node that knows it, and it is versioned. The supervisor consumes it; it does not need the opt-in parent command, since the supervisor is already the parent.

### 3. The artifact in chain state (#6239)

Once a halt proposal carries the image digest (and binary checksums), the supervisor reads the exact artifact for a future halt from chain state before H, with the same verification as for the ledger. `upgrades.json` then mirrors chain state, and the replay and live paths verify against the same kind of record.

### State machine

```text
            ┌──────────────────────────────────────────────┐
            v                                              │
  read committed height ──> resolve leg ──> prepare ──> run child
                                │              │           │
                         invalid ledger   bad artifact     │ child exits
                                │              │           v
                                └──> stop <────┘    height == range end ──> (next leg)
                                       ^            height >  range end ──> stop
                                       │            code 2, other height ──> stop
                                       │            otherwise ──> backoff ──┘
                                       └── restarts exhausted
  Child idle at the range end (RPC) for 2 min: stop it, then as "height == range end".
  Child idle at a halt height the ledger does not record: record it, leave the node up.
  SIGTERM/SIGINT at any point: forward SIGTERM, wait --stop-timeout, then SIGKILL, exit.
```

Live mode (next step) adds one transition: child running, halt signalled (handoff file, or RPC height == `halt_height`) → stop child → resolve the vote's floor → run.

### Artifacts and trust

The ledger records two artifacts per entry: `image.digest` and per-platform `binaries` with `?checksum=sha256:`. The supervisor uses the image:

- The image's `/usr/bin/gnoland` is static (`CGO_ENABLED=0`, `Dockerfile`), so it runs on any linux of the right architecture. The release binaries of v1.5.0 and v1.7.0 require glibc 2.38 (fixed for later releases by #6450).
- The image carries `/gnoroot` from the same commit: the stdlibs a replay needs, without a separate checkout at the tag (#6240).
- It is pulled by digest over the registry HTTP API, with no container runtime and no new dependency: an in-tree client of about 700 lines with comments (`net/http`, `archive/tar`, `compress/gzip`). The index's sha256 must equal the ledger digest; the platform manifest and each layer must match the descriptor naming it (digest and size), and media types are read from the bodies the digests cover, never from headers. Redirects may not leave https; a token realm must be https.
- Only `usr/bin/gnoland` and `gnoroot/**` are extracted, whiteouts are applied, and only regular files, directories and hard links to extracted files are written. Symbolic links are refused: no release's tree has one (`git ls-tree` of v1.2.0, v1.8.0 and master), and a chain of them is the one way an entry could point out of the extraction directory. A layer may write at most 8 GiB.
- An artifact is extracted into a temporary directory and renamed into the cache once complete, with a record holding a hash of every extracted file. A cached artifact is re-hashed before each use and pulled again if it does not match. An interrupted pull leaves nothing behind. Each artifact has a lock file, so supervisors sharing a cache never pull or remove the same artifact at once. The record lives next to the files it hashes, so it guards against damage, not against someone who can write to the cache: the supervisor refuses a cache directory not owned by its user or writable by anyone else.

Trust is the ledger's: the supervisor runs what the ledger in hand pins. It does not fetch or refresh the ledger; the operator supplies it, for example from a pinned checkout. Images are also signed with cosign (`release-docker.yml`), which the live path, where no ledger entry exists yet, should check.

The static binary has no `lmdbdb` or `mdbxdb` (both need cgo). The supervisor reads the committed height of `pebbledb` nodes only, the default, and refuses other backends; release binaries with checksums can follow for nodes that need another backend.

### `halt_min_version` and `skip_upgrade_height`

The replay always starts the release the ledger assigns to the next block, which is exactly what the startup gates accept, so it never needs `skip_upgrade_height`. Left non-zero, that setting disables both gates at that height, so the supervisor refuses to start with it set. A `halt_height` in `config.toml` stops the node at a height the ledger knows nothing about, so it is refused too. A halt the ledger does not record stops a catching-up node like any halt. The supervisor cannot tell a pause from an upgrade missing from its copy of the ledger: mainnet's v1.3.0, v1.4.0 and v1.5.0 halts had no floor, so the old release would have started again and run the new release's blocks. It stops instead, and the operator either updates the ledger or confirms a pause with `--continue-after-halt`.

The state record holds H when a catching-up node exits at a halt. Released binaries save it with a synced write every block (`tm2/pkg/bft/state/store.go` at v1.8.0). On master, block sync defers its writes and takes a flush point at the halt height (`tm2/adr/pr6334_deferred_block_sync_writes.md`, not in any release yet). There, a stop between flush points drops the writes not yet flushed; the state record stays at the last flush point and the supervisor resumes from there, as gnoland does. One case is left to the operator: a crash inside the flush point at the halt height, after the state is written and before the application is, leaves the application below H, and the next release refuses to start; that ADR's remedy (start the previous release once) is not automated, and the supervisor stops after its retries.

The height is read by decoding the whole `sm.State` with this tree's amino types. The record is unchanged since v1.2.0, but a future change to it, or to a type inside it, must stay decodable by the supervisor.

### Packaging

A container image whose process is the supervisor, with `gnoland` as its child: no Docker socket, no `docker` group, and the data and cache directories as volumes. Releases switch inside the container, which is never recreated for a gnoland upgrade. That image and a compose example are a follow-up; this change ships the command.

### Double signing

- The supervisor takes an exclusive lock on the data directory. The child inherits the locked file descriptor, so a supervisor killed outright does not free the directory while its node is still shutting down.
- On Linux the child gets `SIGTERM` if the supervisor dies without forwarding a signal.
- Until the live-halt switch exists, a validator under the supervisor stops at every future halt like any node at the tip, with an error in the log after 10 minutes; replaying and validating stay separate, as below.
- The stop timeout (`--stop-timeout`, 60 seconds) must be shorter than the container's grace period (`stop_grace_period`; Docker's default is 10 seconds), or the container runtime kills gnoland before it has written its last blocks.
- A local lock does not protect against a second machine or a second compose project running the same key, nor against a restart policy (`restart: unless-stopped`) restarting a container whose node is still up elsewhere. `VALIDATOR.md` asks for no restart policy for that reason; under the supervisor the restart policy belongs to the supervisor, and the container's must stay off. Validators should replay on a non-signing node and move the key separately, as today.
- The registry must be reachable when an artifact is not cached yet. The supervisor prepares the next release before the halt that needs it.

### Location

`contribs/gnosupervisor`, with its own `go.mod` like the other `contribs` tools. #6237 suggests `misc/deployments/`, but CI tests and lints every `contribs/*` module automatically (`ci-dir-contribs.yml`, which also runs on `gno.land/` and `tm2/` changes), while `misc/` Go programs are a fixed list in `ci-dir-misc.yml`, and the root module's packages under `misc/deployments/` are not tested by `ci-dir-gnoland.yml`. Moving it is a rename if maintainers prefer otherwise.

## Alternatives considered

- **Release binaries instead of images.** They carry a checksum in the ledger too, but v1.5.0 and v1.7.0 need glibc 2.38, and every release needs a GNOROOT checkout at its tag. Kept as a second backend.
- **Drive Docker (as GnoVisor does) instead of extracting the image.** It needs the Docker socket, which is root-equivalent, and cannot drive a compose service with sentry and signer siblings.
- **Make the stop deterministic with `halt_height` in `config.toml`.** gnoland applies it through the same `SetHaltHeight` (`gno.land/cmd/gnoland/start.go:287-292`), so setting it to each range's end would stop the old release at its last block even if the ledger and the chain disagreed, instead of noticing it one block late. It needs the supervisor to own the node's config (or a `gnoland start` flag), so it is left as a follow-up; until then the startup gates and the past-the-range check stand in for it.
- **Decide the halt from the exit code alone.** Code 2 is any unrecovered panic. The committed height decides; the code is only logged.
- **Read the height from the block store.** In block sync, block H+1 is saved before it is applied (`tm2/pkg/bft/blockchain/reactor.go:531-536`), so the block store can be ahead of the state at a halt. The consensus state record is what the next start's handshake resumes from.
- **A third-party OCI library.** A library such as go-containerregistry would add a large dependency tree to code that only needs anonymous pulls by digest.
- **A single replay release across many halts** (the documented joining path runs v1.7.1 from genesis). That knowledge is not in the ledger, which assigns each range its own release. If it should be, the ledger is the place to record it, not the supervisor.

## Consequences

- An operator joins a chain from genesis with one command: `gnosupervisor run --ledger upgrades.json --genesis genesis.json --data-dir <dir> -- <gnoland flags>`.
- Replaying gnoland-1 range by range, each on its own release, has not been verified yet. The documented joining path is v1.7.1 from genesis to 1651500, then v1.8.0. The images tagged v1.3.0, v1.4.0 and v1.5.0 were built after the fact from the code that ran (`ran_as` records the `chain-mainnet` images that actually did), and whether v1.2.0 block-syncs from today's peers is unknown. A full gnoland-1 replay is the acceptance test before recommending the supervisor for mainnet; open question 3 depends on it.
- The supervisor depends on the ledger format (`gno.land/pkg/upgrades`) and on node facts: where the consensus state lives, how a catching-up node stops at a halt, and the startup gates. The unit tests encode them in a fake gnoland. The real-chain test, which is opt-in and not run in CI, checks them on two releases built from this tree.
- Out of scope here: the live-halt switch (described above), #6528 and #6239 themselves, self-update of the supervisor, metrics, ledger refresh, and upgrades that need a new genesis (`gnogenesis fork`), which stay manual.

## Open questions

1. Is the vote's `halt_min_version` the source of truth for a live switch until #6239 lands, with cosign verification of the image it names?
2. `misc/deployments/supervisor` as #6237 says, or `contribs/` for CI coverage?
3. Should the ledger name a replay release per range when a later release replays it identically (v1.7.1 today), or should every range replay its own release? Today's joining path suggests the first, and the images of v1.3.0 to v1.5.0 (tagged after the fact) are an argument for it.
4. For #6239: the digest in `r/sys/params` next to `halt_min_version`, or in the proposal body? A last-minute fix would then need a new vote, or a rule such as "this digest or a rolling patch of the same version".
5. Any interaction with #6224's in-place migrations that the supervisor should plan for?
