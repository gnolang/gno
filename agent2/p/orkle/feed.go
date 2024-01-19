package orkle

import (
	"github.com/gnolang/gno/agent2/p/orkle/feed"
	"github.com/gnolang/gno/agent2/p/orkle/message"
)

type Feed interface {
	ID() string // necessary?
	Type() feed.Type
	Value() (value feed.Value, dataType string, consumable bool)
	Ingest(funcType message.FuncType, rawMessage, providerAddress string)
	MarshalJSON() ([]byte, error)
	HasAddressWhitelisted(address string) (isWhitelisted, feedHasWhitelist bool)
	Tasks() []feed.Task
	IsActive() bool
}
