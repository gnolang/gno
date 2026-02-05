package batch

import (
	"context"

	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/spec"
)

type sendBatchDelegate func(context.Context, spec.BaseJSONRequests) (spec.BaseJSONResponses, error)

type mockClient struct {
	sendBatchFn sendBatchDelegate
}

func (m *mockClient) SendBatch(ctx context.Context, requests spec.BaseJSONRequests) (spec.BaseJSONResponses, error) {
	if m.sendBatchFn != nil {
		return m.sendBatchFn(ctx, requests)
	}

	return nil, nil
}
