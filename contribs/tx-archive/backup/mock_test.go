package backup

import (
	"context"

	"github.com/gnolang/gno/contribs/tx-archive/backup/client"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
)

type (
	getLatestBlockNumberDelegate func() (uint64, error)
	getBlocksDelegate            func(context.Context, uint64, uint64) ([]*client.Block, error)
	getTxResultsDelegate         func(uint64) ([]*abci.ResponseDeliverTx, error)
)

type mockClient struct {
	getLatestBlockNumberFn getLatestBlockNumberDelegate
	getBlocksFn            getBlocksDelegate
	getTxResultsFn         getTxResultsDelegate
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

func (m *mockClient) GetTxResults(block uint64) ([]*abci.ResponseDeliverTx, error) {
	if m.getTxResultsFn != nil {
		return m.getTxResultsFn(block)
	}

	return nil, nil
}
