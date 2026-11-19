# ADR: Refuse A Valset Proposal In The Transaction That Writes It

## Context

`EndBlocker` dry-runs `valset:proposed` through the apply call tm2's
`ApplyBlock` makes, and drops the whole proposal when it errors (see
`pr6504_endblock_valset_power_bound.md`). That stops a refused proposal from
halting the chain, but `valset:proposed` is not one proposal: every valset
write in a block builds on the one before it, since `r/sys/validators/v0` reads
`GetValsetEffective`, which returns `valset:proposed` while `valset:dirty` is
set. Dropping the proposal at EndBlock drops every valset write of the block.

A key rotation is the costly one. `valopers.UpdateSigningKey` commits the
rotation in the realm before EndBlock runs: the operator's profile and v0's
cache point at the new key, and `signingRegistry` retires the old key for good.
When EndBlock then drops the block's proposal, consensus keeps the old key.
From there:

- a GovDAO removal of the operator resolves its new key, which consensus does
  not hold, and panics `validator does not exist`;
- a further rotation looks up the new key in the set, finds nothing, and
  records `applied=false`;
- `valopers` refuses to register a retired key again, so the operator cannot
  rotate back.

The old key stays a validator that no v0 call can remove. A GovDAO valset
proposal executed in the same block is lost the same way, while GovDAO records
it as executed.

## Decision

Run the same dry run when `valset:proposed` is written. In
`nodeParamsKeeper.WillSetParam`, after the entries parse, read
`valset:current`, compute `currentSet.UpdatesFrom(proposedSet)` as EndBlocker
does, and pass it to `dryRunValsetDiff`, the helper EndBlocker now calls too.
An error panics with `valset:proposed would fail consensus apply: <err>`, so
the transaction that wrote the proposal reverts with its realm state, and the
block's other valset writes are untouched.

`valset:current` changes only at EndBlock, so a write's dry run sees the base
EndBlocker will see. The block's last write, the one EndBlocker reads, passes
both checks or neither. Both read the base through `loadValsetCurrent`.

The same hook refuses a proposal that repeats an address. The dry run accepts
one, since `UpdatesFrom` keeps one entry per address, but once EndBlocker
stores it in `valset:current`, `liveValset` can no longer rebuild the set and
both dry runs are off for every later write. v0 never sends a repeat.

EndBlocker keeps its gate as the fallback for a write that reaches the store
without `WillSetParam`.

A refused write fails the transaction that makes it.

## Alternatives considered

- *Keep the check at EndBlock only and document the residual*: keeps the
  change a rolling upgrade, and leaves a dropped rotation's old key in
  consensus for good.
- *Have `valopers` retire the old key only once consensus applies the
  rotation*: needs EndBlock to call back into the realm, and a realm redeploy.
- *Let v0 remove an operator by any key it has held*: a realm redeploy, and a
  GovDAO change dropped with the block stays lost.
- *Check in the realm*: Gno cannot call tm2's apply code, so the realm would
  carry a second copy of its rules.
- *Move EndBlocker's pubkey-type allow-list and min-validator floor to write
  time too*: `valopers` refuses a disallowed key type at registration and
  rotation, and v0 refuses to publish an empty set, so neither reaches
  EndBlock from v0, and moving them only widens the consensus change.

## Consequences

- Consensus change. A `valset:proposed` write tm2 would refuse now fails its
  transaction. A node without this change accepts that transaction, then halts
  at `ApplyBlock`, unless a later write in the same block replaces the set; in
  that case the two commit different results. Deploy as a coordinated upgrade.
- A proposal with no positive-power entry also fails its transaction now, since
  tm2's apply refuses an empty set, where EndBlocker's floor used to drop it at
  EndBlock. v0 already refuses to publish an empty set.
- Each `valset:proposed` write reads `valset:current` through the params
  keeper, and the transaction pays the read's gas. No gas golden moves.
- Regression tests: `TestNodeParamsKeeperValsetProposedDryRun` in
  `gno.land/pkg/gnoland/valset_power_bound_test.go` writes a replacement at
  the cap over the old set's power and a repeated address, and checks each
  refusal's message, then an in-bounds proposal.
  `params_valset_proposal_power_above_cap.txtar` now expects `ExecuteProposal`
  to fail on both over-cap proposals, delivered in a block rather than refused
  by simulation, and the chain to keep producing blocks.
  `params_valset_rotation_beside_refused_proposal.txtar` rotates a
  validator's key and then executes an over-cap proposal in one delivered
  transaction, checks that both revert, then that the rotation alone lands and
  GovDAO can still remove the operator.
