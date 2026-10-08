package p2p

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/p2p/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// countingTransport counts the connections a transport dialed and accepted
// through a completed handshake
type countingTransport struct {
	*MultiplexTransport

	dialed, accepted atomic.Int32
}

func (c *countingTransport) Dial(ctx context.Context, addr types.NetAddress, behavior PeerBehavior) (PeerConn, error) {
	p, err := c.MultiplexTransport.Dial(ctx, addr, behavior)
	if err == nil {
		c.dialed.Add(1)
	}

	return p, err
}

func (c *countingTransport) Accept(ctx context.Context, behavior PeerBehavior) (PeerConn, error) {
	p, err := c.MultiplexTransport.Accept(ctx, behavior)
	if err == nil {
		c.accepted.Add(1)
	}

	return p, err
}

// newLoopbackSwitch starts a switch on a real transport listening on the
// loopback address. Every node shares that address, so the duplicate-IP guard
// is lifted
func newLoopbackSwitch(t *testing.T, moniker string) (*MultiplexSwitch, *countingTransport) {
	t.Helper()

	tr := &countingTransport{MultiplexTransport: newLoopbackTransport(t, "dev", moniker)}

	sw := NewMultiplexSwitch(tr, WithAllowDuplicateIP(true))
	sw.SetLogger(log.NewNoopLogger())

	require.NoError(t, sw.Start())

	t.Cleanup(func() {
		if err := sw.Stop(); err != nil {
			t.Logf("unable to stop switch %s: %v", moniker, err)
		}
	})

	return sw, tr
}

// connectedAsTieBreakKeeps reports whether the two switches are connected to
// each other through exactly one connection, the one dialed by the node with
// the lower ID, on both sides
func connectedAsTieBreakKeeps(a, b *MultiplexSwitch) bool {
	lower, upper := a, b

	if b.transport.NetAddress().ID < a.transport.NetAddress().ID {
		lower, upper = b, a
	}

	var (
		toUpper = lower.Peers().Get(upper.transport.NetAddress().ID)
		toLower = upper.Peers().Get(lower.transport.NetAddress().ID)
	)

	return toUpper != nil && toUpper.IsOutbound() &&
		toLower != nil && !toLower.IsOutbound() &&
		len(lower.Peers().List()) == 1 &&
		len(upper.Peers().List()) == 1
}

// describePeers renders the switches' peer sets, for failure messages
func describePeers(switches ...*MultiplexSwitch) string {
	var b strings.Builder

	for _, sw := range switches {
		fmt.Fprintf(&b, "[%s:", sw.transport.NetAddress().ID)

		for _, p := range sw.Peers().List() {
			fmt.Fprintf(&b, " %s %s", p.ID(), direction(p))
		}

		b.WriteString("] ")
	}

	return b.String()
}

// TestMultiplexSwitch_SimultaneousOpen is the regression test for
// gnolang/gno#6302: two nodes dial each other at the same moment, over real
// connections. Keeping whichever connection registered first lets each side
// keep a different one and close the other's, tearing both down; both sides
// must instead keep the one dialed by the lower ID. The nodes dial through
// dialPeer directly, past the dial loop's check for an already connected
// peer, so both connections are attempted on every run. Converging proves
// nothing unless the connections overlapped, so at least one attempt must have
// completed the handshake of both
func TestMultiplexSwitch_SimultaneousOpen(t *testing.T) {
	t.Parallel()

	// Attempts in which both connections completed their handshake, so the
	// switches had two connections to resolve
	var overlapped atomic.Int32

	// The group returns once every parallel attempt is done
	t.Run("attempts", func(t *testing.T) {
		for i := range 30 {
			t.Run(fmt.Sprintf("attempt %d", i), func(t *testing.T) {
				t.Parallel()

				var (
					a, aTr = newLoopbackSwitch(t, "a")
					b, bTr = newLoopbackSwitch(t, "b")

					aAddr = a.transport.NetAddress()
					bAddr = b.transport.NetAddress()

					ctx   = t.Context()
					start = make(chan struct{})
					wg    sync.WaitGroup
				)

				wg.Go(func() {
					<-start
					a.dialPeer(ctx, &bAddr)
				})
				wg.Go(func() {
					<-start
					b.dialPeer(ctx, &aAddr)
				})

				close(start)
				wg.Wait()

				if !assert.Eventually(t, func() bool {
					return connectedAsTieBreakKeeps(a, b)
				}, 5*time.Second, 10*time.Millisecond) {
					t.Fatalf("not connected through the lower ID's connection: %s", describePeers(a, b))
				}

				// The connection the tie-break keeps survives once the other is gone
				time.Sleep(time.Second)

				if !connectedAsTieBreakKeeps(a, b) {
					t.Fatalf("the kept connection did not survive: %s", describePeers(a, b))
				}

				if aTr.dialed.Load() > 0 && aTr.accepted.Load() > 0 &&
					bTr.dialed.Load() > 0 && bTr.accepted.Load() > 0 {
					overlapped.Add(1)
				}
			})
		}
	})

	// Converging proves nothing if the dials never overlapped
	require.Positive(t, overlapped.Load(), "no attempt opened both connections at once")
}
