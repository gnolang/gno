package agent

import "gno.land/p/demo/agent/aatask"

type Task interface {
	Definition() aatask.Definition
	Finish(origCaller string)
	GetResult() (result aatask.Result, hasResult bool)
	ID() string
	MarshalJSON() ([]byte, error)
	SubmitResult(origCaller, value string)
}
