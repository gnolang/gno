package blog

import "errors"

var (
	ErrPostAlreadyExists = errors.New("post already exists")

	ErrPostNotFound    = errors.New("post not found")
	ErrCommentNotFound = errors.New("comment not found")
	ErrLikesDisabled   = errors.New("likes are disabled")
	ErrAlreadyLiked    = errors.New("this address has already liked")
	ErrNotLiked        = errors.New("this address hasn't liked yet")

	ErrInvalidPostID    = errors.New("invalid post ID")
	ErrInvalidCommentID = errors.New("invalid comment ID")
	ErrInvalidPost      = errors.New("post not found/ invalid")

	ErrEmptyPrefix  = errors.New("prefix cannot be empty")
	ErrEmptyTitle   = errors.New("title cannot be empty")
	ErrEmptyBody    = errors.New("body cannot be empty")
	ErrEmptySlug    = errors.New("slug cannot be empty")
	ErrEmptyComment = errors.New("content of comment cannot be empty")

	ErrDeleteFailed        = errors.New("delete operation failed")
	ErrUnlikeFailed        = errors.New("failed to unlike post")
	ErrCommentDeleteFailed = errors.New("failed to delete comment")

	ErrInvalidCaller = errors.New("invalid caller address")
	ErrNotSuperuser  = errors.New("address is not owner")
)
