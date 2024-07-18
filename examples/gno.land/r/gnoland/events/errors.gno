package events

import "errors"

var (
	ErrEmptyName        = errors.New("event name cannot be empty")
	ErrInvalidStartTime = errors.New("event start time cannot be less than 0")
	ErrInvalidEndTime   = errors.New("event end time cannot be less than 0 or before start time")
)
