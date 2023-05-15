package file

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/autofile"
	"github.com/gnolang/gno/tm2/pkg/bft/state/txindex/config"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/errors"
)

const (
	IndexerType = "file"
	Path        = "path"
)

var (
	errMissingPath = errors.New("missing path param")
	errInvalidType = errors.New("invalid config for file indexer specified")
)

// TxIndexer is the implementation of a transaction indexer
// that outputs to the local filesystem
type TxIndexer struct {
	headPath string
	group    *autofile.Group
}

// NewTxIndexer creates a new file-based tx indexer
func NewTxIndexer(cfg *config.Config) (*TxIndexer, error) {
	// Parse config params
	if IndexerType != cfg.IndexerType {
		return nil, errInvalidType
	}

	headPath, ok := cfg.GetParam(Path).(string)
	if !ok {
		return nil, errMissingPath
	}

	return &TxIndexer{
		headPath: headPath,
	}, nil
}

// Start starts the file transaction indexer, by opening the autofile group
func (t *TxIndexer) Start() error {
	// Open the group
	group, err := autofile.OpenGroup(t.headPath)
	if err != nil {
		return fmt.Errorf("unable to open file group for writing, %w", err)
	}

	t.group = group

	return nil
}

// Stop stops the file transaction indexer, by closing the autofile group
func (t *TxIndexer) Stop() error {
	// Close off the group
	t.group.Close()

	return nil
}

// GetType returns the file transaction indexer type
func (t *TxIndexer) GetType() string {
	return IndexerType
}

// Index marshals the transaction using amino, and writes it to the disk
func (t *TxIndexer) Index(tx types.TxResult) error {
	// Serialize the transaction using amino
	txRaw, err := amino.MarshalJSON(tx)
	if err != nil {
		return fmt.Errorf("unable to marshal transaction, %w", err)
	}

	// Write the serialized transaction info to the file group
	if err = t.group.WriteLine(string(txRaw)); err != nil {
		return fmt.Errorf("unable to save transaction index, %w", err)
	}

	// Flush output to storage
	if err := t.group.FlushAndSync(); err != nil {
		return fmt.Errorf("unable to flush and sync transaction index, %w", err)
	}

	return nil
}
