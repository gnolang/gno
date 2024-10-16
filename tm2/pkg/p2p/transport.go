package p2p

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/p2p/conn"
	"golang.org/x/sync/errgroup"
)

// TODO make options
const (
	defaultDialTimeout      = time.Second
	defaultHandshakeTimeout = 3 * time.Second
)

// peerInfo is a wrapper for an unverified peer connection
type peerInfo struct {
	addr     *NetAddress // the dial address of the peer
	conn     net.Conn    // the connection associated with the peer
	nodeInfo NodeInfo    // the relevant peer node info
}

// MultiplexTransport accepts and dials tcp connections and upgrades them to
// multiplexed peers.
type MultiplexTransport struct {
	ctx      context.Context
	cancelFn context.CancelFunc

	logger *slog.Logger

	netAddr  NetAddress // the node's P2P dial address, used for handshaking
	nodeInfo NodeInfo   // the node's P2P info, used for handshaking
	nodeKey  NodeKey    // the node's private P2P key, used for handshaking

	listener    net.Listener  // listener for inbound peer connections
	peerCh      chan peerInfo // pipe for inbound peer connections
	activeConns sync.Map      // active peer connections (remote address -> nothing)

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
		peerCh:           make(chan peerInfo, 1),
		handshakeTimeout: defaultHandshakeTimeout,
		mConfig:          mConfig,
		nodeInfo:         nodeInfo,
		nodeKey:          nodeKey,
	}
}

// NetAddress returns the transport's listen address (for p2p connections)
func (mt *MultiplexTransport) NetAddress() NetAddress {
	return mt.netAddr
}

// Accept waits for a verified inbound Peer to connect, and returns it [BLOCKING]
func (mt *MultiplexTransport) Accept(ctx context.Context, behavior PeerBehavior) (Peer, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case info, ok := <-mt.peerCh:
		if !ok {
			return nil, errors.New("transport closed") // TODO make constant
		}

		return mt.newMultiplexPeer(info, behavior, false)
	}
}

// Dial creates an outbound Peer connection, and
// verifies it (performs handshaking) [BLOCKING]
func (mt *MultiplexTransport) Dial(
	ctx context.Context,
	addr NetAddress,
	behavior PeerBehavior,
) (Peer, error) {
	// Set a dial timeout for the connection
	c, err := addr.DialContext(ctx)
	if err != nil {
		return nil, err
	}

	// Process the connection with expected ID
	info, err := mt.processConn(c, addr.ID)
	if err != nil {
		// Close the net peer connection
		_ = c.Close()

		return nil, err
	}

	return mt.newMultiplexPeer(info, behavior, true)
}

// Close stops the multiplex transport
func (mt *MultiplexTransport) Close() error {
	mt.cancelFn()

	if mt.listener == nil {
		return nil
	}

	return mt.listener.Close()
}

// Listen starts an active process of listening for incoming connections [NON-BLOCKING]
func (mt *MultiplexTransport) Listen(addr NetAddress) error {
	// Reserve a port, and start listening
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

	// Run the routine for accepting
	// incoming peer connections
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
				mt.logger.Error(
					"unable to accept p2p connection",
					"err", err,
				)

				continue
			}

			// Process the new connection asynchronously
			go func(c net.Conn) {
				info, err := mt.processConn(c, "")
				if err != nil {
					mt.logger.Error(
						"unable to process p2p connection",
						"err", err,
					)

					// Close the connection
					_ = c.Close()

					return
				}

				select {
				case mt.peerCh <- info:
				case <-mt.ctx.Done():
					// Give up if the transport was closed.
					_ = c.Close()
				}
			}(c)
		}
	}
}

// processConn handles the raw connection by upgrading it and verifying it
func (mt *MultiplexTransport) processConn(c net.Conn, expectedID ID) (peerInfo, error) {
	dialAddr := c.RemoteAddr().String()

	// Check if the connection is a duplicate one
	if _, exists := mt.activeConns.LoadOrStore(dialAddr, struct{}{}); exists {
		return peerInfo{}, errors.New("duplicate peer connection") // TODO make constant
	}

	// Handshake with the peer, through STS
	secretConn, nodeInfo, err := mt.upgradeAndVerifyConn(c)
	if err != nil {
		mt.activeConns.Delete(dialAddr)

		return peerInfo{}, fmt.Errorf("unable to upgrade connection: %w", err)
	}

	// Verify the connection ID
	id := secretConn.RemotePubKey().Address().ID()

	if !expectedID.IsZero() && id.String() != expectedID.String() {
		mt.activeConns.Delete(dialAddr)

		return peerInfo{}, fmt.Errorf(
			"connection ID does not match dialed ID (expected %q got %q)",
			expectedID,
			id,
		)
	}

	netAddr, _ := NewNetAddress(id, c.RemoteAddr())

	return peerInfo{
		addr:     netAddr,
		conn:     secretConn,
		nodeInfo: nodeInfo,
	}, nil
}

