# ADR: Deriving Block Time From The Position Of A Precommit

## Status

Proposed.

## Context

`MedianTime` in `tm2/pkg/bft/state/state.go` derives a block's time from the
previous commit: each precommit contributes its `Timestamp`, weighted by the
voting power of the validator that cast it. It found that validator with
`validators.GetByIndex(vote.ValidatorIndex)` and read `validator.VotingPower`
without checking the result.

`CommitSig.ValidatorIndex` carries no authority on that path.

- `CanonicalizeVote` in `tm2/pkg/bft/types/canonical.go` builds the signed
  payload from `Type`, `Height`, `Round`, `BlockID`, `Timestamp` and the chain
  ID. `ValidatorIndex` is not in it, so rewriting the field leaves every
  signature in the commit valid.
- `VerifyCommit` in `tm2/pkg/bft/types/validator_set.go` verifies position `idx`
  against `vals.GetByIndex(idx)`, and takes the signed bytes from
  `commit.VoteSignBytes(chainID, idx)`, whose `GetVote` writes `idx` into the
  `ValidatorIndex` of the vote it reconstructs. The stored field is already
  discarded there.
- `Commit.ValidateBasic` checks `Type`, `Height`, `Round` and the range of
  `BlockID.PartsHeader.Total`, and nothing else per precommit.
  `Vote.ValidateBasic`, which rejects a negative index, is not called for a
  `CommitSig` on the commit or block validation path.

The bounds check that does exist is on the live-gossip path: `VoteSet.addVote`
in `tm2/pkg/bft/types/vote_set.go` rejects an index outside the set, and also
requires the vote's address to match the validator at that index. Neither reaches
a commit embedded in a block.

Two things follow from weighting by an unauthenticated field.

`GetByIndex` returns a nil `*Validator` for an index outside the set, so an
out-of-range value is a nil dereference in `MedianTime`. `ValidateBlock`
(`validation.go:130`) reaches it with a commit the local node did not build, on
every block a validator checks. `MakeBlock` (`state.go:147`) takes only the
proposer's own `VoteSet` commit. In consensus the panic reaches the
`recover` in `receiveRoutine` (`tm2/pkg/bft/consensus/state.go:638-654`), which
logs `CONSENSUS FAILURE!!!` and then runs `onExit`; the routine stops and does
not restart. Every honest validator runs the same check on the same block, so
one proposer stops all of them at one height.

An in-range value that points at a different validator is the quieter case: it
weights a timestamp with someone else's power and moves the derived time, and
`ValidateBlock` requires `block.Time` to equal it. Measured on three validators
of power 1, 10 and 1000, pointing every precommit at the weakest moves the
median 20 seconds earlier.

The median itself had a second defect. `WeightedMedian` in
`tm2/pkg/bft/types/time/time.go` sorted the timestamps by `UnixNano`, which
wraps for a time before 1678 or after 2262. `Timestamp` is signed, and
nothing on the commit path bounds it, so one validator can sign a vote that
wraps in among the honest ones and becomes the median. Measured on six
validators of power 60, the set `gnoland-1` ran at height 992883, five voting
on 2026-10-22 and one on 2611-05-12: the old order derives 2611-05-12.
`ValidateBlock` accepts it, and the chain keeps committing in that year: each
honest validator stamps its next votes at the locked block's time plus
`TimeIotaMS`, so every later block lands `TimeIotaMS` after the last. Vesting
schedules and session expiry read the block time, so every schedule vests in
full and every session expires.

## Decision

`MedianTime` weights each precommit by the validator at its own position in the
commit, `validators.GetByIndex(i)`, which is the position `VerifyCommit` checked
its signature against. `CommitSig.ValidatorIndex` is not read.

A nil validator is skipped rather than dereferenced. With positional lookup that
needs more precommits than validators, which `ValidateBlock` rejects twice
before `MedianTime` runs, once against `state.LastValidators.Size()` and again
inside `VerifyCommit`. `MakeBlock` checks neither, and takes whatever commit its
caller hands it.

`WeightedMedian` orders the timestamps with `time.Before`, which orders every
time a vote can carry, so a far-future vote sorts after every honest one.
Between 1678 and 2262 both orders agree.

