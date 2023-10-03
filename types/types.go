package types

import "github.com/gnolang/gno/tm2/pkg/std"

// Archive wraps the backed-up chain data
type Archive struct {
	Metadata  *Metadata    `json:"metadata"`
	BlockData []*BlockData `json:"blockData"`
}

// Metadata contains contextual information about the archive
type Metadata struct {
	EarliestBlockHeight uint64 `json:"earliestBlockHeight"`
	LatestBlockHeight   uint64 `json:"latestBlockHeight"`
}

// BlockData contains the historical transaction data
type BlockData struct {
	Txs      []std.Tx `json:"txs"`
	BlockNum uint64   `json:"blockNum"`
}
