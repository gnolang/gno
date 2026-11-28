# ADR: gpao skips an already-live package before verifying it

## Context

gpao asked `vm/qpkgmeta_json` whether a candidate's path was already live with nothing pending only after verifying it, so every settled package paid for a full verification first: a child process, its import closure fetched over `vm/qfile`, a typecheck and a preprocess. That is the common case when catching up with `--start-height` over blocks already approved. A settled `p/nt/ufmt/v0` took about 280 ms; a realm importing it, 6 `vm/qfile` requests.

## Decision

1. `handleCandidate` asks `isSettled` right after the `seen` check, before spawning the verifier, and records a settled package exactly as the later check did: `approved`, "already active on-chain", its bytes marked seen. It waits at most 10 seconds for the node: giving up only means the package is verified, and verifying reports a stalled node as unavailable.
2. The check after verification stays, since the package may go live while it is being verified. It keeps the RPC client's deadline, since giving up there sends an enable, which costs a fee.
3. A live package with a redeploy parked over it is not settled, so it is still verified and enabled.

## Alternatives considered

- **Keep the check after verification only.** Every settled package costs a verification whose verdict cannot change what the oracle does.
- **Move the check rather than duplicate it.** A package enabled by someone else during its verification would then be enabled again, paying a fee to be told it is live.

## Consequences

- A settled package costs one query instead of a verification: about 0.3 ms and no `vm/qfile` request.
- A settled package whose submitted bytes would not have verified is reported `approved` rather than `rejected`. The live package is what the chain runs, and no enable was possible either way.
- A package that is not settled pays one more `vm/qpkgmeta_json` query.
