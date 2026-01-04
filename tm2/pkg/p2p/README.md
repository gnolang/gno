## Overview

The `p2p` package, and its “sub-packages” contain the required building blocks for Tendermint2’s networking layer.

This document aims to explain the `p2p` terminology, and better document the way the `p2p` module works within the TM2
ecosystem, especially in relation to other modules like `consensus`, `blockchain` and `mempool`.

## Common Types

To fully understand the `Concepts` section of the `p2p` documentation, there must be at least a basic understanding of
the terminology of the `p2p` module, because there are types that keep popping up constantly, and it’s worth
understanding what they’re about.

### `NetAddress`

```go
package types

// NetAddress defines information about a peer on the network
// including its ID, IP address, and port
type NetAddress struct {
	ID   ID     `json:"id"`   // unique peer identifier (public key address)
	IP   net.IP `json:"ip"`   // the IP part of the dial address
	Port uint16 `json:"port"` // the port part of the dial address
}
```

A `NetAddress` is simply a wrapper for a unique peer in the network.

This address consists of several parts:

- the peer’s ID, derived from the peer’s public key (it’s the address).
- the peer’s dial address, used for executing TCP dials.

### `ID`

```go
// ID represents the cryptographically unique Peer ID
type ID = crypto.ID
```

The peer ID is the unique peer identifier. It is used for unambiguously resolving who a peer is, during communication.

The reason the peer ID is utilized is because it is derived from the peer’s public key, used to encrypt communication,
and it needs to match the public key used in p2p communication. It can, and should be, considered unique.

### `Reactor`

Without going too much into detail in the terminology section, as a much more detailed explanation is discussed below:

A `Reactor` is an abstraction of a Tendermint2 module, that needs to utilize the `p2p` layer.

Currently active reactors in TM2, that utilize the p2p layer:

- the consensus reactor, that handles consensus message passing
- the blockchain reactor, that handles block syncing
- the mempool reactor, that handles transaction gossiping

All of these functionalities require a live p2p network to work, and `Reactor`s are the answer for how they can be aware
of things happening in the network (like new peers joining, for example).

## Concepts

### Peer

`Peer` is an abstraction over a p2p connection that is:

- **verified**, meaning it went through the handshaking process and the information the other peer shared checked out (
  this process is discussed in detail later).
- **multiplexed over TCP** (the only kind of p2p connections TM2 supports).

```go
package p2p

// Peer is a wrapper for a connected peer
type Peer interface {
	service.Service

	FlushStop()

	ID() types.ID         // peer's cryptographic ID
	RemoteIP() net.IP     // remote IP of the connection
	RemoteAddr() net.Addr // remote address of the connection

	IsOutbound() bool   // did we dial the peer
	IsPersistent() bool // do we redial this peer when we disconnect
	IsPrivate() bool    // do we share the peer

	CloseConn() error // close original connection

	NodeInfo() types.NodeInfo // peer's info
	Status() ConnectionStatus
	SocketAddr() *types.NetAddress // actual address of the socket

	Send(byte, []byte) bool
	TrySend(byte, []byte) bool

	Set(string, any)
	Get(string) any
}
```

There are more than a few things to break down here, so let’s tackle them individually.

The `Peer` abstraction holds callbacks relating to information about the actual live peer connection, such as what kind
of direction it is, what is the connection status, and others.

```go
package p2p

type Peer interface {
	// ...

	ID() types.ID         // peer's cryptographic ID
	RemoteIP() net.IP     // remote IP of the connection
	RemoteAddr() net.Addr // remote address of the connection

	NodeInfo() types.NodeInfo // peer's info
	Status() ConnectionStatus
	SocketAddr() *types.NetAddress // actual address of the socket

	IsOutbound() bool   // did we dial the peer
	IsPersistent() bool // do we redial this peer when we disconnect
	IsPrivate() bool    // do we share the peer

	// ...
}

```

However, there is part of the `Peer` abstraction that outlines the flipped design of the entire `p2p` module, and a
severe limitation of this implementation.

```go
package p2p

type Peer interface {
	// ...

	Send(byte, []byte) bool
	TrySend(byte, []byte) bool

	// ...
}
```

The `Peer` abstraction is used internally in `p2p`, but also by other modules that need to interact with the networking
layer — this is in itself the biggest crux of the current `p2p` implementation: modules *need to understand* how to use
and communicate with peers, regardless of the protocol logic. Networking is not an abstraction for the modules, but a
spec requirement. What this essentially means is there is heavy implementation leaking to parts of the TM2 codebase that
shouldn’t need to know how to handle individual peer broadcasts, or how to trigger custom protocol communication (like
syncing for example).
If `module A` wants to broadcast something to the peer network of the node, it needs to do something like this:

