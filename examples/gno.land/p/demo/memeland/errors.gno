package memeland

import "errors"

var (
	ErrInvalidInput = errors.New("missing meme or time")
	ErrInvalidTime  = errors.New("invalid time passed in")
)
