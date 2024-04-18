package proposal

func NewExecutor(callback func() error) Executor {
	return &executorImpl{
		callback: callback,
		done:     false,
	}
}

type executorImpl struct {
	callback func() error
	done     bool
	success  bool
}

func (exec *executorImpl) execute() error {
	if exec.done {
		return ErrAlreadyDone
	}
	// XXX: assertCalledByGovdao
	err := exec.callback()
	exec.done = true
	exec.success = err == nil
	return err
}

func (exec *executorImpl) Done() bool {
	return exec.done
}

func (exec *executorImpl) Success() bool {
	return exec.success
}
