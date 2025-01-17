package periodic

import (
	"bufio"
	"bytes"
	"errors"

	"gno.land/p/demo/gnorkle/feed"
	"gno.land/p/demo/gnorkle/gnorkle"
	"gno.land/p/demo/gnorkle/ingesters/multi"
	"gno.land/p/demo/gnorkle/message"
	"gno.land/p/demo/gnorkle/storage/simple"
	"gno.land/p/demo/ufmt"
)

// Feed is a periodic feed.
type Feed struct {
	id                string
	valueDataType     string
	timeBetweenCommit uint
	lastCommitTime    uint
	ingester          gnorkle.Ingester
	storage           gnorkle.Storage
	tasks             []feed.Task
}

// NewFeed creates a new periodic feed.
func NewFeed(
	id string,
	valueDataType string,
	timeBetweenCommit uint,
	ingester gnorkle.Ingester,
	storage gnorkle.Storage,
	tasks ...feed.Task,
) *Feed {
	return &Feed{
		id:                id,
		valueDataType:     valueDataType,
		timeBetweenCommit: timeBetweenCommit,
		ingester:          ingester,
		storage:           storage,
		tasks:             tasks,
	}
}

// NewMultiValueFeed is a convenience function  for creating a periodic feed
// that commits values every x times.
func NewMultiValueFeed(
	id string,
	valueDataType string,
	timeBetweenCommit uint,
	tasks ...feed.Task,
) *Feed {
	return NewFeed(
		id,
		valueDataType,
		timeBetweenCommit,
		&multi.ValueIngester{},
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
	return feed.TypePeriodic
}

// Ingest ingests a message into the feed. It either adds the value to the ingester's
// pending values or commits the value to the storage.
func (f *Feed) Ingest(funcType message.FuncType, msg, providerAddress string) error {
	if f == nil {
		return feed.ErrUndefined
	}

	switch funcType {
	case message.FuncTypeIngest:
		_, err := f.ingester.Ingest(msg, providerAddress)
		if err != nil {
			return err
		}

	case message.FuncTypeCommit:
		// Commits the value to the storage only after x times has passed since last commit.
		if f.lastCommitTime+f.timeBetweenCommit > f.storage.GetTime() {
			return errors.New("time between commit not reached")
		}

		if err := f.ingester.CommitValue(f.storage, providerAddress); err != nil {
			return err
		}

		f.lastCommitTime = f.storage.GetTime()

	default:
		return errors.New("invalid message function " + string(funcType))
	}

	return nil
}

// Value returns the feed's latest value, it's data type, and whether or not it can
// be safely consumed. In this case it can be consumed because it's a periodic feed.
func (f Feed) Value() (feed.Value, string, bool) {
	lastValue := f.storage.GetLatest()
	return lastValue, f.valueDataType, lastValue == feed.Value{}
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
	return true
}
