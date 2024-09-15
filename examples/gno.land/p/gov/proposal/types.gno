// Package proposal defines types for proposal execution.
package proposal

import "errors"

// Executor represents a minimal closure-oriented proposal design.
// It is intended to be used by a govdao governance proposal (v1, v2, etc).
type Executor interface {
	// Execute executes the given proposal, and returns any error encountered
	// during the execution
	Execute() error

	// IsDone returns a flag indicating if the proposal was executed
	IsDone() bool

	// IsSuccessful returns a flag indicating if the proposal was executed
	// and is successful
	IsSuccessful() bool // IsDone() && !err

	// IsExpired returns whether the execution had expired or not.
	IsExpired() bool
}

// ErrAlreadyDone is the error returned when trying to execute an already
// executed proposal.
var ErrAlreadyDone = errors.New("already executed")

// Status enum.
type Status string

const (
	NotExecuted Status = "not_executed"
	Succeeded   Status = "succeeded"
	Failed      Status = "failed"
)

const daoPkgPath = "gno.land/r/gov/dao" // TODO: make sure this is configurable through r/sys/vars
