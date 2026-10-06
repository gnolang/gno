# PR6023: Wire config.P2P.Seeds into the P2P switch

## Status

Proposed

## Context

`P2PConfig.Seeds` is declared in `tm2/pkg/p2p/config/config.go` and was never read
anywhere in the codebase:

- no `WithSeeds` option in `tm2/pkg/p2p/switch_option.go`
- `node.go` parsed `PersistentPeers` only, in both `NewNode()` and `OnStart()`
- `ResultDialSeeds` existed in `responses.go` with no handler

Operators setting `seeds` in `config.toml` got silent no-op behaviour. The practical
consequence is that TM2 had no way to bootstrap from a non-persistent peer: any bootstrap
address had to be declared in `persistent_peers`, which is then redialed with exponential
backoff for the lifetime of the node.

This is distinct from peer discovery, which already worked. `discovery.Store` persists known
peers to `config/addrbook.json` and `discovery.Reactor` exchanges them on channel `0x50` when
`pex = true`. What was missing is the entry point into that mechanism.

Fixes #5340.

## Decision

Seeds become a first-class peer category on `MultiplexSwitch`, distinct from persistent
peers, served by a dedicated seed dial service.

### 1. Seeds are stored separately from persistent peers

`WithSeeds([]*types.NetAddress)` populates a `seeds` field on `MultiplexSwitch`. Because
seeds are not in `persistentPeers`, they are not picked up by `runRedialLoop` and are
therefore not kept alive with exponential backoff. That is the intended difference: a seed is
an entry point, not a peer the node wants to stay connected to.

### 2. The redial gate is dial queue emptiness, not peer count

Seeds are redialed when the dial queue holds nothing that can be dialed right now, which
covers both an empty queue and a queue where every item is backing off. The queue is
time-sorted ascending, so a head item scheduled in the future means every queued item is
currently in backoff:

```go
func (sw *MultiplexSwitch) hasDialableItem() bool {
	item := sw.dialQueue.Peek()

	return item != nil && !time.Now().Before(item.Time)
}
```

### 3. A dedicated loop throttles the seed dial rounds

The gate describes a state rather than an event: it stays true for as long as the node has
nothing to dial. The cadence therefore has to come from somewhere, so a dedicated loop ticks
every 30 seconds, and that interval doubles as the minimum delay between two rounds.

`runDialLoop` is not that somewhere. On an empty queue it parks in `waitForPeersToDial` until an
address is pushed, which is precisely the state a seed is meant to break, so a seed dial hosted
there would never fire when it is needed. This holds independently of the busy-spin on backed off
items reported in #6052.

The loop `runSeedDialLoop` runs its first round before the ticker, the same way `runRedialLoop`
calls `redialFn()`, and `node.OnStart` does not dial seeds itself. Bootstrap and fallback therefore
share a single path and the same slot accounting.

### 4. Seeds count against `MaxNumOutboundPeers`, and the node never disconnects

Seeds go through the regular `DialPeers` path and count against `MaxNumOutboundPeers` like
any other dialed peer. `dialSeed` checks the limit itself, before anything else, rather than
relying on `DialPeers` to reject the dial, which would log a warning on every tick on a node
that already has all the outbound peers it needs. The node never closes a seed connection: the
reference point raised in review was CometBFT, where the seed is the side that disconnects
once it has served its purpose.

### 5. A single seed is dialed per round

Queuing every seed at once would fill the outbound slots with bootstrap connections on an
empty node, which is exactly what counting them against the cap is meant to avoid. `dialSeed`
first gathers the seeds that are neither already connected nor already in the dial queue,
then dials exactly one of them, picked at random. If it turns out to be unreachable, the next
round picks another candidate. The random index falls back to the first candidate if the
random source is unavailable, rather than failing the round.

### 6. Seeds require peer exchange

`WithSeeds` is applied only when `pex` is enabled. Without the discovery reactor there is no
way to ask a seed for peers, so the connection, and the redial loop behind it, would serve no
purpose. A warning is logged when seeds are configured with peer exchange off. The parsing
and this gate live in `parseSeedAddrs` in `node.go`, so the node carries no seed state of its
own.

### 7. Discovery is requested on every new outbound connection

