package client

import rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/client"

var (
	_ Client = (*WS)(nil)
)

type WS struct {
	rpc *rpcclient.WSClient

	*baseRPCClient
}

func NewWS(remote, endpoint string) *WS {
	return &WS{
		rpc: rpcclient.NewWSClient(remote, endpoint),
	}
}
