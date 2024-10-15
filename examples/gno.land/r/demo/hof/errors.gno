package hof

import (
	"errors"
)

var (
	ErrNoSuchItem     = errors.New("hof: no such item exists")
	ErrDoubleUpvote   = errors.New("hof: cannot upvote twice")
	ErrDoubleDownvote = errors.New("hof: cannot downvote twice")
)
