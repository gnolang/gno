package v1

import "errors"

var (
	ErrInvalidUsername = errors.New("r/gnoland/users: invalid username")
	ErrPaused          = errors.New("r/gnoland/users: paused")
)
