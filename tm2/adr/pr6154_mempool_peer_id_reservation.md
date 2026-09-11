# ADR: Pairing Mempool Peer ID Reservation With Its Release

## Context

`mempoolIDs` in `tm2/pkg/bft/mempool/reactor.go` hands every peer a `uint16`,
so a transaction is never gossiped back to the peer it arrived from. The pool
holds 65535 of them, and `nextMempoolPeerID` panics once they are all taken.

`AddPeer` took the ID and `RemovePeer` gives it back. The switch calls those
two from different places, and `addPeer` in `tm2/pkg/p2p/switch.go` starts the
peer's send and receive routines before it walks the reactors' `AddPeer`. A
connection error from that point on reaches `StopPeerForError`, which walks the
reactors' `RemovePeer`. So `Reclaim` runs first, finds nothing under that peer
in `peerMap`, and returns. `ReserveForPeer` then takes an ID that nothing will
ever give back.

One raced connection leaks one ID. Peer churn is what drives the race, and the
count only grows: nothing sweeps `activeIDs`. At 65536 the panic fires in
`runAcceptLoop` or `runDialLoop`, neither of which has a `recover`, so the
process ends and the node stops producing blocks.

Moving the reservation to `InitPeer` is not enough on its own, because the
reservation is only half of the pairing. Three properties of the switch are
what actually decide whether reactor state is given back:

- `addPeer` can return an error after `InitPeer` has run, from `p.Start()` and
  from `sw.peers.Add`. Neither `runAcceptLoop` nor `dialPeer` walks the
  reactors' `RemovePeer` on that return, so whatever `InitPeer` took is held
  for the lifetime of the process.
- Reactor per-peer state and the peer set are both keyed on the peer ID, and
  two connections hold the same peer ID while a reconnect races the teardown of
  the connection it supersedes. Reclaiming by ID alone strips whichever
  connection won the race.
