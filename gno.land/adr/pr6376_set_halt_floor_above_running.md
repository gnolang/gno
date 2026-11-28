# ADR: set-halt.sh refuses a halt floor the running binary already meets

## Context

A governance halt sets `halt_height` and `halt_min_version`, and the node checks its binary against the floor at every startup (`checkNodeStartupParams` in `gno.land/pkg/gnoland/node_params.go`): after the halt a binary below the floor is refused (check 1), before it a binary that already meets the floor is refused (check 2). Both are meant for a floor equal to the version operators switch to. A floor the running binary already meets turns them around: every restart of that binary before the halt is refused, with a message telling the operator to "use the previous binary", the very binary being refused, and after the halt the same binary passes check 1 and resumes the chain, so nothing was gated.

A rehearsal on a two-validator scratch chain (v1.5.1, 2026-10-16) showed both halves: with a `v1.5.1` floor, one validator restarted inside the window was refused, the chain stopped short of the halt for lack of quorum until that validator restarted with `skip_upgrade_height`, and after the halt the unchanged v1.5.1 binary resumed. The release flow cannot produce this floor (`misc/release/cut-release.sh` prints the `set-halt` command for the tag it cuts), but `misc/govdao-scripts/set-halt.sh` takes any well-formed version from the person running it.

## Decision

1. Before it builds the proposal, `set-halt.sh` reads `build_version` from the `/status` of the RPC it broadcasts to (`REMOTE`) and refuses a `MIN_VERSION` that version already meets. Cancelling (height `0`) reads nothing.
2. The comparison mirrors `parseReleaseVersion` and `meetsMinVersion` in shell (betanet's `chain/gnolandX.Y` shape, build metadata ignored but checked, semver pre-release precedence, numeric components of any width compared by length then bytes). `TestSetHaltRefusesAFloorTheRunningBinaryMeets` runs the script over a cross product of 19 running versions and 30 floors against tm2's RPC server library answering `/status` and a stub `gnokey`, and requires refusal exactly where `meetsMinVersion` says the running binary meets the floor.
3. The script runs under `LC_ALL=C`, and so do `misc/release/cut-release.sh` and `misc/release/bump-protocol-version.sh`, which check the same shape. Regex bracket ranges follow the locale's collation: under `en_US.UTF-8` on glibc, `[A-Za-z]` also matches letters such as `é`, so without the export the shape check accepts versions `parseReleaseVersion` refuses, the unparseable floor it exists to keep out, and `[[ < ]]`, which orders alphanumeric pre-release identifiers, follows collation too. `TestReleaseToolingMatchesTheParser` requires the export in all three; `TestSetHaltChecksTheFloorShapeInTheCLocale` runs `set-halt.sh` under `en_US.UTF-8`, which catches a missing export only where glibc provides that locale (macOS bash 3.2 and the C locales do not match `é`).
4. A remote that does not answer, or answers without a `build_version`, stops the script: a proposal that cannot be checked is not made. A running version below the floor is printed with a reminder that it is one node. A running version the node does not parse meets no floor the shape check admits (on the node it meets only a byte-identical one), so the script goes on and warns that this one node says nothing about the others.
5. The version is fetched with `curl`, or `wget` when `curl` is missing, with the proxy variables cleared: `gnokey` dials `REMOTE` directly, and through a proxy the check could fail or read another node. An `http://` or `https://` `REMOTE` is read as given; one without a scheme or with `tcp://`, both of which `gnokey` accepts, is read over `http://`. Any other scheme stops the script, `unix://` included, which `gnokey` also accepts.

## Alternatives considered

- **A Go helper (`go run` of a small command, or a subcommand of `misc/deployments/upgrades`).** It would reuse the node's comparison instead of mirroring it, but would make a GovDAO script depend on a Go toolchain and a module download, where the scripts in `misc/govdao-scripts` need `gnokey` and common shell tools (`set-cla.sh` fetches with `wget`). The shell mirror follows the existing pattern for these scripts, whose `VERSION_RE` is held to the parser by `TestReleaseToolingMatchesTheParser`.
- **Warn instead of refusing.** The proposal executes in one transaction once the vote passes, and the damage (a chain that cannot restart its nodes before the halt) is done by then; a warning scrolls past.
- **An override variable.** No legitimate floor is at or below what the network runs; pointing `REMOTE` at a node that runs the current release answers the only case where a refusal could be wrong, a remote already running the new version.
- **Survey every validator.** The script knows one RPC; operators' nodes are not listed anywhere it can read. `UPGRADES.md` tells the person choosing the floor to survey `/status` across the network, and the warning repeats it when the one remote proves nothing.

## Consequences

- `set-halt.sh` needs network access to `REMOTE` before broadcasting (which broadcasting needs anyway) and `curl` or `wget`.
- The check runs when the proposal is created, and once the vote passes, any account can execute the proposal. A network that moves to a newer version during the vote can meet the floor by then; `misc/govdao-scripts/README.md` says to read `/status` again before casting the vote that passes it.
- The mirror can drift from the node; the test runs in the `gno.land` CI job, which triggers on changes to both `gno.land/**` and `set-halt.sh` (`.github/workflows/ci-dir-gnoland.yml`).
- `gno.land/cmd/gnoland/UPGRADES.md` states the rule (the floor is the version operators switch to, above every version the network runs) and points to the check.
