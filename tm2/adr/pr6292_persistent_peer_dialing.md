# PR6292: Dial persistent peers only from the redial loop

## Status

Proposed

## Context

On 2026-10-06, up to 7 of 8 mainnet RPC nodes stopped following the chain for about 40 minutes after a rolling restart of their sentries and seeds (#6287). The nodes run peer exchange behind a NAT, take no inbound peers and allow 25 outbound; their persistent peers are configured by private hostname and advertise a public `external_address`, with `allow_duplicate_ip = false`.

The trace of one sentry on one stalled node shows its private link dying at 08:40:13.6, which queued a redial of the private address, due right away. That address was never dialed in the following hour, while the node dialed the sentry's public address 8 times, the first of them closed by the sentry within a millisecond, consistent with its duplicate-IP guard refusing a second connection from the shared NAT address. The outbound peer limit dropped nothing (the warning it logs appears 0 times), and the dial loop was regularly idle after 08:45, so the private item was not held by a backlog either. What remains, by elimination, is the dial loop's check on a popped item: the item was most likely dropped there, by a check that logged nothing at the time, while the sentry briefly counted as connected over its public address. This is inferred, not logged. The spacing of the later public redials (1 minute, then 13.5 minutes, while the loop was mostly idle) is not explained by the trace. Each short-lived public connection reset the backoff, so the redial loop's later re-queues of the configured address were also due at once, and the trace does not explain why those were not dialed either. The pop-time drop explains the first one, and the fix does not depend on which mechanism dropped them.

This PR also fixes #6291: a persistent peer or seed whose hostname did not resolve at startup was dropped until the next restart.

Independent of the exact mechanism, the code had these defects:

- `StopPeerForError` redialed `peer.SocketAddr()`, the address of the connection that died. A peer is persistent by ID, so once it was reached through an address learned via peer exchange, every redial targeted that address; for an inbound persistent peer, the socket address carries the remote's ephemeral port.
- Persistent and discovered dials shared one time-sorted queue, drained by one serial loop and fed by every peer exchange response (up to 30 addresses, never deduplicated). A persistent dial had no priority, and could be dropped at pop time.
- `addPeer` exempted persistent peers from `MaxNumOutboundPeers` (#6164), and the config comment says "excluding persistent peers", but `DialPeers` and `dialItems` applied the limit to them when queueing.
- `calculateBackoff` shifted the base interval by the attempt count unbounded: from 34 attempts the interval wrapped negative, and the next redial was due immediately. With the 10-minute ceiling, that was about four hours of continuous outage, after which the peer was redialed on every tick.

## Decision

### 1. The redial loop is the only producer of persistent peer dials

Every dial of a persistent peer comes from the redial loop, on the configured address, with backoff. `StopPeerForError` queues nothing; `DialPeers` drops any address whose ID is a persistent peer (peer exchange, the address book, seeds); `Node.OnStart` no longer dials persistent peers, since the redial loop's first pass already does; `dialSeed` does not pick a seed that is also persistent.

Peer lists from the node configuration (`p2p.persistent_peers` and `p2p.seeds`) are parsed with `types.NewConfiguredNetAddress`, which keeps an entry whose hostname does not resolve yet: the address has its ID, `Hostname` and port, and no IP, and `DialContext` dials its hostname (the IP is never filled in). The startup lookup is bounded to 2 seconds, so a slow DNS server cannot stall startup per entry, and malformed entries are still rejected. This replaces master's `OnStart`, which re-parsed `persistent_peers` after the genesis sleep, a second DNS lookup that dialed late-resolving entries once as ordinary peers (outside the persistent set, so a failure or a later drop was never retried); this PR drops that lookup and keeps such entries in the persistent set instead, retried by the redial loop (#6291). An address without an IP is safe because the self-address check (`Same`) matches on the ID or on the dial string, the dial string of an unresolved address (`DialString()`, which `Same` compares) is `<nil>:port`, and the node's own address always carries an IP (`0.0.0.0` or real), so an unresolved configured address is never taken for the node itself. `NewNetAddressFromString` is unchanged: peer-supplied addresses, the address book and the node's own addresses still require a resolvable host.

### 2. Persistent peers get their own dial queue

The redial loop pushes into `persistentDialQueue`, with no outbound limit check, matching `addPeer`. The dial loop takes the persistent head when it is due, except that when both heads are due and the last item it popped was a persistent one, it takes the general head (`peekDialItem`, fed by a flag local to the dial loop); with nothing due, it waits for whichever is due first. The two queues take turns while both are due, so neither starves the other. `hasDialableItem`, the seed service's gate, looks at the dial queue only: seeds exist to refill peer discovery, which a pending persistent dial does not do, and unreachable persistent peers must not shut off the seed fallback.

### 3. Backoff ceiling of 30 seconds, reset on connect

The persistent redial backoff doubles from one second up to 30 seconds (it was 10 minutes), keeping the existing ±10% jitter. With the 5-second tick, attempts are at most about 38 seconds apart once saturated, which bounds the reconnect delay after a peer returns. A `PeerConnected` event clears the peer's backoff. A peer that connects and drops at once is therefore redialed at most once per tick: the tick is its rate limit, and a test pins it.

### 4. The general queue is deduplicated

`DialPeers` skips a peer that is already connected and an exact address already queued. It does not deduplicate by ID alone, so a stale address cannot shadow a working one.

### 5. The backoff never overflows

`calculateBackoff` only takes the shift when its result fits under the ceiling, and returns the ceiling otherwise.

## Alternatives considered

- **Rewrite persistent addresses to the configured one in `DialPeers`** (the proposal in #6287). It fixes the address, but leaves the error path queueing immediately and outside the backoff. Combined with a priority queue, a peer that connects and drops at once would be redialed as fast as the network allows and starve every other dial. Dropping those addresses keeps the backoff in charge of every persistent dial.
- **Keep one queue, deduplicate and cap it.** A persistent dial would still wait behind discovered peers, and still be droppable at pop time behind a flapping connection.
- **A goroutine per persistent peer, as CometBFT does.** It isolates persistent dials too, and it isolates persistent and discovered dials from each other in both directions, at the cost of a second dial path with its own concurrency handling. The second queue keeps a single dial path and accepts the trade-off listed under Consequences.
- **Immediate redial on disconnect, through a signal to the redial loop.** It saves up to one tick (5 seconds) per reconnect, but with the reset on connect, a flapping peer would be redialed in a tight loop again.

## Consequences

- A dropped persistent peer is first redialed on the next tick, 0 to 5 seconds later, instead of an immediate attempt that could wait behind the shared queue (57 seconds in the incident trace). A backoff dial still queued from before the last connection is used instead when present, so the first redial can then wait up to about 33 seconds. A validator behind sentries runs with peer exchange off, so on master its error-path redial ran on an idle queue and was effectively immediate; it now waits up to one tick (5 seconds), or for a leftover backoff dial, as above.
- Persistent and discovered dials alternate in the single dial loop while both are due. The persistent head waits behind at most one discovered dial (up to about 9 seconds with its dial in flight); persistent peers that time out take every other slot, but no longer starve discovered dials or the seeds.
- Alternation fixes starvation, not the cost of serial dials: a persistent peer that just dropped queues, by due time, behind the dials already due for dead persistent peers (up to 3 seconds each), and while discovered dials are due, one discovered dial runs after each of those (up to 9 seconds per discovered slot if remotes hang the handshake), so its wait roughly doubles. In a simulation with 12 dead persistent peers that time out and peer exchange feeding discovered dials, a dropped live persistent peer was redialed after 50s on average and 73s at worst, against 11s and 35s with strict priority; with 4 dead peers, 8s and 17s against 5s and 8s; with peer exchange off, unchanged. The trade is accepted because strict priority let dead persistent peers shut off the seeds and discovered dials indefinitely (zero outbound peers), and the remaining cost is that of serial dials, which only concurrent persistent dials would remove (the per-peer goroutine alternative).
- Peer exchange can no longer stand in for a persistent peer whose configured IP went stale. Since #5020, a persistent peer configured by hostname is re-resolved on every dial, which covers the hostname case; an IP-configured peer needs a configuration update.
- A node started before genesis whose persistent peer or seed does not resolve yet keeps it, and dials it once it resolves.
- `max_num_outbound_peers` no longer blocks persistent dials, even at 0. Persistent peers still count toward the number of outbound peers, so the slots left for discovered peers are unchanged.
- After a restart of a persistent peer, the dialer reconnects within about 38 seconds of it coming back when no other persistent peer is down (otherwise add the queueing described above), instead of up to 10 minutes.
- Repeated peer exchange lists no longer queue the same addresses again.
- A dial dropped because its peer is already connected is logged at debug level.
- Each persistent peer has a single configured address. When `p2p.persistent_peers` lists the same peer ID more than once, only the last address is dialed, and the node logs a warning at start for each ignored entry whose address differs from the dialed one.

Supersedes #6053, which removed the README's claim that persistent peers ignore the peer limit; this change makes that claim true for the outbound limit.
