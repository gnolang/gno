// Package proposal provides a structure for executing proposals.
package proposal

import (
	"std"

	"gno.land/p/demo/context"
)

const daoPkgPath = "gno.land/r/gov/dao" // XXX: make it configurable with r/sys/vars?

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

// execute runs the executor's callback function.
func (exec *executorImpl) Execute() error {
	if exec.done {
		return ErrAlreadyDone
	}
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

// Done returns whether the executor has been executed.
func (exec *executorImpl) Done() bool {
	return exec.done
}

// Success returns whether the execution was successful.
func (exec *executorImpl) Success() bool {
	return exec.success
}

func (exec executorImpl) Status() Status {
	switch {
	case exec.success:
		return Success
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
	return v.(string) == approvedStatus
}

func AssertContextApprovedByGovDAO(ctx context.Context) {
	if !IsApprovedByGovdaoContext(ctx) {
		panic("not approved by govdao")
	}
}

func assertCalledByGovdao() {
	caller := std.CurrentRealm().PkgPath()
	if caller != daoPkgPath {
		panic("only gov/dao can execute proposals")
	}
}

type propContextKey string

func (k propContextKey) String() string { return string(k) }

const (
	statusContextKey = propContextKey("govdao-prop-status")
	approvedStatus   = "approved"
)