```go
package main

func main() {
	// ...

	peers := sw.Peers().List() // fetch the peer list

	for _, p := range peers {
		p.Send(...) // directly message the peer (imitate broadcast)
	}

	// ...
}
```

An additional odd choice in the `Peer` API is the ability to use the peer as a KV store:

```go
package p2p

type Peer interface {
	// ...

	Set(string, any)
	Get(string) any

	// ...
}
```

For example, these methods are used within the `consensus` and `mempool` modules to keep track of active peer states (
like current HRS data, or current peer mempool metadata). Instead of the module handling individual peer state, this
responsibility is shifted to the peer implementation, causing an odd code dependency situation.

The root of this “flipped” design (modules needing to understand how to interact with peers) stems from the fact that
peers are instantiated with a multiplex TCP connection under the hood, and basically just wrap that connection. The
`Peer` API is an abstraction for the multiplexed TCP connection, under the hood.

Changing this dependency stew would require a much larger rewrite of not just the `p2p` module, but other modules (
`consensus`, `blockchain`, `mempool`) as well, and is as such left as-is.

### Switch

In short, a `Switch` is just the middleware layer that handles module <> `Transport` requests, and manages peers on a
high application level (that the `Transport` doesn’t concern itself with).

The `Switch` is the entity that manages active peer connections.

```go
package p2p

// Switch is the abstraction in the p2p module that handles
// and manages peer connections thorough a Transport
type Switch interface {
	// Broadcast publishes data on the given channel, to all peers
	Broadcast(chID byte, data []byte)

	// Peers returns the latest peer set
	Peers() PeerSet

	// Subscribe subscribes to active switch events
	Subscribe(filterFn events.EventFilter) (<-chan events.Event, func())

	// StopPeerForError stops the peer with the given reason
	StopPeerForError(peer Peer, err error)

	// DialPeers marks the given peers as ready for async dialing
	DialPeers(peerAddrs ...*types.NetAddress)
}

```

The API of the `Switch` is relatively straightforward. Users of the `Switch` instantiate it with a `Transport`, and
utilize it as-is.

The API of the `Switch` is geared towards asynchronicity, and as such users of the `Switch` need to adapt to some
limitations, such as not having synchronous dials, or synchronous broadcasts.

#### Services

There are 3 services that run on top of the `MultiplexSwitch`, upon startup:

- **the accept service**
- **the dial service**
- **the redial service**

```go
package p2p

// OnStart implements BaseService. It starts all the reactors and peers.
func (sw *MultiplexSwitch) OnStart() error {
	// Start reactors
	for _, reactor := range sw.reactors {
		if err := reactor.Start(); err != nil {
			return fmt.Errorf("unable to start reactor %w", err)
		}
	}

	// Run the peer accept routine.
	// The accept routine asynchronously accepts
	// and processes incoming peer connections
	go sw.runAcceptLoop(sw.ctx)

	// Run the dial routine.
	// The dial routine parses items in the dial queue
	// and initiates outbound peer connections
	go sw.runDialLoop(sw.ctx)

	// Run the redial routine.
	// The redial routine monitors for important
	// peer disconnects, and attempts to reconnect
	// to them
	go sw.runRedialLoop(sw.ctx)

	return nil
}
```

##### Accept Service

The `MultiplexSwitch` needs to actively listen for incoming connections, and handle them accordingly. These situations
occur when a peer *Dials* (more on this later) another peer, and wants to establish a connection. This connection is
outbound for one peer, and inbound for the other.

Depending on what kind of security policies or configuration the peer has in place, the connection can be accepted, or
rejected for a number of reasons:

- the maximum number of inbound peers is reached
- the multiplex connection fails upon startup (rare)

The `Switch` relies on the `Transport` to return a **verified and valid** peer connection. After the `Transport`
delivers, the `Switch` makes sure having the peer makes sense, given the p2p configuration of the node.

```go
package p2p

func (sw *MultiplexSwitch) runAcceptLoop(ctx context.Context) {
	// ...

	p, err := sw.transport.Accept(ctx, sw.peerBehavior)
	if err != nil {
		sw.Logger.Error(
			"error encountered during peer connection accept",
			"err", err,
		)

		continue
	}

	// Ignore connection if we already have enough peers.
	if in := sw.Peers().NumInbound(); in >= sw.maxInboundPeers {
		sw.Logger.Info(
			"Ignoring inbound connection: already have enough inbound peers",
			"address", p.SocketAddr(),
			"have", in,
			"max", sw.maxInboundPeers,
		)

		sw.transport.Remove(p)

		continue
	}

	// ...
}

```

