package blog

import "errors"

var (
	ErrNoSuchPost = errors.New("no such post")
)
