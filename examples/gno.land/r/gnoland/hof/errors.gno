package hof

import "errors"

var (
	ErrNoSuchItem       = errors.New("hof: no such item exists")
	ErrNoSuchExhibition = errors.New("hof: no such exhibition exists")
	ErrInvalidTitle     = errors.New("hof: invalid title")
	ErrDoubleUpvote     = errors.New("hof: cannot upvote twice")
	ErrDoubleDownvote   = errors.New("hof: cannot downvote twice")
)
