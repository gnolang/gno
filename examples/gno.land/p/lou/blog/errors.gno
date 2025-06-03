package blog

import "errors"

var (
	ErrPostNotFound       = errors.New("post not found")
	ErrCommentNotFound    = errors.New("comment not found")
	ErrInvalidPostID      = errors.New("invalid post ID")
	ErrInvalidCommentID   = errors.New("invalid comment ID")
	ErrEmptyTitle         = errors.New("title cannot be empty")
	ErrEmptyBody          = errors.New("body cannot be empty")
	ErrEmptySlug          = errors.New("slug cannot be empty")
	ErrInvalidPublication = errors.New("invalid publication date format")
)
