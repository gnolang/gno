package core

import (
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	rpctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/gnolang/gno/tm2/pkg/telemetry/traces"
)

// NetInfo returns network info.
func (env *Environment) NetInfo(ctx *rpctypes.Context) (*ctypes.ResultNetInfo, error) {
	_, span := traces.Tracer().Start(ctx.Context(), "NetInfo")
	defer span.End()
	var (
		set     = env.P2PPeers.Peers()
		out, in = set.NumOutbound(), set.NumInbound()
	)

	peers := make([]ctypes.Peer, 0, out+in)
	for _, peer := range set.List() {
		nodeInfo := peer.NodeInfo()
		peers = append(peers, ctypes.Peer{
			NodeInfo:         nodeInfo,
			IsOutbound:       peer.IsOutbound(),
			ConnectionStatus: peer.Status(),
			RemoteIP:         peer.RemoteIP().String(),
		})
	}

	return &ctypes.ResultNetInfo{
		Listening: env.P2PTransport.IsListening(),
		Listeners: env.P2PTransport.Listeners(),
		NPeers:    len(peers),
		Peers:     peers,
	}, nil
}

// Genesis returns the genesis file.
func (env *Environment) Genesis(ctx *rpctypes.Context) (*ctypes.ResultGenesis, error) {
	_, span := traces.Tracer().Start(ctx.Context(), "Genesis")
	defer span.End()
	return &ctypes.ResultGenesis{Genesis: env.GenDoc}, nil
}
