# ADR: gpao verifies again the packages it left pending

## Context

gpao reads each block once, and heights only move forward. A package left pending, on a parked import or by a verification that could not finish, was not reached again: only a resubmission or a restart with `--start-height` at or below its block retried it. One failure strands every package above it. On onyx, two packages ran past the 1-minute prepare budget (their import closures cost 183 and 163 sequential `vm/qfile` queries at about 0.4 s each), and the eight packages importing them, directly or through each other, stayed pending as waiting on a parked import, although nothing was wrong with any of them.

## Decision

1. When `handleCandidate` leaves a submission pending, the verifier remembers its path and submitting height (`stranded`). Any other outcome is a verdict and forgets it. The bytes are not kept.
2. Every 10 minutes the verifier, between candidates, looks at each remembered submission. One the chain no longer holds parked at that height is forgotten: it was enabled, by anyone, or replaced, and a replacement reaches the verifier from its own block. One whose chain imports are all live is verified again, its bytes read from the submitting block through the same decoding and DeliverTx check as block processing. One still blocked keeps the import that blocked it, and its block is not read again until that import is live.
3. Each submission is verified again at most 3 times per run.
4. The look stops as soon as a candidate is queued, so new blocks go first.

## Alternatives considered

- **Scan the chain's parked packages (`vm/qinertpaths`).** It would survive a restart, but the listing reads every parked package and is metered by bytes, so parking enough large packages pushes it past the query gas ceiling and turns retrying off. A path-only index would need a keeper change.
- **Keep the bytes of pending submissions.** It would save the block read, but it lets submitters grow the daemon's memory with content.
- **Retry inside `handleCandidate`.** It would hold the verifier on one package while new blocks wait.

## Consequences

- A package waiting on a parked import is verified within one interval of the import going live, whoever enabled it.
- Nothing outside the oracle can turn retrying off, and the block reader never waits on it.
- What is remembered grows with the submissions this run left pending, as the status board does, and shrinks as the chain enables or replaces them or as they reach 3 attempts.
- A restart forgets what was pending: `--start-height` at or below the oldest such block recovers it, and replaying skips packages already live without fetching them.
- A submission that overran the prepare budget meets the same budget again; retrying helps the packages blocked behind it, not the overrun itself, which needs a larger `--prepare-budget` or faster queries.
- A redeploy parked over a live private realm is not verified again: `vm/qpkgmeta_json` reports the live package for that path.
