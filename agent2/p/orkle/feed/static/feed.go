package static

import (
	"github.com/gnolang/gno/agent2/p/orkle"
	"github.com/gnolang/gno/agent2/p/orkle/feed"
	"github.com/gnolang/gno/agent2/p/orkle/message"
	"gno.land/p/demo/std"
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

func (f *Feed) Ingest(msg string) {
	if f.isLocked {
		panic("feed locked")
	}

	origCaller := string(std.GetOrigCaller())

	msgFunc, msg := message.ParseFunc(msg)
	switch msgFunc {
	case message.FuncTypeIngest:
		canAutoCommit := f.ingester.Ingest(msg, origCaller)

		// Autocommit the ingester's value if it's a single value ingester
		// because this is a static feed and this is the only value it will ever have.
		if canAutoCommit {
			f.ingester.CommitValue(f.storage, origCaller)
			f.isLocked = true
		}

	case message.FuncTypeCommit:
		f.ingester.CommitValue(f.storage, origCaller)
		f.isLocked = true

	default:
		panic("invalid message function " + string(msgFunc))
	}
}

func (f *Feed) Consumable() bool {
	return f.isLocked
}

func (f *Feed) Value() (feed.Value, string) {
	return f.storage.GetLatest(), f.valueDataType
}
