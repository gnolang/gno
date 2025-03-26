package p2p

import (
	"fmt"
	"log/slog"
	"net"
	"slices"

	"github.com/gnolang/gno/tm2/pkg/cmap"
	"github.com/gnolang/gno/tm2/pkg/p2p/conn"
	"github.com/gnolang/gno/tm2/pkg/p2p/types"
	"github.com/gnolang/gno/tm2/pkg/service"
)

type ConnConfig struct {
	MConfig      conn.MConnConfig
	ReactorsByCh map[byte]Reactor
	ChDescs      []*conn.ChannelDescriptor
	OnPeerError  func(PeerConn, error)
}

// ConnInfo wraps the remote peer connection
type ConnInfo struct {
	Outbound   bool     // flag indicating if the connection is dialed
	Persistent bool     // flag indicating if the connection is persistent
	Private    bool     // flag indicating if the peer is private (not shared)
	Conn       net.Conn // the source connection
	RemoteIP   net.IP   // the remote IP of the peer
	SocketAddr *types.NetAddress
}

type multiplexConn interface {
	FlushStop()
	Start() error
	Stop() error
	Send(byte, []byte) bool
	TrySend(byte, []byte) bool
	SetLogger(*slog.Logger)
	Status() conn.ConnectionStatus
	String() string
}

// peer is a wrapper for a remote peer
// Before using a peer, you will need to perform a handshake on connection.
type peer struct {
	service.BaseService

	connInfo *ConnInfo      // Metadata about the connection
	nodeInfo types.NodeInfo // Information about the peer's node
	mConn    multiplexConn  // The multiplexed connection

	data *cmap.CMap // Arbitrary data store associated with the peer
}

// newPeer creates an uninitialized peer instance
func newPeer(
	connInfo *ConnInfo,
	nodeInfo types.NodeInfo,
	mConfig *ConnConfig,
) PeerConn {
	p := &peer{
		connInfo: connInfo,
		nodeInfo: nodeInfo,
		data:     cmap.NewCMap(),
	}

	p.mConn = p.createMConnection(
		connInfo.Conn,
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
		return fmt.Sprintf("Peer{%s %s out}", p.mConn, p.ID())
	}

	return fmt.Sprintf("Peer{%s %s in}", p.mConn, p.ID())
}

// IsOutbound returns true if the connection is outbound, false otherwise.
func (p *peer) IsOutbound() bool {
	return p.connInfo.Outbound
}

// IsPersistent returns true if the peer is persistent, false otherwise.
func (p *peer) IsPersistent() bool {
	return p.connInfo.Persistent
}

// IsPrivate returns true if the peer is private, false otherwise.
func (p *peer) IsPrivate() bool {
	return p.connInfo.Private
}

// SocketAddr returns the address of the socket.
// For outbound peers, it's the address dialed (after DNS resolution).
// For inbound peers, it's the address returned by the underlying connection
// (not what's reported in the peer's NodeInfo).
func (p *peer) SocketAddr() *types.NetAddress {
	return p.connInfo.SocketAddr
}

// CloseConn closes original connection.
// Used for cleaning up in cases where the peer had not been started at all.
func (p *peer) CloseConn() error {
	return p.connInfo.Conn.Close()
}

func (p *peer) SetLogger(l *slog.Logger) {
	p.Logger = l
	p.mConn.SetLogger(l)
}

func (p *peer) OnStart() error {
	if err := p.BaseService.OnStart(); err != nil {
		return fmt.Errorf("unable to start base service, %w", err)
	}

	if err := p.mConn.Start(); err != nil {
		return fmt.Errorf("unable to start multiplex connection, %w", err)
	}

	return nil
}

// FlushStop mimics OnStop but additionally ensures that all successful
// .Send() calls will get flushed before closing the connection.
// NOTE: it is not safe to call this method more than once.
func (p *peer) FlushStop() {
	p.BaseService.OnStop()
	p.mConn.FlushStop() // stop everything and close the conn
}

// OnStop implements BaseService.
func (p *peer) OnStop() {
	p.BaseService.OnStop()

	if err := p.mConn.Stop(); err != nil {
		p.Logger.Error(
			"unable to gracefully close mConn",
			"err",
			err,
		)
	}
}

// ID returns the peer's ID - the hex encoded hash of its pubkey.
func (p *peer) ID() types.ID {
	return p.nodeInfo.ID()
}

// NodeInfo returns a copy of the peer's NodeInfo.
func (p *peer) NodeInfo() types.NodeInfo {
	return p.nodeInfo
}

// Status returns the peer's ConnectionStatus.
func (p *peer) Status() conn.ConnectionStatus {
	return p.mConn.Status()
}

// Send msg bytes to the channel identified by chID byte. Returns false if the
// send queue is full after timeout, specified by MConnection.
func (p *peer) Send(chID byte, msgBytes []byte) bool {
	if !p.IsRunning() || !p.hasChannel(chID) {
		// see MultiplexSwitch#Broadcast, where we fetch the list of peers and loop over
		// them - while we're looping, one peer may be removed and stopped.
		return false
	}

	return p.mConn.Send(chID, msgBytes)
}

// TrySend msg bytes to the channel identified by chID byte. Immediately returns
// false if the send queue is full.
func (p *peer) TrySend(chID byte, msgBytes []byte) bool {
	if !p.IsRunning() || !p.hasChannel(chID) {
		return false
	}

	return p.mConn.TrySend(chID, msgBytes)
}

// Get the data for a given key.
func (p *peer) Get(key string) any {
	return p.data.Get(key)
}

// Set sets the data for the given key.
func (p *peer) Set(key string, data any) {
	p.data.Set(key, data)
}

// hasChannel returns true if the peer reported
// knowing about the given chID.
func (p *peer) hasChannel(chID byte) bool {
	return slices.Contains(p.nodeInfo.Channels, chID)
}

func (p *peer) createMConnection(
	c net.Conn,
	config *ConnConfig,
) *conn.MConnection {
	onReceive := func(chID byte, msgBytes []byte) {
		reactor := config.ReactorsByCh[chID]
		if reactor == nil {
			// Note that its ok to panic here as it's caught in the connm._recover,
			// which does onPeerError.
			panic(fmt.Sprintf("Unknown channel %X", chID))
		}

		reactor.Receive(chID, p, msgBytes)
	}

	onError := func(r error) {
		config.OnPeerError(p, r)
	}

	return conn.NewMConnectionWithConfig(
		c,
		config.ChDescs,
		onReceive,
		onError,
		config.MConfig,
	)
}
