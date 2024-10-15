package p2p

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/p2p/conn"
	"golang.org/x/sync/errgroup"
)

const (
	defaultDialTimeout      = time.Second
	defaultHandshakeTimeout = 3 * time.Second
)

// inboundPeer is a wrapper for incoming peer information
type inboundPeer struct {
	addr     *NetAddress // the dial address of the peer
	conn     net.Conn    // the connection associated with the peer
	nodeInfo NodeInfo    // the relevant peer node info
}

// peerConfig is used to bundle data we need to fully setup a Peer with an
// MConn, provided by the caller of Accept and Dial (currently the Switch). This
// a temporary measure until reactor setup is less dynamic and we introduce the
// concept of PeerBehaviour to communicate about significant Peer lifecycle
// events.
// TODO(xla): Refactor out with more static Reactor setup and PeerBehaviour.
type peerConfig struct {
	chDescs     []*conn.ChannelDescriptor
	onPeerError func(Peer, error)
	outbound    bool
	// isPersistent allows you to set a function, which, given socket address
	// (for outbound peers) OR self-reported address (for inbound peers), tells
	// if the peer is persistent or not.
	isPersistent func(*NetAddress) bool
	reactorsByCh map[byte]Reactor
}

// MultiplexTransport accepts and dials tcp connections and upgrades them to
// multiplexed peers.
type MultiplexTransport struct {
	ctx      context.Context
	cancelFn context.CancelFunc

	netAddr  NetAddress // the node's P2P dial address, used for handshaking
	nodeInfo NodeInfo   // the node's P2P info, used for handshaking
	nodeKey  NodeKey    // the node's private P2P key, used for handshaking

	listener net.Listener     // listener for inbound peer connections
	peerCh   chan inboundPeer // pipe for inbound peer connections

	conns ConnSet // lookup

	handshakeTimeout time.Duration

	// TODO(xla): This config is still needed as we parameterize peerConn and
	// peer currently. All relevant configuration should be refactored into options
	// with sane defaults.
	mConfig conn.MConnConfig
}

// Test multiplexTransport for interface completeness.
var _ Transport = (*MultiplexTransport)(nil)

// NewMultiplexTransport returns a tcp connected multiplexed peer.
func NewMultiplexTransport(
	nodeInfo NodeInfo,
	nodeKey NodeKey,
	mConfig conn.MConnConfig,
) *MultiplexTransport {
	return &MultiplexTransport{
		peerCh:           make(chan inboundPeer, 1),
		handshakeTimeout: defaultHandshakeTimeout,
		mConfig:          mConfig,
		nodeInfo:         nodeInfo,
		nodeKey:          nodeKey,
		conns:            NewConnSet(),
	}
}

// NetAddress implements Transport.
func (mt *MultiplexTransport) NetAddress() NetAddress {
	return mt.netAddr
}

// Accept waits for a verified inbound Peer to connect, and returns it [BLOCKING]
func (mt *MultiplexTransport) Accept(ctx context.Context, cfg peerConfig) (Peer, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case info, ok := <-mt.peerCh:
		if !ok {
			return nil, errors.New("transport closed") // TODO make standard
		}

		// cfg.outbound = false TODO integrate

		return mt.wrapPeer(info, cfg)
	}
}

// Dial creates an outbound verified Peer connection [BLOCKING]
func (mt *MultiplexTransport) Dial(
	ctx context.Context,
	addr NetAddress,
	cfg peerConfig,
) (Peer, error) {
	// Set a dial timeout for the connection
	c, err := addr.DialContext(ctx)
	if err != nil {
		return nil, err
	}

	// Check if the connection is a duplicate one
	if mt.conns.Has(c) {
		// Close the connection
		_ = c.Close()

		return nil, errors.New("duplicate peer connection")
	}

	// Handshake with the peer
	secretConn, nodeInfo, err := mt.upgrade(c, &addr)
	if err != nil {
		return nil, err
	}

	cfg.outbound = true

	info := inboundPeer{
		addr:     &addr,
		conn:     secretConn,
		nodeInfo: nodeInfo,
	}

	return mt.wrapPeer(info, cfg)
}

