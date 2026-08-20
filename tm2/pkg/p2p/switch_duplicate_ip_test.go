package p2p

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/p2p/conn"
	"github.com/gnolang/gno/tm2/pkg/p2p/types"
	"github.com/gnolang/gno/tm2/pkg/versionset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newLoopbackTransport spins up a real MultiplexTransport listening on the
// loopback address.
func newLoopbackTransport(t *testing.T, network, moniker string) *MultiplexTransport {
	t.Helper()

	key := types.GenerateNodeKey()

	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	na, err := types.NewNetAddress(key.ID(), addr)
	require.NoError(t, err)

	info := types.NodeInfo{
		Network:    network,
		NetAddress: na,
		Version:    "v1.0.0-rc.0",
		Moniker:    moniker,
		VersionSet: make(versionset.VersionSet, 0),
		Channels:   []byte{42},
	}

	tr := NewMultiplexTransport(info, *key, conn.DefaultMConnConfig(), log.NewNoopLogger())
	require.NoError(t, tr.Listen(*na))

	t.Cleanup(func() { _ = tr.Close() })

	return tr
}

// dialInbound opens numDialers inbound connections to the given transport, each
// from a distinct node key but all from the loopback address, and returns how
// many the switch ended up accepting.
func dialInbound(t *testing.T, target *MultiplexTransport, sw *MultiplexSwitch, numDialers int) uint64 {
	t.Helper()

	behavior := &reactorPeerBehavior{
		chDescs:            make([]*conn.ChannelDescriptor, 0),
		reactorsByCh:       make(map[byte]Reactor),
		handlePeerErrFn:    func(PeerConn, error) {},
		isPersistentPeerFn: func(types.ID) bool { return false },
		isPrivatePeerFn:    func(types.ID) bool { return false },
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for i := range numDialers {
		clientTr := newLoopbackTransport(t, "dev", fmt.Sprintf("dialer-%d", i))

		p, err := clientTr.Dial(ctx, target.netAddr, behavior)
		require.NoError(t, err, "dial %d", i)
		require.NotNil(t, p)

		t.Cleanup(func() { _ = p.Stop() })
	}

	// Give the accept loop time to settle. Poll for growth, then wait a beat
	// past the last change so rejections are observed too.
	var last uint64
	stable := 0
	for range 60 {
		time.Sleep(50 * time.Millisecond)

		if in := sw.Peers().NumInbound(); in != last {
			last, stable = in, 0

			continue
		}

		if stable++; stable > 4 {
			break
		}
	}

	return sw.Peers().NumInbound()
}

// TestSwitchRejectsDuplicateIP verifies that, by default, a single remote IP
// cannot occupy more than one inbound peer slot. Peer IDs are self-generated
// node keys, so without this guard one host can mint fresh identities and fill
// every slot -- which is what bounds the per-connection recv-buffer budget to a
// single connection's worth of memory rather than the whole inbound limit.
func TestSwitchRejectsDuplicateIP(t *testing.T) {
	t.Parallel()

	serverTr := newLoopbackTransport(t, "dev", "server")

	sw := NewMultiplexSwitch(serverTr, WithMaxInboundPeers(40))
	sw.SetLogger(log.NewNoopLogger())
	require.NoError(t, sw.Start())

	t.Cleanup(func() { _ = sw.Stop() })

	assert.EqualValues(t, 1, dialInbound(t, serverTr, sw, 8),
		"only one inbound slot may be held per source IP")
}

// TestSwitchAllowsDuplicateIPWhenEnabled is the counterpart: local clusters run
// every node on the loopback address, so they must be able to opt out.
func TestSwitchAllowsDuplicateIPWhenEnabled(t *testing.T) {
	t.Parallel()

	serverTr := newLoopbackTransport(t, "dev", "server")

	sw := NewMultiplexSwitch(
		serverTr,
		WithMaxInboundPeers(40),
		WithAllowDuplicateIP(true),
	)
	sw.SetLogger(log.NewNoopLogger())
	require.NoError(t, sw.Start())

	t.Cleanup(func() { _ = sw.Stop() })

	assert.EqualValues(t, 8, dialInbound(t, serverTr, sw, 8),
		"all inbound connections should be accepted when duplicate IPs are allowed")
}

// TestSwitchClosesRejectedDuplicateIPConn verifies that a connection the guard
// turns away is actually closed. transport.Remove only forgets the connection;
// without an explicit close the socket the handshake established lingers until
// the netFD finalizer runs, so a single host could accumulate handshaked
// sockets on the victim faster than the GC reclaims them.
//
// Observed from the dialer's side: once the switch has rejected us, our own
// connection must see the far end go away.
func TestSwitchClosesRejectedDuplicateIPConn(t *testing.T) {
	t.Parallel()

	serverTr := newLoopbackTransport(t, "dev", "server")

	sw := NewMultiplexSwitch(serverTr, WithMaxInboundPeers(40))
	sw.SetLogger(log.NewNoopLogger())
	require.NoError(t, sw.Start())

	t.Cleanup(func() { _ = sw.Stop() })

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dial := func(moniker string, errCh chan<- error) PeerConn {
		t.Helper()

		clientTr := newLoopbackTransport(t, "dev", moniker)

		p, err := clientTr.Dial(ctx, serverTr.netAddr, &reactorPeerBehavior{
			chDescs:      make([]*conn.ChannelDescriptor, 0),
			reactorsByCh: make(map[byte]Reactor),
			handlePeerErrFn: func(_ PeerConn, err error) {
				select {
				case errCh <- err:
				default:
				}
			},
			isPersistentPeerFn: func(types.ID) bool { return false },
			isPrivatePeerFn:    func(types.ID) bool { return false },
		})
		require.NoError(t, err)
		require.NoError(t, p.Start())

		t.Cleanup(func() { _ = p.Stop() })

		return p
	}

	// The first connection takes the only slot this IP gets.
	dial("first", make(chan error, 1))

	// The second is rejected. Our end must notice.
	rejected := make(chan error, 1)
	dial("second", rejected)

	select {
	case err := <-rejected:
		require.Error(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("the rejected connection was left open by the switch")
	}
}
