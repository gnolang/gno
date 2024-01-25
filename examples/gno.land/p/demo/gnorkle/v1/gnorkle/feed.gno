package gnorkle

import (
	"gno.land/p/demo/gnorkle/v1/feed"
	"gno.land/p/demo/gnorkle/v1/message"
)

// Feed is an abstraction used by a gnorkle `Instance` to ingest data from
// agents and provide data feeds to consumers.
type Feed interface {
	ID() string
	Type() feed.Type
	Value() (value feed.Value, dataType string, consumable bool)
	Ingest(funcType message.FuncType, rawMessage, providerAddress string)
	MarshalJSON() ([]byte, error)
	Tasks() []feed.Task
	IsActive() bool
}

// FeedWithWhitelist associates a `Whitelist` with a `Feed`.
type FeedWithWhitelist struct {
	Feed
	Whitelist
}
