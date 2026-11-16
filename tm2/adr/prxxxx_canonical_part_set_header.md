# ADR: Tying A Proposal's PartSetHeader To The Block It Chunks

## Context

`BlockID.PartsHeader` holds a part count and the merkle root over the parts,
and it is half of a block's identity. Nothing on the receive path ties it to the block it
chunks.

`defaultSetProposal` in `tm2/pkg/bft/consensus/state.go` installs the
proposer's header verbatim, `types.NewPartSetFromHeader(proposal.BlockID.PartsHeader)`,
and checks the signature, the height, the round and the POL round above it.
`PartSet.AddPart` verifies each part's merkle proof against that same
proposer-chosen root, so every part of a fabricated chunking verifies. On
completion `addProposalBlockPart` decodes the concatenation with
`amino.UnmarshalSizedReader(..., Block.MaxDataBytes)`, which reads the varint
length prefix and exactly that many bytes: the same block bytes cut at
different boundaries, and the same block bytes with trailing padding, both
decode into the block unchanged.

Two more encodings reach a vote. A block encoded with `DataHash` or
`LastCommitHash` left out decodes with that field empty, and `Block.Hash()`
refills it in place through `fillHeader`, so the block hashes as the one the
proposal names while its parts are the chunking of the unfilled encoding. An
unknown field after the block's last field, inside the length prefix, fails
the decode only after the block is filled.

From there the fabricated header is signed and stored. `defaultDoPrevote`
signs `cs.ProposalBlockParts.Header()`, `enterPrecommit` carries the prevote
majority's `BlockID` forward, and `finalizeCommit` asserts only
`blockParts.HasHeader(blockID.PartsHeader)`, the header against itself,
before `SaveBlock` writes it into the BlockMeta and `ApplyBlock` takes it as
the next height's `LastBlockID`.

`Block.MakePartSet(BlockPartSizeBytes)` runs on the proposer's own block and in
fast sync, never on a received one, so the invariant "the committed
`PartsHeader` is the canonical chunking of the block" is assumed everywhere and
asserted nowhere.

The consequence is permanent. Fast sync recomputes the header locally, in
`tm2/pkg/bft/blockchain/reactor.go`:

```
firstParts := first.MakePartSet(types.BlockPartSizeBytes)
firstPartsHeader := firstParts.Header()
firstID := types.BlockID{Hash: first.Hash(), PartsHeader: firstPartsHeader}
err := state.Validators.VerifyCommit(
	chainID, firstID, first.Height, second.LastCommit)
```

`VerifyCommit` rejects on `!blockID.Equals(commit.BlockID)`, and the commit is
the one +2/3 of validators signed, so it carries the fabricated header. The
next block's `LastBlockID` carries it too, so `validateBlock` also refuses any
state fast sync builds. The commit is history: a fresh node, a restore from
backup and a validator that fell behind are all stuck below that height for
good.

A second freedom rides on the same gap. `PartSetHeader.ValidateBasic` bounds
`Total` by `MaxBlockPartsCount`, 1601 parts derived from the legacy 100MB
`MaxBlockSizeBytes`, while the chain commits at most `Block.MaxDataBytes`, which
is 2,000,000 bytes by default and 8MB at the `MaxBlockDataBytesLimit` ceiling. A
proposal may therefore declare 100MB of parts against a 31-part chain, and
every node buffers what arrives.

## Decision

The checks go in `ConsensusState`, the one place that knows the chain's
consensus params and the decoded block.

### 1. Neither a proposal nor a polka installs more parts than a block fills

A proposal declaring more than `ceil(Block.MaxDataBytes / BlockPartSizeBytes)`
parts is refused with `ErrInvalidProposalPartSetHeader` before any part set is
allocated. That ceiling is exact rather than approximate because
`CreateProposalBlock` trims transactions until `parts.ByteSize()`, the
length-prefixed encoding peers decode, fits `MaxDataBytes`, so it is the most
parts an honest proposer can produce.

A prevote majority installs a part set for its header in two more places,
`enterPrecommit` and `addVote`. Both skip a header over the same ceiling and
keep the part set the node holds, so +2/3 prevotes for a header no whole block
fills cost the node no more than a proposal declaring it. `enterCommit` stays
unbounded: +2/3 have precommitted that header, and the node has to fetch those
parts to finalize the height. A WAL replay lifts the ceiling to
`MaxBlockPartsCount`, per decision 4.

### 2. `addProposalBlockPart` requires the header to be the block's own chunking

Once the set completes and decodes, the block must pass `ValidateBasic` as
decoded, before anything hashes it, and `cs.ProposalBlockParts.Header()` must
equal `cs.ProposalBlock.MakePartSet(types.BlockPartSizeBytes).Header()`.
`ValidateBasic` refuses an empty `DataHash` or `LastCommitHash` that the block's
data or last commit hashes to something else. On a failed decode or either
check the block is dropped, `cs.ProposalBlock = nil`, so the node prevotes nil,
the round times out, and nothing signs or stores the fabricated header, except
under the lock decision 4 restores.

