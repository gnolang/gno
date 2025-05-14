package mock

import "github.com/gnolang/gno/tm2/pkg/bft/types"

type (
	UpdateDelegate  func(types.Txs, int64)
	PendingDelegate func(int64, int64) types.Txs
)

type Mempool struct {
	UpdateFn  UpdateDelegate
	PendingFn PendingDelegate
}

func (m *Mempool) Update(txs types.Txs, updatedTxSize int64) {
	if m.UpdateFn != nil {
		m.UpdateFn(txs, updatedTxSize)
	}
}

func (m *Mempool) Pending(maxSizeBytes int64, maxGas int64) types.Txs {
	if m.PendingFn != nil {
		return m.PendingFn(maxSizeBytes, maxGas)
	}

	return nil
}
