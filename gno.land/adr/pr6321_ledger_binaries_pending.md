# PR6321: A ledger entry's binaries may be pending until CI builds the release

## Context

The upgrade ledger (`misc/deployments/<chain>.gno.land/upgrades.json`, rules in `gno.land/pkg/upgrades`, introduced by PR6228) records one entry per release a network has run, including a `binaries` map of per-platform download URLs carrying their sha256. `misc/release/cut-release.sh` refuses to tag a final release that the ledger does not record. The checksums only exist once CI has built the tagged release, and the ledger refused an entry without them, so no valid entry for a final release could exist at the moment it had to be cut. The gate only passed because it asked whether the entry existed (`upgrades has`, which parses) and never whether the ledger validated.

## Decision

1. `binaries` may be JSON `null`, meaning "not built yet", as `image.digest` already could. `{}` is refused, so a pending entry cannot pass for a built one with no downloads. Go tells them apart through `encoding/json`, which decodes `null` (or an absent key) to a nil map and `{}` to an empty one.
2. Only the last entry may be pending. An earlier release has been built, and a node replaying from genesis needs its downloads, so the binaries are filled in before the next release's entry is added.
3. `cut-release.sh` runs `upgrades check`, the check CI runs, before looking for the version's entry: a final release is refused when the ledger does not validate or its `UPGRADES.md` is stale, and a release candidate is warned and goes on.
4. The daily `ci / upgrades-ledger-digests` job skips null binaries, as it skips a null image digest. Any other non-object value still fails it.

## Alternatives considered

- **Keep `binaries` required and write the final's entry after the cut.** The gate would have to move after the tag, which defeats its purpose: it is the last point at which a release missing from the ledger can be stopped.
- **A pointer to the map, or a custom unmarshaller, to tell null from empty.** Both add code for a distinction `encoding/json` already makes.
- **Allow null binaries on any entry.** Nothing would then catch a release built and never filled in; restricting it to the last entry makes the next final release's gate catch it.
- **Have the daily job fail on an entry pending too long.** A grace period adds a clock to a check that is otherwise a pure comparison; the last-entry rule and the next release's gate already bound how long an entry can stay pending.

## Consequences

- An entry is written before its release is cut, committed on top of the code it names, and the release is cut from that checkout with `--commit` naming the code commit, since the script reads the ledger from the checkout it runs in. RELEASING.md and the release tooling README say so.
- Between the cut and the fill-in, `UPGRADES.md` tells operators to run the last row with no *(pending)* cell, so an unbuilt release is never the one they are pointed to.
- The Go ledger tool accepts a missing `binaries` key as pending, as it does for the other pending fields. The JSON Schema lists every key as required, so an entry checked against it spells `binaries` out.
- `cut-release.sh` does not compare the entry's `commit` with the commit it tags; a stale sha in an entry written early is caught only by review.
