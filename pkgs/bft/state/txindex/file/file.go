package file

import (
	"fmt"

	"github.com/gnolang/gno/pkgs/amino"
	"github.com/gnolang/gno/pkgs/autofile"
	"github.com/gnolang/gno/pkgs/bft/state/txindex/config"
	"github.com/gnolang/gno/pkgs/bft/types"
	"github.com/gnolang/gno/pkgs/errors"
)

const (
	IndexerType = "file-indexer"
	Path        = "path"
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
	headPath, ok := cfg.GetParam(Path).(string)
	if !ok {
		return nil, errors.New("missing path param")
	}

	return &TxIndexer{
		headPath: headPath,
	}, nil
}

func (t *TxIndexer) Start() error {
	// Open the group
	group, err := autofile.OpenGroup(t.headPath)
	if err != nil {
		return fmt.Errorf("unable to open file group for writing, %w", err)
	}

	t.group = group

	return nil
}

func (t *TxIndexer) Stop() error {
	// Close off the group
	t.group.Close()

	return nil
}

func (t *TxIndexer) GetType() string {
	return IndexerType
}

func (t *TxIndexer) Index(tx types.TxResult) error {
	// Serialize the transaction using amino:binary
	txRaw, err := amino.Marshal(tx)
	if err != nil {
		return fmt.Errorf("unable to marshal transaction, %w", err)
	}

	// Write the serialized transaction info to the file group
	if err = t.group.WriteLine(string(txRaw)); err != nil {
		return fmt.Errorf("unable to save transaction index, %w", err)
	}

	return nil
}
