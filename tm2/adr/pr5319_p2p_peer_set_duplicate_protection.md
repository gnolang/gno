# ADR: P2P Peer Set Duplicate Protection

## Context

`set.Add()` in `tm2/pkg/p2p/set.go` unconditionally increments the `inbound`
or `outbound` counter without checking if the peer ID already exists. The map
overwrites duplicates silently, but the counter always increments.

This means N calls to `Add()` with the same peer ID produce 1 map entry but a
counter of N. When all connections close, only one `Remove()` succeeds
(decrementing once), leaving a permanent ghost counter of N-1.

The inbound accept loop (`runAcceptLoop()`) gates solely on `NumInbound()`,
and the transport deduplicates by IP:port — not peer ID. So the same keypair
from different IPs inflates the counter with no guard. With
`maxInboundPeers = 40`, this permanently exhausts all inbound slots.

## Decision

Two complementary fixes:

### 1. `set.Add()` returns an error on duplicate (root cause)

Change the `PeerSet.Add()` signature to return an `error`. If the peer ID
already exists, return an error without modifying the map or counters:

```go
func (s *set) Add(peer PeerConn) error {
    s.mux.Lock()
    defer s.mux.Unlock()
    if _, exists := s.peers[peer.ID()]; exists {
        return errors.New("duplicate peer")
    }
    s.peers[peer.ID()] = peer
    if peer.IsOutbound() {
        s.outbound++
    } else {
        s.inbound++
    }
    return nil
}
```

The `addPeer()` function in `switch.go` checks this error. Both callers
(`runAcceptLoop`, `runDialLoop`) already handle `addPeer()` errors with
proper cleanup (transport removal, peer stop). This also closes the TOCTOU
gap in `runDialLoop` where `Has()` is checked before dialing but the peer
could connect between the check and the `Add()`.

### 2. Reject duplicate peer IDs in `runAcceptLoop()` (defense in depth)

Add `peers.Has(p.ID())` before `addPeer()`. This is a fast-path optimization
that avoids wasted work (peer start, reactor init) when the peer is obviously
a duplicate.

## Alternatives considered

**A. Handle direction changes in `Add()`** — rejected because direction change
inside `Add()` is an implicit side effect. If a caller needs to handle
direction change, it should explicitly `Remove()` then `Add()`, making the
design choice visible at the call site.

**B. Only apply the switch-level guard** — rejected because `set.Add()` would
still violate its own invariants. Any future caller without a prior `Has()`
check would reintroduce the bug.

**C. Make `Add()` silently idempotent (no error)** — rejected because callers
should know a duplicate occurred so they can clean up the already-started peer.

## Consequences

- Ghost counter inflation blocked at two layers
- `set` counters always match map contents
- `PeerSet` interface changed: `Add()` now returns `error`
- TOCTOU gap in `runDialLoop` closed by error return from `Add()`
- 4 new regression tests, no regressions on existing tests
