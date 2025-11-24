package restore

import (
	"context"

	"github.com/gnolang/gno/tm2/pkg/std"
)

type (
	sendTransactionDelegate func(context.Context, *std.Tx) error
)

type mockClient struct {
	sendTransactionFn sendTransactionDelegate
}

func (m *mockClient) SendTransaction(ctx context.Context, tx *std.Tx) error {
	if m.sendTransactionFn != nil {
		return m.sendTransactionFn(ctx, tx)
	}

	return nil
}

type (
	nextDelegate  func(context.Context) (*std.Tx, error)
	closeDelegate func() error
)

type mockSource struct {
	nextFn  nextDelegate
	closeFn closeDelegate
}

func (m *mockSource) Next(ctx context.Context) (*std.Tx, error) {
	if m.nextFn != nil {
		return m.nextFn(ctx)
	}

	return nil, nil
}

func (m *mockSource) Close() error {
	if m.closeFn != nil {
		return m.closeFn()
	}

	return nil
}
