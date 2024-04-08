package client

import (
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/client/ws"
)

var _ Client = (*WS)(nil)

type WS struct {
	rpc *ws.Client

	*baseRPCClient
}

// func NewWS(remote, endpoint string) *WS {
// 	return &WS{
// 		rpc: ws.NewClient(remote, endpoint),
// 	}
// }
//
// // NewBatch creates a new rpcBatch client for this HTTP client.
// func (c *WS) NewBatch() *Batch {
// 	batch := rpcclient.NewRPCRequestBatch(c.rpc)
//
// 	return &Batch{
// 		rpcBatch: batch,
// 		baseRPCClient: &baseRPCClient{
// 			caller: batch,
// 		},
// 	}
// }
