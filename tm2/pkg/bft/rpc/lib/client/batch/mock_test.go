package batch

import (
	"context"

	types "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
)

type sendBatchDelegate func(context.Context, types.RPCRequests) (types.RPCResponses, error)

type mockClient struct {
	sendBatchFn sendBatchDelegate
}

func (m *mockClient) SendBatch(ctx context.Context, requests types.RPCRequests) (types.RPCResponses, error) {
	if m.sendBatchFn != nil {
		return m.sendBatchFn(ctx, requests)
	}

	return nil, nil
}
