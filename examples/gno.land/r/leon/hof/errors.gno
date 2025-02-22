package hof

import (
	"errors"
	"gno.land/p/leon/pkgerr"
)

var (
	ErrNoSuchItem     = pkgerr.New(errors.New("no such item exists"))
	ErrDoubleUpvote   = pkgerr.New(errors.New("cannot upvote twice"))
	ErrDoubleDownvote = pkgerr.New(errors.New("cannot downvote twice"))
)
