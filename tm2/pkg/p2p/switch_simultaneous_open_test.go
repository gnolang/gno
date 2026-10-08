package p2p

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/stretchr/testify/assert"
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
// connections, and both must keep the connection dialed by the lower ID. Both
// nodes dial through dialPeer directly, past the dial loop's check for an
// already connected peer, so both connections are attempted on every run.
// Which one each side registers first depends on scheduling, and before the
// fix about a third of the attempts ended without the lower ID's connection on
// both sides
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
		})
	}
}
