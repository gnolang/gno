package gnorkle

import (
	"bufio"
	"bytes"

	"gno.land/p/demo/gnorkle/feed"
	"gno.land/p/demo/gnorkle/message"
	"gno.land/p/demo/ufmt"
)

// Feed is an abstraction used by a gnorkle `Instance` to ingest data from
// agents and provide data feeds to consumers.
type Feed interface {
	ID() string
	Type() feed.Type
	Value() (value feed.Value, dataType string, consumable bool)
	Ingest(funcType message.FuncType, rawMessage, providerAddress string) error
	Tasks() []feed.Task
	IsActive() bool
}

// FeedWithWhitelist associates a `Whitelist` with a `Feed`.
type FeedWithWhitelist struct {
	Feed
	Whitelist
}

// MarshalJSON marshals the components of the feed that are needed for
// an agent to execute tasks and send values for ingestion.
func MarshalFeedJSON(feed Feed) ([]byte, error) {
    buf := new(bytes.Buffer)
	w := bufio.NewWriter(buf)

	_, dataType, _ := feed.Value()
	w.Write([]byte(
		`{"id":"` + feed.ID() +
			`","type":"` + ufmt.Sprintf("%d", int(feed.Type())) +
			`","value_type":"` + dataType +
			`","tasks":[`),
	)

	first := true
	for _, task := range feed.Tasks() {
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