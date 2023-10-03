package client

import "github.com/gnolang/gno/tm2/pkg/std"

type Client interface {
	GetLatestBlockNumber() (uint64, error)
	GetBlockTransactions(uint64) ([]std.Tx, error)
}
