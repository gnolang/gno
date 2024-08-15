package executor

import (
	"errors"
	"std"
)

var errInvalidCaller = errors.New("invalid executor caller")

// NewCallbackExecutor creates a new callback executor with the provided callback function
func NewCallbackExecutor(callback func() error, path string) *CallbackExecutor {
	return &CallbackExecutor{
		callback:   callback,
		daoPkgPath: path,
	}
}

// CallbackExecutor is an implementation of the dao.Executor interface,
// based on a specific callback.
// The given callback should verify the validity of the govdao call
type CallbackExecutor struct {
	callback   func() error // the callback to be executed
	daoPkgPath string       // the active pkg path of the govdao
}

// Execute runs the executor's callback function.
func (exec *CallbackExecutor) Execute() error {
	// Verify the caller is an adequate Realm
	caller := std.CurrentRealm().PkgPath()
	if caller != exec.daoPkgPath {
		return errInvalidCaller
	}

	if exec.callback != nil {
		return exec.callback()
	}

	return nil
}

// IsExpired returns a flag indicating if the executor is expired.
// The CallbackExecutor never expires
func (exec *CallbackExecutor) IsExpired() bool {
	return false
}