In fact, this is the central point in the relationship between the `Switch` and `Transport`.
The `Transport` is responsible for establishing the connection, and the `Switch` is responsible for handling it after
it’s been established.

When TM2 modules communicate with the `p2p` module, they communicate *with the `Switch`, not the `Transport`* to execute
peer-related actions.

##### Dial Service

Peers are dialed asynchronously in the `Switch`, as is suggested by the `Switch` API:

```go
DialPeers(peerAddrs ...*types.NetAddress)
```

The `MultiplexSwitch` implementation utilizes a concept called a *dial queue*.

A dial queue is a priority-based queue (sorted by dial time, ascending) from which dial requests are taken out of and
executed in the form of peer dialing (through the `Transport`, of course).

The queue needs to be sorted by the dial time, since there are asynchronous dial requests that need to be executed as
soon as possible, while others can wait to be executed up until a certain point in time.

```go
package p2p

func (sw *MultiplexSwitch) runDialLoop(ctx context.Context) {
	// ...

	// Grab a dial item
	item := sw.dialQueue.Peek()
	if item == nil {
		// Nothing to dial
		continue
	}

	// Check if the dial time is right
	// for the item
	if time.Now().Before(item.Time) {
		// Nothing to dial
		continue
	}

	// Pop the item from the dial queue
	item = sw.dialQueue.Pop()

	// Dial the peer
	sw.Logger.Info(
		"dialing peer",
		"address", item.Address.String(),
	)

	// ...
}
```

To follow the outcomes of dial requests, users of the `Switch` can subscribe to peer events (more on this later).

##### Redial Service

The TM2 `p2p` module has a concept of something called *persistent peers*.

Persistent peers are specific peers whose connections must be preserved, at all costs. They are specified in the
top-level node P2P configuration, under `p2p.persistent_peers`.

These peer connections are special, as they don’t adhere to high-level configuration limits like the maximum peer cap,
instead, they are monitored and handled actively.

A good candidate for a persistent peer is a bootnode, that bootstraps and facilitates peer discovery for the network.

If a persistent peer connection is lost for whatever reason (for ex, the peer disconnects), the redial service of the
`MultiplexSwitch` will create a dial request for the dial service, and attempt to re-establish the lost connection.

```go
package p2p

func (sw *MultiplexSwitch) runRedialLoop(ctx context.Context) {
	// ...

	var (
		peers       = sw.Peers()
		peersToDial = make([]*types.NetAddress, 0)
	)

	sw.persistentPeers.Range(func(key, value any) bool {
		var (
			id   = key.(types.ID)
			addr = value.(*types.NetAddress)
		)

		// Check if the peer is part of the peer set
		// or is scheduled for dialing
		if peers.Has(id) || sw.dialQueue.Has(addr) {
			return true
		}

		peersToDial = append(peersToDial, addr)

		return true
	})

	if len(peersToDial) == 0 {
		// No persistent peers are missing
		return
	}

	// Add the peers to the dial queue
	sw.DialPeers(peersToDial...)

	// ...
}
```

#### Events

The `Switch` is meant to be asynchronous.

This means that processes like dialing peers, removing peers, doing broadcasts and more, is not a synchronous blocking
process for the `Switch` user.

To be able to tap into the outcome of these asynchronous events, the `Switch` utilizes a simple event system, based on
event filters.

```go
package main

func main() {
	// ...

	// Subscribe to live switch events
	ch, unsubFn := multiplexSwitch.Subscribe(func(event events.Event) bool {
		// This subscription will only return "PeerConnected" events
		return event.Type() == events.PeerConnected
	})

	defer unsubFn() // removes the subscription

	select {
	// Events are sent to the channel as soon as
	// they appear and pass the subscription filter
	case ev <- ch:
		e := ev.(*events.PeerConnectedEvent)
		// use event data...
	case <-ctx.Done():
		// ...
	}

	// ...
}
```

An event setup like this is useful for example when the user of the `Switch` wants to capture successful peer dial
events, in realtime.

#### What is “peer behavior”?

```go
package p2p

// PeerBehavior wraps the Reactor and MultiplexSwitch information a Transport would need when
// dialing or accepting new Peer connections.
// It is worth noting that the only reason why this information is required in the first place,
// is because Peers expose an API through which different TM modules can interact with them.
// In the future™, modules should not directly "Send" anything to Peers, but instead communicate through
// other mediums, such as the P2P module
type PeerBehavior interface {
	// ReactorChDescriptors returns the Reactor channel descriptors
	ReactorChDescriptors() []*conn.ChannelDescriptor

	// Reactors returns the node's active p2p Reactors (modules)
	Reactors() map[byte]Reactor

	// HandlePeerError propagates a peer connection error for further processing
	HandlePeerError(Peer, error)

	// IsPersistentPeer returns a flag indicating if the given peer is persistent
	IsPersistentPeer(types.ID) bool

	// IsPrivatePeer returns a flag indicating if the given peer is private
	IsPrivatePeer(types.ID) bool
}

```

