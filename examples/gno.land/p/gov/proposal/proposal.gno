// Package proposal provides a structure for executing proposals.
package proposal

import (
	"errors"
	"std"

	"gno.land/p/demo/context"
)

var errNotGovDAO = errors.New("only r/gov/dao can be the caller")

// NewExecutor creates a new executor with the provided callback function.
func NewExecutor(callback func() error) Executor {
	return &executorImpl{
		callback: callback,
		done:     false,
	}
}

// NewCtxExecutor creates a new executor with the provided callback function.
func NewCtxExecutor(callback func(ctx context.Context) error) Executor {
	return &executorImpl{
		callbackCtx: callback,
		done:        false,
	}
}

// executorImpl is an implementation of the Executor interface.
type executorImpl struct {
	callback    func() error
	callbackCtx func(ctx context.Context) error
	done        bool
	success     bool
}

// Execute runs the executor's callback function.
func (exec *executorImpl) Execute() error {
	if exec.done {
		return ErrAlreadyDone
	}

	// Verify the executor is r/gov/dao
	assertCalledByGovdao()

	var err error
	if exec.callback != nil {
		err = exec.callback()
	} else if exec.callbackCtx != nil {
		ctx := context.WithValue(context.Empty(), statusContextKey, approvedStatus)
		err = exec.callbackCtx(ctx)
	}
	exec.done = true
	exec.success = err == nil

	return err
}

// IsDone returns whether the executor has been executed.
func (exec *executorImpl) IsDone() bool {
	return exec.done
}

// IsSuccessful returns whether the execution was successful.
func (exec *executorImpl) IsSuccessful() bool {
	return exec.success
}

func (exec executorImpl) GetStatus() Status {
	switch {
	case exec.success:
		return Succeeded
	case exec.done:
		return Failed
	default:
		return NotExecuted
	}
}

func IsApprovedByGovdaoContext(ctx context.Context) bool {
	v := ctx.Value(statusContextKey)
	if v == nil {
		return false
	}
	vs, ok := v.(string)
	return ok && vs == approvedStatus
}

func AssertContextApprovedByGovDAO(ctx context.Context) {
	if !IsApprovedByGovdaoContext(ctx) {
		panic("not approved by govdao")
	}
}

// assertCalledByGovdao asserts that the calling Realm is /r/gov/dao
func assertCalledByGovdao() {
	caller := std.CurrentRealm().PkgPath()

	if caller != daoPkgPath {
		panic(errNotGovDAO)
	}
}

type propContextKey string

func (k propContextKey) String() string { return string(k) }

const (
	statusContextKey = propContextKey("govdao-prop-status")
	approvedStatus   = "approved"
)
