package agent

import "github.com/gnolang/gno/agent/p/agent/task"

type Task interface {
	Definition() task.Definition
	Finish(origCaller string)
	GetResult() (result task.Result, hasResult bool)
	ID() string
	MarshalJSON() ([]byte, error)
	SubmitResult(origCaller, value string)
}
