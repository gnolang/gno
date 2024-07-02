// uassert is an adapted lighter version of https://github.com/stretchr/testify/assert.
package uassert

import "std"

// NoError asserts that a function returned no error (i.e. `nil`).
func NoError(t TestingT, err error, msgs ...string) bool {
	t.Helper()
	if err != nil {
		return fail(t, msgs, "unexpected error: %s", err.Error())
	}
	return true
}

// Error asserts that a function returned an error (i.e. not `nil`).
func Error(t TestingT, err error, msgs ...string) bool {
	t.Helper()
	if err == nil {
		return fail(t, msgs, "an error is expected but got nil")
	}
	return true
}

// True asserts that the specified value is true.
func True(t TestingT, value bool, msgs ...string) bool {
	t.Helper()
	if !value {
		return fail(t, msgs, "should be true")
	}
	return true
}

// False asserts that the specified value is false.
func False(t TestingT, value bool, msgs ...string) bool {
	t.Helper()
	if value {
		return fail(t, msgs, "should be false")
	}
	return true
}

// ErrorIs asserts the given error matches the target error
func ErrorIs(t TestingT, err, target error, msgs ...string) bool {
	t.Helper()

	if err == nil || target == nil {
		return err == target
	}

	// XXX: if errors.Is(err, target) return true

	if err.Error() != target.Error() {
		return fail(t, msgs, "error mismatch, expected %s, got %s", target.Error(), err.Error())
	}

	return true
}

// PanicsWithError asserts that the code inside the specified func panics,
// and that the recovered panic value is an error that satisfies the given message
func PanicsWithError(t TestingT, errString string, f func(), msgs ...string) bool {
	t.Helper()

	didPanic, panicValue := checkDidPanic(f)
	if !didPanic {
		return fail(t, msgs, "func should panic\n\tPanic value:\t%v", panicValue)
	}

	if panicValue != errString {
		return fail(t, msgs, "func should panic with error:\t%s\n\tPanic value:\t%s", panicValue, errString)
	}
	return true
}

// NotPanics asserts that the code inside the specified func does NOT panic.
func NotPanics(t TestingT, f func(), msgs ...string) bool {
	t.Helper()

	didPanic, panicValue := checkDidPanic(f)

	if didPanic {
		return fail(t, msgs, "func should not panic\n\tPanic value:\t%s", panicValue)
	}
	return true
}

// Equal asserts that two objects are equal.
func Equal(t TestingT, expected, actual interface{}, msgs ...string) bool {
	t.Helper()

	if expected == nil || actual == nil {
		return expected == actual
	}

	// XXX: errors
	// XXX: slices
	// XXX: pointers

	switch ev := expected.(type) {
	case string:
		av, ok := actual.(string)
		return ok && ev == av
	case int:
		av, ok := actual.(int)
		return ok && ev == av
	case int8:
		av, ok := actual.(int8)
		return ok && ev == av
	case int16:
		av, ok := actual.(int16)
		return ok && ev == av
	case int32:
		av, ok := actual.(int32)
		return ok && ev == av
	case int64:
		av, ok := actual.(int64)
		return ok && ev == av
	case uint:
		av, ok := actual.(uint)
		return ok && ev == av
	case uint8:
		av, ok := actual.(uint8)
		return ok && ev == av
	case uint16:
		av, ok := actual.(uint16)
		return ok && ev == av
	case uint32:
		av, ok := actual.(uint32)
		return ok && ev == av
	case uint64:
		av, ok := actual.(uint64)
		return ok && ev == av
	case bool:
		av, ok := actual.(bool)
		return ok && ev == av
	case float32:
		av, ok := actual.(float32)
		return ok && ev == av
	case float64:
		av, ok := actual.(float64)
		return ok && ev == av
	default:
		fail(t, msgs, "uassert.Equal: unsupported type")
	}

	return false
}
