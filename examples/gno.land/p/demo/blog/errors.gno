package blog

import "errors"

var ErrNoSuchPost = errors.New("no such post")
var ErrNoSuchComment = errors.New("no such comment")
