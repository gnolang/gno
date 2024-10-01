package p2p

import (
	"fmt"
	"log/slog"
	"net"

	"github.com/gnolang/gno/tm2/pkg/cmap"
	connm "github.com/gnolang/gno/tm2/pkg/p2p/conn"
	"github.com/gnolang/gno/tm2/pkg/service"
)

// ConnInfo wraps the remote peer connection
type ConnInfo struct {
	Outbound   bool     // flag indicating if the connection is dialed
	Persistent bool     // flag indicating if the connection is persistent
	Conn       net.Conn // the source connection
	RemoteIP   net.IP   // the remote IP of the peer
	SocketAddr *NetAddress
}

// peer is a wrapper for a remote peer
// Before using a peer, you will need to perform a handshake on connection.
type peer struct {
	service.BaseService

	connInfo *ConnInfo
	remoteIP net.IP
	mconn    *connm.MConnection

	nodeInfo NodeInfo
	data     *cmap.CMap
}

// TODO cleanup
func New(
	connInfo *ConnInfo,
	mConfig connm.MConnConfig,
	nodeInfo NodeInfo,
	reactorsByCh map[byte]Reactor,
	chDescs []*connm.ChannelDescriptor,
	onPeerError func(Peer, interface{}),
) *peer {
	p := &peer{
		connInfo: connInfo,
		nodeInfo: nodeInfo,
		data:     cmap.NewCMap(),
	}

	p.mconn = createMConnection(
		connInfo.Conn,
		p,
		reactorsByCh,
		chDescs,
		onPeerError,
		mConfig,
	)

	p.BaseService = *service.NewBaseService(nil, "Peer", p)

	return p
}

// RemoteIP returns the IP from the remote connection
func (p *peer) RemoteIP() net.IP {
	return p.connInfo.RemoteIP
}

// RemoteAddr returns the address from the remote connection
func (p *peer) RemoteAddr() net.Addr {
	return p.connInfo.Conn.RemoteAddr()
}

func (p *peer) String() string {
	if p.connInfo.Outbound {
		return fmt.Sprintf("Peer{%s %s out}", p.mconn, p.ID())
	}

	return fmt.Sprintf("Peer{%s %s in}", p.mconn, p.ID())
}

func (p *peer) SetLogger(l *slog.Logger) {
	p.Logger = l
	p.mconn.SetLogger(l)
}

func (p *peer) OnStart() error {
	if err := p.BaseService.OnStart(); err != nil {
		return err
	}

	if err := p.mconn.Start(); err != nil {
		return err
	}

	return nil
}

// FlushStop mimics OnStop but additionally ensures that all successful
// .Send() calls will get flushed before closing the connection.
// NOTE: it is not safe to call this method more than once.
func (p *peer) FlushStop() {
	p.BaseService.OnStop()
	p.mconn.FlushStop() // stop everything and close the conn
}

// OnStop implements BaseService.
func (p *peer) OnStop() {
	p.BaseService.OnStop()

	if err := p.mconn.Stop(); err != nil {
		p.Logger.Error(
			"unable to gracefully close mconn",
			"err",
			err,
		)
	}
}

// ID returns the peer's ID - the hex encoded hash of its pubkey.
func (p *peer) ID() ID {
	return p.nodeInfo.NetAddress.ID
}

// IsOutbound returns true if the connection is outbound, false otherwise.
func (p *peer) IsOutbound() bool {
	return p.connInfo.Outbound
}

// IsPersistent returns true if the peer is persistent, false otherwise.
func (p *peer) IsPersistent() bool {
	return p.connInfo.Persistent
}

// NodeInfo returns a copy of the peer's NodeInfo.
func (p *peer) NodeInfo() NodeInfo {
	return p.nodeInfo
}

// SocketAddr returns the address of the socket.
// For outbound peers, it's the address dialed (after DNS resolution).
// For inbound peers, it's the address returned by the underlying connection
// (not what's reported in the peer's NodeInfo).
func (p *peer) SocketAddr() *NetAddress {
	return p.connInfo.SocketAddr
}

// Status returns the peer's ConnectionStatus.
func (p *peer) Status() connm.ConnectionStatus {
	return p.mconn.Status()
}

// Send msg bytes to the channel identified by chID byte. Returns false if the
// send queue is full after timeout, specified by MConnection.
func (p *peer) Send(chID byte, msgBytes []byte) bool {
	if !p.IsRunning() || !p.hasChannel(chID) {
		// see Switch#Broadcast, where we fetch the list of peers and loop over
		// them - while we're looping, one peer may be removed and stopped.
		return false
	}

	return p.mconn.Send(chID, msgBytes)
}

// TrySend msg bytes to the channel identified by chID byte. Immediately returns
// false if the send queue is full.
func (p *peer) TrySend(chID byte, msgBytes []byte) bool {
	if !p.IsRunning() || !p.hasChannel(chID) {
		return false
	}

	return p.mconn.TrySend(chID, msgBytes)
}

// Get the data for a given key.
func (p *peer) Get(key string) interface{} {
	return p.data.Get(key)
}

// Set sets the data for the given key.
func (p *peer) Set(key string, data interface{}) {
	p.data.Set(key, data)
}

// hasChannel returns true if the peer reported
// knowing about the given chID.
func (p *peer) hasChannel(chID byte) bool {
	for _, ch := range p.nodeInfo.Channels {
		if ch == chID {
			return true
		}
	}
	// NOTE: probably will want to remove this
	// but could be helpful while the feature is new
	p.Logger.Debug(
		"Unknown channel for peer",
		"channel",
		chID,
		"channels",
		p.nodeInfo.Channels,
	)
	return false
}

// CloseConn closes original connection. Used for cleaning up in cases where the peer had not been started at all.
func (p *peer) CloseConn() error {
	return p.connInfo.Conn.Close()
}

func createMConnection(
	conn net.Conn,
	p Peer,
	reactorsByCh map[byte]Reactor,
	chDescs []*connm.ChannelDescriptor,
	onPeerError func(Peer, interface{}),
	config connm.MConnConfig,
) *connm.MConnection {
	onReceive := func(chID byte, msgBytes []byte) {
		reactor := reactorsByCh[chID]
		if reactor == nil {
			// Note that its ok to panic here as it's caught in the connm._recover,
			// which does onPeerError.
			panic(fmt.Sprintf("Unknown channel %X", chID))
		}
		reactor.Receive(chID, p, msgBytes)
	}

	onError := func(r interface{}) {
		onPeerError(p, r)
	}

	return connm.NewMConnectionWithConfig(
		conn,
		chDescs,
		onReceive,
		onError,
		config,
	)
}