// Remove removes the peer resources from the transport
func (mt *MultiplexTransport) Remove(p Peer) {
	mt.activeConns.Delete(p.RemoteAddr().String())
}

// upgradeAndVerifyConn upgrades the connections (performs the handshaking process)
// and verifies that the connecting peer is valid
func (mt *MultiplexTransport) upgradeAndVerifyConn(c net.Conn) (*conn.SecretConnection, NodeInfo, error) {
	// Upgrade to a secret connection
	secretConn, err := upgradeToSecretConn(
		c,
		mt.handshakeTimeout,
		mt.nodeKey.PrivKey,
	)
	if err != nil {
		return nil, NodeInfo{}, fmt.Errorf("unable to upgrade p2p connection, %w", err)
	}

	// Exchange node information
	nodeInfo, err := exchangeNodeInfo(secretConn, mt.handshakeTimeout, mt.nodeInfo)
	if err != nil {
		return nil, NodeInfo{}, fmt.Errorf("unable to exchange node information, %w", err)
	}

	// Ensure the connection ID matches the node's reported ID
	connID := secretConn.RemotePubKey().Address().ID()

	if connID != nodeInfo.ID() {
		return nil, NodeInfo{}, fmt.Errorf(
			"connection ID does not match node info ID (expected %q got %q)",
			connID.String(),
			nodeInfo.ID().String(),
		)
	}

	// Check compatibility with the node
	if err = mt.nodeInfo.CompatibleWith(nodeInfo); err != nil {
		return nil, NodeInfo{}, fmt.Errorf("incompatible node info, %w", err)
	}

	return secretConn, nodeInfo, nil
}

// newMultiplexPeer creates a new multiplex Peer, using
// the provided Peer behavior and info
func (mt *MultiplexTransport) newMultiplexPeer(
	info peerInfo,
	behavior PeerBehavior,
	isOutbound bool,
) (Peer, error) {
	// Check for peer persistence using the dial address,
	// as well as the self-reported address
	persistent := behavior.IsPersistentPeer(info.addr) ||
		behavior.IsPersistentPeer(info.nodeInfo.NetAddress)

	// Extract the host
	host, _, err := net.SplitHostPort(info.conn.RemoteAddr().String())
	if err != nil {
		return nil, fmt.Errorf("unable to extract peer host, %w", err)
	}

	// Look up the IPs
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, fmt.Errorf("unable to lookup peer IPs, %w", err)
	}

	// Wrap the info related to the connection
	peerConn := &ConnInfo{
		Outbound:   isOutbound,
		Persistent: persistent,
		Conn:       info.conn,
		RemoteIP:   ips[0], // IPv4
		SocketAddr: info.addr,
	}

	// Create the info related to the multiplex connection
	mConfig := &MultiplexConnConfig{
		MConfig:      mt.mConfig,
		ReactorsByCh: behavior.Reactors(),
		ChDescs:      behavior.ReactorChDescriptors(),
		OnPeerError:  behavior.HandlePeerError,
	}

	return NewPeer(peerConn, info.nodeInfo, mConfig), nil
}

// exchangeNodeInfo performs a "handshake", where node
// info is exchanged between the current node and a peer
func exchangeNodeInfo(
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

	// Validate the received node information
	if err := nodeInfo.Validate(); err != nil {
		return NodeInfo{}, fmt.Errorf("unable to validate node info, %w", err)
	}

	return peerNodeInfo, nil
}

// upgradeToSecretConn takes an active TCP connection,
// and upgrades it to a verified, handshaked connection through
// the STS protocol
func upgradeToSecretConn(
	c net.Conn,
	timeout time.Duration,
	privKey crypto.PrivKey,
) (*conn.SecretConnection, error) {
	if err := c.SetDeadline(time.Now().Add(timeout)); err != nil {
		return nil, err
	}

	// Handshake (STS)
	sc, err := conn.MakeSecretConnection(c, privKey)
	if err != nil {
		return nil, err
	}

	return sc, sc.SetDeadline(time.Time{})
}
