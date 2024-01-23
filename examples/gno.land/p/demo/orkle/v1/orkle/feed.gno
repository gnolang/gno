package orkle

import (
	"gno.land/p/demo/orkle/v1/feed"
	"gno.land/p/demo/orkle/v1/message"
)

type Feed interface {
	ID() string // necessary?
	Type() feed.Type
	Value() (value feed.Value, dataType string, consumable bool)
	Ingest(funcType message.FuncType, rawMessage, providerAddress string)
	MarshalJSON() ([]byte, error)
	Tasks() []feed.Task
	IsActive() bool
}

type FeedWithWhitelist struct {
	Feed
	Whitelist
}