In short, the previously-mentioned crux of the `p2p` implementation (having `Peer`s be directly managed by different TM2
modules)  requires information on how to behave when interacting with other peers.

TM2 modules on `peer A` communicate through something called *channels* to the same modules on `peer B`. For example, if
the `mempool` module on `peer A` wants to share a transaction to the mempool module on `peer B`, it will utilize a
dedicated (and unique!) channel for it (ex. `0x30`). This is a protocol that lives on top of the already-established
multiplexed connection, and metadata relating to it is passed down through *peer behavior*.

### Transport

As previously mentioned, the `Transport` is the infrastructure layer of the `p2p` module.

In contrast to the `Switch`, which is concerned with higher-level application logic (like the number of peers, peer
limits, etc), the `Transport` is concerned with actually establishing and maintaining peer connections on a much lower
level.

```go
package p2p

// Transport handles peer dialing and connection acceptance. Additionally,
// it is also responsible for any custom connection mechanisms (like handshaking).
// Peers returned by the transport are considered to be verified and sound
type Transport interface {
	// NetAddress returns the Transport's dial address
	NetAddress() types.NetAddress

	// Accept returns a newly connected inbound peer
	Accept(context.Context, PeerBehavior) (Peer, error)

	// Dial dials a peer, and returns it
	Dial(context.Context, types.NetAddress, PeerBehavior) (Peer, error)

	// Remove drops any resources associated
	// with the Peer in the transport
	Remove(Peer)
}
```

When peers dial other peers in TM2, they are in fact dialing their `Transport`s, and the connection is being handled
here.

- `Accept` waits for an **incoming** connection, parses it and returns it.
- `Dial` attempts to establish an **outgoing** connection, parses it and returns it.

There are a few important steps that happen when establishing a p2p connection in TM2, between 2 different peers:

1. The peers go through a handshaking process, and establish something called a *secret connection*. The handshaking
   process is based on the [STS protocol](https://github.com/tendermint/tendermint/blob/0.1/docs/sts-final.pdf), and
   after it is completed successfully, all communication between the 2 peers is **encrypted**.
2. After establishing a secret connection, the peers exchange their respective node information. The purpose of this
   step is to verify that the peers are indeed compatible with each other, and should be establishing a connection in
   the first place (same network, common protocols , etc).
3. Once the secret connection is established, and the node information is exchanged, the connection to the peer is
   considered valid and verified — it can now be used by the `Switch` (accepted, or rejected, based on `Switch`
   high-level constraints). Note the distinction here that the `Transport` establishes and maintains the connection, but
   it can ultimately be scraped by the `Switch` at any point in time.

### Peer Discovery

There is a final service that runs alongside the previously-mentioned `Switch` services — peer discovery.

Every blockchain node needs an adequate amount of peers to communicate with, in order to ensure smooth functioning. For
validator nodes, they need to be *loosely connected* to at least 2/3+ of the validator set in order to participate and
not cause block misses or mis-votes (loosely connected means that there always exists a path between different peers in
the network topology, that allows them to be reachable to each other).

The peer discovery service ensures that the given node is always learning more about the overall network topology, and
filling out any empty connection slots (outbound peers).

This background service works in the following (albeit primitive) way:

1. At specific intervals, `node A` checks its peer table, and picks a random peer `P`, from the active peer list.
2. When `P` is picked, `node A` initiates a discovery protocol process, in which:
    - `node A` sends a request to peer `P` for his peer list (max 30 peers)
    - peer `P` responds to the request

3. Once `node A` has the peer list from `P`, it adds the entire peer list into the dial queue, to establish outbound
   peer connections.

This process repeats at specific intervals. It is worth nothing that if the limit of outbound peers is reached, the peer
dials have no effect.

#### Bootnodes (Seeds)

Bootnodes are specialized network nodes that play a critical role in the initial peer discovery process for new nodes
joining the network.

When a blockchain client starts, it needs to find and connect to other nodes to synchronize data and participate in the
network. Bootnodes provide a predefined list of accessible and reliable entry points that act as a gateway for
discovering other active nodes (through peer discovery).

These nodes are provided as part of the node’s p2p configuration. Once connected to a bootnode, the client uses peer
discovery to discover and connect to additional peers, enabling full participation and unlocking other client
protocols (consensus, mempool…).

Bootnodes usually do not store the full blockchain or participate in consensus; their primary role is to facilitate
connectivity in the network (act as a peer relay).