`VerifyCommit` refuses a precommit whose `ValidatorIndex` is not its position,
beside the check that its `ValidatorAddress` names the validator there. Those
are the two fields of a precommit no signature covers and no other check pins,
so after both checks every field a validated commit carries is either signed or
equal to what its position implies, and a later reader of the index needs no
bound of its own.

## Alternatives considered

**A. Bounds-check `vote.ValidatorIndex` and keep reading it.** Rejected. It
stops the crash and leaves the skew: an index inside the set still weights a
timestamp with the wrong validator's power, and that is the form an attacker
reaches for when a halt is too loud. Position is also the only reading
consistent with what `VerifyCommit` proved.

**B. Reject `ValidatorIndex != i` in `Commit.ValidateBasic` instead.**
Rejected. It would refuse the shape a step earlier on a fast-sync message, but
both paths that run `ValidateBasic` on a commit, `ValidateBlock` and the
fast-sync reactor, reach `VerifyCommit` before the commit is used or stored,
and `VerifyCommit` already holds the check on the other unsigned field, the
address.

**C. Give `MedianTime` an error return.** Rejected. Both callers would have to
handle a branch that cannot be reached on the validated path, and `MakeBlock`
has no error return of its own to fold it into.

**D. Cut each precommit in a commit down to a flag, the address, a timestamp
and a signature, as Tendermint 0.33 did, or drop the address too.** Not now. A
`Commit` stores a full `Vote` per validator, so the index, the address and the
block repeat per precommit and each needs its own check. Dropping them
removes the class rather than checking it, but changes the encoding of every
commit. Nodes keep their data dir across this release's halt, so the new
binary would have to decode, hash and serve both encodings by height.

## Consequences

- An honest commit derives the same time as before. `VoteSet.MakeCommit` fills
  position `i` from `voteSet.votes[i]`, and `addVote` has already checked that
  the vote's index and address match that position, so `ValidatorIndex == i`
  throughout.
- A block whose `LastCommit` carries a precommit off its position is refused,
  and a fast-sync peer's commit carrying one is refused before it is stored.
  `MedianTime` still reads the position rather than the field, so `MakeBlock`,
  whose commit `VerifyCommit` never sees, derives the honest time either way.
- `CommitSig.ValidatorIndex` stays on the wire, since `CommitSig` is `Vote` and
  the type is amino-encoded into `Commit.Hash()`.
- Nothing about state, the wire format or a hash moves. A commit whose
  precommits each carry their own position as `ValidatorIndex`, with
  timestamps between 1678 and 2262, derives the same time on a patched and an
  unpatched node. Each half of the change splits them on a different attack.
  The index half needs a byzantine proposer: a patched node refuses the block,
  where an unpatched one accepts it with an in-range index and halts on it with
  an out-of-range one. The comparator half needs one byzantine validator and
  no difference in power: `Timestamp` is inside
  `CanonicalizeVote`, so an honest proposer carries the signed vote into the
  next block. On the six validators of power 60 above, the new order derives
  2026-10-22 where the old one derives 2611-05-12, and each side rejects the
  block the other proposes.
- Upgrading is still per node, since neither attack leaves an unpatched chain
  as it was: an index attack halts it or skews its time, and a far-future vote
  moves its time past 2262 for good. While both binaries run, the same attack
  splits them, and only a side holding more than 2/3 of the voting power can
  commit.
- A patched node also re-derives the time of every block it fast-syncs:
  `ApplyBlock` calls `ValidateBlock`, and `blockchain/reactor.go` panics on a
  mismatch. Replay at start-up skips that check through `ExecCommitBlock` up
  to the last block. A patched node therefore joins a chain only where no
  committed block carries either shape. On `gnoland-1`, read over RPC on
  2026-10-22, the `last_commit` of every block from height 2 to 992877 carries
  neither: 4742212 precommits, each `ValidatorIndex` equal to its position and
  every timestamp between 1678 and 2262.
- `MedianTime` skipping a precommit past the end of the set is covered by
  `TestMedianTimeSkipsPrecommitsPastTheValidatorSet`; removing the guard panics
  it.
