// Package proposal defines types for proposal execution.
package proposal

import "errors"

// Executor represents a minimal closure-oriented proposal design.
// It is intended to be used by a govdao governance proposal (v1, v2, etc).
type Executor interface {
	Execute() error
	Done() bool
	Success() bool // Done() && !err
}

// ErrAlreadyDone is the error returned when trying to execute an already
// executed proposal.
var ErrAlreadyDone = errors.New("already executed")
