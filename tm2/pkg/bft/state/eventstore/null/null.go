package null

import (
	"github.com/gnolang/gno/tm2/pkg/bft/state/eventstore"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
)

var _ eventstore.TxEventStore = (*TxEventStore)(nil)

const (
	EventStoreType = "none"
)

// TxEventStore acts as a /dev/null
type TxEventStore struct{}

func NewNullEventStore() *TxEventStore {
	return &TxEventStore{}
}

func (t TxEventStore) Start() error {
	return nil
}

func (t TxEventStore) Stop() error {
	return nil
}

func (t TxEventStore) Append(_ types.TxResult) error {
	return nil
}

func (t TxEventStore) GetType() string {
	return EventStoreType
}
