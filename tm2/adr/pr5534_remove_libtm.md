# PR5534: Remove libtm — integration analysis and verdict

## Context

`tm2/pkg/libtm` is a minimal, whitepaper-faithful implementation of the
Tendermint consensus algorithm
([paper](https://arxiv.org/pdf/1807.04938.pdf)). It has been sitting in-tree
as a vendored sibling Go module (`github.com/gnolang/libtm`) since
introduction, with the stated long-term intent of replacing the consensus
core inside `tm2/pkg/bft/consensus` and eventually being extracted to its
own repo.

Current state before this analysis:

- `libtm` is its own Go module, unreachable from `github.com/gnolang/gno`.
- Its generated `messages.pb.go` is not committed (matches the repo-wide
  `*.pb.go` gitignore), so it doesn't build from a fresh checkout without
  running `protoc`.
- Zero imports from any other package in the monorepo. Effectively dead
  code at the moment.

`tm2/pkg/bft/consensus` carries the current, production implementation:

- `state.go` — 1811 lines
- `reactor.go` — 1646 lines
- `types/height_vote_set.go` — 259 lines
- plus `round_state.go`, `peer_round_state.go`, `wal.go`, `replay.go`, etc.

This ADR answers a single question: **does it make sense to replace the
consensus core in `tm2/pkg/bft/consensus` with `libtm`?**

## Surface comparison

`libtm` exposes a very small surface:

- `core.Tendermint.RunSequence(ctx, height) *FinalizedProposal` — runs a
  single consensus height.
- Four caller-implemented interfaces: `Verifier`, `Node`, `Broadcast`,
  `Signer`.
- Three message types: `ProposalMessage`, `PrevoteMessage`,
  `PrecommitMessage` (protobuf).
- State: atomic step (propose/prevote/precommit), `lockedValue` and
  `validValue` stored as opaque `[]byte`, no block internals.

`tm2/pkg/bft/consensus` does substantially more.

## What tm2 does that libtm does not

1. **Block parts / chunked proposal delivery.**
   `state.go:1441 addProposalBlockPart()` reassembles large blocks from
   `types.PartSet` chunks (256KB default). `reactor.go:452–498
   gossipDataRoutine()` actively gossips missing parts. libtm assumes
   proposals arrive as a single `[]byte`; it has no concept of parts.

2. **Write-ahead log and crash recovery.**
   `walm.WAL` persists all consensus messages (proposals, votes, timeouts)
   before state transitions (`state.go:21, 127–130, 166–169, 292–309`) and
   replays them on startup (`state.go:323–330`). libtm has no persistence
   whatsoever.

3. **Evidence pool / double-sign detection.**
   `state.go:1519` detects conflicting votes from the same validator and
   routes them to `evpool.AddEvidence()` for slashing. libtm has no
   evidence handling.

4. **Lock and Valid Block semantics with POL.**
   `round_state.go:78–84` tracks `LockedBlock`/`LockedBlockParts` and
   `ValidBlock`/`ValidBlockParts` as first-class typed state, with
   explicit `POLRound` in `types.Proposal` (`types/proposal.go:27`).
   `state.go:926–928` falls back to `ValidBlock` if the current proposal
   is invalid. libtm stores `lockedValue`/`validValue` as bare `[]byte`
   with no block identity or POL round.

5. **Wall-clock timeouts tied to WAL / RPC.**
   `state.go:114, 163` uses a `TimeoutTicker` whose `timeoutInfo` is
   logged to the WAL and replayed on startup. `round_state.go:74–75`
   exposes `StartTime`/`CommitTime` for RPC and UI. libtm's timeouts
   are in-memory only.

6. **Peer round-state tracking and gossip.**
   `reactor.go` (~1646 lines total) maintains per-peer `PeerState` and
   drives gossip: block parts, POL bit arrays, has-vote notices, catchup.
   libtm has no peer state; all P2P is delegated to the caller as a flat
   `Broadcast` interface.

7. **Dynamic validator sets and state sync.**
   `state.go:108` holds a persistent `sm.State` with `Validators` /
   `LastValidators` per height, updated via ABCI `EndBlock`. libtm has no
   validator-set management; `Verifier.IsValidator()` is per-call.

8. **Commit persistence and block store integration.**
   `state.go:1332–1358` saves finalized blocks to `blockStore` and logs a
   `MetaMessage{Height+1}` to the WAL as a recovery marker. libtm's
   `RunSequence` returns the `FinalizedProposal` immediately and leaves
   persistence to the caller.

9. **Mempool integration for propose timing.**
   `state.go:81–83` wires a `txNotifier` from the mempool to avoid
   sleeping through the propose timeout when no txs are available. libtm
   has no hook for this.

## Type-boundary impedance

libtm uses `[]byte` for proposals, IDs, and senders. tm2 uses rich typed
structures:

| Field | tm2 | libtm |
|-------|-----|-------|
| Proposal | `types.Proposal{Height, Round, POLRound, BlockID{Hash, PartsHeader}, Signature}` | `[]byte` |
| BlockID | `{Hash []byte, PartsHeader{Total, Hash}}` | `[]byte` |
| Vote | `types.Vote{Height, Round, BlockID, Timestamp, ValidatorIndex, ValidatorAddress, Signature}` | flat `[]byte` sender |
| Block | `types.Block{Header, Data, LastCommit}` | `[]byte` |
| Validator address | `crypto.Address` (20B) | `[]byte` |

Integration would have to embed amino-encoded tm2 types inside libtm's
`[]byte`, which forces tm2 to dual-track the structured copy alongside the
opaque copy (for RPC, evidence, WAL, block-parts gossip, and POL
reconstruction). The abstraction boundary erases the very information tm2
needs at every other layer.

## Algorithmic differences

Both implement the Tendermint paper, but tm2 layers on production safety
properties that libtm (as a reference implementation) does not:

- **Lock & Unlock** semantics at the block level.
- **Proof-of-Lock** enforcement across rounds (not just the
  `lockedRound`/`validRound` counters the paper specifies).
- **Valid Block reuse** — hold the block object itself across rounds, not
  just the `[]byte`, so block-parts gossip can continue.

Integrating libtm would require **forking it** to add these — i.e. the
"reference" nature of libtm is lost the moment we integrate it.

## Line-count estimate

| | lines |
|---|---:|
| Removed from `state.go`+`reactor.go` core state machine | −500 to −600 |
| Added: WAL wrapper, block-parts adapter, evidence router, validator-set adapter, RPC type adapters, the 4 libtm interfaces | +800 to +1000 |
| **Net** | **+200 to +400** |

Integration makes tm2 **larger**, not smaller.

## Decision

**Do not integrate `libtm` into `tm2/pkg/bft/consensus`.**

The two codebases sit at different abstraction layers: `libtm` is a
whitepaper-faithful reference engine; `tm2/pkg/bft/consensus` is a
production consensus stack with WAL, evidence, block parts, dynamic
validator sets, RPC interfaces, and safety extensions beyond the paper.
`libtm`'s `[]byte`-based abstractions can't carry the structured data
tm2's other layers need, and most of tm2's complexity is outside libtm's
scope anyway. Integration would increase code, not reduce it.

## Alternatives considered

- **Fold libtm into the main module and wire it up incrementally.**
  Rejected: see line-count estimate and impedance section above.
- **Use libtm only to replace `HeightVoteSet`.** Rejected: surfaces look
  similar (~260 vs ~180 lines) but semantics diverge (libtm's collector
  is generic dedup-by-sender, `HeightVoteSet` bakes in +2/3 majority, POL
  tracking, peer catchup rounds, validator voting power, amino JSON).
  Downstream rewrites would dwarf any savings.
- **Fork libtm to add tm2's features.** Rejected: this defeats libtm's
  stated purpose as a minimal, paper-faithful reference. The fork would
  diverge and libtm would no longer be extractable to a standalone repo.
- **Refactor `tm2/pkg/bft/consensus` directly** — separate state machine,
  WAL, gossip, and ABCI into cleaner packages without adopting libtm as a
  dependency. Out of scope for this ADR but a more promising direction
  if simplification is the goal.

## Recommendation for `tm2/pkg/libtm/`

Given the "do not integrate" verdict, `tm2/pkg/libtm/` is now known-dead
code: vendored, unbuildable from a fresh checkout (missing `.pb.go`),
imported by zero packages, and with no credible path to activation.

Three options, ordered by preference:

1. **Delete it.** Clean up dead code. Git history preserves it for anyone
   who wants to resurrect the standalone-repo plan. Also removes the
   misleading README note that suggests integration is pending. Net diff:
   roughly **−2500 lines** including tests.

2. **Extract it now** to `github.com/gnolang/libtm` as originally
   envisioned. The README's "we'll extract once integration stabilizes"
   rationale is moot — integration isn't happening, so there is no reason
   to hold the code hostage in this monorepo. Requires maintainer
   coordination and a new repo.

3. **Keep it** as a reference / test harness. Accept the maintenance
   drift (the `.pb.go` rot, the dual-module awkwardness, the stale README
   note). Only makes sense if someone plans to use it as a test oracle or
   teaching artifact, which nobody currently does.

**This PR recommends option 1** (delete). If maintainers prefer 2 or 3,
the outcome should be recorded in a follow-up ADR and the misleading
README note revised accordingly.

## Consequences

- `tm2/pkg/bft/consensus` remains the consensus core. No behavior
  changes.
- Pending maintainer input on the libtm-fate question, follow-up PRs
  would either delete `tm2/pkg/libtm/` or coordinate the extraction.
