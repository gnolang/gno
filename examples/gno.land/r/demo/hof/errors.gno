package hof

import (
	"errors"
)

var (
	ErrNoSuchItem     = errors.New("hof: no such item exists")
	ErrAlreadyExists  = errors.New("hof: alerady exists")
	ErrDoubleUpvote   = errors.New("hof: cannot upvote twice")
	ErrDoubleDownvote = errors.New("hof: cannot downvote twice")
	ErrNonCodeCall    = errors.New("hof: non-code call")
	ErrPaused         = errors.New("hof: registering is paused")
)
