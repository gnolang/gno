# PRxxxx: Reset a persistent peer's backoff only after a stable connection

## Status

Proposed

## Context

#6292 made the redial loop the only producer of persistent peer dials (see `pr6292_persistent_peer_dialing.md`). Its decision 3 cleared a peer's redial backoff on every `PeerConnected` event, and it rejected an immediate redial on disconnect, because with that reset a peer that connects and drops at once would be redialed in a tight loop. Two consequences followed (#6301):

- a persistent peer that completes the handshake and then drops us at once (inbound slots full, duplicate-IP guard) was redialed on every 5-second tick, forever;
- a dropped persistent peer waited for the next tick, up to 5 seconds; a validator behind sentries, with peer exchange off, used to redial at once on master.

A related gap: a tick can queue a backoff dial while that peer's previous dial is still in flight. If that dial connects, the queued item survives the connection, and after a later drop the redial loop waits for it, up to about 33 seconds.

## Decision

1. The redial loop subscribes to `PeerConnected` and `PeerDisconnected` for persistent peers, on one channel so they are handled in the order they are notified. It keeps `attempts` and a new `connectedAt` map, both owned by its goroutine.
2. On connect, it records the time and removes any dial queued for the peer's configured address (`dial.Queue.Remove`). The backoff is left as is. A connect handled after the peer already dropped is ignored.
3. On disconnect, it resets the backoff if the connection lasted `persistentStableUptime`, set to the 30-second backoff ceiling, then queues that peer at once: due right away after a stable connection, after the next backoff step otherwise. A disconnect with no recorded connect counts as unstable.
4. The 5-second tick is unchanged; it covers startup, dropped events and failed dials.
5. The dial loop pops through `dial.Queue.PopDue`, which checks and pops the head under the queue's lock, since the redial loop's removal can now run between the dial loop's peek and pop.

## Alternatives considered

- **A 5-second stability threshold** (one redial interval): a broken connection is recognised sooner, but a peer up 6 seconds at a time would be redialed at once every 6 seconds. Tying the threshold to the ceiling bounds the worst case at about one dial per 30 seconds, whatever the uptime pattern.
- **Rescheduling the leftover dial on disconnect:** needs a reschedule operation and a re-sort, and stale items stay queued while the peer is connected.
- **Tracking dials in flight to avoid leftovers:** shared state between the dial loop and the redial loop.
- **Resetting the backoff on the tick** for peers connected long enough: splits the logic between the tick and the events, and needs the connect times anyway.
- **Keeping the connect time on the peer and deciding in `StopPeerForError`:** puts redial decisions back on the error path, against #6292's single-producer rule.

## Consequences

- A persistent peer whose connections last under 30 seconds is redialed after 1, 2, 4, 8, 16, then 30 seconds, instead of every 5 seconds.
- A persistent peer that drops after 30 seconds or more of uptime is redialed at once, without waiting for the tick, so a validator reconnects to its sentry as fast as before #6292.
- A queued dial no longer outlives the connection it was meant to establish, unless that connection's `PeerConnected` event is dropped (see below).
- A dropped `PeerConnected` event makes the next drop count as unstable, and lets a leftover dial survive, bounded by the 30-second ceiling; a dropped `PeerDisconnected` event falls back to the tick, within 5 seconds.
- Events and peer-set changes are not strictly ordered, but a connect handled after its connection already dropped is ignored, so it neither stamps a start time nor removes the dial the disconnect queued. The one remaining window is a connection shorter than the event-handling latency, where a leftover dial can survive, bounded by the 30-second ceiling.
