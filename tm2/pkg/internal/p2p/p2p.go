// Package p2p contains testing code that is moved over, and adapted from p2p/test_utils.go.
// This isn't a good way to simulate the networking layer in TM2 modules.
// It actually isn't a good way to simulate the networking layer, in anything.
//
// Code is carried over to keep the testing code of p2p-dependent modules happy
// and "working". We should delete this entire package the second TM2 module unit tests don't
// need to rely on a live p2p cluster to pass.
package p2p

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/p2p"
	p2pcfg "github.com/gnolang/gno/tm2/pkg/p2p/config"
	"github.com/gnolang/gno/tm2/pkg/p2p/conn"
	"github.com/gnolang/gno/tm2/pkg/p2p/events"
	p2pTypes "github.com/gnolang/gno/tm2/pkg/p2p/types"
	"github.com/gnolang/gno/tm2/pkg/service"
	"github.com/gnolang/gno/tm2/pkg/versionset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

// TestingConfig is the P2P cluster testing config
type TestingConfig struct {
	P2PCfg        *p2pcfg.P2PConfig          // the common p2p configuration
	Count         int                        // the size of the cluster
	SwitchOptions map[int][]p2p.SwitchOption // multiplex switch options
	Channels      []byte                     // the common p2p peer multiplex channels
}

// MakeConnectedPeers creates a cluster of peers, with the given options.
// Used to simulate the networking layer for a TM2 module
func MakeConnectedPeers(
	t *testing.T,
	ctx context.Context,
	cfg TestingConfig,
) ([]*p2p.MultiplexSwitch, []*p2p.MultiplexTransport) {
	t.Helper()

	// Each pair of switches is connected from the lower-indexed side only (see
	// the DialPeers call below), so switch 0 makes cfg.Count-1 outbound dials and
	// accepts no inbound ones. MultiplexSwitch checks MaxNumOutboundPeers (default
	// 10) only when a dial is *enqueued*, and we enqueue every peer before any
	// dial completes, so the cap currently isn't hit even above 10 — but that
	// relies on enqueueing racing ahead of connection setup. Rather than depend
	// on that, fail fast with a clear message if a caller ever exceeds the cap,
	// instead of surfacing it as a confusing 1-minute connect timeout.
	maxOutbound := int(p2pcfg.DefaultP2PConfig().MaxNumOutboundPeers)
	require.LessOrEqualf(t, cfg.Count-1, maxOutbound,
		"cluster of %d requires %d outbound dials from the lowest switch, "+
			"exceeding the outbound peer cap of %d",
		cfg.Count, cfg.Count-1, maxOutbound)

	// Initialize collections for switches, transports, and addresses.
	var (
		sws   = make([]*p2p.MultiplexSwitch, cfg.Count)
		ts    = make([]*p2p.MultiplexTransport, 0, cfg.Count)
		addrs = make([]*p2pTypes.NetAddress, 0, cfg.Count)
	)

	createTransport := func(index int) *p2p.MultiplexTransport {
		// Generate a fresh key
		key := p2pTypes.GenerateNodeKey()

		addr, err := p2pTypes.NewNetAddress(
			key.ID(),
			&net.TCPAddr{
				IP:   net.ParseIP("127.0.0.1"),
				Port: 0, // random free port
			},
		)
		require.NoError(t, err)

		info := p2pTypes.NodeInfo{
			VersionSet: versionset.VersionSet{
				versionset.VersionInfo{Name: "p2p", Version: "v0.0.0"},
			},
			NetAddress: addr,
			Network:    "testing",
			Software:   "p2ptest",
			Version:    "v1.2.3-rc.0-deadbeef",
			Channels:   cfg.Channels,
			Moniker:    fmt.Sprintf("node-%d", index),
			Other: p2pTypes.NodeInfoOther{
				TxIndex:    "off",
				RPCAddress: fmt.Sprintf("127.0.0.1:%d", 0),
			},
		}

		transport := p2p.NewMultiplexTransport(
			info,
			*key,
			conn.MConfigFromP2P(cfg.P2PCfg),
			log.NewNoopLogger(),
		)

		require.NoError(t, transport.Listen(*addr))
		t.Cleanup(func() { assert.NoError(t, transport.Close()) })

		return transport
	}

	// Create transports and gather addresses
	for i := range cfg.Count {
		transport := createTransport(i)
		addr := transport.NetAddress()

		addrs = append(addrs, &addr)
		ts = append(ts, transport)
	}

	// Connect switches and ensure all peers are connected
	connectPeers := func(switchIndex int) error {
		multiplexSwitch := p2p.NewMultiplexSwitch(
			ts[switchIndex],
			cfg.SwitchOptions[switchIndex]...,
		)

		ch, unsubFn := multiplexSwitch.Subscribe(func(event events.Event) bool {
			return event.Type() == events.PeerConnected
		})
		defer unsubFn()

		// Start the switch
		require.NoError(t, multiplexSwitch.Start())

		// Save it
		sws[switchIndex] = multiplexSwitch

		if cfg.Count == 1 {
			// No peers to dial, switch is alone
			return nil
		}

		// Async dial the peers with a higher index, so each pair connects in
		// a single direction. If both ends of a pair dial each other
		// concurrently, each side can register its own outbound connection and
		// then reject the other side's inbound as a duplicate, tearing down
		// both connections and leaving the pair permanently disconnected
		multiplexSwitch.DialPeers(addrs[switchIndex+1:]...)

		// Set up an exit timer
		timer := time.NewTimer(1 * time.Minute)
		defer timer.Stop()

		var (
			connectedPeers = make(map[p2pTypes.ID]struct{})
			targetPeers    = cfg.Count - 1
		)

		for {
			select {
			case evRaw := <-ch:
				ev := evRaw.(events.PeerConnectedEvent)

				connectedPeers[ev.PeerID] = struct{}{}

				if len(connectedPeers) == targetPeers {
					return nil
				}
			case <-timer.C:
				return errors.New("timed out waiting for peer switches to connect")
			}
		}
	}

	g, _ := errgroup.WithContext(ctx)
	for i := range cfg.Count {
		g.Go(func() error { return connectPeers(i) })
	}

	require.NoError(t, g.Wait())

	return sws, ts
}

