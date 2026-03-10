package boards2

import "gno.land/p/gnoland/boards"

// BoardMeta defines a type for board metadata.
type BoardMeta struct {
	// HiddenThreads contains hidden board threads.
	HiddenThreads boards.PostStorage
}
