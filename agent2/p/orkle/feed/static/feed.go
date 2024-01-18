package static

import (
	"github.com/gnolang/gno/agent2/p/orkle"
	"github.com/gnolang/gno/agent2/p/orkle/feed"
	"github.com/gnolang/gno/agent2/p/orkle/message"
)

type Feed struct {
	id            string
	isLocked      bool
	valueDataType string
	ingester      orkle.Ingester
	storage       orkle.Storage
}

func (f *Feed) ID() string {
	return f.id
}

func (f *Feed) Type() feed.Type {
	return feed.TypeStatic
}

func (f *Feed) Ingest(funcType message.FuncType, msg, providerAddress string) {
	if f.isLocked {
		panic("feed locked")
	}

	switch funcType {
	case message.FuncTypeIngest:
		// Autocommit the ingester's value if it's a single value ingester
		// because this is a static feed and this is the only value it will ever have.
		if canAutoCommit := f.ingester.Ingest(msg, providerAddress); canAutoCommit {
			f.ingester.CommitValue(f.storage, providerAddress)
			f.isLocked = true
		}

	case message.FuncTypeCommit:
		f.ingester.CommitValue(f.storage, providerAddress)
		f.isLocked = true

	default:
		panic("invalid message function " + string(funcType))
	}
}

func (f *Feed) Value() (feed.Value, string, bool) {
	return f.storage.GetLatest(), f.valueDataType, f.isLocked
}
