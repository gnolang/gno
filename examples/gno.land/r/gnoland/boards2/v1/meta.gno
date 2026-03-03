package boards2

import "gno.land/p/gnoland/boards"

// Meta defines a type for board metadata.
type Meta struct {
	// HiddenThreads contains hidden board threads.
	HiddenThreads boards.PostStorage
}
