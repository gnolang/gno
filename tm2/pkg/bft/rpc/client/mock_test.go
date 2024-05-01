package client

import (
	"context"

	types "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
)

type (
	sendRequestDelegate func(context.Context, types.RPCRequest) (*types.RPCResponse, error)
	sendBatchDelegate   func(context.Context, types.RPCRequests) (types.RPCResponses, error)
	closeDelegate       func() error
)

type mockClient struct {
	sendRequestFn sendRequestDelegate
	sendBatchFn   sendBatchDelegate
	closeFn       closeDelegate
}

func (m *mockClient) SendRequest(ctx context.Context, request types.RPCRequest) (*types.RPCResponse, error) {
	if m.sendRequestFn != nil {
		return m.sendRequestFn(ctx, request)
	}

	return nil, nil
}

func (m *mockClient) SendBatch(ctx context.Context, requests types.RPCRequests) (types.RPCResponses, error) {
	if m.sendBatchFn != nil {
		return m.sendBatchFn(ctx, requests)
	}

	return nil, nil
}

func (m *mockClient) Close() error {
	if m.closeFn != nil {
		return m.closeFn()
	}

	return nil
}
