package backup

import (
	"github.com/gnolang/tx-archive/backup/client"
)

type (
	getLatestBlockNumberDelegate func() (uint64, error)
	getBlockDelegate             func(uint64) (*client.Block, error)
)

type mockClient struct {
	getLatestBlockNumberFn getLatestBlockNumberDelegate
	getBlockFn             getBlockDelegate
}

func (m *mockClient) GetLatestBlockNumber() (uint64, error) {
	if m.getLatestBlockNumberFn != nil {
		return m.getLatestBlockNumberFn()
	}

	return 0, nil
}

func (m *mockClient) GetBlock(blockNum uint64) (*client.Block, error) {
	if m.getBlockFn != nil {
		return m.getBlockFn(blockNum)
	}

	return nil, nil
}
