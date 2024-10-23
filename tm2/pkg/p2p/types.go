package p2p

import (
	"context"
	"net"

	"github.com/gnolang/gno/tm2/pkg/p2p/conn"
	"github.com/gnolang/gno/tm2/pkg/p2p/events"
	"github.com/gnolang/gno/tm2/pkg/p2p/types"
	"github.com/gnolang/gno/tm2/pkg/service"
)

type (
	ChannelDescriptor = conn.ChannelDescriptor
	ConnectionStatus  = conn.ConnectionStatus
)

// Peer is a wrapper for a connected peer
type Peer interface {
	service.Service

	FlushStop()

	ID() types.ID         // peer's cryptographic ID
	RemoteIP() net.IP     // remote IP of the connection
	RemoteAddr() net.Addr // remote address of the connection

	IsOutbound() bool   // did we dial the peer
	IsPersistent() bool // do we redial this peer when we disconnect

	CloseConn() error // close original connection

	NodeInfo() types.NodeInfo // peer's info
	Status() ConnectionStatus
	SocketAddr() *types.NetAddress // actual address of the socket

	Send(byte, []byte) bool
	TrySend(byte, []byte) bool

	Set(string, any)
	Get(string) any
}

// PeerSet has a (immutable) subset of the methods of PeerSet.
type PeerSet interface {
	Add(peer Peer)
	Remove(key types.ID) bool
	Has(key types.ID) bool
	HasIP(ip net.IP) bool
	Get(key types.ID) Peer
	List() []Peer // TODO consider implementing an iterator
	Size() int    // TODO remove

	NumInbound() uint64  // returns the number of connected inbound nodes
	NumOutbound() uint64 // returns the number of connected outbound nodes
}

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

// Switch is the abstraction in the p2p module that handles
// and manages peer connections thorough a Transport
type Switch interface {
	// Subscribe subscribes to peer events on the switch
	Subscribe(filterFn events.EventFilter) (<-chan events.Event, func())

	// Broadcast publishes data on the given channel, to all peers
	Broadcast(chID byte, data []byte)

	// Peers returns the latest peer set
	Peers() PeerSet

	// StopPeerForError stops the peer with the given reason
	StopPeerForError(peer Peer, err error)

	// DialPeers marks the given peers as ready for async dialing
	DialPeers(peerAddrs ...*types.NetAddress)
}

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
	IsPersistentPeer(*types.NetAddress) bool
}
