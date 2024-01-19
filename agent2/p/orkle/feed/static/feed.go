package static

import (
	"bufio"
	"bytes"
	ufmt "fmt"

	"github.com/gnolang/gno/agent2/p/orkle"
	"github.com/gnolang/gno/agent2/p/orkle/feed"
	"github.com/gnolang/gno/agent2/p/orkle/ingester/single"
	"github.com/gnolang/gno/agent2/p/orkle/message"
	"github.com/gnolang/gno/agent2/p/orkle/storage"
	"gno.land/p/demo/avl"
)

type Feed struct {
	id            string
	isLocked      bool
	valueDataType string
	whitelist     *avl.Tree
	ingester      orkle.Ingester
	storage       orkle.Storage
	tasks         []feed.Task
}

func NewFeed(
	id string,
	valueDataType string,
	whitelist []string,
	ingester orkle.Ingester,
	storage orkle.Storage,
	tasks ...feed.Task,
) *Feed {
	feed := &Feed{
		id:            id,
		valueDataType: valueDataType,
		ingester:      ingester,
		storage:       storage,
		tasks:         tasks,
	}

	if len(whitelist) != 0 {
		feed.whitelist = avl.NewTree()
		for _, address := range whitelist {
			feed.whitelist.Set(address, struct{}{})
		}
	}

	return feed
}

func NewSingleValueFeed(
	id string,
	valueDataType string,
	whitelist []string,
	tasks ...feed.Task,
) *Feed {
	return NewFeed(
		id,
		valueDataType,
		whitelist,
		&single.ValueIngester{},
		storage.NewSimple(1),
		tasks...,
	)
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

func (f *Feed) MarshalJSON() ([]byte, error) {
	buf := new(bytes.Buffer)
	w := bufio.NewWriter(buf)

	w.Write([]byte(
		`{"id":"` + f.id +
			`","type":"` + ufmt.Sprintf("%d", f.Type()) +
			`","value_type":"` + f.valueDataType +
			`","tasks":[`),
	)

	first := true
	for _, task := range f.tasks {
		if !first {
			w.WriteString(",")
		}

		taskJSON, err := task.MarshalToJSON()
		if err != nil {
			return nil, err
		}

		w.Write(taskJSON)
	}

	w.Write([]byte("]}"))
	w.Flush()

	return buf.Bytes(), nil
}

func (f *Feed) HasAddressWhitelisted(address string) (isWhitelisted, feedHasWhitelist bool) {
	if f.whitelist == nil {
		return true, false
	}

	return f.whitelist.Has(address), true
}

func (f *Feed) Tasks() []feed.Task {
	return f.tasks
}

func (f *Feed) IsActive() bool {
	return !f.isLocked
}
