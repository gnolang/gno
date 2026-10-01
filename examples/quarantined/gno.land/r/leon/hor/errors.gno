package hor

import (
	"gno.land/p/leon/pkgerr"
)

var (
	ErrNoSuchItem     = pkgerr.New("no such item exists")
	ErrDoubleUpvote   = pkgerr.New("cannot upvote twice")
	ErrDoubleDownvote = pkgerr.New("cannot downvote twice")
)
