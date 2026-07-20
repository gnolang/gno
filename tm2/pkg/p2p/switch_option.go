package p2p

import (
	"github.com/gnolang/gno/tm2/pkg/p2p/types"
)

// SwitchOption is a callback used for configuring the p2p MultiplexSwitch
type SwitchOption func(*MultiplexSwitch)

// WithReactor sets the p2p switch reactors
func WithReactor(name string, reactor Reactor) SwitchOption {
	return func(sw *MultiplexSwitch) {
		for _, chDesc := range reactor.GetChannels() {
			chID := chDesc.ID

			// No two reactors can share the same channel
			if sw.peerBehavior.reactorsByCh[chID] != nil {
				continue
			}

			sw.peerBehavior.chDescs = append(sw.peerBehavior.chDescs, chDesc)
			sw.peerBehavior.reactorsByCh[chID] = reactor
		}

		sw.reactors[name] = reactor

		reactor.SetSwitch(sw)
	}
}

// WithPersistentPeers sets the p2p switch's persistent peer set from raw
// "id@host:port" address strings. The original strings are retained
// alongside the resolved addresses so FQDN-based persistent peers can be
// re-resolved on each reconnect attempt (see redialFn in switch.go) instead
// of reusing a possibly-stale resolved IP forever (see #2580). Invalid
// addresses are skipped; callers that want to surface parse errors (e.g.
// for logging) should validate persistentPeerAddrs themselves beforehand,
// such as with types.NewNetAddressFromStrings.
func WithPersistentPeers(persistentPeerAddrs []string) SwitchOption {
	return func(sw *MultiplexSwitch) {
		for _, raw := range persistentPeerAddrs {
			addr, err := types.NewNetAddressFromString(raw)
			if err != nil {
				continue
			}

			sw.persistentPeers.Store(addr.ID, addr)
			sw.persistentPeerAddrStrs.Store(addr.ID, raw)
		}
	}
}

// WithPrivatePeers sets the p2p switch's private peer set
func WithPrivatePeers(peerIDs []types.ID) SwitchOption {
	return func(sw *MultiplexSwitch) {
		for _, id := range peerIDs {
			sw.privatePeers.Store(id, struct{}{})
		}
	}
}

// WithMaxInboundPeers sets the p2p switch's maximum inbound peer limit
func WithMaxInboundPeers(maxInbound uint64) SwitchOption {
	return func(sw *MultiplexSwitch) {
		sw.maxInboundPeers = maxInbound
	}
}

// WithMaxOutboundPeers sets the p2p switch's maximum outbound peer limit
func WithMaxOutboundPeers(maxOutbound uint64) SwitchOption {
	return func(sw *MultiplexSwitch) {
		sw.maxOutboundPeers = maxOutbound
	}
}