- `stopAndRemovePeer` removes from the peer set last, on purpose
  (tendermint#3338). So its `Remove` can run before `addPeer` reaches its
  `Add`, and `addPeer` then inserts a peer that has already been stopped and
  that nothing will ever remove again.

## Decision

### 1. `InitPeer` reserves the ID

The switch calls `InitPeer` before `p.Start()`, so the reservation is in place
before the connection can fail. `AddPeer` keeps the broadcast routine and
nothing else. This is the hook CometBFT settled on in `mempool/v0/reactor.go`.

The consensus reactor uses `InitPeer` too, but it is not precedent for keeping
the state here: it stores per-peer state on the peer with
`peer.Set(types.PeerStateKey, ...)`, so the state is collected with the peer,
and the warning on `Reactor.InitPeer` — do not keep peer state in the reactor
itself, because it is never cleaned up — does not reach it. `mempoolIDs` is
reactor-held, which is what the rest of these decisions exist to pair.

### 2. Reservations are keyed on the connection

`peerMap` is `map[p2p.PeerConn]uint16` rather than `map[p2pTypes.ID]uint16`.
Two connections carrying one peer ID each get an ID of their own and give back
only the ID they took.

Keyed on the peer ID instead, a superseded connection's `Reclaim` deletes the
entry the connection that replaced it depends on, and that connection spends
the rest of its life reading under `UnknownPeerID` — the ID reserved for
locally submitted transactions, so `senders` reports every one of them as
already seen and the node never gossips its own transactions to that peer.

The switch hands every reactor hook the same `PeerConn` for a connection:
`peer.createMConnection` closes over it for `Receive` and `OnPeerError`, and no
reactor's `InitPeer` substitutes a different one. A connection is therefore its
own identity here.

### 3. `addPeer` unwinds reactor state on its error returns

`removeReactorPeerState` walks the reactors' `RemovePeer` for the two returns
that follow `InitPeer`, so a connection the switch refuses does not keep what
`InitPeer` gave it.

The walk is unconditional, including for a connection another has superseded.
Skipping it there would be the leak again: keyed on the connection, a reactor
gives back only what that connection took, and giving nothing back keeps it
forever. Only the peer set entry needs the distinction, in decision 6.

### 4. `addPeer` refuses a duplicate before any reactor sees it

`sw.peers.Has` runs before the `InitPeer` loop. A connection `sw.peers.Add`
would refuse anyway now neither starts nor reaches a reactor, which is what
`pr5319_p2p_peer_set_duplicate_protection.md` claimed the peer set guard was
worth. `Add` stays the authoritative check for the residual race.

### 5. `addPeer` refuses a peer stopped while it was being added

After `sw.peers.Add`, `addPeer` checks `p.IsRunning()` and rolls the insertion
back if the peer is already stopped. Otherwise that peer holds an inbound slot
and its peer ID for the lifetime of the process: `sw.peers.Remove` is only ever
reached from `stopAndRemovePeer`, which has already run for it. 40 of those —
the `MaxNumInboundPeers` default — and the node accepts no further inbound
connections, which is 1600 times cheaper than the 65535 the ID pool costs.

### 6. `stopAndRemovePeer` leaves a superseded connection's peer set entry

A connection that lost the race for the peer set still closes its own socket
and gives back its own reactor state, but the entry under its peer ID is the
winning connection's, so it neither removes it nor announces a disconnect it
never announced a connection for.

### 7. The outbound peer limit is enforced where peers are added

`MaxNumOutboundPeers` was checked only where an address is queued for dialing,
in `DialPeers` and `dialItems`, and `NumOutbound` cannot change while either
loop runs. So one batch of queued dials overshot the limit without bound, and
`activeIDs` was not bounded by the peer limits at all. `addPeer` now enforces
the limit where the peer is actually added, exempting persistent peers as
`MaxNumOutboundPeers` documents.

A `Response` in `tm2/pkg/p2p/discovery` is the remote way to fill that queue,
and `ValidateBasic` bounded only its lower end. It now rejects a response
carrying more peers than `maxPeersShared`, the number the protocol itself
shares.

## Alternatives considered

**A. Record the early `Reclaim` and skip the late `ReserveForPeer`.** Rejected
because the record is only consumed by a reservation that follows it. A
`Reclaim` for a peer that never reserves leaves an entry nothing deletes, which
trades a bounded pool for an unbounded map.

**B. Guard `AddPeer` on `peer.IsRunning()`.** Rejected as the primary fix: it
narrows the window rather than closing it, since the peer can be stopped
between the check and the reservation. With the reservation in `InitPeer` it
guards nothing, and `broadcastTxRoutine` already returns on a peer that is not
running. The `IsRunning` check in decision 5 is a different one: it guards the
peer set insertion, which `stopAndRemovePeer` will not undo a second time.

**C. Re-reserve in `AddPeer` for a peer whose ID was reclaimed.** Rejected
because it reintroduces the leak this ADR removes. `stopAndRemovePeer` can
complete between `addPeer`'s liveness check and its `AddPeer` loop, and the
reservation that followed it would then have no `RemovePeer` left to pair with.
Connection keying fixes the same problem without a second reservation point.

**D. Return an error from `nextMempoolPeerID` instead of panicking.** Left as a
separate decision: it changes what a caller must handle. Decisions 3 to 7 are
what make the premise true — with them, `activeIDs` is bounded by
`MaxNumInboundPeers` plus `MaxNumOutboundPeers` plus the persistent peers, so
the panic is unreachable rather than merely unlikely. It is worth revisiting on
its own merits, since a bounded pool backed by a process-killing panic in a
goroutine with no `recover` is a poor trade whatever the bound.

## Consequences

- `activeIDs` is bounded by the peer limits, because every path that reaches
  `InitPeer` now pairs it with a `RemovePeer`: `stopAndRemovePeer` for a
  connection that joined, `removeReactorPeerState` for one the switch refused,
  and nothing at all for one that never got past the duplicate check.
- Reactors keeping per-peer state in `InitPeer` can rely on `addPeer` unwinding
  it, which is the contract the `Reactor` doc said did not exist. `RemovePeer`
  has to tolerate being walked for a connection another has superseded, and
  being walked twice for one connection. Both were already reachable, since
  `StopPeerForError` has no once-guard. `BlockPool` in `tm2/pkg/bft/blockchain`
  is the one reactor still keying per-peer state on the peer ID; it tolerates
  both, but it cannot tell a superseded connection from the one that replaced
  it, and it is also written from `Receive`, so its lifecycle is a separate
  problem.
- A peer that reconnects before the old connection's `RemovePeer` finishes gets
  an ID of its own, and keeps it when the old connection's `Reclaim` runs.
- A duplicate connection does not disturb the ID of the peer it duplicates, and
  does not evict it from the peer set.
- Keying on the connection means `mempoolIDs` holds a `PeerConn` reference
  until `Reclaim`. That is bounded by the same pairing as the IDs themselves.
- The mempool reactor depends on the switch calling `InitPeer` before
  `p.Start()`, and on it handing the same `PeerConn` to every hook.
  `TestMultiplexSwitch_AddPeerRemovedBeforeAdded` pins the order and the
  interleaving it protects against, and it now also pins the peer set state, so
  it fails if the insertion is not rolled back.
