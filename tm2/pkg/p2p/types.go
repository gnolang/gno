package p2p

import (
	"net"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/p2p/conn"
	connm "github.com/gnolang/gno/tm2/pkg/p2p/conn"
	"github.com/gnolang/gno/tm2/pkg/service"
)

type (
	ChannelDescriptor = conn.ChannelDescriptor
	ConnectionStatus  = conn.ConnectionStatus
)

type ID = crypto.ID

// Peer is a wrapper for a connected peer
type Peer interface {
	service.Service

	FlushStop()

	ID() ID               // peer's cryptographic ID
	RemoteIP() net.IP     // remote IP of the connection
	RemoteAddr() net.Addr // remote address of the connection

	IsOutbound() bool   // did we dial the peer
	IsPersistent() bool // do we redial this peer when we disconnect

	CloseConn() error // close original connection

	NodeInfo() NodeInfo // peer's info
	Status() connm.ConnectionStatus
	SocketAddr() *NetAddress // actual address of the socket

	Send(byte, []byte) bool
	TrySend(byte, []byte) bool

	Set(string, any)
	Get(string) any
}

// PeerSet has a (immutable) subset of the methods of PeerSet.
type PeerSet interface {
	Add(peer Peer)
	Remove(key ID) bool
	Has(key ID) bool
	HasIP(ip net.IP) bool
	Get(key ID) Peer
	List() []Peer
	Size() int
}
