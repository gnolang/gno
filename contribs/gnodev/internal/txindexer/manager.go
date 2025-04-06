package txindexer

import "context"

// Manager provides the means to manage the tx-indexer application by providing
// functionality to start, and reload the process.
type Manager interface {
	Start(ctx context.Context) error
	Reload(ctx context.Context) error
}
