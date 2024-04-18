package proposal

type Executor interface {
	execute() error
	Done() bool
}

var ErrAlreadyDone = errors.New("already executed")
