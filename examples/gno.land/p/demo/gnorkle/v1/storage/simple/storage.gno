package simple

import (
	"time"

	"gno.land/p/demo/gnorkle/v1/feed"
)

type Storage struct {
	values    []feed.Value
	maxValues int
}

func NewStorage(maxValues int) *Storage {
	return &Storage{
		maxValues: maxValues,
	}
}

func (s *Storage) Put(value string) {
	s.values = append(s.values, feed.Value{String: value, Time: time.Now()})
	if len(s.values) > s.maxValues && !(len(s.values) == 1 && s.maxValues == 0) {
		s.values = s.values[1:]
	}
}

func (s *Storage) GetLatest() feed.Value {
	if len(s.values) == 0 {
		return feed.Value{}
	}

	return s.values[len(s.values)-1]
}

func (s *Storage) GetHistory() []feed.Value {
	return s.values
}
