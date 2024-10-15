package static

import (
	"bufio"
	"bytes"
	"errors"

	"gno.land/p/demo/gnorkle/feed"
	"gno.land/p/demo/gnorkle/gnorkle"
	"gno.land/p/demo/gnorkle/ingesters/single"
	"gno.land/p/demo/gnorkle/message"
	"gno.land/p/demo/gnorkle/storage/simple"
	"gno.land/p/demo/ufmt"
)

// Feed is a static feed.
type Feed struct {
	id            string
	isLocked      bool
	valueDataType string
	ingester      gnorkle.Ingester
	storage       gnorkle.Storage
	tasks         []feed.Task
}

// NewFeed creates a new static feed.
func NewFeed(
	id string,
	valueDataType string,
	ingester gnorkle.Ingester,
	storage gnorkle.Storage,
	tasks ...feed.Task,
) *Feed {
	return &Feed{
		id:            id,
		valueDataType: valueDataType,
		ingester:      ingester,
		storage:       storage,
		tasks:         tasks,
	}
}

// NewSingleValueFeed is a convenience function  for creating a static feed
// that autocommits a value after a single ingestion.
func NewSingleValueFeed(
	id string,
	valueDataType string,
	tasks ...feed.Task,
) *Feed {
	return NewFeed(
		id,
		valueDataType,
		&single.ValueIngester{},
		simple.NewStorage(1),
		tasks...,
	)
}

// ID returns the feed's ID.
func (f Feed) ID() string {
	return f.id
}

// Type returns the feed's type.
func (f Feed) Type() feed.Type {
	return feed.TypeStatic
}

// Ingest ingests a message into the feed. It either adds the value to the ingester's
// pending values or commits the value to the storage.
func (f *Feed) Ingest(funcType message.FuncType, msg, providerAddress string) error {
	if f == nil {
		return feed.ErrUndefined
	}

	if f.isLocked {
		return errors.New("feed locked")
	}

	switch funcType {
	case message.FuncTypeIngest:
		// Autocommit the ingester's value if it's a single value ingester
		// because this is a static feed and this is the only value it will ever have.
		if canAutoCommit, err := f.ingester.Ingest(msg, providerAddress); canAutoCommit && err == nil {
			if err := f.ingester.CommitValue(f.storage, providerAddress); err != nil {
				return err
			}

			f.isLocked = true
		} else if err != nil {
			return err
		}

	case message.FuncTypeCommit:
		if err := f.ingester.CommitValue(f.storage, providerAddress); err != nil {
			return err
		}

		f.isLocked = true

	default:
		return errors.New("invalid message function " + string(funcType))
	}

	return nil
}

// Value returns the feed's latest value, it's data type, and whether or not it can
// be safely consumed. In this case it uses `f.isLocked` because, this being a static
// feed, it will only ever have one value; once that value is committed the feed is locked
// and there is a valid, non-empty value to consume.
func (f Feed) Value() (feed.Value, string, bool) {
	return f.storage.GetLatest(), f.valueDataType, f.isLocked
}

// MarshalJSON marshals the components of the feed that are needed for
// an agent to execute tasks and send values for ingestion.
func (f Feed) MarshalJSON() ([]byte, error) {
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

		taskJSON, err := task.MarshalJSON()
		if err != nil {
			return nil, err
		}

		w.Write(taskJSON)
		first = false
	}

	w.Write([]byte("]}"))
	w.Flush()

	return buf.Bytes(), nil
}

// Tasks returns the feed's tasks. This allows task consumers to extract task
// contents without having to marshal the entire feed.
func (f Feed) Tasks() []feed.Task {
	return f.tasks
}

// IsActive returns true if the feed is accepting ingestion requests from agents.
func (f Feed) IsActive() bool {
	return !f.isLocked
}
