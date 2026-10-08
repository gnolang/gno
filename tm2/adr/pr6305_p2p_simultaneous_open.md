# PR6305: Resolve a simultaneous open by keeping the lower ID's connection

## Status

Proposed

## Context

The switch refused a second connection for a peer ID already registered by arrival order, in the accept loop's duplicate check, in `addPeer`'s fast-path check, and in `set.Add`. When two nodes X and Y dial each other at the same moment there are two connections, c1 dialed by X and c2 dialed by Y. If X registers c1 first and Y registers c2 first, each side rejects the other's inbound connection and closes it, which is the connection the other side kept, and both connections die. #5781 measured this at about one run in three for the test clusters and worked around it there, by having each pair dial in one direction only; the switch was unchanged. CometBFT v0.38.21 has the same arrival-order rule.

It matters for persistent peers since #6292: two nodes that list each other redial on each node's fixed tick phase, so two nodes whose phases are close can collide on every tick until timing breaks the tie (gnolang/gno#6302).

## Decision

### 1. The connection dialed by the lower ID survives

`keepsRegistered(registered, incoming)` resolves a duplicate. For two connections in opposite directions, the preferred direction is outbound when the local ID is lower than the peer's and inbound otherwise, and the registered connection is kept only if it has the preferred direction. Both ends compare the same two IDs, so they keep the same connection without exchanging anything. Two connections in the same direction keep the registered one, as before, and so does a switch that does not know its own ID yet.

### 2. The decision is made where a connection registers, atomically

The accept and dial loops can register the same peer concurrently, so a switch-level registry lock covers every step that reads the entry of a peer ID and then changes it. `registerPeer` adds a connection, refuses it with `errDuplicatePeer`, or replaces the registered one and returns it. `addPeer` tears the replaced connection down before the new one's reactors run `AddPeer`. The replaced connection gets its `RemovePeer` exactly once: it runs before the new connection's `AddPeer`, unless the replaced connection's own teardown, after the remote closed it, began first, in which case it may run concurrently with that `AddPeer`. A connection refused at registration gets `addPeer`'s unwind `RemovePeer` instead, which can run after the kept connection's `AddPeer`. The early duplicate checks in `addPeer` and the accept loop apply the same rule, so a winner is not turned away before it reaches registration.

### 3. Removal checks and removes in one step

`stopAndRemovePeer` used to check whether its connection was superseded and then remove the entry by ID, as two steps. A replaced connection that errors between them, for example because the other side just closed it while resolving the same simultaneous open, would delete the entry of the connection that replaced it. `removeUnlessSuperseded` does both under the registry lock, and `addPeer`'s rollback of a peer stopped while being added uses it too. A replaced connection announces no disconnect, since its peer stays connected. `stopAndRemovePeer` stops a connection before closing its socket, which makes its teardown run once: it also removes the second `RemovePeer` and the second `PeerDisconnected` event that any switch-initiated stop produced when the recv routine reported the socket close.

### 4. A replacement goes through the guards

A replacement moves a connection from one direction to the other, so it takes a slot in its own direction and goes through the inbound or outbound peer limit like any other connection. Two adjustments keep a genuine simultaneous open working: the duplicate-IP guard does not count a connection of the arriving peer's own ID, since registration resolves two connections of one ID itself, and a persistent peer's replacing inbound connection is exempt from the inbound peer limit, so two persistent peers dialing each other at once still converge on a node whose inbound slots are full. The exemption holds while the connection being replaced is still registered when the accept loop checks; if it is already gone, the new connection goes through the inbound limit like a new peer, and the redial loop recovers. Persistent peers are configured by the operator, so no remote party can use that exemption. The outbound limit applies to a replacing outbound connection; persistent peers were already exempt from it.

## Alternatives considered

- **Newest connection wins.** It fails for opposite directions: each side's newest connection is the one the other side drops, which is the same teardown.
- **Jitter on persistent redials.** It makes a collision less likely without resolving it, does not cover discovered peers dialing each other, and delays every first redial.
- **Negotiating the surviving connection over the wire**, as libp2p does. It needs a protocol change and complicates networks mixing versions, for a decision both ends can make locally.
- **Exempting every replacement from the peer limits and the duplicate-IP guard.** An accept-loop replacement always swaps one of our outbound connections for an inbound one, so any peer we dialed with a lower ID could convert its connection at will, past both guards, and peer exchange would refill the freed outbound slot: one host with many identities could accumulate connections from a single IP without bound. The dial path has the mirror problem with the outbound limit.

## Consequences

- When both ends run this rule, a simultaneous open leaves the connection dialed by the lower ID on both sides, except in the cases below.
- Between a node running this rule and one keeping the first-registered connection, a simultaneous open survives only when the latter happened to register the same connection first: about as likely as before, in different orderings. Convergence is guaranteed once both ends run the rule, outside the cases below.
- A simultaneous open can still end with both connections torn down when the lower-ID peer is not a persistent peer of the higher-ID node and that node is at its inbound peer limit, when the winning connection is refused by the duplicate-IP guard because another peer holds its IP (persistent or not), or, on the dial side, when the node is at its outbound peer limit and the peer is not persistent. The redial loop recovers persistent peers.
- A peer re-dialing in the same direction while its previous connection lingers is still refused until that connection times out.
- #5781's one-direction dialing in `tm2/pkg/internal/p2p` is no longer needed for correctness and is left in place.
- A replacement logs at Info, `replacing connection to resolve a simultaneous open`.
