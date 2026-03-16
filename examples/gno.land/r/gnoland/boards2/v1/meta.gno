package boards2

import "gno.land/p/gnoland/boards"

// BoardMeta defines a type for board metadata.
type BoardMeta struct {
	// HiddenThreads contains hidden board threads.
	HiddenThreads boards.PostStorage
}

// ThreadMeta defines a type for thread metadata.
type ThreadMeta struct {
	// AllReplies contains all existing thread comments and replies.
	AllReplies boards.PostStorage
}
