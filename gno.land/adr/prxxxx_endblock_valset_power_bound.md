# ADR: Bound A Valset Proposal By The Apply It Will Meet

Superseded in part by `prxxxx_valset_proposed_write_dry_run.md`: the same dry
run now refuses the write of `valset:proposed`, so the EndBlocker gate below
is a fallback, and the change as a whole is a coordinated upgrade rather than
one nodes take independently.

## Context

`EndBlocker` in `gno.land/pkg/gnoland/app.go` is the last gate before a
governance valset proposal becomes an ABCI `ResponseEndBlock.ValidatorUpdates`
delta. It parses `valset:proposed`, applies a min-validator floor and a
pubkey-type allow-list, commits `valset:current := proposed`, and returns the
delta. Nothing in it bounds the entries' numeric voting power, and the
realm-side write path accepts any power `ParseUint(…, 10, 63)` can hold — up to
`MaxInt64`.

tm2 does enforce magnitude, but only later: `ApplyBlock` calls
`ValidatorSet.UpdateWithABCIValidatorUpdates` on `state.NextValidators`, whose
`processChanges` rejects a per-entry power above `MaxTotalVotingPower` and
whose `verifyUpdates` walks the delta's updates against the current total,
failing when the running total exceeds the cap. That runs AFTER `EndBlock`
returns and after block H is already in the block store. When it fails, the
node kills itself; on restart, replay re-executes the same stored block
deterministically and fails identically, so the chain cannot progress past H
without a coordinated emergency binary. A governance proposal that merely
looks large therefore becomes an unrecoverable halt.

The accounting has a non-obvious shape: `verifyUpdates` does not credit
removals that arrive in the same delta. It starts from the current set's total
and adds each update's delta (new power for an addition, difference for an
update). So `current {cap/2, cap/2} → proposed {new: cap}`, whose target set's
own total is exactly the cap, is refused by `verifyUpdates`: the cap is odd, so
the two halves sum to `cap - 1`, and it sees `(cap - 1) + cap`. A gate that only
bounds the proposed set's total accepts that proposal and the chain halts.

## Decision

Mirror the apply in the gate. Before committing `valset:current`, build a
throwaway `ValidatorSet` from `valset:current` and run the delta through the
same call consensus will use:

```go
bft.NewValidatorSetFromABCIValidatorUpdates(liveSet).UpdateWithABCIValidatorUpdates(diff)
```

`valset:current` holds the last accepted proposal as written, and a proposal
may carry a power-0 entry, the removal form `SetValsetProposal` documents, or
the same address twice. `UpdatesFrom` already reads both, keyed by address with
the last entry winning, and the apply drops a validator whose entry is power 0;
the set constructor refuses both. So `liveSet` is `valset:current` with its
power-0 entries skipped. An address listed twice leaves no answer: the encoding
sorts its entries by power, so the stored order no longer says which one
consensus applied, and the gate then skips the dry-run and decides the proposal
as a node without it does. What the constructor still refuses, a power or a
total above the cap, only store damage or a chain-internal bug produces; the
gate pre-checks it and panics with the existing `valset:current corrupted
(chain-internal)` message instead of letting the constructor's panic surface.

On error, whole-reject: log, clear `valset:dirty`, do not write
`valset:current`, return no updates. The rejection is the same shape as the
existing min-floor and pubkey-type rejections, so a future, smaller proposal
can land.

On success, write `valset:current` exactly as before this change: the
proposal, canonically encoded. Storing the dry-run's resulting set instead
would drop a removal entry or a duplicate from the stored bytes, so a node
running this change would commit a different app hash from one that is not,
and the fix could no longer roll out node by node.

Because the call is the apply path itself, it covers `processChanges`'
per-entry bound and duplicate rejection, `verifyUpdates`' running total with
its removal accounting, and `verifyRemovals`. No second, drifting estimator of
the same rules is introduced.

## Alternatives considered