`discovery.Reactor` only asked for peers on its own tick, picking one peer at random out of
the whole peer set every 3 seconds. A freshly dialed peer had to wait to be selected, which
takes interval x N on average: with the default caps of 10 outbound and 40 inbound peers,
around 2.5 minutes on a full node. At bootstrap this is invisible, but a seed dialed later
sits idle for minutes while holding an outbound slot here, and an inbound slot on the seed
itself, which is a scarce resource for a node meant to serve thousands of peers.

`Reactor.AddPeer` is implemented so a discovery request goes out as soon as an outbound
connection is established. This applies to every outbound peer, not only to seeds: the
reactor has no notion of a seed, and any freshly dialed peer benefits from the same
treatment. Inbound peers are left alone, since we did not choose them and they are their own
source of addresses.

## Alternatives considered

- **Redial the seeds when outbound peers fall below half of `MaxNumOutboundPeers`**: rejected
  because peer count is the wrong signal. A node with 2 peers and 800 addresses in the
  address book would redial the seeds when it should pick an address from the book instead.
- **Reuse the address starvation gate raised in review**, described there as CometBFT's
  (fewer than 1000 *not bad* addresses in the book): rejected because TM2 has no good/bad
  qualification for address book entries, so implementing the same gate goes beyond the scope
  of this change.
- **Exempt seeds from `MaxNumOutboundPeers`**: rejected as useless. If the node has filled its
  peer list up to the cap, there is no reason to dial a seed, since it already has all the peers
  it needs. The exemption the original issue assumed for seeds did not exist in the code either:
  `DialPeers` already enforced the limit for every address.
- **Keep the bootstrap dial in `node.OnStart`**: rejected. `DialPeers` checks the outbound cap
  at enqueue time only, and the counter is still zero at `OnStart`, so every configured seed
  was queued regardless of the cap. Reproduced with 6 seeds and
  `max_num_outbound_peers = 2`: 6 outbound connections were established.
- **Thread a seed flag down into `discovery.Reactor`** so only seeds are asked for peers on
  connection: rejected, because it introduces into the reactor a distinction it does not
  otherwise make, for no benefit over asking every outbound peer.
- **Handle the idle-seed problem in a follow-up PR**: possible, but the feature is incomplete
  without it, and it is needed before a dedicated seed node implementation lands.

## Consequences

- `p2p.seeds` is honored: a node can bootstrap from an address it does not keep redialing.
- Seeds are dialed one at a time, on start and whenever the node runs out of peers to dial,
  and count against `p2p.max_num_outbound_peers`.
- Startup ordering is not deterministic. If persistent peers are already queued when
  `runSeedDialLoop` starts, the queue is dialable and the first seed dial happens 30 seconds
  later instead. This is benign, since the node already has somewhere to go.
- An unreachable first seed costs 30 seconds before the next candidate is tried.
- Every outbound peer, not only seeds, is now asked for its peer set on connection. Peer
  discovery converges faster, at the cost of one extra request per outbound connection.
- Seeds configured with `p2p.pex` disabled are ignored, with a warning, instead of being
  silently dialed for nothing.

## Tests

`tm2/pkg/p2p/switch_test.go`:

- `TestMultiplexSwitch_DialSeed`, eight cases: no seeds configured, a dialable item in the
  queue, queued items fully backed off, an empty queue, the outbound peer limit reached, a
  connected seed skipped, a queued seed not duplicated, and a single seed dialed per round.
- `TestMultiplexSwitch_SeedDialLoop`: the bootstrap round runs before the first tick, and the
  loop exits on context cancellation.
- `TestMultiplexSwitch_Options`, case `seeds`: `WithSeeds` populates the switch's seed set.

`tm2/pkg/p2p/discovery/discovery_test.go`:

- `TestReactor_AddPeer`: an outbound peer is asked for peers on connection, an inbound peer is
  not.

`tm2/pkg/bft/node/node_test.go`:

- `TestParseSeedAddrs`: seeds are parsed, invalid entries are dropped, seeds are ignored when
  peer exchange is disabled, and no seeds configured yields no addresses.

## Out of scope

No change is needed to the handshake. `NodeInfo.CompatibleWith` requires only an identical
`Network`, a `VersionSet` compatible on non-optional entries, and one channel in common, and
`discovery.Channel` is appended to the node's channels whenever `pex = true`. A peer
advertising only channel `0x50` is therefore already accepted.

Seed-side behaviour, meaning crawling, address book maintenance and respond-then-disconnect,
is out of scope and requires no core change.
