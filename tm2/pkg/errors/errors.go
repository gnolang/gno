package errors

import (
	"fmt"
	"runtime"
)

// ----------------------------------------
// Convenience method.

func Wrap(cause interface{}, format string, args ...interface{}) Error {
	if causeCmnError, ok := cause.(*cmnError); ok { //nolint:gocritic
		msg := fmt.Sprintf(format, args...)
		return causeCmnError.Stacktrace().Trace(1, msg)
	} else if cause == nil {
		return newCmnError(FmtError{format, args}).Stacktrace()
	} else {
		// NOTE: causeCmnError is a typed nil here.
		msg := fmt.Sprintf(format, args...)
		return newCmnError(cause).Stacktrace().Trace(1, msg)
	}
}

func Cause(err error) error {
	if cerr, ok := err.(*cmnError); ok {
		return cerr.Data().(error)
	} else {
		return err
	}
}

// ----------------------------------------
// Error & cmnError

/*
Usage with arbitrary error data:

```go

	// Error construction
	type MyError struct{}
	var err1 error = NewWithData(MyError{}, "my message")
	...
	// Wrapping
	var err2 error  = Wrap(err1, "another message")
	if (err1 != err2) { panic("should be the same")
	...
	// Error handling
	switch err2.Data().(type){
		case MyError: ...
	    default: ...
	}

```
*/
type Error interface {
	Error() string
	Stacktrace() Error
	Trace(offset int, format string, args ...interface{}) Error
	Data() interface{}
}

// New Error with formatted message.
// The Error's Data will be a FmtError type.
func New(format string, args ...interface{}) Error {
	err := FmtError{format, args}
	return newCmnError(err)
}

// New Error with specified data.
func NewWithData(data interface{}) Error {
	return newCmnError(data)
}

type cmnError struct {
	data       interface{}    // associated data
	msgtraces  []msgtraceItem // all messages traced
	stacktrace []uintptr      // first stack trace
}

var _ Error = &cmnError{}

// NOTE: do not expose.
func newCmnError(data interface{}) *cmnError {
	return &cmnError{
		data:       data,
		msgtraces:  nil,
		stacktrace: nil,
	}
}

// Implements error.
func (err *cmnError) Error() string {
	return fmt.Sprintf("%v", err)
}

// Implements Unwrap method for compat with stdlib errors.Is()/As().
func (err *cmnError) Unwrap() error {
	if err.data == nil {
		return nil
	}
	werr, ok := err.data.(error)
	if !ok {
		return nil
	}
	return werr
}

// Captures a stacktrace if one was not already captured.
func (err *cmnError) Stacktrace() Error {
	if err.stacktrace == nil {
		offset := 3
		depth := 32
		err.stacktrace = captureStacktrace(offset, depth)
	}
	return err
}

// Add tracing information with msg.
// Set n=0 unless wrapped with some function, then n > 0.
func (err *cmnError) Trace(offset int, format string, args ...interface{}) Error {
	msg := fmt.Sprintf(format, args...)
	return err.doTrace(msg, offset)
}

// Return the "data" of this error.
// Data could be used for error handling/switching,
// or for holding general error/debug information.
func (err *cmnError) Data() interface{} {
	return err.data
}

func (err *cmnError) doTrace(msg string, n int) Error {
	// Ignoring linting on `runtime.Caller` for now, as it's
	// a critical method that can't be currently reworked
	//nolint:dogsled
	pc, _, _, _ := runtime.Caller(n + 2) // +1 for doTrace().  +1 for the caller.
	// Include file & line number & msg.
	// Do not include the whole stack trace.
	err.msgtraces = append(err.msgtraces, msgtraceItem{
		pc:  pc,
		msg: msg,
	})
	return err
}

func (err *cmnError) Format(s fmt.State, verb rune) {
	switch {
	case verb == 'p':
		s.Write([]byte(fmt.Sprintf("%p", &err)))
	case verb == 'v' && s.Flag('+'):
		s.Write([]byte("--= Error =--\n"))
		// Write data.
		fmt.Fprintf(s, "Data: %+v\n", err.data)
		// Write msg trace items.
		s.Write([]byte("Msg Traces:\n"))
		for i, msgtrace := range err.msgtraces {
			fmt.Fprintf(s, " %4d  %s\n", i, msgtrace.String())
		}
		s.Write([]byte("--= /Error =--\n"))
	case verb == 'v' && s.Flag('#'):
		s.Write([]byte("--= Error =--\n"))
		// Write data.
		fmt.Fprintf(s, "Data: %#v\n", err.data)
		// Write msg trace items.
		s.Write([]byte("Msg Traces:\n"))
		for i, msgtrace := range err.msgtraces {
			fmt.Fprintf(s, " %4d  %s\n", i, msgtrace.String())
		}
		// Write stack trace.
		if err.stacktrace != nil {
			s.Write([]byte("Stack Trace:\n"))
			frames := runtime.CallersFrames(err.stacktrace)
			for i := 0; ; i++ {
				frame, more := frames.Next()
				fmt.Fprintf(s, " %4d  %s:%d\n", i, frame.File, frame.Line)
				if !more {
					break
				}
			}
		}
		s.Write([]byte("--= /Error =--\n"))
	default:
		// Write msg.
		fmt.Fprintf(s, "%v", err.data)
	}
}

// ----------------------------------------
// stacktrace & msgtraceItem

func captureStacktrace(offset int, depth int) []uintptr {
	pcs := make([]uintptr, depth)
	n := runtime.Callers(offset, pcs)
	return pcs[0:n]
}

type msgtraceItem struct {
	pc  uintptr
	msg string
}

func (mti msgtraceItem) String() string {
	fnc := runtime.FuncForPC(mti.pc)
	file, line := fnc.FileLine(mti.pc)
	return fmt.Sprintf("%s:%d - %s",
		file, line,
		mti.msg,
	)
}

// ----------------------------------------
// fmt error

/*
FmtError is the data type for New() (e.g. New().Data().(FmtError))
Theoretically it could be used to switch on the format string.

```go

	// Error construction
	var err1 error = New("invalid username %v", "BOB")
	var err2 error = New("another kind of error")
	...
	// Error handling
	switch err1.Data().(cmn.FmtError).Format() {
		case "invalid username %v": ...
		case "another kind of error": ...
	    default: ...
	}

```
*/
type FmtError struct {
	format string
	args   []interface{}
}

func (fe FmtError) Error() string {
	if len(fe.args) == 0 {
		return fe.format
	}
	return fmt.Sprintf(fe.format, fe.args...)
}

func (fe FmtError) String() string {
	return fmt.Sprintf("FmtError{format:%v,args:%v}",
		fe.format, fe.args)
}

func (fe FmtError) Format() string {
	return fe.format
}
