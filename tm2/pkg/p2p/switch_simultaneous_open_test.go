package p2p

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/stretchr/testify/require"
)

// newLoopbackSwitch starts a switch on a real transport listening on the
// loopback address. Every node shares that address, so the duplicate-IP guard
// is lifted
func newLoopbackSwitch(t *testing.T, moniker string) *MultiplexSwitch {
	t.Helper()

	sw := NewMultiplexSwitch(
		newLoopbackTransport(t, "dev", moniker),
		WithAllowDuplicateIP(true),
	)
	sw.SetLogger(log.NewNoopLogger())

	require.NoError(t, sw.Start())

	t.Cleanup(func() {
		if err := sw.Stop(); err != nil {
			t.Logf("unable to stop switch %s: %v", moniker, err)
		}
	})

	return sw
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

// TestMultiplexSwitch_SimultaneousOpen is the regression test for
// gnolang/gno#6302: two nodes dial each other at the same moment, over real
// connections. Keeping whichever connection registered first lets each side
// keep a different one and close the other's, tearing both down; both sides
// must instead keep the one dialed by the lower ID. The nodes dial through
// dialPeer directly, past the dial loop's check for an already connected
// peer, so both connections are attempted on every run
func TestMultiplexSwitch_SimultaneousOpen(t *testing.T) {
	t.Parallel()

	for i := range 30 {
		t.Run(fmt.Sprintf("attempt %d", i), func(t *testing.T) {
			t.Parallel()

			var (
				a = newLoopbackSwitch(t, "a")
				b = newLoopbackSwitch(t, "b")

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

			require.Eventually(t, func() bool {
				return connectedAsTieBreakKeeps(a, b)
			}, 5*time.Second, 10*time.Millisecond)

			// The connection the tie-break keeps survives once the other is gone
			time.Sleep(time.Second)

			require.True(t, connectedAsTieBreakKeeps(a, b))
		})
	}
}
