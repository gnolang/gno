package p2p

import (
	"context"
	goerrors "errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/p2p/conn"
	"github.com/gnolang/gno/tm2/pkg/p2p/types"
	"golang.org/x/sync/errgroup"
)

// defaultHandshakeTimeout is the timeout for the STS handshaking protocol
const defaultHandshakeTimeout = 3 * time.Second

var (
	errTransportClosed        = errors.New("transport is closed")
	errDuplicateConnection    = errors.New("duplicate peer connection")
	errPeerIDNodeInfoMismatch = errors.New("connection ID does not match node info ID")
	errPeerIDDialMismatch     = errors.New("connection ID does not match dialed ID")
	errIncompatibleNodeInfo   = errors.New("incompatible node info")
)

type connUpgradeFn func(io.ReadWriteCloser, ed25519.PrivKeyEd25519) (*conn.SecretConnection, error)

type secretConn interface {
	net.Conn

	RemotePubKey() ed25519.PubKeyEd25519
}

// peerInfo is a wrapper for an unverified peer connection
type peerInfo struct {
	addr     *types.NetAddress // the dial address of the peer
	conn     net.Conn          // the connection associated with the peer
	nodeInfo types.NodeInfo    // the relevant peer node info
}

// MultiplexTransport accepts and dials tcp connections and upgrades them to
// multiplexed peers.
type MultiplexTransport struct {
	ctx      context.Context
	cancelFn context.CancelFunc

	logger *slog.Logger

	netAddr  types.NetAddress // the node's P2P dial address, used for handshaking
	nodeInfo types.NodeInfo   // the node's P2P info, used for handshaking
	nodeKey  types.NodeKey    // the node's private P2P key, used for handshaking

	listener    net.Listener  // listener for inbound peer connections
	peerCh      chan peerInfo // pipe for inbound peer connections
	activeConns sync.Map      // active peer connections (remote address -> nothing)

	connUpgradeFn connUpgradeFn // Upgrades the connection to a secret connection

	// TODO(xla): This config is still needed as we parameterize peerConn and
	// peer currently. All relevant configuration should be refactored into options
	// with sane defaults.
	mConfig conn.MConnConfig
}

// NewMultiplexTransport returns a tcp connected multiplexed peer.
func NewMultiplexTransport(
	nodeInfo types.NodeInfo,
	nodeKey types.NodeKey,
	mConfig conn.MConnConfig,
	logger *slog.Logger,
) *MultiplexTransport {
	ctx, cancel := context.WithCancel(context.Background())
	return &MultiplexTransport{
		ctx:           ctx,
		cancelFn:      cancel,
		peerCh:        make(chan peerInfo, 1),
		mConfig:       mConfig,
		nodeInfo:      nodeInfo,
		nodeKey:       nodeKey,
		logger:        logger,
		connUpgradeFn: conn.MakeSecretConnection,
	}
}

// NetAddress returns the transport's listen address (for p2p connections)
func (mt *MultiplexTransport) NetAddress() types.NetAddress {
	return mt.netAddr
}

// Accept waits for a verified inbound Peer to connect, and returns it [BLOCKING]
func (mt *MultiplexTransport) Accept(ctx context.Context, behavior PeerBehavior) (PeerConn, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case info, ok := <-mt.peerCh:
		if !ok {
			return nil, errTransportClosed
		}

		return mt.newMultiplexPeer(info, behavior, false)
	}
}

// Dial creates an outbound Peer connection, and
// verifies it (performs handshaking) [BLOCKING]
func (mt *MultiplexTransport) Dial(
	ctx context.Context,
	addr types.NetAddress,
	behavior PeerBehavior,
) (PeerConn, error) {
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

		return nil, fmt.Errorf("unable to process connection, %w", err)
	}

	return mt.newMultiplexPeer(info, behavior, true)
}

// Close stops the multiplex transport
func (mt *MultiplexTransport) Close() error {
	if mt.listener == nil {
		return nil
	}

	mt.cancelFn()

	return mt.listener.Close()
}

// Listen starts an active process of listening for incoming connections [NON-BLOCKING]
func (mt *MultiplexTransport) Listen(addr types.NetAddress) error {
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
			ln.Close()
			return fmt.Errorf("error finding port (after listening on port 0): %w", err)
		}

		addr.Port = uint16(tcpAddr.Port)
	}

	mt.netAddr = addr
	mt.listener = ln

	// Run the routine for accepting
	// incoming peer connections
	go mt.runAcceptLoop(mt.ctx)

	return nil
}

// runAcceptLoop runs the loop where incoming peers are:
//
// 1. accepted by the transport
// 2. filtered
// 3. upgraded (handshaked + verified)
func (mt *MultiplexTransport) runAcceptLoop(ctx context.Context) {
	var wg sync.WaitGroup
	defer func() {
		wg.Wait() // Wait for all process routines
		close(mt.peerCh)
	}()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel() // cancel sub-connection process

	for {
		// Accept an incoming peer connection
		c, err := mt.listener.Accept()

		switch {
		case err == nil: // ok
		case goerrors.Is(err, net.ErrClosed):
			// Listener has been closed, this is not recoverable.
			mt.logger.Debug("listener has been closed")
			return // exit
		default:
			// An error occurred during accept, report and continue
			mt.logger.Warn("accept p2p connection error", "err", err)
			continue
		}

		// Process the new connection asynchronously
		wg.Add(1)

		go func(c net.Conn) {
			defer wg.Done()

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
			case <-ctx.Done():
				// Give up if the transport was closed.
				_ = c.Close()
			}
		}(c)
	}
}

