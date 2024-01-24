package gnorkle

import "gno.land/p/demo/gnorkle/v1/feed"

// Storage defines how published feed values should be read
// and written.
type Storage interface {
	Put(value string)
	GetLatest() feed.Value
	GetHistory() []feed.Value
}
