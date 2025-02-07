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

// WithPersistentPeers sets the p2p switch's persistent peer set
func WithPersistentPeers(peerAddrs []*types.NetAddress) SwitchOption {
	return func(sw *MultiplexSwitch) {
		for _, addr := range peerAddrs {
			sw.persistentPeers.Store(addr.ID, addr)
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
