package events

import (
	"errors"
	"strconv"
)

var (
	ErrEmptyName        = errors.New("event name cannot be empty")
	ErrInvalidStartTime = errors.New("event start time cannot be less than 0")
	ErrInvalidEndTime   = errors.New("event end time cannot be before start time")
	ErrNoSuchID         = errors.New("event with specified ID does not exist")
	ErrWidgetMinAmt     = errors.New("you need to request at least 1 event to render")
	ErrWidgetMaxAmt     = errors.New("maximum number of events in widget is" + strconv.Itoa(maxWidgetSize))
)
