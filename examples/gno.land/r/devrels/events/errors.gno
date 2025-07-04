package events

import (
	"errors"
	"strconv"
)

var (
	ErrEmptyName                 = errors.New("event name cannot be empty")
	ErrNoSuchID                  = errors.New("event with specified ID does not exist")
	ErrMinWidgetSize             = errors.New("you need to request at least 1 event to render")
	ErrMaxWidgetSize             = errors.New("maximum number of events in widget is" + strconv.Itoa(MaxWidgetSize))
	ErrDescriptionTooLong        = errors.New("event description is too long")
	ErrInvalidStartTime          = errors.New("invalid start time format")
	ErrInvalidEndTime            = errors.New("invalid end time format")
	ErrEndBeforeStart            = errors.New("end time cannot be before start time")
	ErrStartEndTimezonemMismatch = errors.New("start and end timezones are not the same")
)
