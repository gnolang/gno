package executor

import (
	"errors"
	"std"

	"gno.land/r/sys/vars"
)

var errNotGovDAO = errors.New("only r/gov/dao can be the caller")

// NewCallbackExecutor creates a new callback executor with the provided callback function
func NewCallbackExecutor(callback func() error) *CallbackExecutor {
	return &CallbackExecutor{
		callback: callback,
	}
}

// CallbackExecutor is an implementation of the dao.Executor interface,
// based on a specific callback.
// The given callback should verify the validity of the govdao call
type CallbackExecutor struct {
	callback func() error // the callback to be executed
}

// Execute runs the executor's callback function.
func (exec *CallbackExecutor) Execute() error {
	// Verify the executor is r/gov/dao
	assertCalledByGovdao()

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

// assertCalledByGovdao asserts that the calling Realm is /r/gov/dao
func assertCalledByGovdao() {
	caller := std.CurrentRealm().PkgPath()

	if caller != vars.GetStringValue(vars.GovdaoRealm) {
		panic(errNotGovDAO)
	}
}
