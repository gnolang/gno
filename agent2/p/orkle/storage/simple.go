package storage

import (
	"time"

	"github.com/gnolang/gno/agent2/p/orkle/feed"
)

type Simple struct {
	values    []feed.Value
	maxValues int
}

func (s *Simple) Put(value string) {
	s.values = append(s.values, feed.Value{String: value, Time: time.Now()})
	if len(s.values) > s.maxValues {
		s.values = s.values[1:]
	}
}

func (s *Simple) GetLatest() feed.Value {
	if len(s.values) == 0 {
		return feed.Value{}
	}

	return s.values[len(s.values)-1]
}

func (s *Simple) GetHistory() []feed.Value {
	return s.values
}
