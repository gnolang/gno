package std

import "errors"

const (
	ModuleName = "vm"
	RouterKey  = ModuleName
)
// std result
type Result struct {
	Response
	// GasWanted int64
	// GasUsed   int64
	// additional fields?
}

func GnoResult(r Response) *Result {
	return &Result{Response: r}
}

type Event string

func (e Event) AssertABCIEvent() {}

type Response struct {
	ErrMsg    string  // TODO: user err_code instead, for safety
	Data      []byte
	Events    []Event
	// gas limit?

	Log  string // nondeterministic
	Info string // nondeterministic
}

func NewResponse() Response {
	return Response{}
}

func (r Response) WithData(data []byte) Response {
	r.Data = data
	return r
}

func (r Response) WithEvents(events []Event) Response {
	r.Events = events
	return r
}

func (r Response) WithLog(log string) Response {
	r.Log = log
	return r
}

func (r Response) WithInfo(info string) Response {
	r.Info = info
	return r
}