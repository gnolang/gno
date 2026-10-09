# PR6303: Reset a persistent peer's backoff only after a stable connection

## Status

Proposed

## Context

#6292 made the redial loop the only producer of persistent peer dials (see `pr6292_persistent_peer_dialing.md`). Its decision 3 cleared a peer's redial backoff on every `PeerConnected` event, and it rejected an immediate redial on disconnect, because with that reset a peer that connects and drops at once would be redialed in a tight loop. Two consequences followed (#6301):

- a persistent peer that completes the handshake and then drops us at once (inbound slots full, duplicate-IP guard) was redialed on every 5-second tick, forever;
- a dropped persistent peer waited for the next tick, up to 5 seconds; a validator behind sentries, with peer exchange off, used to redial at once on master.

A related gap: a tick can queue a backoff dial while that peer's previous dial is still in flight. If that dial connects, the queued item survives the connection, and after a later drop the redial loop waits for it, up to about 33 seconds.

## Decision

1. The redial loop subscribes to `PeerConnected` and `PeerDisconnected` for persistent peers, on one channel so they are handled in the order they are notified. It keeps `attempts` and a new `connectedAt` map, both owned by its goroutine.
2. On connect, it records the time and removes any dial queued for the peer's configured address (`dial.Queue.Remove`). The backoff is left as is. A connect handled after the peer already dropped is ignored, and also drops any connect time left from a connection whose disconnect was missed.
3. On disconnect, it resets the backoff if the connection lasted the switch's `stableUptime`, which defaults to `persistentStableUptime`, the 30-second backoff ceiling, and which tests shorten; then it queues that peer at once: due right away after a stable connection, after the next backoff step otherwise. A disconnect with no recorded connect counts as unstable.
4. The 5-second tick is unchanged; it covers startup, dropped events and failed dials.
5. The dial loop pops through `dial.Queue.PopDue`, which checks and pops the head under the queue's lock, since the redial loop's removal can now run between the dial loop's peek and pop.
6. The switch notifies a disconnect only when the peer actually left the peer set, so each connect has one matching disconnect.

## Alternatives considered

- **A 5-second stability threshold** (one redial interval): a broken connection is recognised sooner, but the worst pattern is a stable connection followed by a few short ones, since the stable drop resets the backoff and the short ones walk it up again (dials at 0, 1, 3, 7 and 15 seconds, then a connection of the threshold). With the threshold at the 30-second ceiling that is four short connections per stable one, about one dial every 9 seconds at worst (about 400 an hour), below #6292's one per 5-second tick (720 an hour); with a 5-second threshold the worst such pattern, two short connections per stable one, reaches about one dial every 3 seconds, above it.
- **Rescheduling the leftover dial on disconnect:** needs a reschedule operation and a re-sort, and stale items stay queued while the peer is connected.
- **Tracking dials in flight to avoid leftovers:** shared state between the dial loop and the redial loop.
- **Resetting the backoff on the tick** for peers connected long enough: splits the logic between the tick and the events, and needs the connect times anyway.
- **Keeping the connect time on the peer and deciding in `StopPeerForError`:** puts redial decisions back on the error path, against #6292's single-producer rule.

## Consequences

- A persistent peer whose connections last under 30 seconds is redialed with a backoff that steps up one per drop from where the attempt count stood, so after a reset 1, 2, 4, 8, 16, then 30 seconds, and a peer that needed several dials before a short connection resumes at that step, instead of every 5 seconds.
- A link that stays up between 5 and 30 seconds per cycle is, once backed off, reconnected after up to about 33 seconds instead of within the 5-second tick; a link up 20 seconds per cycle is connected about 40% of the time instead of about 80%. That is the price of bounding the dial rate.
- A persistent peer that drops after 30 seconds or more of uptime has its dial queued at once, without waiting for the tick. The dial loop serves it behind the persistent dials already due and, while discovered dials are due, after at most one discovered dial per persistent dial ahead of it (#6292's alternation). With no other dial due, a validator reconnects to its sentry as fast as before #6292; behind dead persistent peers and peer exchange, the queueing dominates (see the measurements below).
- A queued dial no longer outlives the connection it was meant to establish, unless that connection's `PeerConnected` event is dropped (see below).
- A dropped `PeerConnected` event makes the next drop count as unstable, and lets a leftover dial survive, bounded by the 30-second ceiling; a dropped `PeerDisconnected` event leaves the redial to the tick, which queues the peer with the backoff it had before that connection, so up to about 38 seconds after the drop (the next tick, then the backoff) instead of at once.
- Events and peer-set changes are not strictly ordered, but a connect handled after its connection already dropped is ignored, so it neither stamps a start time nor removes the dial the disconnect queued. The remaining window is a connection shorter than the event-handling latency, where a leftover dial can survive, bounded by the 30-second ceiling.
- Both ends of a stable persistent link see the drop at about the same time and queue a redial at once, so two nodes that list each other as persistent peers dial each other at the same moment more often than under the tick phases described in PR6305's context; PR6305 resolves such a simultaneous open deterministically.

The redial delay was measured with #6292's simulation harness (`testing/synctest`): a live persistent peer drops every 97 seconds, after a stable connection, while dead persistent peers time out after 3 seconds and, with peer exchange, 30 unreachable addresses are queued every 10 seconds. Delay from the drop to the redial, average and worst over 20 drops, the same in three runs within a second:

| Dead persistent peers | Peer exchange | #6292 | This PR |
|---|---|---|---|
| 12 | on | 50s / 73s | 47s / 74s |
| 4 | on | 8s / 17s | 6s / 14s |
| 12 | off | 10s / 35s | 8s / 32s |
| 0 | off | — | under 1s |

With 12 dead peers and peer exchange, the queueing behind the dead peers' due dials and the alternation dominate; strict priority would serve the redial in 8s and 32s, but lets dead persistent peers starve the seeds, as #6292 records. The seed-starvation scenario from #6292's review is unchanged: 24 seed connections in 12 minutes with 12 or 16 dead persistent peers.
