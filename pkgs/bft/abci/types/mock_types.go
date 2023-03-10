package abci

import (
	"time"
)

// Only used for tests.
type MockHeader struct {
	Version  string    `json:"version"`
	ChainID  string    `json:"chain_id"`
	Height   int64     `json:"height"`
	Time     time.Time `json:"time"`
	NumTxs   int64     `json:"num_txs"`
	TotalTxs int64     `json:"total_txs"`
}

func (MockHeader) AssertABCIHeader() {}

func (mh MockHeader) GetChainID() string {
	return mh.ChainID
}

func (mh MockHeader) GetHeight() int64 {
	return mh.Height
}

func (mh MockHeader) GetTime() time.Time {
	return mh.Time
}
