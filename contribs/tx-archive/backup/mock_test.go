package backup

import (
	"context"

	"github.com/gnolang/gno/contribs/tx-archive/backup/client"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
)

type (
	getLatestBlockNumberDelegate func() (uint64, error)
	getChainIDDelegate           func() (string, error)
	getBlocksDelegate            func(context.Context, uint64, uint64) ([]*client.Block, error)
	getTxResultsDelegate         func(uint64) ([]*abci.ResponseDeliverTx, error)
	getAccountAtHeightDelegate   func(crypto.Address, uint64) (uint64, uint64, error)
)

type mockClient struct {
	getLatestBlockNumberFn getLatestBlockNumberDelegate
	getChainIDFn           getChainIDDelegate
	getBlocksFn            getBlocksDelegate
	getTxResultsFn         getTxResultsDelegate
	getAccountAtHeightFn   getAccountAtHeightDelegate
}

func (m *mockClient) GetLatestBlockNumber() (uint64, error) {
	if m.getLatestBlockNumberFn != nil {
		return m.getLatestBlockNumberFn()
	}

	return 0, nil
}

func (m *mockClient) GetChainID() (string, error) {
	if m.getChainIDFn != nil {
		return m.getChainIDFn()
	}

	return "test-chain", nil
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

func (m *mockClient) GetAccountAtHeight(addr crypto.Address, height uint64) (uint64, uint64, error) {
	if m.getAccountAtHeightFn != nil {
		return m.getAccountAtHeightFn(addr, height)
	}

	return 0, 0, nil
}
