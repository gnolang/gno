package p2p

// WithPeerFilters sets the filters for rejection of new peers.
func WithPeerFilters(filters ...PeerFilterFunc) SwitchOption {
	return func(sw *Switch) {
		sw.peerFilters = filters
	}
}

// WithNodeInfo sets the node info for the p2p switch
func WithNodeInfo(ni NodeInfo) SwitchOption {
	return func(sw *Switch) {
		sw.nodeInfo = ni
	}
}

// WithNodeKey sets the node p2p key, utilized by the switch
func WithNodeKey(key *NodeKey) SwitchOption {
	return func(sw *Switch) {
		sw.nodeKey = key
	}
}
