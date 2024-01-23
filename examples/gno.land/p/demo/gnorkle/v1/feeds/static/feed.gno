package static

import (
	"bufio"
	"bytes"

	"gno.land/p/demo/gnorkle/v1/feed"
	"gno.land/p/demo/gnorkle/v1/gnorkle"
	"gno.land/p/demo/gnorkle/v1/ingesters/single"
	"gno.land/p/demo/gnorkle/v1/message"
	"gno.land/p/demo/gnorkle/v1/storage/simple"
	"gno.land/p/demo/ufmt"
)

type Feed struct {
	id            string
	isLocked      bool
	valueDataType string
	ingester      gnorkle.Ingester
	storage       gnorkle.Storage
	tasks         []feed.Task
}

func NewFeed(
	id string,
	valueDataType string,
	ingester gnorkle.Ingester,
	storage gnorkle.Storage,
	tasks []feed.Task,
) *Feed {
	return &Feed{
		id:            id,
		valueDataType: valueDataType,
		ingester:      ingester,
		storage:       storage,
		tasks:         tasks,
	}
}

func NewSingleValueFeed(
	id string,
	valueDataType string,
	tasks []feed.Task,
) *Feed {
	return NewFeed(
		id,
		valueDataType,
		&single.ValueIngester{},
		simple.NewStorage(1),
		tasks,
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
			`","type":"` + ufmt.Sprintf("%d", int(f.Type())) +
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

func (f *Feed) Tasks() []feed.Task {
	return f.tasks
}

func (f *Feed) IsActive() bool {
	return !f.isLocked
}
