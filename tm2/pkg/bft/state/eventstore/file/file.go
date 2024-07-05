package file

import (
	"fmt"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/autofile"
	storetypes "github.com/gnolang/gno/tm2/pkg/bft/state/eventstore/types"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/errors"
)

const (
	EventStoreType = "file"
	Path           = "path"
)

var (
	errMissingPath = errors.New("missing path param")
	errInvalidType = errors.New("invalid config for file event store specified")
)

// TxEventStore is the implementation of a transaction event store
// that outputs to the local filesystem
type TxEventStore struct {
	headPath string
	group    *autofile.Group
}

// NewTxEventStore creates a new file-based tx event store
func NewTxEventStore(cfg *storetypes.Config) (*TxEventStore, error) {
	// Parse config params
	if EventStoreType != cfg.EventStoreType {
		return nil, errInvalidType
	}

	headPath, ok := cfg.GetParam(Path).(string)
	if !ok {
		return nil, errMissingPath
	}

	return &TxEventStore{
		headPath: headPath,
	}, nil
}

// Start starts the file transaction event store, by opening the autofile group
func (t *TxEventStore) Start() error {
	// Open the group
	group, err := autofile.OpenGroup(t.headPath)
	if err != nil {
		return fmt.Errorf("unable to open file group for writing, %w", err)
	}

	t.group = group

	return nil
}

// Stop stops the file transaction event store, by closing the autofile group
func (t *TxEventStore) Stop() error {
	// Close off the group
	t.group.Close()

	return nil
}

// GetType returns the file transaction event store type
func (t *TxEventStore) GetType() string {
	return EventStoreType
}

// Append marshals the transaction using amino, and writes it to the disk
func (t *TxEventStore) Append(tx types.TxResult) error {
	// Serialize the transaction using amino
	txRaw, err := amino.MarshalJSON(tx)
	if err != nil {
		return fmt.Errorf("unable to marshal transaction, %w", err)
	}

	// Write the serialized transaction info to the file group
	if err = t.group.WriteLine(string(txRaw)); err != nil {
		return fmt.Errorf("unable to save transaction event, %w", err)
	}

	// Flush output to storage
	if err := t.group.FlushAndSync(); err != nil {
		return fmt.Errorf("unable to flush and sync transaction event, %w", err)
	}

	return nil
}
