package client

import (
	"context"

	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/spec"
)

type (
	sendRequestDelegate func(context.Context, *spec.BaseJSONRequest) (*spec.BaseJSONResponse, error)
	sendBatchDelegate   func(context.Context, spec.BaseJSONRequests) (spec.BaseJSONResponses, error)
	closeDelegate       func() error
)

type mockClient struct {
	sendRequestFn sendRequestDelegate
	sendBatchFn   sendBatchDelegate
	closeFn       closeDelegate
}

func (m *mockClient) SendRequest(ctx context.Context, request *spec.BaseJSONRequest) (*spec.BaseJSONResponse, error) {
	if m.sendRequestFn != nil {
		return m.sendRequestFn(ctx, request)
	}

	return nil, nil
}

func (m *mockClient) SendBatch(ctx context.Context, requests spec.BaseJSONRequests) (spec.BaseJSONResponses, error) {
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
