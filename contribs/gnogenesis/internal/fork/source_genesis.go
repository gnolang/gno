package fork

import (
	"context"

	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
)

// GenesisSource provides the source chain's genesis document. Two
// implementations live alongside this file, picked via mutually-exclusive
// --source-genesis-* flags:
//
//   - rpcGenesisSource  (--source-genesis-rpc <url>):    source_genesis_rpc.go
//   - fileGenesisSource (--source-genesis-file <path>):  source_genesis_file.go
//
// Historical transactions are fetched separately via a TxsSource
// (see source_txs.go).
type GenesisSource interface {
	// Description returns a human-readable source type label.
	Description() string

	// FetchGenesis returns the source chain's genesis document.
	FetchGenesis(ctx context.Context) (*bft.GenesisDoc, error)

	// Close releases any resources held by the source.
	Close() error
}