// createRoutableAddr generates a valid, routable NetAddress for the given node ID using a secure random IP
func createRoutableAddr(t *testing.T, id p2pTypes.ID) *p2pTypes.NetAddress {
	t.Helper()

	generateIP := func() string {
		ip := make([]byte, 4)

		_, err := rand.Read(ip)
		require.NoError(t, err)

		return fmt.Sprintf("%d.%d.%d.%d", ip[0], ip[1], ip[2], ip[3])
	}

	for {
		addrStr := fmt.Sprintf("%s@%s:26656", id, generateIP())

		netAddr, err := p2pTypes.NewNetAddressFromString(addrStr)
		require.NoError(t, err)

		if netAddr.Routable() {
			return netAddr
		}
	}
}

// Peer is a live peer, utilized for testing purposes.
// This Peer implementation is NOT thread safe
type Peer struct {
	*service.BaseService
	ip   net.IP
	id   p2pTypes.ID
	addr *p2pTypes.NetAddress
	kv   map[string]any

	Outbound, Persistent, Private bool
}

// NewPeer creates and starts a new mock peer.
// It generates a new routable address for the peer
func NewPeer(t *testing.T) *Peer {
	t.Helper()

	var (
		nodeKey = p2pTypes.GenerateNodeKey()
		netAddr = createRoutableAddr(t, nodeKey.ID())
	)

	mp := &Peer{
		ip:   netAddr.IP,
		id:   nodeKey.ID(),
		addr: netAddr,
		kv:   make(map[string]any),
	}

	mp.BaseService = service.NewBaseService(nil, "MockPeer", mp)

	require.NoError(t, mp.Start())

	return mp
}

func (mp *Peer) FlushStop()                    { mp.Stop() }
func (mp *Peer) TrySend(_ byte, _ []byte) bool { return true }
func (mp *Peer) Send(_ byte, _ []byte) bool    { return true }
func (mp *Peer) NodeInfo() p2pTypes.NodeInfo {
	return p2pTypes.NodeInfo{
		NetAddress: mp.addr,
	}
}
func (mp *Peer) Status() conn.ConnectionStatus { return conn.ConnectionStatus{} }
func (mp *Peer) ID() p2pTypes.ID               { return mp.id }
func (mp *Peer) IsOutbound() bool              { return mp.Outbound }
func (mp *Peer) IsPersistent() bool            { return mp.Persistent }
func (mp *Peer) IsPrivate() bool               { return mp.Private }
func (mp *Peer) Get(key string) any {
	if value, ok := mp.kv[key]; ok {
		return value
	}
	return nil
}

func (mp *Peer) Set(key string, value any) {
	mp.kv[key] = value
}
func (mp *Peer) RemoteIP() net.IP                 { return mp.ip }
func (mp *Peer) SocketAddr() *p2pTypes.NetAddress { return mp.addr }
func (mp *Peer) RemoteAddr() net.Addr             { return &net.TCPAddr{IP: mp.ip, Port: 8800} }
func (mp *Peer) CloseConn() error                 { return nil }