// processConn handles the raw connection by upgrading it and verifying it
func (mt *MultiplexTransport) processConn(c net.Conn, expectedID types.ID) (peerInfo, error) {
	dialAddr := c.RemoteAddr().String()

	// Check if the connection is a duplicate one
	if _, exists := mt.activeConns.LoadOrStore(dialAddr, struct{}{}); exists {
		return peerInfo{}, errDuplicateConnection
	}

	// Handshake with the peer, through STS
	secretConn, nodeInfo, err := mt.upgradeAndVerifyConn(c)
	if err != nil {
		mt.activeConns.Delete(dialAddr)

		return peerInfo{}, fmt.Errorf("unable to upgrade connection, %w", err)
	}

	// Grab the connection ID.
	// At this point, the connection and information shared
	// with the peer is considered valid, since full handshaking
	// and verification took place
	id := secretConn.RemotePubKey().Address().ID()

	// The reason the dial ID needs to be verified is because
	// for outbound peers (peers the node dials), there is an expected peer ID
	// when initializing the outbound connection, that can differ from the exchanged one.
	// For inbound peers, the ID is whatever the peer exchanges during the
	// handshaking process, and is verified separately
	if !expectedID.IsZero() && id.String() != expectedID.String() {
		mt.activeConns.Delete(dialAddr)

		return peerInfo{}, fmt.Errorf(
			"%w (expected %q got %q)",
			errPeerIDDialMismatch,
			expectedID,
			id,
		)
	}

	netAddr, _ := types.NewNetAddress(id, c.RemoteAddr())

	return peerInfo{
		addr:     netAddr,
		conn:     secretConn,
		nodeInfo: nodeInfo,
	}, nil
}

// Remove removes the peer resources from the transport
func (mt *MultiplexTransport) Remove(p PeerConn) {
	mt.activeConns.Delete(p.RemoteAddr().String())
}

// upgradeAndVerifyConn upgrades the connections (performs the handshaking process)
// and verifies that the connecting peer is valid
func (mt *MultiplexTransport) upgradeAndVerifyConn(c net.Conn) (secretConn, types.NodeInfo, error) {
	// Upgrade to a secret connection.
	// A secret connection is a connection that has passed
	// an initial handshaking process, as defined by the STS
	// protocol, and is considered to be secure and authentic
	sc, err := mt.upgradeToSecretConn(
		c,
		defaultHandshakeTimeout,
		mt.nodeKey.PrivKey,
	)
	if err != nil {
		return nil, types.NodeInfo{}, fmt.Errorf("unable to upgrade p2p connection, %w", err)
	}

	// Exchange node information
	nodeInfo, err := exchangeNodeInfo(sc, defaultHandshakeTimeout, mt.nodeInfo)
	if err != nil {
		return nil, types.NodeInfo{}, fmt.Errorf("unable to exchange node information, %w", err)
	}

	// Ensure the connection ID matches the node's reported ID
	connID := sc.RemotePubKey().Address().ID()

	if connID != nodeInfo.ID() {
		return nil, types.NodeInfo{}, fmt.Errorf(
			"%w (expected %q got %q)",
			errPeerIDNodeInfoMismatch,
			connID.String(),
			nodeInfo.ID().String(),
		)
	}

	// Check compatibility with the node
	if err = mt.nodeInfo.CompatibleWith(nodeInfo); err != nil {
		return nil, types.NodeInfo{}, fmt.Errorf("%w, %w", errIncompatibleNodeInfo, err)
	}

	return sc, nodeInfo, nil
}

// newMultiplexPeer creates a new multiplex Peer, using
// the provided Peer behavior and info
func (mt *MultiplexTransport) newMultiplexPeer(
	info peerInfo,
	behavior PeerBehavior,
	isOutbound bool,
) (PeerConn, error) {
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
		Persistent: behavior.IsPersistentPeer(info.addr.ID),
		Private:    behavior.IsPrivatePeer(info.nodeInfo.ID()),
		Conn:       info.conn,
		RemoteIP:   ips[0], // IPv4
		SocketAddr: info.addr,
	}

	// Create the info related to the multiplex connection
	mConfig := &ConnConfig{
		MConfig:      mt.mConfig,
		ReactorsByCh: behavior.Reactors(),
		ChDescs:      behavior.ReactorChDescriptors(),
		OnPeerError:  behavior.HandlePeerError,
	}

	return newPeer(peerConn, info.nodeInfo, mConfig), nil
}

// exchangeNodeInfo performs a data swap, where node
// info is exchanged between the current node and a peer async
func exchangeNodeInfo(
	c secretConn,
	timeout time.Duration,
	nodeInfo types.NodeInfo,
) (types.NodeInfo, error) {
	if err := c.SetDeadline(time.Now().Add(timeout)); err != nil {
		return types.NodeInfo{}, err
	}

	var (
		peerNodeInfo types.NodeInfo
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
			types.MaxNodeInfoSize,
		)

		return err
	})

	if err := g.Wait(); err != nil {
		return types.NodeInfo{}, err
	}

	// Validate the received node information
	if err := nodeInfo.Validate(); err != nil {
		return types.NodeInfo{}, fmt.Errorf("unable to validate node info, %w", err)
	}

	return peerNodeInfo, c.SetDeadline(time.Time{})
}

// upgradeToSecretConn takes an active TCP connection,
// and upgrades it to a verified, handshaked connection through
// the STS protocol
func (mt *MultiplexTransport) upgradeToSecretConn(
	c net.Conn,
	timeout time.Duration,
	privKey ed25519.PrivKeyEd25519,
) (secretConn, error) {
	if err := c.SetDeadline(time.Now().Add(timeout)); err != nil {
		return nil, err
	}

	// Handshake (STS)
	sc, err := mt.connUpgradeFn(c, privKey)
	if err != nil {
		return nil, err
	}

	return sc, sc.SetDeadline(time.Time{})
}
