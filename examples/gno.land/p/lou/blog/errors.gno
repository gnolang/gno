package blog

import "errors"

var (
	ErrPostAlreadyExists  = errors.New("post already exists")

	ErrPostNotFound        = errors.New("post not found")
	ErrCommentNotFound     = errors.New("comment not found")
	
	ErrInvalidPostID       = errors.New("invalid post ID")
	ErrInvalidCommentID    = errors.New("invalid comment ID")
	
	ErrEmptyPrefix         = errors.New("prefix cannot be empty")
	ErrEmptyTitle          = errors.New("title cannot be empty")
	ErrEmptyBody           = errors.New("body cannot be empty")
	ErrEmptySlug           = errors.New("slug cannot be empty")
	ErrEmptyComment        = errors.New("content of comment cannot be empty")

	ErrDeleteFailed        = errors.New("delete operation failed")
	ErrPostAddFailed       = errors.New("failed to add post")
	ErrCommentDeleteFailed = errors.New("failed to delete comment")
	ErrCommentAddFailed    = errors.New("failed to add comment")
)
