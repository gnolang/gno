package types

import "github.com/gnolang/gno/tm2/pkg/std"

// Archive wraps the backed-up chain data
type Archive struct {
	ChainData ChainData `json:"chainData"`
	Metadata  Metadata  `json:"metadata"`
}

// Metadata contains contextual information about the archive
type Metadata struct {
	EarliestBlockHeight uint64 `json:"earliestBlockHeight"`
	EarliestTxHash      uint64 `json:"earliestTxHash"`

	LatestBlockHeight uint64 `json:"latestBlockHeight"`
	LatestTxHash      uint64 `json:"latestTxHash"`
}

// ChainData contains the historical transaction data
type ChainData struct {
	Txs      []std.Tx `json:"txs"`
	BlockNum uint64   `json:"blockNum"`
}
