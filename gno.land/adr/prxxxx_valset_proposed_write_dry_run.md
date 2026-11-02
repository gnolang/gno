# ADR: Refuse A Valset Proposal In The Transaction That Writes It

## Context

`EndBlocker` dry-runs `valset:proposed` through the apply call tm2's
`ApplyBlock` makes, and drops the whole proposal when it errors (see
`prxxxx_endblock_valset_power_bound.md`). That stops a refused proposal from
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

A refused write aborts the GovDAO transaction executing the proposal, so the
proposal stays open, where before this change its first execution closed it,
and any account could execute it later once the set had room for it. No Gno
code can turn the refusal into an error that `ExecuteOrRejectProposal` would
close the proposal on: a panic from a native param write aborts the
transaction even inside `r/sys/params`. So `r/gov/dao/impl/v0` gives every
proposal an execution deadline instead. `VoteOnProposal` records `PassedAt`,
the height of the vote that first brings YES to supermajority, and the
proposal expires `executionDeadlineBlocks` (170,000, one week at gnoland-1's
3.55 s blocks) after it. One `expired` check serves three places: a vote on an
expired proposal is refused with `proposal expired`, the proposal renders as
`EXPIRED`, and an execution attempt closes it as denied, with reason
`expired: not executed within 170000 blocks of passing`.

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
- *Close a refused proposal by letting members change their vote, or by
  withdrawal*: both need members to act on every stuck proposal, and vote
  changes alter voting for every proposal. The deadline closes it unattended.
- *Keep the deadline in `Law`*: a law-update request built without the new
  field would set it to zero and expire every open proposal.
- *Count the deadline from creation*: gnoland-1's proposal 6 reached its
  passing vote at least 757,864 blocks (30 days) after creation, so a week from
  creation would have closed it before it could pass.

## Consequences

- Consensus change. A `valset:proposed` write tm2 would refuse now fails its
  transaction. A node without this change accepts that transaction, then halts
  at `ApplyBlock`, unless a later write in the same block replaces the set; in
  that case the two commit different results. Deploy as a coordinated upgrade.
- A proposal with no positive-power entry also fails its transaction now, since
  tm2's apply refuses an empty set, where EndBlocker's floor used to drop it at
  EndBlock. v0 already refuses to publish an empty set.
- A refused GovDAO proposal stays executable until its deadline, and any
  account can execute it once the set has room for it within that week.
- Every passed GovDAO proposal must now execute within 170,000 blocks of its
  passing vote, about a week at gnoland-1's current pace; a slower or faster
  chain moves the wall-clock length. Voting has no deadline. A proposal that
  reaches supermajority without a vote, because members left, or that passes
  at height 0, records no `PassedAt` and has no deadline.
- A YES vote now computes the YES tally to record `PassedAt`, which walks the
  member list: about 2.9M more gas on a one-member DAO (16.5M to 19.4M in
  `govdao_proposal_add_member.txtar`). NO and ABSTAIN votes cannot pass a
  proposal and skip it. Eighteen txtars' YES-vote gas limits rise by 4M.
- The change ships as a new GovDAO implementation deployed by transaction and
  switched in by GovDAO; statuses live on the implementation instance, so the
  deadline applies to proposals created under the new one. Deploy it at its own
  package path. A fork genesis that patched this `impl/v0` in place would
  replay gnoland-1's history under the deadline at the source heights.
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
  `examples/gno.land/r/gov/dao/impl/v0/z_execution_deadline_test.gno` executes
  a passed proposal at its deadline and closes one a block past it without
  running its executor, executes a proposal whose deciding vote came a deadline
  after creation, and refuses a vote on an expired proposal that renders as
  expired.