// Close implements TransportLifecycle.
func (mt *MultiplexTransport) Close() error {
	mt.cancelFn()

	if mt.listener == nil {
		return nil
	}

	return mt.listener.Close()
}

// Listen implements TransportLifecycle.
func (mt *MultiplexTransport) Listen(addr NetAddress) error {
	ln, err := net.Listen("tcp", addr.DialString())
	if err != nil {
		return fmt.Errorf("unable to listen on address, %w", err)
	}

	if addr.Port == 0 {
		// net.Listen on port 0 means the kernel will auto-allocate a port
		// - find out which one has been given to us.
		tcpAddr, ok := ln.Addr().(*net.TCPAddr)
		if !ok {
			return fmt.Errorf("error finding port (after listening on port 0): %w", err)
		}

		addr.Port = uint16(tcpAddr.Port)
	}

	mt.netAddr = addr
	mt.listener = ln

	go mt.runAcceptLoop()

	return nil
}

// runAcceptLoop runs the loop where incoming peers are:
// - 1. accepted by the transport
// - 2. filtered
// - 3. upgraded (handshaked + verified)
func (mt *MultiplexTransport) runAcceptLoop() {
	defer close(mt.peerCh)

	for {
		select {
		case <-mt.ctx.Done():
			return
		default:
			// Accept an incoming peer connection
			c, err := mt.listener.Accept()
			if err != nil {
				// TODO log accept error
				continue
			}

			// Check if the connection is a duplicate one
			if mt.conns.Has(c) {
				// TODO add warn log
				continue
			}

			// Connection upgrade and filtering should be asynchronous to avoid
			// Head-of-line blocking[0].
			// Reference:  https://github.com/tendermint/classic/issues/2047
			//
			// [0] https://en.wikipedia.org/wiki/Head-of-line_blocking
			go func(c net.Conn) {
				// TODO extract common logic with Dial()
				var (
					nodeInfo   NodeInfo
					secretConn *conn.SecretConnection
				)

				secretConn, nodeInfo, err = mt.upgrade(c, nil)
				if err != nil {
					// TODO add error log
					return
				}

				var (
					addr       = c.RemoteAddr()
					id         = secretConn.RemotePubKey().Address().ID()
					netAddr, _ = NewNetAddress(id, addr)
				)

				p := inboundPeer{
					addr:     netAddr,
					conn:     c,
					nodeInfo: nodeInfo,
				}

				select {
				case mt.peerCh <- p:
				case <-mt.ctx.Done():
					// Give up if the transport was closed.
					_ = c.Close()
				}
			}(c)
		}
	}
}

// Remove removes the given address from the connections set and
// closes the connection.
func (mt *MultiplexTransport) Remove(p Peer) {
	mt.conns.RemoveAddr(p.RemoteAddr())
	_ = p.CloseConn()
}

func (mt *MultiplexTransport) cleanup(c net.Conn) error {
	mt.conns.Remove(c)

	return c.Close()
}

