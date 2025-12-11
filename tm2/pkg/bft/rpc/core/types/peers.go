package core_types

import (
	"github.com/gnolang/gno/tm2/pkg/p2p"
	p2pTypes "github.com/gnolang/gno/tm2/pkg/p2p/types"
)

// Peers exposes access to the current P2P peer set
type Peers interface {
	// Peers returns the current peer set
	Peers() p2p.PeerSet
}

// Transport exposes read-only access to the P2P transport
type Transport interface {
	// Listeners returns the addresses the node is currently listening on
	Listeners() []string

	// IsListening reports whether the node is currently accepting incoming connections
	IsListening() bool

	// NodeInfo returns the local node's P2P identity and metadata
	NodeInfo() p2pTypes.NodeInfo
}
