# ADR: A ledger entry is written once the chain runs the release

## Context

The upgrade ledger (`misc/deployments/<chain>.gno.land/upgrades.json`, rules in `gno.land/pkg/upgrades`) records one entry per release a network has run. #6228 made facts that have not happened yet `null`, rendered *(pending)*, and had `misc/release/cut-release.sh` refuse a final release the ledger does not record (its Decision 7). #6321 let the binaries be pending too, so the entry could be written before the release is cut, and had the script check that the entry names the commit being tagged. A coordinated upgrade's entry carries its halt height, and `cut-release.sh --halt-height` also writes that height into the annotated tag's message.

Until the GovDAO proposal setting it has executed, that height is an estimate: a target time converted to a height at the current block rate, which moves whenever the vote or the date does. A release cut before then records the estimate: in the ledger entry, which a commit can correct, and in the tag message, which nothing can, since a pushed tag never moves. v1.6.0 was cut with an estimated height before its mainnet date moved, and its tag message keeps that height.

An entry written early also takes one follow-up commit per fact that arrives later: the digest and binaries once CI has built the tag, the proposal once it exists, and the halt time once the chain has halted. Each is pushed to the chain branch and ported to `master`. The halt time is the last of them, and it exists only after the halt.

## Decision

1. An entry is written once the chain runs the release, after the halt for a coordinated upgrade and once validators run it for a rolling one, with nothing pending.
2. `cut-release.sh` warns, instead of refusing, when the ledger has no entry for a final release. It still refuses a final release whose ledger fails `upgrades check`, or whose entry names another commit than the one being tagged. A release candidate still only warns: the final release cut after it, on its commit, is checked the same way.
3. The tag message of a coordinated upgrade records `halt_min_version` and no height. `--halt-height` still prints the GovDAO proposal to go with the release. A release page carries the halt height once the proposal has executed.

## Alternatives considered

- **Keep writing the entry before the cut, and correct the height when it moves.** Each value written before it is known is one more value to correct, and the tag message cannot be corrected at all.
- **Write the entry once the halt proposal has executed, before the halt.** The height is right by then, but the halt time is still pending, so the entry still needs a follow-up commit. Writing it once the chain runs the release needs none.
- **Drop `--halt-height` from `cut-release.sh`.** The proposal it prints is the one the release manager submits; only the tag message was wrong.
- **Remove the pending allowance and the commit check, which an entry written before the cut needed.** The allowance is part of the ledger format, which the JSON Schema and the daily `ci / upgrades-ledger-digests` job also read; an entry written after the upgrade leaves it unused, and removing it is a format change for no behaviour. The commit check still catches an entry written ahead of the cut, which nothing forbids, and costs one lookup.
- **A CI check that every released version has an entry.** A chain runs a release some time after its tag, so the check would need a grace period, a clock in a check that is otherwise a comparison; #6321 declined a similar check for the same reason. The release's tracking issue (one per release, such as #6372) has a checklist item for each chain's entry.

## Consequences

- #6228's refusal of a final release without an entry, which #6321 kept, rested on the cut being the last point at which a missing entry can be stopped. That does not hold here: the entry is written after the cut by design. Nothing in CI catches an entry that is never written; the release's tracking issue does.
- The ledger format is unchanged. A complete entry has no *(pending)* cell, so the last row with none, which `UPGRADES.md` and `VALIDATOR.md` point operators at, is the last row.
- Before a coordinated halt, `UPGRADES.md` has no row for it. Operators learn of it from the release page and the announcement, which name the height once the proposal has executed.
- Tags already pushed keep their messages, v1.6.0's estimated height included; its release page says so.