func (mt *MultiplexTransport) upgrade(
	c net.Conn,
	dialedAddr *NetAddress,
) (secretConn *conn.SecretConnection, nodeInfo NodeInfo, err error) {
	defer func() {
		if err != nil {
			_ = mt.cleanup(c)
		}
	}()

	secretConn, err = upgradeSecretConn(c, mt.handshakeTimeout, mt.nodeKey.PrivKey)
	if err != nil {
		return nil, NodeInfo{}, RejectedError{
			conn:          c,
			err:           fmt.Errorf("secret conn failed: %w", err),
			isAuthFailure: true,
		}
	}

	// For outgoing conns, ensure connection key matches dialed key.
	connID := secretConn.RemotePubKey().Address().ID()
	if dialedAddr != nil {
		if dialedID := dialedAddr.ID; connID.String() != dialedID.String() {
			return nil, NodeInfo{}, RejectedError{
				conn: c,
				id:   connID,
				err: fmt.Errorf(
					"conn.ID (%v) dialed ID (%v) mismatch",
					connID,
					dialedID,
				),
				isAuthFailure: true,
			}
		}
	}

	nodeInfo, err = handshake(secretConn, mt.handshakeTimeout, mt.nodeInfo)
	if err != nil {
		return nil, NodeInfo{}, RejectedError{
			conn:          c,
			err:           fmt.Errorf("handshake failed: %w", err),
			isAuthFailure: true,
		}
	}

	if err := nodeInfo.Validate(); err != nil {
		return nil, NodeInfo{}, RejectedError{
			conn:              c,
			err:               err,
			isNodeInfoInvalid: true,
		}
	}

	// Ensure connection key matches self reported key.
	if connID != nodeInfo.ID() {
		return nil, NodeInfo{}, RejectedError{
			conn: c,
			id:   connID,
			err: fmt.Errorf(
				"conn.ID (%v) NodeInfo.ID (%v) mismatch",
				connID,
				nodeInfo.ID(),
			),
			isAuthFailure: true,
		}
	}

	// Reject self.
	if mt.nodeInfo.ID() == nodeInfo.ID() {
		addr, err := NewNetAddress(nodeInfo.ID(), c.RemoteAddr())
		if err != nil {
			return nil, NodeInfo{}, NetAddressInvalidError{
				Addr: c.RemoteAddr().String(),
				Err:  err,
			}
		}

		return nil, NodeInfo{}, RejectedError{
			addr:   *addr,
			conn:   c,
			id:     nodeInfo.ID(),
			isSelf: true,
		}
	}

	if err := mt.nodeInfo.CompatibleWith(nodeInfo); err != nil {
		return nil, NodeInfo{}, RejectedError{
			conn:           c,
			err:            err,
			id:             nodeInfo.ID(),
			isIncompatible: true,
		}
	}

	return secretConn, nodeInfo, nil
}

func (mt *MultiplexTransport) wrapPeer(
	info inboundPeer,
	cfg peerConfig,
) (Peer, error) {
	persistent := false
	if cfg.isPersistent != nil {
		if cfg.outbound {
			persistent = cfg.isPersistent(info.addr)
		} else {
			selfReportedAddr := info.nodeInfo.NetAddress
			persistent = cfg.isPersistent(selfReportedAddr)
		}
	}

	// Extract the host
	host, _, err := net.SplitHostPort(info.conn.RemoteAddr().String())
	if err != nil {
		return nil, fmt.Errorf("unable to extract peer host, %w", err)
	}

	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, fmt.Errorf("unable to lookup peer IPs, %w", err)
	}

	peerConn := &ConnInfo{
		Outbound:   cfg.outbound,
		Persistent: persistent,
		Conn:       info.conn,
		RemoteIP:   ips[0],
		SocketAddr: info.addr,
	}

	mConfig := &MultiplexConnConfig{
		MConfig:      mt.mConfig,
		ReactorsByCh: cfg.reactorsByCh,
		ChDescs:      cfg.chDescs,
		OnPeerError:  cfg.onPeerError,
	}

	return NewPeer(peerConn, info.nodeInfo, mConfig), nil
}

func handshake(
	c net.Conn,
	timeout time.Duration,
	nodeInfo NodeInfo,
) (NodeInfo, error) {
	if err := c.SetDeadline(time.Now().Add(timeout)); err != nil {
		return NodeInfo{}, err
	}

	var (
		peerNodeInfo NodeInfo
		ourNodeInfo  = nodeInfo
	)

	g, _ := errgroup.WithContext(context.Background())

	g.Go(func() error {
		_, err := amino.MarshalSizedWriter(c, ourNodeInfo)

		return err
	})

	g.Go(func() error {
		_, err := amino.UnmarshalSizedReader(
			c,
			&peerNodeInfo,
			MaxNodeInfoSize,
		)

		return err
	})

	if err := g.Wait(); err != nil {
		return NodeInfo{}, err
	}

	return peerNodeInfo, c.SetDeadline(time.Time{})
}

func upgradeSecretConn(
	c net.Conn,
	timeout time.Duration,
	privKey crypto.PrivKey,
) (*conn.SecretConnection, error) {
	if err := c.SetDeadline(time.Now().Add(timeout)); err != nil {
		return nil, err
	}

	sc, err := conn.MakeSecretConnection(c, privKey)
	if err != nil {
		return nil, err
	}

	return sc, sc.SetDeadline(time.Time{})
}
