// Package once provides utilities for one-time execution patterns.
// It extends the concept of sync.Once with error handling and panic options.
package once

import (
	"errors"
)

// Once represents a one-time execution guard
type Once struct {
	done    bool
	err     error
	paniced bool
	value   interface{} // stores the result of the execution
}

// New creates a new Once instance
func New() *Once {
	return &Once{}
}

// Do executes fn only once and returns nil on subsequent calls
func (o *Once) Do(fn func()) {
	if o.done {
		return
	}
	fn()
	o.done = true
}

// DoErr executes fn only once and returns the same error on subsequent calls
func (o *Once) DoErr(fn func() error) error {
	if o.done {
		return o.err
	}
	o.err = fn()
	o.done = true
	return o.err
}

// DoOrPanic executes fn only once and panics on subsequent calls
func (o *Once) DoOrPanic(fn func()) {
	if o.done {
		panic("once: multiple execution attempted")
	}
	fn()
	o.done = true
}

// DoValue executes fn only once and returns its value, subsequent calls return the cached value
func (o *Once) DoValue(fn func() interface{}) interface{} {
	if o.done {
		return o.value
	}
	o.value = fn()
	o.done = true
	return o.value
}

// DoValueErr executes fn only once and returns its value and error
// Subsequent calls return the cached value and error
func (o *Once) DoValueErr(fn func() (interface{}, error)) (interface{}, error) {
	if o.done {
		return o.value, o.err
	}
	o.value, o.err = fn()
	o.done = true
	return o.value, o.err
}

// Reset resets the Once instance to its initial state
// This is mainly useful for testing purposes
func (o *Once) Reset() {
	o.done = false
	o.err = nil
	o.paniced = false
	o.value = nil
}

// IsDone returns whether the Once has been executed
func (o *Once) IsDone() bool {
	return o.done
}

// Error returns the error from the last execution if any
func (o *Once) Error() error {
	return o.err
}

var (
	ErrNotExecuted = errors.New("once: not executed yet")
)

// Value returns the stored value and an error if not executed yet
func (o *Once) Value() (interface{}, error) {
	if !o.done {
		return nil, ErrNotExecuted
	}
	return o.value, nil
}
