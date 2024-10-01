package hof

import (
	"errors"
	"strconv"
)

var (
	ErrNoSuchItem       = errors.New("hof: no such item exists")
	ErrNoSuchExhibition = errors.New("hof: no such exhibition exists")
	ErrInvalidTitle     = errors.New("hof: invalid title")
	ErrDoubleUpvote     = errors.New("hof: cannot upvote twice")
	ErrDoubleDownvote   = errors.New("hof: cannot downvote twice")
	ErrMinWidgetSize    = errors.New("hof: widget size cannot be less than 1")
	ErrMaxWidgetSize    = errors.New("hof: widget size cannot be more than " + strconv.Itoa(MaxWidgetSize))
	ErrNonCodeCall      = errors.New("hof: non-code call")
)
