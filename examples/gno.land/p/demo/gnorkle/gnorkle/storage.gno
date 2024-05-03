package gnorkle

import "gno.land/p/demo/gnorkle/feed"

// Storage defines how published feed values should be read
// and written.
type Storage interface {
	Put(value string) error
	GetLatest() feed.Value
	GetHistory() []feed.Value
}
