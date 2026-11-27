# ADR: Rejecting A Precommit's Part-Set Total Before It Is Canonicalized

## Context

`Commit.ValidateBasic` in `tm2/pkg/bft/types/block.go` checks each precommit's
`Type`, `Height` and `Round`, and nothing else. It never runs `Vote.ValidateBasic`,
which for a standalone vote reaches `BlockID.ValidateBasic` →
`PartSetHeader.ValidateBasic` and bounds `PartSetHeader.Total`.

`PartSetHeader.Total` is a platform `int` while the canonical wire form is a
`uint32`: `CanonicalizePartSetHeader` in `tm2/pkg/bft/types/canonical.go`
panics on a `Total` that is negative or above `math.MaxUint32`. A commit reaches
that function through `VoteSignBytes` → `CanonicalizeVote` →
`CanonicalizeBlockID`. `Total` is part of the signed payload as a `uint32`, but
the panic fires while the sign bytes are being built, before the signature is
checked, so the crafted commit needs no valid signature at all.

The panic has two ways in. In fast sync, `blockchain/reactor.go`'s
`poolRoutine` calls `state.Validators.VerifyCommit` on a peer-supplied block's
`LastCommit`; `VerifyCommit` calls `commit.ValidateBasic()` and then, per
precommit, `commit.VoteSignBytes(...)`. `poolRoutine` runs as a bare goroutine
with no `recover`, so any peer the node syncs from ends the node process.

In consensus, the round's proposer writes its block's `LastCommit`. That commit
arrives inside the proposal block rather than as a vote message, so the
`Vote.ValidateBasic` that `ConsensusReactor.Receive` runs on votes never sees
it. `defaultDoPrevote` calls `state.ValidateBlock` on the proposal, which
reaches `VerifyCommit`. The `receiveRoutine` recover logs `CONSENSUS FAILURE!!!`
and runs `onExit`, so the consensus routine stops on every node that validates
the proposal, until its process is restarted.

## Decision

`Commit.ValidateBasic` rejects any precommit whose
`BlockID.PartsHeader.Total` is negative or above `math.MaxUint32`, with the same
range `CanonicalizePartSetHeader` requires. The check is on `Total` alone: it
mirrors exactly what the canonicalization needs, and it leaves every other shape
the commit path accepts today unchanged.

The one check closes both ways in: `Block.ValidateBasic` runs it on the
proposal's `LastCommit`, and `VerifyCommit` runs it before building any sign
bytes.

## Alternatives considered

- *Call `Vote.ValidateBasic` per precommit*: it would additionally reject a
  zero `ValidatorAddress`, an incomplete `BlockID` and a missing signature,
  shapes a `CommitSig` can carry today and the reconstruction path tolerates. A
  larger behaviour change than the panic needs.
- *Bound `Total` by `MaxBlockPartsCount`*: stricter than the canonicalization
  range and a consensus rule the commit path does not have today.
- *A `defer`/`recover` in `poolRoutine`*: turns the crash into a silent stop of
  the fast-sync routine and hides every later panic on the same goroutine. It is
  defense in depth, not the fix, and it is left out of this change.

## Consequences

- A block whose `LastCommit` carries an out-of-range `Total` fails
  `Block.ValidateBasic`, so fast sync refuses it in `Receive` and disconnects
  the peer; the node stays up. `VerifyCommit` refuses the same commit before
  any signature work for any other caller.
- A proposal carrying such a commit fails `state.ValidateBlock`, so an honest
  validator prevotes nil and its consensus routine keeps running.
- An honest commit is unaffected: its `Total` comes from chunking a real block
  and is a small positive number.
- The regression test drives the attack shape through `VerifyCommit`; with the
  check removed, the test panics in `CanonicalizePartSetHeader`.
