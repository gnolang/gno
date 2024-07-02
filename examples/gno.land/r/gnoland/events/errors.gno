package events

import "errors"

var (
	errEmptyName        = errors.New("event name cannot be empty")
	errInvalidStartTime = errors.New("event start time cannot be less than 0")
)
