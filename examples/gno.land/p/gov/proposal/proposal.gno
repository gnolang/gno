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
}

func (exec *executorImpl) execute() error {
	if exec.done {
		return ErrAlreadyDone
	}
	// XXX: verify prevrealm or framestack is govdao
	return exec.callback()
}

func (exec *executorImpl) Done() bool {
	return exec.done
}
