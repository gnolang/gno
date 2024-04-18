package proposal

// Executor represents a minimal closure-oriented proposal design, intended to be used by a govdao governance proposal (v1, v2, etc).
type Executor interface {
	execute() error
	Done() bool
	Success() bool // Done() && !err
}

var ErrAlreadyDone = errors.New("already executed")