The checks are skipped in the commit step. By then +2/3 have precommitted the
header these parts carry, so refusing the block protects no vote and only
strands a node that reached the commit step before the parts arrived, a
catching-up or lagging one, at that height for good. A node that dropped the
block earlier, in the propose step, keeps its complete part set, which
`enterCommit` resets only on a different header, and no further part arrives
to decode it again. `enterCommit` therefore decodes the block from those parts
once +2/3 precommit their header. In both places a decode that fails after
filling the block keeps it when it hashes to the committed hash and
`state.ValidateBlock` accepts it, which is what an unpatched node commits.

### 3. A block counts as the committed one only under the committed header

`enterCommit` takes the locked block, or the block it holds, only when its parts
carry the header +2/3 precommitted; otherwise it fetches the parts under that
header. When a prevote majority names a block the node holds under other parts,
`addVote` drops the block as it replaces the part set, for a header within the
ceiling of decision 1. A hash-only comparison at either place hands
`finalizeCommit` a part set that is not the commit's, which halts the node with
`Expected ProposalBlockParts header to be commit header`, or an empty one, which
halts it in `SaveBlock`. Both predate this change.

`enterPrecommit` relocks, or locks the block it holds, only under the parts the
prevote majority names. A majority for the same block under other parts unlocks
the node and gets a nil precommit, so a patched node never signs a header its
own checks refused, except under a lock an unpatched binary took, below.

When `enterPrecommit` replaces the part set with the majority's header, it fires
`EventNewValidBlock`, as `addVote` and `enterCommit` do.
Peers send parts only for the header a node announced, so a node that met the
polka before the proposal would otherwise wait in the commit step for parts no
peer sends it.

### 4. A lock the unpatched binary took survives the restart

A validator restarted onto this binary mid-height replays its WAL through the
same checks, so it drops a padded block the unpatched binary locked on and
holds no lock on it once the replay ends. Its own precommit for the block is
still replayed, so in a later round it would prevote another block after
precommitting this one, breaking the lock rule.

`addProposalBlockPart` therefore keeps the last complete part set whose block
it refused, until the height ends, and `addVote` locks again on the block the
node's own precommit names, decoded from those parts when they carry the
precommit's header and the decoded block hashes to the precommit's block. The
lock rule then has the node prevote the padded block in later rounds, under the
header it already precommitted. The node locks only on
a precommit from a round after its current lock, and only while no round after
the precommit's, up to the current one, holds a prevote majority for another
block or for nil, since such a polka unlocked the unpatched binary. That holds
when the own precommit is handled after the round moved on, which happens when
the unpatched binary skipped to the next round on +2/3 of that round's
prevotes before handling its own vote. A nil precommit of its own and another
validator's precommit lock nothing. Only one refused part set is kept per
height, so a second refused proposal arriving before the node handles its own
precommit for the first leaves that precommit nothing to lock on.

The unpatched binary takes a block past decision 1's ceiling, up to
`MaxBlockPartsCount` parts, so `maxBlockParts` returns that bound while the WAL
replays. The parts of such a block, from its proposal or from a polka ahead of
it, then reach the kept set too. A node that refused such a proposal live still
comes back unlocked, since only its own replayed precommit locks.

Re-deriving the header from the decoded block assumes
`marshal(unmarshal(b)) == b`. That identity is already load-bearing: fast sync
recomputes every block's header the same way before verifying its commit.

## Alternatives considered

**A. Re-derive `MaxBlockPartsCount` from `MaxBlockDataBytesLimit`.** Rejected
as the fix. It bounds the coarse global ceiling, 8MB, rather than the chain's
actual `MaxDataBytes`; it moves the shape-level fixtures added for the bound
itself; and it says nothing about re-chunking, which is where the permanent
damage is.

**B. Check the header in `finalizeCommit`.** Too late. A +2/3 majority has
signed it by then, so the check saves this node and leaves the commit on the
chain. Refusing at the part-set level means the header never reaches a vote.

**C. Bound the assembled byte size inside `PartSet.AddPart`.** The part set
does not know the chain's consensus params, and a size bound alone still admits
a smaller non-canonical chunking.

**D. Keep the block and vote the canonical header the node computes itself.**
This preserves the round against a byzantine proposer, but the node's own part
set no longer matches what it voted, so `finalizeCommit`'s `HasHeader`
assertion, `ValidBlockParts` and the gossip path all have to be re-chunked
locally. Dropping the block costs one round; this costs a redesign of the
proposal path.

**E. Hash the decoded block before re-chunking it.** `Hash()` refills an empty
`DataHash` or `LastCommitHash`, so the re-chunk then differs from the unfilled
parts and refuses them too. `ValidateBasic` refuses the block before anything
changes it, and its error names the field.

**F. Skip the checks while replaying the WAL.** The restarted node would come
back locked as the unpatched binary left it, but a patched node that refused
the block live, with its own nil prevote and nil precommit in the WAL, would
then come back locked on the padded block it refused.

