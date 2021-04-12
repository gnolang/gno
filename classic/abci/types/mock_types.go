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

func (_ MockHeader) AssertABCIHeader() {}
