# ADR: Keep ProposalBlockParts consistent with ProposalBlock under equivocating proposals

## Context

A block's `PartSetHeader` is `(Total, merkle root of the chunked serialized
block)`. The same block bytes can be chunked in more than one way, so two
proposals can carry the same block hash but different parts headers. A
byzantine round proposer can send one node the block under one cut and the
rest of the validators the same block under another, so the votes the node
sees name a `BlockID` whose parts header differs from the parts it holds.

Before the canonical part-set check, `addVote` then kept `ProposalBlock` beside a fresh, empty part set
for the voted header, and `finalizeCommit` handed that incomplete set to
`SaveBlock`, which panicked and stopped the node's consensus until a restart.

The canonical part-set check (`pr6504_canonical_part_set_header.md`) closed that halt. Outside the
commit step a node now drops any completed part set that is not
`MakePartSet(BlockPartSizeBytes)` of its block, `addVote` and `enterPrecommit`
drop the held block when they swap in another header's part set, and
`enterCommit` moves the locked pair to the proposal only when it carries the
committed header.

Three places still compare a held block with a vote by hash alone:

- `addVote`'s unlock rule keeps a lock when a later polka names the locked
  block under another header.
- `enterPrecommit` unlocks a node locked on another cut of the polka's block,
  though when the polka names the `BlockPartSizeBytes` cut it could rebuild
  those parts from the block. Such a lock is what `relockOnOwnPrecommit`
  restores from a WAL an earlier binary wrote.
- `addVote`'s `Valid*` update takes the held block with parts the polka did
  not name.

## Decision

1. A held block's parts are rebuilt only as `MakePartSet(BlockPartSizeBytes)`,
   and only when the vote names that cut. Fast sync re-chunks at
   `BlockPartSizeBytes`, and the canonical part-set check refuses every other cut before a vote, so a
   node must not sign one.
2. `repairLockedParts`: when a polka names the locked block under another
   header, swap in that rebuild if the polka names it, moving
   `ValidBlockParts` along when the valid block is the locked one, or unlock.
   `enterPrecommit` calls it before the relock, and `addVote`'s unlock rule
   calls it, so that rule now compares `BlockID`s rather than hashes.
3. `repairProposalParts`: the same for the proposal pair, in `addVote`'s
   `Valid*` update and in `enterCommit`, dropping the block where it cannot
   rebuild. In `addVote`, a header with more parts than a whole block fills
   keeps the block, as the canonical part-set check's swap there does.

## Alternatives considered

- *Also rebuild the equal cut* `ceil(len/Total)`: a node locks on, precommits
  and later proposes a cut the canonical part-set check refuses, and its own `addProposalBlockPart`
  then refuses the proposal it signed.
- *Repair in `addProposalBlockPart`*: a polka for any cut it cannot rebuild
  discards the complete canonical block the node just received, and installs
  part sets past that check's ceiling. `enterPrecommit` already swaps to the polka's
  header, under the ceiling.
- *Rebuild the committed cut from `LockedBlock` in `enterCommit`*: the canonical part-set check's guard
  moves the locked pair only when it carries the committed header, so the
  rebuild never runs.
- *Have `finalizeCommit` tolerate an incomplete part set*: it would store an
  incomplete block.

## Consequences

- With every validator running the canonical part-set check, no polka names a cut other than
  `BlockPartSizeBytes`. The repairs then act on a lock `relockOnOwnPrecommit`
  restored on another cut, which moves to the canonical cut instead of
  unlocking.
- During a rollout, a polka for another cut from earlier binaries makes the
  node drop its held block or release its lock and precommit nil. It never
  signs that cut.
- Cost: one `MarshalSized` of the block and one merkle tree per header
  mismatch at each call site; each site switches itself off afterwards.
- Regression tests in `tm2/pkg/bft/consensus/byz_parts_header_test.go`. With
  the equal-cut rebuild restored, `TestByzProposerEquivocatingPartsHeaderEqualCut`
  fails; with the ceiling check removed from `repairProposalParts`,
  `TestByzProposerEquivocatingPartsHeaderPolkaPastCeiling` fails; with the
  repair restored in `addProposalBlockPart`, the canonical part-set check's
  `TestStatePolkaAheadOfProposalAnnouncesParts` fails.
