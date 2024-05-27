package queue

import (
	"errors"
)

// Error definitions for various edge cases
var (
	ErrResourceError    = errors.New("resource error: queue is full")
	ErrEmptyArrayPop    = errors.New("empty array pop: queue is empty")
	ErrArrayOutOfBounds = errors.New("array out of bounds")

	ErrNonExistedValue = errors.New("non-existed value")
)
