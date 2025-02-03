package backup

import (
	"context"

	"github.com/gnolang/tx-archive/backup/client"
)

type (
	getLatestBlockNumberDelegate func() (uint64, error)
	getBlocksDelegate            func(context.Context, uint64, uint64) ([]*client.Block, error)
)

type mockClient struct {
	getLatestBlockNumberFn getLatestBlockNumberDelegate
	getBlocksFn            getBlocksDelegate
}

func (m *mockClient) GetLatestBlockNumber() (uint64, error) {
	if m.getLatestBlockNumberFn != nil {
		return m.getLatestBlockNumberFn()
	}

	return 0, nil
}

func (m *mockClient) GetBlocks(ctx context.Context, from, to uint64) ([]*client.Block, error) {
	if m.getBlocksFn != nil {
		return m.getBlocksFn(ctx, from, to)
	}

	return nil, nil
}
