# ADR: A ledger entry's binaries may be pending until CI builds the release

## Context

The upgrade ledger (`misc/deployments/<chain>.gno.land/upgrades.json`, rules in `gno.land/pkg/upgrades`, introduced by #6228) records one entry per release a network has run, including a `binaries` map of per-platform download URLs carrying their sha256. `misc/release/cut-release.sh` refuses to tag a final release that the ledger does not record. The checksums only exist once CI has built the tagged release, and the ledger refused an entry without them, so no valid entry for a final release could exist at the moment it had to be cut. The gate only passed because it asked whether the entry existed (`upgrades has`, which parses) and never whether the ledger validated.

## Decision

1. `binaries` may be JSON `null`, meaning "not built yet", as `image.digest` already could. `{}` is refused, so a pending entry cannot pass for a built one with no downloads. Go tells them apart through `encoding/json`, which decodes `null` (or an absent key) to a nil map and `{}` to an empty one.
2. Only the last entry may be pending. An earlier release has been built, and a node replaying from genesis needs its downloads, so the binaries are filled in before the next release's entry is added. The image digest is filled in together with them, never before: `UPGRADES.md` does not show the binaries, so a filled digest must mean a built release.
3. `cut-release.sh` runs `upgrades check`, the check CI runs, before looking for the version's entry: a final release is refused when the ledger does not validate or its `UPGRADES.md` is stale, and a release candidate is warned and goes on. It then compares the commit the entry names with the commit it is about to tag, read with `upgrades commit`: a final release is refused on a mismatch, a candidate warned. The entry is committed on top of the code it names, so the branch tip is the ledger commit, and a cut without `--commit` would otherwise tag it.
4. The daily `ci / upgrades-ledger-digests` job lists null binaries and null image digests as pending instead of checking them. A string, number or boolean in `binaries` still fails it; an array or an empty object is left to `upgrades check`, which refuses them.

## Alternatives considered

- **Keep `binaries` required and write the final's entry after the cut.** The gate would have to move after the tag, which defeats its purpose: it is the last point at which a release missing from the ledger can be stopped.
- **A pointer to the map, or a custom unmarshaller, to tell null from empty.** Both add code for a distinction `encoding/json` already makes.
- **Allow null binaries on any entry.** Nothing would then catch a release built and never filled in; restricting it to the last entry makes the next final release's gate catch it.
- **Have the daily job fail on an entry pending too long.** A grace period adds a clock to a check that is otherwise a pure comparison; the last-entry rule and the next release's gate already bound how long an entry can stay pending.

## Consequences

- An entry is written before its release is cut, committed on top of the code it names, and the release is cut from that checkout with `--commit` naming the code commit, since the script reads the ledger from the checkout it runs in. RELEASING.md and the release tooling README say so.
- Between the cut and the fill-in, `UPGRADES.md` and `VALIDATOR.md` tell operators to run the last row with no *(pending)* cell; since the digest is filled in with the binaries, an unbuilt release is never the one they are pointed to. An upgrade's row stays *(pending)* until its halt time is recorded, so the halt time is recorded as soon as the halt has happened.
- The Go ledger tool accepts a missing `binaries` key as pending, as it does for the other pending fields. The JSON Schema lists every key as required, so an entry checked against it spells `binaries` out.
- An entry written before the cut can carry a stale `commit` if the branch moves before the tag; `cut-release.sh` refuses the final release until the entry and the tagged commit agree.
