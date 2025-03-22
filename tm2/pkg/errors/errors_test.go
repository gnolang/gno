package errors

import (
	"errors"
	fmt "fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorPanic(t *testing.T) {
	t.Parallel()

	type pnk struct {
		msg string
	}

	capturePanic := func() (err Error) {
		defer func() {
			if r := recover(); r != nil {
				err = Wrap(r, "This is the message in Wrap(r, message).")
			}
		}()
		panic(pnk{"something"})
	}

	err := capturePanic()

	assert.Equal(t, pnk{"something"}, err.Data())
	assert.Equal(t, "{something}", fmt.Sprintf("%v", err))
	assert.Contains(t, fmt.Sprintf("%#v", err), "This is the message in Wrap(r, message).")
	assert.Contains(t, fmt.Sprintf("%#v", err), "Stack Trace:\n    0")
}

func TestWrapSomething(t *testing.T) {
	t.Parallel()

	err := Wrapf("something", "formatter%v%v", 0, 1)

	assert.Equal(t, "something", err.Data())
	assert.Equal(t, "something", fmt.Sprintf("%v", err))
	assert.Regexp(t, `formatter01\n`, fmt.Sprintf("%#v", err))
	assert.Contains(t, fmt.Sprintf("%#v", err), "Stack Trace:\n    0")
}

func TestWrapNothing(t *testing.T) {
	t.Parallel()

	err := Wrapf(nil, "formatter%v%v", 0, 1)

	assert.Equal(t,
		FmtError{"formatter%v%v", []any{0, 1}},
		err.Data())
	assert.Equal(t, "formatter01", fmt.Sprintf("%v", err))
	assert.Contains(t, fmt.Sprintf("%#v", err), `Data: errors.FmtError{format:"formatter%v%v", args:[]interface {}{0, 1}}`)
	assert.Contains(t, fmt.Sprintf("%#v", err), "Stack Trace:\n    0")
}

func TestErrorNew(t *testing.T) {
	t.Parallel()

	err := New("formatter%v%v", 0, 1)

	assert.Equal(t,
		FmtError{"formatter%v%v", []any{0, 1}},
		err.Data())
	assert.Equal(t, "formatter01", fmt.Sprintf("%v", err))
	assert.Contains(t, fmt.Sprintf("%#v", err), `Data: errors.FmtError{format:"formatter%v%v", args:[]interface {}{0, 1}}`)
	assert.NotContains(t, fmt.Sprintf("%#v", err), "Stack Trace")
}

func TestErrorNewWithDetails(t *testing.T) {
	t.Parallel()

	err := New("formatter%v%v", 0, 1)
	err.Trace(0, "trace %v", 1)
	err.Trace(0, "trace %v", 2)
	err.Trace(0, "trace %v", 3)
	assert.Contains(t, fmt.Sprintf("%+v", err), `Data: formatter01`)
	assert.Contains(t, fmt.Sprintf("%+v", err), "Msg Traces:\n    0")
}

func TestErrorNewWithStacktrace(t *testing.T) {
	t.Parallel()

	err := New("formatter%v%v", 0, 1).Stacktrace()

	assert.Equal(t,
		FmtError{"formatter%v%v", []any{0, 1}},
		err.Data())
	assert.Equal(t, "formatter01", fmt.Sprintf("%v", err))
	assert.Contains(t, fmt.Sprintf("%#v", err), `Data: errors.FmtError{format:"formatter%v%v", args:[]interface {}{0, 1}}`)
	assert.Contains(t, fmt.Sprintf("%#v", err), "Stack Trace:\n    0")
}

func TestErrorNewWithTrace(t *testing.T) {
	t.Parallel()

	err := New("formatter%v%v", 0, 1)
	err.Trace(0, "trace %v", 1)
	err.Trace(0, "trace %v", 2)
	err.Trace(0, "trace %v", 3)

	assert.Equal(t,
		FmtError{"formatter%v%v", []any{0, 1}},
		err.Data())
	assert.Equal(t, "formatter01", fmt.Sprintf("%v", err))
	assert.Contains(t, fmt.Sprintf("%#v", err), `Data: errors.FmtError{format:"formatter%v%v", args:[]interface {}{0, 1}}`)
	dump := fmt.Sprintf("%#v", err)
	assert.NotContains(t, dump, "Stack Trace")
	assert.Regexp(t, `errors/errors_test\.go:[0-9]+ - trace 1`, dump)
	assert.Regexp(t, `errors/errors_test\.go:[0-9]+ - trace 2`, dump)
	assert.Regexp(t, `errors/errors_test\.go:[0-9]+ - trace 3`, dump)
}

func TestWrapError(t *testing.T) {
	t.Parallel()

	var err1 error = New("my message")
	var err2 error = Wrap(err1, "another message")
	assert.Equal(t, err1, err2)
	assert.True(t, errors.Is(err2, err1))

	err1 = fmt.Errorf("my message")
	err2 = Wrap(err1, "another message")
	assert.NotEqual(t, err1, err2)
	assert.True(t, errors.Is(err2, err1))
}
