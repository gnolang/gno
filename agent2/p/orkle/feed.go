package orkle

import "github.com/gnolang/gno/agent2/p/orkle/feed"

type Feed interface {
	ID() string // necessary?
	Type() feed.Type
	Value() (value feed.Value, dataType string)
	Ingest(rawMessage, providerAddress string)
	Consumable() bool
	MarshalJSON() ([]byte, error)
}
