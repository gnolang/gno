# ADR: Reading A Precommit's ValidatorAddress From Its Position

## Context

`Commit.Precommits[i].ValidatorAddress` carries no authority. `CanonicalizeVote`
in `tm2/pkg/bft/types/canonical.go` builds the signed payload from `Type`,
`Height`, `Round`, `BlockID`, `Timestamp` and the chain ID, so rewriting the
address leaves every signature in a genuine commit valid. `Commit.ValidateBasic`
checks per precommit only `Type`, `Height` and `Round`, and it never runs
`Vote.ValidateBasic`, which rejects a zero address, on a `CommitSig` on the
commit or block validation path. The field is redundant with the precommit's
position:
`VerifyCommit` verifies position `idx` against `vals.GetByIndex(idx)`, and
`VoteSet.MakeCommit` fills position `i` from the vote `addVote` already matched
to that position by index and address.

Three readers exist on the commit path.

- `Commit.GetVote` in `tm2/pkg/bft/types/block.go` copied the field into the
  vote it rebuilds for sign bytes and for `CommitToVoteSet`. The sign-bytes use
  is inert, since `CanonicalizeVote` discards the field; the reconstruction use
  was not.
- `CommitToVoteSet`, the inverse of `MakeCommit`, feeds those votes into
  `VoteSet.AddVote`, which rejects a vote whose address does not match the
  validator at its index and panics with `Failed to reconstruct LastCommit`. Its
  only production caller is `reconstructLastCommit` in
  `tm2/pkg/bft/consensus/state.go`, run at node construction and at the
  fast-sync to consensus handoff.
- `VerifyFutureCommit` in `tm2/pkg/bft/types/validator_set.go` looks the address
  up in the old validator set through `GetByAddress`. A mismatched address
  degrades to skipping that precommit, and the old-power threshold then rejects
  the commit. It has no caller in the tree.

The defect is reachable from a fast-sync peer. The peer hands a node the genuine
block at height h and a self-consistent block at h+1 whose `LastCommit` is the
genuine commit for h with one `ValidatorAddress` rewritten.
`Block.ValidateBasic` requires only `LastCommitHash == LastCommit.Hash()`, which
the attacker computes, and `VerifyCommit` checked signatures positionally and
nothing else, so `state.Validators.VerifyCommit` in
`tm2/pkg/bft/blockchain/reactor.go` accepted the tampered commit and `SaveBlock`
persisted it as the seen-commit for h. At the handoff,
`SwitchToConsensus` runs `reconstructLastCommit`, whose `CommitToVoteSet`
panicked inside `poolRoutine`, a goroutine with no recover, ending the process.
The poisoned commit stays in the block store under the seen-commit key for h, so
every restart runs `NewConsensusState` into the same panic until an operator
repairs or wipes the store.

## Decision

1. `CommitToVoteSet` derives `vote.ValidatorAddress` from `vals.GetByIndex(idx)`
   instead of copying the wire value. Reconstruction is deterministic for any
   genuinely signed commit, so a node that already synced a poisoned seen-commit
   rebuilds the correct vote set and starts.
2. `VerifyCommit` rejects a precommit whose `ValidatorAddress` differs from the
   validator at its positional index. A tampered commit is refused at the
   fast-sync boundary and never persisted as the seen-commit.

The field stays on the wire and in `Commit.Hash()`: `CommitSig` is `Vote` and
the address is amino-encoded, so dropping it is a wire and hash change for a
field the positional reading already neutralizes. A commit-specific type is
where that belongs.

## Alternatives considered

**A. Bounds-check `GetByIndex` in `CommitToVoteSet`.** Not needed.
`CommitToVoteSet` takes only the address `GetByIndex` returns, which is the zero
address past the end of the set, so a precommit count that exceeds the set
reaches `AddVote`'s own index check and keeps its descriptive panic. That shape
panics on both sides of this change, and `VerifyCommit` refuses it before a
commit reaches the store.

**B. Reject the mismatch in `Commit.ValidateBasic`.** Rejected. It has no
validator set to compare against, so it could at most reject a zero address and
leave a rewrite to another validator's address open.

**C. Remove `ValidatorAddress` from `CommitSig`.** Rejected for this change.
`CommitSig` is `Vote`, the field is amino-encoded into `Commit.Hash()`, and
commits are stored and exchanged by that hash, so removing it moves the wire
format and every stored commit's hash.

**D. Return an error from `CommitToVoteSet`.** Rejected. Its caller treats a
failure as unrecoverable state corruption and panics; the fix removes the
reachable failure instead of widening the signature.

**E. Fix `CommitToVoteSet` alone.** Rejected as insufficient. The node would
stop crashing and keep accepting and storing tampered commits, and the
`VerifyCommit` rule is what keeps the poisoned commit out of the store.

## Consequences

- A tampered commit is refused where fast sync accepts it, and a poisoned commit
  already in a block store rebuilds into the correct vote set, so a node that
  synced one starts after the upgrade and no longer crash-loops.
- The `VerifyCommit` check tightens consensus validation: a block whose
  `LastCommit` carried a mismatched address, which the old code accepted, is now
  rejected, including on WAL replay and in `ValidateBlock`. A chain produced by
  this code holds no such block, because `VoteSet.MakeCommit` fills every
  position from a vote `addVote` already matched to that position by index and
  address. Only a chain on which +2/3 of voting power accepted a tampered commit
  under the old code could carry one, and patched nodes cannot fast-sync past
  such a block, including from genesis. Shipping this as a rolling `class:node`
  upgrade therefore assumes that no validator relabels a commit while more than
  2/3 of voting power is unpatched, and that existing history holds no such
  block. A chain that cannot assume both needs the rule gated by height.
- `Commit.GetVote` still copies the address for the sign-bytes path, where
  `CanonicalizeVote` discards it, so the field's only live use is positional.
- `VerifyFutureCommit` keeps reading `precommit.ValidatorAddress` through
  `GetByAddress`. A commit that reaches it has passed `VerifyCommit`, so an
  address absent from the old set is a validator the old set did not hold, and
  the precommit is skipped rather than misattributed.
- `VerifyCommit` also refuses a precommit whose `ValidatorIndex` is not its
  position, the other unsigned field. See
  `pr6504_mediantime_positional_index.md`.
