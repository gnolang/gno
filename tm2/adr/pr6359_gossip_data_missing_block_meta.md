# ADR: Serving A Peer That Names A Height The Blockstore Never Held

## Status

Proposed.

## Context

`gossipDataRoutine` in `tm2/pkg/bft/consensus/reactor.go` helps a peer that is
behind. When the peer's height sits between 1 and the node's own, and the peer
has sent no block parts yet, the routine reads
`blockStore.LoadBlockMeta(prs.Height)` to learn the part-set header to offer,
and it panicked when that read came back nil.

`prs.Height` is whatever the peer said it was. `NewRoundStepMessage` arrives on
`StateChannel`, which `Receive` handles before the `FastSync()` gate that guards
`DataChannel` and `VoteChannel`. `NewRoundStepMessage.ValidateBasic` checks that
the height is not negative, that the round is not negative, that the step is one
of the declared values, and that `LastCommitRound` is at least -1 above height 1.
It does not compare the height against anything local.
`ApplyNewRoundStepMessage` then copies the value into `ps.PRS.Height` and, since
the height changed, sets `ps.PRS.ProposalBlockParts` to nil, which is the branch
holding the read.

A blockstore holds no block below `InitialHeight`. `MakeGenesisState` sets
`LastBlockHeight` to `InitialHeight - 1`, so a chain starting at 100 opens with
a consensus height of 100 and an empty store, and heights 1 to 99 are never
produced. `BlockStore.SaveBlock` accepts any height as the first one for exactly
this reason. `gno.land/pkg/gnoland/app.go` records that dev and test networks
run `InitialHeight > 1` outside a hardfork, and
`gno.land/pkg/gnoland/testdata/slim_genesis.json` carries an `initial_height` of
815001 against a `past_chain_ids` of `gnoland1`.

The routine starts as a bare `go conR.gossipDataRoutine(peer, peerState)` in
`AddPeer`, with no `defer`/`recover` of its own. A panic there is not the
`CONSENSUS FAILURE!!!` path in `receiveRoutine`, which recovers; it unwinds the
whole process.

Three sibling reads take the same peer-supplied height to the blockstore.
`gossipDataForCatchup`, ten lines below, logs at error and sleeps on a nil
`LoadBlockMeta`, and again on a nil `LoadBlockPart`. `queryMaj23Routine` logs at
warn and loops without sleeping on a nil `LoadCommit`, but that branch is
unreachable: `CatchupCommitRound` leaves -1 only after a commit for
`prs.Height` has been loaded and sent, and `ApplyNewRoundStepMessage` resets it
on every height change.

`gossipVotesRoutine` is the fourth, and its nil branch loops with no sleep at
all: `continue OUTER_LOOP` on a nil `LoadBlockCommit` jumps over the
`time.Sleep` that ends the loop body. The same peer height that reached the
panic reaches this branch, so the two fire together. Measured on one peer
claiming height 50 against a chain starting at 100, with the test config's 5
millisecond gossip sleep: 1587544 warn records in one second. That figure is
bounded by the processor rather than by any sleep, so the default 100
milliseconds does not reduce it.

## Decision

A nil `blockMeta` logs at error with the two heights and sleeps for
`PeerGossipSleepDuration` before the next pass, which is the shape
`gossipDataForCatchup` uses on the same nil, with the same two attributes.

The message says `for peer block parts` rather than the sibling's bare `Failed to
load block meta`. Both run on the same logger with the same attributes, so one
string would leave an operator unable to tell which branch fired. The suffix
follows the file's own convention, `for catchup` and `for queryMaj23`.

`gossipVotesRoutine` sleeps for `PeerGossipSleepDuration` before its own
`continue`, which is what every other branch of that loop does. Under the same
measurement it reports 197 records in a second rather than 1587544.

The peer is not disconnected, and the height it claims is not validated where the
message arrives.

## Alternatives considered

**A. Stop the peer for error.** Rejected. `Receive` calls
`Switch.StopPeerForError` where the message itself is broken, on a decode failure
and on a `ValidateBasic` failure. A nil `blockMeta` is a fact about this node's
store rather than about the message, and every other blockstore read driven by
`prs.Height` waits the nil out, this branch's second change included. One that
disconnects instead would fire on an honest node started against the wrong
genesis as readily as on an attacker.

**B. Reject a height below `InitialHeight` in `NewRoundStepMessage.ValidateBasic`
or in `ApplyNewRoundStepMessage`.** Rejected here, not on the merits.
`ValidateBasic` runs on the message alone and has no state to compare against,
so the check would have to move into `ApplyNewRoundStepMessage`, and every
consumer of `prs.Height` would then be relying on a bound declared somewhere
else. The read that panics is the one this change is answering, and it is
guarded at the read.

**C. Recover inside `gossipDataRoutine`.** Rejected. It turns one known nil into
a catch-all that also swallows the next unrelated panic, and the routine would
have to decide whether to restart itself.

## Consequences

- A node on a chain with `InitialHeight > 1` survives a `NewRoundStepMessage`
  naming any height below it. Before, one such message ended the process.
- While a peer holds that height, the node logs one error line and one warn line
  per `PeerGossipSleepDuration`, 100 milliseconds by default, for that peer. Both
  name the peer and the height it claims, and the error line adds the node's
  height and the store's, so an operator separates a stuck peer from a hostile one
  by the heights alone. `gossipDataForCatchup` keeps the same unthrottled shape on
  its own nil, which this branch leaves alone.
- Nothing about state, the wire format, block validation or a hash moves, so a
  patched and an unpatched node agree on every block. They differ only in whether
  the unpatched one is still running.
- `TestGossipRoutinesHeightBelowInitialHeight` drives both routines against a
  peer claiming height 50 on a chain that starts at 100. Restoring the panic
  crashes the test binary from the gossip goroutine.
