// Package proposal provides a structure for executing proposals.
package proposal

// NewExecutor creates a new executor with the provided callback function.
func NewExecutor(callback func() error) Executor {
	return &executorImpl{
		callback: callback,
		done:     false,
	}
}

// executorImpl is an implementation of the Executor interface.
type executorImpl struct {
	callback func() error
	done     bool
	success  bool
}

// execute runs the executor's callback function.
func (exec *executorImpl) Execute() error {
	if exec.done {
		return ErrAlreadyDone
	}
	// XXX: assertCalledByGovdao
	err := exec.callback()
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
