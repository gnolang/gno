# ADR: Publish a release early, promote it once mainnet's switch is set

## Context

#6228 (its Decision 4) gave GitHub's "Latest" release and the CLI images' `latest` tag to the highest final release, set at publication: `release / chain-tag` published the highest final tag with `--latest`, `release / docker` pushed `latest` for `gno`, `gnokey`, `gnodev` and `gnoweb` on it, and `misc/install.sh` installs the highest `v*` release that is not a pre-release. Publishing a final release therefore made it what operators land on by default, while mainnet still ran the previous one.

After v1.6.0, the release runbook cut a final only once mainnet's switch was set: for a coordinated upgrade, once the GovDAO halt proposal had executed, shortly before the halt. The build takes about 40 minutes, so the binaries and images would arrive shortly before the halt, leaving operators and infrastructure (which prepares machine images ahead of time) little time to fetch and check them, and no slack for a build failure.

v1.6.0's tag message records an estimated halt height, which a pushed tag can never correct. #6381 removed the height from tag messages, so nothing a release publishes (tag, binaries, images) names a height: Latest was the only reason left to wait for the proposal.

v1.7.0 was published the evening before mainnet's halt, and held by hand as each workflow ended: the release marked a pre-release, Latest given back to v1.5.1, and `latest` pointed back at v1.5.1's images by putting their manifests under it again.

## Decision

1. A final release is cut and published as soon as its last candidate has passed on the testnet, before mainnet's switch is set, and without `--halt-height`. The halt height goes only where it can be edited (the release notes, the release's tracking issue, the announcement to operators), as an estimate until the proposal has executed. The proposal is created with `govdao-exec.sh set-halt` once the height is chosen.
2. `release / chain-tag` creates and publishes every release as a pre-release, the final one included, and never as Latest. `release / docker` pushes each image's version tag only, never `latest`.
3. `misc/release/promote-release.sh <version> [--push]` promotes a final release once mainnet's switch is set (for a coordinated upgrade, once the halt proposal has executed; for a rolling one, on the agreed day): it puts each CLI image's version manifest under `latest`, byte for byte, then makes the release a full release and GitHub's "Latest". Without `--push` it runs every check and changes nothing.
4. A promotion only moves forward from the highest full final release, the one `misc/install.sh` installs: not GitHub's "Latest" marker, which a `chain/<name>` genesis record created by hand can hold, and not the highest tag, so a patch to the running line is promoted while a higher MINOR is held for its halt. The script refuses a release candidate, a draft, an older release, a MINOR whose halt mainnet has not scheduled (mainnet's `halt_min_version`, read from its RPC, must name it; a PATCH and the release mainnet runs need none), a MAJOR (a new network, which mainnet never schedules), and an image it cannot read or whose bytes do not match their digest; it checks every image before writing anything, while a missing write permission shows only at the first write, where ghcr.io checks it.

## Alternatives considered

- **Keep cutting the final once the switch is set.** It leaves operators and infrastructure under an hour with the release, and a failed build in that hour delays the halt.
- **Publish early and hold by hand, as v1.7.0 was.** Every hold is manual, and the workflows give the release Latest and `latest` first, so for some minutes `misc/install.sh` and `latest` serve a release mainnet has not scheduled.
- **Publish the final under another name, and cut `vX.Y.Z` once the switch is set.** The binaries carry their tag as their version, and `halt_min_version` names the final; a build cut later is a different build from the one operators fetched and checked, since the release builds are not reproducible.
- **A manually triggered workflow for the promotion.** It would hold write access to releases and packages in a workflow anyone with write access can trigger. A script run by the release manager with their own credentials, like `cut-release.sh`, shows the result in a dry run first and adds no such workflow.
- **Promote automatically once the mainnet ledger records the release.** The ledger entry is written once the chain runs the release (#6381), after the halt; the release has to be promoted once the halt is scheduled, so that an operator installing during the window gets the release the chain switches to.

## Consequences

- Between publication and promotion a final release is a pre-release: `misc/install.sh` installs it only with `--version`, and `latest` stays on the release mainnet runs. Operators pin version tags in any case (`VALIDATOR.md`).
- A node started on the final release before the halt proposal executes is not refused: no halt is scheduled yet. The docs and notices name the release the chain runs.
- A final that is never promoted stays a pre-release for good: a final mainnet skips, as v1.6.0 is, or a patch to an older line.
- Should mainnet skip a release after its promotion (a cancelled halt), the same script restores the running release once the skipped one is a pre-release again (RELEASING.md, "Publishing early, promoting once the switch is set").
- Each final release takes one more step once the switch is set. Forgetting it leaves Latest and `latest` one release behind, which affects what `misc/install.sh` and `latest` serve, not the chain; the release's tracking issue lists the step.
- A coordinated release's tag message does not name its `halt_min_version`: the final is cut without `--halt-height`, which also prints the halt proposal only while the tag does not exist yet. #6381's Decision 3 is superseded in part.
- #6228's Decision 4 is superseded in part: `latest` and Latest follow promotion, not tag order at publication.
- CI checks the workflows' flags, the script's version checks, its choice of the release to move forward from, its header parsing, and that it never moves `gnoland:latest`; it does not run the registry or release writes. The manifest write is the request used to hold v1.7.0, which ghcr.io accepted.
- A tag push runs the release workflows of the tagged commit, so this change applies to a release once it is on `chain/mainnet`, its tests and `promote-release.sh` with it.
