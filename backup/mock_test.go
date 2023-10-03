package backup

import "github.com/gnolang/gno/tm2/pkg/std"

type (
	getLatestBlockNumberDelegate func() (uint64, error)
	getBlockTransactionsDelegate func(uint64) ([]std.Tx, error)
)

type mockClient struct {
	getLatestBlockNumberFn getLatestBlockNumberDelegate
	getBlockTransactionsFn getBlockTransactionsDelegate
}

func (m *mockClient) GetLatestBlockNumber() (uint64, error) {
	if m.getLatestBlockNumberFn != nil {
		return m.getLatestBlockNumberFn()
	}

	return 0, nil
}

func (m *mockClient) GetBlockTransactions(blockNum uint64) ([]std.Tx, error) {
	if m.getBlockTransactionsFn != nil {
		return m.getBlockTransactionsFn(blockNum)
	}

	return nil, nil
}