**G. Lock again only in the round the own precommit was signed, from the
round's part set.** No new field and no scan over rounds, but an own precommit
handled after a round skip finds the round's parts cleared and restores
nothing, though no polka unlocked the block, so the node prevotes the next
round's proposal.

**H. Match the own precommit by the hash of the block in the kept parts, and
count only a polka for another hash as an unlock.** That is the unpatched
binary's own rule, and it closes the two orderings in Consequences that the
header match misses. Declined on a reading of the code, not a run: a patched
node that refused a block under one chunking, then locked and precommitted it
under its canonical parts, and was unlocked by a polka for the same block
under other parts, would lock again on the refused parts if it handled its own
precommit after that polka, and prevote them.

**I. Bound `enterCommit`'s install too.** Every install a peer header reaches
would then sit under one ceiling. A node would wait for good at a height +2/3
precommitted under a header over the ceiling, which an unpatched majority
commits during the rollout, since the unpatched binary has no ceiling.

**J. Leave the polka installs unbounded and record the gap.** No change to the
vote path. A prevote majority, unpatched validators during the rollout or more
than 1/3 byzantine, then installs up to `MaxBlockPartsCount` parts, about 105MB,
and a refused set stays in `droppedParts` until the height ends. That costs
what the base costs, which takes the same parts on the proposal path too.

**K. Leave the announce unfixed.** No new event. A node that meets a polka for
the block under other parts ahead of the proposal then waits in the commit step
for parts no peer sends it. The base fails the same ordering too, by panicking
in `finalizeCommit` on a part set that is not the commit's.

**L. Set `ValidBlock` on the lock decision 4 restores.** It only helps a
restored node that proposes in a later round of that height: it would
re-propose its locked block, where it now builds a fresh one that validators
locked on the block do not prevote, so that round is likely lost.

**M. Reuse the kept part set in `enterCommit`.** A node that refused the block
and moved to a later round would decode it from the parts it kept, rather than
fetch every part again once +2/3 precommit them. It saves only that
re-download; the node commits once peers send the parts again.

## Consequences

- A byzantine proposer can no longer commit a non-canonical header. Its round
  proceeds as if the proposal never arrived.
- A peer-supplied part set costs at most one block's worth of memory, from a
  proposal or a prevote majority, instead of up to 100MB. Only +2/3 precommits
  install a larger one, and the node then stores what they committed, as the
  base does. The consensus WAL is not
  bounded by this: `cs.wal.Write(mi)` runs on every peer message before
  `handleMsg` sees it, which is a proposal-independent path and untouched here.
- One `ValidateBasic`, one extra `MarshalSized` and a merkle root per completed
  proposal block outside the commit step, the same work the proposer already
  did, against a decode of the same bytes. A node keeps at most one refused
  block's parts until the height ends, bounded by decision 1, or after a WAL
  replay by `MaxBlockPartsCount`, what the unpatched binary holds live.
- The rule is new, so it only holds once more than 1/3 of voting power runs
  it: a patched node holds no block, never locks and precommits nil, so the
  rest cannot reach +2/3 and such a round fails rather than commits. It
  repairs no chain already attacked: such a height stays unverifiable by fast
  sync whatever a node runs. Honest proposals are canonical by construction, so
  patched and unpatched nodes agree on every block an honest proposer makes.
- A partial rollout opens one window for a byzantine proposer. With patched
  power under 1/3 and unpatched power at most 2/3, unpatched nodes can lock on
  a padded block that patched nodes keep refusing, and no nil polka can unlock
  them, so the height commits only when the byzantine validators vote with the
  unpatched ones. Patched power above 1/3 closes the window, since the lock
  then never forms. An unpatched validator locked that way and restarted onto
  this binary keeps its lock, per decision 4.
- Decision 4 matches the own precommit by its parts header, while the
  unpatched binary relocks and keeps its lock by block hash alone. Two
  orderings therefore bring a restarted validator back unlocked although the
  unpatched binary held the lock. In the first, it locked on a block under one
  chunking, relocked on a polka for the same block under a second chunking
  whose parts it never completed, and precommitted that second header, which
  matches no kept parts. In the second, its own precommit is handled after a
  polka for the same block under other parts, which `polkaAfter` counts as an
  unlock. Both need a byzantine proposer chunking one block two ways, +2/3
  prevoting the second chunking, and the validator restarted onto this binary
  mid-height. Alternative H closes both and was declined.
- A patched node commits what +2/3 committed, so it stores the parts an
  unpatched node stores. `BlockStore.LoadBlock` decodes padded,
  trailing-field and unfilled records as consensus does, and panics on a
  record whose block does not hash to its meta.
- `PartSetHeader.ValidateBasic`'s `MaxBlockPartsCount` bound stays as the
  only bound on `enterCommit`'s install, alternative I; the consensus bound is
  the one that tracks the chain's real limit.
- Nothing here compares the decoded block's hash against
  `cs.Proposal.BlockID.Hash`, which is still unread in `state.go`. The pair a
  node votes remains self-consistent and canonical either way, so it is a
  separate question.