- *Bound each proposed entry by `MaxTotalVotingPower` and the sum of proposed
  powers by the cap* (the proposed patch on the finding): passes
  `current {cap/2, cap/2} → proposed {new: cap}` and the chain still halts.
  The finding's own fix reviewer rejected it for exactly this case.
- *Bound only the sum of proposed powers*: same gap, and it also misses
  `processChanges`' per-entry bound.
- *Clamp or drop the offending entries instead of whole-rejecting*: partial
  application would make `valset:current` disagree with the actual set and
  re-open the divergence the whole-apply design closed.
- *Store the dry-run's resulting set instead of the proposal*: `valset:current`
  would never hold a removal entry or a duplicate, but a proposal carrying one
  would write different bytes than a node without this change writes, which
  moves the app hash and turns the fix into a hardfork.
- *Reject power-0 or duplicate entries at the gate*: refuses a proposal a node
  without this change applies cleanly, the documented removal form among them,
  so a network running both splits on it.
- *Enforce the bound in the realm*: the realm cannot see `valset:current`'s
  relationship to `state.NextValidators`, and the ABCI boundary is where the
  delta is formed.

## Consequences

- A proposal whose delta tm2 would refuse is dropped at EndBlock: the flag is
  cleared, `valset:current` does not advance, and no updates are emitted.
  Consensus applies nothing and the chain keeps producing blocks.
- An in-bounds proposal is unaffected: the dry-run accepts exactly what
  `ApplyBlock` accepts, including a removal-only proposal that lowers the
  total, which `verifyUpdates` never refuses.
- Cost: one set construction, sort and apply over `valset:current` plus at most
  `maxValsetEntries` (100) proposed entries, only on a dirty EndBlock.
  `valset:current` itself is genesis-seeded and count-uncapped, so the base's
  size is the chain's validator count.
- For every proposal the gate accepts, `valset:current` and the returned delta
  are byte for byte what a node without this change produces, so nodes can
  take the change independently. The two differ only on a proposal the gate
  rejects, which a node without it halts on while the agreement below holds.
- The gate assumes `valset:current`, read as `UpdatesFrom` reads it, and
  `state.NextValidators` agree, which the delta computation
  (`currentSet.UpdatesFrom(proposedSet)`) already requires; the dry-run adds no
  new assumption. One case breaks that agreement before this change and still
  does: the encoding sorts a duplicate address's entries by power, so a
  proposal listing the higher power first leaves `valset:current` resolving to
  it while consensus applied the lower. The gate skips its dry-run whenever
  `valset:current` repeats an address, so that disagreement never reaches its
  corruption check, which would otherwise halt every node carrying the change.
  `r/sys/validators/v0`, the one realm `SetValsetProposal` admits, builds each
  proposal from a map keyed by address and never repeats one.
- A power-0 entry stays in `valset:current` until a later proposal replaces it.
  A later proposal that omits that address emits its removal again, which the
  apply refuses for a validator already gone, so the gate rejects that proposal
  where a node without this change halts. `r/sys/validators/v0` builds each
  proposal from the stored set, so it carries the entry forward.
- Regression tests: `gno.land/pkg/gnoland/valset_power_bound_test.go` drives
  the public `EndBlocker` through the existing mock harness, covering a
  single over-cap entry, two entries whose sum crosses the cap, the
  replacement case above, a total exactly at the cap, a removal-only
  proposal, a removal entry stored as proposed, a power-0 entry and a duplicate
  address in `valset:current` that the next proposal must read without
  panicking, including one listed higher power first with that power above the
  cap, a later proposal omitting an address stored at power 0, a power above
  the cap and a total above the cap in `valset:current` reported as
  corruption, and a `valset:current` totalling exactly the cap read by the
  next proposal. `gno.land/pkg/integration/testdata/params_valset_proposal_power_above_cap.txtar`
  runs a GovDAO proposal above the per-validator cap and one above the total
  cap on an in-memory node, and checks the chain keeps producing blocks after
  each.
