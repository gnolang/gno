package p2p

import "fmt"

// SwitchOption is a callback used for configuring the p2p Switch
type SwitchOption func(*Switch)

// WithReactor sets the p2p switch reactors
func WithReactor(name string, reactor Reactor) SwitchOption {
	return func(sw *Switch) {
		for _, chDesc := range reactor.GetChannels() {
			chID := chDesc.ID
			// No two reactors can share the same channel
			if sw.reactorsByCh[chID] != nil {
				panic(fmt.Sprintf("Channel %X has multiple reactors %v & %v", chID, sw.reactorsByCh[chID], reactor))
			}

			sw.chDescs = append(sw.chDescs, chDesc)
			sw.reactorsByCh[chID] = reactor
		}

		sw.reactors[name] = reactor

		reactor.SetSwitch(sw)
	}
}

// WithPersistentPeers sets the p2p switch's persistent peer set
func WithPersistentPeers(peerAddrs []*NetAddress) SwitchOption {
	return func(sw *Switch) {
		for _, addr := range peerAddrs {
			sw.persistentPeers.Store(addr.ID, addr)
		}
	}
}
