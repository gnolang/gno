# PR5348: Avoid Re-signing Known Self Votes in Consensus

## Context

`signAddVote` is the function responsible for signing a prevote or precommit and
dispatching it as an internal consensus message. It was called unconditionally on
every invocation, even when the validator had already signed a vote for the same
height/round/type.

This caused two problems:

**1. Unnecessary work with the local signer.**
`signVote` always calls `cs.wal.FlushAndSync()` before signing, an expensive disk
I/O. When `signAddVote` is triggered again for an already-signed vote (WAL replay
on restart, a timer re-firing, or a round re-evaluation), this flush was wasted.

**2. Incompatibility with strict remote signers (KMS/HSM).**
`types.PrivValidator` is an interface. The local `privval.PrivValidator`
implementation tolerates re-sign requests via its `sameHRS` +
`CheckVotesOnlyDifferByTimestamp` logic — the conflicting vote is silently dropped
by the vote set (`ErrAddingVote` is discarded, logged as INFO) and consensus
proceeds normally. However, a strict remote signer (e.g., a KMS) may reject
re-sign requests entirely, causing the node to never cast its vote for that round
and potentially stalling consensus.

The concrete trigger observed in tests: in a 2-validator network where one node's
prevote is delayed, the other node can reach a state where `signAddVote` is invoked
a second time for a vote already present in the vote set (e.g., received back via
the P2P reactor).

A secondary bug was also present: `cs.privValidator.PubKey().Address()` was called
before the `cs.privValidator == nil` guard, causing a nil pointer dereference for
observer/light-client nodes that have no private validator set.

## Decision

### Check the vote set before signing

Before calling `signVote`, `signAddVote` now looks up the current round's vote set
for an existing vote by this validator:

- **Vote exists, same block ID** — reuse it silently; skip signing and WAL flush.
- **Vote exists, different block ID** — log an error and return; do not attempt to
  sign a conflicting vote at the consensus layer.
- **No existing vote** — proceed with signing as before.

The vote set (`cs.Votes`) is already the authoritative record of what this validator
has committed to in the current round. Using it as a gate before signing is the
correct layer for this check and keeps the logic independent of the privval
implementation.

### Fix nil dereference ordering

The `cs.privValidator == nil` guard was moved above the `PubKey().Address()` call
so that observer nodes (nil privValidator) do not panic.

### Pass address as parameter

The address was being computed twice — once in `signAddVote` and once inside
`existingSignedVote`. The helper now accepts the address as a parameter, removing
the redundant computation and the now-unnecessary nil check inside the helper.

## Alternatives considered

**Rely solely on privval's `sameHRS` handling.**
Works for the local signer but breaks for strict remote signers. Also still triggers
the WAL flush unnecessarily. Rejected.

**Guard at each `signAddVote` call site.**
There are nine call sites. Scattering the same guard logic across all of them is
harder to maintain and easier to miss. Rejected.

## Key files

| File | Role |
|------|------|
| `tm2/pkg/bft/consensus/state.go` | `signAddVote`, `existingSignedVote` |
| `tm2/pkg/bft/consensus/state_test.go` | Unit tests: reuse and conflict paths |
| `tm2/pkg/bft/consensus/reactor_test.go` | Integration test: no re-sign in 2-node net |
| `tm2/pkg/bft/privval/privval.go` | Local `PrivValidator.SignVote` — `sameHRS` logic |

## Consequences

- `signAddVote` no longer triggers a WAL flush or a signer call when the vote is
  already in the vote set for the current round.
- With a local signer, re-triggered `signAddVote` calls were harmless but wasteful:
  the conflicting vote was silently dropped, the original vote retained, and
  consensus unaffected. The fix eliminates the unnecessary WAL flush and log noise.
- With a strict remote signer, the fix prevents the node from failing to cast its
  vote after a restart, which could stall consensus.
- The nil pointer dereference for observer nodes is eliminated.
- A new round resets the vote set, so signing proceeds normally after a round
  change — no behavioral change for the normal consensus path.
