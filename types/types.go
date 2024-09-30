package types

import (
	"github.com/gnolang/gno/tm2/pkg/std"
)

// TxData contains the single block transaction,
// along with the block information
type TxData struct {
	Tx       std.Tx `json:"tx"`
	BlockNum uint64 `json:"blockNum"`

	// Timestamp contains the block creation time in unix milliseconds
	Timestamp int64 `json:"bt"` //nolint:tagliatelle // this name reduces disk space usage
}
