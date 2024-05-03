package simple

import (
	"time"

	"gno.land/p/demo/gnorkle/feed"
	"gno.land/p/demo/gnorkle/storage"
)

// Storage is simple, bounded storage for published feed values.
type Storage struct {
	values    []feed.Value
	maxValues uint
}

// NewStorage creates a new Storage with the given maximum number of values.
// If maxValues is 0, the storage is bounded to a size of one. If this is not desirable,
// then don't provide a value of 0.
func NewStorage(maxValues uint) *Storage {
	if maxValues == 0 {
		maxValues = 1
	}

	return &Storage{
		maxValues: maxValues,
	}
}

// Put adds a new value to the storage. If the storage is full, the oldest value
// is removed. If maxValues is 0, the storage is bounded to a size of one.
func (s *Storage) Put(value string) error {
	if s == nil {
		return storage.ErrUndefined
	}

	s.values = append(s.values, feed.Value{String: value, Time: time.Now()})
	if uint(len(s.values)) > s.maxValues {
		s.values = s.values[1:]
	}

	return nil
}

// GetLatest returns the most recently added value, or an empty value if none exist.
func (s Storage) GetLatest() feed.Value {
	if len(s.values) == 0 {
		return feed.Value{}
	}

	return s.values[len(s.values)-1]
}

// GetHistory returns all values in the storage, from oldest to newest.
func (s Storage) GetHistory() []feed.Value {
	return s.values
}
