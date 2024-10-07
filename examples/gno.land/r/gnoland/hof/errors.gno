package hof

import (
	"errors"
)

var (
	ErrNoSuchItem       = errors.New("hof: no such item exists")
	ErrAlreadyExists    = errors.New("hof: alerady exists")
	ErrNoSuchExhibition = errors.New("hof: no such exhibition exists")
	ErrDoubleUpvote     = errors.New("hof: cannot upvote twice")
	ErrDoubleDownvote   = errors.New("hof: cannot downvote twice")
	ErrNonCodeCall      = errors.New("hof: non-code call")
)
