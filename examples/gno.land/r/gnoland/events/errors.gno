package events

import "errors"

var (
	errEmptyName        = errors.New("event name cannot be empty")
	errInvalidStartTime = errors.New("event start time cannot be less than 0")
	errDescTooLong      = errors.New("event description cannot be more than 80 chars")
)
