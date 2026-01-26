package net

import (
	"context"

	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/metadata"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/spec"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/telemetry/traces"
)

// Handler is the net RPC handler
type Handler struct {
	genesisDoc *types.GenesisDoc

	peers     ctypes.Peers
	transport ctypes.Transport
}

// NewHandler creates a new instance of the net RPC handler
func NewHandler(
	peers ctypes.Peers,
	transport ctypes.Transport,
	genesisDoc *types.GenesisDoc,
) *Handler {
	return &Handler{
		peers:      peers,
		transport:  transport,
		genesisDoc: genesisDoc,
	}
}

// NetInfoHandler fetches the current network info
//
//	No params
func (h *Handler) NetInfoHandler(_ *metadata.Metadata, p []any) (any, *spec.BaseJSONError) {
	if len(p) > 0 {
		return nil, spec.GenerateInvalidParamError(1)
	}

	_, span := traces.Tracer().Start(context.Background(), "NetInfo")
	defer span.End()

	var (
		set     = h.peers.Peers()
		out, in = set.NumOutbound(), set.NumInbound()
	)

	peers := make([]Peer, 0, out+in)
	for _, peer := range set.List() {
		peers = append(peers, Peer{
			NodeInfo:         peer.NodeInfo(),
			IsOutbound:       peer.IsOutbound(),
			ConnectionStatus: peer.Status(),
			RemoteIP:         peer.RemoteIP().String(),
		})
	}

	return &ResultNetInfo{
		Listening: h.transport.IsListening(),
		Listeners: h.transport.Listeners(),
		NPeers:    len(peers),
		Peers:     peers,
	}, nil
}

// GenesisHandler fetches the genesis document (genesis.json)
//
//	No params
func (h *Handler) GenesisHandler(_ *metadata.Metadata, p []any) (any, *spec.BaseJSONError) {
	if len(p) > 0 {
		return nil, spec.GenerateInvalidParamError(1)
	}

	_, span := traces.Tracer().Start(context.Background(), "Genesis")
	defer span.End()

	return &ResultGenesis{
		Genesis: h.genesisDoc,
	}, nil
}
